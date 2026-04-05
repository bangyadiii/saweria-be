package websocket

import (
	"context"
	"crypto/sha256"
	"fmt"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	gws "github.com/gorilla/websocket"
)

const (
	pingInterval = 5 * time.Second  // how often server sends a ping to client
	pongWait     = 10 * time.Second // read deadline; must be > pingInterval
)

// OverlayLookup determines whether a stream key is valid
type OverlayLookup interface {
	StreamKeyExists(ctx context.Context, hash string) (bool, error)
}

var upgrader = gws.Upgrader{
	CheckOrigin: func(r *http.Request) bool { return true },
}

type Handler struct {
	hub         *Hub
	overlayRepo OverlayLookup
}

func NewHandler(hub *Hub, lookup OverlayLookup) *Handler {
	return &Handler{hub: hub, overlayRepo: lookup}
}

// Connect handles: GET /ws?key=<plain_stream_key>
func (h *Handler) Connect(c *gin.Context) {
	plainKey := c.Query("key")
	if plainKey == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "missing key"})
		return
	}

	hash := hashKey(plainKey)

	ok, err := h.overlayRepo.StreamKeyExists(c.Request.Context(), hash)
	if err != nil || !ok {
		c.JSON(http.StatusForbidden, gin.H{"error": "invalid stream key"})
		return
	}

	conn, err := upgrader.Upgrade(c.Writer, c.Request, nil)
	if err != nil {
		return
	}
	defer conn.Close()

	h.hub.Register(hash, conn)
	defer h.hub.Unregister(hash, conn)

	// set initial read deadline — will be extended on every pong received
	conn.SetReadDeadline(time.Now().Add(pongWait))

	// reset deadline each time the client sends a pong back
	conn.SetPongHandler(func(string) error {
		return conn.SetReadDeadline(time.Now().Add(pongWait))
	})

	// done signals the ping goroutine to stop when the read loop exits
	done := make(chan struct{})
	defer close(done) // executed first (LIFO) — stops ping goroutine before Unregister

	// ping goroutine: server is responsible for keepalive
	go func() {
		ticker := time.NewTicker(pingInterval)
		defer ticker.Stop()
		for {
			select {
			case <-ticker.C:
				conn.SetWriteDeadline(time.Now().Add(writeWait))
				if err := conn.WriteMessage(gws.PingMessage, nil); err != nil {
					// write failed (OBS closed or hung) — force read loop to exit
					conn.Close()
					return
				}
			case <-done:
				return
			}
		}
	}()

	// read loop — blocks until read deadline exceeded or client disconnects
	for {
		if _, _, err := conn.ReadMessage(); err != nil {
			break
		}
	}
	// LIFO defers: close(done) → Unregister → conn.Close()
}

func hashKey(key string) string {
	sum := sha256.Sum256([]byte(key))
	return fmt.Sprintf("%x", sum)
}
