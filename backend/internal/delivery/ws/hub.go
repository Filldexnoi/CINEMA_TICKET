package ws

import (
	"sync"

	"cinema-ticket/backend/internal/usecase/ports"
)

type Hub struct {
	mu    sync.RWMutex
	rooms map[string]map[*Client]struct{}
}

func NewHub() *Hub {
	return &Hub{rooms: make(map[string]map[*Client]struct{})}
}

var _ ports.Broadcaster = (*Hub)(nil)

func (h *Hub) Register(showtimeID string, c *Client) {
	h.mu.Lock()
	defer h.mu.Unlock()
	room, ok := h.rooms[showtimeID]
	if !ok {
		room = make(map[*Client]struct{})
		h.rooms[showtimeID] = room
	}
	room[c] = struct{}{}
}

func (h *Hub) Unregister(showtimeID string, c *Client) {
	h.mu.Lock()
	defer h.mu.Unlock()
	if room, ok := h.rooms[showtimeID]; ok {
		delete(room, c)
		if len(room) == 0 {
			delete(h.rooms, showtimeID)
		}
	}
}

func (h *Hub) BroadcastToShowtime(showtimeID string, payload []byte) {
	h.mu.RLock()
	defer h.mu.RUnlock()
	for c := range h.rooms[showtimeID] {
		c.Send(payload)
	}
}
