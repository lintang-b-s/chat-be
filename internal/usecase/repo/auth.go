package repo

import (
	"context"
	"fmt"

	"github.com/lintangbs/chat-be/internal/entity"
	"gorm.io/gorm"
)

type AuthRepo struct {
	db *gorm.DB
}

type User struct {
	gorm.Model
	Username string
	Password string
	Email    string
}

func NewAuthRepo(db *gorm.DB) *AuthRepo {
	return &AuthRepo{db}
}

func (r *AuthRepo) CreateUser(ctx context.Context, e entity.CreateUserRequest) (entity.CreateUserResponse, error) {
	// e.password sudah dihash
	user := User{Username: e.Username, Password: e.Password, Email: e.Password}
	if result := r.db.Create(&user); result.Error != nil {
		return entity.CreateUserResponse{}, fmt.Errorf("AuthRepo - Createuser - r.db.Create: %w", result.Error)
	}

	res := entity.CreateUserResponse{
		Username: user.Username,
		Email:    user.Email,
	}
	return res, nil
}
