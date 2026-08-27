// Package session composes the listener, the budget, and the browser.
//
// The load-bearing trick, inherited from interject: classification runs while
// the student is still talking, not in the gap after they stop. A local model
// needs a second or two to form an opinion, far too slow to answer "speak
// now?" the moment a pause opens. So the opinion is always already there, and
// the pause only reads it.
//
// What this session adds over interject's: a ledger of everything the tutor
// chose NOT to say. Every held opinion is recorded with its reason, and when
// the student later catches their own mistake, the held entry is marked
// caught. That ledger — not the interruptions — is the end-of-session report.
package session

import (
	"context"
	"fmt"
	"log/slog"
	"strings"
	"sync"
	"time"

	"hacktutor/internal/budget"
	"hacktutor/internal/listen"
	"hacktutor/internal/stream"
)

const (
	transcriptLimit = 14              // utterances kept as context
	idleBleed       = 2 * time.Second // silence before pressure starts falling
	tickEvery       = 500 * time.Millisecond
)

// Interruption is one spent turn, kept so the listener can judge it.
type Interruption struct {
	ID        int     `json:"id"`
	Line      string  `json:"line"`
	Kind      string  `json:"kind"`
	Conf      string  `json:"conf"`
	Pressure  float64 `json:"pressure"`
	Threshold float64 `json:"threshold"`
	At        string  `json:"at"`
}

// Held is one thing the tutor noticed and chose not to say. CaughtAt is filled
// in when the student later corrects the mistake themselves — which is the
// outcome the whole ration exists to buy.
type Held struct {
	At       string `json:"at"`
	Kind     string `json:"kind"`
	Conf     string `json:"conf"`
	Heard    string `json:"heard"`  // the student line that triggered it
	Line     string `json:"line"`   // what the tutor would have asked
	Reason   string `json:"reason"` // why the budget said no
	CaughtAt string `json:"caught_at,omitempty"`
}

// Session is one student, one budget, one running problem.
type Session struct {
	log    *slog.Logger
	scorer *listen.Scorer
	hub    *stream.Hub
	cfg    budget.Config

	trigger chan struct{}

	mu         sync.Mutex
	ledger     *budget.Ledger
	start      time.Time
	transcript []string
	spoken     []string
	spent      []Interruption
	held       []Held
	pending    []Interruption
	nextID     int
	lastHeard  time.Time
	scoring    bool
}

func New(log *slog.Logger, scorer *listen.Scorer, hub *stream.Hub, cfg budget.Config) *Session {
	now := time.Now()
	return &Session{
		log:     log,
		scorer:  scorer,
		hub:     hub,
		cfg:     cfg,
		trigger: make(chan struct{}, 1),
		ledger:  budget.New(cfg, now),
		start:   now,
	}
}

// Run drives the two background loops: classify whenever there is new speech,
// and publish state on a tick so the browser sees the bar move on its own.
func (s *Session) Run(ctx context.Context) {
	ticker := time.NewTicker(tickEvery)
	defer ticker.Stop()
	for {
		select {
		case <-ctx.Done():
			return
		case <-s.trigger:
			s.classify(ctx)
		case <-ticker.C:
			s.bleed()
			s.publishState()
		}
	}
}

// Hear takes one line of the student working out loud. It returns immediately:
// the classifier runs behind it, and drops requests that arrive while it is
// busy, so a fast talker cannot queue up a backlog of stale opinions.
func (s *Session) Hear(text string) {
	text = strings.TrimSpace(text)
	if text == "" {
		return
	}
	s.mu.Lock()
	s.transcript = append(s.transcript, text)
	if len(s.transcript) > transcriptLimit {
		s.transcript = s.transcript[len(s.transcript)-transcriptLimit:]
	}
	s.lastHeard = time.Now()
	s.mu.Unlock()

	s.hub.Publish("heard", map[string]any{"text": text})
	select {
	case s.trigger <- struct{}{}:
	default:
	}
}

func (s *Session) classify(ctx context.Context) {
	s.mu.Lock()
	transcript := append([]string(nil), s.transcript...)
	spoken := append([]string(nil), s.spoken...)
	depth := len(s.transcript)
	heard := transcript[len(transcript)-1]
	s.scoring = true
	s.mu.Unlock()

	s.hub.Publish("thinking", map[string]any{"active": true})
	reading, err := s.scorer.Read(ctx, transcript, spoken)

	s.mu.Lock()
	s.scoring = false
	s.mu.Unlock()
	s.hub.Publish("thinking", map[string]any{"active": false})

	if err != nil {
		if ctx.Err() != nil {
			return
		}
		s.log.Error("classify speech", "err", err)
		s.hub.Publish("fault", map[string]any{"message": err.Error()})
		return
	}
	s.apply(reading, depth, heard)
}

// apply folds one reading into the budget and takes a turn if it is worth it.
func (s *Session) apply(reading listen.Reading, depth int, heard string) {
	kind := budget.ParseKind(reading.Kind)
	conf := budget.ParseConfidence(reading.Conf)
	raw := budget.Raw(kind, conf, depth)
	if raw > 0 && !budget.Askable(reading.Line) {
		raw = 0 // a tutor that cannot phrase it as one short question does not get a turn
	}

	now := time.Now()
	s.mu.Lock()
	s.ledger.Observe(raw)
	decision := s.ledger.Decide(now)
	interruption, spoke := s.commit(decision, reading, kind, conf, now)
	if kind == budget.KindSelfCorrect {
		s.resolveHeld(now)
	}
	if !spoke && raw > 0 {
		s.hold(now, kind, conf, heard, reading.Line, decision.Reason)
	}
	s.mu.Unlock()

	s.hub.Publish("reading", map[string]any{
		"kind": string(kind), "conf": string(conf), "line": reading.Line,
		"raw": round(raw), "latency_ms": reading.Latency.Milliseconds(),
		"decision": decision,
	})
	if spoke {
		s.log.Info("interrupted", "kind", kind, "pressure", round(decision.Pressure), "line", reading.Line)
		s.hub.Publish("interrupt", interruption)
	}
	s.publishState()
}

// commit spends the budget under the caller's lock. It returns whether the
// turn was actually taken.
func (s *Session) commit(decision budget.Decision, reading listen.Reading, kind budget.Kind, conf budget.Confidence, now time.Time) (Interruption, bool) {
	if !decision.Speak {
		return Interruption{}, false
	}
	s.nextID++
	interruption := Interruption{
		ID:        s.nextID,
		Line:      reading.Line,
		Kind:      string(kind),
		Conf:      string(conf),
		Pressure:  round(decision.Pressure),
		Threshold: round(decision.Threshold),
		At:        s.clock(now),
	}
	s.ledger.Spend(now)
	s.spoken = append(s.spoken, reading.Line)
	s.spent = append(s.spent, interruption)
	s.pending = append(s.pending, interruption)
	return interruption, true
}

// hold records a held opinion for the report, under the caller's lock.
// Consecutive readings of the same opinion collapse into one entry, so a
// student who talks past their mistake does not flood the ledger.
func (s *Session) hold(now time.Time, kind budget.Kind, conf budget.Confidence, heard, line, reason string) {
	if len(s.held) > 0 {
		last := &s.held[len(s.held)-1]
		if last.Kind == string(kind) && last.Line == line && last.CaughtAt == "" {
			last.Reason = reason
			return
		}
	}
	s.held = append(s.held, Held{
		At:     s.clock(now),
		Kind:   string(kind),
		Conf:   string(conf),
		Heard:  heard,
		Line:   line,
		Reason: reason,
	})
	s.hub.Publish("held", s.held[len(s.held)-1])
}

// resolveHeld marks the newest uncaught held mistake as caught by the
// student, under the caller's lock. The link is positional, not semantic: the
// mistake a student corrects is overwhelmingly the one they just made. Only
// mistakes resolve — getting unstuck is not "catching" anything.
func (s *Session) resolveHeld(now time.Time) {
	for i := len(s.held) - 1; i >= 0; i-- {
		if s.held[i].CaughtAt != "" {
			continue
		}
		if s.held[i].Kind != string(budget.KindSlip) && s.held[i].Kind != string(budget.KindMisconception) {
			continue
		}
		s.held[i].CaughtAt = s.clock(now)
		s.hub.Publish("caught", s.held[i])
		return
	}
}

// Verdict folds a judgement of one interruption into the bar.
func (s *Session) Verdict(id int, verdict budget.Verdict) {
	s.mu.Lock()
	kept := s.pending[:0]
	found := false
	for _, item := range s.pending {
		if item.ID == id {
			found = true
			continue
		}
		kept = append(kept, item)
	}
	s.pending = kept
	if found {
		s.ledger.Calibrate(verdict)
	}
	s.mu.Unlock()
	s.publishState()
}

// Reset starts a fresh session, keeping nothing.
func (s *Session) Reset() {
	now := time.Now()
	s.mu.Lock()
	s.ledger = budget.New(s.cfg, now)
	s.start = now
	s.transcript = nil
	s.spoken = nil
	s.spent = nil
	s.held = nil
	s.pending = nil
	s.mu.Unlock()
	s.hub.Publish("reset", map[string]any{})
	s.publishState()
}

// Report is the product: everything the tutor said, and everything it chose
// not to. A parent reads the held list and understands the app in one pass.
type Report struct {
	Elapsed string         `json:"elapsed"`
	Total   int            `json:"total"`
	Spent   []Interruption `json:"spent"`
	Held    []Held         `json:"held"`
}

// Report renders the session so far without changing it.
func (s *Session) Report() Report {
	now := time.Now()
	s.mu.Lock()
	defer s.mu.Unlock()
	return Report{
		Elapsed: s.clock(now),
		Total:   s.cfg.Total,
		Spent:   append([]Interruption{}, s.spent...), // non-nil so the JSON is [], never null
		Held:    append([]Held{}, s.held...),
	}
}

// bleed lets pressure fall when nobody is talking, so a missed opening cools
// off instead of lying in wait.
func (s *Session) bleed() {
	s.mu.Lock()
	defer s.mu.Unlock()
	if s.ledger.Pressure() <= 0 {
		return
	}
	if time.Since(s.lastHeard) < idleBleed {
		return
	}
	s.ledger.Observe(0)
}

// State is the whole visible condition of the session.
type State struct {
	Pressure  float64 `json:"pressure"`
	Threshold float64 `json:"threshold"`
	Floor     float64 `json:"floor"`
	Remaining int     `json:"remaining"`
	Total     int     `json:"total"`
	Trim      float64 `json:"trim"`
	Held      int     `json:"held"`
	ResetsIn  int     `json:"resets_in"`
	Cooldown  int     `json:"cooldown"`
	Reason    string  `json:"reason"`
	Scoring   bool    `json:"scoring"`
	Model     string  `json:"model"`
}

// Snapshot reports the current state without changing it.
func (s *Session) Snapshot() State {
	now := time.Now()
	s.mu.Lock()
	defer s.mu.Unlock()
	decision := s.ledger.Decide(now)
	return State{
		Pressure:  round(s.ledger.Pressure()),
		Threshold: round(decision.Threshold),
		Floor:     s.cfg.Floor,
		Remaining: decision.Remaining,
		Total:     s.cfg.Total,
		Trim:      round(s.ledger.Trim()),
		Held:      len(s.held),
		ResetsIn:  int(s.ledger.Resets(now).Seconds()),
		Cooldown:  int(s.cfg.Cooldown.Seconds()),
		Reason:    decision.Reason,
		Scoring:   s.scoring,
		Model:     s.scorer.Model(),
	}
}

func (s *Session) publishState() {
	s.hub.Publish("state", s.Snapshot())
}

// clock renders a moment as mm:ss into the session, the timestamps the report
// speaks in. Callers hold the lock.
func (s *Session) clock(now time.Time) string {
	elapsed := int(now.Sub(s.start).Seconds())
	if elapsed < 0 {
		elapsed = 0
	}
	return fmt.Sprintf("%d:%02d", elapsed/60, elapsed%60)
}

func round(v float64) float64 {
	return float64(int(v*10+0.5)) / 10
}
