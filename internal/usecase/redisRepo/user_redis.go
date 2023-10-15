package redisRepo

import (
	"context"
	"fmt"
	"github.com/lintangbs/chat-be/internal/entity"
	"github.com/lintangbs/chat-be/pkg/redispkg"
	"time"
)

type UserRedisRepo struct {
	rds *redispkg.Redis
}

const (
	keyUserStatus         = "userStatus"
	keyUserServerLocation = "userServer"
)

func NewUserRedisrepo(rds *redispkg.Redis) *UserRedisRepo {
	return &UserRedisRepo{rds}
}

// UserSetOnline set user online status di redis
func (r *UserRedisRepo) UserSetOnline(uuid string) error {
	key := r.getKeyUserStatus(uuid)
	return r.rds.Client.Set(context.Background(), key, time.Now().String(), 30*time.Second).Err()
}

func (r *UserRedisRepo) getKeyUserStatus(userUUID string) string {
	return fmt.Sprintf("%s.%s", keyUserStatus, userUUID)
}

// UserIsOnline get user online status dari rediss
func (r *UserRedisRepo) UserIsOnline(uuid string) bool {
	key := r.getKeyUserStatus(uuid)
	err := r.rds.Client.Get(context.Background(), key).Err()
	if err == nil {
		// user online karena key ada di di redis
		return true
	}
	// user offline karena key tidak ada di redis (ada error)
	return false
}

func (r *UserRedisRepo) constructKey(key string, userId string) string {
	return fmt.Sprintf("%s.%s", key, userId)
}

// SetUserServerLocation Set user chat-server location in redis
func (r *UserRedisRepo) SetUserServerLocation(userId string) error {
	key := r.constructKey(keyUserServerLocation, userId)
	return r.rds.Client.HSet(context.Background(), key, userId, entity.ChatServerNameGlobal.ChatServerName).Err()
}

// GetUserServerLocation  get user chat-server location in redis
func (r *UserRedisRepo) GetUserServerLocation(userId string) (string, error) {
	key := r.constructKey(keyUserServerLocation, userId)
	res := r.rds.Client.HGet(context.Background(), key, userId)
	if err := res.Err(); err != nil {
		return "", err
	}
	return res.Val(), nil
}
