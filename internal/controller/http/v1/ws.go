package v1

import (
	"github.com/gin-gonic/gin"
	"github.com/lintangbs/chat-be/internal/usecase"
	"github.com/lintangbs/chat-be/pkg/logger"
	"net/http"
)

type WebbsocketRoutes struct {
	w usecase.Websocket
	l logger.Interface
}

func NewWebsocketRoutes(h *gin.RouterGroup, w usecase.Websocket, l logger.Interface) {

	r := &WebbsocketRoutes{w, l}

	h.GET("/ws", r.websocketHandler)
}

// websocketHandler handler saat buka koneksi websocket
func (r *WebbsocketRoutes) websocketHandler(c *gin.Context) {
	err := r.w.WebsocketHandler(c.Writer, c.Request, c.Request.Context())
	if err != nil {
		if err == usecase.WebsocketUnauthorizedError {
			//w.WriteHeader(http.StatusUnauthorized)
			ErrorResponse(c, http.StatusUnauthorized, "Websocket connection unauthorized")
			return
		}

		ErrorResponse(c, http.StatusInternalServerError, "Websocket service error")
		return
	}

}
