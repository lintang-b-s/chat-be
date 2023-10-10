package v1

import (
	"errors"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/lintangbs/chat-be/internal/entity"
	"github.com/lintangbs/chat-be/internal/usecase"
	"github.com/lintangbs/chat-be/pkg/logger"
	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm"
	"net/http"
	"time"
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

type createUserRequest struct {
	Username string `json:"username" binding:"required,alphanum"`
	Password string `json:"password" binding:"required,min=3"`
	Email    string `json:"email" binding:"required, email"`
}

type userResponse struct {
	Id       uint   `json:"id"`
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
	var request createUserRequest

	if err := c.ShouldBindJSON(&request); err != nil {
		r.l.Error(err, "http - v1 - registerUser")
		ErrorResponse(c, http.StatusBadRequest, "invalid request body")
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

		errorM := errors.Unwrap(err)
		if errorM.Error() == "AuthRepo - Createuser - r.db.Where: User With same username or email already exists" {
			ErrorResponse(c, http.StatusBadRequest, "Bad request:  User With same username or email already exists")
			return
		}

		r.l.Error("http - v1- registerUser")
		ErrorResponse(c, http.StatusInternalServerError, "auth service problems: "+errorM.Error())
		return
	}

	resp := userResponse{Username: user.Username, Email: user.Email}
	c.JSON(http.StatusCreated, resp)
}

// ---- loginUser ----
type loginUserRequest struct {
	Email    string `json:"email" binding:"required,alphanum"`
	Password string `json:"password" binding:"required,min=3"`
}

type loginUserResponse struct {
	SessionId             uuid.UUID    `json:"session_id"`
	AccessToken           string       `json:"access_token"`
	AccessTokenExpiresAt  time.Time    `json:"access_token_expires_at"`
	RefreshToken          string       `json:"refresh_token"`
	RefreshTokenExpiresAt time.Time    `json:"refresh_token_expires_at"`
	User                  userResponse `json:"user"`
}

// @Summary     Login User
// @Description   Login User
// @ID          loginUser
// @Tags  	    user
// @Accept      json
// @Produce     json
// @Param       request body LoginUserRequest true "Login  user"
// @Success     200 {object} LoginUserResponse
// @Failure     400 {object} response
// @Failure     500 {object} response
// @Router      /subscription [post]
// Author: https://github.com/lintang-b-s
func (r *authRoutes) loginUser(c *gin.Context) {

	var request loginUserRequest

	if err := c.ShouldBindJSON(&request); err != nil {
		r.l.Error(err, "http - v1 - loginUser")
		ErrorResponse(c, http.StatusInternalServerError, "invalid request body")
		return
	}

	loginResponse, err := r.a.Login(
		c.Request.Context(),
		entity.LoginUserRequest{
			Email:    request.Email,
			Password: request.Password,
		},
	)
	if err != nil {

		unwrapedErr := errors.Unwrap(err)
		if unwrapedErr == bcrypt.ErrMismatchedHashAndPassword {
			ErrorResponse(c, http.StatusUnauthorized, "Wrong password")
			return
		}

		if unwrapedErr == gorm.ErrRecordNotFound {
			ErrorResponse(c, http.StatusBadRequest, "User not found: "+unwrapedErr.Error())
			return
		}

		r.l.Error("http - v1- loginUser")
		ErrorResponse(c, http.StatusInternalServerError, "translation service problems: "+err.Error())
		return
	}

	resp := loginUserResponse{
		SessionId:             loginResponse.SessionId,
		AccessToken:           loginResponse.AccessToken,
		AccessTokenExpiresAt:  loginResponse.AccessTokenExpiresAt,
		RefreshToken:          loginResponse.RefreshToken,
		RefreshTokenExpiresAt: loginResponse.RefreshTokenExpiresAt,
		User: userResponse{
			Id:       loginResponse.User.Id,
			Username: loginResponse.User.Email,
			Email:    loginResponse.User.Email,
		},
	}
	c.JSON(http.StatusCreated, resp)
}
