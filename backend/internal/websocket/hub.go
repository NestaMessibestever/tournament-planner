// internal/websocket/hub.go
// WebSocket hub manages client connections and message broadcasting

package websocket

import (
	"encoding/json"
	"log"
	"sync"

	"tournament-planner/internal/services"
)

// Hub maintains active websocket connections and broadcasts messages
type Hub struct {
	// Registered clients by tournament ID
	tournaments map[string]map[*Client]bool

	// Registered clients by user ID
	users map[string]*Client

	// Register client
	register chan *Client

	// Unregister client
	unregister chan *Client

	// Broadcast messages to tournament
	broadcast chan *Message

	// Services
	services *services.Container
	logger   *log.Logger

	// Mutex for concurrent access
	mu sync.RWMutex
}

// Message represents a WebSocket message
type Message struct {
	Type         string      `json:"type"`
	TournamentID string      `json:"tournament_id,omitempty"`
	UserID       string      `json:"user_id,omitempty"`
	Data         interface{} `json:"data"`
}

// NewHub creates a new WebSocket hub
func NewHub(services *services.Container, logger *log.Logger) *Hub {
	return &Hub{
		tournaments: make(map[string]map[*Client]bool),
		users:       make(map[string]*Client),
		register:    make(chan *Client),
		unregister:  make(chan *Client),
		broadcast:   make(chan *Message, 256),
		services:    services,
		logger:      logger,
	}
}

// Run starts the hub's main loop
func (h *Hub) Run() {
	for {
		select {
		case client := <-h.register:
			h.registerClient(client)

		case client := <-h.unregister:
			h.unregisterClient(client)

		case message := <-h.broadcast:
			h.broadcastMessage(message)
		}
	}
}

// registerClient adds a new client to the hub
func (h *Hub) registerClient(client *Client) {
	h.mu.Lock()
	defer h.mu.Unlock()

	// Register user connection
	if client.userID != "" {
		// Close existing connection for this user
		if existing, exists := h.users[client.userID]; exists {
			existing.close()
			h.removeClient(existing)
		}
		h.users[client.userID] = client
	}

	// Register tournament connections
	for _, tournamentID := range client.tournaments {
		if h.tournaments[tournamentID] == nil {
			h.tournaments[tournamentID] = make(map[*Client]bool)
		}
		h.tournaments[tournamentID][client] = true
	}

	h.logger.Printf("Client registered: %s (tournaments: %v)", client.userID, client.tournaments)
}

// unregisterClient removes a client from the hub
func (h *Hub) unregisterClient(client *Client) {
	h.mu.Lock()
	defer h.mu.Unlock()

	h.removeClient(client)
	client.close()

	h.logger.Printf("Client unregistered: %s", client.userID)
}

// removeClient removes client from all registrations
func (h *Hub) removeClient(client *Client) {
	// Remove from user map
	if client.userID != "" {
		delete(h.users, client.userID)
	}

	// Remove from tournament maps
	for _, tournamentID := range client.tournaments {
		if clients, exists := h.tournaments[tournamentID]; exists {
			delete(clients, client)
			if len(clients) == 0 {
				delete(h.tournaments, tournamentID)
			}
		}
	}
}

// broadcastMessage sends a message to relevant clients
func (h *Hub) broadcastMessage(message *Message) {
	h.mu.RLock()
	defer h.mu.RUnlock()

	data, err := json.Marshal(message)
	if err != nil {
		h.logger.Printf("Failed to marshal message: %v", err)
		return
	}

	// Broadcast to tournament participants
	if message.TournamentID != "" {
		if clients, exists := h.tournaments[message.TournamentID]; exists {
			for client := range clients {
				select {
				case client.send <- data:
				default:
					// Client's send channel is full, close it
					h.removeClient(client)
					client.close()
				}
			}
		}
	}

	// Send to specific user
	if message.UserID != "" {
		if client, exists := h.users[message.UserID]; exists {
			select {
			case client.send <- data:
			default:
				// Client's send channel is full, close it
				h.removeClient(client)
				client.close()
			}
		}
	}
}

// BroadcastTournamentUpdate broadcasts an update to all tournament participants
func (h *Hub) BroadcastTournamentUpdate(tournamentID string, updateType string, data interface{}) {
	message := &Message{
		Type:         updateType,
		TournamentID: tournamentID,
		Data:         data,
	}
	h.broadcast <- message
}

// SendToUser sends a message to a specific user
func (h *Hub) SendToUser(userID string, messageType string, data interface{}) {
	message := &Message{
		Type:   messageType,
		UserID: userID,
		Data:   data,
	}
	h.broadcast <- message
}

// SubscribeToTournament subscribes a client to tournament updates
func (h *Hub) SubscribeToTournament(client *Client, tournamentID string) {
	h.mu.Lock()
	defer h.mu.Unlock()

	// Add tournament to client's list
	client.tournaments = append(client.tournaments, tournamentID)

	// Add client to tournament's subscriber list
	if h.tournaments[tournamentID] == nil {
		h.tournaments[tournamentID] = make(map[*Client]bool)
	}
	h.tournaments[tournamentID][client] = true

	h.logger.Printf("Client %s subscribed to tournament %s", client.userID, tournamentID)
}

// UnsubscribeFromTournament unsubscribes a client from tournament updates
func (h *Hub) UnsubscribeFromTournament(client *Client, tournamentID string) {
	h.mu.Lock()
	defer h.mu.Unlock()

	// Remove tournament from client's list
	for i, id := range client.tournaments {
		if id == tournamentID {
			client.tournaments = append(client.tournaments[:i], client.tournaments[i+1:]...)
			break
		}
	}

	// Remove client from tournament's subscriber list
	if clients, exists := h.tournaments[tournamentID]; exists {
		delete(clients, client)
		if len(clients) == 0 {
			delete(h.tournaments, tournamentID)
		}
	}

	h.logger.Printf("Client %s unsubscribed from tournament %s", client.userID, tournamentID)
}
