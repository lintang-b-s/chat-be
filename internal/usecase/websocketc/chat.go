package websocketc

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"github.com/gorilla/websocket"

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
	io sync.Mutex
	//Conn   io.ReadWriteCloser
	//WsConn net.Conn
	Conn *websocket.Conn

	Id   uint
	Name string
	Chat *Chat
}

type MessageFromWs struct {
	UUID              string    `json:"uuid"`
	Type              string    `json:"type"`
	SenderUsername    string    `json:"sender_username"`
	RecipientUsername string    `json:"recipient_username"`
	GroupId           string    `json:"group_id"`
	Message           string    `json:"message"`
	CreatedAt         time.Time `json:"created_at"`
}

type Chat struct {
	mu        sync.RWMutex
	seq       uint
	rds       *redispkg.Redis
	edenAiApi webapi.EdenAIAPI
	pg        repo.UserRepo
}

func NewChat(rds *redispkg.Redis, ed webapi.EdenAIAPI, pg repo.UserRepo) *Chat {
	return &Chat{rds: rds, edenAiApi: ed, pg: pg}
}

var (
	// pongWait : berapa lama server menunggu message pong dari client
	pongWait = 10 * time.Second
	// pingInterval : setiap 0.9 detik server mengirim ping message ke client. pingInterval haruslah lebih kecil dari pongWait
	pingInterval = (pongWait * 9) / 10
)

// Receive reads next message from user's underlying connection.
// It blocks until full message received.
func (u *User) Receive() error {

	// Set Max Size of Messages in Bytes
	u.Conn.SetReadLimit(1024)

	// Configure Wait time for Pong response, use Current time + pongWait
	// This has to be done here to set the first initial timer.
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

	msgWs := &MessageFromWs{}
	if err = json.Unmarshal(msg, msgWs); err != nil {
		log.Println("json.Unmarshal: ", err)
		return err
	}

	if msgWs.Type == "chatBot_private" {
		// private chat dg chatbot
		resText := u.Chat.edenAiApi.GenerateText(msgWs.Message)
		msgWs.Message = resText
		err = u.Write(websocket.TextMessage, msgWs)
		if err != nil {
			log.Println("Write: ", err)
			return err
		}

	}

	if msgWs.Type == "private_chat" {
		// private chat dg user lain yang sudah ditambahkan kontaknya
		isFriendErr := u.Chat.pg.GetUserFriend(context.Background(), msgWs.SenderUsername, msgWs.RecipientUsername)
		if isFriendErr != nil {
			fmt.Println("u.Chat.pg.GetUserFriend: ", isFriendErr)
			msgWs.Message = msgWs.RecipientUsername + " is not your friend"
			err = u.Write(websocket.TextMessage, msgWs)
			if err != nil {
				log.Println("Write: ", err)
				return err
			}
			return isFriendErr

		}
		err = u.Chat.Broadcast(msgWs.RecipientUsername, msgWs)
		if err != nil {
			fmt.Println("u.chat.Broadcast: ", err)
			return err
		}
	}

	return nil
}

// pongHandler handle message pong yang dikirim oleh client, mereset durasi readdeadline
func (u *User) pongHandler(pongMsg string) error {
	log.Println("pong received from client !!")
	return u.Conn.SetReadDeadline(time.Now().Add(pongWait))
}

// Broadcast sends message to all alive users.
func (c *Chat) Broadcast(to string, msg *MessageFromWs) error {
	buff := bytes.NewBufferString("")
	enc := json.NewEncoder(buff)
	err := enc.Encode(msg)

	if err != nil {
		fmt.Println(err)
	}

	err = c.rds.Client.Publish(context.Background(), to, buff.String()).Err()
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
	pubSub := c.rds.Client.Subscribe(ctx, username)

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
			msg := &MessageFromWs{}
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

			log.Println("send ping messages to client!!")

			err := u.Conn.WriteMessage(websocket.PingMessage, []byte{})
			//fmt.Println("pingCode: ", ws.OpPing)
			//s, err := conn.Write(ws.CompiledPing)
			if err != nil {
				log.Println("wsutil.WriteServerMessage", err)
				return
			}
		}
	}
}

func (u *User) Write(op int, message *MessageFromWs) error {
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
