package repo

import (
	"context"
	"fmt"
	"github.com/lintangbs/chat-be/internal/entity"
	"gorm.io/gorm"
	"time"
)

type AuthRepo struct {
	db *gorm.DB
}

type User struct {
	gorm.Model
	Username       string
	HashedPassword string
	Email          string
	CreatedAt      time.Time
	UpdatedAt      time.Time
}

func NewAuthRepo(db *gorm.DB) *AuthRepo {
	return &AuthRepo{db}
}

func (r *AuthRepo) CreateUser(ctx context.Context, e entity.CreateUserRequest) (entity.User, error) {
	// e.password sudah dihash
	var userDb User

	result := r.db.Where(&User{Username: e.Username}).Or(&User{Email: e.Email}).First(&userDb)

	if result.RowsAffected > 0 {
		//bad request
		return entity.User{}, fmt.Errorf("AuthRepo - Createuser - r.db.Where: User With same username or email already exists")
	}

	user := User{Username: e.Username, HashedPassword: e.Password, Email: e.Password}
	if result := r.db.Create(&user); result.Error != nil {
		// internal server error
		return entity.User{}, fmt.Errorf("AuthRepo - Createuser - r.db.Create: %w", result.Error)
	}

	res := entity.User{
		Username: user.Username,
		Email:    user.Email,
	}
	return res, nil
}

func (r *AuthRepo) GetUser(ctx context.Context, email string) (entity.GetUser, error) {
	//TODO implement me
	var userDb User
	result := r.db.Where(&User{Email: email}).First(&userDb)
	if err := result.Error; err != nil {
		return entity.GetUser{}, err
	}

	user := entity.GetUser{
		Id:             userDb.ID,
		Username:       userDb.Username,
		Email:          userDb.Email,
		HashedPassword: userDb.HashedPassword,
	}

	return user, nil
}
