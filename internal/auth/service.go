package auth

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
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

type RefreshResult struct {
	AccessToken  string
	RefreshToken string
	ExpiresIn    int64
}

func (s *Service) createRefreshToken(ctx context.Context, userID uuid.UUID) (string, error) {
	token, err := generateRefreshToken()
	if err != nil {
		return "", err
	}

	tokenHash := hashRefreshToken(token)

	tokenID := uuid.New()
	expiresAt := time.Now().Add(s.refreshTokenTTL)

	err = s.refreshTokens.Create(ctx, db.CreateRefreshTokenParams{
		ID:        tokenID.String(),
		UserID:    userID.String(),
		TokenHash: tokenHash,
		ExpiresAt: expiresAt,
	})
	if err != nil {
		return "", fmt.Errorf("storing refresh token: %w", err)
	}

	return token, nil
}

func (s *Service) Refresh(ctx context.Context, refreshToken string) (RefreshResult, error) {
	if refreshToken == "" {
		return RefreshResult{}, fmt.Errorf("refresh token is required")
	}

	tokenHash := hashRefreshToken(refreshToken)

	storedToken, err := s.refreshTokens.GetByHash(ctx, tokenHash)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return RefreshResult{}, fmt.Errorf("invalid refresh token")
		}

		return RefreshResult{}, fmt.Errorf("getting refresh token: %w", err)
	}

	if storedToken.RevokedAt.Valid {
		return RefreshResult{}, fmt.Errorf("refresh token has been revoked")
	}

	if time.Now().After(storedToken.ExpiresAt) {
		return RefreshResult{}, fmt.Errorf("refresh token has expired")
	}

	userID, err := uuid.Parse(storedToken.UserID)
	if err != nil {
		return RefreshResult{}, fmt.Errorf("parse refresh token user id: %w", err)
	}

	newAccessToken, err := s.tokens.GenerateAccessToken(userID, s.accessTokenTTL)
	if err != nil {
		return RefreshResult{}, fmt.Errorf("generate access token: %w", err)
	}

	newRefreshToken, err := generateRefreshToken()
	if err != nil {
		return RefreshResult{}, fmt.Errorf("generate refresh token: %w", err)
	}

	newRefreshTokenHash := hashRefreshToken(newRefreshToken)

	newRefreshTokenID := uuid.New()
	newRefreshTokenExpiresAt := time.Now().Add(s.refreshTokenTTL)

	oldRefreshTokenID, err := uuid.Parse(storedToken.ID)
	if err != nil {
		return RefreshResult{}, fmt.Errorf("parse refresh token id: %w", err)
	}

	err = s.refreshTokens.Rotate(ctx, oldRefreshTokenID, userID, newRefreshTokenID, newRefreshTokenHash, newRefreshTokenExpiresAt)
	if err != nil {
		return RefreshResult{}, fmt.Errorf("rotate refresh token: %w", err)
	}

	return RefreshResult{
		AccessToken:  newAccessToken,
		RefreshToken: newRefreshToken,
		ExpiresIn:    int64(s.accessTokenTTL.Seconds()),
	}, nil
}
