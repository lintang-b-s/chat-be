package entity

import "github.com/google/uuid"

var (
	ServerName = "chat-server" + uuid.New().String()
)
