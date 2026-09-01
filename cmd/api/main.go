package main

import (
	"log"
	"net/http"

	"github.com/ModstDev/Chatter/internal/auth"
	"github.com/ModstDev/Chatter/internal/config"
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

	tokenManager := auth.NewTokenManager(cfg.Auth.JWTSecret)

	userRepository := user.NewRepository(queries)

	userService := user.NewService(userRepository)

	userHandler := user.NewHandler(userService)

	router := server.NewRouter(userHandler)

	log.Println("server listening on :8080")

	if err := http.ListenAndServe(":8080", router); err != nil {
		log.Fatal(err)
	}

	_ = tokenManager
}
