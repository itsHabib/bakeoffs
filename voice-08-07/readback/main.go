// hack-readback: voice-gated irreversible actions for a simulated ops world.
// Order -> canonical readback -> spoken confirm token -> execute. Nothing
// mutates without a matching confirm; every exchange lands in the audit log.
package main

import (
	_ "embed"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"sync"

	"hack-readback/policy"
)

//go:embed world.json
var seed []byte

//go:embed ui/index.html
var indexHTML []byte

type hearRequest struct {
	Text string `json:"text"`
}

type hearResponse struct {
	Reply policy.Reply `json:"reply"`
	State policy.State `json:"state"`
}

// server holds the current gate; reset swaps in a fresh world so the canned
// demo is repeatable from the same seed.
type server struct {
	mu   sync.Mutex
	gate *policy.Gate
}

func (s *server) current() *policy.Gate {
	s.mu.Lock()
	defer s.mu.Unlock()
	return s.gate
}

func (s *server) reset() error {
	var world policy.World
	if err := json.Unmarshal(seed, &world); err != nil {
		return fmt.Errorf("bad world.json: %w", err)
	}
	s.mu.Lock()
	s.gate = policy.New(&world)
	s.mu.Unlock()
	return nil
}

func main() {
	srv := &server{}
	if err := srv.reset(); err != nil {
		log.Fatal(err)
	}

	mux := http.NewServeMux()
	mux.HandleFunc("GET /", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/html; charset=utf-8")
		w.Write(indexHTML)
	})
	mux.HandleFunc("POST /api/hear", func(w http.ResponseWriter, r *http.Request) {
		var req hearRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}
		gate := srv.current()
		reply := gate.Hear(req.Text)
		writeJSON(w, hearResponse{Reply: reply, State: gate.Snapshot()})
	})
	mux.HandleFunc("GET /api/state", func(w http.ResponseWriter, r *http.Request) {
		writeJSON(w, srv.current().Snapshot())
	})
	mux.HandleFunc("POST /api/reset", func(w http.ResponseWriter, r *http.Request) {
		if err := srv.reset(); err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		writeJSON(w, srv.current().Snapshot())
	})

	addr := "localhost:8014"
	if fromEnv := os.Getenv("READBACK_ADDR"); fromEnv != "" {
		addr = fromEnv
	}
	fmt.Printf("readback console: http://%s\n", addr)
	log.Fatal(http.ListenAndServe(addr, mux))
}

func writeJSON(w http.ResponseWriter, v any) {
	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(v); err != nil {
		log.Printf("encode: %v", err)
	}
}
