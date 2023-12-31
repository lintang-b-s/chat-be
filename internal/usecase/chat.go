package usecase

import (
	"context"
	"encoding/json"
	"github.com/gorilla/websocket"
	"github.com/lintangbs/chat-be/internal/entity"
	sonyflake2 "github.com/lintangbs/chat-be/internal/util/sonyflake"
	"github.com/sony/sonyflake"
	"sort"

	"github.com/lintangbs/chat-be/pkg/redispkg"
	"log"
	"strings"
	"sync"
	"time"
)

// User represents user connection.
// It contains logic of receiving and sending messages.
// That is, there are no active reader or writer. Some other layer of the
// application should call Receive() to read user's incoming message.
type User struct {
	io   sync.Mutex
	Conn *websocket.Conn

	Id   uint
	Name string
	Chat *ChatHub

	inbox chan *entity.MessageWs
}

var sf *sonyflake.Sonyflake

// ChatHub utk menyimpan semua client websocket yang terhubung ke chat-server ini
type ChatHub struct {
	mu        sync.RWMutex
	seq       uint
	PubSub    PubSubRedis
	Rds       *redispkg.Redis
	edenAiApi EdenAiApi
	userPg    UserRepo
	usrRedis  UserRedisRepo
	pChat     PrivateChatRepo
	idGen     sonyflake2.IdGenerator
	gpRepo    GroupRepo
	gcRepo    GroupChatRepo

	us        []*User
	broadcast chan *entity.MessageWs

	// Register requests from the clients.
	register chan *User

	// Unregister requests from clients.
	unregister chan *User
}

func NewChat(pubSub PubSubRedis,
	ed EdenAiApi,
	userPg UserRepo,
	rds *redispkg.Redis,
	ud UserRedisRepo,
	pc PrivateChatRepo,
	idGen sonyflake2.IdGenerator,
	gpRepo GroupRepo,
	gcRepo GroupChatRepo,
) *ChatHub {

	return &ChatHub{PubSub: pubSub,

		edenAiApi:  ed,
		userPg:     userPg,
		Rds:        rds,
		usrRedis:   ud,
		broadcast:  make(chan *entity.MessageWs),
		unregister: make(chan *User),
		register:   make(chan *User),
		pChat:      pc,
		idGen:      idGen,
		gpRepo:     gpRepo,
		gcRepo:     gcRepo,
	}
}

func (c *ChatHub) Run() {
	for {
		select {
		case user := <-c.register:
			user.Id = c.seq
			user.Name = user.Name

			c.us = append(c.us, user)
			c.seq++

		case user := <-c.unregister:
			c.mu.Lock()
			// binary search utk cari index user di array us
			i := sort.Search(len(c.us), func(i int) bool {
				return c.us[i].Id >= user.Id
			})

			// hapus client dari array chat.us
			without := make([]*User, len(c.us)-1) // us = nil
			copy(without[:i], c.us[:i])
			copy(without[i:], c.us[i+1:])
			c.us = without
			c.mu.Unlock()

		case message := <-c.broadcast:
			// menerima message da	ri user lain yg chat-servernya sama dg user
			// mengirim ke user dg username sama dg recipient username di messageWs
			c.sendToSpecificUserInboxInServer(message)
		}
	}
}

func (c *ChatHub) sendToSpecificUserInboxInServer(message *entity.MessageWs) {
	for _, user := range c.us {

		// mengirim ke user dg username sama dg recipient username di messageWs
		recipientUsername := message.PrivateChat.RecipientUsername
		rcpFanoutUsername := message.MsgOnlineStatusFanout.UserToGetNotified
		rcpGroupChat := message.MsgGroupChat.RecipientUsername
		rcpGroupChatBot := message.MsgGroupChatBot.RecipientUsername

		switch message.Type {
		case entity.MessageTypePrivateChat:
			if user.Name == recipientUsername {
				select {
				case user.inbox <- message:
				}
			}
		case entity.MessageTypeOnlineStatusFanOut:
			if user.Name == rcpFanoutUsername {
				select {
				case user.inbox <- message:
				}
			}

		case entity.MessageTypeGroupChat:
			if user.Name == rcpGroupChat {
				select {
				case user.inbox <- message:
				}
			}
		case entity.MessageTypeGroupChatBot:
			if user.Name == rcpGroupChatBot {
				select {
				case user.inbox <- message:
				}
			}
		}
	}
}

var (
	// pongWait : berapa lama server menunggu message pong dari client (30 detik)
	pongWait = 30 * time.Second
	// pingInterval : setiap 5 detik server mengirim ping message ke client.
	// pingInterval haruslah lebih kecil dari pongWait
	pingInterval = (pongWait * 5) / 10
)

// Receive membaca next message websocket dari client
// It blocks until full message received.
func (u *User) Receive() error {

	defer func() {
		u.Chat.unregister <- u
		u.Chat.userOnlineStatusFanout(u.Name, false)
		u.Conn.Close()
	}()
	// Set Max Size of Messages in Bytes
	u.Conn.SetReadLimit(1024)

	// Konfigurasi waktu tunggu pong response, pong response dari client harus dalam kurun waktu time  + pongWait (10 detiK)
	// initial timernya di pemanggilan fungsi ini.
	// Jika dalam 30 detik client tidak membalas ping message dg pong meessage, koneksi websocket dg client di close
	if err := u.Conn.SetReadDeadline(time.Now().Add(pongWait)); err != nil {
		log.Println(err)
		u.Chat.unregister <- u
		u.Chat.userOnlineStatusFanout(u.Name, false)
		u.Conn.Close()
		return err
	}

	u.Conn.SetPongHandler(u.pongHandler)

	for {

		// ReadMessage dari client websocket
		_, msg, err := u.Conn.ReadMessage()

		if err != nil {
			break
		}

		msgWs := &entity.MessageWs{}
		if err = json.Unmarshal(msg, msgWs); err != nil {
			log.Println("json.Unmarshal: ", err)
			break
		}

		switch msgWs.Type {
		case entity.MessageTypePrivateChatBot:
			// jika tipe message dari frontend adalah privateChatBot
			resText, err := u.Chat.edenAiApi.GenerateText(msgWs.PrivateChat.Message)
			if err != nil {
				msgWs.PrivateChat.MessageId, _ = u.Chat.idGen.GenerateId()
				msgWs.PrivateChat.CreatedAt = time.Now()
				msgWs.PrivateChat.Message = err.Error()
				err = u.Write(websocket.TextMessage, msgWs)
				continue
			}
			msgWs.PrivateChat.MessageId, _ = u.Chat.idGen.GenerateId()
			msgWs.PrivateChat.CreatedAt = time.Now()
			msgWs.PrivateChat.SenderUsername = "ChatBot-edenAI-GPT"
			msgWs.PrivateChat.Message = resText
			err = u.Write(websocket.TextMessage, msgWs)
			if err != nil {
				log.Println("Write: ", err)
				continue
			}
		case entity.MessageTypePrivateChat:
			// jika tipe message dari frontend private chat dg user lain yang sudah ditambahkan kontaknya
			msgWs.PrivateChat.MessageId, _ = u.Chat.idGen.GenerateId()
			msgWs.PrivateChat.CreatedAt = time.Now()
			friendUsername := msgWs.PrivateChat.RecipientUsername

			isFriendErr := u.Chat.userPg.GetUserFriend(
				context.Background(),
				msgWs.PrivateChat.SenderUsername,
				friendUsername,
			)
			if isFriendErr != nil {
				log.Println("u.Chat.pg.GetUserFriend: ", isFriendErr)

				msgWs.PrivateChat.Message = msgWs.PrivateChat.RecipientUsername + " is not your friend"
				err = u.Write(websocket.TextMessage, msgWs)
				if err != nil {
					log.Println("Write: ", err)
					continue
				}
				continue

			}
			friend, _ := u.Chat.userPg.GetUserByUsername(friendUsername)
			isFriendInSameServer, friendServerLocation := u.Chat.isFriendInSameServer(friend.Id.String())
			sender, err := u.Chat.userPg.GetUserByUsername(msgWs.PrivateChat.SenderUsername)

			//	 Save Private Chat message to db
			pc := entity.InsertPrivateChatRequest{
				MessageId:   msgWs.PrivateChat.MessageId,
				MessageTo:   friend.Id,
				MessageFrom: sender.Id,
				Content:     msgWs.PrivateChat.Message,
			}
			_, err = u.Chat.pChat.InsertPrivateChat(pc)
			if err != nil {
				log.Println("Recive() - u.Chat.pChat.InsertPrivateChat:", err)
			}
			if isFriendInSameServer == true {
				// Jika friend/recipient message berada di chat-server yg sama dg chat-server user
				u.Chat.broadcast <- msgWs
				continue
			}

			// jika teman user berada di server yg berbeda dg server user sender
			// publish ke chat-server teman
			u.Chat.PubSub.PublishToChannel(friendServerLocation, msgWs)
		case entity.MessageTypeGroupChat:
			// Jika tipe message dari frontend adalah group chat
			msgWs.MsgGroupChat.MessageId, _ = u.Chat.idGen.GenerateId() // generate message id menggunakan sonyflake
			// mendapatkan entitas user sender dari db
			sender, err := u.Chat.userPg.GetUserByUsername(msgWs.MsgGroupChat.SenderUsername)
			if err != nil {
				msgWs.MsgGroupChat.Content = err.Error()
				msgWs.PrivateChat.CreatedAt = time.Now()
				u.Write(websocket.TextMessage, msgWs)
				continue
			}
			// mendapatkan entitas group dari db
			groupDb, err := u.Chat.gpRepo.GetGroupByName(msgWs.MsgGroupChat.GroupName, sender.Id)
			if err != nil {
				msgWs.MsgGroupChat.Content = err.Error()
				msgWs.PrivateChat.CreatedAt = time.Now()
				u.Write(websocket.TextMessage, msgWs)
				continue
			}
			// mendapatkan groupchat members
			group, err := u.Chat.gpRepo.GetGroupMembers(groupDb.Id, sender.Id)
			if err != nil {
				msgWs.MsgGroupChat.Content = err.Error()
				msgWs.PrivateChat.CreatedAt = time.Now()
				u.Write(websocket.TextMessage, msgWs)
				continue
			}

			gcMessageDb := entity.GroupChatMessage{
				GroupId:   groupDb.Id,
				MessageId: msgWs.MsgGroupChat.MessageId,
				UserId:    sender.Id,
				Content:   msgWs.MsgGroupChat.Content,
			}
			// inser chat ke table groupchat
			u.Chat.gcRepo.InsertNewChat(gcMessageDb)

			// fanout message ke semua member group chat
			groupMembers := group.Members
			for _, memberId := range groupMembers {
				friend, _ := u.Chat.userPg.GetUserById(memberId)
				isFriendInSameServer, friendServerLocation := u.Chat.isFriendInSameServer(memberId.String())
				msgWs.MsgGroupChat.RecipientUsername = friend.Username
				if friend.Username == msgWs.MsgGroupChat.SenderUsername {
					continue
				}
				if isFriendInSameServer == true {
					// Jika friend/recipient message berada di chat-server yg sama dg chat-server user
					u.Chat.broadcast <- msgWs
					continue
				}

				// jika teman user berada di server yg berbeda dg server user sender
				// publish ke chat-server teman
				u.Chat.PubSub.PublishToChannel(friendServerLocation, msgWs)
			}

		case entity.MessageTypeGroupChatBot:
			msgWs.MsgGroupChatBot.MessageId, _ = u.Chat.idGen.GenerateId() // generate message id menggunakan sonyflake
			msgWs.MsgGroupChatBot.CreatedAt = time.Now()
			resTextChatBot, err := u.Chat.edenAiApi.GenerateText(msgWs.MsgGroupChatBot.Content)
			if err != nil {
				err = u.Write(websocket.TextMessage, msgWs)
				continue
			}
			sender, err := u.Chat.userPg.GetUserByUsername(msgWs.MsgGroupChatBot.SenderUsername)
			if err != nil {
				msgWs.MsgGroupChatBot.Content = err.Error()
				u.Write(websocket.TextMessage, msgWs)
			}
			groupDb, err := u.Chat.gpRepo.GetGroupByName(msgWs.MsgGroupChatBot.GroupName, sender.Id)
			if err != nil {
				msgWs.MsgGroupChatBot.Content = err.Error()
				u.Write(websocket.TextMessage, msgWs)
				continue
			}
			group, err := u.Chat.gpRepo.GetGroupMembers(groupDb.Id, sender.Id)
			if err != nil {
				msgWs.MsgGroupChatBot.Content = err.Error()
				u.Write(websocket.TextMessage, msgWs)
				continue
			}

			// insert message prompt dari user
			gcMessageDb := entity.GroupChatMessage{
				GroupId:   groupDb.Id,
				MessageId: msgWs.MsgGroupChatBot.MessageId,
				UserId:    sender.Id,
				Content:   msgWs.MsgGroupChatBot.Content,
			}
			u.Chat.gcRepo.InsertNewChat(gcMessageDb)

			// insert message jawaban chatbot
			newMsgId, _ := u.Chat.idGen.GenerateId()
			gcMessageDb = entity.GroupChatMessage{
				GroupId:   groupDb.Id,
				MessageId: newMsgId,
				UserId:    sender.Id,
				Content:   resTextChatBot,
			}
			u.Chat.gcRepo.InsertNewChat(gcMessageDb)

			//	fanout messsage ke semua member group
			groupMembers := group.Members
			msgWs.MsgGroupChatBot.SenderUsername = "ChatBot-edenAI-GPT"
			msgWs.MsgGroupChatBot.Content = resTextChatBot
			u.Write(websocket.TextMessage, msgWs)
			for _, memberId := range groupMembers {
				friend, _ := u.Chat.userPg.GetUserById(memberId)
				isFriendInSameServer, friendServerLocation := u.Chat.isFriendInSameServer(memberId.String())
				msgWs.MsgGroupChatBot.RecipientUsername = friend.Username
				if friend.Username == msgWs.MsgGroupChatBot.SenderUsername {
					continue
				}
				if isFriendInSameServer == true {
					// Jika friend/recipient message berada di chat-server yg sama dg chat-server user
					u.Chat.broadcast <- msgWs
					continue
				}

				// jika teman user berada di server yg berbeda dg server user sender
				// publish ke chat-server teman
				u.Chat.PubSub.PublishToChannel(friendServerLocation, msgWs)
			}

		}
	}
	return nil
}

// pongHandler handle message pong yang dikirim oleh client
// -> mereset durasi readDeadline (tambah 30 detik lagi) & set user online di dalam redis
// dan juga mengirim status online user ke semua kontaknya
func (u *User) pongHandler(pongMsg string) error {

	user, err := u.Chat.userPg.GetUserByUsername(u.Name)
	if err != nil {
		return err
	}

	// Set User online in Redis
	u.Chat.usrRedis.UserSetOnline(user.Id.String())

	// Fanout User Online Status ke semua kontaknya
	u.Chat.userOnlineStatusFanout(u.Name, true)

	return u.Conn.SetReadDeadline(time.Now().Add(pongWait))
}

// isFriendInSameServer  Jika friend/recipient message berada di chat-server yg sama dg chat-server user sender
// return bool, friendServerLocation
func (c *ChatHub) isFriendInSameServer(friendId string) (bool, string) {
	friendServerLocation, _ := c.usrRedis.GetUserServerLocation(friendId)
	isFriendOnline := c.usrRedis.UserIsOnline(friendId)
	if friendServerLocation == entity.ChatServerNameGlobal.ChatServerName && isFriendOnline == true {
		// Jika teman user berada di server yg sama dg server user sender
		return true, ""
	}
	return false, friendServerLocation
}

// userOnlineStatusFanout fanout user online status ke semua kontaknya
func (c *ChatHub) userOnlineStatusFanout(username string, online bool) {
	// Fanout User Online Status ke semua kontaknya
	userDb, _ := c.userPg.GetUserFriends(context.Background(), username)
	for _, uFriend := range userDb.Friends {
		// send notification ke semua kontaknya bahwa user masih online
		// user yg online (user yang mengirim pong message)
		msgOnlineStatusFanout := entity.MessageOnlineStatusFanout{
			FriendId:          userDb.Id.String(),
			FriendUsername:    userDb.Username,
			FriendEmail:       userDb.Email,
			Online:            online,
			UserToGetNotified: uFriend.Username,
		}
		msgWs := &entity.MessageWs{
			Type:                  entity.MessageTypeOnlineStatusFanOut,
			MsgOnlineStatusFanout: msgOnlineStatusFanout,
		}
		isFriendInSameServer, friendServerLocation := c.isFriendInSameServer(uFriend.Id.String())
		if isFriendInSameServer == true {
			// Jika friend/recipient message berada di chat-server yg sama dg chat-server user sender
			c.broadcast <- msgWs
			continue
		}

		// jika chat server friend/recipient berbeda dg chat-server user sender
		c.PubSub.PublishToChannel(friendServerLocation, msgWs)
	}
}

// getAllFriendsOnlineStatus user akan mendapatkan status online/tidaknya semua teman/kontaknya
func (u *User) getAllFriendsOnlineStatus(ctx context.Context, username string) {
	userDb, _ := u.Chat.userPg.GetUserFriends(ctx, username)
	totFriend := len(userDb.Friends)
	totOnline := 0

	var messageWsFriendsStatus entity.MessageFriendsOnlineStatus
	var friends []entity.Friend
	// Set online status setiap kontak/teman  user
	for _, uFriend := range userDb.Friends {
		isFriendOnline := u.Chat.usrRedis.UserIsOnline(uFriend.Id.String())
		if isFriendOnline == true {
			totOnline += 1
		}

		friends = append(friends, entity.Friend{
			FriendId:       uFriend.Id.String(),
			FriendUsername: uFriend.Username,
			FriendEmail:    uFriend.Email,
			Online:         isFriendOnline,
		})
	}
	messageWsFriendsStatus.TotalFriends = totFriend
	messageWsFriendsStatus.TotalOnline = totOnline
	messageWsFriendsStatus.Friends = friends

	messageWs := &entity.MessageWs{
		Type:                   entity.MessageTypeFriendsOnlineStatus,
		MsgFriendsOnlineStatus: messageWsFriendsStatus,
	}

	u.Write(websocket.TextMessage, messageWs)
}

// Register registers new connection as a User.
func (c *ChatHub) Register(ctx context.Context, conn *websocket.Conn, username string, userId string,
) *User {
	user := &User{
		Chat:  c,
		Conn:  conn,
		inbox: make(chan *entity.MessageWs),
		Name:  username,
	}

	user.Chat.register <- user

	// Register user chat-server location in redis
	c.usrRedis.SetUserServerLocation(userId)

	// Set User Online status (key,value) in redis
	c.usrRedis.UserSetOnline(userId)

	// fanout user online status ke semua kontaknya
	c.userOnlineStatusFanout(username, true)

	// Get online status semua kontak yang dimiliki user
	user.getAllFriendsOnlineStatus(ctx, username)

	// gorotuine untuk membaca message websocket yang dikriim dari frontend
	go user.Receive()
	// goroutine untuk menulis message websocket ke frontend
	go user.writePump()

	return user
}

// SubscribePubSubAndSendToClient Subscribe ke channel chat-servernya lalu mempublish message ke specific user inbox
// subcribe pubsub redis jika recipient berada di chat-server berbeda dg sender
func (c *ChatHub) SubscribePubSubAndSendToClient(channelRedis *redispkg.ChannelPubSub) {
	defer channelRedis.Closed()
	for {
		select {
		case data := <-channelRedis.Channel():

			msg := &entity.MessageWs{}
			dec := json.NewDecoder(strings.NewReader(data.Payload))

			err := dec.Decode(msg)

			if err != nil {
				log.Println("dec.Decode", err)
				return
			} else {

				c.sendToSpecificUserInboxInServer(msg)
			}
		}
	}
}

// writePump mengirim message websocket ke user/client/frontend
// 1 goroutine yg menjalankan writePump dijalankan di setiap koneksi client websocket.
func (u *User) writePump() {

	// Create a ticker that triggers a ping at given interval
	ticker := time.NewTicker(pingInterval)

	defer func() {
		ticker.Stop()
		u.Chat.userOnlineStatusFanout(u.Name, false)
		u.Conn.Close()
	}()
	for {

		select {

		case <-ticker.C:

			err := u.Conn.WriteMessage(websocket.PingMessage, []byte{})
			if err != nil {
				log.Println("wsutil.WriteServerMessage", err)
				return
			}
		case msgWs := <-u.inbox:
			// menerima message dari inbox user, llau send wesbsocket message to client/user
			u.Write(websocket.TextMessage, msgWs)
		}
	}
}

// Write mengirim messsage websocket ke client/frontend
func (u *User) Write(op int, message *entity.MessageWs) error {
	data, err := json.Marshal(message)
	if err != nil {
		log.Println("json.Marshal", err)
		return err
	}
	u.Conn.WriteMessage(op, data)
	if err != nil {
		log.Println("json.Marshal", err)
		return err
	}

	return nil
}
