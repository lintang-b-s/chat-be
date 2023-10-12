package redispkg

import (
	"log"
	"sync"

	"github.com/redis/go-redis/v9"
)

type Redis struct {
	Client             *redis.Client
	ChannelsPubSub     map[string]*ChannelPubSub
	ChannelsPubSubSync *sync.RWMutex
}

type ChannelPubSub struct {
	CloseChan  chan struct{}
	ClosedChan chan struct{}
	PubSub     *redis.PubSub
}

func (channel *ChannelPubSub) Channel() <-chan *redis.Message {
	return channel.PubSub.Channel()
}

func (channel *ChannelPubSub) Close() <-chan struct{} {
	return channel.CloseChan
}

func (channel *ChannelPubSub) Closed() <-chan struct{} {
	return channel.ClosedChan
}

func NewRedis(addr, password string) (*Redis, error) {
	log.Println("Initialized redispkg client", addr, password)

	opt := &redis.Options{
		Addr: addr,
	}

	if password != "" {
		opt.Password = password
	}

	c := redis.NewClient(opt)

	r := &Redis{
		Client:             c,
		ChannelsPubSub:     make(map[string]*ChannelPubSub, 0),
		ChannelsPubSubSync: &sync.RWMutex{},
	}

	return r, nil
}
