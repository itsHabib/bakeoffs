package analyze

import (
	"math"
	"path/filepath"
	"testing"
	"time"

	"hack3-cache-max/internal/pricing"
	"hack3-cache-max/internal/transcript"
)

// testRates are FIXED constants local to the tests — never the repo's TODO
// table — so the asserted dollar deltas are deterministic regardless of the
// real prices the operator later fills in.
var testRates = pricing.Rates{
	FreshInput:   3.00,
	Output:       15.00,
	CacheRead:    0.30,
	CacheWrite5m: 3.75,
	CacheWrite1h: 6.00,
}

func load(t *testing.T, name string) SessionStats {
	t.Helper()
	s, err := transcript.ParseFile(filepath.Join("..", "..", "fixtures", name))
	if err != nil {
		t.Fatalf("parse %s: %v", name, err)
	}
	return Session(s, Default())
}

func approx(a, b float64) bool { return math.Abs(a-b) < 1e-9 }

func TestSteadySessionHasNoBusts(t *testing.T) {
	st := load(t, "steady.jsonl")
	if st.Assistants != 4 {
		t.Fatalf("assistants = %d, want 4", st.Assistants)
	}
	if len(st.Busts) != 0 {
		t.Fatalf("busts = %d, want 0: %+v", len(st.Busts), st.Busts)
	}
	// reads 15000+20000+21000+22000=78000; creates 5000+1000+1000+1000=8000; input 0.
	if st.SumRead != 78000 || st.SumCreate != 8000 {
		t.Fatalf("read=%d create=%d, want 78000/8000", st.SumRead, st.SumCreate)
	}
	want := 78000.0 / (78000.0 + 8000.0)
	if !approx(st.HitRatio(), want) {
		t.Fatalf("hit ratio = %.6f, want %.6f", st.HitRatio(), want)
	}
}

func TestBustySessionDetectsExactlyTwoBusts(t *testing.T) {
	st := load(t, "busty.jsonl")
	if len(st.Busts) != 2 {
		t.Fatalf("busts = %d, want 2: %+v", len(st.Busts), st.Busts)
	}

	b1, b2 := st.Busts[0], st.Busts[1]

	// Bust #1: prev-read 15000 -> threshold 3750; creation 25000 @5m.
	if b1.Threshold != 3750 || b1.Excess != 21250 {
		t.Fatalf("bust1 threshold=%d excess=%d, want 3750/21250", b1.Threshold, b1.Excess)
	}
	if b1.Excess5m != 21250 || b1.Excess1h != 0 {
		t.Fatalf("bust1 split 1h=%d 5m=%d, want 0/21250", b1.Excess1h, b1.Excess5m)
	}
	if b1.Cause != "tool:Bash" {
		t.Fatalf("bust1 cause = %q, want tool:Bash", b1.Cause)
	}

	// Bust #2: prev-read 25000 -> threshold 6250; creation 30000 @1h.
	if b2.Threshold != 6250 || b2.Excess != 23750 {
		t.Fatalf("bust2 threshold=%d excess=%d, want 6250/23750", b2.Threshold, b2.Excess)
	}
	if b2.Excess1h != 23750 || b2.Excess5m != 0 {
		t.Fatalf("bust2 split 1h=%d 5m=%d, want 23750/0", b2.Excess1h, b2.Excess5m)
	}
	if b2.Cause != "tool:Read" {
		t.Fatalf("bust2 cause = %q, want tool:Read", b2.Cause)
	}
}

func TestCounterfactualDollarDelta(t *testing.T) {
	st := load(t, "busty.jsonl")
	cost := SessionCost(st, testRates)

	// Savings = excess repriced from write to read rate.
	// bust1: 21250 @5m * (3.75-0.30)/1e6 = 0.0733125
	// bust2: 23750 @1h * (6.00-0.30)/1e6 = 0.135375
	wantSavings := 0.0733125 + 0.135375
	if !approx(cost.Savings(), wantSavings) {
		t.Fatalf("savings = %.9f, want %.9f", cost.Savings(), wantSavings)
	}
	if cost.Counterfactual >= cost.AsHappened {
		t.Fatalf("counterfactual %.6f should be below as-happened %.6f", cost.Counterfactual, cost.AsHappened)
	}
}

func TestCorpusRanksCausesByRewrittenTokens(t *testing.T) {
	stats := []SessionStats{load(t, "steady.jsonl"), load(t, "busty.jsonl")}
	c := Aggregate(stats, testRates)

	if c.TotalBusts != 2 {
		t.Fatalf("corpus busts = %d, want 2", c.TotalBusts)
	}
	if len(c.Causes) != 2 {
		t.Fatalf("cause rows = %d, want 2: %+v", len(c.Causes), c.Causes)
	}
	// Read bust re-wrote 30000 tokens; Bash bust re-wrote 25000 — Read ranks first.
	if c.Causes[0].Bucket != "tool:Read" || c.Causes[0].Rewritten != 30000 {
		t.Fatalf("top cause = %q/%d, want tool:Read/30000", c.Causes[0].Bucket, c.Causes[0].Rewritten)
	}
	if c.Causes[1].Bucket != "tool:Bash" || c.Causes[1].Rewritten != 25000 {
		t.Fatalf("2nd cause = %q/%d, want tool:Bash/25000", c.Causes[1].Bucket, c.Causes[1].Rewritten)
	}
}

func TestMinPrevReadFloorSuppressesColdStart(t *testing.T) {
	// A message with a tiny previous cache-read must not count as a bust, even
	// with large creation — you cannot bust a prefix that was never cached.
	s := transcript.Session{Turns: []transcript.Turn{
		{Assistant: true, Usage: transcript.Usage{CacheRead: 100, Create5m: 5000}},
		{Assistant: true, Usage: transcript.Usage{CacheRead: 200, Create5m: 90000}},
	}}
	st := Session(s, Default())
	if len(st.Busts) != 0 {
		t.Fatalf("busts = %d, want 0 (prev-read below floor)", len(st.Busts))
	}
}

func TestIdleGapAttributedToTTLExpiry(t *testing.T) {
	base := time.Date(2026, 8, 1, 12, 0, 0, 0, time.UTC)
	// Two assistant turns, no new context between them, 10-minute gap. The
	// second is a bust (creation ≫ threshold) with an empty context diff, so it
	// must be blamed on idle cache-TTL expiry, not left unattributed.
	s := transcript.Session{Turns: []transcript.Turn{
		{Assistant: true, Usage: transcript.Usage{CacheRead: 20000}, Time: base, HasTime: true},
		{Assistant: true, Usage: transcript.Usage{CacheRead: 1000, Create1h: 40000},
			Time: base.Add(10 * time.Minute), HasTime: true},
	}}
	st := Session(s, Default())
	if len(st.Busts) != 1 {
		t.Fatalf("busts = %d, want 1", len(st.Busts))
	}
	if st.Busts[0].Cause != "idle-ttl-expiry" {
		t.Fatalf("cause = %q, want idle-ttl-expiry", st.Busts[0].Cause)
	}
}

func TestNewContentBeatsIdleGap(t *testing.T) {
	base := time.Date(2026, 8, 1, 12, 0, 0, 0, time.UTC)
	// A bust with BOTH a long idle gap AND a big new tool result: the content
	// diff wins, per the brief's primary attribution.
	s := transcript.Session{Turns: []transcript.Turn{
		{Assistant: true, Usage: transcript.Usage{CacheRead: 20000}, Time: base, HasTime: true},
		{Chunks: []transcript.Chunk{{Bucket: "tool:Bash", Chars: 50000}}},
		{Assistant: true, Usage: transcript.Usage{CacheRead: 1000, Create1h: 40000},
			Time: base.Add(30 * time.Minute), HasTime: true},
	}}
	st := Session(s, Default())
	if len(st.Busts) != 1 || st.Busts[0].Cause != "tool:Bash" {
		t.Fatalf("cause = %+v, want tool:Bash", st.Busts)
	}
}

func TestExtrapolationScalesLinearly(t *testing.T) {
	stats := []SessionStats{load(t, "busty.jsonl")}
	c := Aggregate(stats, testRates)
	billable := float64(c.SumRead + c.SumCreate + c.SumInput + c.SumOutput)
	got := c.ExtrapolatedSavings(billable) // scale factor exactly 1
	if !approx(got, c.Cost.Savings()) {
		t.Fatalf("extrapolated at 1x = %.9f, want %.9f", got, c.Cost.Savings())
	}
	if !approx(c.ExtrapolatedSavings(2*billable), 2*c.Cost.Savings()) {
		t.Fatalf("extrapolation not linear")
	}
}
