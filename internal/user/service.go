package user

import (
	"context"
	"fmt"

	db "github.com/ModstDev/Chatter/internal/database/sqlc"
	"github.com/google/uuid"
	"golang.org/x/crypto/bcrypt"
)

type RegisterInput struct {
	Username string
	Email    string
	Password string
}

type Service struct {
	repository Repository
}

func NewService(repository Repository) *Service {
	return &Service{
		repository: repository,
	}
}
func (s *Service) Register(ctx context.Context, input RegisterInput) (db.User, error) {
	if input.Username == "" {
		return db.User{}, fmt.Errorf("username is required")
	}
	if input.Email == "" {
		return db.User{}, fmt.Errorf("email is required")
	}
	if input.Password == "" {
		return db.User{}, fmt.Errorf("password is required")
	}

	passwordHash, err := bcrypt.GenerateFromPassword(
		[]byte(input.Password),
		bcrypt.DefaultCost,
	)
	if err != nil {
		return db.User{}, fmt.Errorf("hashing password: %w", err)
	}

	userID := uuid.New()

	err = s.repository.Create(ctx, db.CreateUserParams{
		ID:           userID.String(),
		Username:     input.Username,
		Email:        input.Email,
		PasswordHash: string(passwordHash),
	})
	if err != nil {
		return db.User{}, fmt.Errorf("creating user: %w", err)
	}

	user, err := s.repository.GetByID(ctx, userID)
	if err != nil {
		return db.User{}, fmt.Errorf("getting created user: %w", err)
	}

	return user, nil
}

func (s *Service) GetByID(ctx context.Context, id uuid.UUID) (db.User, error) {
	user, err := s.repository.GetByID(ctx, id)
	if err != nil {
		return db.User{}, fmt.Errorf("getting user: %w", err)
	}

	return user, nil
}

func (s *Service) List(ctx context.Context) ([]db.User, error) {
	return s.repository.List(ctx)
}
