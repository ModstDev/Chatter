package main

import (
	"log"

	"github.com/ModstDev/Chatter/internal/config"
	server "github.com/ModstDev/Chatter/internal/server"
	"github.com/joho/godotenv"
)

func main() {
	if err := godotenv.Load(); err != nil {
		log.Printf("loading .env: %v", err)
	}

	cfg, err := config.Load()
	if err != nil {
		log.Fatal(err)
	}

	app, err := server.New(&cfg)
	if err != nil {
		log.Fatal(err)
	}

	log.Println("server listening on :8080")

	if err := app.Run(); err != nil {
		log.Fatal(err)
	}
}
