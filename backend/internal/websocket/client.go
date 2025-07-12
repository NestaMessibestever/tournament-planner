// internal/websocket/client.go
// WebSocket client connection handler

package websocket

import (
	"encoding/json"
	"log"
	"time"

	"github.com/gorilla/websocket"
)

const (
	// Time allowed to write a message to the peer
	writeWait = 10 * time.Second

	// Time allowed to read the next pong message from the peer
	pongWait = 60 * time.Second

	// Send pings to peer with this period (must be less than pongWait)
	pingPeriod = (pongWait * 9) / 10

	// Maximum message size allowed from peer
	maxMessageSize = 512 * 1024 // 512KB
)

// Client represents a websocket client connection
type Client struct {
	hub         *Hub
	conn        *websocket.Conn
	send        chan []byte
	userID      string
	tournaments []string
}

// ClientMessage represents a message from client
type ClientMessage struct {
	Type   string          `json:"type"`
	Action string          `json:"action"`
	Data   json.RawMessage `json:"data"`
}

// readPump pumps messages from the websocket connection to the hub
func (c *Client) readPump() {
	defer func() {
		c.hub.unregister <- c
		c.conn.Close()
	}()

	c.conn.SetReadLimit(maxMessageSize)
	c.conn.SetReadDeadline(time.Now().Add(pongWait))
	c.conn.SetPongHandler(func(string) error {
		c.conn.SetReadDeadline(time.Now().Add(pongWait))
		return nil
	})

	for {
		var msg ClientMessage
		err := c.conn.ReadJSON(&msg)
		if err != nil {
			if websocket.IsUnexpectedCloseError(err, websocket.CloseGoingAway, websocket.CloseAbnormalClosure) {
				log.Printf("WebSocket error: %v", err)
			}
			break
		}

		// Handle different message types
		switch msg.Type {
		case "subscribe":
			c.handleSubscribe(msg)
		case "unsubscribe":
			c.handleUnsubscribe(msg)
		case "ping":
			c.handlePing()
		default:
			log.Printf("Unknown message type: %s", msg.Type)
		}
	}
}

// writePump pumps messages from the hub to the websocket connection
func (c *Client) writePump() {
	ticker := time.NewTicker(pingPeriod)
	defer func() {
		ticker.Stop()
		c.conn.Close()
	}()

	for {
		select {
		case message, ok := <-c.send:
			c.conn.SetWriteDeadline(time.Now().Add(writeWait))
			if !ok {
				// The hub closed the channel
				c.conn.WriteMessage(websocket.CloseMessage, []byte{})
				return
			}

			if err := c.conn.WriteMessage(websocket.TextMessage, message); err != nil {
				return
			}

		case <-ticker.C:
			c.conn.SetWriteDeadline(time.Now().Add(writeWait))
			if err := c.conn.WriteMessage(websocket.PingMessage, nil); err != nil {
				return
			}
		}
	}
}

// handleSubscribe handles tournament subscription requests
func (c *Client) handleSubscribe(msg ClientMessage) {
	var data struct {
		TournamentID string `json:"tournament_id"`
	}

	if err := json.Unmarshal(msg.Data, &data); err != nil {
		log.Printf("Failed to unmarshal subscribe data: %v", err)
		return
	}

	if data.TournamentID != "" {
		c.hub.SubscribeToTournament(c, data.TournamentID)

		// Send confirmation
		response := Message{
			Type: "subscribed",
			Data: map[string]string{
				"tournament_id": data.TournamentID,
			},
		}

		if responseData, err := json.Marshal(response); err == nil {
			c.send <- responseData
		}
	}
}

// handleUnsubscribe handles tournament unsubscription requests
func (c *Client) handleUnsubscribe(msg ClientMessage) {
	var data struct {
		TournamentID string `json:"tournament_id"`
	}

	if err := json.Unmarshal(msg.Data, &data); err != nil {
		log.Printf("Failed to unmarshal unsubscribe data: %v", err)
		return
	}

	if data.TournamentID != "" {
		c.hub.UnsubscribeFromTournament(c, data.TournamentID)

		// Send confirmation
		response := Message{
			Type: "unsubscribed",
			Data: map[string]string{
				"tournament_id": data.TournamentID,
			},
		}

		if responseData, err := json.Marshal(response); err == nil {
			c.send <- responseData
		}
	}
}

// handlePing responds to ping messages
func (c *Client) handlePing() {
	response := Message{
		Type: "pong",
		Data: map[string]int64{
			"timestamp": time.Now().Unix(),
		},
	}

	if responseData, err := json.Marshal(response); err == nil {
		c.send <- responseData
	}
}

// close cleanly closes the client connection
func (c *Client) close() {
	close(c.send)
}
