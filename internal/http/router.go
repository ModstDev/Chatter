package server

import (
	"net/http"

	"github.com/ModstDev/Chatter/internal/auth"
	"github.com/ModstDev/Chatter/internal/conversation"
	"github.com/ModstDev/Chatter/internal/user"
)

func NewRouter(userHandler *user.Handler,
	authHandler *auth.Handler,
	conversationHandler *conversation.Handler,
	tokenManager *auth.TokenManager,
) http.Handler {
	mux := http.NewServeMux()

	mux.Handle("GET /users",
		auth.AuthMiddleware(tokenManager,
			http.HandlerFunc(userHandler.List)))

	mux.Handle("GET /conversations",
		auth.AuthMiddleware(tokenManager,
			http.HandlerFunc(conversationHandler.List)))

	mux.HandleFunc("POST /register", userHandler.Register)
	mux.HandleFunc("POST /auth/login", authHandler.Login)
	mux.HandleFunc("POST /auth/refresh", authHandler.Refresh)
	mux.HandleFunc("POST /auth/logout", authHandler.Logout)

	mux.Handle("POST /conversations",
		auth.AuthMiddleware(tokenManager,
			http.HandlerFunc(conversationHandler.Create)))

	return mux
}
