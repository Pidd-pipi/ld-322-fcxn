package websocket

import (
	"github.com/gorilla/websocket"
	"net/http"
	"sync"
)

type Hub struct {
	clients  map[*websocket.Conn]struct{}
	mu       sync.RWMutex
	upgrader websocket.Upgrader
}

func NewHub() *Hub {
	return &Hub{clients: make(map[*websocket.Conn]struct{}), upgrader: websocket.Upgrader{CheckOrigin: func(*http.Request) bool { return true }}}
}
func (h *Hub) Handle(w http.ResponseWriter, r *http.Request) {
	conn, err := h.upgrader.Upgrade(w, r, nil)
	if err != nil {
		return
	}
	h.mu.Lock()
	h.clients[conn] = struct{}{}
	h.mu.Unlock()
	defer func() { h.mu.Lock(); delete(h.clients, conn); h.mu.Unlock(); _ = conn.Close() }()
	for {
		if _, _, err = conn.ReadMessage(); err != nil {
			return
		}
	}
}
func (h *Hub) Broadcast(event string, data any) {
	h.mu.RLock()
	defer h.mu.RUnlock()
	for conn := range h.clients {
		if err := conn.WriteJSON(map[string]any{"event": event, "data": data}); err != nil {
			_ = conn.Close()
		}
	}
}
