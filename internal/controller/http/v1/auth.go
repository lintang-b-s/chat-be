package v1

import (
	"errors"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/lintangbs/chat-be/internal/entity"
	"github.com/lintangbs/chat-be/internal/usecase"
	"github.com/lintangbs/chat-be/internal/util/jwt"
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
		h.POST("/login", r.loginUser)
		h.POST("/token", r.renewAccessToken)
	}

}

type createUserRequest struct {
	Username string `json:"username" binding:"required,alphanum"`
	Password string `json:"password" binding:"required,min=3"`
	Email    string `json:"email" binding:"required,email"`
}

type userResponse struct {
	Id       uuid.UUID `json:"id"`
	Username string    `json:"username"`
	Email    string    `json:"email"`
}

// @Summary     Register User in Db
// @Description   Register User in Db
// @ID          registerUser
// @Tags  	    user
// @Accept      json
// @Produce     json
// @Param       request body createUserRequest true "Set up user"
// @Success     200 {object} userResponse
// @Failure     400 {object} response
// @Failure     500 {object} response
// @Router      /v1/auth/register [post]
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
		ErrorResponse(c, http.StatusInternalServerError, "register service problems: "+errorM.Error())
		return
	}

	resp := userResponse{Id: user.Id, Username: user.Username, Email: user.Email}
	c.JSON(http.StatusCreated, resp)
}

// ---- loginUser ----
type loginUserRequest struct {
	Email    string `json:"email" binding:"required"`
	Password string `json:"password" binding:"required,min=3"`
}

type loginUserResponse struct {
	SessionId             uuid.UUID    `json:"session_id"`
	AccessToken           string       `json:"access_token"`
	AccessTokenExpiresAt  time.Time    `json:"access_token_expires_at"`
	RefreshToken          string       `json:"refresh_token"`
	RefreshTokenExpiresAt time.Time    `json:"refresh_token_expires_at"`
	User                  userResponse `json:"user"`
	Otp                   string       `json:"otp"`
}

// @Summary     Login User
// @Description   Login User
// @ID          loginUser
// @Tags  	    user
// @Accept      json
// @Produce     json
// @Param       request body loginUserRequest true "Login  user"
// @Success     200 {object} loginUserResponse
// @Failure     400 {object} response
// @Failure     500 {object} response
// @Router      /v1/auth/login [post]
// Author: https://github.com/lintang-b-s
func (r *authRoutes) loginUser(c *gin.Context) {

	var request loginUserRequest

	if err := c.ShouldBindJSON(&request); err != nil {
		r.l.Error(err, "http - v1 - loginUser")
		ErrorResponse(c, http.StatusBadRequest, "invalid request body")
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
		ErrorResponse(c, http.StatusInternalServerError, "loginUser service problems: "+err.Error())
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
			Username: loginResponse.User.Username,
			Email:    loginResponse.User.Email,
		},
		Otp: loginResponse.Otp,
	}
	c.JSON(http.StatusCreated, resp)
}

type renewAccessTokenRequest struct {
	RefreshToken string `json:"refresh_token" binding:"required"`
}

type renewAccessTokenResponse struct {
	AccessToken          string    `json:"access_token"`
	AccessTokenExpiresAt time.Time `json:"access_token_expires_at"`
}

// @Summary     renew Access Token using user refreshToken
// @Description    renew Access Token using user refreshToken
// @ID          renewAccessToken
// @Tags  	    user
// @Accept      json
// @Produce     json
// @Param       request body renewAccessTokenRequest true "Login  user"
// @Success     200 {object} renewAccessTokenResponse
// @Failure     400 {object} response
// @Failure     500 {object} response
// @Router      /v1/auth/token [post]
// Author: https://github.com/lintang-b-s
func (r *authRoutes) renewAccessToken(c *gin.Context) {
	var request renewAccessTokenRequest

	if err := c.ShouldBindJSON(&request); err != nil {

		unwrapedErr := errors.Unwrap(err)
		// jika refresh token invalid / expired
		if unwrapedErr == jwt.ErrInvalidToken || unwrapedErr == jwt.ErrExpiredToken {
			ErrorResponse(c, http.StatusUnauthorized, "Token invalid or token already expired")
			return
		}
		// jika row pada db tidak ditemukan
		if unwrapedErr == gorm.ErrRecordNotFound {
			ErrorResponse(c, http.StatusBadRequest, "User not found: "+unwrapedErr.Error())
			return
		}

		if err.Error() == "Invalid session" {
			ErrorResponse(c, http.StatusUnauthorized, "Refresh Token mismatch with refresh token in database")
			return
		}

		r.l.Error(err, "http - v1 - loginUser")
		ErrorResponse(c, http.StatusBadRequest, "invalid request body")
		return
	}

	renewResponse, err := r.a.RenewAccessToken(
		c.Request.Context(),
		entity.RenewAccessTokenRequest{
			RefreshToken: request.RefreshToken,
		},
	)
	if err != nil {
		r.l.Error("http - v1- renewAccessToken")
		ErrorResponse(c, http.StatusInternalServerError, "renewAccessToken service problems: "+err.Error())
		return
	}

	resp := renewAccessTokenResponse{
		AccessToken:          renewResponse.AccessToken,
		AccessTokenExpiresAt: renewResponse.AccessTokenExpiresAt,
	}

	c.JSON(http.StatusCreated, resp)
}
