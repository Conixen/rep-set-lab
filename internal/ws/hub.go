package ws

import (
	"encoding/json"
	"log/slog"
	"net/http"
	"sync"

	"github.com/gin-gonic/gin"
	"github.com/gorilla/websocket"
	"github.com/leonj/rep-set-lab/internal/auth"
)

const EventXPUpdate = "xp_update"

type Event struct {
	Type string `json:"type"`
	Data any    `json:"data"`
}

type client struct {
	userID int64
	conn   *websocket.Conn
	send   chan []byte
}

type Hub struct {
	mu      sync.RWMutex
	clients map[int64]*client
	logger  *slog.Logger
}

var upgrader = websocket.Upgrader{
	CheckOrigin: func(r *http.Request) bool { return true },
}

func NewHub(logger *slog.Logger) *Hub {
	return &Hub{clients: make(map[int64]*client), logger: logger}
}

// Broadcast sends an event to the connected client for the given userID.
// No-ops silently if the user has no active connection.
func (h *Hub) Broadcast(userID int64, event Event) {
	h.mu.RLock()
	c, ok := h.clients[userID]
	h.mu.RUnlock()
	if !ok {
		return
	}
	data, err := json.Marshal(event)
	if err != nil {
		h.logger.Error("ws: marshal event", "error", err)
		return
	}
	select {
	case c.send <- data:
	default:
		h.remove(userID)
	}
}

func (h *Hub) Handler(c *gin.Context) {
	claims := auth.GetClaims(c)
	if claims == nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
		return
	}
	conn, err := upgrader.Upgrade(c.Writer, c.Request, nil)
	if err != nil {
		h.logger.Error("ws: upgrade failed", "error", err)
		return
	}
	cl := &client{userID: claims.UserID, conn: conn, send: make(chan []byte, 32)}

	h.mu.Lock()
	h.clients[claims.UserID] = cl
	h.mu.Unlock()

	go h.writePump(cl)
	h.readPump(cl)
}

func (h *Hub) readPump(c *client) {
	defer func() {
		h.remove(c.userID)
		c.conn.Close()
	}()
	for {
		if _, _, err := c.conn.ReadMessage(); err != nil {
			break
		}
	}
}

func (h *Hub) writePump(c *client) {
	for msg := range c.send {
		if err := c.conn.WriteMessage(websocket.TextMessage, msg); err != nil {
			break
		}
	}
}

func (h *Hub) remove(userID int64) {
	h.mu.Lock()
	defer h.mu.Unlock()
	if c, ok := h.clients[userID]; ok {
		close(c.send)
		delete(h.clients, userID)
	}
}
