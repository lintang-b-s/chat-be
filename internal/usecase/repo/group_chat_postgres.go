package repo

import (
	"fmt"
	"github.com/google/uuid"
	"github.com/lintangbs/chat-be/internal/entity"
	"gorm.io/gorm"
	"time"
)

type GroupChatRepo struct {
	db *gorm.DB
}

type GroupChat struct {
	gorm.Model
	Id        uuid.UUID
	MessageId uint64
	UserId    uuid.UUID
	Content   string
	CreatedAt time.Time
	UpdatedAt time.Time
}

func NewGroupChatRepo(db *gorm.DB) *GroupChatRepo {
	return &GroupChatRepo{db}
}

// InsertNewChat insert group Chat to Database postgres
func (r *GroupChatRepo) InsertNewChat(gcMessage entity.GroupChatMessage) (entity.GroupChatMessage, error) {
	msg := GroupChat{Id: gcMessage.GroupId,
		MessageId: gcMessage.MessageId,
		UserId:    gcMessage.UserId,
		Content:   gcMessage.Content,
	}

	if result := r.db.Create(&msg); result.Error != nil {
		return entity.GroupChatMessage{}, fmt.Errorf("GroupChatRepo - InsertNewChat - r.db.Create(&msg): %w", result.Error)
	}

	msgRes := entity.GroupChatMessage{
		GroupId:   msg.Id,
		MessageId: msg.MessageId,
		UserId:    msg.UserId,
		Content:   msg.Content,
		CreatedAt: msg.CreatedAt,
		UpdatedAt: msg.UpdatedAt,
	}

	return msgRes, nil
}

// GetMessagesByGroupId getMessages by groupId
func (r *GroupChatRepo) GetMessagesByGroupId(groupId uuid.UUID) (entity.GroupChatMessages, error) {
	var groupChat []GroupChat
	if res := r.db.Where(&GroupChat{Id: groupId}).Find(&groupChat); res.Error != nil {
		return entity.GroupChatMessages{}, fmt.Errorf("GroupChatRepo - GetMessagesByGroupId -  r.db.Where(&Group{Id: groupId}).Find: %w", res.Error)
	}

	var msgs []entity.GroupChatMessage
	for _, gChat := range groupChat {
		msgs = append(msgs, entity.GroupChatMessage{
			GroupId:   gChat.Id,
			MessageId: gChat.MessageId,
			UserId:    gChat.UserId,
			Content:   gChat.Content,
			CreatedAt: gChat.CreatedAt,
			UpdatedAt: gChat.UpdatedAt,
		})
	}

	res := entity.GroupChatMessages{
		Messages: msgs,
	}

	return res, nil
}
