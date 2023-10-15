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

// UserSetOnline set (key,value) = (userStatus:userId, timestampNow) in redis untuk tunjukkan user dg userId masih online selama 30 detik kedepan
func (r *UserRedisRepo) UserSetOnline(userId string) error {
	key := r.getKeyUserStatus(userId)
	return r.rds.Client.Set(context.Background(), key, time.Now().String(), 30*time.Second).Err()
}

// UserSetOffline delete key= userStatus:userId di redis untuk tunjukkan user dg userId sudah offline
func (r *UserRedisRepo) UserSetOffline(userId string) error {
	key := r.getKeyUserStatus(userId)
	return r.rds.Client.Del(context.Background(), key).Err()
}

// getKeyUserStatus membuat key dari online status di redis (userStatus:userId)
func (r *UserRedisRepo) getKeyUserStatus(userUUID string) string {
	return fmt.Sprintf("%s.%s", keyUserStatus, userUUID)
}

// UserIsOnline cek apakah user masih online, jika masih online err == null (ada recrd di redis) , jika offline return err ( tidak ada record di redis)
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
