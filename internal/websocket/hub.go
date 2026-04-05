package websocket

import (
	"sync"
	"time"

	"github.com/gorilla/websocket"
)

const writeWait = 5 * time.Second

// subathonState holds a live countdown timer state per stream key.
// totalSeconds is the value at lastUpdated; if running, elapsed time is subtracted on read.
type subathonState struct {
	totalSeconds int
	running      bool
	lastUpdated  time.Time
}

func (s *subathonState) current() int {
	if !s.running {
		return s.totalSeconds
	}
	elapsed := int(time.Since(s.lastUpdated).Seconds())
	rem := s.totalSeconds - elapsed
	if rem < 0 {
		return 0
	}
	return rem
}

// Hub manages WebSocket connections keyed by stream key hash
type Hub struct {
	mu       sync.RWMutex
	rooms    map[string][]*websocket.Conn
	subMu    sync.RWMutex
	subState map[string]*subathonState
}

func NewHub() *Hub {
	return &Hub{
		rooms:    make(map[string][]*websocket.Conn),
		subState: make(map[string]*subathonState),
	}
}

// SetSubathonState stores (or updates) the timer state for a stream key.
// totalSeconds is the value right now; running indicates whether it is counting down.
func (h *Hub) SetSubathonState(hash string, totalSeconds int, running bool) {
	h.subMu.Lock()
	defer h.subMu.Unlock()
	h.subState[hash] = &subathonState{
		totalSeconds: totalSeconds,
		running:      running,
		lastUpdated:  time.Now(),
	}
}

// GetSubathonSeconds returns the current remaining seconds, running status, and whether
// state has been set for this hash. Accounts for time that has elapsed since last update.
func (h *Hub) GetSubathonSeconds(hash string) (totalSeconds int, running bool, ok bool) {
	h.subMu.RLock()
	defer h.subMu.RUnlock()
	s, exists := h.subState[hash]
	if !exists {
		return 0, false, false
	}
	return s.current(), s.running, true
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
