package policy

import (
	"strings"
	"testing"
	"time"
)

// testGate returns a gate with a controllable clock and a fixed "random"
// token source, so every path is deterministic.
func testGate() (*Gate, *time.Time) {
	now := time.Date(2026, 8, 7, 12, 0, 0, 0, time.UTC)
	g := New(testWorld())
	g.now = func() time.Time { return now }
	g.rnd = func(n int) int { return 37 % n }
	return g, &now
}

func lastEvent(g *Gate) string {
	if len(g.audit) == 0 {
		return ""
	}
	return g.audit[len(g.audit)-1].Event
}

func TestCleanOrderConfirmExecutes(t *testing.T) {
	g, _ := testGate()
	r := g.Hear("merge roxiq pr fourteen")
	if r.Event != "readback" || r.Pending == nil {
		t.Fatalf("expected readback with pending card, got %+v", r)
	}
	if r.Pending.Token != "14" || r.Pending.TokenSpoken != "one four" {
		t.Fatalf("token must be the risky digits, got %q / %q", r.Pending.Token, r.Pending.TokenSpoken)
	}
	if !strings.Contains(r.Say, "one four") {
		t.Fatalf("readback must spell digits individually, got %q", r.Say)
	}

	r = g.Hear("confirm one four")
	if r.Event != "executed" {
		t.Fatalf("matching confirm must execute, got %q (%q)", r.Event, r.Say)
	}
	repo := g.world.repo("roxiq")
	if repo.prOpen(14) || len(repo.Merged) != 3 {
		t.Fatalf("world not mutated after confirm: open=%v merged=%v", repo.OpenPRs, repo.Merged)
	}
	if r.Pending != nil {
		t.Fatal("pending must clear after execution")
	}
}

func TestConfirmAcceptsWordAndDigitForms(t *testing.T) {
	for _, phrase := range []string{"confirm one four", "confirm fourteen", "confirm 14", "confirmed one four"} {
		g, _ := testGate()
		g.Hear("merge roxiq pr fourteen")
		if r := g.Hear(phrase); r.Event != "executed" {
			t.Errorf("%q should execute, got %q", phrase, r.Event)
		}
	}
}

func TestWrongTokenNeverExecutes(t *testing.T) {
	g, _ := testGate()
	g.Hear("merge roxiq pr forty")  // the mishear: operator meant fourteen
	r := g.Hear("confirm one four") // operator confirms what they MEANT
	if r.Event != "mismatch" {
		t.Fatalf("token mismatch must not execute, got %q", r.Event)
	}
	if r.Pending == nil {
		t.Fatal("order must stay armed after a mismatch")
	}
	repo := g.world.repo("roxiq")
	if !repo.prOpen(40) || !repo.prOpen(14) {
		t.Fatalf("mismatch mutated the world: open=%v", repo.OpenPRs)
	}
}

func TestNegativeCancelsWithoutEffect(t *testing.T) {
	g, _ := testGate()
	g.Hear("merge roxiq pr forty")
	r := g.Hear("negative")
	if r.Event != "negative" || r.Pending != nil {
		t.Fatalf("negative must cancel, got %+v", r)
	}
	if !g.world.repo("roxiq").prOpen(40) {
		t.Fatal("negative mutated the world")
	}
}

func TestStaleTokenCannotArmNextOrder(t *testing.T) {
	g, _ := testGate()
	g.Hear("merge roxiq pr fourteen")
	g.Hear("negative")
	g.Hear("merge roxiq pr forty")
	r := g.Hear("confirm one four") // stale confirm from the previous order
	if r.Event != "mismatch" {
		t.Fatalf("stale token must not arm a different order, got %q", r.Event)
	}
	if !g.world.repo("roxiq").prOpen(40) {
		t.Fatal("stale confirm executed the order")
	}
}

func TestConfirmWithNothingArmed(t *testing.T) {
	g, _ := testGate()
	r := g.Hear("confirm one four")
	if r.Event != "stale" {
		t.Fatalf("confirm with nothing armed must be inert, got %q", r.Event)
	}
}

func TestArmedOrderExpires(t *testing.T) {
	g, now := testGate()
	g.Hear("merge roxiq pr fourteen")
	*now = now.Add(DefaultTTL + time.Second)
	r := g.Hear("confirm one four")
	if r.Event == "executed" {
		t.Fatal("expired order must not execute")
	}
	if !g.world.repo("roxiq").prOpen(14) {
		t.Fatal("expired order mutated the world")
	}
	found := false
	for _, e := range g.audit {
		if e.Event == "expired" {
			found = true
		}
	}
	if !found {
		t.Fatal("expiry must be audited")
	}
}

func TestSnapshotExpiresStaleOrder(t *testing.T) {
	g, now := testGate()
	g.Hear("merge roxiq pr fourteen")
	*now = now.Add(DefaultTTL + time.Second)
	s := g.Snapshot()
	if s.Pending != nil {
		t.Fatal("snapshot must expire a stale order")
	}
	if lastEvent(g) != "expired" {
		t.Fatalf("expected expired audit entry, got %q", lastEvent(g))
	}
}

func TestOtherSpeechWhileArmedIsHeldNotExecuted(t *testing.T) {
	g, _ := testGate()
	g.Hear("merge roxiq pr fourteen")
	r := g.Hear("deploy checkout to production")
	if r.Event != "held" {
		t.Fatalf("speech while armed must be held, got %q", r.Event)
	}
	if r.Pending == nil || r.Pending.Token != "14" {
		t.Fatal("original order must stay armed")
	}
	if g.world.service("checkout").Versions["production"] != "v12" {
		t.Fatal("held speech mutated the world")
	}
}

func TestClarifyDoesNotArm(t *testing.T) {
	g, _ := testGate()
	r := g.Hear("merge roxiq") // two PRs open: must clarify, not guess
	if r.Event != "clarify" || r.Pending != nil {
		t.Fatalf("ambiguous order must clarify without arming, got %+v", r)
	}
	if !strings.Contains(r.Say, "one four") || !strings.Contains(r.Say, "four zero") {
		t.Fatalf("clarify must list the candidates in spoken digits, got %q", r.Say)
	}
}

func TestDeployTokenIsVersionDigits(t *testing.T) {
	g, _ := testGate()
	r := g.Hear("deploy checkout to production")
	if r.Pending == nil || r.Pending.Token != "13" {
		t.Fatalf("deploy token must be the version digits, got %+v", r.Pending)
	}
	if !strings.Contains(r.Say, "PRODUCTION") {
		t.Fatalf("readback must stress the environment, got %q", r.Say)
	}
	r = g.Hear("confirm one three")
	if r.Event != "executed" {
		t.Fatalf("confirm one three should execute, got %q", r.Event)
	}
	svc := g.world.service("checkout")
	if svc.Versions["production"] != "v13" || svc.PrevProd != "v12" {
		t.Fatalf("deploy effect wrong: %+v", svc)
	}
}

func TestFlagTokenIsRandomChallenge(t *testing.T) {
	g, _ := testGate()
	r := g.Hear("flag dark mode on")
	if r.Pending == nil || r.Pending.Token != "47" { // 10 + 37%90 with the fixed rnd
		t.Fatalf("flag token must come from the challenge source, got %+v", r.Pending)
	}
	if g.Hear("confirm four seven").Event != "executed" {
		t.Fatal("flag confirm failed")
	}
	if !g.world.flag("dark-mode").On {
		t.Fatal("flag not flipped")
	}
}

// The audit trail is the product: a full mishear exchange must leave a
// readable record and zero world changes.
func TestMishearLeavesAuditAndNoDamage(t *testing.T) {
	g, _ := testGate()
	g.Hear("merge roxiq pr forty") // heard "forty", operator meant fourteen
	g.Hear("negative")
	events := make([]string, len(g.audit))
	for i, e := range g.audit {
		events[i] = e.Event
	}
	want := []string{"order", "negative"}
	if len(events) != len(want) || events[0] != want[0] || events[1] != want[1] {
		t.Fatalf("audit = %v, want %v", events, want)
	}
	repo := g.world.repo("roxiq")
	if len(repo.OpenPRs) != 2 || len(repo.Merged) != 2 {
		t.Fatalf("mishear exchange changed the world: %+v", repo)
	}
}
