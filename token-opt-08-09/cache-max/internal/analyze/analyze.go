// Package analyze is the policy layer: cache accounting, the mechanical bust
// rule, correlational bust attribution, and the what-if repricing. Every number
// here is computed from transcript facts and priced from the one rate table;
// nothing is model-judged. This is the layer the tests pin.
package analyze

import (
	"sort"
	"time"

	"hack3-cache-max/internal/pricing"
	"hack3-cache-max/internal/transcript"
)

// Params tune the mechanical bust rule.
type Params struct {
	// BustFrac: a message busts when its cache-creation exceeds this fraction
	// of the PREVIOUS message's cache-read. Default 0.25.
	BustFrac float64
	// MinPrevRead floors bust eligibility: a message can only "bust" a prefix
	// that was actually cached. Below this many previous cache-read tokens the
	// message is skipped (avoids cold-start noise where prev-read is ~0).
	MinPrevRead int
	// IdleTTL: when a bust has no new context block, a wall-clock gap since the
	// previous message longer than this is attributed to cache-TTL expiry (the
	// session went idle and the prefix fell out of cache). Default 5m.
	IdleTTL time.Duration
}

// Default returns the brief's default parameters.
func Default() Params {
	return Params{BustFrac: 0.25, MinPrevRead: 1024, IdleTTL: 5 * time.Minute}
}

// Bust is one detected prefix-busting message.
type Bust struct {
	TurnIndex int
	Create    int // total cache-creation (re-written) tokens on this message
	Create1h  int
	Create5m  int
	PrevRead  int
	Threshold int
	Excess    int // creation above the threshold — what stability would recover
	Excess1h  int
	Excess5m  int
	Cause     string // bucket of the largest new context contributor (correlational)
}

// SessionStats is one session's accounting.
type SessionStats struct {
	Path       string
	Assistants int
	SumRead    int
	SumCreate  int
	Sum1h      int
	Sum5m      int
	SumInput   int
	SumOutput  int
	Busts      []Bust
	Duration   time.Duration
	HasSpan    bool
}

// HitRatio is cache reads over all input-side tokens (reads + writes + fresh).
func (s SessionStats) HitRatio() float64 {
	denom := s.SumRead + s.SumCreate + s.SumInput
	if denom == 0 {
		return 0
	}
	return float64(s.SumRead) / float64(denom)
}

// Session walks one parsed session in file order and applies the bust rule.
func Session(s transcript.Session, p Params) SessionStats {
	st := SessionStats{Path: s.Path}
	var (
		haveTime      bool
		first, last   time.Time
		prevRead      int
		prevTime      time.Time
		prevHasTime   bool
		seenAssistant bool
		pending       []transcript.Chunk // context since the last assistant turn
	)
	for i, t := range s.Turns {
		if !t.Assistant {
			pending = append(pending, t.Chunks...)
			continue
		}
		u := t.Usage
		st.Assistants++
		st.SumRead += u.CacheRead
		st.SumCreate += u.Create()
		st.Sum1h += u.Create1h
		st.Sum5m += u.Create5m
		st.SumInput += u.Input
		st.SumOutput += u.Output
		if t.HasTime {
			if !haveTime {
				first, haveTime = t.Time, true
			}
			last = t.Time
		}
		gap, hasGap := turnGap(prevTime, prevHasTime, t)
		if b, ok := detectBust(i, u, prevRead, pending, seenAssistant, p, gap, hasGap); ok {
			st.Busts = append(st.Busts, b)
		}
		seenAssistant = true
		prevRead = u.CacheRead
		prevTime, prevHasTime = t.Time, t.HasTime
		pending = nil
	}
	if haveTime {
		st.Duration = last.Sub(first)
		st.HasSpan = st.Duration > 0
	}
	return st
}

func detectBust(idx int, u transcript.Usage, prevRead int, pending []transcript.Chunk, seenAssistant bool, p Params, gap time.Duration, hasGap bool) (Bust, bool) {
	if !seenAssistant || prevRead < p.MinPrevRead {
		return Bust{}, false
	}
	threshold := int(p.BustFrac * float64(prevRead))
	if u.Create() <= threshold {
		return Bust{}, false
	}
	excess := u.Create() - threshold
	e1h, e5m := splitProportional(excess, u.Create1h, u.Create5m)
	return Bust{
		TurnIndex: idx,
		Create:    u.Create(),
		Create1h:  u.Create1h,
		Create5m:  u.Create5m,
		PrevRead:  prevRead,
		Threshold: threshold,
		Excess:    excess,
		Excess1h:  e1h,
		Excess5m:  e5m,
		Cause:     attribute(pending, gap, hasGap, p.IdleTTL),
	}, true
}

// turnGap is the wall-clock time between the previous assistant turn and this
// one, when both are timestamped.
func turnGap(prevTime time.Time, prevHasTime bool, t transcript.Turn) (time.Duration, bool) {
	if !prevHasTime || !t.HasTime {
		return 0, false
	}
	return t.Time.Sub(prevTime), true
}

// attribute names the bust cause. New context that entered the window is the
// primary explanation (the brief's context diff); when nothing entered, a long
// idle gap is attributed to cache-TTL expiry — the prefix fell out of cache
// while the session sat idle. Both are computed from transcript facts.
func attribute(pending []transcript.Chunk, gap time.Duration, hasGap bool, idleTTL time.Duration) string {
	if len(pending) > 0 {
		return largestContributor(pending)
	}
	if hasGap && gap >= idleTTL {
		return "idle-ttl-expiry"
	}
	return "unattributed"
}

// splitProportional apportions an excess between the 1h and 5m classes in the
// same ratio as the message's own creation split.
func splitProportional(excess, c1h, c5m int) (e1h, e5m int) {
	total := c1h + c5m
	if total == 0 {
		return excess, 0
	}
	e1h = excess * c1h / total
	return e1h, excess - e1h
}

// largestContributor returns the bucket that contributed the most characters of
// new context since the previous assistant turn. Attribution is correlational:
// it names what changed most, not a proven cause.
func largestContributor(pending []transcript.Chunk) string {
	if len(pending) == 0 {
		return "unattributed"
	}
	byBucket := map[string]int{}
	for _, c := range pending {
		byBucket[c.Bucket] += c.Chars
	}
	best, bestChars := "unattributed", -1
	for b, ch := range byBucket {
		if ch > bestChars || (ch == bestChars && b < best) {
			best, bestChars = b, ch
		}
	}
	return best
}

// Cost is the dollar cost of a set of tokens, split by class.
type Cost struct {
	AsHappened     float64 // cost as the tokens were actually billed
	Counterfactual float64 // cost if every bust's excess creation were a read
}

// Savings is the dollars stable prefixes would have recovered.
func (c Cost) Savings() float64 { return c.AsHappened - c.Counterfactual }

// SessionCost prices one session as-happened and under the counterfactual where
// each bust's excess cache-creation tokens are re-billed at the read rate.
func SessionCost(st SessionStats, r pricing.Rates) Cost {
	as := pricing.USD(st.SumRead, r.CacheRead) +
		pricing.USD(st.Sum1h, r.CacheWrite1h) +
		pricing.USD(st.Sum5m, r.CacheWrite5m) +
		pricing.USD(st.SumInput, r.FreshInput) +
		pricing.USD(st.SumOutput, r.Output)
	saved := 0.0
	for _, b := range st.Busts {
		saved += pricing.USD(b.Excess1h, r.CacheWrite1h-r.CacheRead)
		saved += pricing.USD(b.Excess5m, r.CacheWrite5m-r.CacheRead)
	}
	return Cost{AsHappened: as, Counterfactual: as - saved}
}

// Corpus aggregates every session into the headline figures.
type Corpus struct {
	Sessions   []SessionStats
	Rates      pricing.Rates
	SumRead    int
	SumCreate  int
	SumInput   int
	SumOutput  int
	TotalBusts int
	SpanHours  float64
	Cost       Cost
	Causes     []CauseRow // ranked, largest re-written tokens first
}

// CauseRow is one attributed bust-cause bucket.
type CauseRow struct {
	Bucket    string
	Busts     int
	Rewritten int // total cache-creation tokens on busts blamed on this bucket
	Sessions  int // distinct sessions where this bucket is the top cause
}

// HitRatio is the corpus-wide cache hit ratio.
func (c Corpus) HitRatio() float64 {
	denom := c.SumRead + c.SumCreate + c.SumInput
	if denom == 0 {
		return 0
	}
	return float64(c.SumRead) / float64(denom)
}

// BustsPerHour is corpus busts divided by total observed session span.
func (c Corpus) BustsPerHour() float64 {
	if c.SpanHours <= 0 {
		return 0
	}
	return float64(c.TotalBusts) / c.SpanHours
}

// Aggregate rolls session stats into corpus totals, ranks bust causes, and
// prices the whole corpus.
func Aggregate(stats []SessionStats, r pricing.Rates) Corpus {
	c := Corpus{Sessions: stats, Rates: r}
	causeTokens := map[string]int{}
	causeBusts := map[string]int{}
	causeSessions := map[string]map[string]bool{}
	for _, st := range stats {
		c.SumRead += st.SumRead
		c.SumCreate += st.SumCreate
		c.SumInput += st.SumInput
		c.SumOutput += st.SumOutput
		c.TotalBusts += len(st.Busts)
		if st.HasSpan {
			c.SpanHours += st.Duration.Hours()
		}
		sc := SessionCost(st, r)
		c.Cost.AsHappened += sc.AsHappened
		c.Cost.Counterfactual += sc.Counterfactual
		for _, b := range st.Busts {
			causeTokens[b.Cause] += b.Create
			causeBusts[b.Cause]++
			if causeSessions[b.Cause] == nil {
				causeSessions[b.Cause] = map[string]bool{}
			}
			causeSessions[b.Cause][st.Path] = true
		}
	}
	for bucket, toks := range causeTokens {
		c.Causes = append(c.Causes, CauseRow{
			Bucket:    bucket,
			Busts:     causeBusts[bucket],
			Rewritten: toks,
			Sessions:  len(causeSessions[bucket]),
		})
	}
	sort.Slice(c.Causes, func(i, j int) bool {
		if c.Causes[i].Rewritten != c.Causes[j].Rewritten {
			return c.Causes[i].Rewritten > c.Causes[j].Rewritten
		}
		return c.Causes[i].Bucket < c.Causes[j].Bucket
	})
	return c
}

// ExtrapolatedSavings scales the observed savings to a target monthly token
// volume (e.g. 5B), assuming the same cache behavior at scale. Billable tokens
// are reads + writes + fresh input + output.
func (c Corpus) ExtrapolatedSavings(targetMonthlyTokens float64) float64 {
	billable := float64(c.SumRead + c.SumCreate + c.SumInput + c.SumOutput)
	if billable <= 0 {
		return 0
	}
	return c.Cost.Savings() * (targetMonthlyTokens / billable)
}
