package usecase

import (
	"context"
	"fmt"
	"github.com/lintangbs/chat-be/internal/entity"
	"github.com/lintangbs/chat-be/internal/usecase/redisRepo"
	"github.com/lintangbs/chat-be/internal/usecase/repo"
	"github.com/lintangbs/chat-be/internal/util"
	"github.com/lintangbs/chat-be/internal/util/jwt"
	"time"
)

type AuthUseCase struct {
	userRepo      repo.UserRepo
	jwtTokenMaker jwt.JwtTokenMaker
	sessionRepo   SessionRepo
	otpRepo       redisRepo.OtpRepo
}

func NewAuthUseCase(r repo.UserRepo, j jwt.JwtTokenMaker, s SessionRepo, otpRepo redisRepo.OtpRepo) *AuthUseCase {
	return &AuthUseCase{
		userRepo:      r,
		jwtTokenMaker: j,
		sessionRepo:   s,
		otpRepo:       otpRepo,
	}
}

func (uc *AuthUseCase) Register(ctx context.Context, c entity.CreateUserRequest) (entity.UserResponse, error) {
	hashedPassword, err := util.HashPassword(c.Password)
	if err != nil {
		// internal server error
		return entity.UserResponse{}, fmt.Errorf("AuthUseCase - Register -  util.HashPassword: %w", err)
	}
	c.Password = hashedPassword
	createdUser, err := uc.userRepo.CreateUser(ctx, c)
	if err != nil {
		// internal server error/ bad request
		return entity.UserResponse{}, fmt.Errorf("AuthUseCase - Register - uc.userRepo.CreateUser: %w", err)
	}

	return createdUser, nil
}

// Login: logic login use case
func (uc *AuthUseCase) Login(ctx context.Context, l entity.LoginUserRequest) (entity.LoginUserResponse, error) {
	user, err := uc.userRepo.GetUser(ctx, l.Email)
	if err != nil {
		// Bad request User with email not found in DB
		return entity.LoginUserResponse{}, fmt.Errorf("AuthUseCase - Login - uc.userRepo.GetUser: %w", err)
	}

	err = util.CheckPassword(l.Password, user.HashedPassword)
	if err != nil {
		// unauthorized
		return entity.LoginUserResponse{}, fmt.Errorf("AuthUseCase - Login - util.CheckPassword: %w", err)
	}

	accessToken, accessPayload, err := uc.jwtTokenMaker.CreateToken(
		user.Username,
		7*time.Hour,
	)

	if err != nil {
		// internal server error
		return entity.LoginUserResponse{}, fmt.Errorf("AuthUseCase - Login - uc.jwtTokenMaker.CreateToken: %w", err)
	}

	refreshToken, refreshPayload, err := uc.jwtTokenMaker.CreateToken(
		user.Username,
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

	//Create Otp for Websocket
	otp, err := uc.otpRepo.CreateOtp(ctx, user.Username)
	if err != nil {
		return entity.LoginUserResponse{}, fmt.Errorf("AuthUseCase - Login - uc.otpRepo.CreateOtp: %w", err)
	}

	userRes := entity.UserResponse{
		Id:       user.Id,
		Username: user.Username,
		Email:    user.Email,
	}

	res := entity.LoginUserResponse{
		SessionId:             session.ID,
		AccessToken:           accessToken,
		AccessTokenExpiresAt:  accessPayload.ExpiredAt,
		RefreshToken:          refreshToken,
		RefreshTokenExpiresAt: refreshPayload.ExpiredAt,
		User:                  userRes,
		Otp:                   otp,
	}
	return res, nil
}

func (uc *AuthUseCase) RenewAccessToken(ctx context.Context, r entity.RenewAccessTokenRequest) (entity.RenewAccessTokenResponse, error) {
	refreshPayload, err := uc.jwtTokenMaker.VerifyToken(r.RefreshToken)
	if err != nil {
		// Unauthorized , token yg dikrim tidak sama dg yg ada di database
		return entity.RenewAccessTokenResponse{}, fmt.Errorf("AuthUseCase - RenewAccessToken - uc.jwtTokenMaker.VerifyToken: %w", err)
	}

	session, err := uc.sessionRepo.GetSession(ctx, refreshPayload.ID)
	if err != nil {
		return entity.RenewAccessTokenResponse{}, fmt.Errorf("AuthUseCase - RenewAccessToken - uc.sessionRepo.GetSession: %w", err)
	}

	if session.Email != refreshPayload.Username {
		return entity.RenewAccessTokenResponse{}, fmt.Errorf("Invalid session")
	}
	if session.RefreshToken != r.RefreshToken {
		return entity.RenewAccessTokenResponse{}, fmt.Errorf("Invalid session")
	}

	if time.Now().After(session.ExpiresAt) {
		return entity.RenewAccessTokenResponse{}, fmt.Errorf("Invalid session")
	}

	accessToken, accessTokenPayload, err := uc.jwtTokenMaker.CreateToken(
		refreshPayload.Username,
		7*time.Hour,
	)
	if err != nil {
		return entity.RenewAccessTokenResponse{}, fmt.Errorf("AuthUseCase - RenewAccessToken - uc.jwtTokenMaker.CreateToken: %w", err)
	}

	res := entity.RenewAccessTokenResponse{
		AccessToken:          accessToken,
		AccessTokenExpiresAt: accessTokenPayload.ExpiredAt,
	}
	return res, nil
}
