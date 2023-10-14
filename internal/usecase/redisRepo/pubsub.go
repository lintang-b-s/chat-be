package redisRepo

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"github.com/lintangbs/chat-be/internal/entity"
	"github.com/lintangbs/chat-be/pkg/redispkg"
	"github.com/redis/go-redis/v9"
)

type PubSubRedis struct {
	rds *redispkg.Redis
}

func NewPubSubRedis(rds *redispkg.Redis) *PubSubRedis {
	return &PubSubRedis{rds}
}

func (p *PubSubRedis) SubscribeToChannel(ctx context.Context, username string) *redis.PubSub {
	return p.rds.Client.Subscribe(ctx, username)
}

func (p *PubSubRedis) PublishToChannel(to string, msg *entity.MessageWs) error {
	buff := bytes.NewBufferString("")
	enc := json.NewEncoder(buff)
	err := enc.Encode(msg)

	if err != nil {
		fmt.Println(err)
	}

	return p.rds.Client.Publish(context.Background(), to, buff.String()).Err()
}
