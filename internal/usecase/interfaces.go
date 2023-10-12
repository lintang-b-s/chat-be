// Package usecase implements application business logic. Each logic group in own file.
package usecase

import (
	"context"
	"github.com/google/uuid"
	"github.com/lintangbs/chat-be/internal/entity"
	"github.com/lintangbs/chat-be/internal/usecase/websocket"
	"net"
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
	}

	// AuthRepo
	AuthRepo interface {
		CreateUser(context.Context, entity.CreateUserRequest) (entity.UserResponse, error)
		GetUser(context.Context, string) (entity.GetUser, error)
	}

	// SessionRepo
	SessionRepo interface {
		CreateSession(context.Context, entity.CreateSessionRequest) (entity.Session, error)
		GetSession(context.Context, uuid.UUID) (entity.Session, error)
	}

	Websocket interface {
		WebsocketHandler(http.ResponseWriter, *http.Request, context.Context) error
	}

	OtpRepo interface {
		GetOtp(string, context.Context) (string, error)
		CreateOtp(context.Context) error
	}

	Chat interface {
		Register(ctx context.Context, conn net.Conn, username string) *websocket.User
	}
)
