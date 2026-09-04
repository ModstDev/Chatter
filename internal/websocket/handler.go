package websocket

import (
	"context"
	"errors"
	"net/http"
	"strings"

	"github.com/ModstDev/Chatter/internal/auth"
	"github.com/coder/websocket"
	"github.com/google/uuid"
)

type Handler struct {
	tokenManager *auth.TokenManager
	hub          *Hub
}

func NewHandler(tokenManager *auth.TokenManager, hub *Hub) *Handler {
	return &Handler{
		tokenManager: tokenManager,
		hub:          hub,
	}
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
		_, _, err := client.conn.Read(ctx)
		if err != nil {
			return
		}
	}
}
