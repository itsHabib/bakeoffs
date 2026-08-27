package policy

import (
	"fmt"
	"math/rand/v2"
	"sync"
	"time"
)

// DefaultTTL is how long an armed order waits for its confirm before it
// expires harmlessly.
const DefaultTTL = 25 * time.Second

// Gate is the readback/hearback state machine. One order may be armed at a
// time; while armed, ONLY the exact confirm token or a negative is accepted.
// Nothing mutates the world except a confirm whose token matches.
type Gate struct {
	mu      sync.Mutex
	world   *World
	now     func() time.Time
	rnd     func(n int) int
	ttl     time.Duration
	pending *pending
	audit   []Entry
}

type pending struct {
	order    Order
	token    int
	deadline time.Time
}

// Entry is one line of the audit trail: what was heard, what the gate did
// about it, and what (if anything) changed.
type Entry struct {
	TS     time.Time `json:"ts"`
	Heard  string    `json:"heard,omitempty"`
	Event  string    `json:"event"`
	Detail string    `json:"detail"`
}

// Reply is what the gate says and shows in response to one utterance.
type Reply struct {
	Say     string       `json:"say"`
	Event   string       `json:"event"`
	Pending *PendingView `json:"pending"`
}

// PendingView is the armed-order card for the UI.
type PendingView struct {
	Canonical   string      `json:"canonical"`
	Fields      [][2]string `json:"fields"`
	Token       string      `json:"token"`
	TokenSpoken string      `json:"token_spoken"`
	Deadline    time.Time   `json:"deadline"`
	TTLSeconds  int         `json:"ttl_seconds"`
}

// State is the full snapshot the UI polls.
type State struct {
	World   *World       `json:"world"`
	Pending *PendingView `json:"pending"`
	Audit   []Entry      `json:"audit"`
}

func New(w *World) *Gate {
	return &Gate{
		world: w,
		now:   time.Now,
		rnd:   rand.IntN,
		ttl:   DefaultTTL,
	}
}

// Hear processes one utterance — typed or transcribed, the path is
// identical — and returns what to say back.
func (g *Gate) Hear(text string) Reply {
	g.mu.Lock()
	defer g.mu.Unlock()
	g.expireLocked()
	toks := Normalize(text)
	if g.pending != nil {
		return g.hearArmed(text, toks)
	}
	return g.hearIdle(text, toks)
}

func (g *Gate) hearArmed(heard string, toks []string) Reply {
	p := g.pending
	if isNegative(toks) {
		g.log(heard, "negative", "cancelled: "+p.order.Canonical()+" — nothing executed")
		g.pending = nil
		return g.reply("negative. order cancelled, nothing executed.", "negative")
	}
	if len(toks) > 0 && isConfirmWord(toks[0]) {
		return g.hearConfirm(heard, toks[1:])
	}
	g.log(heard, "held", "ignored while an order is armed")
	return g.reply(fmt.Sprintf("an order is armed: %s. say, confirm %s. or say, negative.",
		p.order.Spoken(), spokenDigits(p.token)), "held")
}

func (g *Gate) hearConfirm(heard string, numToks []string) Reply {
	p := g.pending
	n, ok := numberFrom(numToks)
	if !ok {
		g.log(heard, "mismatch", "confirm without a readable token")
		return g.reply(fmt.Sprintf("say, confirm %s. or say, negative.", spokenDigits(p.token)), "mismatch")
	}
	if n != p.token {
		g.log(heard, "mismatch", fmt.Sprintf("confirm %s does not match token %s — not executed",
			spokenDigits(n), spokenDigits(p.token)))
		return g.reply(fmt.Sprintf("no match. armed order is %s. say, confirm %s. or say, negative.",
			p.order.Spoken(), spokenDigits(p.token)), "mismatch")
	}
	effect, err := g.world.Apply(p.order)
	g.pending = nil
	if err != nil {
		g.log(heard, "reject", "confirm matched but world refused: "+err.Error())
		return g.reply("unable. "+err.Error()+". order dropped.", "reject")
	}
	g.log(heard, "confirm", "token "+spokenDigits(p.token)+" matched")
	g.log("", "executed", effect)
	return g.reply("confirmed. "+effect, "executed")
}

func (g *Gate) hearIdle(heard string, toks []string) Reply {
	if isNegative(toks) || (len(toks) > 0 && isConfirmWord(toks[0])) {
		g.log(heard, "stale", "confirm or negative with no order armed — nothing to arm")
		return g.reply("no order is armed. nothing to confirm.", "stale")
	}
	res := Parse(g.world, toks)
	switch res.Outcome {
	case Clarified:
		g.log(heard, "clarify", res.Say)
		return g.reply(res.Say, "clarify")
	case Rejected:
		g.log(heard, "reject", res.Say)
		return g.reply(res.Say, "reject")
	}
	token := res.Order.Token(g.rnd)
	g.pending = &pending{order: res.Order, token: token, deadline: g.now().Add(g.ttl)}
	readback := res.Order.Readback(token)
	g.log(heard, "order", fmt.Sprintf("armed: %s · token %s · readback: %q",
		res.Order.Canonical(), spokenDigits(token), readback))
	return g.reply(readback, "readback")
}

// Snapshot expires any stale order, then returns the full state.
func (g *Gate) Snapshot() State {
	g.mu.Lock()
	defer g.mu.Unlock()
	g.expireLocked()
	return State{World: g.world, Pending: g.pendingView(), Audit: g.audit}
}

func (g *Gate) expireLocked() {
	if g.pending == nil {
		return
	}
	if g.now().Before(g.pending.deadline) {
		return
	}
	g.log("", "expired", g.pending.order.Canonical()+" expired unconfirmed — nothing executed")
	g.pending = nil
}

func (g *Gate) log(heard, event, detail string) {
	g.audit = append(g.audit, Entry{TS: g.now(), Heard: heard, Event: event, Detail: detail})
}

func (g *Gate) reply(say, event string) Reply {
	return Reply{Say: say, Event: event, Pending: g.pendingView()}
}

func (g *Gate) pendingView() *PendingView {
	if g.pending == nil {
		return nil
	}
	p := g.pending
	return &PendingView{
		Canonical:   p.order.Canonical(),
		Fields:      p.order.Fields(),
		Token:       fmt.Sprintf("%d", p.token),
		TokenSpoken: spokenDigits(p.token),
		Deadline:    p.deadline,
		TTLSeconds:  int(g.ttl / time.Second),
	}
}

func isNegative(toks []string) bool {
	if len(toks) == 0 {
		return false
	}
	switch toks[0] {
	case "negative", "cancel", "abort", "disregard", "no":
		return true
	}
	return false
}

func isConfirmWord(tok string) bool {
	return tok == "confirm" || tok == "confirmed" || tok == "affirm"
}
