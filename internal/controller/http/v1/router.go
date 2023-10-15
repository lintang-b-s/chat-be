// Package v1 implements routing paths. Each services in own file.
package v1

import (
	"github.com/lintangbs/chat-be/internal/util/jwt"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	swaggerFiles "github.com/swaggo/files"
	ginSwagger "github.com/swaggo/gin-swagger"

	// Swagger docs.
	_ "github.com/lintangbs/chat-be/docs"
	"github.com/lintangbs/chat-be/internal/usecase"
	"github.com/lintangbs/chat-be/pkg/logger"
)

// NewRouter -.
// Swagger spec:
// @title       Go Clean Template API
// @description Using a translation service as an example
// @version     1.0
// @host        localhost:8080
// @BasePath    /v1
func NewRouter(handler *gin.Engine, l logger.Interface, a usecase.Auth, ws usecase.Websocket, cont usecase.ContactUseCase, jwt jwt.JwtTokenMaker) {
	// Options
	handler.Use(gin.Logger())
	handler.Use(gin.Recovery())

	// Swagger
	swaggerHandler := ginSwagger.DisablingWrapHandler(swaggerFiles.Handler, "DISABLE_SWAGGER_HTTP_HANDLER")
	handler.GET("/swagger/*any", swaggerHandler)

	// K8s probe
	handler.GET("/healthz", func(c *gin.Context) { c.Status(http.StatusOK) })

	// Prometheus metrics
	handler.GET("/metrics", gin.WrapH(promhttp.Handler()))

	// Routers
	h := handler.Group("/v1")
	{
		newAuthRoutes(h, a, l, jwt)
		NewWebsocketRoutes(h, ws, l)
		NewContactRoutes(h, &cont, l, jwt)
	}
}
