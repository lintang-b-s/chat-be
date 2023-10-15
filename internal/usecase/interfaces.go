// Package usecase implements application business logic. Each logic group in own file.
package usecase

import (
	"context"
	"github.com/google/uuid"
	"github.com/gorilla/websocket"
	"github.com/lintangbs/chat-be/internal/entity"
	"github.com/lintangbs/chat-be/internal/usecase/websocketc"
	"github.com/lintangbs/chat-be/pkg/redispkg"
	"github.com/redis/go-redis/v9"
	"net/http"
)

//go:generate mockgen -source=interfaces.go -destination=./mocks_test.go -package=usecase_test

type (
	// Translation -.
	Translation interface {
		Translate(context.Context, entity.Translation) (entity.Translation, error)
		History(context.Context) ([]entity.Translation, error)
	}

	// TranslationRepo -.
	TranslationRepo interface {
		Store(context.Context, entity.Translation) error
		GetHistory(context.Context) ([]entity.Translation, error)
	}

	// TranslationWebAPI -.
	TranslationWebAPI interface {
		Translate(entity.Translation) (entity.Translation, error)
	}

	// Auth Use Case
	Auth interface {
		Register(context.Context, entity.CreateUserRequest) (entity.UserResponse, error)
		Login(context.Context, entity.LoginUserRequest) (entity.LoginUserResponse, error)
		RenewAccessToken(context.Context, entity.RenewAccessTokenRequest) (entity.RenewAccessTokenResponse, error)
		DeleteRefreshToken(context.Context, entity.DeleteRefreshTokenRequest) error
	}

	// AuthRepo
	UserRepo interface {
		CreateUser(context.Context, entity.CreateUserRequest) (entity.UserResponse, error)
		GetUser(context.Context, string) (entity.GetUser, error)
		AddFriend(context.Context, string, string) (entity.UserResponse, error)
		GetUserFriends(context.Context, string) (entity.UserResponse, error)
		GetUserFriend(context.Context, string, string) error
		//GetAllUsers(context.Context) ([]entity.UserResponse, error)
		GetUserByUsername(string) (entity.GetUser, error)
	}

	// SessionRepo
	SessionRepo interface {
		CreateSession(context.Context, entity.CreateSessionRequest) (entity.Session, error)
		GetSession(context.Context, uuid.UUID) (entity.Session, error)
		DeleteSession(context.Context, uuid.UUID) error
	}

	// Websocket usecase
	Websocket interface {
		WebsocketHandler(http.ResponseWriter, *http.Request, context.Context) error
	}

	// OtpRepo OtpRepo
	OtpRepo interface {
		GetOtp(string, context.Context, string) error
		CreateOtp(context.Context, string) (string, error)
	}

	// Chat
	ChatHub interface {
		Register(context.Context, *websocket.Conn, string, string) *websocketc.User
		SubscribePubSubAndSendToClient(*redispkg.ChannelPubSub)
	}

	// EdenAiApi
	EdenAiApi interface {
		GenerateText(string) string
	}

	Contact interface {
		AddContact(context.Context, entity.AddFriendRequest) (entity.UserResponse, error)
		GetContact(context.Context, entity.GetContactRequest) (entity.UserResponse, error)
	}

	//	PubSubRedis
	PubSubRedis interface {
		SubscribeToChannel(context.Context, string) *redis.PubSub
		PublishToChannel(string, *entity.MessageWs) error
	}

	// UserRedisRepo
	UserRedisRepo interface {
		UserSetOnline(string) error
		UserIsOnline(string) bool
		SetUserServerLocation(string) error
		GetUserServerLocation(string) (string, error)
	}
)
