// Package policy is the correctness kernel: parsing, gating, and world
// mutation are deterministic code. No model runs anywhere in this package.
package policy

import (
	"fmt"
	"strconv"
	"strings"
)

// World is the simulated ops estate the console operates on.
type World struct {
	Repos    []*Repo    `json:"repos"`
	Services []*Service `json:"services"`
	Flags    []*Flag    `json:"flags"`
}

type Repo struct {
	Name    string `json:"name"`
	OpenPRs []int  `json:"open_prs"`
	Merged  []int  `json:"merged"`
}

type Service struct {
	Name     string            `json:"name"`
	Versions map[string]string `json:"versions"` // env -> version, e.g. "production": "v12"
	PrevProd string            `json:"prev_prod,omitempty"`
}

type Flag struct {
	Name string `json:"name"`
	On   bool   `json:"on"`
}

func (w *World) repo(name string) *Repo {
	for _, r := range w.Repos {
		if r.Name == name {
			return r
		}
	}
	return nil
}

func (w *World) service(name string) *Service {
	for _, s := range w.Services {
		if s.Name == name {
			return s
		}
	}
	return nil
}

func (w *World) flag(name string) *Flag {
	for _, f := range w.Flags {
		if f.Name == name {
			return f
		}
	}
	return nil
}

func (r *Repo) prOpen(n int) bool {
	for _, pr := range r.OpenPRs {
		if pr == n {
			return true
		}
	}
	return false
}

// Apply executes a parsed-and-confirmed order against the world and returns
// a past-tense effect line for the audit log. It re-validates so a confirm
// can never mutate state the parser did not see.
func (w *World) Apply(o Order) (string, error) {
	switch o.Verb {
	case VerbMerge:
		return w.applyMerge(o)
	case VerbDeploy:
		return w.applyDeploy(o)
	case VerbRollback:
		return w.applyRollback(o)
	case VerbFlag:
		return w.applyFlag(o)
	}
	return "", fmt.Errorf("unknown verb %q", o.Verb)
}

func (w *World) applyMerge(o Order) (string, error) {
	r := w.repo(o.Repo)
	if r == nil {
		return "", fmt.Errorf("no repo %q", o.Repo)
	}
	if !r.prOpen(o.PR) {
		return "", fmt.Errorf("%s has no open PR %d", o.Repo, o.PR)
	}
	kept := r.OpenPRs[:0]
	for _, pr := range r.OpenPRs {
		if pr != o.PR {
			kept = append(kept, pr)
		}
	}
	r.OpenPRs = kept
	r.Merged = append(r.Merged, o.PR)
	return fmt.Sprintf("merged %s#%d. %d open PR(s) remain on %s.", o.Repo, o.PR, len(r.OpenPRs), o.Repo), nil
}

func (w *World) applyDeploy(o Order) (string, error) {
	s := w.service(o.Service)
	if s == nil {
		return "", fmt.Errorf("no service %q", o.Service)
	}
	if o.Env == "production" {
		s.PrevProd = s.Versions["production"]
	}
	s.Versions[o.Env] = o.Version
	return fmt.Sprintf("deployed %s %s to %s.", o.Service, o.Version, o.Env), nil
}

func (w *World) applyRollback(o Order) (string, error) {
	s := w.service(o.Service)
	if s == nil {
		return "", fmt.Errorf("no service %q", o.Service)
	}
	if s.PrevProd == "" {
		return "", fmt.Errorf("nothing to roll %s back to", o.Service)
	}
	s.Versions["production"] = o.Version
	s.PrevProd = ""
	return fmt.Sprintf("rolled %s production back to %s.", o.Service, o.Version), nil
}

func (w *World) applyFlag(o Order) (string, error) {
	f := w.flag(o.Flag)
	if f == nil {
		return "", fmt.Errorf("no flag %q", o.Flag)
	}
	f.On = o.On
	return fmt.Sprintf("flag %s is now %s.", o.Flag, onOff(o.On)), nil
}

func onOff(on bool) string {
	if on {
		return "ON"
	}
	return "off"
}

// versionNumber extracts the integer from a "v13"-style version string.
func versionNumber(v string) (int, bool) {
	n, err := strconv.Atoi(strings.TrimPrefix(v, "v"))
	if err != nil {
		return 0, false
	}
	return n, true
}

func bumpVersion(v string) string {
	n, ok := versionNumber(v)
	if !ok {
		return v + "+1"
	}
	return fmt.Sprintf("v%d", n+1)
}
