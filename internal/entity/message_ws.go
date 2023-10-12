package entity

import "time"

type DataPubSubMessage struct {
	UUID          string    `json:"UUID"`
	SenderUUID    string    `json:"SenderUUID"`
	RecipientUUID string    `json:"RecipientUUID"`
	Message       string    `json:"Message"`
	CreatedAt     time.Time `json:"CreatedAt"`
}
