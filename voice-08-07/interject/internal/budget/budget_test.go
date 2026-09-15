package budget

import (
	"testing"
	"time"
)

var epoch = time.Date(2026, 8, 2, 9, 0, 0, 0, time.UTC)

func testConfig() Config {
	cfg := DefaultConfig()
	cfg.Attack = 1 // step straight to the raw value so cases read plainly
	cfg.Decay = 1
	return cfg
}

func TestRaw(t *testing.T) {
	const deep = 12
	cases := []struct {
		name  string
		kind  Kind
		conf  Confidence
		depth int
		want  float64
	}{
		{"certain error", KindError, ConfHigh, deep, 88},
		{"hedged error", KindError, ConfLow, deep, 45.76},
		{"certain contradiction", KindContradiction, ConfHigh, deep, 80},
		{"medium blindspot", KindBlindspot, ConfMed, deep, 48.36},
		{"circling", KindCircling, ConfHigh, deep, 48},
		{"nothing said", KindNone, ConfHigh, deep, 0},
		{"confident nothing is still nothing", KindNone, ConfLow, deep, 0},
		{"contradiction with nothing to contradict", KindContradiction, ConfHigh, 1, 62},
		{"contradiction just past the depth", KindContradiction, ConfHigh, contextDepth, 80},
		{"depth does not discount an error", KindError, ConfHigh, 1, 88},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			if got := Raw(tc.kind, tc.conf, tc.depth); !near(got, tc.want) {
				t.Fatalf("Raw(%s,%s,%d) = %.2f, want %.2f", tc.kind, tc.conf, tc.depth, got, tc.want)
			}
		})
	}
}

// A shallow contradiction must not clear an untouched bar, because that is
// exactly the mistake a small model makes at the start of a session.
func TestShallowContradictionStaysSilent(t *testing.T) {
	ledger := New(DefaultConfig(), epoch)
	ledger.Observe(Raw(KindContradiction, ConfHigh, 1))
	if got := ledger.Decide(epoch); got.Speak {
		t.Fatalf("spoke on an unfounded contradiction: pressure %.1f, threshold %.1f", got.Pressure, got.Threshold)
	}
	deep := New(DefaultConfig(), epoch)
	deep.Observe(Raw(KindContradiction, ConfHigh, 8))
	if got := deep.Decide(epoch); !got.Speak {
		t.Fatalf("stayed silent on a founded contradiction: pressure %.1f, threshold %.1f", got.Pressure, got.Threshold)
	}
}

func TestEnvelopeRisesFastAndFallsSlow(t *testing.T) {
	ledger := New(DefaultConfig(), epoch)
	if got := ledger.Observe(100); got < 79 {
		t.Fatalf("attack too slow: reached %.1f on first observation", got)
	}
	peak := ledger.Pressure()
	after := ledger.Observe(0)
	if after >= peak*0.75 {
		t.Fatalf("decay too slow: %.1f from peak %.1f", after, peak)
	}
	if after <= 0 {
		t.Fatalf("decay too fast: pressure vanished to %.1f in one step", after)
	}
}

func TestDecide(t *testing.T) {
	cases := []struct {
		name      string
		raw       float64
		spends    []time.Duration // offsets from epoch of prior interruptions
		at        time.Duration
		wantSpeak bool
	}{
		{"strong reason, full budget", 88, nil, 0, true},
		{"weak reason, full budget", 30, nil, 0, false},
		{"mid reason clears an on-pace bar", 62, nil, 0, true},
		{"same mid reason blocked once overspent", 62, []time.Duration{time.Minute, 2 * time.Minute}, 3 * time.Minute, false},
		{"strong reason still lands when overspent", 88, []time.Duration{time.Minute, 2 * time.Minute}, 3 * time.Minute, true},
		{"budget exhausted refuses everything", 100, []time.Duration{0, time.Minute, 2 * time.Minute}, 3 * time.Minute, false},
		{"speaking again immediately is too expensive", 88, []time.Duration{0}, 2 * time.Second, false},
		{"a merely adequate reason waits out the surcharge", 62, []time.Duration{0}, 12 * time.Second, false},
		{"an urgent reason buys in before the surcharge clears", 96, []time.Duration{0}, 12 * time.Second, true},
		{"the surcharge decays away", 88, []time.Duration{0}, 30 * time.Second, true},
		{"budget returns after the window", 88, []time.Duration{0, time.Minute, 2 * time.Minute}, 11 * time.Minute, true},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			ledger := New(testConfig(), epoch)
			for _, offset := range tc.spends {
				ledger.Spend(epoch.Add(offset))
			}
			ledger.Observe(tc.raw)
			got := ledger.Decide(epoch.Add(tc.at))
			if got.Speak != tc.wantSpeak {
				t.Fatalf("Speak = %v, want %v (pressure %.0f, threshold %.0f, remaining %d, %q)",
					got.Speak, tc.wantSpeak, got.Pressure, got.Threshold, got.Remaining, got.Reason)
			}
			if got.Reason == "" {
				t.Fatal("decision carried no reason")
			}
		})
	}
}

func TestThresholdRisesWithOverspending(t *testing.T) {
	ledger := New(testConfig(), epoch)
	onPace := ledger.Threshold(epoch)
	ledger.Spend(epoch.Add(time.Minute))
	ledger.Spend(epoch.Add(2 * time.Minute))
	overspent := ledger.Threshold(epoch.Add(2 * time.Minute))
	if overspent <= onPace {
		t.Fatalf("threshold did not rise after overspending: %.1f then %.1f", onPace, overspent)
	}
}

func TestThresholdFallsWhenHoarding(t *testing.T) {
	ledger := New(testConfig(), epoch)
	start := ledger.Threshold(epoch)
	// Two thirds of the window gone with nothing spent: the bar should relax.
	hoarding := ledger.Threshold(epoch.Add(400 * time.Second))
	if hoarding >= start {
		t.Fatalf("threshold did not relax while hoarding: %.1f then %.1f", start, hoarding)
	}
	if hoarding < testConfig().Floor {
		t.Fatalf("threshold %.1f fell through the floor", hoarding)
	}
}

func TestRecencySurchargeDecays(t *testing.T) {
	ledger := New(testConfig(), epoch)
	ledger.Spend(epoch)
	fresh := ledger.Threshold(epoch)
	mid := ledger.Threshold(epoch.Add(12 * time.Second))
	cleared := ledger.Threshold(epoch.Add(30 * time.Second))
	if !(fresh > mid && mid > cleared) {
		t.Fatalf("surcharge did not decay: %.1f then %.1f then %.1f", fresh, mid, cleared)
	}
}

func TestFloorHoldsAgainstAnEmptyWindow(t *testing.T) {
	ledger := New(testConfig(), epoch)
	ledger.Observe(Raw(KindCircling, ConfLow, 9)) // 24.96, real but trivial
	got := ledger.Decide(epoch.Add(9 * time.Minute))
	if got.Speak {
		t.Fatalf("spoke below the floor with budget to burn: pressure %.1f, threshold %.1f", got.Pressure, got.Threshold)
	}
}

func TestSpendDischargesPressure(t *testing.T) {
	ledger := New(testConfig(), epoch)
	ledger.Observe(90)
	ledger.Spend(epoch)
	if ledger.Pressure() != 0 {
		t.Fatalf("pressure survived the spend: %.1f", ledger.Pressure())
	}
}

func TestCalibrate(t *testing.T) {
	cases := []struct {
		name     string
		verdicts []Verdict
		wantUp   bool
	}{
		{"wasted raises the bar", []Verdict{VerdictWasted}, true},
		{"worth lowers the bar", []Verdict{VerdictWorth}, false},
		{"repeated waste compounds", []Verdict{VerdictWasted, VerdictWasted}, true},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			ledger := New(testConfig(), epoch)
			before := ledger.Threshold(epoch)
			for _, verdict := range tc.verdicts {
				ledger.Calibrate(verdict)
			}
			after := ledger.Threshold(epoch)
			if tc.wantUp && after <= before {
				t.Fatalf("threshold %.1f did not rise from %.1f", after, before)
			}
			if !tc.wantUp && after >= before {
				t.Fatalf("threshold %.1f did not fall from %.1f", after, before)
			}
		})
	}
}

func TestCalibrationIsBounded(t *testing.T) {
	ledger := New(testConfig(), epoch)
	for range 40 {
		ledger.Calibrate(VerdictWasted)
	}
	if ledger.Trim() > 26 {
		t.Fatalf("calibration ran away to %.1f", ledger.Trim())
	}
	for range 80 {
		ledger.Calibrate(VerdictWorth)
	}
	if ledger.Trim() < -18 {
		t.Fatalf("calibration ran away to %.1f", ledger.Trim())
	}
}

func TestRemainingRecoversWithTheWindow(t *testing.T) {
	ledger := New(testConfig(), epoch)
	ledger.Spend(epoch)
	ledger.Spend(epoch.Add(time.Second))
	if got := ledger.Remaining(epoch.Add(time.Minute)); got != 1 {
		t.Fatalf("Remaining = %d, want 1", got)
	}
	if got := ledger.Remaining(epoch.Add(11 * time.Minute)); got != 3 {
		t.Fatalf("Remaining after the window = %d, want 3", got)
	}
}

func TestParseFailsClosed(t *testing.T) {
	cases := []struct {
		raw      string
		wantKind Kind
	}{
		{"error", KindError},
		{"  ERROR ", KindError},
		{"contradiction", KindContradiction},
		{"urgent", KindNone},
		{"", KindNone},
	}
	for _, tc := range cases {
		t.Run(tc.raw, func(t *testing.T) {
			if got := ParseKind(tc.raw); got != tc.wantKind {
				t.Fatalf("ParseKind(%q) = %q, want %q", tc.raw, got, tc.wantKind)
			}
		})
	}
	for _, raw := range []string{"", "certain", "HIGH!"} {
		if got := ParseConfidence(raw); got != ConfLow {
			t.Fatalf("ParseConfidence(%q) = %q, want low", raw, got)
		}
	}
}

func near(a, b float64) bool {
	diff := a - b
	return diff < 0.01 && diff > -0.01
}
