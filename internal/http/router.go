package server

import (
	"net/http"

	"github.com/ModstDev/Chatter/internal/user"
)

func NewRouter(userHandler *user.Handler) http.Handler {
	mux := http.NewServeMux()

	mux.HandleFunc("POST /register", userHandler.Register)

	return mux
}
