package auth

import (
	"context"

	db "github.com/ModstDev/Chatter/internal/database/sqlc"
	"github.com/google/uuid"
)

type RefreshTokenRepository interface {
	Create(ctx context.Context, params db.CreateRefreshTokenParams) error
	GetByHash(ctx context.Context, tokenHash string) (db.RefreshToken, error)
	Revoke(ctx context.Context, id uuid.UUID) error
	RevokeAllForUser(ctx context.Context, userID uuid.UUID) error
}

type refreshTokenRepository struct {
	queries *db.Queries
}

func NewRefreshTokenRepository(queries *db.Queries) RefreshTokenRepository {
	return &refreshTokenRepository{
		queries: queries,
	}
}

func (r *refreshTokenRepository) Create(ctx context.Context, params db.CreateRefreshTokenParams) error {
	return r.queries.CreateRefreshToken(ctx, params)
}

func (r *refreshTokenRepository) GetByHash(ctx context.Context, tokenHash string) (db.RefreshToken, error) {
	return r.queries.GetRefreshTokenByHash(ctx, tokenHash)
}

func (r *refreshTokenRepository) Revoke(ctx context.Context, id uuid.UUID) error {
	return r.queries.RevokeRefreshToken(ctx, id.String())
}

func (r *refreshTokenRepository) RevokeAllForUser(ctx context.Context, userID uuid.UUID) error {
	return r.queries.RevokeAllUserRefreshTokens(ctx, userID.String())
}
