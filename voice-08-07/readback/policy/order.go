package policy

import (
	"fmt"
	"strings"
)

type Verb string

const (
	VerbMerge    Verb = "merge"
	VerbDeploy   Verb = "deploy"
	VerbRollback Verb = "rollback"
	VerbFlag     Verb = "flag"
)

// Order is one exact, fully-resolved action. By the time an Order exists,
// every field is a validated member of the world — parsing never emits a
// "probably" order.
type Order struct {
	Verb    Verb   `json:"verb"`
	Repo    string `json:"repo,omitempty"`
	PR      int    `json:"pr,omitempty"`
	Service string `json:"service,omitempty"`
	Env     string `json:"env,omitempty"`
	Version string `json:"version,omitempty"`
	Flag    string `json:"flag,omitempty"`
	On      bool   `json:"on,omitempty"`
}

// Canonical is the exact screen form shown on the order card.
func (o Order) Canonical() string {
	switch o.Verb {
	case VerbMerge:
		return fmt.Sprintf("merge %s#%d → main", o.Repo, o.PR)
	case VerbDeploy:
		return fmt.Sprintf("deploy %s %s → %s", o.Service, o.Version, strings.ToUpper(o.Env))
	case VerbRollback:
		return fmt.Sprintf("rollback %s production → %s", o.Service, o.Version)
	case VerbFlag:
		return fmt.Sprintf("flag %s → %s", o.Flag, onOff(o.On))
	}
	return string(o.Verb)
}

// Spoken is the canonical action read over the audio channel: digits spelled
// individually, dangerous parameters made unmistakable ("to PRODUCTION").
func (o Order) Spoken() string {
	switch o.Verb {
	case VerbMerge:
		return fmt.Sprintf("merge, %s, P R %s, into main", o.Repo, spokenDigits(o.PR))
	case VerbDeploy:
		env := o.Env
		if env == "production" {
			env = "PRODUCTION"
		}
		return fmt.Sprintf("deploy %s, version %s, to %s", o.Service, spokenVersion(o.Version), env)
	case VerbRollback:
		return fmt.Sprintf("roll back %s PRODUCTION, to version %s", o.Service, spokenVersion(o.Version))
	case VerbFlag:
		return fmt.Sprintf("flag %s, %s", strings.ReplaceAll(o.Flag, "-", " "), onOff(o.On))
	}
	return string(o.Verb)
}

// Readback is the full spoken challenge: canonical action plus the exact
// confirm phrase the operator must echo back.
func (o Order) Readback(token int) string {
	return fmt.Sprintf("read back. %s. say, confirm %s. or say, negative.", o.Spoken(), spokenDigits(token))
}

// Token picks the confirm token for this order. When the order carries risky
// digits (a PR number, a version) the token IS those digits, so confirming
// forces the operator to re-speak the exact value that could be misheard.
// Otherwise it is a fresh two-digit challenge from rnd.
func (o Order) Token(rnd func(n int) int) int {
	switch o.Verb {
	case VerbMerge:
		return o.PR
	case VerbDeploy, VerbRollback:
		if n, ok := versionNumber(o.Version); ok {
			return n
		}
	}
	return 10 + rnd(90)
}

// Fields are label/value pairs for the structured card in the UI.
func (o Order) Fields() [][2]string {
	switch o.Verb {
	case VerbMerge:
		return [][2]string{{"verb", "merge"}, {"repo", o.Repo}, {"pr", fmt.Sprintf("#%d", o.PR)}, {"into", "main"}}
	case VerbDeploy:
		return [][2]string{{"verb", "deploy"}, {"service", o.Service}, {"version", o.Version}, {"env", strings.ToUpper(o.Env)}}
	case VerbRollback:
		return [][2]string{{"verb", "rollback"}, {"service", o.Service}, {"env", "PRODUCTION"}, {"to", o.Version}}
	case VerbFlag:
		return [][2]string{{"verb", "flag"}, {"flag", o.Flag}, {"state", onOff(o.On)}}
	}
	return nil
}

func spokenVersion(v string) string {
	if n, ok := versionNumber(v); ok {
		return "v " + spokenDigits(n)
	}
	return v
}
