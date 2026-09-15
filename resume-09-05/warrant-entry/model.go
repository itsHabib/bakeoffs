package main

import (
	"bytes"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"

	w "github.com/itsHabib/warrant"
)

const upstream = "33094f0b99a5a200fdcfebc87c28b39efab5ef8f"

func digest(b []byte) string { h := sha256.Sum256(b); return "sha256:" + hex.EncodeToString(h[:]) }
func encoded(v any) []byte {
	b, e := json.Marshal(v)
	if e != nil {
		panic(e)
	}
	return b
}
func identity(source string, candidate []byte) string {
	return digest(encoded([]string{source, digest(candidate)}))
}
func operation(run, step, subject, input string) string {
	return digest(encoded([]string{run, step, subject, input}))
}
func readJSON(path string, v any) error {
	b, e := os.ReadFile(path)
	if e != nil {
		return e
	}
	return json.Unmarshal(b, v)
}

// Process-crash boundary: flush file contents before rename; no power-loss claim.
func save(path string, b []byte) error {
	if e := os.MkdirAll(filepath.Dir(path), 0700); e != nil {
		return e
	}
	f, e := os.CreateTemp(filepath.Dir(path), ".write-")
	if e != nil {
		return e
	}
	defer os.Remove(f.Name())
	if _, e = f.Write(b); e != nil {
		f.Close()
		return e
	}
	if e = f.Sync(); e != nil {
		f.Close()
		return e
	}
	if e = f.Close(); e != nil {
		return e
	}
	return os.Rename(f.Name(), path)
}
func saveJSON(path string, v any) error { return save(path, append(encoded(v), '\n')) }

type Source struct {
	Revision string `json:"revision"`
	Value    int    `json:"value"`
}
type Candidate struct {
	Source string `json:"source"`
	Result int    `json:"result"`
}

func sourceBytes(revision string) ([]byte, error) {
	switch revision {
	case "S1":
		return encoded(Source{revision, 7}), nil
	case "S2":
		return encoded(Source{revision, 11}), nil
	}
	return nil, fmt.Errorf("unknown revision %q", revision)
}
func candidateBytes(source []byte) []byte {
	var s Source
	_ = json.Unmarshal(source, &s)
	return encoded(Candidate{digest(source), s.Value * 2})
}

// Both the real candidate and planted defect pass through this exact checker.
func check(source, candidate []byte) bool {
	var s Source
	var c Candidate
	dec := json.NewDecoder(bytes.NewReader(candidate))
	dec.DisallowUnknownFields()
	return json.Unmarshal(source, &s) == nil && dec.Decode(&c) == nil && bytes.Equal(encoded(c), candidate) && c.Source == digest(source) && c.Result == s.Value*2
}
func evidence(source, candidate []byte) w.Evidence {
	verdict := w.Refuted
	if check(source, candidate) {
		verdict = w.Supported
	}
	return w.Evidence{Claim: "quality", Subject: identity(digest(source), candidate), Verdict: verdict,
		Recipe: w.Recipe{Tool: "double-check", Version: "1", Command: "resume check -source SOURCE.json -candidate CANDIDATE.json", InputDigest: digest(candidate), Seed: "none", Bounds: []string{"exact JSON source digest and integer doubling"}}}
}

func definition(mode string) w.Pipeline {
	replay := w.Deduplicated
	if mode == "opaque" {
		replay = w.Manual
	}
	demands := []w.Requirement{{Kind: w.Supports, Claim: "quality"}, {Kind: w.NegativeControl, Claim: "quality"}}
	p := w.Pipeline{Name: "resume-candidate", Version: 1, Start: "prepare", MaxAttempts: 1,
		Steps: []w.StepSpec{{Name: "prepare", ReplayClass: w.Idempotent}, {Name: "check", ReplayClass: w.Idempotent}, {Name: "control", ReplayClass: w.Idempotent, Control: true}, {Name: "deliver", ReplayClass: replay, Requires: demands}, {Name: "complete", ReplayClass: w.Idempotent, Requires: append(append([]w.Requirement{}, demands...), w.Requirement{Kind: w.Supports, Claim: "delivered"})}}}
	for i, s := range p.Steps {
		action := w.RuleAction{Kind: w.Finish}
		if i+1 < len(p.Steps) {
			action = w.RuleAction{Kind: w.GoTo, Step: p.Steps[i+1].Name}
		}
		p.Rules = append(p.Rules, w.Rule{Step: s.Name, Outcome: w.Success, Action: action}, w.Rule{Step: s.Name, Outcome: w.Failure, Action: w.RuleAction{Kind: w.Stop, Reason: s.Name + " failed"}})
	}
	p.Digest = digest(encoded(p))
	return p
}

type Receipt struct {
	ID      string `json:"id"`
	Key     string `json:"key"`
	Run     string `json:"run"`
	Step    string `json:"step"`
	Source  string `json:"source"`
	Subject string `json:"subject"`
	Input   string `json:"input"`
	Payload []byte `json:"payload"`
	Number  int    `json:"number"`
}
type Outcome struct {
	Status      string       `json:"status"`
	Reason      string       `json:"reason,omitempty"`
	Operation   string       `json:"operation,omitempty"`
	Run         string       `json:"run"`
	Incarnation string       `json:"incarnation"`
	Revision    string       `json:"revision"`
	Source      string       `json:"source"`
	Subject     string       `json:"subject"`
	Candidate   string       `json:"candidate"`
	Evidence    []w.Evidence `json:"evidence,omitempty"`
	Receipt     *Receipt     `json:"receipt,omitempty"`
	View        w.View       `json:"view"`
	Next        w.NextAction `json:"next"`
}
