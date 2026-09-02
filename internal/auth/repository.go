package auth

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	db "github.com/ModstDev/Chatter/internal/database/sqlc"
	"github.com/google/uuid"
)

type RefreshTokenRepository interface {
	Create(ctx context.Context, params db.CreateRefreshTokenParams) error
	GetByHash(ctx context.Context, tokenHash string) (db.RefreshToken, error)
	Revoke(ctx context.Context, id uuid.UUID) error
	RevokeAllForUser(ctx context.Context, userID uuid.UUID) error
	Rotate(ctx context.Context, oldTokenID uuid.UUID, userID uuid.UUID, newTokenID uuid.UUID, newTokenHash string, expiresAt time.Time) error
}

type refreshTokenRepository struct {
	db      *sql.DB
	queries *db.Queries
}

func NewRefreshTokenRepository(db *sql.DB, queries *db.Queries) RefreshTokenRepository {
	return &refreshTokenRepository{
		db:      db,
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

func (r *refreshTokenRepository) Rotate(
	ctx context.Context,
	oldTokenID uuid.UUID,
	userID uuid.UUID,
	newTokenID uuid.UUID,
	newTokenHash string,
	expiresAt time.Time,
) error {
	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return fmt.Errorf("begin refresh token rotation transaction: %w", err)
	}

	txQueries := r.queries.WithTx(tx)

	result, err := txQueries.RevokeRefreshTokenIfActive(ctx, oldTokenID.String())
	if err != nil {
		_ = tx.Rollback()
		return fmt.Errorf("revoke old refresh token: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		_ = tx.Rollback()
		return fmt.Errorf("get affected rows: %w", err)
	}

	if rowsAffected != 1 {
		_ = tx.Rollback()
		return fmt.Errorf("refresh token is no longer active")
	}

	err = txQueries.CreateRefreshToken(ctx, db.CreateRefreshTokenParams{
		ID:        newTokenID.String(),
		UserID:    userID.String(),
		TokenHash: newTokenHash,
		ExpiresAt: expiresAt,
	})
	if err != nil {
		_ = tx.Rollback()
		return fmt.Errorf("create new refresh token: %w", err)
	}

	if err := tx.Commit(); err != nil {
		return fmt.Errorf("commit refresh token rotation transaction: %w", err)
	}

	return nil
}
