// Package app configures and runs application.
package app

import (
	"context"
	"fmt"
	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	uuid2 "github.com/google/uuid"
	"github.com/lintangbs/chat-be/internal/entity"
	"github.com/lintangbs/chat-be/internal/usecase/redisRepo"
	"github.com/lintangbs/chat-be/internal/usecase/webapi"
	"github.com/lintangbs/chat-be/internal/usecase/websocketc"
	"github.com/lintangbs/chat-be/internal/util/jwt"
	"github.com/lintangbs/chat-be/pkg/redispkg"
	"os"
	"os/signal"
	"syscall"

	"github.com/lintangbs/chat-be/config"
	v1 "github.com/lintangbs/chat-be/internal/controller/http/v1"
	"github.com/lintangbs/chat-be/internal/usecase"
	"github.com/lintangbs/chat-be/internal/usecase/repo"
	"github.com/lintangbs/chat-be/pkg/gorm"
	"github.com/lintangbs/chat-be/pkg/httpserver"
	"github.com/lintangbs/chat-be/pkg/logger"
)

// Run creates objects via constructors.
func Run(cfg *config.Config) {
	l := logger.New(cfg.Log.Level)
	// Redis repo
	redis, err := redispkg.NewRedis(cfg.Redis.Address, cfg.Redis.Password)
	if err != nil {
		l.Fatal(fmt.Errorf("app - Run - redispkg - redispkg.NewRedis: %w", err))
	}

	// gorm repo
	gorm, err := gorm.NewGorm(cfg.Postgres.Username, cfg.Postgres.Password)

	// jwt
	jwtTokenMaker, err := jwt.NewJWTMaker("eyJhbGciOiJIUzI1NiJ9.eyJSb2xlIjoiQWRtaW4iLCJJc3N1ZXIiOiJJc3N1ZXIiLCJVc2VybmFtZSI6IkphdmFJblVzZSIsImV4cCI6MTY5NjkyMDA0MywiaWF0IjoxNjk2OTIwMDQzfQ.6x0sgC9T1l64c2IpuCT3WBnw02ZRmHZI-iq4rP5cA9s")

	// EdenAi API
	edenAi := webapi.NewEdenAIAPI(cfg.EdenAi.ApiKey)

	if err != nil {
		l.Fatal(fmt.Errorf("app - Run - jwtTokenMaker - jwt.NewJWTMaker: %w", err))
	}

	authUseCase := usecase.NewAuthUseCase(
		repo.NewUserRepo(gorm.Pool),
		jwtTokenMaker,
		repo.NewSessionRepo(gorm.Pool),
		redisRepo.NewOtp(redis),
		redisRepo.NewPubSubRedis(redis),
		redisRepo.NewUserRedisrepo(redis),
	)

	chat := websocketc.NewChat(
		redisRepo.NewPubSubRedis(redis),
		edenAi,
		repo.NewUserRepo(gorm.Pool),
		redis,
		redisRepo.NewUserRedisrepo(redis),
	)

	go chat.Run()

	entity.ChatServerNameGlobal = &entity.ServerName{
		ChatServerName: "chat-server" + uuid2.New().String(),
	}

	fmt.Println("chat-server: ", entity.ChatServerNameGlobal)

	webSocketUseCase := usecase.NewWebsocket(
		redisRepo.NewOtp(redis),
		chat,
		repo.NewUserRepo(gorm.Pool),
	)

	contactUseCase := usecase.NewContactUseCase(
		repo.NewUserRepo(gorm.Pool),
	)

	// HTTP Server
	handler := gin.New()

	handler.Use(cors.Default())

	v1.NewRouter(handler, l, authUseCase, webSocketUseCase, *contactUseCase, jwtTokenMaker)
	httpServer := httpserver.New(handler, httpserver.Port(cfg.HTTP.Port))

	// start subscriber channel chat-server-serverName
	// Client subscribe to redis channel (nama channel ya username si user sendiri)
	// Untuk menerima message dari kontaknya
	pubSub := chat.PubSub.SubscribeToChannel(context.Background(), entity.ChatServerNameGlobal.ChatServerName)

	newChannelPubSub := &redispkg.ChannelPubSub{
		CloseChan:  make(chan struct{}, 1),
		ClosedChan: make(chan struct{}, 1),
		PubSub:     pubSub,
	}

	chat.Rds.ChannelsPubSubSync.Lock()

	if _, ok := chat.Rds.ChannelsPubSub[entity.ChatServerNameGlobal.ChatServerName]; !ok {
		chat.Rds.ChannelsPubSub[entity.ChatServerNameGlobal.ChatServerName] = newChannelPubSub
	}
	chat.Rds.ChannelsPubSubSync.Unlock()

	go chat.SubscribePubSubAndSendToClient(newChannelPubSub)

	// Waiting signal
	interrupt := make(chan os.Signal, 1)
	signal.Notify(interrupt, os.Interrupt, syscall.SIGTERM)

	select {
	case s := <-interrupt:
		l.Info("app - Run - signal: " + s.String())
	case err = <-httpServer.Notify():
		l.Error(fmt.Errorf("app - Run - httpServer.Notify: %w", err))

	}

	// Shutdown
	err = httpServer.Shutdown()
	if err != nil {
		l.Error(fmt.Errorf("app - Run - httpServer.Shutdown: %w", err))
	}

}
