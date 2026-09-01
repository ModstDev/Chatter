package config

import (
	"fmt"
	"os"
	"time"
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
	AccessTokenTTL  time.Duration
	RefreshTokenTTL time.Duration
}

func Load() (Config, error) {
	database, err := loadDatabaseConfig()
	if err != nil {
		return Config{}, err
	}

	auth, err := loadAuthConfig()
	if err != nil {
		return Config{}, err
	}

	return Config{
		Database: database,
		Auth:     auth,
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

func loadAuthConfig() (AuthConfig, error) {
	jwtSecret := os.Getenv("JWT_SECRET")
	if jwtSecret == "" {
		return AuthConfig{}, fmt.Errorf("JWT_SECRET is required")
	}

	accessTokenTTL, err := time.ParseDuration(os.Getenv("ACCESS_TOKEN_TTL"))
	if err != nil {
		return AuthConfig{}, fmt.Errorf("parsing ACCESS_TOKEN_TTL: %w", err)
	}

	if accessTokenTTL <= 0 {
		return AuthConfig{}, fmt.Errorf("ACCESS_TOKEN_TTL must be greater than zero")
	}

	refreshTokenTTL, err := time.ParseDuration(os.Getenv("REFRESH_TOKEN_TTL"))
	if err != nil {
		return AuthConfig{}, fmt.Errorf("parsing REFRESH_TOKEN_TTL: %w", err)
	}

	if refreshTokenTTL <= 0 {
		return AuthConfig{}, fmt.Errorf("REFRESH_TOKEN_TTL must be greater than zero")
	}

	return AuthConfig{
		JWTSecret:       jwtSecret,
		AccessTokenTTL:  accessTokenTTL,
		RefreshTokenTTL: refreshTokenTTL,
	}, nil
}
