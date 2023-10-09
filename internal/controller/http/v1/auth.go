package v1

import (
	"github.com/gin-gonic/gin"
	"github.com/lintangbs/chat-be/internal/entity"
	"github.com/lintangbs/chat-be/internal/usecase"
	"github.com/lintangbs/chat-be/pkg/logger"
	"net/http"
)

type authRoutes struct {
	a usecase.Auth
	l logger.Interface
}

func newAuthRoutes(handler *gin.RouterGroup, a usecase.Auth, l logger.Interface) {
	r := &authRoutes{a, l}

	h := handler.Group("/auth")
	{
		h.POST("/register", r.registerUser)
	}

}

type CreateUserRequest struct {
	Username string `json:"username"`
	Password string `json:"password"`
	Email    string `json:"email"`
}

type CreateUserResponse struct {
	Id       string `json:"id"`
	Username string `json:"username"`
	Email    string `json:"email"`
}

// @Summary     Register User in Db
// @Description   Register User in Db
// @ID          registerUser
// @Tags  	    user
// @Accept      json
// @Produce     json
// @Param       request body CreateUserRequest true "Set up user"
// @Success     200 {object} CreateUserResponse
// @Failure     400 {object} response
// @Failure     500 {object} response
// @Router      /subscription [post]
// Author: https://github.com/lintang-b-s
func (r *authRoutes) registerUser(c *gin.Context) {
	var request CreateUserRequest

	if err := c.ShouldBindJSON(&request); err != nil {
		r.l.Error(err, "http - v1 - registerUser")
		errorResponse(c, http.StatusBadRequest, "invalid request body")
		return
	}

	user, err := r.a.Register(
		c.Request.Context(),
		entity.CreateUserRequest{
			Username: request.Username,
			Password: request.Password,
			Email:    request.Email,
		},
	)

	if err != nil {
		r.l.Error("http - v1- registerUser")
		errorResponse(c, http.StatusBadRequest, "Bad request: %w"+err.Error())
		return
	}

	resp := CreateUserResponse{Username: user.Username, Email: user.Email}
	c.JSON(http.StatusCreated, resp)
}
