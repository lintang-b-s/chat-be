// Package usecase implements application business logic. Each logic group in own file.
package usecase

import (
	"context"

	"github.com/lintangbs/chat-be/internal/entity"
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
		Register(context.Context, entity.CreateUserRequest) (entity.User, error)
		Login(context.Context, entity.LoginUserRequest) (entity.LoginUserResponse, error)
	}

	// AuthRepo
	AuthRepo interface {
		CreateUser(context.Context, entity.CreateUserRequest) (entity.User, error)
		GetUser(context.Context, string) (entity.GetUser, error)
		CreateSession(context.Context, entity.CreateSessionRequest) (entity.Session, error)
	}

	// SessionRepo
	SessionRepo interface {
		CreateSession(context.Context, entity.CreateSessionRequest) (entity.Session, error)
	}
)
