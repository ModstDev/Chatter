package websocket

import (
	"log"
	"sync"

	"github.com/google/uuid"
)

type Hub struct {
	mu      sync.RWMutex
	clients map[uuid.UUID]map[*Client]struct{}
}

func NewHub() *Hub {
	return &Hub{
		clients: make(map[uuid.UUID]map[*Client]struct{}),
	}
}

func (h *Hub) Register(client *Client) {
	h.mu.Lock()
	defer h.mu.Unlock()

	clients := h.clients[client.UserID()]

	if clients == nil {
		clients = make(map[*Client]struct{})
		h.clients[client.UserID()] = clients
	}

	clients[client] = struct{}{}
}

func (h *Hub) Unregister(client *Client) {
	h.mu.Lock()
	defer h.mu.Unlock()

	clients := h.clients[client.UserID()]

	if clients == nil {
		return
	}

	delete(clients, client)

	if len(clients) == 0 {
		delete(h.clients, client.userID)
	}
}

func (h *Hub) SendToUser(userID uuid.UUID, message []byte) {
	h.mu.RLock()
	defer h.mu.RUnlock()

	for client := range h.clients[userID] {
		select {
		case client.send <- message:
		default:
			// Client is too slow
			log.Printf(
				"client send buffer full: user=%s",
				client.UserID(),
			)
		}
	}
}
