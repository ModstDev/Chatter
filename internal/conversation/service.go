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

	userLow, userHigh := normalizeUserPair(userID, otherUserID)

	existingID, err := s.repository.FindDirectConversation(ctx, userLow, userHigh)
	if err == nil {
		return existingID, nil
	}

	if !errors.Is(err, sql.ErrNoRows) {
		return uuid.Nil, fmt.Errorf("find direct conversation: %w", err)
	}

	conversationID, err := s.repository.CreateDirect(
		ctx,
		userLow,
		userHigh,
	)
	if err != nil {
		return uuid.Nil, fmt.Errorf(
			"create direct conversation: %w",
			err,
		)
	}

	return conversationID, nil
}

func normalizeUserPair(userID1 uuid.UUID, userID2 uuid.UUID) (uuid.UUID, uuid.UUID) {
	if userID1.String() < userID2.String() {
		return userID1, userID2
	}

	return userID2, userID1
}
