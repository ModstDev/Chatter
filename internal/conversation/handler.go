package conversation

import (
	"encoding/json"
	"net/http"

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
