// Package app configures and runs application.
package app

import (
	"fmt"
	"os"
	"os/signal"
	"syscall"

	"github.com/gin-gonic/gin"

	"github.com/lintangbs/chat-be/config"
	v1 "github.com/lintangbs/chat-be/internal/controller/http/v1"
	"github.com/lintangbs/chat-be/internal/usecase"
	"github.com/lintangbs/chat-be/internal/usecase/repo"
	"github.com/lintangbs/chat-be/internal/usecase/webapi"
	"github.com/lintangbs/chat-be/pkg/gorm"
	"github.com/lintangbs/chat-be/pkg/httpserver"
	"github.com/lintangbs/chat-be/pkg/logger"
	"github.com/lintangbs/chat-be/pkg/postgres"
	"github.com/lintangbs/chat-be/pkg/redis"
)

// Run creates objects via constructors.
func Run(cfg *config.Config) {
	l := logger.New(cfg.Log.Level)

	// Repository
	pg, err := postgres.New(cfg.PG.URL, postgres.MaxPoolSize(cfg.PG.PoolMax))
	if err != nil {
		l.Fatal(fmt.Errorf("app - Run - postgres.New: %w", err))
	}
	defer pg.Close()

	// Redis repo
	redis, err := redis.NewRedis(cfg.Redis.Address, cfg.Redis.Password)
	if err != nil {
		l.Fatal(fmt.Errorf("app - Run - redis - redis.NewRedis: %w", err))
	}

	// gorm repo
	gorm, err := gorm.NewGorm()

	// Use case
	translationUseCase := usecase.New(
		repo.New(pg),
		webapi.New(),
	)

	authUseCase := usecase.NewAuthUseCase(
		repo.NewAuthRepo(gorm.Pool),
	)

	// RabbitMQ RPC Server
	// rmqRouter := amqprpc.NewRouter(translationUseCase)

	// rmqServer, err := server.New(cfg.RMQ.URL, cfg.RMQ.ServerExchange, rmqRouter, l)
	if err != nil {
		l.Fatal(fmt.Errorf("app - Run - rmqServer - server.New: %w", err))
	}

	// HTTP Server
	handler := gin.New()
	v1.NewRouter(handler, l, translationUseCase, authUseCase)
	httpServer := httpserver.New(handler, httpserver.Port(cfg.HTTP.Port))

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
