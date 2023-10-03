package redis

import (
	"fmt"
	"log"

	"github.com/go-redis/redis"
)

type Redis struct {
	client *redis.Client
}


func NewRedis(addr, password string) (*Redis, error) {
	log.Println("Initialized redis client", addr, password)

	opt := &redis.Options{
		Addr: addr,
	}

	if password != ""{
		opt.Password = password
	}

	c := redis.NewClient(opt)
	if err := c.Ping().Err(); err != nil{
		return nil, fmt.Errorf("redis - NewRedis - NewClient == 0: %w", err)
	}

	r := &Redis {
		client: c,
	}
	return r, nil
}
