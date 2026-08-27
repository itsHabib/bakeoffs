// Package stream fans server events out to browsers over SSE.
//
// One direction, no framing library, no upgrade handshake. The browser posts
// speech in and reads events back; that is the whole transport.
//
// Transplanted from interject unchanged.
package stream

import (
	"encoding/json"
	"fmt"
	"net/http"
	"sync"
)

// Event is one named message delivered to every attached browser.
type Event struct {
	Name string
	Data any
}

// Hub holds the attached browsers.
type Hub struct {
	mu      sync.Mutex
	next    int
	clients map[int]chan Event
}

func NewHub() *Hub {
	return &Hub{clients: make(map[int]chan Event)}
}

// Publish delivers to every attached browser, dropping for any client too slow
// to keep up. A stalled tab must never block the decision loop.
func (h *Hub) Publish(name string, data any) {
	h.mu.Lock()
	defer h.mu.Unlock()
	for _, client := range h.clients {
		select {
		case client <- Event{Name: name, Data: data}:
		default:
		}
	}
}

func (h *Hub) attach() (int, chan Event) {
	h.mu.Lock()
	defer h.mu.Unlock()
	h.next++
	events := make(chan Event, 32)
	h.clients[h.next] = events
	return h.next, events
}

func (h *Hub) detach(id int) {
	h.mu.Lock()
	defer h.mu.Unlock()
	delete(h.clients, id)
}

// ServeHTTP streams events until the browser goes away.
func (h *Hub) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	flusher, ok := w.(http.Flusher)
	if !ok {
		http.Error(w, "streaming unsupported", http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "text/event-stream")
	w.Header().Set("Cache-Control", "no-cache")
	w.Header().Set("Connection", "keep-alive")
	w.WriteHeader(http.StatusOK)
	flusher.Flush()

	id, events := h.attach()
	defer h.detach(id)

	for {
		select {
		case <-r.Context().Done():
			return
		case event := <-events:
			payload, err := json.Marshal(event.Data)
			if err != nil {
				continue
			}
			if _, err := fmt.Fprintf(w, "event: %s\ndata: %s\n\n", event.Name, payload); err != nil {
				return
			}
			flusher.Flush()
		}
	}
}
