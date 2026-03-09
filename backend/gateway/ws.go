package main

import (
	"context"
	"encoding/json"
	"log"
	"net/http"
	"sync"

	"github.com/gorilla/websocket"
	"github.com/redis/go-redis/v9"
)

const redisChannelAvatarSaved = "avatar.saved"

var upgrader = websocket.Upgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
	CheckOrigin: func(r *http.Request) bool {
		return true // CORS handled at gateway; allow all for WS
	},
}

// wsHub maps user_id -> list of WebSocket connections.
type wsHub struct {
	mu          sync.RWMutex
	byUser      map[string]map[*wsConn]struct{}
	register    chan *wsConn
	unregister  chan *wsConn
	broadcast   chan []byte // not used; we push by user from Redis
	redisNotify chan []byte // payload from Redis avatar.saved
}

type wsConn struct {
	userID string
	send   chan []byte
	conn   *websocket.Conn
}

func newHub() *wsHub {
	return &wsHub{
		byUser:      make(map[string]map[*wsConn]struct{}),
		register:    make(chan *wsConn),
		unregister:  make(chan *wsConn),
		redisNotify: make(chan []byte, 64),
	}
}

func (h *wsHub) run() {
	for {
		select {
		case c := <-h.register:
			h.mu.Lock()
			if h.byUser[c.userID] == nil {
				h.byUser[c.userID] = make(map[*wsConn]struct{})
			}
			h.byUser[c.userID][c] = struct{}{}
			h.mu.Unlock()
		case c := <-h.unregister:
			h.mu.Lock()
			if m := h.byUser[c.userID]; m != nil {
				delete(m, c)
				if len(m) == 0 {
					delete(h.byUser, c.userID)
				}
			}
			close(c.send)
			h.mu.Unlock()
		case payload := <-h.redisNotify:
			var msg struct {
				UserID  string `json:"user_id"`
				ImageID string `json:"image_id"`
			}
			if err := json.Unmarshal(payload, &msg); err != nil {
				continue
			}
			h.mu.RLock()
			conns := h.byUser[msg.UserID]
			h.mu.RUnlock()
			for c := range conns {
				select {
				case c.send <- payload:
				default:
					// skip if send buffer full
				}
			}
		}
	}
}

func (h *wsHub) handleWS(w http.ResponseWriter, r *http.Request) {
	userID := r.URL.Query().Get("user_id")
	if userID == "" {
		http.Error(w, "user_id required", http.StatusBadRequest)
		return
	}
	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Printf("ws upgrade: %v", err)
		return
	}
	c := &wsConn{userID: userID, send: make(chan []byte, 8), conn: conn}
	h.register <- c
	defer func() { h.unregister <- c }()

	go func() {
		for b := range c.send {
			if err := c.conn.WriteMessage(websocket.TextMessage, b); err != nil {
				return
			}
		}
	}()
	// Block until client disconnects (read loop)
	for {
		if _, _, err := c.conn.ReadMessage(); err != nil {
			break
		}
	}
}

// runRedisSub subscribes to avatar.saved and forwards to hub.
func runRedisSub(rdb *redis.Client, hub *wsHub) {
	ctx := context.Background()
	pubsub := rdb.Subscribe(ctx, redisChannelAvatarSaved)
	defer pubsub.Close()
	for msg := range pubsub.Channel() {
		if msg.Channel == redisChannelAvatarSaved && len(msg.Payload) > 0 {
			hub.redisNotify <- []byte(msg.Payload)
		}
	}
}
