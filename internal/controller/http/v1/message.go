package v1

import (
	"errors"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/lintangbs/chat-be/internal/entity"
	api "github.com/lintangbs/chat-be/internal/middleware"
	"github.com/lintangbs/chat-be/internal/usecase"
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
	Message map[uuid.UUID][]privateChatMessage `json:"message"`
}

// @Summary     Get user messages
// @Description    Get user messages
// @ID          getMessages
// @Tags  	    user
// @Accept      json
// @Produce     json
// @Security ApiKeyAuth
// @Success     200 {object} getMessageResponse
// @Failure     400 {object} response
// @Failure     500 {object} response
// @Router      /v1/contact/add [post]
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
		if unwrapedErr == gorm.ErrRecordNotFound {
			ErrorResponse(c, http.StatusBadRequest, unwrapedErr.Error())
			return
		}
		r.l.Error("http - v1- getMessages")
		ErrorResponse(c, http.StatusInternalServerError, "getMessages service problems: "+err.Error())
		return
	}

	res := privateChatUsersResponse{
		Message: make(map[uuid.UUID][]privateChatMessage),
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
		res.Message[key] = pcMsgs
	}

	c.JSON(http.StatusOK, res)

}
