package auth

import (
	"context"
	"time"

	db "github.com/ModstDev/Chatter/internal/database/sqlc"
	"github.com/ModstDev/Chatter/internal/user"
	"github.com/google/uuid"
)

type Service struct {
	users           user.Repository
	refreshTokens   RefreshTokenRepository
	tokens          *TokenManager
	accessTokenTTL  time.Duration
	refreshTokenTTL time.Duration
}

func NewService(
	users user.Repository,
	refreshTokens RefreshTokenRepository,
	tokens *TokenManager,
	accessTokenTTL time.Duration,
	refreshTokenTTL time.Duration,
) *Service {
	return &Service{
		users:           users,
		refreshTokens:   refreshTokens,
		tokens:          tokens,
		accessTokenTTL:  accessTokenTTL,
		refreshTokenTTL: refreshTokenTTL,
	}
}

func (s *Service) createRefreshToken(ctx context.Context, userID uuid.UUID) (string, error) {
	token, err := generateRefreshToken(token)
	if err != nil {
		return "", err
	}

	tokenHash := hashRefreshToken(token)

	tokenID := uuid.New()
	expiresAt := time.Now().Add(s.refreshTokenTTL)

	err = s.refreshTokens.Create(ctx, db.CreateRefreshTokenParams{
		ID:        tokenID,
		UserID:    userID,
		TokenHash: tokenHash,
		ExpiresAt: expiresAt,
	})

	return token, nil
}
