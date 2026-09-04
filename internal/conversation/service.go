package conversation

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"time"

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

type ConverstionListItem struct {
	ID            uuid.UUID
	CreatedAt     time.Time
	OtherUserID   uuid.UUID
	OtherUsername string
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

func (s *Service) ListForUser(ctx context.Context, userID uuid.UUID) ([]ConverstionListItem, error) {
	rows, err := s.repository.ListDirectForUser(ctx, userID)
	if err != nil {
		return nil, fmt.Errorf("list conversations: %w", err)
	}

	conversations := make([]ConverstionListItem, 0, len(rows))

	for _, row := range rows {
		conversationID, err := uuid.Parse(row.ID)
		if err != nil {
			return nil, fmt.Errorf("parse conversation id: %w", err)
		}

		otherUserID, err := uuid.Parse(row.OtherUserID)
		if err != nil {
			return nil, fmt.Errorf("parse other user id: %w", err)
		}

		conversations = append(conversations, ConverstionListItem{
			ID:            conversationID,
			CreatedAt:     row.CreatedAt,
			OtherUserID:   otherUserID,
			OtherUsername: row.OtherUsername,
		})
	}

	return conversations, nil
}

func (s *Service) ListMembers(ctx context.Context, conversationID uuid.UUID) ([]uuid.UUID, error) {
	rows, err := s.repository.ListMemberIDs(
		ctx,
		conversationID,
	)
	if err != nil {
		return nil, err
	}

	memberIDs := make([]uuid.UUID, 0, len(rows))

	for _, row := range rows {
		id, err := uuid.Parse(row)
		if err != nil {
			return nil, fmt.Errorf("parse member id: %w", err)
		}

		memberIDs = append(memberIDs, id)
	}

	return memberIDs, nil
}

func normalizeUserPair(userID1 uuid.UUID, userID2 uuid.UUID) (uuid.UUID, uuid.UUID) {
	if userID1.String() < userID2.String() {
		return userID1, userID2
	}

	return userID2, userID1
}
