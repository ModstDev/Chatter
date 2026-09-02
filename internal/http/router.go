package server

import (
	"net/http"

	"github.com/ModstDev/Chatter/internal/auth"
	"github.com/ModstDev/Chatter/internal/user"
)

func NewRouter(userHandler *user.Handler,
	authHandler *auth.Handler,
) http.Handler {
	mux := http.NewServeMux()

	mux.HandleFunc("POST /register", userHandler.Register)
	mux.HandleFunc("POST /auth/login", authHandler.Login)
	mux.HandleFunc("POST /auth/refresh", authHandler.Refresh)
	mux.HandleFunc("POST /auth/logout", authHandler.Logout)

	return mux
}
