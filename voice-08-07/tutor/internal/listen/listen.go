// Package listen turns a stretch of student speech into one classification.
//
// It is a sensor and nothing more: it reports what it heard and never decides
// what to do about it. Any chat endpoint speaking the OpenAI wire format will
// do, which is why a local Ollama and a hosted model need no separate clients.
//
// Transplanted from interject unchanged except for this comment.
package listen

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"os"
	"strings"
	"sync"
	"time"
)

// Reading is one classification of the most recent speech.
type Reading struct {
	Kind    string `json:"kind"`
	Conf    string `json:"conf"`
	Line    string `json:"line"`
	Latency time.Duration
}

// Config points the scorer at a model.
type Config struct {
	BaseURL string
	Model   string
	APIKey  string
	Prompt  string // path to the system prompt, reloaded per call in dev
}

// Scorer classifies speech against a chat-completions endpoint.
type Scorer struct {
	cfg    Config
	http   *http.Client
	mu     sync.Mutex
	prompt string
}

func New(cfg Config) *Scorer {
	if strings.TrimSpace(cfg.BaseURL) == "" {
		cfg.BaseURL = "http://localhost:11434/v1"
	}
	return &Scorer{
		cfg:  cfg,
		http: &http.Client{Timeout: 25 * time.Second},
	}
}

// Model reports which model backs this scorer, for the operator's benefit.
func (s *Scorer) Model() string { return s.cfg.Model }

// Ready reports whether the endpoint answers, so a dead Ollama is a startup
// error rather than a mystery silence during a demo.
func (s *Scorer) Ready(ctx context.Context) error {
	request, err := http.NewRequestWithContext(ctx, http.MethodGet, s.cfg.BaseURL+"/models", nil)
	if err != nil {
		return fmt.Errorf("build readiness request: %w", err)
	}
	s.authorize(request)
	response, err := s.http.Do(request)
	if err != nil {
		return fmt.Errorf("reach %s: %w", s.cfg.BaseURL, err)
	}
	defer func() { _ = response.Body.Close() }()
	_, _ = io.Copy(io.Discard, io.LimitReader(response.Body, 1<<16))
	if response.StatusCode < 200 || response.StatusCode >= 300 {
		return fmt.Errorf("reach %s: status %d", s.cfg.BaseURL, response.StatusCode)
	}
	return nil
}

// Read classifies the newest utterance against the ones before it. spoken
// carries lines already said aloud so the model can avoid re-raising a point
// that has been made.
func (s *Scorer) Read(ctx context.Context, transcript []string, spoken []string) (Reading, error) {
	prompt, err := s.systemPrompt()
	if err != nil {
		return Reading{}, err
	}
	user := framed(transcript, spoken)
	body, err := json.Marshal(map[string]any{
		"model": s.cfg.Model,
		"messages": []map[string]string{
			{"role": "system", "content": prompt},
			{"role": "user", "content": user},
		},
		"stream":          false,
		"temperature":     0.1,
		"max_tokens":      80,
		"keep_alive":      "30m",
		"response_format": map[string]string{"type": "json_object"},
	})
	if err != nil {
		return Reading{}, fmt.Errorf("encode classification request: %w", err)
	}
	started := time.Now()
	content, err := s.complete(ctx, body)
	if err != nil {
		return Reading{}, err
	}
	reading, err := decodeReading(content)
	if err != nil {
		return Reading{}, err
	}
	reading.Latency = time.Since(started)
	return reading, nil
}

// framed separates the newest utterance from its context. Handing the model
// one undifferentiated blob makes it keep re-reporting the most interesting
// thing anywhere in the transcript, long after that moment has passed; the
// only question is what the last thing said was.
func framed(transcript, spoken []string) string {
	if len(transcript) == 0 {
		return ""
	}
	var out strings.Builder
	if earlier := transcript[:len(transcript)-1]; len(earlier) > 0 {
		out.WriteString("Earlier in this session, for context only:\n")
		for _, line := range earlier {
			out.WriteString("  " + line + "\n")
		}
		out.WriteString("\n")
	}
	if len(spoken) > 0 {
		out.WriteString("You have already asked these aloud. Do not raise them again:\n")
		for _, line := range spoken {
			out.WriteString("  - " + line + "\n")
		}
		out.WriteString("\n")
	}
	out.WriteString("Classify ONLY this, the most recent thing the student said:\n  ")
	out.WriteString(transcript[len(transcript)-1])
	return out.String()
}

func (s *Scorer) complete(ctx context.Context, body []byte) (string, error) {
	request, err := http.NewRequestWithContext(ctx, http.MethodPost, s.cfg.BaseURL+"/chat/completions", bytes.NewReader(body))
	if err != nil {
		return "", fmt.Errorf("build classification request: %w", err)
	}
	request.Header.Set("Content-Type", "application/json")
	s.authorize(request)
	response, err := s.http.Do(request)
	if err != nil {
		return "", fmt.Errorf("classify speech: %w", err)
	}
	defer func() { _ = response.Body.Close() }()
	payload, err := io.ReadAll(io.LimitReader(response.Body, 1<<20))
	if err != nil {
		return "", fmt.Errorf("read classification: %w", err)
	}
	if response.StatusCode < 200 || response.StatusCode >= 300 {
		return "", fmt.Errorf("classify speech: status %d: %s", response.StatusCode, truncate(payload))
	}
	var decoded struct {
		Choices []struct {
			Message struct {
				Content string `json:"content"`
			} `json:"message"`
		} `json:"choices"`
	}
	if err := json.Unmarshal(payload, &decoded); err != nil {
		return "", fmt.Errorf("decode classification envelope: %w", err)
	}
	if len(decoded.Choices) == 0 {
		return "", errors.New("classify speech: model returned no choices")
	}
	return decoded.Choices[0].Message.Content, nil
}

func (s *Scorer) authorize(request *http.Request) {
	if s.cfg.APIKey == "" {
		return
	}
	request.Header.Set("Authorization", "Bearer "+s.cfg.APIKey)
}

// systemPrompt reloads the prompt file on every call so it can be tuned
// against a live session without a restart. A missing file falls back to the
// copy read at startup.
func (s *Scorer) systemPrompt() (string, error) {
	if s.cfg.Prompt == "" {
		return "", errors.New("no classifier prompt configured")
	}
	raw, err := os.ReadFile(s.cfg.Prompt)
	if err == nil {
		s.mu.Lock()
		s.prompt = string(raw)
		s.mu.Unlock()
		return string(raw), nil
	}
	s.mu.Lock()
	cached := s.prompt
	s.mu.Unlock()
	if cached == "" {
		return "", fmt.Errorf("read classifier prompt: %w", err)
	}
	return cached, nil
}

// decodeReading tolerates a model that wraps its JSON in prose or a fence,
// because a small model occasionally will.
func decodeReading(content string) (Reading, error) {
	trimmed := strings.TrimSpace(content)
	start := strings.Index(trimmed, "{")
	end := strings.LastIndex(trimmed, "}")
	if start < 0 || end <= start {
		return Reading{}, fmt.Errorf("classification held no JSON object: %q", truncate([]byte(trimmed)))
	}
	var reading Reading
	if err := json.Unmarshal([]byte(trimmed[start:end+1]), &reading); err != nil {
		return Reading{}, fmt.Errorf("decode classification: %w", err)
	}
	reading.Line = strings.TrimSpace(reading.Line)
	return reading, nil
}

func truncate(payload []byte) string {
	const limit = 180
	if len(payload) <= limit {
		return string(payload)
	}
	return string(payload[:limit]) + "..."
}
