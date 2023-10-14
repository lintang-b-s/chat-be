package websocketc

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/google/uuid"
	"github.com/gorilla/websocket"
	"github.com/lintangbs/chat-be/internal/entity"
	"github.com/lintangbs/chat-be/internal/usecase/redisRepo"

	"github.com/lintangbs/chat-be/internal/usecase/repo"
	"github.com/lintangbs/chat-be/internal/usecase/webapi"
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
	Chat *Chat
}

type Chat struct {
	mu        sync.RWMutex
	seq       uint
	pubSub    redisRepo.PubSubRedis
	rds       redispkg.Redis
	edenAiApi webapi.EdenAIAPI
	userPg    repo.UserRepo
	usrRedis  redisRepo.UserRedisRepo
}

func NewChat(pubSub redisRepo.PubSubRedis,
	ed webapi.EdenAIAPI,
	userPg repo.UserRepo,
	rds redispkg.Redis,
	ud redisRepo.UserRedisRepo,
) *Chat {
	return &Chat{pubSub: pubSub, edenAiApi: ed, userPg: userPg, rds: rds, usrRedis: ud}
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

	// Set Max Size of Messages in Bytes
	u.Conn.SetReadLimit(1024)

	// Konfigurasi waktu tunggu pong response, pong response dari client harus dalam kurun waktu time  + pongWait (10 detiK)
	// initial timernya di pemanggilan fungsi ini.
	// Jika dalam 30 detik client tidak membalas ping message dg pong meessage, koneksi websocket dg client di close
	if err := u.Conn.SetReadDeadline(time.Now().Add(pongWait)); err != nil {
		log.Println(err)
		u.Conn.Close()
		return err
	}

	u.Conn.SetPongHandler(u.pongHandler)

	// ReadMessage dari client websocket
	_, msg, err := u.Conn.ReadMessage()

	if err != nil {
		u.Conn.Close()
		return err
	}

	msgWs := &entity.MessageWs{}
	if err = json.Unmarshal(msg, msgWs); err != nil {
		log.Println("json.Unmarshal: ", err)
		return err
	}

	switch msgWs.Type {
	case entity.MessageTypePrivateChatBot:
		resText := u.Chat.edenAiApi.GenerateText(msgWs.PrivateChat.Message)
		msgWs.PrivateChat.MessageId = uuid.New().String()
		msgWs.PrivateChat.CreatedAt = time.Now()
		msgWs.PrivateChat.Message = resText
		err = u.Write(websocket.TextMessage, msgWs)
		if err != nil {
			log.Println("Write: ", err)
			return err
		}
	case entity.MessageTypePrivateChat:
		// private chat dg user lain yang sudah ditambahkan kontaknya
		msgWs.PrivateChat.MessageId = uuid.New().String()
		msgWs.PrivateChat.CreatedAt = time.Now()
		isFriendErr := u.Chat.userPg.GetUserFriend(
			context.Background(),
			msgWs.PrivateChat.SenderUsername,
			msgWs.PrivateChat.RecipientUsername,
		)
		if isFriendErr != nil {
			fmt.Println("u.Chat.pg.GetUserFriend: ", isFriendErr)

			msgWs.PrivateChat.Message = msgWs.PrivateChat.RecipientUsername + " is not your friend"
			err = u.Write(websocket.TextMessage, msgWs)
			if err != nil {
				log.Println("Write: ", err)
				return err
			}
			return isFriendErr

		}
		err = u.Chat.Broadcast(msgWs.PrivateChat.RecipientUsername, msgWs)
		if err != nil {
			fmt.Println("u.chat.Broadcast: ", err)
			return err
		}
	}

	return nil
}

// pongHandler handle message pong yang dikirim oleh client
// -> mereset durasi readDeadline (tambah 30 detik lagi) & set user online di dalam redis
// dan juga mengirim status online user ke semua kontaknya
func (u *User) pongHandler(pongMsg string) error {
	log.Println("pong received from client " + u.Name + " !!")

	user, err := u.Chat.userPg.GetUserByUsername(u.Name)
	if err != nil {
		return err
	}

	// Set User online in Redis
	u.Chat.usrRedis.UserSetOnline(user.Id.String())

	// Fanout User Online Status ke semua kontaknya
	uFriends, err := u.Chat.userPg.GetUserFriends(context.Background(), u.Name)
	for _, uFriend := range uFriends.Friends {
		msgOnlineStatusFanout := entity.MessageOnlineStatusFanout{
			FriendId:       uFriend.Id.String(),
			FriendUsername: uFriends.Username,
			FriendEmail:    uFriends.Email,
			Online:         true,
		}
		msgWs := &entity.MessageWs{
			Type:                  entity.MessageTypeOnlineStatusFanOut,
			MsgOnlineStatusFanout: msgOnlineStatusFanout,
		}
		u.Chat.pubSub.PublishToChannel(uFriend.Username, msgWs)
	}

	return u.Conn.SetReadDeadline(time.Now().Add(pongWait))
}

// Broadcast sends message to all alive users.
func (c *Chat) Broadcast(to string, msg *entity.MessageWs) error {
	err := c.pubSub.PublishToChannel(to, msg)
	if err != nil {
		fmt.Println("c.rds.Client.Publish", err)
	}

	return nil
}

// Register registers new connection as a User.
func (c *Chat) Register(ctx context.Context, conn *websocket.Conn, username string) *User {
	user := &User{
		Chat: c,
		Conn: conn,
	}

	c.mu.Lock()
	{
		user.Id = c.seq
		user.Name = username

		c.seq++
	}
	c.mu.Unlock()

	// Client subscribe to redis channel (nama channel ya username si user sendiri)
	pubSub := c.pubSub.SubscribeToChannel(ctx, username)

	newChannelPubSub := &redispkg.ChannelPubSub{
		CloseChan:  make(chan struct{}, 1),
		ClosedChan: make(chan struct{}, 1),
		PubSub:     pubSub,
	}

	c.rds.ChannelsPubSubSync.Lock()

	if _, ok := c.rds.ChannelsPubSub[username]; !ok {
		c.rds.ChannelsPubSub[username] = newChannelPubSub
	}
	c.rds.ChannelsPubSubSync.Unlock()

	go user.subscribePubSubAndSendToClient(newChannelPubSub)

	return user
}

// subscribePubSubAndSendToClient subscribe channel (nama channel username si user) , setiap ada message kirim ke client ini/recipient
func (u *User) subscribePubSubAndSendToClient(channelRedis *redispkg.ChannelPubSub) {
	defer channelRedis.Closed()
	// Create a ticker that triggers a ping at given interval
	ticker := time.NewTicker(pingInterval)

	defer func() {
		ticker.Stop()
		u.Conn.Close()
	}()
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
				err = u.Write(websocket.TextMessage, msg)
				if err != nil {
					log.Println(err)
					return
				}
			}
		case <-ticker.C:

			log.Println("send ping messages to client " + u.Name + " !!")

			err := u.Conn.WriteMessage(websocket.PingMessage, []byte{})
			if err != nil {
				log.Println("wsutil.WriteServerMessage", err)
				return
			}
		}
	}
}

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
