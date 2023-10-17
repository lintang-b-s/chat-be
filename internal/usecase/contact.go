package usecase

import (
	"context"
	"fmt"
	"github.com/lintangbs/chat-be/internal/entity"
)

// Bussines logic untuk mengelola kontak dari user
type ContactUseCase struct {
	userRepo UserRepo
}

func NewContactUseCase(u UserRepo) *ContactUseCase {
	return &ContactUseCase{
		userRepo: u,
	}
}

// Menambah teman di dalam kontak usernya
func (uc *ContactUseCase) AddContact(ctx context.Context, a entity.AddFriendRequest) (entity.UserResponse, error) {
	user, err := uc.userRepo.AddFriend(ctx, a.MyUsername, a.FriendUsername)
	if err != nil {
		// Bad request User Not Found
		return entity.UserResponse{}, fmt.Errorf("AuthUseCase - AddContact - uc.userRepo.AddFriend: %w", err)
	}
	return user, nil
}

// mendapatkan semua kontak yang dimiliki user.
func (uc *ContactUseCase) GetContact(ctx context.Context, g entity.GetContactRequest) (entity.UserResponse, error) {
	user, err := uc.userRepo.GetUserFriends(ctx, g.MyUsername)
	if err != nil {
		// Bad request User Not Found
		return entity.UserResponse{}, fmt.Errorf("AuthUseCase - GetContact - uc.userRepo.GetUserFriends: %w", err)
	}

	return user, nil
}
