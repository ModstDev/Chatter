package config

import (
	"fmt"
	"os"
)

type Config struct {
	Database DatabaseConfig
	Auth     AuthConfig
}

type DatabaseConfig struct {
	Host     string
	Port     string
	User     string
	Password string
	Name     string
}

type AuthConfig struct {
	JWTSecret       string
	AccessTokenTTL  string
	RefreshTokenTTL string
}

func Load() (Config, error) {
	database, err := loadDatabaseConfig()
	if err != nil {
		return Config{}, err
	}

	return Config{
		Database: database,
	}, nil
}

func loadDatabaseConfig() (DatabaseConfig, error) {
	config := DatabaseConfig{
		Host:     os.Getenv("DB_HOST"),
		Port:     os.Getenv("DB_PORT"),
		User:     os.Getenv("DB_USER"),
		Password: os.Getenv("DB_PASSWORD"),
		Name:     os.Getenv("DB_NAME"),
	}

	if config.Host == "" {
		return DatabaseConfig{}, fmt.Errorf("DB_HOST is required")
	}

	if config.Port == "" {
		return DatabaseConfig{}, fmt.Errorf("DB_PORT is required")
	}

	if config.User == "" {
		return DatabaseConfig{}, fmt.Errorf("DB_USER is required")
	}

	if config.Password == "" {
		return DatabaseConfig{}, fmt.Errorf("DB_PASSWORD is required")
	}

	if config.Name == "" {
		return DatabaseConfig{}, fmt.Errorf("DB_NAME is required")
	}

	return config, nil
}
