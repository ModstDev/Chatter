package conversation

import (
	"context"
	"database/sql"
	"errors"
	"fmt"

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

func (s *Service) CreateDirect(ctx context.Context, userID uuid.UUID, otherUserID uuid.UUID) (uuid.UUID, error) {
	if userID == uuid.Nil {
		return uuid.Nil, fmt.Errorf("invalid user id")
	}

	if otherUserID == uuid.Nil {
		return uuid.Nil, fmt.Errorf("invalid other user id")
	}

	if userID == otherUserID {
		return uuid.Nil, fmt.Errorf("cannot create conversation with yourself")
	}

	existingID, err := s.repository.FindDirectConversation(ctx, userID, otherUserID)
	if err == nil {
		return existingID, nil
	}

	if !errors.Is(err, sql.ErrNoRows) {
		return uuid.Nil, fmt.Errorf("find direct conversation: %w", err)
	}

	conversationID := uuid.New()

	if err := s.repository.Create(ctx, conversationID); err != nil {
		return uuid.Nil, fmt.Errorf("create conversation: %w", err)
	}

	if err := s.repository.AddMember(ctx, conversationID, userID); err != nil {
		return uuid.Nil, fmt.Errorf("add first conversation member: %w", err)
	}

	if err := s.repository.AddMember(ctx, conversationID, otherUserID); err != nil {
		return uuid.Nil, fmt.Errorf("add second conversation member: %w", err)
	}

	return conversationID, nil
}
