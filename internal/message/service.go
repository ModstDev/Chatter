package message

import (
	"context"
	"fmt"
	"time"

	"github.com/ModstDev/Chatter/internal/conversation"
	"github.com/google/uuid"
)

const (
	defaultMessageLimit = 50
	maxMessageLimit     = 100
)

type Service struct {
	messages      Repository
	conversations conversation.Repository
}

func NewService(messages Repository, conversations conversation.Repository) *Service {
	return &Service{
		messages:      messages,
		conversations: conversations,
	}
}

type Message struct {
	ID             uuid.UUID
	ConversationID uuid.UUID
	SenderID       uuid.UUID
	Content        string
	CreatedAt      time.Time
}

type MessagePage struct {
	Messages   []Message
	NextCursor string
	HasMore    bool
}

func (s *Service) ListHistory(
	ctx context.Context,
	userID uuid.UUID,
	conversationID uuid.UUID,
	limit int,
) (MessagePage, error) {
	if userID == uuid.Nil {
		return MessagePage{}, fmt.Errorf("invalid user id")
	}

	if conversationID == uuid.Nil {
		return MessagePage{}, fmt.Errorf("invalid conversation id")
	}

	if limit <= 0 {
		limit = defaultMessageLimit
	}

	if limit > maxMessageLimit {
		limit = maxMessageLimit
	}

	isMember, err := s.conversations.IsMember(ctx, conversationID, userID)
	if err != nil {
		return MessagePage{}, fmt.Errorf("check conversation membership: %w", err)
	}

	if !isMember {
		return MessagePage{}, fmt.Errorf("user is not a conversation member")
	}

	rows, err := s.messages.List(ctx, conversationID, int32(limit+1))
	if err != nil {
		return MessagePage{}, fmt.Errorf("list messages: %w", err)
	}

	hasMore := len(rows) > limit

	if hasMore {
		rows = rows[:limit]
	}

	messages := make([]Message, 0, len(rows))

	for _, row := range rows {
		id, err := uuid.Parse(row.ID)
		if err != nil {
			return MessagePage{}, fmt.Errorf("parse message id: %w", err)
		}

		senderID, err := uuid.Parse(row.SenderID)
		if err != nil {
			return MessagePage{}, fmt.Errorf("parse sender id: %w", err)
		}

		messages = append(messages, Message{
			ID:             id,
			ConversationID: conversationID,
			SenderID:       senderID,
			Content:        row.Content,
			CreatedAt:      row.CreatedAt,
		})
	}

	return MessagePage{
		Messages: messages,
		HasMore:  hasMore,
	}, nil
}
