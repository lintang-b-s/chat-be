package repo

import (
	"context"
	"errors"
	"fmt"
	"github.com/google/uuid"
	"github.com/lintangbs/chat-be/internal/entity"
	"gorm.io/gorm"
	"time"
)

type UserRepo struct {
	db *gorm.DB
}

var (
	ErrNotFoundContactErr       = errors.New("Not Found Contacts")
	AlreadyYourInYourContactErr = errors.New("Already in your contacts")
)

type User struct {
	gorm.Model
	ID             uuid.UUID
	Username       string
	HashedPassword string
	Email          string
	CreatedAt      time.Time
	UpdatedAt      time.Time
	Friends        []*User `gorm:"many2many:contacts"`
}

type contact struct {
	UserID    uuid.UUID
	FriendId  uuid.UUID
	CreatedAt time.Time
	UpdatedAt time.Time
	DeletedAt time.Time
}

func NewUserRepo(db *gorm.DB) *UserRepo {
	return &UserRepo{db}
}

// CreateUser membuat user baru di database
func (r *UserRepo) CreateUser(ctx context.Context, e entity.CreateUserRequest) (entity.UserResponse, error) {
	// e.password sudah dihash
	var userDb User

	result := r.db.Where(&User{Username: e.Username}).Or(&User{Email: e.Email}).First(&userDb)

	if result.RowsAffected > 0 {
		//bad request
		return entity.UserResponse{}, fmt.Errorf("AuthRepo - Createuser - r.db.Where: User With same username or email already exists")
	}

	userId := uuid.New()
	user := User{ID: userId, Username: e.Username, HashedPassword: e.Password, Email: e.Email, Friends: []*User{}}
	if result := r.db.Create(&user); result.Error != nil {
		// internal server error
		return entity.UserResponse{}, fmt.Errorf("AuthRepo - Createuser - r.db.Create: %w", result.Error)
	}

	res := entity.UserResponse{
		Id:       userId,
		Username: user.Username,
		Email:    user.Email,
	}
	return res, nil
}

// GetUser mendapatkan user dg yg memiliki email=email di db
func (r *UserRepo) GetUser(ctx context.Context, email string) (entity.GetUser, error) {
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

// AddFriend menambahkan daftar kontak yg dimiliki user
func (r *UserRepo) AddFriend(ctx context.Context, myUsername string, fUsername string) (entity.UserResponse, error) {
	var user User
	var friend User
	queryResF := r.db.Where(&User{Username: fUsername}).First(&friend)
	if err := queryResF.Error; err != nil {
		return entity.UserResponse{}, err
	}
	queryRes := r.db.Where(&User{Username: myUsername}).First(&user)
	if err := queryRes.Error; err != nil {
		return entity.UserResponse{}, err
	}

	var result []contact
	queryRes = r.db.Raw("SELECT * FROM contacts WHERE user_id=? AND friend_id=? ", user.ID, friend.ID).Scan(&result)
	if err := queryRes.Error; err != nil {
		return entity.UserResponse{}, err
	}
	if len(result) > 0 {
		return entity.UserResponse{}, AlreadyYourInYourContactErr
	}

	user.Friends = append(user.Friends, &friend)
	friend.Friends = append(friend.Friends, &user)
	r.db.Save(&user)
	r.db.Save(&friend)

	var friendRes []entity.UserResponse
	for _, uFriend := range user.Friends {

		frnd := entity.UserResponse{
			Id:       uFriend.ID,
			Username: uFriend.Username,
			Email:    uFriend.Email,
		}
		friendRes = append(friendRes, frnd)
	}

	resp := entity.UserResponse{
		Id:       user.ID,
		Username: user.Username,
		Email:    user.Email,
		Friends:  friendRes,
	}
	return resp, nil
}

// GetUserFriends mendpatkan kontak yang dimiliki user
func (r *UserRepo) GetUserFriends(ctx context.Context, username string) (entity.UserResponse, error) {
	var user User
	queryRes := r.db.Where(&User{Username: username}).Preload("Friends").First(&user)
	if err := queryRes.Error; err != nil {
		return entity.UserResponse{}, err
	}

	var friendRes []entity.UserResponse
	for _, uFriend := range user.Friends {

		frnd := entity.UserResponse{
			Id:       uFriend.ID,
			Username: uFriend.Username,
			Email:    uFriend.Email,
		}
		friendRes = append(friendRes, frnd)
	}

	res := entity.UserResponse{
		Id:       user.ID,
		Username: user.Username,
		Email:    user.Email,
		Friends:  friendRes,
	}
	return res, nil
}

// GetUserFriend untuk mengecek apakah friendUsername teman dari user
func (r *UserRepo) GetUserFriend(ctx context.Context, myUsername string, friendUsername string) error {
	var user User
	var friend User
	queryResFriend := r.db.Where(&User{Username: friendUsername}).First(&friend)
	if err := queryResFriend.Error; err != nil {
		return err
	}
	queryRes := r.db.Where(&User{Username: myUsername}).First(&user)
	if err := queryRes.Error; err != nil {
		return err
	}
	var result []contact
	queryRes = r.db.Raw("SELECT * FROM contacts WHERE user_id=? AND friend_id=? ", user.ID, friend.ID).Scan(&result)
	if err := queryRes.Error; err != nil {
		return err
	}
	if len(result) == 0 {
		return ErrNotFoundContactErr
	}

	return nil
}

//// GetAllUsers Mendapatkan semua user
//func (r *UserRepo) GetAllUsers(ctx context.Context) ([]entity.UserResponse, error) {
//
//}

// GetUser mendapatkan user berdasarkan username di db
func (r *UserRepo) GetUserByUsername(username string) (entity.GetUser, error) {
	//TODO implement me
	var userDb User
	result := r.db.Where(&User{Username: username}).First(&userDb)
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
