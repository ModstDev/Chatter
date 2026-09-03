package conversation

import (
	"context"
	"database/sql"
	"errors"
	"fmt"

	db "github.com/ModstDev/Chatter/internal/database/sqlc"
	"github.com/go-sql-driver/mysql"
	"github.com/google/uuid"
)

type Repository interface {
	CreateDirect(ctx context.Context, userLow uuid.UUID, userHigh uuid.UUID) (uuid.UUID, error)
	AddMember(ctx context.Context, conversationID, userID uuid.UUID) error
	GetByID(ctx context.Context, id uuid.UUID) (db.GetConversationByIDRow, error)
	FindDirectConversation(ctx context.Context, userLow uuid.UUID, userHigh uuid.UUID) (uuid.UUID, error)
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

func (r *repository) CreateDirect(ctx context.Context, userLow uuid.UUID, userHigh uuid.UUID) (uuid.UUID, error) {
	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return uuid.Nil, fmt.Errorf("begin conversatio ntransaction: %w", err)
	}

	txQueries := r.queries.WithTx(tx)

	conversationID := uuid.New()

	err = txQueries.CreateConversation(ctx, db.CreateConversationParams{
		ID:       conversationID.String(),
		UserLow:  sql.NullString{String: userLow.String(), Valid: true},
		UserHigh: sql.NullString{String: userHigh.String(), Valid: true},
	})
	if err != nil {
		_ = tx.Rollback()

		if isDuplicateKeyError(err) {
			return uuid.Nil, errors.New("direct conversation already exists")
		}

		return uuid.Nil, fmt.Errorf("create direct conversation: %w", err)
	}

	if err := txQueries.AddConversationMember(ctx, db.AddConversationMemberParams{
		ConversationID: conversationID.String(),
		UserID:         userLow.String(),
	}); err != nil {
		_ = tx.Rollback()

		if isDuplicateKeyError(err) {
			return uuid.Nil, errors.New("direct conversation already exists")
		}

		return uuid.Nil, fmt.Errorf("add first conversation member %w", err)
	}

	if err := txQueries.AddConversationMember(ctx, db.AddConversationMemberParams{
		ConversationID: conversationID.String(),
		UserID:         userHigh.String(),
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

func (r *repository) GetByID(ctx context.Context, id uuid.UUID) (db.GetConversationByIDRow, error) {
	return r.queries.GetConversationByID(ctx, id.String())
}

func (r *repository) FindDirectConversation(ctx context.Context, userLow uuid.UUID, userHigh uuid.UUID) (uuid.UUID, error) {
	id, err := r.queries.FindDirectConversation(ctx, db.FindDirectConversationParams{
		UserLow:  sql.NullString{String: userLow.String(), Valid: true},
		UserHigh: sql.NullString{String: userHigh.String(), Valid: true},
	})
	if err != nil {
		return uuid.Nil, err
	}

	return uuid.Parse(id)
}

func isDuplicateKeyError(err error) bool {
	var mysqlErr *mysql.MySQLError

	return errors.As(err, &mysqlErr) &&
		mysqlErr.Number == 1062
}
