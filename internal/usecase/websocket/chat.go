package websocket

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"github.com/gobwas/ws"
	"github.com/gobwas/ws/wsutil"
	"github.com/lintangbs/chat-be/pkg/redispkg"
	"io"
	"log"
	"net"
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
	Conn io.ReadWriteCloser

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
	mu  sync.RWMutex
	seq uint
	rds *redispkg.Redis
}

func NewChat(rds *redispkg.Redis) *Chat {
	return &Chat{rds: rds}
}

// Receive reads next message from user's underlying connection.
// It blocks until full message received.
func (u *User) Receive() error {
	//req, err := u.readRequest()
	msg, _, err := wsutil.ReadClientData(u.Conn)

	if err != nil {
		u.Conn.Close()
		return err
	}

	if msg == nil {
		// Handled some control message.
		return nil
	}

	msgWs := &MessageFromWs{}
	if err = json.Unmarshal(msg, msgWs); err != nil {
		log.Println("json.Unmarshal")
	}

	if msgWs.Type == "private_chat" {
		err = u.Chat.Broadcast(msgWs.RecipientUsername, msgWs)
		if err != nil {
			fmt.Println("u.chat.Broadcast")
		}
	}

	return nil
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
func (c *Chat) Register(ctx context.Context, conn net.Conn, username string) *User {
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

	go c.subscribePubSubAndSendToClient(conn, newChannelPubSub)

	return user
}

// subscribePubSubAndSendToClient subscribe channel (nama channel username si user) , setiap ada message kirim ke client ini/recipient
func (c *Chat) subscribePubSubAndSendToClient(conn net.Conn, channelRedis *redispkg.ChannelPubSub) {
	defer channelRedis.Closed()
	for {
		select {
		case data := <-channelRedis.Channel():
			msg := &MessageFromWs{}
			dec := json.NewDecoder(strings.NewReader(data.Payload))

			err := dec.Decode(msg)
			if err != nil {
				log.Println("dec.Decode", err)
			} else {
				err := Write(conn, ws.OpText, msg)
				if err != nil {
					log.Println(err)
				}
			}

		}
	}
}

// Write to websocket client
func Write(conn io.ReadWriter, op ws.OpCode, message *MessageFromWs) error {

	data, err := json.Marshal(message)
	if err != nil {
		log.Println(err)
		return nil
	}
	log.Println("write socket message:", string(data))
	err = wsutil.WriteServerMessage(conn, op, data)
	if err != nil {
		log.Println(err)
		return err
	}

	return nil
}
