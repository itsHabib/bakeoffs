// Package transcript is the mechanism layer: it turns a Claude Code session
// JSONL file into an ordered stream of typed turns. It carries no analysis
// policy — no bust rule, no pricing — it only reads what the usage fields and
// content entries directly state.
package transcript

import (
	"bufio"
	"encoding/json"
	"io"
	"os"
	"strings"
	"time"
)

// Usage mirrors the billable token counts on one assistant message.
type Usage struct {
	Input     int // fresh (uncached) input tokens
	CacheRead int // tokens billed at the cheap cache-read rate
	Create1h  int // ephemeral 1h cache-creation (write) tokens
	Create5m  int // ephemeral 5m cache-creation (write) tokens
	Output    int
}

// Create is the total cache-creation (write) tokens for the message.
func (u Usage) Create() int { return u.Create1h + u.Create5m }

// Chunk is one piece of content that entered the context window between two
// assistant turns, bucketed by what produced it. Chars is a size proxy (raw
// content length); it is only used to rank contributors within one bust.
type Chunk struct {
	Bucket string
	Chars  int
}

// Turn is one ordered event in a session. Assistant turns carry Usage; context
// turns carry the Chunks that entered the window before the next assistant turn.
type Turn struct {
	Assistant bool
	Usage     Usage
	Time      time.Time
	HasTime   bool
	Chunks    []Chunk
}

// Session is one transcript file parsed into file-order turns.
type Session struct {
	Path  string
	Turns []Turn
}

// raw is the subset of a JSONL line we read.
type raw struct {
	Type      string          `json:"type"`
	Timestamp string          `json:"timestamp"`
	Message   json.RawMessage `json:"message"`
}

type rawMessage struct {
	Role    string          `json:"role"`
	Content json.RawMessage `json:"content"`
	Usage   *rawUsage       `json:"usage"`
}

type rawUsage struct {
	Input          int `json:"input_tokens"`
	CacheCreation  int `json:"cache_creation_input_tokens"`
	CacheRead      int `json:"cache_read_input_tokens"`
	Output         int `json:"output_tokens"`
	CacheCreationO struct {
		Eph1h int `json:"ephemeral_1h_input_tokens"`
		Eph5m int `json:"ephemeral_5m_input_tokens"`
	} `json:"cache_creation"`
}

type contentItem struct {
	Type      string          `json:"type"`
	Name      string          `json:"name"`        // tool_use
	ID        string          `json:"id"`          // tool_use
	ToolUseID string          `json:"tool_use_id"` // tool_result
	Text      string          `json:"text"`        // text
	Content   json.RawMessage `json:"content"`     // tool_result payload (string or list)
}

// ParseFile reads one session transcript. Unparseable lines are skipped rather
// than failing the whole file — real corpora carry mixed record types.
func ParseFile(path string) (Session, error) {
	f, err := os.Open(path)
	if err != nil {
		return Session{}, err
	}
	defer f.Close()
	s := Session{Path: path}
	// tool_use_id -> tool name, populated as assistant tool_use entries appear.
	toolName := map[string]string{}

	sc := bufio.NewScanner(f)
	sc.Buffer(make([]byte, 0, 1024*1024), 32*1024*1024)
	for sc.Scan() {
		line := sc.Bytes()
		if len(line) == 0 {
			continue
		}
		var r raw
		if json.Unmarshal(line, &r) != nil {
			continue
		}
		var m rawMessage
		if len(r.Message) > 0 {
			_ = json.Unmarshal(r.Message, &m)
		}
		if r.Type == "assistant" && m.Usage != nil {
			registerToolNames(m.Content, toolName)
			s.Turns = append(s.Turns, assistantTurn(r, m))
			continue
		}
		// Everything else is potential context: user turns, tool results,
		// attachments. Measure what entered the window.
		if chunks := contextChunks(r, m, toolName, len(line)); len(chunks) > 0 {
			s.Turns = append(s.Turns, Turn{Chunks: chunks})
		}
	}
	if err := sc.Err(); err != nil && err != io.EOF {
		return s, err
	}
	return s, nil
}

func assistantTurn(r raw, m rawMessage) Turn {
	u := Usage{
		Input:     m.Usage.Input,
		CacheRead: m.Usage.CacheRead,
		Create1h:  m.Usage.CacheCreationO.Eph1h,
		Create5m:  m.Usage.CacheCreationO.Eph5m,
		Output:    m.Usage.Output,
	}
	// Fall back to the flat total if the 1h/5m split is absent.
	if u.Create1h == 0 && u.Create5m == 0 && m.Usage.CacheCreation > 0 {
		u.Create5m = m.Usage.CacheCreation
	}
	t := Turn{Assistant: true, Usage: u}
	if ts, err := time.Parse(time.RFC3339, r.Timestamp); err == nil {
		t.Time, t.HasTime = ts, true
	}
	return t
}

func registerToolNames(content json.RawMessage, toolName map[string]string) {
	for _, it := range decodeItems(content) {
		if it.Type == "tool_use" && it.ID != "" {
			toolName[it.ID] = it.Name
		}
	}
}

func contextChunks(r raw, m rawMessage, toolName map[string]string, rawLen int) []Chunk {
	if r.Type == "attachment" {
		return []Chunk{{Bucket: "attachment", Chars: rawLen}}
	}
	items := decodeItems(m.Content)
	if len(items) == 0 {
		// A plain string user message.
		var s string
		if json.Unmarshal(m.Content, &s) == nil && s != "" {
			return []Chunk{{Bucket: userBucket(s), Chars: len(s)}}
		}
		return nil
	}
	var out []Chunk
	for _, it := range items {
		switch it.Type {
		case "tool_result":
			name := toolName[it.ToolUseID]
			if name == "" {
				name = "?"
			}
			out = append(out, Chunk{Bucket: "tool:" + name, Chars: payloadChars(it.Content)})
		case "text":
			if it.Text != "" {
				out = append(out, Chunk{Bucket: userBucket(it.Text), Chars: len(it.Text)})
			}
		}
	}
	return out
}

func userBucket(s string) string {
	if strings.Contains(s, "system-reminder") {
		return "system-reminder"
	}
	return "user-message"
}

// payloadChars measures a tool_result payload whether it is a string or a list
// of content blocks.
func payloadChars(raw json.RawMessage) int {
	if len(raw) == 0 {
		return 0
	}
	var s string
	if json.Unmarshal(raw, &s) == nil {
		return len(s)
	}
	return len(raw)
}

func decodeItems(raw json.RawMessage) []contentItem {
	if len(raw) == 0 {
		return nil
	}
	var items []contentItem
	if json.Unmarshal(raw, &items) == nil {
		return items
	}
	return nil
}
