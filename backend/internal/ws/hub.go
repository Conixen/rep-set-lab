package ws

import (
	"encoding/json"
	"log/slog"
	"net/http"
	"sync"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/gorilla/websocket"
	"github.com/leonj/rep-set-lab/internal/auth"
)

const (
	pingInterval   = 30 * time.Second
	pongDeadline   = 60 * time.Second
	writeDeadline  = 10 * time.Second
	maxMessageSize = 512
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
	mu       sync.RWMutex
	clients  map[int64]*client
	logger   *slog.Logger
	upgrader websocket.Upgrader
}

func NewHub(logger *slog.Logger, allowedOrigins []string) *Hub {
	allowed := make(map[string]struct{}, len(allowedOrigins))
	for _, o := range allowedOrigins {
		allowed[o] = struct{}{}
	}
	return &Hub{
		clients: make(map[int64]*client),
		logger:  logger,
		upgrader: websocket.Upgrader{
			CheckOrigin: func(r *http.Request) bool {
				origin := r.Header.Get("Origin")
				if origin == "" {
					return true // non-browser clients (curl, mobile apps)
				}
				_, ok := allowed[origin]
				return ok
			},
		},
	}
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
	conn, err := h.upgrader.Upgrade(c.Writer, c.Request, nil)
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
	c.conn.SetReadLimit(maxMessageSize)
	_ = c.conn.SetReadDeadline(time.Now().Add(pongDeadline))
	c.conn.SetPongHandler(func(string) error {
		return c.conn.SetReadDeadline(time.Now().Add(pongDeadline))
	})
	for {
		if _, _, err := c.conn.ReadMessage(); err != nil {
			break
		}
	}
}

func (h *Hub) writePump(c *client) {
	ticker := time.NewTicker(pingInterval)
	defer ticker.Stop()
	for {
		select {
		case msg, ok := <-c.send:
			_ = c.conn.SetWriteDeadline(time.Now().Add(writeDeadline))
			if !ok {
				_ = c.conn.WriteMessage(websocket.CloseMessage, []byte{})
				return
			}
			if err := c.conn.WriteMessage(websocket.TextMessage, msg); err != nil {
				return
			}
		case <-ticker.C:
			_ = c.conn.SetWriteDeadline(time.Now().Add(writeDeadline))
			if err := c.conn.WriteMessage(websocket.PingMessage, nil); err != nil {
				return
			}
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
