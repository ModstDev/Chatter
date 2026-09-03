package conversation

import (
	"context"
	"database/sql"
	"fmt"

	db "github.com/ModstDev/Chatter/internal/database/sqlc"
	"github.com/google/uuid"
)

type Repository interface {
	Create(ctx context.Context, userID1 uuid.UUID, userID2 uuid.UUID) (uuid.UUID, error)
	AddMember(ctx context.Context, conversationID, userID uuid.UUID) error
	GetByID(ctx context.Context, id uuid.UUID) (db.Conversation, error)
	FindDirectConversation(ctx context.Context, userID1 uuid.UUID, userID2 uuid.UUID) (uuid.UUID, error)
}

type repository struct {
	db      *sql.DB
	queries *db.Queries
}

func NewRepository(db *sql.DB, queries *db.Queries) Repository {
	return &repository{
		db:      db,
		queries: queries,
	}
}

func (r *repository) Create(ctx context.Context, userID1 uuid.UUID, userID2 uuid.UUID) (uuid.UUID, error) {
	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return uuid.Nil, fmt.Errorf("begin conversatio ntransaction: %w", err)
	}

	txQueries := r.queries.WithTx(tx)

	conversationID := uuid.New()

	if err := txQueries.CreateConversation(ctx, conversationID.String()); err != nil {
		_ = tx.Rollback()

		return uuid.Nil, fmt.Errorf("create conversation: %w", err)
	}

	if err := txQueries.AddConversationMember(ctx, db.AddConversationMemberParams{
		ConversationID: conversationID.String(),
		UserID:         userID1.String(),
	}); err != nil {
		_ = tx.Rollback()

		return uuid.Nil, fmt.Errorf("add first conversation member %w", err)
	}

	if err := txQueries.AddConversationMember(ctx, db.AddConversationMemberParams{
		ConversationID: conversationID.String(),
		UserID:         userID2.String(),
	}); err != nil {
		_ = tx.Rollback()

		return uuid.Nil, fmt.Errorf("add second conversation member: %w", err)
	}

	if err := tx.Commit(); err != nil {
		return uuid.Nil, fmt.Errorf("commit conversation transaction: %w", err)
	}

	return conversationID, nil
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
