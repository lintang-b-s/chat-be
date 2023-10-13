package v1

import (
	"errors"
	"github.com/gin-gonic/gin"
	"github.com/lintangbs/chat-be/internal/entity"
	api "github.com/lintangbs/chat-be/internal/middleware"
	"github.com/lintangbs/chat-be/internal/usecase"
	"github.com/lintangbs/chat-be/internal/usecase/repo"
	"github.com/lintangbs/chat-be/internal/util/jwt"
	"github.com/lintangbs/chat-be/pkg/logger"
	"gorm.io/gorm"
	"net/http"
)

type contactRoutes struct {
	c   usecase.Contact
	l   logger.Interface
	jwt jwt.JwtTokenMaker
}

func NewContactRoutes(handler *gin.RouterGroup, c usecase.Contact, l logger.Interface, jwt jwt.JwtTokenMaker) {
	r := &contactRoutes{c, l, jwt}

	h := handler.Group("/contact").Use(api.AuthMiddleware(r.jwt))
	{
		h.POST("/add", r.addContact)
		h.GET("/", r.getContact)
	}
}

type addFriendRequest struct {
	FriendUsername string `json:"friend_username"`
}

// @Summary     Add Contact
// @Description    Add Contact
// @ID          addContact
// @Tags  	    user
// @Accept      json
// @Produce     json
// @Security ApiKeyAuth
// @Param       request body addFriendRequest true "set up addFriendRequest"
// @Success     200 {object} userResponse
// @Failure     400 {object} response
// @Failure     500 {object} response
// @Router      /v1/contact/add [post]
// Author: https://github.com/lintang-b-s
func (r *contactRoutes) addContact(c *gin.Context) {
	var request addFriendRequest

	if err := c.ShouldBindJSON(&request); err != nil {
		r.l.Error(err, "http - v1 - addContact")
		ErrorResponse(c, http.StatusBadRequest, "invalid request body")
		return
	}
	authPayload := c.MustGet(api.AuthorizationPayloadKey).(*jwt.Payload)

	user, err := r.c.AddContact(
		c.Request.Context(),
		entity.AddFriendRequest{
			MyUsername:     authPayload.Username,
			FriendUsername: request.FriendUsername,
		},
	)

	if err != nil {
		unwrapedErr := errors.Unwrap(err)
		if unwrapedErr == gorm.ErrRecordNotFound {
			ErrorResponse(c, http.StatusBadRequest, "User not found: "+unwrapedErr.Error())
			return
		}

		if unwrapedErr == repo.AlreadyYourInYourContactErr {
			ErrorResponse(c, http.StatusBadRequest, " Bad Request : "+unwrapedErr.Error())
			return
		}

		r.l.Error("http - v1- addContact")
		ErrorResponse(c, http.StatusInternalServerError, "addContact service problems: "+err.Error())
		return
	}

	res := userResponse{
		Username: user.Username,
		Id:       user.Id,
		Email:    user.Email,
	}

	c.JSON(http.StatusCreated, res)
}

type getContactResponse struct {
	Contacts []userResponse `json:"contacts"`
}

// @Summary     Get  User Contact
// @Description    Get User Contact
// @ID          getContact
// @Tags  	    user
// @Accept      json
// @Produce     json
// @Security ApiKeyAuth
// @Success     200 {object} getContactResponse
// @Failure     400 {object} response
// @Failure     500 {object} response
// @Router      /v1/contact/add [GET]
// Author: https://github.com/lintang-b-s
func (r *contactRoutes) getContact(c *gin.Context) {
	authPayload := c.MustGet(api.AuthorizationPayloadKey).(*jwt.Payload)

	contacts, err := r.c.GetContact(
		c.Request.Context(),
		entity.GetContactRequest{
			MyUsername: authPayload.Username,
		},
	)

	if err != nil {
		unwrapedErr := errors.Unwrap(err)
		if unwrapedErr == gorm.ErrRecordNotFound {
			ErrorResponse(c, http.StatusBadRequest, "User not found: "+unwrapedErr.Error())
			return
		}

		ErrorResponse(c, http.StatusInternalServerError, "getContact service problems: "+err.Error())
		return
	}

	var uFriends []userResponse
	for _, contact := range contacts.Friends {
		uFriend := userResponse{
			Id:       contact.Id,
			Username: contact.Username,
			Email:    contact.Email,
		}
		uFriends = append(uFriends, uFriend)
	}

	res := getContactResponse{
		Contacts: uFriends,
	}

	c.JSON(http.StatusOK, res)
}
