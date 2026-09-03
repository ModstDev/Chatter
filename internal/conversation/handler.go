package conversation

import (
	"encoding/json"
	"net/http"
	"time"

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

type createConversationRequest struct {
	UserID string `json:"user_id"`
}

type conversationResponse struct {
	ID string `json:"id"`
}

type conversationListItemResponse struct {
	ID            string    `json:"id"`
	OtherUserID   string    `json:"other_user_id"`
	OtherUsername string    `json:"other_username"`
	CreatedAt     time.Time `json:"created_at"`
}

func (h *Handler) Create(w http.ResponseWriter, r *http.Request) {
	var req createConversationRequest

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "invalid request body", http.StatusBadRequest)
		return
	}

	otherUserID, err := uuid.Parse(req.UserID)
	if err != nil {
		http.Error(w, "invalid user_id", http.StatusBadRequest)
		return
	}

	userID, ok := auth.UserIDFromContext(r.Context())
	if !ok {
		http.Error(w, "unauthorized", http.StatusUnauthorized)
		return
	}

	conversationID, err := h.service.CreateDirect(r.Context(), userID, otherUserID)
	if err != nil {
		http.Error(w, "failed to create conversation", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)

	_ = json.NewEncoder(w).Encode(conversationResponse{
		ID: conversationID.String(),
	})
}

func (h *Handler) List(w http.ResponseWriter, r *http.Request) {
	userID, ok := auth.UserIDFromContext(r.Context())
	if !ok {
		http.Error(w, "unauthorized", http.StatusUnauthorized)
		return
	}

	conversations, err := h.service.ListForUser(r.Context(), userID)
	if err != nil {
		http.Error(w, "failed to get conversations", http.StatusInternalServerError)
		return
	}

	response := make([]conversationListItemResponse, 0, len(conversations))

	for _, conversation := range conversations {
		response = append(response, conversationListItemResponse{
			ID:            conversation.ID.String(),
			OtherUserID:   conversation.OtherUserID.String(),
			OtherUsername: conversation.OtherUsername,
			CreatedAt:     conversation.CreatedAt,
		})
	}

	w.Header().Set("Content-Type", "application/json")

	if err := json.NewEncoder(w).Encode(response); err != nil {
		return
	}
}
