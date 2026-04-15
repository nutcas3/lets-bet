package games

import (
	"encoding/json"
	"log"
	"sync"
	"time"

	"github.com/betting-platform/internal/core/domain"
	"github.com/google/uuid"
	"github.com/gorilla/websocket"
	"github.com/shopspring/decimal"
)

// GameState represents the current state of a crash game
type GameState struct {
	GameID         uuid.UUID       `json:"game_id"`
	RoundNumber    int64           `json:"round_number"`
	Status         domain.GameStatus `json:"status"`
	CurrentMultiplier decimal.Decimal `json:"current_multiplier"`
	CrashPoint     *decimal.Decimal `json:"crash_point,omitempty"` // Hidden until crash
	TimeRemaining  int             `json:"time_remaining"` // Seconds until next phase
	ActivePlayers  int             `json:"active_players"`
}

// Client represents a connected WebSocket client
type Client struct {
	ID       uuid.UUID
	UserID   uuid.UUID
	Conn     *websocket.Conn
	Send     chan []byte
	GameID   uuid.UUID
}

// Hub maintains the set of active clients and broadcasts messages
type Hub struct {
	// Registered clients by game ID
	games      map[uuid.UUID]map[*Client]bool
	gamesMutex sync.RWMutex
	
	// Inbound messages from clients
	broadcast  chan BroadcastMessage
	
	// Register requests from clients
	register   chan *Client
	
	// Unregister requests from clients
	unregister chan *Client
}

type BroadcastMessage struct {
	GameID  uuid.UUID
	Message []byte
}

func NewHub() *Hub {
	return &Hub{
		games:      make(map[uuid.UUID]map[*Client]bool),
		broadcast:  make(chan BroadcastMessage, 256),
		register:   make(chan *Client),
		unregister: make(chan *Client),
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
			h.broadcastToGame(message)
		}
	}
}

func (h *Hub) registerClient(client *Client) {
	h.gamesMutex.Lock()
	defer h.gamesMutex.Unlock()
	
	if h.games[client.GameID] == nil {
		h.games[client.GameID] = make(map[*Client]bool)
	}
	h.games[client.GameID][client] = true
	
	log.Printf("Client %s joined game %s. Total: %d", client.ID, client.GameID, len(h.games[client.GameID]))
}

func (h *Hub) unregisterClient(client *Client) {
	h.gamesMutex.Lock()
	defer h.gamesMutex.Unlock()
	
	if clients, ok := h.games[client.GameID]; ok {
		if _, ok := clients[client]; ok {
			delete(clients, client)
			close(client.Send)
			
			if len(clients) == 0 {
				delete(h.games, client.GameID)
			}
		}
	}
}

func (h *Hub) broadcastToGame(msg BroadcastMessage) {
	h.gamesMutex.RLock()
	clients := h.games[msg.GameID]
	h.gamesMutex.RUnlock()
	
	for client := range clients {
		select {
		case client.Send <- msg.Message:
		default:
			// Client's send buffer is full, disconnect
			h.unregisterClient(client)
		}
	}
}

// GetActivePlayerCount returns the number of connected players for a game
func (h *Hub) GetActivePlayerCount(gameID uuid.UUID) int {
	h.gamesMutex.RLock()
	defer h.gamesMutex.RUnlock()
	
	if clients, ok := h.games[gameID]; ok {
		return len(clients)
	}
	return 0
}

// BroadcastGameState sends the current game state to all connected clients
func (h *Hub) BroadcastGameState(state *GameState) {
	data, err := json.Marshal(state)
	if err != nil {
		log.Printf("Error marshaling game state: %v", err)
		return
	}
	
	h.broadcast <- BroadcastMessage{
		GameID:  state.GameID,
		Message: data,
	}
}

// ============ CLIENT READER/WRITER ============

// ReadPump pumps messages from the websocket connection to the hub
func (c *Client) ReadPump(hub *Hub) {
	defer func() {
		hub.unregister <- c
		c.Conn.Close()
	}()
	
	c.Conn.SetReadDeadline(time.Now().Add(60 * time.Second))
	c.Conn.SetPongHandler(func(string) error {
		c.Conn.SetReadDeadline(time.Now().Add(60 * time.Second))
		return nil
	})
	
	for {
		_, message, err := c.Conn.ReadMessage()
		if err != nil {
			if websocket.IsUnexpectedCloseError(err, websocket.CloseGoingAway, websocket.CloseAbnormalClosure) {
				log.Printf("WebSocket error: %v", err)
			}
			break
		}
		
		// Handle incoming messages (e.g., place bet, cashout)
		handleClientMessage(c, message)
	}
}

// WritePump pumps messages from the hub to the websocket connection
func (c *Client) WritePump() {
	ticker := time.NewTicker(54 * time.Second)
	defer func() {
		ticker.Stop()
		c.Conn.Close()
	}()
	
	for {
		select {
		case message, ok := <-c.Send:
			c.Conn.SetWriteDeadline(time.Now().Add(10 * time.Second))
			if !ok {
				c.Conn.WriteMessage(websocket.CloseMessage, []byte{})
				return
			}
			
			w, err := c.Conn.NextWriter(websocket.TextMessage)
			if err != nil {
				return
			}
			w.Write(message)
			
			// Add queued messages to the current websocket message
			n := len(c.Send)
			for i := 0; i < n; i++ {
				w.Write([]byte{'\n'})
				w.Write(<-c.Send)
			}
			
			if err := w.Close(); err != nil {
				return
			}
			
		case <-ticker.C:
			c.Conn.SetWriteDeadline(time.Now().Add(10 * time.Second))
			if err := c.Conn.WriteMessage(websocket.PingMessage, nil); err != nil {
				return
			}
		}
	}
}

func handleClientMessage(c *Client, message []byte) {
	var msg map[string]interface{}
	if err := json.Unmarshal(message, &msg); err != nil {
		log.Printf("Invalid message from client %s: %v", c.ID, err)
		return
	}
	
	action, ok := msg["action"].(string)
	if !ok {
		return
	}
	
	switch action {
	case "place_bet":
		// Handle bet placement
		log.Printf("Client %s placing bet", c.ID)
	case "cashout":
		// Handle cashout request
		log.Printf("Client %s cashing out", c.ID)
	}
}
