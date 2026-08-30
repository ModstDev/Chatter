package user

import (
	"context"

	db "github.com/ModstDev/Chatter/internal/database/sqlc"
	"github.com/google/uuid"
)

type Repository interface {
	Create(ctx context.Context, user db.User) error
	GetByID(ctx context.Context, id uuid.UUID) (db.User, error)
	GetByEmail(ctx context.Context, email string) (db.User, error)
	GetByUsername(ctx context.Context, username string) (db.User, error)
}
