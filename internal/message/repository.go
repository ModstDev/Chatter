package message

import (
	"context"
	"time"

	db "github.com/ModstDev/Chatter/internal/database/sqlc"
	"github.com/google/uuid"
)

type Repository interface {
	List(
		ctx context.Context,
		conversatition uuid.UUID,
		limit int32,
	) ([]db.Message, error)

	ListBefore(
		ctx context.Context,
		conversationID uuid.UUID,
		createdAt time.Time,
		messageID uuid.UUID,
		limit int32,
	) ([]db.Message, error)

	ListAfter(
		ctx context.Context,
		conversationID uuid.UUID,
		createdAt time.Time,
		messageID uuid.UUID,
		limit int32,
	) ([]db.Message, error)

	Create(ctx context.Context, params db.CreateMessageParams) error
	GetByID(ctx context.Context, id uuid.UUID) (db.Message, error)
}

type repository struct {
	queries *db.Queries
}

func NewRepository(queries *db.Queries) Repository {
	return &repository{
		queries: queries,
	}
}

func (r *repository) List(
	ctx context.Context,
	conversationID uuid.UUID,
	limit int32,
) ([]db.Message, error) {
	return r.queries.ListMessages(ctx, db.ListMessagesParams{
		ConversationID: conversationID.String(),
		Limit:          limit,
	})
}

func (r *repository) ListBefore(
	ctx context.Context,
	conversationID uuid.UUID,
	createdAt time.Time,
	messageID uuid.UUID,
	limit int32,
) ([]db.Message, error) {
	return r.queries.ListMessagesBefore(ctx, db.ListMessagesBeforeParams{
		ConversationID: conversationID.String(),
		CreatedAt:      createdAt,
		CreatedAt_2:    createdAt,
		ID:             messageID.String(),
		Limit:          limit,
	})
}

func (r *repository) ListAfter(
	ctx context.Context,
	conversationID uuid.UUID,
	createdAt time.Time,
	messageID uuid.UUID,
	limit int32,
) ([]db.Message, error) {
	return r.queries.ListMessagesAfter(
		ctx,
		db.ListMessagesAfterParams{
			ConversationID: conversationID.String(),
			CreatedAt:      createdAt,
			CreatedAt_2:    createdAt,
			ID:             messageID.String(),
			Limit:          limit,
		},
	)
}

func (r *repository) Create(ctx context.Context, params db.CreateMessageParams) error {
	return r.queries.CreateMessage(ctx, params)
}

func (r *repository) GetByID(ctx context.Context, id uuid.UUID) (db.Message, error) {
	return r.queries.GetMessageByID(ctx, id.String())
}
