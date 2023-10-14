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
	key := getKeyUserStatus(uuid)
	return r.rds.Client.Set(context.Background(), key, time.Now().String(), 30*time.Second).Err()
}

func getKeyUserStatus(userUUID string) string {
	return fmt.Sprintf("%s.%s", keyUserStatus, userUUID)
}
