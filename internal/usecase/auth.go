package usecase

import (
	"context"
	"fmt"
	"github.com/lintangbs/chat-be/internal/entity"
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
	createdUser, err := uc.authRepo.CreateUser(ctx, c)
	if err != nil {
		return entity.CreateUserResponse{}, fmt.Errorf("AuthUseCase - Register - uc.authRepo.CreateUser: %w", err)
	}

	return createdUser, nil
}
