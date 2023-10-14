package redisRepo

import (
	"context"
	"fmt"
	"github.com/lintangbs/chat-be/pkg/redispkg"
	"time"
)

type UserRedisRepo struct {
	rds *redispkg.Redis
}

const (
	keyUserStatus = "userStatus"
)

func NewUserRedisrepo(rds *redispkg.Redis) *UserRedisRepo {
	return &UserRedisRepo{rds}
}

func (r *UserRedisRepo) UserSetOnline(uuid string) error {
	key := r.getKeyUserStatus(uuid)
	return r.rds.Client.Set(context.Background(), key, time.Now().String(), 30*time.Second).Err()
}

func (r *UserRedisRepo) getKeyUserStatus(userUUID string) string {
	return fmt.Sprintf("%s.%s", keyUserStatus, userUUID)
}

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
