package entity

import "time"

type DataPubSubMessage struct {
	UUID          string    `json:"UUID"`
	SenderUUID    string    `json:"SenderUUID"`
	RecipientUUID string    `json:"RecipientUUID"`
	Message       string    `json:"Message"`
	CreatedAt     time.Time `json:"CreatedAt"`
}

// MessageWs message yang dipertukarkan di websocket connection
type MessageWs struct {
	Type                  MessageType               `json:"type"`
	PrivateChat           MessagePrivateChat        `json:"private_chat,omitempty"`
	MsgOnlineStatusFanout MessageOnlineStatusFanout `json:"msg_online_status_fanout,omitempty"`
}

// MessagePrivateChat message untuk private chat
type MessagePrivateChat struct {
	MessageId         string `json:"message_id,omitempty"`
	SenderUsername    string `json:"sender_username"`
	RecipientUsername string `json:"recipient_username"`
	//GroupId           string      `json:"group_id"`
	Message   string    `json:"message"`
	CreatedAt time.Time `json:"created_at,omitempty"`
}

type MessageOnlineStatusFanout struct {
	FriendId       string `json:"friend_id"`
	FriendUsername string `json:"friend_username"`
	FriendEmail    string `json:"friend_email"`
	Online         bool   `json:"online"`
}

type (
	MessageType string
)

const (
	MessageTypeLogin              MessageType = "login"
	MessageTypeLogout             MessageType = "logout"
	MessageTypePrivateChat        MessageType = "private_chat"
	MessageTypePrivateChatBot     MessageType = "chatBot_private"
	MessageTypeOnlineStatusFanOut MessageType = "online_status_fanout"
	MessageTypeGroupChat          MessageType = "group_chat"
	MessageTypePrivateChatJoin    MessageType = "private_chat_join"
	MessageTypeGroupChatJoin      MessageType = "group_chat_join"
)
