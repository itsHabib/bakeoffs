package main

import (
	"bytes"
	"embed"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"io/fs"
	"log"
	"net/http"
	"os"
	"strings"
	"time"

	"hack2-language-drill/internal/drill"
)

//go:embed web/*
var webFiles embed.FS

type scoreRequest struct {
	PhraseID string `json:"phrase_id"`
	Attempt  string `json:"attempt"`
}

func main() {
	assets, err := fs.Sub(webFiles, "web")
	if err != nil {
		log.Fatal(err)
	}
	mux := http.NewServeMux()
	mux.Handle("GET /", http.FileServer(http.FS(assets)))
	mux.HandleFunc("GET /api/pack", jsonHandler(drill.Pack))
	mux.HandleFunc("GET /api/canned", jsonHandler(drill.CannedFixtures))
	mux.HandleFunc("POST /api/score", score)
	mux.HandleFunc("GET /api/realtime-token", realtimeToken)

	address := "127.0.0.1:8740"
	if port := strings.TrimSpace(os.Getenv("PORT")); port != "" {
		address = "127.0.0.1:" + port
	}
	server := &http.Server{
		Addr:              address,
		Handler:           requestLog(mux),
		ReadHeaderTimeout: 5 * time.Second,
	}
	log.Printf("language drill listening on http://%s", server.Addr)
	if err := server.ListenAndServe(); !errors.Is(err, http.ErrServerClosed) {
		log.Fatal(err)
	}
}

func jsonHandler(value any) http.HandlerFunc {
	return func(response http.ResponseWriter, _ *http.Request) {
		writeJSON(response, http.StatusOK, value)
	}
}

func score(response http.ResponseWriter, request *http.Request) {
	request.Body = http.MaxBytesReader(response, request.Body, 32<<10)
	var input scoreRequest
	if err := json.NewDecoder(request.Body).Decode(&input); err != nil {
		writeJSON(response, http.StatusBadRequest, map[string]string{"error": "invalid score request"})
		return
	}
	result, err := drill.Score(input.PhraseID, input.Attempt)
	if err != nil {
		writeJSON(response, http.StatusBadRequest, map[string]string{"error": err.Error()})
		return
	}
	writeJSON(response, http.StatusOK, result)
}

func realtimeToken(response http.ResponseWriter, request *http.Request) {
	apiKey := strings.TrimSpace(os.Getenv("OPENAI_API_KEY"))
	if apiKey == "" {
		writeJSON(response, http.StatusServiceUnavailable, map[string]string{"error": "OPENAI_API_KEY is not configured on the server"})
		return
	}
	payload := map[string]any{"session": map[string]any{
		"type":              "realtime",
		"model":             "gpt-realtime-2.1",
		"instructions":      "You are a Spanish restaurant drill partner. Speak Spanish only. When asked to model a phrase, say exactly that phrase once and nothing else. In scene mode, act as a concise waiter and stay within ordinary restaurant dialogue. Never grade the learner.",
		"output_modalities": []string{"audio"},
		"audio": map[string]any{
			"input": map[string]any{
				"turn_detection": map[string]any{"type": "semantic_vad", "create_response": true, "interrupt_response": true},
			},
			"output": map[string]any{"voice": "marin"},
		},
	}}
	body, err := json.Marshal(payload)
	if err != nil {
		writeJSON(response, http.StatusInternalServerError, map[string]string{"error": "could not configure voice session"})
		return
	}
	upstreamRequest, err := http.NewRequestWithContext(request.Context(), http.MethodPost, "https://api.openai.com/v1/realtime/client_secrets", bytes.NewReader(body))
	if err != nil {
		writeJSON(response, http.StatusInternalServerError, map[string]string{"error": "could not create voice request"})
		return
	}
	upstreamRequest.Header.Set("Authorization", "Bearer "+apiKey)
	upstreamRequest.Header.Set("Content-Type", "application/json")
	upstreamRequest.Header.Set("OpenAI-Safety-Identifier", "hack2-language-drill-local-demo")
	client := &http.Client{Timeout: 20 * time.Second}
	upstream, err := client.Do(upstreamRequest)
	if err != nil {
		writeJSON(response, http.StatusBadGateway, map[string]string{"error": "voice service unavailable"})
		return
	}
	defer upstream.Body.Close()
	upstreamBody, err := io.ReadAll(io.LimitReader(upstream.Body, 1<<20))
	if err != nil {
		writeJSON(response, http.StatusBadGateway, map[string]string{"error": "invalid voice service response"})
		return
	}
	if upstream.StatusCode < 200 || upstream.StatusCode >= 300 {
		log.Printf("Realtime client secret request failed with status %d", upstream.StatusCode)
		writeJSON(response, http.StatusBadGateway, map[string]string{"error": fmt.Sprintf("voice session rejected with status %d", upstream.StatusCode)})
		return
	}
	response.Header().Set("Content-Type", "application/json")
	response.Header().Set("Cache-Control", "no-store")
	response.WriteHeader(http.StatusOK)
	_, _ = response.Write(upstreamBody)
}

func requestLog(next http.Handler) http.Handler {
	return http.HandlerFunc(func(response http.ResponseWriter, request *http.Request) {
		started := time.Now()
		next.ServeHTTP(response, request)
		log.Printf("%s %s %s", request.Method, request.URL.Path, time.Since(started).Round(time.Millisecond))
	})
}

func writeJSON(response http.ResponseWriter, status int, value any) {
	response.Header().Set("Content-Type", "application/json")
	response.Header().Set("Cache-Control", "no-store")
	response.WriteHeader(status)
	if err := json.NewEncoder(response).Encode(value); err != nil {
		log.Printf("encode JSON: %v", err)
	}
}
