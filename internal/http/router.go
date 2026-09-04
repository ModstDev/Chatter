package server

import (
	"net/http"

	"github.com/ModstDev/Chatter/internal/auth"
	"github.com/ModstDev/Chatter/internal/conversation"
	"github.com/ModstDev/Chatter/internal/message"
	"github.com/ModstDev/Chatter/internal/user"
	"github.com/ModstDev/Chatter/internal/websocket"
)

func NewRouter(userHandler *user.Handler,
	authHandler *auth.Handler,
	conversationHandler *conversation.Handler,
	messageHandler *message.Handler,
	websocketHandler *websocket.Handler,
	tokenManager *auth.TokenManager,
) http.Handler {
	mux := http.NewServeMux()

	mux.Handle("GET /api/v1/users",
		auth.AuthMiddleware(tokenManager,
			http.HandlerFunc(userHandler.List)))

	mux.Handle("GET /api/v1/conversations",
		auth.AuthMiddleware(tokenManager,
			http.HandlerFunc(conversationHandler.List)))

	mux.Handle(
		"GET /api/v1/conversations/{conversationID}/messages",
		auth.AuthMiddleware(
			tokenManager,
			http.HandlerFunc(messageHandler.ListHistory),
		),
	)

	mux.Handle(
		"GET /api/v1/ws",
		http.HandlerFunc(websocketHandler.ServeHTTP),
	)

	mux.HandleFunc("POST /api/v1/register", userHandler.Register)
	mux.HandleFunc("POST /api/v1/auth/login", authHandler.Login)
	mux.HandleFunc("POST /api/v1/auth/refresh", authHandler.Refresh)
	mux.HandleFunc("POST /api/v1/auth/logout", authHandler.Logout)

	mux.Handle("POST /api/v1/conversations",
		auth.AuthMiddleware(tokenManager,
			http.HandlerFunc(conversationHandler.Create)))

	return mux
}
