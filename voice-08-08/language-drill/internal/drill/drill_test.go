package drill

import "testing"

func TestPackHasFortyPhrasesWithTraps(t *testing.T) {
	if len(Pack) != 40 {
		t.Fatalf("pack length = %d, want 40", len(Pack))
	}
	for _, phrase := range Pack {
		if len(phrase.Traps) < 2 {
			t.Errorf("phrase %q has %d traps, want at least 2", phrase.ID, len(phrase.Traps))
		}
	}
}

func TestScore(t *testing.T) {
	tests := []struct {
		name       string
		phraseID   string
		attempt    string
		wantPass   bool
		wantTrap   string
		minimum    float64
		matchedOne string
	}{
		{name: "clean punctuation", phraseID: "check", attempt: "¡La cuenta, por favor!", wantPass: true, minimum: 1},
		{name: "accent tolerant", phraseID: "coffee", attempt: "un cafe solo por favor", wantPass: true, minimum: 1},
		{name: "accepted variant", phraseID: "table-two", attempt: "mesa para dos por favor", wantPass: true, minimum: 1, matchedOne: "mesa para dos, por favor"},
		{name: "gender trap", phraseID: "check", attempt: "el cuenta por favor", wantTrap: "gender article", minimum: .7},
		{name: "English carry-over", phraseID: "check", attempt: "el check por favor", wantTrap: "English carry-over"},
		{name: "agreement trap", phraseID: "vegetarian", attempt: "tienen opciones vegetarianos", wantTrap: "agreement", minimum: .6},
		{name: "missing politeness", phraseID: "check", attempt: "la cuenta", wantTrap: "politeness", minimum: .45},
		{name: "total miss", phraseID: "sparkling", attempt: "queremos pagar ahora", wantPass: false},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			got, err := Score(test.phraseID, test.attempt)
			if err != nil {
				t.Fatal(err)
			}
			if got.Passed != test.wantPass {
				t.Errorf("Passed = %v, want %v (score %.2f, traps %#v)", got.Passed, test.wantPass, got.Score, got.DetectedTrap)
			}
			if got.Score < test.minimum {
				t.Errorf("Score = %.2f, want at least %.2f", got.Score, test.minimum)
			}
			if test.matchedOne != "" && got.Matched != test.matchedOne {
				t.Errorf("Matched = %q, want %q", got.Matched, test.matchedOne)
			}
			if test.wantTrap != "" && !hasTrap(got, test.wantTrap) {
				t.Errorf("traps = %#v, want %q", got.DetectedTrap, test.wantTrap)
			}
		})
	}
}

func TestScoreRejectsBadInput(t *testing.T) {
	if _, err := Score("missing", ""); err == nil {
		t.Fatal("empty attempt unexpectedly accepted")
	}
	if _, err := Score("not-a-phrase", "hola"); err == nil {
		t.Fatal("unknown phrase unexpectedly accepted")
	}
}

func hasTrap(result Result, name string) bool {
	for _, trap := range result.DetectedTrap {
		if trap.Name == name {
			return true
		}
	}
	return false
}
