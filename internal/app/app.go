// Package app configures and runs application.
package app

import (
	"flag"
	"fmt"
	"github.com/gin-contrib/cors"
	"github.com/lintangbs/chat-be/internal/usecase/redisRepo"
	"github.com/lintangbs/chat-be/internal/usecase/webapi"
	"github.com/lintangbs/chat-be/internal/usecase/websocketc"
	"github.com/lintangbs/chat-be/internal/util/gopool"
	"github.com/lintangbs/chat-be/internal/util/jwt"
	"github.com/lintangbs/chat-be/pkg/redispkg"
	"github.com/mailru/easygo/netpoll"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/gin-gonic/gin"

	"github.com/lintangbs/chat-be/config"
	v1 "github.com/lintangbs/chat-be/internal/controller/http/v1"
	"github.com/lintangbs/chat-be/internal/usecase"
	"github.com/lintangbs/chat-be/internal/usecase/repo"
	"github.com/lintangbs/chat-be/pkg/gorm"
	"github.com/lintangbs/chat-be/pkg/httpserver"
	"github.com/lintangbs/chat-be/pkg/logger"
)

var (
	workers   = flag.Int("workers", 128, "max workers count")
	queue     = flag.Int("queue", 1, "workers task queue size")
	ioTimeout = flag.Duration("io_timeout", time.Millisecond*100, "i/o operations timeout")
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

	poller, err := netpoll.New(nil)
	if err != nil {
		l.Fatal(fmt.Errorf("app - Run - netpoll.New - netpoll.New: %w", err))
	}

	authUseCase := usecase.NewAuthUseCase(
		*repo.NewUserRepo(gorm.Pool),
		jwtTokenMaker,
		repo.NewSessionRepo(gorm.Pool),
		*redisRepo.NewOtp(redis),
		redisRepo.NewUserRedisrepo(redis),
		redisRepo.NewPubSubRedis(redis),
	)

	var (
		// Make pool of X size, Y sized work queue and one pre-spawned
		// goroutine.
		pool = gopool.NewPool(*workers, *queue, 1)
	)

	webSocketUseCase := usecase.NewWebsocket(
		*redisRepo.NewOtp(redis),
		*websocketc.NewChat(*redisRepo.NewPubSubRedis(redis),
			*edenAi,
			*repo.NewUserRepo(gorm.Pool),
			*redis,
			*redisRepo.NewUserRedisrepo(redis),
		),
		poller,
		pool,
		*repo.NewUserRepo(gorm.Pool),
	)

	contactUseCase := usecase.NewContactUseCase(
		*repo.NewUserRepo(gorm.Pool),
	)

	// HTTP Server
	handler := gin.New()

	handler.Use(cors.Default())

	v1.NewRouter(handler, l, authUseCase, webSocketUseCase, *contactUseCase, jwtTokenMaker)
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
