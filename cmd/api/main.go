package main

import (
	"log"

	"github.com/ModstDev/Chatter/internal/config"
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

	log.Printf("database configured for %s:%s", cfg.Database.Host, cfg.Database.Port)
}
