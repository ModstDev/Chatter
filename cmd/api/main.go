package main

import (
	"log"
	"net/http"

	"github.com/ModstDev/Chatter/internal/auth"
	"github.com/ModstDev/Chatter/internal/config"
	"github.com/ModstDev/Chatter/internal/conversation"
	"github.com/ModstDev/Chatter/internal/database"
	sqlc "github.com/ModstDev/Chatter/internal/database/sqlc"
	server "github.com/ModstDev/Chatter/internal/http"
	"github.com/ModstDev/Chatter/internal/user"
	"github.com/joho/godotenv"
)

func main() {
	if err := godotenv.Load(); err != nil {
		log.Printf("loading .env %v", err)
	}

	cfg, err := config.Load()
	if err != nil {
		log.Fatal(err)
	}

	db, err := database.Connect(cfg.Database)
	if err != nil {
		log.Fatal(err)
	}
	defer db.Close()

	queries := sqlc.New(db)

	userRepository := user.NewRepository(queries)
	userService := user.NewService(userRepository)
	userHandler := user.NewHandler(userService)

	refreshTokenRepository := auth.NewRefreshTokenRepository(db, queries)
	tokenManager := auth.NewTokenManager(cfg.Auth.JWTSecret)
	authService := auth.NewService(userRepository, refreshTokenRepository, tokenManager, cfg.Auth.AccessTokenTTL, cfg.Auth.RefreshTokenTTL)
	authHandler := auth.NewHandler(authService)

	conversationRepository := conversation.NewRepository(db, queries)
	conversationService := conversation.NewService(conversationRepository)
	conversationHandler := conversation.NewHandler(conversationService)

	router := server.NewRouter(userHandler, authHandler, conversationHandler, tokenManager)

	log.Println("server listening on :8080")

	if err := http.ListenAndServe(":8080", router); err != nil {
		log.Fatal(err)
	}

	_ = tokenManager
	_ = refreshTokenRepository
}
