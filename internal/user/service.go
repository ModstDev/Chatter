package user

import (
	"context"
	"fmt"

	db "github.com/ModstDev/Chatter/internal/database/sqlc"
	"github.com/google/uuid"
)

type Service struct {
	repository Repository
}

func NewService(repository Repository) *Service {
	return &Service{
		repository: repository,
	}
}

func (s *Service) GetByID(ctx context.Context, id uuid.UUID) (db.User, error) {
	user, err := s.repository.GetByID(ctx, id)
	if err != nil {
		return db.User{}, fmt.Errorf("getting user: %w", err)
	}

	return user, nil
}
