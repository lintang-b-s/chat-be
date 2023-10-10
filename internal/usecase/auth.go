package usecase

import (
	"context"
	"fmt"
	"github.com/lintangbs/chat-be/internal/entity"
	"github.com/lintangbs/chat-be/internal/util"
	"github.com/lintangbs/chat-be/internal/util/jwt"
	"time"
)

type AuthUseCase struct {
	authRepo      AuthRepo
	jwtTokenMaker jwt.JwtTokenMaker
	sessionRepo   SessionRepo
}

func NewAuthUseCase(r AuthRepo, j jwt.JwtTokenMaker, s SessionRepo) *AuthUseCase {
	return &AuthUseCase{
		authRepo:      r,
		jwtTokenMaker: j,
		sessionRepo:   s,
	}
}

func (uc *AuthUseCase) Register(ctx context.Context, c entity.CreateUserRequest) (entity.User, error) {
	hashedPassword, err := util.HashPassword(c.Password)
	if err != nil {
		// internal server error
		return entity.User{}, fmt.Errorf("AuthUseCase - Register -  util.HashPassword: %w", err)
	}
	c.Password = hashedPassword
	createdUser, err := uc.authRepo.CreateUser(ctx, c)
	if err != nil {
		// internal server error/ bad request
		return entity.User{}, fmt.Errorf("AuthUseCase - Register - uc.authRepo.CreateUser: %w", err)
	}

	return createdUser, nil
}

// Login: logic login use case
func (uc *AuthUseCase) Login(ctx context.Context, l entity.LoginUserRequest) (entity.LoginUserResponse, error) {
	user, err := uc.authRepo.GetUser(ctx, l.Email)
	if err != nil {
		// Bad request User with email not found in DB
		return entity.LoginUserResponse{}, fmt.Errorf("AuthUseCase - Login - uc.authRepo.GetUser: %w", err)
	}

	err = util.CheckPassword(l.Password, user.HashedPassword)
	if err != nil {
		// unauthorized
		return entity.LoginUserResponse{}, fmt.Errorf("AuthUseCase - Login - util.CheckPassword: %w", err)
	}

	accessToken, accessPayload, err := uc.jwtTokenMaker.CreateToken(
		l.Email,
		3*time.Hour,
	)

	if err != nil {
		// internal server error
		return entity.LoginUserResponse{}, fmt.Errorf("AuthUseCase - Login - uc.jwtTokenMaker.CreateToken: %w", err)
	}

	refreshToken, refreshPayload, err := uc.jwtTokenMaker.CreateToken(
		l.Email,
		168*time.Hour,
	)
	if err != nil {
		// internal server error
		return entity.LoginUserResponse{}, fmt.Errorf("AuthUseCase - Login - uc.jwtTokenMaker.CreateToken: %w", err)
	}

	createSessionReq := entity.CreateSessionRequest{
		ID:           refreshPayload.ID,
		Email:        l.Email,
		RefreshToken: refreshToken,
		ExpiresAt:    refreshPayload.ExpiredAt,
	}

	session, err := uc.sessionRepo.CreateSession(
		ctx,
		createSessionReq,
	)

	if err != nil {
		return entity.LoginUserResponse{}, fmt.Errorf("AuthUseCase - Login - uc.sessionRepo.CreateSession: %w", err)
	}

	userRes := entity.User{
		Id:       user.Id,
		Username: user.Email,
		Email:    user.Email,
	}

	res := entity.LoginUserResponse{
		SessionId:             session.ID,
		AccessToken:           accessToken,
		AccessTokenExpiresAt:  accessPayload.ExpiredAt,
		RefreshToken:          refreshToken,
		RefreshTokenExpiresAt: refreshPayload.ExpiredAt,
		User:                  userRes,
	}
	return res, nil
}
