package repo

import (
	"fmt"
	"github.com/google/uuid"
	"github.com/lintangbs/chat-be/internal/entity"
	"gorm.io/gorm"
	"time"
)

type PrivateChatRepo struct {
	db *gorm.DB
}

type PrivateChat struct {
	gorm.Model
	Id          uint64
	MessageFrom uuid.UUID
	MessageTo   uuid.UUID
	Content     string `gorm:"type:text"`
	CreatedAt   time.Time
	UpdatedAt   time.Time
}

func NewPrivateChatRepo(db *gorm.DB) *PrivateChatRepo {
	return &PrivateChatRepo{db}
}

// InsertPrivateChat insert private chat to private chat table
func (r *PrivateChatRepo) InsertPrivateChat(e entity.InsertPrivateChatRequest) (entity.PrivateChatMessage, error) {

	msg := PrivateChat{Id: e.MessageId, MessageFrom: e.MessageFrom, MessageTo: e.MessageTo, Content: e.Content}
	if result := r.db.Create(&msg); result.Error != nil {
		return entity.PrivateChatMessage{}, fmt.Errorf("PrivateChatRepo -  InsertPrivateChat - r.db.Create", result.Error)
	}

	res := entity.PrivateChatMessage{
		MessageId:   e.MessageId,
		MessageFrom: e.MessageFrom,
		MessageTo:   e.MessageTo,
		Content:     e.Content,
		CreatedAt:   msg.CreatedAt,
		UpdatedAt:   msg.UpdatedAt,
	}

	return res, nil
}

// GetPrivateChatByUser get private chat by userId , bisa receiver & sender
func (r *PrivateChatRepo) GetPrivateChatByUser(e entity.GetPrivateChatQueryByUserRequest) (entity.PrivateChatUsers, error) {
	var msgs []PrivateChat
	//var satuMsg PrivateChat

	userId := e.UserId
	//
	if result := r.db.Where(&PrivateChat{MessageFrom: userId}).Or(&PrivateChat{MessageTo: userId}).Find(&msgs); result.Error != nil {
		return entity.PrivateChatUsers{}, fmt.Errorf("PrivateChatRepo - GetPrivateChatByUser -  r.db.Where: %w", result.Error)
	}

	pc := entity.PrivateChatUsers{
		Message: make(map[uuid.UUID][]entity.PrivateChatMessage),
	}

	for _, msg := range msgs {
		newPcMessage := entity.PrivateChatMessage{
			MessageId:   msg.Id,
			MessageFrom: msg.MessageFrom,
			MessageTo:   msg.MessageTo,
			Content:     msg.Content,
			CreatedAt:   msg.CreatedAt,
			UpdatedAt:   msg.UpdatedAt,
			DeletedAt:   msg.UpdatedAt,
		}
		if msg.MessageFrom == userId {
			// jika from = user yg login
			if pcMsg, prs := pc.Message[msg.MessageTo]; !prs {
				var toMessages []entity.PrivateChatMessage

				pc.Message[msg.MessageTo] = append(toMessages, newPcMessage)
			} else {
				pc.Message[msg.MessageTo] = append(pcMsg, newPcMessage)
			}
		} else if msg.MessageTo == userId {
			// jika sender = user yg login
			if pcMsg, prs := pc.Message[msg.MessageFrom]; !prs {
				var toMessages []entity.PrivateChatMessage

				pc.Message[msg.MessageFrom] = append(toMessages, newPcMessage)
			} else {
				pc.Message[msg.MessageFrom] = append(pcMsg, newPcMessage)
			}

		}
	}

	return pc, nil

}
