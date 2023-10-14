package usecase

import (
	"context"
	"errors"
	"github.com/gorilla/websocket"
	"github.com/lintangbs/chat-be/internal/usecase/redisRepo"
	"github.com/lintangbs/chat-be/internal/usecase/repo"
	"github.com/lintangbs/chat-be/internal/usecase/websocketc"
	"github.com/lintangbs/chat-be/internal/util/gopool"
	"github.com/mailru/easygo/netpoll"
	"net/http"
)

var (
	WebsocketConnectionError   = errors.New("websocketc connection error")
	WebsocketUnauthorizedError = errors.New("websocketc unauthorized error")
)

// WebsocketUseCase bussines logic websocketc
type WebsocketUseCase struct {
	otpRepo redisRepo.OtpRepo
	userPg  repo.UserRepo
	chat    websocketc.Chat
	poller  netpoll.Poller
	gopool  *gopool.Pool
}

// NewWebsocket Create new websocketUseCase
func NewWebsocket(otp redisRepo.OtpRepo, chat websocketc.Chat, p netpoll.Poller, gp *gopool.Pool,
	uPg repo.UserRepo) *WebsocketUseCase {
	return &WebsocketUseCase{
		otp, uPg, chat, p, gp}
}

var upgrader = websocket.Upgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
}

func (uc *WebsocketUseCase) WebsocketHandler(w http.ResponseWriter, r *http.Request, ctx context.Context) error {

	otp := r.URL.Query().Get("otp")
	username := r.URL.Query().Get("username")
	// Grab the OTP in the Get param
	if otp == "" {
		// Tell the user its not authorized
		return WebsocketUnauthorizedError
	}
	if username == "" {
		return WebsocketUnauthorizedError
	}

	userDb, err := uc.userPg.GetUserByUsername(username)
	if err != nil {
		// Tell the user its not authorized
		return WebsocketUnauthorizedError
	}

	err = uc.otpRepo.GetOtp(otp, ctx, username)
	if err != nil {
		return WebsocketUnauthorizedError
	}

	// handle is a new incoming connection handler.
	// It upgrades TCP connection to WebSocket, registers netpoll listener on
	// it and stores it as a chat user in Chat instance.
	//
	// We will call it below within accept() loop.
	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		return WebsocketConnectionError
	}

	// Register incoming user in chat.
	user := uc.chat.Register(ctx, conn, username, userDb.Id.String())

	// Create netpoll event descriptor for conn.
	// We want to handle only read events of it.
	desc := netpoll.Must(netpoll.HandleRead(conn.UnderlyingConn()))

	// Subscribe to events about conn.
	uc.poller.Start(desc, func(ev netpoll.Event) {
		if ev&(netpoll.EventReadHup|netpoll.EventHup) != 0 {
			// When ReadHup or Hup received, this mean that client has
			// closed at least write end of the connection or connections
			// itself. So we want to stop receive events about such conn
			// and remove it from the chat registry.
			uc.poller.Stop(desc)
			return
		}
		// Here we can read some new message from connection.
		// We can not read it right here in callback, because then we will
		// block the h.poller's inner loop.
		// We do not want to spawn a new goroutine to read single message.
		// But we want to reuse previously spawned goroutine.
		uc.gopool.Schedule(func() {
			if err := user.Receive(); err != nil {
				// When receive failed, we can only disconnect broken
				// connection and stop to receive events about it.
				uc.poller.Stop(desc)
			}
		})
	})

	return nil
}
