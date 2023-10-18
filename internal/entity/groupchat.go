package entity

import (
	"github.com/google/uuid"
	"time"
)

// GroupChatMessage entitas pesan group chat
type GroupChatMessage struct {
	GroupId   uuid.UUID `json:"id"`
	MessageId uint64    `json:"message_id"`
	UserId    uuid.UUID `json:"user_id"`
	Content   string    `json:"content"`
	CreatedAt time.Time `json:"created_at,omitempty"`
	UpdatedAt time.Time `json:"updated_at,omitempty"`
}

// GroupChatMessages array of pesan group chat
type GroupChatMessages struct {
	Messages []GroupChatMessage `json:"messages"`
}

type GroupChatMsgRequest struct {
	GroupName string `json:"group_name"`
	UserName  string `json:"user_name"`
}
