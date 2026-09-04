package auth

import (
	"fmt"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
)

const tokenIssuer = "chatter"

type TokenManager struct {
	secret string
}

func NewTokenManager(secret string) *TokenManager {
	return &TokenManager{
		secret: secret,
	}
}

func (tm *TokenManager) GenerateAccessToken(userID uuid.UUID, duration time.Duration) (string, error) {
	now := time.Now()

	claims := jwt.RegisteredClaims{
		Issuer:    tokenIssuer,
		Subject:   userID.String(),
		IssuedAt:  jwt.NewNumericDate(now),
		ExpiresAt: jwt.NewNumericDate(now.Add(duration)),
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)

	signed, err := token.SignedString([]byte(tm.secret))
	if err != nil {
		return "", fmt.Errorf("signing access token: %w", err)
	}

	return signed, nil
}

func (tm *TokenManager) ValidateAccessToken(tokenString string) (uuid.UUID, error) {
	token, err := jwt.ParseWithClaims(tokenString, &jwt.RegisteredClaims{}, func(token *jwt.Token) (interface{}, error) {
		return []byte(tm.secret), nil
	})
	if err != nil {
		return uuid.Nil, fmt.Errorf("parsing access token: %w", err)
	}

	claims, ok := token.Claims.(*jwt.RegisteredClaims)
	if !ok || !token.Valid {
		return uuid.Nil, fmt.Errorf("invalid access token")
	}

	userID, err := uuid.Parse(claims.Subject)
	if err != nil {
		return uuid.Nil, fmt.Errorf("invalid user ID in access token: %w", err)
	}

	return userID, nil
}
