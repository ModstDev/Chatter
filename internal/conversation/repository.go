package conversation

import (
	"context"

	db "github.com/ModstDev/Chatter/internal/database/sqlc"
	"github.com/google/uuid"
)

type Repository interface {
	Create(ctx context.Context, id uuid.UUID) error
	AddMember(ctx context.Context, conversationID, userID uuid.UUID) error
	GetByID(ctx context.Context, id uuid.UUID) (db.Conversation, error)
	FindDirectConversation(ctx context.Context, userID1 uuid.UUID, userID2 uuid.UUID) (uuid.UUID, error)
}

type repository struct {
	queries *db.Queries
}

func NewRepository(queries *db.Queries) Repository {
	return &repository{
		queries: queries,
	}
}

func (r *repository) Create(ctx context.Context, id uuid.UUID) error {
	return r.queries.CreateConversation(ctx, id.String())
}

func (r *repository) AddMember(ctx context.Context, conversationID, userID uuid.UUID) error {
	return r.queries.AddConversationMember(ctx, db.AddConversationMemberParams{
		ConversationID: conversationID.String(),
		UserID:         userID.String(),
	})
}

func (r *repository) GetByID(ctx context.Context, id uuid.UUID) (db.Conversation, error) {
	return r.queries.GetConversationByID(ctx, id.String())
}

func (r *repository) FindDirectConversation(ctx context.Context, userID1, userID2 uuid.UUID) (uuid.UUID, error) {
	id, err := r.queries.FindDirectConversation(ctx, db.FindDirectConversationParams{
		UserID:   userID1.String(),
		UserID_2: userID2.String(),
	})
	if err != nil {
		return uuid.Nil, err
	}

	return uuid.Parse(id)
}
