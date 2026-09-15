package main

import (
	"bufio"
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"time"
)

var controllerBinary string

type observation struct {
	Name        string     `json:"name"`
	Expected    string     `json:"expected"`
	Actual      string     `json:"actual"`
	Passed      bool       `json:"passed"`
	Effects     []receipt  `json:"effects"`
	Completions []terminal `json:"completions"`
	Errors      []string   `json:"errors,omitempty"`
}
type scenario struct {
	dir, name, mode, mutant string
	commands                []string
	errors                  []string
	step                    int
}

func (s *scenario) fail(format string, args ...any) {
	s.errors = append(s.errors, fmt.Sprintf(format, args...))
}
func (s *scenario) require(ok bool, format string, args ...any) {
	if !ok {
		s.fail(format, args...)
	}
}
func (s *scenario) args(inc, bound string) []string {
	a := []string{"worker", "--dir", s.dir, "--mode", s.mode, "--run", "run-1", "--inc", inc}
	if bound != "" {
		a = append(a, "--boundary", bound)
	}
	if s.mutant != "" {
		a = append(a, "--mutant", s.mutant)
	}
	return a
}
func (s *scenario) command(args []string, bound string) (string, error) {
	s.step++
	s.commands = append(s.commands, quoteCommand(append([]string{controllerBinary}, args...)))
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	cmd := exec.CommandContext(ctx, controllerBinary, args...)
	var errout bytes.Buffer
	cmd.Stderr = &errout
	if bound == "" {
		out, e := cmd.Output()
		log := string(out) + errout.String()
		durable(filepath.Join(s.dir, fmt.Sprintf("process-%02d.log", s.step)), []byte(log))
		return log, e
	}
	stdout, e := cmd.StdoutPipe()
	if e != nil {
		return "", e
	}
	stdin, e := cmd.StdinPipe()
	if e != nil {
		return "", e
	}
	if e = cmd.Start(); e != nil {
		return "", e
	}
	defer stdin.Close()
	// A blocking read is bounded by CommandContext, never a timed crash guess.
	sc := bufio.NewScanner(stdout)
	var log strings.Builder
	seen := false
	for sc.Scan() {
		line := sc.Text()
		log.WriteString(line + "\n")
		if line == "BOUNDARY "+bound {
			seen = true
			break
		}
	}
	if !seen {
		cmd.Process.Kill()
		cmd.Wait()
		return log.String(), fmt.Errorf("missing handshake %s: %s", bound, errout.String())
	}
	e = cmd.Process.Kill()
	if e != nil {
		return log.String(), e
	}
	waitErr := cmd.Wait()
	stdin.Close()
	s.require(waitErr != nil, "worker unexpectedly exited without a kill")
	log.WriteString("CONTROLLER: SIGKILL after " + bound + "; replacement will be a new process\n")
	log.WriteString(errout.String())
	durable(filepath.Join(s.dir, fmt.Sprintf("process-%02d.log", s.step)), []byte(log.String()))
	// Await the sink's EOF cleanup, not the semantic crash boundary (already acknowledged).
	deadline := time.NewTimer(2 * time.Second)
	defer deadline.Stop()
	tick := time.NewTicker(time.Millisecond)
	defer tick.Stop()
	for {
		if _, e := os.Stat(filepath.Join(s.dir, "sink-private", "lock")); os.IsNotExist(e) {
			break
		}
		select {
		case <-deadline.C:
			return log.String(), fmt.Errorf("sink did not exit after worker death")
		case <-tick.C:
		}
	}
	return log.String(), nil
}
func quoteCommand(a []string) string {
	var q []string
	for _, v := range a {
		q = append(q, "'"+strings.ReplaceAll(v, "'", "'\\''")+"'")
	}
	return strings.Join(q, " ")
}
func (s *scenario) worker(inc, bound string) {
	out, e := s.command(s.args(inc, bound), bound)
	if e != nil {
		s.fail("worker %s: %v: %s", inc, e, out)
	}
}
func (s *scenario) refuse(inc, reason string) {
	out, e := s.command(s.args(inc, ""), "")
	s.require(e != nil && strings.Contains(out, reason), "expected refusal %s, got error=%v output=%s", reason, e, out)
}
func (s *scenario) records() []record {
	j, e := loadJournal(filepath.Join(s.dir, "journal.jsonl"))
	if e != nil {
		s.fail("load journal: %v", e)
		return nil
	}
	return j.Records
}
func (s *scenario) effects() []receipt {
	b, e := os.ReadFile(filepath.Join(s.dir, "sink-private", "effects.jsonl"))
	if os.IsNotExist(e) {
		return nil
	}
	if e != nil {
		s.fail("read effects: %v", e)
		return nil
	}
	var out []receipt
	for _, line := range bytes.Split(bytes.TrimSpace(b), []byte{'\n'}) {
		if len(line) == 0 {
			continue
		}
		var r receipt
		if e = json.Unmarshal(line, &r); e != nil {
			s.fail("invalid private effect: %v", e)
			continue
		}
		out = append(out, r)
	}
	return out
}
func (s *scenario) completions() []terminal {
	var out []terminal
	for _, r := range s.records() {
		if r.Kind == "completed" {
			out = append(out, decode[terminal](r.Data))
		}
	}
	return out
}
func (s *scenario) count(kind string) int {
	n := 0
	for _, r := range s.records() {
		if r.Kind == kind {
			n++
		}
	}
	return n
}

// Independent oracle: rebuild expected candidate bytes here (do not call worker checker),
// inspect retained bytes/evidence and private effects, then verify all bound identities.
func (s *scenario) verifyCompleted(source string) {
	sd := digest([]byte(source))
	expected := []byte(fmt.Sprintf(`{"Source":%q,"Body":%q}`, sd, "artifact for "+source))
	cd := digest(expected)
	dir := filepath.Join(s.dir, "subjects", sd)
	artifact, e := os.ReadFile(filepath.Join(dir, "candidate.json"))
	s.require(e == nil && bytes.Equal(artifact, expected), "candidate bytes do not match oracle for %s", source)
	eb, e := os.ReadFile(filepath.Join(dir, "evidence.json"))
	s.require(e == nil, "evidence missing")
	var ev evidence
	if json.Unmarshal(eb, &ev) != nil {
		s.fail("evidence invalid")
	}
	bad, e := os.ReadFile(filepath.Join(dir, "defect.json"))
	s.require(e == nil && !bytes.Equal(bad, expected) && digest(bad) == ev.Defect, "negative artifact absent or valid")
	s.require(ev.Source == sd && ev.Candidate == cd && ev.NegativeError != "", "evidence subject/negative control invalid")
	var ts []terminal
	for _, t := range s.completions() {
		if t.Subject.Source == sd {
			ts = append(ts, t)
		}
	}
	s.require(len(ts) == 1, "expected one completion for %s, got %d", source, len(ts))
	if len(ts) != 1 {
		return
	}
	t := ts[0]
	expectedKey := digest(encode([]string{"run-1", "deliver", sd, cd, digest(eb)}))
	s.require(t.Subject == (subject{"run-1", sd, cd, digest(eb), expectedKey}), "completion subject wrong")
	n := 0
	for _, r := range s.effects() {
		if r.Key == expectedKey {
			n++
			s.require(r.Payload == cd && r == t.Receipt && r.Mode == s.mode, "receipt and private effect differ")
		}
	}
	s.require(n == 1, "expected one effect for operation %s; got %d", expectedKey, n)
}
func (s *scenario) assertUnresolved(effects int) {
	s.require(len(s.effects()) == effects, "expected %d effects, got %d", effects, len(s.effects()))
	s.require(len(s.completions()) == 0, "fabricated completion")
	s.require(s.count("unresolved") >= 1, "missing unresolved identity")
	for _, r := range s.records() {
		if r.Kind == "unresolved" {
			v := decode[subject](r.Data)
			s.require(v.Key != "" && v.Source == digest([]byte("S1")), "unresolved operation identity missing")
		}
	}
}
func (s *scenario) changeCandidate() {
	p := filepath.Join(s.dir, "subjects", digest([]byte("S1")), "candidate.json")
	durable(filepath.Join(s.dir, "changed-candidate.json"), []byte(`{"Body":"changed after check"}`))
	if e := durable(p, []byte(`{"Body":"changed after check"}`)); e != nil {
		s.fail("mutate: %v", e)
	}
}
func (s *scenario) lateTerminal() {
	s.worker("old", "receipt_durable")
	var packet terminal
	for _, r := range s.records() {
		if r.Kind == "receipt" {
			packet = decode[terminal](r.Data)
		}
	}
	s.worker("replacement", "activated")
	p := filepath.Join(s.dir, "late-terminal.json")
	durable(p, encode(packet))
	a := []string{"terminal", "--dir", s.dir, "--packet", p}
	if s.mutant != "" {
		a = append(a, "--mutant", s.mutant)
	}
	out, e := s.command(a, "")
	s.require(e == nil, "terminal command failed: %v %s", e, out)
	s.require(s.count("completed") == 0, "old incarnation advanced the run")
	found := false
	for _, r := range s.records() {
		if r.Kind == "rejected_terminal" {
			v := decode[map[string]string](r.Data)
			found = v["reason"] == "STALE_INCARNATION" && v["received"] == "old" && v["active"] == "replacement"
		}
	}
	s.require(found, "missing retained stale-incarnation refusal")
	if s.mutant == "" {
		s.worker("newest", "")
		s.verifyCompleted("S1")
		s.require(len(s.effects()) == 1, "late result duplicated effect")
	}
}
func runSuite(root string) error {
	if controllerBinary == "" {
		var e error
		controllerBinary, e = os.Executable()
		if e != nil {
			return e
		}
	}
	if e := os.MkdirAll(root, 0700); e != nil {
		return e
	}
	out, e := os.MkdirTemp(root, "run-")
	if e != nil {
		return e
	}
	out, e = filepath.Abs(out)
	if e != nil {
		return e
	}
	type testCase struct {
		name, mode, mutant, expected string
		body                         func(*scenario)
	}
	cases := []testCase{
		{"good", "queryable", "", "one checked S1 completion and one effect", func(s *scenario) {
			s.worker("one", "")
			s.verifyCompleted("S1")
			s.require(len(s.effects()) == 1, "wrong total effects")
		}},
		{"intent-crash", "queryable", "", "recovery queries absent operation then delivers once", func(s *scenario) {
			s.worker("one", "intent_durable")
			s.require(len(s.effects()) == 0, "effect before intent boundary")
			s.worker("two", "")
			s.verifyCompleted("S1")
		}},
		{"opaque-intent-crash", "opaque", "", "conservative unresolved, zero effects", func(s *scenario) { s.worker("one", "intent_durable"); s.worker("two", ""); s.assertUnresolved(0) }},
		{"lost-ack", "queryable", "", "retained receipt recovered, exactly one effect", func(s *scenario) {
			s.worker("one", "sink_committed")
			s.require(len(s.effects()) == 1 && s.count("receipt") == 0, "wrong lost-ack boundary")
			s.worker("two", "")
			s.verifyCompleted("S1")
		}},
		{"receipt-crash", "queryable", "", "durable receipt reused, exactly one effect", func(s *scenario) {
			s.worker("one", "receipt_durable")
			s.require(s.count("receipt") == 1 && s.count("completed") == 0, "wrong receipt boundary")
			s.worker("two", "")
			s.verifyCompleted("S1")
		}},
		{"opaque-lost-ack", "opaque", "", "named unresolved, one effect and zero completions", func(s *scenario) {
			s.worker("one", "sink_committed")
			s.require(s.count("receipt") == 0, "receipt crossed lost-ack boundary")
			s.worker("two", "")
			s.worker("three", "")
			s.assertUnresolved(1)
		}},
		{"late-terminal", "queryable", "", "old terminal rejected with reason, fresh worker completes", func(s *scenario) { s.lateTerminal() }},
		{"source-change", "queryable", "", "S1 historical effect retained; S2 fresh evidence and completion", func(s *scenario) {
			s.worker("one", "")
			durable(filepath.Join(s.dir, "source.txt"), []byte("S2"))
			s.worker("two", "checked")
			s.require(len(s.effects()) == 1 && len(s.completions()) == 1, "S2 falsely delivered before fresh completion")
			s.require(s.count("checked") == 2, "S2 reused S1 checks")
			s.worker("three", "")
			s.verifyCompleted("S1")
			s.verifyCompleted("S2")
			s.require(len(s.effects()) == 2, "expected distinct source effects")
		}},
		{"candidate-change", "queryable", "", "changed bytes refused; restored bytes subsequently complete", func(s *scenario) {
			s.worker("one", "checked")
			s.changeCandidate()
			s.refuse("two", "REFUSED_CANDIDATE_CHANGED")
			s.require(len(s.effects()) == 0 && len(s.completions()) == 0, "changed candidate advanced")
			durable(filepath.Join(s.dir, "subjects", digest([]byte("S1")), "candidate.json"), candidateBytes([]byte("S1")))
			s.worker("three", "")
			s.verifyCompleted("S1")
		}},
		{"truncated-queryable", "queryable", "", "complete prefix recovered and one effect reconciled", func(s *scenario) {
			s.worker("one", "sink_committed")
			appendTail(s)
			s.worker("two", "")
			s.verifyCompleted("S1")
		}},
		{"truncated-opaque", "opaque", "", "complete prefix recovered and unresolved retained", func(s *scenario) {
			s.worker("one", "sink_committed")
			appendTail(s)
			s.worker("two", "")
			s.assertUnresolved(1)
		}},
		{"corrupt-record", "queryable", "", "complete record corruption explicitly refused", func(s *scenario) {
			s.worker("one", "intent_durable")
			p := filepath.Join(s.dir, "journal.jsonl")
			b, _ := os.ReadFile(p)
			b = bytes.Replace(b, []byte("run-1"), []byte("run-X"), 1)
			durable(p, b)
			s.refuse("two", "REFUSED_CORRUPT_JOURNAL")
			s.require(len(s.effects()) == 0, "effect despite corrupt journal")
		}},
		{"definition-mismatch", "queryable", "", "different workflow definition explicitly refused", func(s *scenario) {
			s.worker("one", "intent_durable")
			p := filepath.Join(s.dir, "journal.jsonl")
			b, _ := os.ReadFile(p)
			b = bytes.Replace(b, []byte(definition), []byte("other-definition"), 1)
			durable(p, b)
			s.refuse("two", "REFUSED_DEFINITION")
			s.require(len(s.effects()) == 0, "effect despite definition mismatch")
		}},
		{"sink-conflict", "queryable", "", "same key/payload deduplicated; conflicting payload refused", func(s *scenario) {
			s.worker("one", "")
			effects := s.effects()
			if len(effects) != 1 {
				s.fail("missing initial effect")
				return
			}
			r := effects[0]
			a := []string{"sink", "--dir", s.dir, "--mode", "queryable", "--key", r.Key, "--payload", r.Payload}
			out, e := s.command(a, "")
			s.require(e == nil && strings.TrimSpace(out) == string(encode(r)), "repeat failed to return original receipt")
			a[len(a)-1] = "different"
			out, e = s.command(a, "")
			s.require(e != nil && strings.Contains(out, "REFUSED_KEY_CONFLICT"), "conflict not refused")
			s.require(len(s.effects()) == 1 && s.effects()[0] == r, "prior effect changed")
			s.verifyCompleted("S1")
		}},
		{"run-binding", "queryable", "", "journal cannot be renamed to another logical run", func(s *scenario) {
			s.worker("one", "intent_durable")
			a := s.args("two", "")
			for i := range a {
				if a[i] == "run-1" {
					a[i] = "run-other"
				}
			}
			out, e := s.command(a, "")
			s.require(e != nil && strings.Contains(out, "REFUSED_RUN_BINDING"), "run renamed")
			s.require(len(s.effects()) == 0, "effect on run mismatch")
		}},
		{"mutant-blind-retry", "opaque", "blind-retry", "oracle catches duplicate effect and fabricated completion", func(s *scenario) { s.worker("one", "sink_committed"); s.worker("two", ""); s.assertUnresolved(1) }},
		{"mutant-old-terminal", "queryable", "old-terminal", "oracle catches old incarnation advancement", func(s *scenario) { s.lateTerminal() }},
		{"mutant-stale-evidence", "queryable", "stale-evidence", "oracle catches changed bytes certified by old evidence", func(s *scenario) {
			s.worker("one", "checked")
			s.changeCandidate()
			s.worker("two", "")
			s.verifyCompleted("S1")
			s.require(len(s.completions()) == 0 && len(s.effects()) == 0, "changed candidate advanced")
		}},
	}
	var results []observation
	all := true
	for _, tc := range cases {
		s := &scenario{dir: filepath.Join(out, tc.name), name: tc.name, mode: tc.mode, mutant: tc.mutant}
		os.Mkdir(s.dir, 0700)
		durable(filepath.Join(s.dir, "source.txt"), []byte("S1"))
		tc.body(s)
		// Corrupt journals are intentionally unreadable; preserve them without deriving state.
		var completions []terminal
		if tc.name != "corrupt-record" && tc.name != "definition-mismatch" {
			completions = s.completions()
		}
		passed := len(s.errors) == 0
		actual := "verified"
		if tc.mutant != "" {
			joined := strings.Join(s.errors, "\n")
			switch tc.mutant {
			case "blind-retry":
				passed = strings.Contains(joined, "expected 1 effects, got 2") && len(s.effects()) == 2 && len(completions) == 1
			case "old-terminal":
				passed = strings.Contains(joined, "old incarnation advanced the run") && len(completions) == 1 && completions[0].Incarnation == "old"
			case "stale-evidence":
				passed = strings.Contains(joined, "candidate bytes do not match oracle") && len(completions) == 1 && len(s.effects()) == 1
			}
			for _, err := range s.errors {
				if strings.HasPrefix(err, "worker ") || strings.Contains(err, "missing handshake") {
					passed = false
				}
			}
			actual = "mutant detected by independent assertions"
		}
		if !passed {
			all = false
			actual = "FAILED"
		}
		r := observation{tc.name, tc.expected, actual, passed, s.effects(), completions, s.errors}
		results = append(results, r)
		durable(filepath.Join(s.dir, "observation.json"), append(encode(r), '\n'))
		durable(filepath.Join(s.dir, "commands.sh"), []byte("# Recorded process invocations; crash boundaries require the controller.\n"+strings.Join(s.commands, "\n")+"\n"))
		fmt.Printf("%-24s %s effects=%d completions=%d\n", tc.name, actual, len(r.Effects), len(r.Completions))
	}
	durable(filepath.Join(out, "results.json"), append(encode(results), '\n'))
	durable(filepath.Join(out, "REPLAY.txt"), []byte("From repo: go run . demo --dir artifacts\nEvery execution creates a fresh run directory.\nRecorded boundary commands block until the controller kills their process; replay through demo.\n"))
	fmt.Println("ARTIFACTS " + out)
	if !all {
		return fmt.Errorf("scenario assertions failed; see %s", out)
	}
	fmt.Printf("PASS: %d required/extra cases and 3 detected mutants\n", len(cases)-3)
	return nil
}
func appendTail(s *scenario) {
	p := filepath.Join(s.dir, "journal.jsonl")
	b, _ := os.ReadFile(p)
	durable(filepath.Join(s.dir, "journal-before-truncation.jsonl"), b)
	durable(p, append(b, []byte(`{"Seq":999,"incomplete":`)...))
}
