package message

import (
	"encoding/json"
	"net/http"
	"strconv"

	"github.com/ModstDev/Chatter/internal/auth"
	"github.com/google/uuid"
)

type Handler struct {
	service *Service
}

func NewHandler(service *Service) *Handler {
	return &Handler{
		service: service,
	}
}

type messageResponse struct {
	ID        uuid.UUID `json:"id"`
	SenderID  uuid.UUID `json:"sender_id"`
	Content   string    `json:"content"`
	CreatedAt string    `json:"created_at"`
}

type messagePageResponse struct {
	Messages   []messageResponse `json:"messages"`
	NextCursor string            `json:"next_cursor"`
	HasMore    bool              `json:"has_more"`
}

func (h *Handler) ListHistory(w http.ResponseWriter, r *http.Request) {
	userID, ok := auth.UserIDFromContext(r.Context())
	if !ok {
		http.Error(w, "unauthorized", http.StatusUnauthorized)
		return
	}

	conversationID, err := uuid.Parse(r.PathValue("conversationID"))
	if err != nil {
		http.Error(w, "invalid conversation ID", http.StatusBadRequest)
		return
	}

	limit := 50

	if value := r.URL.Query().Get("limit"); value != "" {
		limit, err = strconv.Atoi(value)
		if err != nil {
			http.Error(w, "invalid limit", http.StatusBadRequest)
			return
		}
	}

	before := r.URL.Query().Get("before")

	page, err := h.service.ListHistory(r.Context(), userID, conversationID, limit, before)
	if err != nil {
		http.Error(w, "failed to get messages", http.StatusInternalServerError)
		return
	}

	response := messagePageResponse{
		Messages:   make([]messageResponse, 0, len(page.Messages)),
		NextCursor: page.NextCursor,
		HasMore:    page.HasMore,
	}

	for _, msg := range page.Messages {
		response.Messages = append(response.Messages, messageResponse{
			ID:        msg.ID,
			SenderID:  msg.SenderID,
			Content:   msg.Content,
			CreatedAt: msg.CreatedAt.Format("2006-01-02T15:04:05.999999Z07:00"),
		})
	}

	w.Header().Set("Content-Type", "application/json")

	if err := json.NewEncoder(w).Encode(response); err != nil {
		return
	}
}
