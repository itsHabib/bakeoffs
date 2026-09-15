// Package budget holds the policy that decides whether speaking is worth it.
//
// The classifier is only a sensor: it reports what it heard. Everything about
// scarcity — how many interruptions remain, how fast they are being spent, how
// good a reason has to be right now — is decided here, in deterministic code.
// That split is the point. A rationed speaker whose restraint came from a
// model's mood would not be rationed at all.
package budget

import (
	"fmt"
	"strings"
	"time"
)

// Kind is what the classifier heard in the most recent speech.
type Kind string

const (
	KindError         Kind = "error"
	KindContradiction Kind = "contradiction"
	KindBlindspot     Kind = "blindspot"
	KindCircling      Kind = "circling"
	KindNone          Kind = "none"
)

// Weight is the raw pressure a kind exerts at full confidence. These are the
// house's opinion about what is worth a person's train of thought: a mistake
// that costs money outranks a case they merely have not reached yet.
func (k Kind) Weight() float64 {
	switch k {
	case KindError:
		return 88
	case KindContradiction:
		return 80
	case KindBlindspot:
		return 62
	case KindCircling:
		return 48
	default:
		return 0
	}
}

// Valid reports whether k is a kind the classifier is allowed to emit.
func (k Kind) Valid() bool {
	return k.Weight() > 0 || k == KindNone
}

// Confidence scales a kind's weight. A hedged contradiction is worth less of
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

// contextDepth is how many utterances must precede a claim of self-
// contradiction before it is worth full weight. A contradiction asserted two
// sentences into a session has almost nothing to contradict, and a small model
// will confidently invent one; it is treated as a blindspot until there is a
// transcript to point at.
const contextDepth = 4

// Raw converts one classification into instantaneous pressure, 0-100. depth is
// how many utterances of context the classifier had.
func Raw(kind Kind, conf Confidence, depth int) float64 {
	weight := kind.Weight()
	if kind == KindContradiction && depth < contextDepth {
		weight = KindBlindspot.Weight()
	}
	return weight * conf.Scale()
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

// DefaultConfig is tuned so the ration is visible inside a short sitting. A
// realistic deployment would be nearer three per hour; three per ten minutes
// is the same policy at demo speed.
func DefaultConfig() Config {
	return Config{
		Total:    3,
		Window:   10 * time.Minute,
		Cooldown: 25 * time.Second,
		Base:     56,
		Floor:    44,
		Gain:     12,
		Recent:   22,
		Attack:   0.8,
		Decay:    0.3,
	}
}

// Verdict is the listener's judgement of an interruption they received.
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
	trim     float64 // calibration drift from listener verdicts
}

func New(cfg Config, now time.Time) *Ledger {
	return &Ledger{cfg: cfg, start: now}
}

func (l *Ledger) Config() Config { return l.cfg }

// Observe folds one classification into the pressure envelope and returns the
// new smoothed value. Pressure rises fast and falls slow: something worth
// saying should register immediately, and a missed opening should stay warm
// for a moment rather than vanishing between two words.
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

// fairShare is how much budget a speaker spending evenly would still be
// holding at this point in the window. Comparing it to what is actually left
// is what makes the ration self-pacing rather than first-come-first-served.
func (l *Ledger) fairShare(now time.Time) float64 {
	elapsed := now.Sub(l.start)
	if elapsed >= l.cfg.Window {
		return float64(l.cfg.Total)
	}
	left := 1 - elapsed.Seconds()/l.cfg.Window.Seconds()
	return float64(l.cfg.Total) * left
}

// Threshold is how good a reason has to be, right now, to be worth saying.
// It rises as the speaker outruns its own pace and falls when it has been
// hoarding, which is what a budget is for.
//
// Having just spoken is priced in here rather than enforced as a separate
// veto. A fixed cooldown cannot tell the difference between a trivial remark
// and a serious mistake spotted eight seconds later, and would swallow both;
// as a surcharge that decays, urgency can still buy its way in early while
// nothing merely adequate can.
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
// because a silence nobody can account for is indistinguishable from a bug.
type Decision struct {
	Speak     bool    `json:"speak"`
	Pressure  float64 `json:"pressure"`
	Threshold float64 `json:"threshold"`
	Remaining int     `json:"remaining"`
	Reason    string  `json:"reason"`
}

// Decide answers whether to take the floor. It never mutates; Spend commits.
func (l *Ledger) Decide(now time.Time) Decision {
	decision := Decision{
		Pressure:  l.pressure,
		Threshold: l.Threshold(now),
		Remaining: l.Remaining(now),
	}
	if decision.Remaining <= 0 {
		decision.Reason = "budget spent for this window"
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

// Calibrate folds the listener's judgement back into the bar. Being told an
// interruption was not worth it makes every later one more expensive.
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
// KindNone. An unrecognised label must cost the listener nothing.
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
