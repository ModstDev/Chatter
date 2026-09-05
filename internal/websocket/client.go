package websocket

import (
	"context"

	"github.com/google/uuid"

	"github.com/coder/websocket"
)

type Client struct {
	conn   *websocket.Conn
	userID uuid.UUID

	send chan []byte
}

func NewClient(conn *websocket.Conn, userID uuid.UUID) *Client {
	return &Client{
		conn:   conn,
		userID: userID,
		send:   make(chan []byte, 256),
	}
}

func (c *Client) UserID() uuid.UUID {
	return c.userID
}

func (c *Client) writeLoop(ctx context.Context) error {
	for {
		select {
		case <-ctx.Done():
			return ctx.Err()

		case message, ok := <-c.send:
			if !ok {
				return nil
			}

			if err := c.conn.Write(ctx, websocket.MessageText, message); err != nil {
				return err
			}
		}
	}
}
