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
	userRepo      UserRepo
	jwtTokenMaker jwt.JwtTokenMaker
	sessionRepo   SessionRepo
	otpRepo       OtpRepo
	pubSubRds     PubSubRedis
	userRdsRepo   UserRedisRepo
}

func NewAuthUseCase(r UserRepo, j jwt.JwtTokenMaker, s SessionRepo,
	otpRepo OtpRepo,
	redis PubSubRedis, userRdsRepo UserRedisRepo) *AuthUseCase {
	return &AuthUseCase{
		userRepo:      r,
		jwtTokenMaker: j,
		sessionRepo:   s,
		otpRepo:       otpRepo,
		pubSubRds:     redis,
		userRdsRepo:   userRdsRepo,
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
		Username:     user.Username,
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

	if session.Username != refreshPayload.Username {
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

func (uc *AuthUseCase) DeleteRefreshToken(ctx context.Context, d entity.DeleteRefreshTokenRequest) error {

	refreshPayload, err := uc.jwtTokenMaker.VerifyToken(d.RefreshToken)
	if err != nil {
		// Unauthorized , token yg dikrim tidak sama dg yg ada di database
		return fmt.Errorf("AuthUseCase - DeleteRefreshToken - uc.jwtTokenMaker.VerifyToken: %w", err)
	}

	session, err := uc.sessionRepo.GetSession(ctx, refreshPayload.ID)
	if err != nil {
		return fmt.Errorf("AuthUseCase - DeleteRefreshToken - uc.sessionRepo.GetSession: %w", err)
	}
	if session.Username != refreshPayload.Username {
		return fmt.Errorf("Invalid session")
	}
	if session.RefreshToken != d.RefreshToken {
		return fmt.Errorf("Invalid session")
	}

	if time.Now().After(session.ExpiresAt) {
		return fmt.Errorf("Invalid session")
	}
	err = uc.sessionRepo.DeleteSession(ctx, refreshPayload.ID)
	if err != nil {
		return fmt.Errorf("AuthUseCase - DeleteRefreshToken - uc.sessionRepo.DeleteSession: %w", err)
	}

	user, err := uc.userRepo.GetUserFriends(ctx, refreshPayload.Username)
	if err != nil {
		return fmt.Errorf("AuthUseCase - DeleteRefreshToken - uc.userRepo.GetUserFriends: %w", err)
	}
	for _, uFriend := range user.Friends {
		msgOnStatusFanout := entity.MessageOnlineStatusFanout{
			FriendId:          user.Id.String(),
			FriendUsername:    user.Username,
			FriendEmail:       user.Email,
			Online:            false,
			UserToGetNotified: uFriend.Username,
		}
		msgWs := &entity.MessageWs{
			Type:                  entity.MessageTypeOnlineStatusFanOut,
			MsgOnlineStatusFanout: msgOnStatusFanout,
		}
		isFriendOnline := uc.userRdsRepo.UserIsOnline(uFriend.Username)
		friendChatServerLocation, _ := uc.userRdsRepo.GetUserServerLocation(uFriend.Id.String())
		if friendChatServerLocation == entity.ChatServerNameGlobal.ChatServerName && isFriendOnline == true {
			// Jika teman user berada di server yg sama dg server user sender
			// publish ke channel chat-server sendiri
			fmt.Println("server sama", friendChatServerLocation, "  ", entity.ChatServerNameGlobal.ChatServerName)
			uc.pubSubRds.PublishToChannel(entity.ChatServerNameGlobal.ChatServerName, msgWs)
			continue
		}
		// publish ke channel ke chat-server teman
		fmt.Println("server beda", friendChatServerLocation, "  ", entity.ChatServerNameGlobal.ChatServerName)
		uc.pubSubRds.PublishToChannel(friendChatServerLocation, msgWs)
		continue
	}
	return nil
}
