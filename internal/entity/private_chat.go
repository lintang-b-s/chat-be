package entity

import (
	"github.com/google/uuid"
	"time"
)

// PrivateChat messages
type PrivateChatMessage struct {
	MessageId   uint64    `json:"message_id"`
	MessageFrom uuid.UUID `json:"message_from"`
	MessageTo   uuid.UUID `json:"message_to"`
	Content     string    `json:"content"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
	DeletedAt   time.Time `json:"deleted_at"`
}

type PrivateChatUsers struct {
	Message map[uuid.UUID][]PrivateChatMessage `json:"message"`
}

type InsertPrivateChatRequest struct {
	MessageId   uint64    `json:"message_id"`
	MessageFrom uuid.UUID `json:"message_from"`
	MessageTo   uuid.UUID `json:"message_to"`
	Content     string    `json:"content"`
}

type GetPrivateChatQueryByUserRequest struct {
	UserId uuid.UUID `json:"user_id"`
}

type GetPrivateChatByUserRequest struct {
	Username string `json:"username"`
}
