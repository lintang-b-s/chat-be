package usecase

import (
	"context"
	"errors"
	"github.com/gorilla/websocket"
	"net/http"
)

var (
	WebsocketConnectionError   = errors.New("websocketc connection error")
	WebsocketUnauthorizedError = errors.New("websocketc unauthorized error")
)

// WebsocketUseCase bussines logic websocketc
type WebsocketUseCase struct {
	otpRepo OtpRepo
	userPg  UserRepo
	chat    *Chat
}

// NewWebsocket Create new websocketUseCase
func NewWebsocket(otp OtpRepo, chat *Chat,
	uPg UserRepo) *WebsocketUseCase {
	return &WebsocketUseCase{
		otp, uPg, chat}
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
	_ = uc.chat.Register(ctx, conn, username, userDb.Id.String())

	return nil
}
