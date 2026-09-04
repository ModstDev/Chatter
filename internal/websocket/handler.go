package websocket

import (
	"context"
	"encoding/json"
	"errors"
	"log"
	"net/http"
	"strings"

	"github.com/ModstDev/Chatter/internal/auth"
	"github.com/ModstDev/Chatter/internal/conversation"
	"github.com/ModstDev/Chatter/internal/message"
	"github.com/coder/websocket"
	"github.com/google/uuid"
)

type Handler struct {
	tokenManager  *auth.TokenManager
	hub           *Hub
	messages      *message.Service
	conversations *conversation.Service
}

func NewHandler(tokenManager *auth.TokenManager, hub *Hub, messages *message.Service, conversations *conversation.Service) *Handler {
	return &Handler{
		tokenManager:  tokenManager,
		hub:           hub,
		messages:      messages,
		conversations: conversations,
	}
}

type messageRequest struct {
	Type           string `json:"type"`
	ConversationID string `json:"conversation_id"`
	Content        string `json:"content"`
}

type messageEvent struct {
	Type    string           `json:"type"`
	Message *message.Message `json:"message"`
}

func (h *Handler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	userID, err := h.authenticate(r)
	if err != nil {
		http.Error(w, "unauthorized", http.StatusUnauthorized)
		return
	}

	conn, err := websocket.Accept(w, r, nil)
	if err != nil {
		return
	}

	client := NewClient(conn, userID)

	h.hub.Register(client)
	defer h.hub.Unregister(client)

	ctx := r.Context()

	go client.writeLoop(ctx)

	h.readLoop(ctx, client)
}

func (h *Handler) authenticate(r *http.Request) (uuid.UUID, error) {
	header := r.Header.Get("Authorization")

	const prefix = "Bearer "

	if !strings.HasPrefix(header, prefix) {
		return uuid.Nil, errors.New("unauthorized")
	}

	token := strings.TrimSpace(strings.TrimPrefix(header, prefix))

	if token == "" {
		return uuid.Nil, errors.New("unauthorized")
	}

	return h.tokenManager.ValidateAccessToken(token)
}

func (h *Handler) readLoop(ctx context.Context, client *Client) {
	for {
		_, data, err := client.conn.Read(ctx)
		if err != nil {
			return
		}

		var req messageRequest
		if err := json.Unmarshal(data, &req); err != nil {
			continue
		}

		if req.Type != "message" {
			continue
		}

		conversationID, err := uuid.Parse(req.ConversationID)
		if err != nil {
			continue
		}

		msg, err := h.messages.Send(ctx, client.userID, conversationID, req.Content)
		if err != nil {
			continue
		}

		event, err := json.Marshal(messageEvent{
			Type:    "messaage",
			Message: msg,
		})
		if err != nil {
			continue
		}

		memberIDs, err := h.conversations.ListMembers(ctx, conversationID)
		if err != nil {
			log.Printf("list conversation members: %v", err)
			continue
		}

		for _, memberID := range memberIDs {
			h.hub.SendToUser(memberID, event)
		}
	}
}
