// Command hack-tutor runs a math tutor that is allowed to interrupt three
// times per session, and has to decide — while the student is still working —
// whether what it has to say is worth one of them. Everything it holds back is
// the end-of-session report.
package main

import (
	"context"
	"embed"
	"encoding/json"
	"errors"
	"flag"
	"fmt"
	"io/fs"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"time"

	"hacktutor/internal/budget"
	"hacktutor/internal/listen"
	"hacktutor/internal/session"
	"hacktutor/internal/stream"
)

//go:embed web
var embedded embed.FS

func main() {
	if err := run(); err != nil {
		fmt.Fprintln(os.Stderr, "hack-tutor:", err)
		os.Exit(1)
	}
}

func run() error {
	cfg := budget.DefaultConfig()
	addr := flag.String("addr", "127.0.0.1:8731", "listen address")
	model := flag.String("model", "qwen2.5:7b", "classifier model")
	baseURL := flag.String("base-url", "http://localhost:11434/v1", "OpenAI-compatible endpoint")
	promptPath := flag.String("prompt", "prompts/classify.md", "classifier prompt, reloaded per call")
	webDir := flag.String("web", "", "serve UI from this directory instead of the embedded copy")
	flag.IntVar(&cfg.Total, "budget", cfg.Total, "interruptions allowed per session window")
	flag.DurationVar(&cfg.Window, "window", cfg.Window, "budget window")
	flag.DurationVar(&cfg.Cooldown, "cooldown", cfg.Cooldown, "how long a fresh interruption stays expensive")
	flag.Float64Var(&cfg.Base, "base", cfg.Base, "threshold when spending is on pace")
	flag.Float64Var(&cfg.Floor, "floor", cfg.Floor, "pressure below which it never speaks")
	flag.Parse()

	log := slog.New(slog.NewTextHandler(os.Stderr, &slog.HandlerOptions{Level: slog.LevelInfo}))

	scorer := listen.New(listen.Config{
		BaseURL: *baseURL,
		Model:   *model,
		APIKey:  os.Getenv("OPENAI_API_KEY"),
		Prompt:  *promptPath,
	})

	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt)
	defer stop()

	probe, cancel := context.WithTimeout(ctx, 4*time.Second)
	defer cancel()
	if err := scorer.Ready(probe); err != nil {
		return fmt.Errorf("classifier unreachable — is `ollama serve` running? %w", err)
	}

	hub := stream.NewHub()
	current := session.New(log, scorer, hub, cfg)
	go current.Run(ctx)

	assets, err := webAssets(*webDir)
	if err != nil {
		return err
	}

	mux := http.NewServeMux()
	mux.Handle("GET /", http.FileServer(http.FS(assets)))
	mux.Handle("GET /api/events", hub)
	mux.HandleFunc("GET /api/state", func(w http.ResponseWriter, r *http.Request) {
		writeJSON(w, current.Snapshot())
	})
	mux.HandleFunc("GET /api/report", func(w http.ResponseWriter, r *http.Request) {
		writeJSON(w, current.Report())
	})
	mux.HandleFunc("POST /api/hear", func(w http.ResponseWriter, r *http.Request) {
		var body struct {
			Text string `json:"text"`
		}
		if err := decode(r, &body); err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}
		current.Hear(body.Text)
		w.WriteHeader(http.StatusAccepted)
	})
	mux.HandleFunc("POST /api/verdict", func(w http.ResponseWriter, r *http.Request) {
		var body struct {
			ID      int    `json:"id"`
			Verdict string `json:"verdict"`
		}
		if err := decode(r, &body); err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}
		verdict, err := parseVerdict(body.Verdict)
		if err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}
		current.Verdict(body.ID, verdict)
		w.WriteHeader(http.StatusAccepted)
	})
	mux.HandleFunc("POST /api/reset", func(w http.ResponseWriter, r *http.Request) {
		current.Reset()
		w.WriteHeader(http.StatusAccepted)
	})

	server := &http.Server{
		Addr:              *addr,
		Handler:           mux,
		ReadHeaderTimeout: 10 * time.Second,
	}
	go func() {
		<-ctx.Done()
		shutdown, cancel := context.WithTimeout(context.Background(), 3*time.Second)
		defer cancel()
		_ = server.Shutdown(shutdown)
	}()

	log.Info("listening",
		"url", "http://"+*addr,
		"model", *model,
		"budget", fmt.Sprintf("%d per %s", cfg.Total, cfg.Window),
	)
	if err := server.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
		return fmt.Errorf("serve: %w", err)
	}
	return nil
}

// webAssets prefers a directory when one is given, so the UI can be edited
// against a running server.
func webAssets(dir string) (fs.FS, error) {
	if dir != "" {
		return os.DirFS(dir), nil
	}
	assets, err := fs.Sub(embedded, "web")
	if err != nil {
		return nil, fmt.Errorf("open embedded UI: %w", err)
	}
	return assets, nil
}

func parseVerdict(raw string) (budget.Verdict, error) {
	switch budget.Verdict(raw) {
	case budget.VerdictWorth:
		return budget.VerdictWorth, nil
	case budget.VerdictWasted:
		return budget.VerdictWasted, nil
	}
	return "", fmt.Errorf("unknown verdict %q", raw)
}

func decode(r *http.Request, into any) error {
	decoder := json.NewDecoder(http.MaxBytesReader(nil, r.Body, 1<<16))
	if err := decoder.Decode(into); err != nil {
		return fmt.Errorf("decode request: %w", err)
	}
	return nil
}

func writeJSON(w http.ResponseWriter, payload any) {
	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(payload); err != nil {
		slog.Error("write response", "err", err)
	}
}
