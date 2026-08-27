// Package budget holds the policy that decides whether an interruption is
// worth one of the tutor's three turns.
//
// The classifier is only a sensor: it reports what kind of thing the student
// just said. Everything about scarcity — how many turns remain, how fast they
// are being spent, how good a reason has to be right now — is decided here, in
// deterministic code. A tutor whose restraint came from a model's mood would
// not be rationed at all.
//
// The weights are the pedagogy. A misconception poisons every future problem,
// so it outranks everything. Being stuck might deserve a nudge. A procedural
// slip sits below the floor on purpose: students catch those themselves, and
// teaching them to wait for the buzzer is the failure this whole app exists to
// avoid. Productive struggle is never worth a turn at any budget.
package budget

import (
	"fmt"
	"strings"
	"time"
)

// Kind is what the classifier heard in the student's most recent line.
type Kind string

const (
	KindMisconception Kind = "misconception"       // wrong mental model — will poison future problems
	KindStuck         Kind = "stuck"               // circling, no progress — might deserve a nudge
	KindSlip          Kind = "procedural-slip"     // sign error, dropped term — cheap, usually self-caught
	KindStruggle      Kind = "productive-struggle" // wrong but progressing — never interrupt
	KindSelfCorrect   Kind = "self-correct"        // caught their own mistake — the whole point
	KindNone          Kind = "none"
)

// Weight is the raw pressure a kind exerts at full confidence. Misconception
// clears an on-pace bar outright; stuck needs to persist before it does; a
// slip can never reach the floor however rich the budget; struggle and
// self-correction are worth zero turns by definition.
func (k Kind) Weight() float64 {
	switch k {
	case KindMisconception:
		return 90
	case KindStuck:
		return 62
	case KindSlip:
		return 34
	case KindStruggle:
		return 16
	default:
		return 0
	}
}

// Valid reports whether k is a kind the classifier is allowed to emit.
func (k Kind) Valid() bool {
	return k.Weight() > 0 || k == KindNone || k == KindSelfCorrect
}

// Confidence scales a kind's weight. A hedged misconception is worth less of
// the floor than a certain one.
type Confidence string

const (
	ConfLow  Confidence = "low"
	ConfMed  Confidence = "med"
	ConfHigh Confidence = "high"
)

func (c Confidence) Scale() float64 {
	switch c {
	case ConfHigh:
		return 1.0
	case ConfMed:
		return 0.78
	default:
		return 0.52
	}
}

// stuckDepth is how many utterances must precede a claim that the student is
// stuck before it is worth full weight. Two sentences into a problem nobody is
// stuck, they are reading it; a small model will confidently call hesitation
// stuck anyway, so an early claim is discounted to struggle weight.
const stuckDepth = 3

// Raw converts one classification into instantaneous pressure, 0-100. depth is
// how many utterances of context the classifier had.
func Raw(kind Kind, conf Confidence, depth int) float64 {
	weight := kind.Weight()
	if kind == KindStuck && depth < stuckDepth {
		weight = KindStruggle.Weight()
	}
	return weight * conf.Scale()
}

// askableWordLimit caps a tutor line. An interruption that needs a paragraph
// is a lecture, and lectures are what the budget exists to prevent.
const askableWordLimit = 16

// Askable reports whether a line is allowed to be spent: non-empty, short, and
// a question. The tutor asks, it never tells — and that is enforced here
// rather than hoped for in the prompt. An opinion the model cannot phrase as
// one short question costs the student nothing.
func Askable(line string) bool {
	line = strings.TrimSpace(line)
	if line == "" {
		return false
	}
	if !strings.HasSuffix(line, "?") {
		return false
	}
	return len(strings.Fields(line)) <= askableWordLimit
}

// Config is the operator-tunable shape of the ration.
type Config struct {
	Total    int           // interruptions allowed per window
	Window   time.Duration // the window they are allowed in
	Cooldown time.Duration // how long a fresh interruption stays expensive
	Base     float64       // threshold when spending is exactly on pace
	Floor    float64       // pressure below this never speaks, however rich
	Gain     float64       // threshold points added per interruption overspent
	Recent   float64       // threshold points added immediately after speaking
	Attack   float64       // envelope weight when pressure is rising
	Decay    float64       // envelope weight when pressure is falling
}

// DefaultConfig is tuned so one worked problem shows the whole arc. A certain
// misconception (90) clears the 56 bar on its first observation; a certain
// stuck signal (62) only clears it by persisting; a slip (34) can never reach
// the 40 floor at any confidence.
func DefaultConfig() Config {
	return Config{
		Total:    3,
		Window:   10 * time.Minute,
		Cooldown: 25 * time.Second,
		Base:     56,
		Floor:    40,
		Gain:     12,
		Recent:   25,
		Attack:   0.8,
		Decay:    0.3,
	}
}

// Verdict is the student's (or the watching adult's) judgement of an
// interruption they received.
type Verdict string

const (
	VerdictWorth  Verdict = "worth"
	VerdictWasted Verdict = "wasted"
)

// Ledger tracks one session's spending and calibration. It is not safe for
// concurrent use; the session owns the lock.
type Ledger struct {
	cfg      Config
	start    time.Time
	spent    []time.Time
	pressure float64
	trim     float64 // calibration drift from verdicts
}

func New(cfg Config, now time.Time) *Ledger {
	return &Ledger{cfg: cfg, start: now}
}

func (l *Ledger) Config() Config { return l.cfg }

// Observe folds one classification into the pressure envelope and returns the
// new smoothed value. Pressure rises fast and falls slow: a misconception
// should register immediately, and a missed opening should stay warm for a
// moment rather than vanishing between two lines of working.
func (l *Ledger) Observe(raw float64) float64 {
	weight := l.cfg.Decay
	if raw > l.pressure {
		weight = l.cfg.Attack
	}
	l.pressure = weight*raw + (1-weight)*l.pressure
	if l.pressure < 0.5 {
		l.pressure = 0
	}
	return l.pressure
}

func (l *Ledger) Pressure() float64 { return l.pressure }

// Remaining is how many interruptions are still available in the current window.
func (l *Ledger) Remaining(now time.Time) int {
	return l.cfg.Total - len(l.live(now))
}

// live returns the spends that still count against the window.
func (l *Ledger) live(now time.Time) []time.Time {
	cut := now.Add(-l.cfg.Window)
	live := make([]time.Time, 0, len(l.spent))
	for _, at := range l.spent {
		if at.After(cut) {
			live = append(live, at)
		}
	}
	return live
}

// fairShare is how much budget a tutor spending evenly would still be holding
// at this point in the window. Comparing it to what is actually left is what
// makes the ration self-pacing rather than first-come-first-served.
func (l *Ledger) fairShare(now time.Time) float64 {
	elapsed := now.Sub(l.start)
	if elapsed >= l.cfg.Window {
		return float64(l.cfg.Total)
	}
	left := 1 - elapsed.Seconds()/l.cfg.Window.Seconds()
	return float64(l.cfg.Total) * left
}

// Threshold is how good a reason has to be, right now, to be worth a turn.
// It rises as the tutor outruns its own pace and falls when it has been
// hoarding, which is what a budget is for.
//
// Having just spoken is priced in here rather than enforced as a separate
// veto. A fixed cooldown cannot tell the difference between a nudge and a
// misconception spotted eight seconds later, and would swallow both; as a
// surcharge that decays, urgency can still buy its way in early while nothing
// merely adequate can.
func (l *Ledger) Threshold(now time.Time) float64 {
	over := l.fairShare(now) - float64(l.Remaining(now))
	threshold := l.cfg.Base + l.cfg.Gain*over + l.recency(now) + l.trim
	if threshold < l.cfg.Floor {
		return l.cfg.Floor
	}
	if threshold > 100 {
		return 100
	}
	return threshold
}

// recency is the surcharge for having spoken recently, decaying to nothing
// across the cooldown.
func (l *Ledger) recency(now time.Time) float64 {
	left := l.cooldownLeft(now)
	if left <= 0 {
		return 0
	}
	return l.cfg.Recent * (left.Seconds() / l.cfg.Cooldown.Seconds())
}

// Decision explains one speak-or-stay-silent call. The reason travels with it
// because a silence nobody can account for is indistinguishable from a bug —
// and because the reasons become the end-of-session report.
type Decision struct {
	Speak     bool    `json:"speak"`
	Pressure  float64 `json:"pressure"`
	Threshold float64 `json:"threshold"`
	Remaining int     `json:"remaining"`
	Reason    string  `json:"reason"`
}

// Decide answers whether to take a turn. It never mutates; Spend commits.
func (l *Ledger) Decide(now time.Time) Decision {
	decision := Decision{
		Pressure:  l.pressure,
		Threshold: l.Threshold(now),
		Remaining: l.Remaining(now),
	}
	if decision.Remaining <= 0 {
		decision.Reason = "budget spent for this session"
		return decision
	}
	if l.pressure < l.cfg.Floor {
		decision.Reason = "below the floor — never worth a turn"
		return decision
	}
	if l.pressure < decision.Threshold {
		decision.Reason = fmt.Sprintf("worth %.0f, bar is %.0f%s", l.pressure, decision.Threshold, l.surcharge(now))
		return decision
	}
	decision.Speak = true
	decision.Reason = fmt.Sprintf("worth %.0f against a %.0f bar", l.pressure, decision.Threshold)
	return decision
}

// surcharge annotates a refusal when speaking recently is what made it too
// expensive, so a held opinion is never unexplained.
func (l *Ledger) surcharge(now time.Time) string {
	premium := l.recency(now)
	if premium < 1 {
		return ""
	}
	return fmt.Sprintf(" (+%.0f, just spoke)", premium)
}

func (l *Ledger) cooldownLeft(now time.Time) time.Duration {
	if len(l.spent) == 0 {
		return 0
	}
	last := l.spent[len(l.spent)-1]
	if wait := l.cfg.Cooldown - now.Sub(last); wait > 0 {
		return wait
	}
	return 0
}

// Spend commits one interruption and discharges the pressure that earned it.
func (l *Ledger) Spend(now time.Time) {
	l.spent = append(l.spent, now)
	l.pressure = 0
}

// Calibrate folds a verdict back into the bar. Being told an interruption was
// not worth it makes every later one more expensive.
func (l *Ledger) Calibrate(verdict Verdict) {
	switch verdict {
	case VerdictWasted:
		l.trim += 7
	case VerdictWorth:
		l.trim -= 4
	}
	l.trim = clamp(l.trim, -18, 26)
}

func (l *Ledger) Trim() float64 { return l.trim }

// Resets is when the oldest live spend ages out and budget returns.
func (l *Ledger) Resets(now time.Time) time.Duration {
	live := l.live(now)
	if len(live) == 0 {
		return 0
	}
	if wait := l.cfg.Window - now.Sub(live[0]); wait > 0 {
		return wait
	}
	return 0
}

func clamp(v, lo, hi float64) float64 {
	if v < lo {
		return lo
	}
	if v > hi {
		return hi
	}
	return v
}

// ParseKind maps classifier output onto a known kind, failing closed to
// KindNone. An unrecognised label must cost the student nothing.
func ParseKind(raw string) Kind {
	kind := Kind(strings.ToLower(strings.TrimSpace(raw)))
	if !kind.Valid() {
		return KindNone
	}
	return kind
}

// ParseConfidence maps classifier output onto a confidence, failing closed to
// the least assertive value.
func ParseConfidence(raw string) Confidence {
	switch Confidence(strings.ToLower(strings.TrimSpace(raw))) {
	case ConfHigh:
		return ConfHigh
	case ConfMed:
		return ConfMed
	default:
		return ConfLow
	}
}
