package policy

import "testing"

func TestNumberFrom(t *testing.T) {
	tests := []struct {
		in   string
		want int
		ok   bool
	}{
		{"fourteen", 14, true},
		{"forty", 40, true},
		{"one four", 14, true},
		{"four zero", 40, true},
		{"four oh", 40, true},
		{"40", 40, true},
		{"1 4", 14, true},
		{"forty five", 45, true},
		{"seven", 7, true},
		{"oh seven", 7, true},
		{"ninety nine", 99, true},
		{"zero", 0, true},
		{"one hundred", 0, false},
		{"banana", 0, false},
		{"", 0, false},
		{"one four banana", 0, false},
	}
	for _, tt := range tests {
		got, ok := numberFrom(Normalize(tt.in))
		if ok != tt.ok || got != tt.want {
			t.Errorf("numberFrom(%q) = %d, %v; want %d, %v", tt.in, got, ok, tt.want, tt.ok)
		}
	}
}

func TestSpokenDigits(t *testing.T) {
	tests := []struct {
		in   int
		want string
	}{
		{14, "one four"},
		{40, "four zero"},
		{7, "seven"},
		{0, "zero"},
		{113, "one one three"},
	}
	for _, tt := range tests {
		if got := spokenDigits(tt.in); got != tt.want {
			t.Errorf("spokenDigits(%d) = %q, want %q", tt.in, got, tt.want)
		}
	}
}

// The round trip is the safety property: however the readback spells a
// number, saying that spelling back must parse to the same number.
func TestReadbackRoundTrip(t *testing.T) {
	for _, n := range []int{0, 7, 14, 40, 45, 99, 113} {
		got, ok := numberFrom(Normalize(spokenDigits(n)))
		if !ok || got != n {
			t.Errorf("round trip %d -> %q -> %d, %v", n, spokenDigits(n), got, ok)
		}
	}
}

func TestMatch(t *testing.T) {
	repos := []string{"roxiq", "dossier", "ship"}
	tests := []struct {
		spoken    string
		want      string
		ambiguous bool
		ok        bool
	}{
		{"roxiq", "roxiq", false, true},
		{"ROXIQ", "roxiq", false, true},
		{"roxick", "roxiq", false, true},
		{"dosier", "dossier", false, true},
		{"ship", "ship", false, true},
		{"shop", "ship", false, true},
		{"kubernetes", "", false, false},
		{"", "", false, false},
	}
	for _, tt := range tests {
		got, ambiguous, ok := Match(tt.spoken, repos)
		if got != tt.want || ambiguous != tt.ambiguous || ok != tt.ok {
			t.Errorf("Match(%q) = %q, %v, %v; want %q, %v, %v",
				tt.spoken, got, ambiguous, ok, tt.want, tt.ambiguous, tt.ok)
		}
	}
}

func TestMatchTieIsAmbiguousNotGuessed(t *testing.T) {
	_, ambiguous, ok := Match("abcz", []string{"abcx", "abcy"})
	if ok || !ambiguous {
		t.Fatalf("equidistant candidates must report ambiguous, got ok=%v ambiguous=%v", ok, ambiguous)
	}
}
