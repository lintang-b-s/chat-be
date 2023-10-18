package v1

import (
	"errors"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/lintangbs/chat-be/internal/entity"
	api "github.com/lintangbs/chat-be/internal/middleware"
	"github.com/lintangbs/chat-be/internal/usecase"
	"github.com/lintangbs/chat-be/internal/usecase/repo"
	"github.com/lintangbs/chat-be/internal/util/jwt"
	"github.com/lintangbs/chat-be/pkg/logger"
	"gorm.io/gorm"
	"net/http"
	"time"
)

type messageRoutes struct {
	m   usecase.Message
	l   logger.Interface
	jwt jwt.JwtTokenMaker
}

func NewMessageRoutes(handler *gin.RouterGroup, m usecase.Message, l logger.Interface, jwt jwt.JwtTokenMaker) {
	r := &messageRoutes{m, l, jwt}

	h := handler.Group("/messages").Use(api.AuthMiddleware(r.jwt))
	{
		h.GET("", r.getMessages)
		h.GET("/friend", r.getMessagesByFriend)
		h.GET("/group", r.getMessagesByGroupChat)
	}

}

// PrivateChat messages
type privateChatMessage struct {
	MessageId   uint64    `json:"message_id"`
	MessageFrom uuid.UUID `json:"message_from"`
	MessageTo   uuid.UUID `json:"message_to"`
	Content     string    `json:"content"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
	DeletedAt   time.Time `json:"deleted_at"`
}

type privateChatUsersResponse struct {
	Message map[string]map[uuid.UUID][]privateChatMessage `json:"message"`
}

// @Summary     Get user messages
// @Description    Get user messages
// @ID          getMessages
// @Tags  	    messages
// @Accept      json
// @Produce     json
// @Security OAuth2Application
// @Success     200 {object} privateChatUsersResponse
// @Failure     400 {object} response
// @Failure     500 {object} response
// @Router      /v1/messages [get]
// Author: https://github.com/lintang-b-s
func (r *messageRoutes) getMessages(c *gin.Context) {
	authPayload := c.MustGet(api.AuthorizationPayloadKey).(*jwt.Payload)

	msgs, err := r.m.GetMessageByUserLogin(
		c.Request.Context(),
		entity.GetPrivateChatByUserRequest{
			Username: authPayload.Username,
		},
	)
	if err != nil {
		unwrapedErr := errors.Unwrap(err)
		errRepo := errors.Unwrap(unwrapedErr)
		if errRepo == gorm.ErrRecordNotFound {
			ErrorResponse(c, http.StatusBadRequest, errRepo.Error())
			return
		}
		r.l.Error("http - v1- getMessages")
		ErrorResponse(c, http.StatusInternalServerError, "getMessages service problems: "+err.Error())
		return
	}

	res := privateChatUsersResponse{
		Message: make(map[string]map[uuid.UUID][]privateChatMessage),
	}
	for key, val := range msgs.Message {
		var pcMsgs []privateChatMessage
		for _, msgVal := range val {
			pcMsgs = append(pcMsgs, privateChatMessage{
				MessageId:   msgVal.MessageId,
				MessageFrom: msgVal.MessageFrom,
				MessageTo:   msgVal.MessageTo,
				Content:     msgVal.Content,
				CreatedAt:   msgVal.CreatedAt,
				UpdatedAt:   msgVal.UpdatedAt,
				DeletedAt:   msgVal.DeletedAt,
			})
		}
		innerMap := make(map[uuid.UUID][]privateChatMessage)
		innerMap[key] = pcMsgs
		res.Message["friendId"] = innerMap
	}

	c.JSON(http.StatusOK, res)
}

type getMessagesByFriendResponse struct {
	Messages []privateChatMessage `json:"message"`
}

// @Summary     Get user messages by friend
// @Description    Get user messages by friend
// @ID          getMessagesByFriend
// @Tags  	    messages
// @Accept      json
// @Produce     json
// @Security OAuth2Application
// @Param        friendUsername    query     string  false  "friendName search by friendUsername"
// @Success     200 {object} getMessagesByFriendResponse
// @Failure     400 {object} response
// @Failure     500 {object} response
// @Router      /v1/messages/friend [get]
// Author: https://github.com/lintang-b-s
func (r *messageRoutes) getMessagesByFriend(c *gin.Context) {
	friendUsername := c.Query("friendUsername")
	authPayload := c.MustGet(api.AuthorizationPayloadKey).(*jwt.Payload)

	if friendUsername == "" {
		ErrorResponse(c, http.StatusBadRequest, "invalid request body")
		return
	}

	msgs, err := r.m.GetMessagesByRecipient(
		c.Request.Context(),
		entity.GetPCBySdrAndRcvrRequest{
			SenderUsername:   authPayload.Username,
			ReceiverUsername: friendUsername,
		},
	)
	if err != nil {
		unwrapedErr := errors.Unwrap(err)
		errRepo := errors.Unwrap(unwrapedErr)
		if errRepo == gorm.ErrRecordNotFound {
			ErrorResponse(c, http.StatusBadRequest, errRepo.Error())
			return
		}
		r.l.Error("http - v1- getMessages")
		ErrorResponse(c, http.StatusInternalServerError, "getMessages service problems: "+err.Error())
		return
	}

	var pcs []privateChatMessage

	for _, msg := range msgs.Messages {
		pcs = append(pcs, privateChatMessage{
			MessageId:   msg.MessageId,
			MessageFrom: msg.MessageFrom,
			MessageTo:   msg.MessageTo,
			Content:     msg.Content,
			CreatedAt:   msg.CreatedAt,
			UpdatedAt:   msg.UpdatedAt,
			DeletedAt:   msg.DeletedAt,
		})
	}

	res := getMessagesByFriendResponse{
		Messages: pcs,
	}

	c.JSON(http.StatusOK, res)
}

type groupChatMessage struct {
	GroupId   uuid.UUID `json:"id"`
	MessageId uint64    `json:"message_id"`
	UserId    uuid.UUID `json:"user_id"`
	Content   string    `json:"content"`
	CreatedAt time.Time `json:"created_at,omitempty"`
	UpdatedAt time.Time `json:"updated_at,omitempty"`
}

type getMessagesByGroupName struct {
	Messages []groupChatMessage `json:"messages"`
}

// @Summary     Get user messages by group Chat
// @Description    Get user messages by group Chat
// @ID          getMessagesByGroupChat
// @Tags  	    messages
// @Accept      json
// @Produce     json
// @Security OAuth2Application
// @Param        groupName    query     string  false  "groupName search by group"
// @Success     200 {object} getMessagesByGroupName
// @Failure     400 {object} response
// @Failure     500 {object} response
// @Router      /v1/messages/group [get]
// Author: https://github.com/lintang-b-s
func (r *messageRoutes) getMessagesByGroupChat(c *gin.Context) {
	groupName := c.Query("groupName")
	authPayload := c.MustGet(api.AuthorizationPayloadKey).(*jwt.Payload)

	if groupName == "" {
		ErrorResponse(c, http.StatusBadRequest, "invalid request body")
		return
	}

	msgs, err := r.m.GetMessagesByGroupChat(
		c.Request.Context(),
		entity.GroupChatMsgRequest{
			GroupName: groupName,
			UserName:  authPayload.Username,
		},
	)

	if err != nil {
		unwrapedErr := errors.Unwrap(err)
		errRepo := errors.Unwrap(unwrapedErr)
		if errRepo == gorm.ErrRecordNotFound || errRepo == repo.UserNotMemberErr {
			ErrorResponse(c, http.StatusBadRequest, errRepo.Error())
			return
		}

		r.l.Error("http - v1- getMessages")
		ErrorResponse(c, http.StatusInternalServerError, "getMessages service problems: "+err.Error())
		return
	}

	var gcMsgs []groupChatMessage
	for _, msg := range msgs.Messages {
		gcMsgs = append(gcMsgs, groupChatMessage{
			GroupId:   msg.GroupId,
			MessageId: msg.MessageId,
			UserId:    msg.UserId,
			Content:   msg.Content,
			CreatedAt: msg.CreatedAt,
			UpdatedAt: msg.UpdatedAt,
		})
	}

	res := getMessagesByGroupName{
		Messages: gcMsgs,
	}
	c.JSON(http.StatusOK, res)
}
