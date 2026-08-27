package policy

import "testing"

func testWorld() *World {
	return &World{
		Repos: []*Repo{
			{Name: "roxiq", OpenPRs: []int{14, 40}, Merged: []int{12, 13}},
			{Name: "dossier", OpenPRs: []int{7}, Merged: []int{5, 6}},
			{Name: "ship", OpenPRs: []int{}, Merged: []int{21, 22}},
		},
		Services: []*Service{
			{Name: "checkout", Versions: map[string]string{"staging": "v13", "production": "v12"}, PrevProd: "v11"},
			{Name: "gateway", Versions: map[string]string{"staging": "v8", "production": "v8"}},
		},
		Flags: []*Flag{
			{Name: "dark-mode", On: false},
			{Name: "new-billing", On: false},
		},
	}
}

func TestParse(t *testing.T) {
	tests := []struct {
		name    string
		heard   string
		outcome Outcome
		order   Order
	}{
		{"merge word number", "merge roxiq pr fourteen", Parsed, Order{Verb: VerbMerge, Repo: "roxiq", PR: 14}},
		{"merge spelled digits", "merge roxiq pr one four", Parsed, Order{Verb: VerbMerge, Repo: "roxiq", PR: 14}},
		{"merge digits", "merge roxiq PR 40", Parsed, Order{Verb: VerbMerge, Repo: "roxiq", PR: 40}},
		{"merge p r spelled", "merge roxiq P R one four", Parsed, Order{Verb: VerbMerge, Repo: "roxiq", PR: 14}},
		{"merge pull request", "merge roxiq pull request forty", Parsed, Order{Verb: VerbMerge, Repo: "roxiq", PR: 40}},
		{"merge no keyword", "merge roxiq fourteen", Parsed, Order{Verb: VerbMerge, Repo: "roxiq", PR: 14}},
		{"merge fuzzy repo", "merge roxick pr fourteen", Parsed, Order{Verb: VerbMerge, Repo: "roxiq", PR: 14}},
		{"merge fuzzy repo dossier", "merge dosier", Parsed, Order{Verb: VerbMerge, Repo: "dossier", PR: 7}},
		{"merge sole pr inferred", "merge dossier", Parsed, Order{Verb: VerbMerge, Repo: "dossier", PR: 7}},
		{"merge ambiguous two prs", "merge roxiq", Clarified, Order{}},
		{"merge no open prs", "merge ship", Rejected, Order{}},
		{"merge pr not open", "merge roxiq pr seven", Rejected, Order{}},
		{"merge unknown repo", "merge kubernetes pr two", Rejected, Order{}},
		{"merge missing repo", "merge", Rejected, Order{}},
		{"deploy to production", "deploy checkout to production", Parsed,
			Order{Verb: VerbDeploy, Service: "checkout", Env: "production", Version: "v13"}},
		{"deploy prod alias", "deploy checkout to prod", Parsed,
			Order{Verb: VerbDeploy, Service: "checkout", Env: "production", Version: "v13"}},
		{"deploy to staging bumps", "deploy gateway to staging", Parsed,
			Order{Verb: VerbDeploy, Service: "gateway", Env: "staging", Version: "v9"}},
		{"deploy missing env", "deploy checkout", Clarified, Order{}},
		{"deploy bad env", "deploy checkout to mars", Rejected, Order{}},
		{"deploy unknown service", "deploy warehouse to production", Rejected, Order{}},
		{"rollback", "rollback checkout", Parsed,
			Order{Verb: VerbRollback, Service: "checkout", Env: "production", Version: "v11"}},
		{"roll back two words", "roll back checkout", Parsed,
			Order{Verb: VerbRollback, Service: "checkout", Env: "production", Version: "v11"}},
		{"rollback nothing recorded", "rollback gateway", Rejected, Order{}},
		{"flag on", "flag dark mode on", Parsed, Order{Verb: VerbFlag, Flag: "dark-mode", On: true}},
		{"flag hyphen name", "flag dark-mode on", Parsed, Order{Verb: VerbFlag, Flag: "dark-mode", On: true}},
		{"flag already off", "flag dark mode off", Rejected, Order{}},
		{"flag missing state", "flag dark mode", Clarified, Order{}},
		{"flag unknown", "flag turbo on", Rejected, Order{}},
		{"unknown verb", "launch the missiles", Rejected, Order{}},
		{"empty", "", Rejected, Order{}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			w := testWorld()
			got := Parse(w, Normalize(tt.heard))
			if got.Outcome != tt.outcome {
				t.Fatalf("outcome = %v (say %q), want %v", got.Outcome, got.Say, tt.outcome)
			}
			if tt.outcome == Parsed && got.Order != tt.order {
				t.Fatalf("order = %+v, want %+v", got.Order, tt.order)
			}
			if tt.outcome != Parsed && got.Say == "" {
				t.Fatalf("clarify/reject must carry a spoken line")
			}
		})
	}
}

func TestParseNeverMutatesWorld(t *testing.T) {
	w := testWorld()
	Parse(w, Normalize("merge roxiq pr fourteen"))
	Parse(w, Normalize("deploy checkout to production"))
	Parse(w, Normalize("flag dark mode on"))
	if len(w.repo("roxiq").OpenPRs) != 2 {
		t.Fatal("parse mutated the world: PR list changed")
	}
	if w.service("checkout").Versions["production"] != "v12" {
		t.Fatal("parse mutated the world: deploy happened")
	}
	if w.flag("dark-mode").On {
		t.Fatal("parse mutated the world: flag flipped")
	}
}

func TestNormalize(t *testing.T) {
	tests := []struct {
		in   string
		want []string
	}{
		{"Merge Roxiq PR-14!", []string{"merge", "roxiq", "pr", "14"}},
		{"merge the roxiq pr number fourteen", []string{"merge", "roxiq", "pr", "fourteen"}},
		{"P R one four", []string{"pr", "one", "four"}},
		{"", nil},
	}
	for _, tt := range tests {
		got := Normalize(tt.in)
		if len(got) != len(tt.want) {
			t.Fatalf("Normalize(%q) = %v, want %v", tt.in, got, tt.want)
		}
		for i := range got {
			if got[i] != tt.want[i] {
				t.Fatalf("Normalize(%q) = %v, want %v", tt.in, got, tt.want)
			}
		}
	}
}
