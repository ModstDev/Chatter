package websocket

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
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
	After          string `json:"after"`
}

type messageEvent struct {
	Type    string           `json:"type"`
	Message *message.Message `json:"message"`
}

type errorEvent struct {
	Type  string `json:"type"`
	Error string `json:"error"`
}

type syncEvent struct {
	Type     string
	Messages []message.Message
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
	defer conn.Close(websocket.StatusNormalClosure, "")

	conn.SetReadLimit(64 * 1024)

	client := NewClient(conn, userID)

	h.hub.Register(client)
	defer h.hub.Unregister(client)

	ctx, cancel := context.WithCancel(r.Context())
	defer cancel()

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
			h.sendError(client, fmt.Errorf("invalid message format"))
			continue
		}

		switch req.Type {
		case "message":
			h.handleMessage(ctx, client, req)

		case "sync":
			h.handleSync(ctx, client, req)

		default:
			h.sendError(client, fmt.Errorf("unsupported message type"))
		}
	}
}

func (h *Handler) sendError(client *Client, err error) {
	event, marshalErr := json.Marshal(errorEvent{
		Type:  "error",
		Error: err.Error(),
	})
	if marshalErr != nil {
		log.Printf("encode websocket error: %v", err)
		return
	}

	select {
	case client.send <- event:
	default:
		log.Printf("client send buffer is full")
	}
}

func (h *Handler) handleMessage(
	ctx context.Context,
	client *Client,
	req messageRequest,
) {
	conversationID, err := uuid.Parse(req.ConversationID)
	if err != nil {
		h.sendError(client, fmt.Errorf("invalid conversation id"))
		return
	}

	msg, err := h.messages.Send(
		ctx,
		client.userID,
		conversationID,
		req.Content,
	)
	if err != nil {
		log.Printf("send message: %v", err)
		h.sendError(client, err)
		return
	}

	event, err := json.Marshal(messageEvent{
		Type:    "message",
		Message: msg,
	})
	if err != nil {
		log.Printf("encode message event: %v", err)
		return
	}

	memberIDs, err := h.conversations.ListMembers(
		ctx,
		conversationID,
	)
	if err != nil {
		log.Printf("list conversation members: %v", err)
		h.sendError(client, fmt.Errorf("failed to deliver message"))
		return
	}

	for _, memberID := range memberIDs {
		h.hub.SendToUser(memberID, event)
	}
}

func (h *Handler) handleSync(
	ctx context.Context,
	client *Client,
	req messageRequest,
) {
	conversationID, err := uuid.Parse(req.ConversationID)
	if err != nil {
		h.sendError(client, fmt.Errorf("invalid conversation id"))
		return
	}

	if req.After == "" {
		h.sendError(client, fmt.Errorf("missing sync cursor"))
		return
	}

	cursor, err := message.DecodeCursor(req.After)
	if err != nil {
		h.sendError(client, fmt.Errorf("invalid cursor"))
		return
	}

	messages, err := h.messages.ListAfter(
		ctx,
		client.userID,
		conversationID,
		&cursor,
	)
	if err != nil {
		log.Printf("sync messages: %v", err)
		h.sendError(client, err)
		return
	}

	event, err := json.Marshal(syncEvent{
		Type:     "sync",
		Messages: messages,
	})
	if err != nil {
		log.Printf("encode sync event: %v", err)
		return
	}

	select {
	case client.send <- event:
	default:
		log.Printf("client send buffer full")
	}
}
