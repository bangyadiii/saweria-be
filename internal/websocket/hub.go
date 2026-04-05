package websocket

import (
	"sync"
	"time"

	"github.com/gorilla/websocket"
)

const writeWait = 5 * time.Second

// Hub manages WebSocket connections keyed by stream key hash
type Hub struct {
	mu    sync.RWMutex
	rooms map[string][]*websocket.Conn
}

func NewHub() *Hub {
	return &Hub{rooms: make(map[string][]*websocket.Conn)}
}

func (h *Hub) Register(streamKeyHash string, conn *websocket.Conn) {
	h.mu.Lock()
	defer h.mu.Unlock()
	h.rooms[streamKeyHash] = append(h.rooms[streamKeyHash], conn)
}

func (h *Hub) Unregister(streamKeyHash string, conn *websocket.Conn) {
	h.mu.Lock()
	defer h.mu.Unlock()
	conns := h.rooms[streamKeyHash]
	updated := conns[:0]
	for _, c := range conns {
		if c != conn {
			updated = append(updated, c)
		}
	}
	if len(updated) == 0 {
		delete(h.rooms, streamKeyHash)
	} else {
		h.rooms[streamKeyHash] = updated
	}
}

func (h *Hub) Broadcast(streamKeyHash string, message []byte) {
	h.mu.RLock()
	conns := make([]*websocket.Conn, len(h.rooms[streamKeyHash]))
	copy(conns, h.rooms[streamKeyHash])
	h.mu.RUnlock()

	var dead []*websocket.Conn
	for _, conn := range conns {
		conn.SetWriteDeadline(time.Now().Add(writeWait))
		if err := conn.WriteMessage(websocket.TextMessage, message); err != nil {
			dead = append(dead, conn)
		}
	}
	for _, conn := range dead {
		h.Unregister(streamKeyHash, conn)
		conn.Close()
	}
}
