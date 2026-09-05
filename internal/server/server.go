package server

import (
	"database/sql"
	"net/http"

	"github.com/ModstDev/Chatter/internal/auth"
	"github.com/ModstDev/Chatter/internal/config"
	"github.com/ModstDev/Chatter/internal/conversation"
	"github.com/ModstDev/Chatter/internal/database"
	sqlc "github.com/ModstDev/Chatter/internal/database/sqlc"
	httpserver "github.com/ModstDev/Chatter/internal/http"
	"github.com/ModstDev/Chatter/internal/message"
	"github.com/ModstDev/Chatter/internal/user"
	"github.com/ModstDev/Chatter/internal/websocket"
)

type Server struct {
	httpServer *http.Server
	db         *sql.DB
}

func New(cfg *config.Config) (*Server, error) {
	db, err := database.Connect(cfg.Database)
	if err != nil {
		return nil, err
	}

	queries := sqlc.New(db)

	userRepository := user.NewRepository(queries)
	userService := user.NewService(userRepository)
	userHandler := user.NewHandler(userService)

	refreshTokenRepository := auth.NewRefreshTokenRepository(db, queries)
	tokenManager := auth.NewTokenManager(cfg.Auth.JWTSecret)

	authService := auth.NewService(
		userRepository,
		refreshTokenRepository,
		tokenManager,
		cfg.Auth.AccessTokenTTL,
		cfg.Auth.RefreshTokenTTL,
	)
	authHandler := auth.NewHandler(authService)

	conversationRepository := conversation.NewRepository(db, queries)
	conversationService := conversation.NewService(conversationRepository)
	conversationHandler := conversation.NewHandler(conversationService)

	messageRepository := message.NewRepository(queries)
	messageService := message.NewService(
		messageRepository,
		conversationRepository,
	)
	messageHandler := message.NewHandler(messageService)

	wsHub := websocket.NewHub()
	wsHandler := websocket.NewHandler(
		tokenManager,
		wsHub,
		messageService,
		conversationService,
	)

	router := httpserver.NewRouter(
		userHandler,
		authHandler,
		conversationHandler,
		messageHandler,
		wsHandler,
		tokenManager,
	)

	return &Server{
		httpServer: &http.Server{
			Addr:    ":8080",
			Handler: router,
		},
		db: db,
	}, nil
}

func (s *Server) Run() error {
	defer s.db.Close()

	return s.httpServer.ListenAndServe()
}
