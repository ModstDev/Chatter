package main

import (
	"log"

	"github.com/ModstDev/Chatter/internal/config"
	"github.com/ModstDev/Chatter/internal/database"
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

	log.Printf("database conncetion established")
}
