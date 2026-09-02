package user

import (
	"context"

	db "github.com/ModstDev/Chatter/internal/database/sqlc"
	"github.com/google/uuid"
)

type Repository interface {
	Create(ctx context.Context, params db.CreateUserParams) error
	GetByID(ctx context.Context, id uuid.UUID) (db.User, error)
	GetByEmail(ctx context.Context, email string) (db.User, error)
	GetByUsername(ctx context.Context, username string) (db.User, error)
	List(ctx context.Context) ([]db.User, error)
}

type repository struct {
	queries *db.Queries
}

func NewRepository(queries *db.Queries) Repository {
	return &repository{
		queries: queries,
	}
}

func (r *repository) Create(ctx context.Context, params db.CreateUserParams) error {
	return r.queries.CreateUser(ctx, params)
}

func (r *repository) GetByID(ctx context.Context, id uuid.UUID) (db.User, error) {
	return r.queries.GetUserByID(ctx, id.String())
}

func (r *repository) GetByEmail(ctx context.Context, email string) (db.User, error) {
	return r.queries.GetUserByEmail(ctx, email)
}

func (r *repository) GetByUsername(ctx context.Context, username string) (db.User, error) {
	return r.queries.GetUserByUsername(ctx, username)
}

func (r *repository) List(ctx context.Context) ([]db.User, error) {
	return r.queries.ListUsers(ctx)
}
