package amqprpc

import (
	"github.com/lintangbs/chat-be/internal/usecase"
	"github.com/lintangbs/chat-be/pkg/rabbitmq/rmq_rpc/server"
)

// NewRouter -.
func NewRouter(t usecase.Translation) map[string]server.CallHandler {
	routes := make(map[string]server.CallHandler)
	{
		newTranslationRoutes(routes, t)
	}

	return routes
}
