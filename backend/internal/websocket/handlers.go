// internal/websocket/handlers.go
// WebSocket connection handlers

package websocket

import (
	"encoding/json"
	"log"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/gorilla/websocket"
)

var upgrader = websocket.Upgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
	CheckOrigin: func(r *http.Request) bool {
		// In production, implement proper origin checking
		return true
	},
}

// HandleConnection handles new WebSocket connections
func HandleConnection(hub *Hub) gin.HandlerFunc {
	return func(c *gin.Context) {
		// Get user ID from context (set by auth middleware)
		userID, _ := c.Get("user_id")
		userIDStr := ""
		if userID != nil {
			userIDStr = userID.(string)
		}

		// Upgrade HTTP connection to WebSocket
		conn, err := upgrader.Upgrade(c.Writer, c.Request, nil)
		if err != nil {
			log.Printf("Failed to upgrade connection: %v", err)
			return
		}

		// Create new client
		client := &Client{
			hub:         hub,
			conn:        conn,
			send:        make(chan []byte, 256),
			userID:      userIDStr,
			tournaments: make([]string, 0),
		}

		// Register client with hub
		hub.register <- client

		// Send welcome message
		welcomeMsg := Message{
			Type: "welcome",
			Data: map[string]interface{}{
				"message": "Connected to Tournament Planner WebSocket",
				"user_id": userIDStr,
			},
		}

		if data, err := json.Marshal(welcomeMsg); err == nil {
			client.send <- data
		}

		// Start client pumps in goroutines
		go client.writePump()
		go client.readPump()
	}
}

// Message types for WebSocket communication
const (
	// Tournament updates
	MessageTournamentCreated   = "tournament_created"
	MessageTournamentUpdated   = "tournament_updated"
	MessageTournamentPublished = "tournament_published"
	MessageTournamentStarted   = "tournament_started"
	MessageTournamentCompleted = "tournament_completed"

	// Match updates
	MessageMatchScheduled    = "match_scheduled"
	MessageMatchStarted      = "match_started"
	MessageMatchScoreUpdated = "match_score_updated"
	MessageMatchCompleted    = "match_completed"

	// Participant updates
	MessageParticipantRegistered = "participant_registered"
	MessageParticipantWithdrawn  = "participant_withdrawn"
	MessageParticipantCheckedIn  = "participant_checked_in"

	// Bracket updates
	MessageBracketUpdated    = "bracket_updated"
	MessageFixturesGenerated = "fixtures_generated"

	// Notifications
	MessageNotification = "notification"
	MessageAlert        = "alert"
)
