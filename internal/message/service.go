package message

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"time"

	"github.com/ModstDev/Chatter/internal/conversation"
	db "github.com/ModstDev/Chatter/internal/database/sqlc"
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

type messageCursor struct {
	CreatedAt time.Time `json:"created_at"`
	MessageID uuid.UUID `json:"message_id"`
}

func (s *Service) ListHistory(
	ctx context.Context,
	userID uuid.UUID,
	conversationID uuid.UUID,
	limit int,
	before string,
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

	var rows []db.Message

	queryLimit := int32(limit + 1)

	if before == "" {
		rows, err = s.messages.List(ctx, conversationID, queryLimit)
	} else {
		cursor, err := decodeCursor(before)
		if err != nil {
			return MessagePage{}, err
		}
		rows, err = s.messages.ListBefore(ctx, conversationID, cursor.CreatedAt, cursor.MessageID, queryLimit)
	}
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

	var nextCursor string

	if hasMore && len(messages) > 0 {
		last := messages[len(messages)-1]

		cursor := messageCursor{
			CreatedAt: last.CreatedAt,
			MessageID: last.ID,
		}

		encoded, err := encodeCursor(cursor)
		if err != nil {
			return MessagePage{}, fmt.Errorf("encode next cursor: %w", err)
		}

		nextCursor = encoded
	}

	return MessagePage{
		Messages:   messages,
		NextCursor: nextCursor,
		HasMore:    hasMore,
	}, nil
}

func encodeCursor(cursor messageCursor) (string, error) {
	data, err := json.Marshal(cursor)
	if err != nil {
		return "", err
	}

	return base64.RawStdEncoding.EncodeToString(data), nil
}

func decodeCursor(value string) (messageCursor, error) {
	data, err := base64.RawURLEncoding.DecodeString(value)
	if err != nil {
		return messageCursor{}, fmt.Errorf("invalid message cursor: %w", err)
	}

	var cursor messageCursor

	if err := json.Unmarshal(data, &cursor); err != nil {
		return messageCursor{}, fmt.Errorf("invalid message cursor: %w", err)
	}

	if cursor.MessageID == uuid.Nil || cursor.CreatedAt.IsZero() {
		return messageCursor{}, fmt.Errorf("invalid message cursor: %w", err)
	}

	return cursor, nil
}
