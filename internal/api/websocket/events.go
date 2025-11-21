package websocket

import (
	"encoding/json"
	"net/http"
	"sync"
	"time"

	"github.com/gorilla/websocket"
	"steri-connect-go/internal/logging"
)

var upgrader = websocket.Upgrader{
	CheckOrigin: func(r *http.Request) bool {
		// Allow connections from localhost
		origin := r.Header.Get("Origin")
		return origin == "" ||
			origin == "http://localhost:3000" ||
			origin == "http://127.0.0.1:3000" ||
			origin == "http://localhost:8080" ||
			origin == "http://127.0.0.1:8080"
	},
}

// Hub maintains the set of active clients and broadcasts messages to clients
type Hub struct {
	// Registered clients
	clients map[*Client]bool

	// Inbound messages from clients
	broadcast chan []byte

	// Register requests from clients
	register chan *Client

	// Unregister requests from clients
	unregister chan *Client

	mu sync.RWMutex
}

// Client is a middleman between the websocket connection and the hub
type Client struct {
	hub *Hub

	// The websocket connection
	conn *websocket.Conn

	// Buffered channel of outbound messages
	send chan []byte
}

// Event represents a WebSocket event
type Event struct {
	Event     string                 `json:"event"`
	Timestamp string                 `json:"timestamp"`
	Data      map[string]interface{} `json:"data"`
}

var globalHub *Hub

// GetHub returns the global WebSocket hub
func GetHub() *Hub {
	if globalHub == nil {
		globalHub = &Hub{
			clients:    make(map[*Client]bool),
			broadcast:  make(chan []byte, 256),
			register:   make(chan *Client),
			unregister: make(chan *Client),
		}
		go globalHub.run()
	}
	return globalHub
}

// run starts the hub's message broadcasting loop
func (h *Hub) run() {
	for {
		select {
		case client := <-h.register:
			h.mu.Lock()
			h.clients[client] = true
			h.mu.Unlock()

		case client := <-h.unregister:
			h.mu.Lock()
			if _, ok := h.clients[client]; ok {
				delete(h.clients, client)
				close(client.send)
			}
			h.mu.Unlock()

		case message := <-h.broadcast:
			h.mu.RLock()
			for client := range h.clients {
				select {
				case client.send <- message:
				default:
					close(client.send)
					delete(h.clients, client)
				}
			}
			h.mu.RUnlock()
		}
	}
}

// BroadcastEvent broadcasts an event to all connected clients
func BroadcastEvent(event Event) error {
	event.Timestamp = getCurrentTimestamp()
	data, err := json.Marshal(event)
	if err != nil {
		return err
	}

	hub := GetHub()
	hub.broadcast <- data
	return nil
}

// HandleWebSocket handles WebSocket connections
func HandleWebSocket(w http.ResponseWriter, r *http.Request) {
	logger := logging.Get()
	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		logger.Error("WebSocket upgrade failed", "error", err)
		return
	}

	hub := GetHub()
	client := &Client{
		hub:  hub,
		conn: conn,
		send: make(chan []byte, 256),
	}

	client.hub.register <- client

	// Allow collection of memory referenced by the caller by doing all work in new goroutines
	go client.writePump()
	go client.readPump()
}

// readPump pumps messages from the websocket connection to the hub
func (c *Client) readPump() {
	defer func() {
		c.hub.unregister <- c
		c.conn.Close()
	}()

	c.conn.SetReadDeadline(getReadDeadline())
	c.conn.SetPongHandler(func(string) error {
		c.conn.SetReadDeadline(getReadDeadline())
		return nil
	})

		for {
		_, _, err := c.conn.ReadMessage()
		if err != nil {
			if websocket.IsUnexpectedCloseError(err, websocket.CloseGoingAway, websocket.CloseAbnormalClosure) {
				logger := logging.Get()
				logger.Error("WebSocket error", "error", err)
			}
			break
		}
	}
}

// writePump pumps messages from the hub to the websocket connection
func (c *Client) writePump() {
	defer c.conn.Close()

	for {
		select {
		case message, ok := <-c.send:
			c.conn.SetWriteDeadline(getWriteDeadline())
			if !ok {
				c.conn.WriteMessage(websocket.CloseMessage, []byte{})
				return
			}

			w, err := c.conn.NextWriter(websocket.TextMessage)
			if err != nil {
				return
			}
			w.Write(message)

			// Add queued messages to the current websocket message
			n := len(c.send)
			for i := 0; i < n; i++ {
				w.Write([]byte{'\n'})
				w.Write(<-c.send)
			}

			if err := w.Close(); err != nil {
				return
			}
		}
	}
}

// Helper functions
func getReadDeadline() time.Time {
	return time.Now().Add(60 * time.Second)
}

func getWriteDeadline() time.Time {
	return time.Now().Add(10 * time.Second)
}

func getCurrentTimestamp() string {
	return time.Now().UTC().Format(time.RFC3339)
}

