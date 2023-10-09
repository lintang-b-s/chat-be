package usecase

import (
	"context"
	"fmt"
	"github.com/lintangbs/chat-be/internal/entity"
	"github.com/lintangbs/chat-be/internal/util"
)

type AuthUseCase struct {
	authRepo AuthRepo
}

func NewAuthUseCase(r AuthRepo) *AuthUseCase {
	return &AuthUseCase{
		authRepo: r,
	}
}

func (uc *AuthUseCase) Register(ctx context.Context, c entity.CreateUserRequest) (entity.CreateUserResponse, error) {
	hashedPassword, err := util.HashPassword(c.Password)
	if err != nil {
		return entity.CreateUserResponse{}, fmt.Errorf("AuthUseCase - Register -  util.HashPassword: %w", err)
	}
	c.Password = hashedPassword
	createdUser, err := uc.authRepo.CreateUser(ctx, c)
	if err != nil {
		return entity.CreateUserResponse{}, fmt.Errorf("AuthUseCase - Register - uc.authRepo.CreateUser: %w", err)
	}

	return createdUser, nil
}
