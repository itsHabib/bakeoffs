package main

import (
	"bufio"
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"time"

	w "github.com/itsHabib/warrant"
)

type child struct {
	cmd    *exec.Cmd
	lines  chan string
	done   chan error
	input  *os.File
	log    *os.File
	waited bool
}

func launch(exe string, args []string, logPath string, mutants bool) (*child, error) {
	log, e := os.Create(logPath)
	if e != nil {
		return nil, e
	}
	cmd := exec.Command(exe, args...)
	cmd.Stderr = log
	if mutants {
		cmd.Env = append(os.Environ(), "BAKEOFF_MUTANTS=enabled")
	}
	stdout, e := cmd.StdoutPipe()
	if e != nil {
		log.Close()
		return nil, e
	}
	r, in, e := os.Pipe()
	if e != nil {
		log.Close()
		return nil, e
	}
	cmd.Stdin = r
	c := &child{cmd: cmd, lines: make(chan string, 128), done: make(chan error, 1), input: in, log: log}
	if e = cmd.Start(); e != nil {
		r.Close()
		in.Close()
		log.Close()
		return nil, e
	}
	r.Close()
	go func() {
		s := bufio.NewScanner(stdout)
		s.Buffer(make([]byte, 4096), 1024*1024)
		for s.Scan() {
			line := s.Text()
			fmt.Fprintln(log, line)
			c.lines <- line
		}
		close(c.lines)
		c.done <- cmd.Wait()
	}()
	return c, nil
}
func (c *child) until(prefix string) error {
	timer := time.NewTimer(15 * time.Second)
	defer timer.Stop()
	for {
		select {
		case line, ok := <-c.lines:
			if !ok {
				return fmt.Errorf("process ended before handshake %q", prefix)
			}
			if strings.HasPrefix(line, prefix) {
				return nil
			}
		case <-timer.C:
			return fmt.Errorf("timeout waiting for %q", prefix)
		}
	}
}
func (c *child) wait() error {
	if c.waited {
		return nil
	}
	c.waited = true
	defer c.input.Close()
	defer c.log.Close()
	select {
	case e := <-c.done:
		return e
	case <-time.After(20 * time.Second):
		_ = c.cmd.Process.Kill()
		<-c.done
		return fmt.Errorf("child exceeded 20 second bound")
	}
}
func (c *child) kill() {
	if c != nil && !c.waited {
		_ = c.cmd.Process.Kill()
		_ = c.wait()
	}
}

type scenario struct {
	exe, root, dir, private, socket, mode string
	sink                                  *child
	workers                               int
	mutant                                string
}

func newScenario(exe, root, name, mode, mutant string) (*scenario, error) {
	s := &scenario{exe: exe, root: filepath.Join(root, name), mode: mode, mutant: mutant}
	s.dir = filepath.Join(s.root, "run")
	s.private = filepath.Join(s.root, "controller-private")
	if e := os.MkdirAll(s.dir, 0700); e != nil {
		return nil, e
	}
	sockDir, e := os.MkdirTemp("/tmp", "wsk-")
	if e != nil {
		return nil, e
	}
	s.socket = filepath.Join(sockDir, "s")
	args := []string{"sink", "-dir", s.private, "-socket", s.socket, "-mode", mode}
	s.sink, e = launch(exe, args, filepath.Join(s.root, "sink.log"), false)
	if e != nil {
		return nil, e
	}
	s.recordCommand(args)
	if e = s.sink.until("ready"); e != nil {
		s.close()
		return nil, e
	}
	_ = saveJSON(filepath.Join(s.root, "inputs.json"), map[string]any{"logical_run_identity": "journal header run field", "S1": string(mustSource("S1")), "S2": string(mustSource("S2")), "mode": mode, "source_commit": upstream, "mutant": mutant})
	_ = save(filepath.Join(s.root, "REPLAY.sh"), []byte("#!/bin/sh\nset -eu\n"+shellQuote(exe)+" replay -dir "+shellQuote(s.dir)+" -mode "+mode+"\n# Recreate controlled execution: ./run.sh demo\n"))
	return s, nil
}
func mustSource(rev string) []byte {
	b, e := sourceBytes(rev)
	if e != nil {
		panic(e)
	}
	return b
}
func shellQuote(v string) string { return "'" + strings.ReplaceAll(v, "'", "'\"'\"'") + "'" }
func (s *scenario) recordCommand(args []string) {
	f, e := os.OpenFile(filepath.Join(s.root, "commands.ndjson"), os.O_CREATE|os.O_APPEND|os.O_WRONLY, 0600)
	if e == nil {
		_ = json.NewEncoder(f).Encode(append([]string{s.exe}, args...))
		f.Close()
	}
}
func (s *scenario) close() { s.sink.kill(); _ = os.RemoveAll(filepath.Dir(s.socket)) }
func (s *scenario) start(extra ...string) (*child, error) {
	s.workers++
	args := []string{"worker", "-dir", s.dir, "-socket", s.socket, "-mode", s.mode}
	if s.mutant != "" {
		args = append(args, "-mutant", s.mutant)
	}
	args = append(args, extra...)
	s.recordCommand(args)
	return launch(s.exe, args, filepath.Join(s.root, fmt.Sprintf("worker-%d.log", s.workers)), s.mutant != "")
}
func (s *scenario) run(extra ...string) (Outcome, error) {
	c, e := s.start(extra...)
	if e != nil {
		return Outcome{}, e
	}
	e = c.wait()
	var out Outcome
	readErr := readJSON(filepath.Join(s.dir, "outcome.json"), &out)
	if readErr != nil {
		return out, readErr
	}
	_ = saveJSON(filepath.Join(s.root, fmt.Sprintf("worker-%d.outcome.json", s.workers)), out)
	return out, e
}
func (s *scenario) crash(point string, hold bool) error {
	args := []string{"-pause", point}
	if hold {
		args = append(args, "-hold")
	}
	c, e := s.start(args...)
	if e != nil {
		return e
	}
	defer c.kill()
	if hold {
		e = s.sink.until("reply_withheld ")
	} else {
		e = c.until(point)
	}
	if e != nil {
		return e
	}
	if e = saveJSON(filepath.Join(s.root, "crash.json"), map[string]any{"worker_pid": c.cmd.Process.Pid, "signal": "SIGKILL", "handshake": point, "ack_withheld": hold}); e != nil {
		return e
	}
	c.kill()
	data, e := os.ReadFile(filepath.Join(s.dir, "journal.ndjson"))
	if e != nil {
		return e
	}
	return save(filepath.Join(s.root, "journal-at-crash.ndjson"), data)
}

// This observer does not trust worker success flags, checker functions, reducer
// certification, or sink response counts. It recomputes bytes and inspects private
// sink records as the controller, then joins identities across the artifacts.
func auditCompleted(s *scenario, revision string, wantEffects int) error {
	var out Outcome
	if e := readJSON(filepath.Join(s.dir, "outcome.json"), &out); e != nil {
		return e
	}
	effects, e := sinkRecords(s.private)
	if e != nil {
		return e
	}
	if len(effects) != wantEffects {
		return fmt.Errorf("effect count: got %d want %d", len(effects), wantEffects)
	}
	source, e := os.ReadFile(filepath.Join(s.dir, "source-"+revision+".json"))
	if e != nil {
		return e
	}
	candidate, e := os.ReadFile(filepath.Join(s.dir, "candidate-"+revision+".json"))
	if e != nil {
		return e
	}
	var src Source
	var cand Candidate
	if json.Unmarshal(source, &src) != nil || json.Unmarshal(candidate, &cand) != nil {
		return fmt.Errorf("invalid actual artifacts")
	}
	expected := 7
	if revision == "S2" {
		expected = 11
	}
	if src.Revision != revision || src.Value != expected || cand.Source != digest(source) || cand.Result != expected*2 {
		return fmt.Errorf("actual candidate bytes are invalid for %s", revision)
	}
	subject := digest(encoded([]string{digest(source), digest(candidate)}))
	var positive, negative bool
	control, e := os.ReadFile(filepath.Join(s.dir, "control-"+revision+".json"))
	if e != nil {
		return e
	}
	var bad Candidate
	if json.Unmarshal(control, &bad) != nil || bad.Source != digest(source) || bad.Result == expected*2 {
		return fmt.Errorf("no real planted defect")
	}
	controlSubject := digest(encoded([]string{digest(source), digest(control)}))
	for _, ev := range out.Evidence {
		if ev.Claim == "quality" && ev.Subject == subject && ev.Verdict == w.Supported && ev.Recipe.InputDigest == digest(candidate) {
			positive = true
		}
		if ev.Claim == "quality" && ev.Subject == controlSubject && ev.Verdict == w.Refuted && ev.Recipe.InputDigest == digest(control) {
			negative = true
		}
	}
	if !positive || !negative {
		return fmt.Errorf("completion lacks actual positive/negative byte-bound evidence")
	}
	if out.Status != "completed" || out.Source != digest(source) || out.Subject != subject || out.Candidate != digest(candidate) || out.Receipt == nil {
		return fmt.Errorf("completion identity mismatch")
	}
	matches := 0
	for _, r := range effects {
		if r.Subject != subject {
			continue
		}
		matches++
		expectedKey := digest(encoded([]string{out.Run, "deliver", subject, digest(candidate)}))
		rID := r.ID
		unsigned := r
		unsigned.ID = ""
		if r.Key != expectedKey || r.Input != digest(candidate) || string(r.Payload) != string(candidate) || r.Run != out.Run || r.Source != digest(source) || r.Step != "deliver" || rID != digest(encoded(unsigned)) || string(encoded(r)) != string(encoded(out.Receipt)) {
			return fmt.Errorf("actual effect identity or receipt mismatch")
		}
	}
	if matches != 1 {
		return fmt.Errorf("expected exactly one effect for current subject, got %d", matches)
	}
	j, e := loadJournal(s.dir, s.mode)
	if e != nil {
		return e
	}
	if j.Receipt == nil || j.Receipt.ID != out.Receipt.ID {
		return fmt.Errorf("receipt not retained in journal")
	}
	if string(encoded(j.ledger())) != string(encoded(out.Evidence)) {
		return fmt.Errorf("completion evidence differs from durable journal")
	}
	return nil
}
func auditUnresolved(s *scenario, wantEffects int) error {
	effects, e := sinkRecords(s.private)
	if e != nil {
		return e
	}
	if len(effects) != wantEffects {
		return fmt.Errorf("effect count: got %d want %d", len(effects), wantEffects)
	}
	var out Outcome
	if e = readJSON(filepath.Join(s.dir, "outcome.json"), &out); e != nil {
		return e
	}
	j, e := loadJournal(s.dir, s.mode)
	if e != nil {
		return e
	}
	pending := j.Events[len(j.Events)-1].Event
	if pending.Type != w.EffectPrepared || pending.Step != "deliver" || out.Operation != pending.EffectID || out.Status != "unresolved" || out.Reason == "" || out.Receipt != nil {
		return fmt.Errorf("lost operation was not honestly retained unresolved")
	}
	for _, r := range effects {
		if r.Key != pending.EffectID || r.Input != pending.InputDigest || digest(r.Payload) != r.Input || r.Source != j.Source {
			return fmt.Errorf("unresolved operation differs from actual effect")
		}
	}
	return nil
}

type CaseResult struct {
	Case     string `json:"case"`
	Expected string `json:"expected"`
	Actual   string `json:"actual"`
	Passed   bool   `json:"passed"`
	Error    string `json:"error,omitempty"`
	Artifact string `json:"artifact"`
}
type caseSpec struct{ name, mode, mutant, expected string }

var cases = []caseSpec{
	{"good-queryable", "queryable", "", "completed, one effect"}, {"good-opaque", "opaque", "", "completed, one effect"},
	{"intent-queryable", "queryable", "", "replacement completes, one effect"}, {"intent-opaque", "opaque", "", "unresolved, zero effects"},
	{"lost-ack-queryable", "queryable", "", "retained receipt reconciled, one effect"}, {"receipt-queryable", "queryable", "", "receipt reused, one effect"}, {"receipt-opaque", "opaque", "", "receipt reused, one effect"},
	{"lost-ack-opaque", "opaque", "", "unresolved named operation, one effect"}, {"late-terminal", "queryable", "", "old terminal rejected before advancement; continuation completes"},
	{"source-change", "queryable", "", "old evidence refused for S2; two historical effects; S2 completes"}, {"candidate-change", "queryable", "", "changed bytes refused, zero effects"},
	{"torn-tail-queryable", "queryable", "", "prefix recovered, one effect"}, {"torn-tail-opaque", "opaque", "", "prefix recovered, unresolved, one effect"},
	{"corrupt-record", "queryable", "", "explicit refusal, zero effects"}, {"definition-mismatch", "queryable", "", "explicit refusal, zero effects"},
	{"conflicting-key", "queryable", "", "conflict refused; identical retry retains one receipt"},
	{"replay-is-read-only", "queryable", "", "replay changes neither journal nor effects; continuation completes"},
	{"mutant-blind-retry", "opaque", "blind-retry", "independent oracle detects duplicate opaque effect"}, {"mutant-old-terminal", "queryable", "old-terminal", "independent oracle detects stale advancement"}, {"mutant-stale-candidate", "queryable", "stale-candidate", "independent oracle detects invalid candidate/evidence"},
}

func demo(exe, root string) error {
	var e error
	if root == "" {
		root, e = os.MkdirTemp("/tmp", "warrant-demo-")
		if e != nil {
			return e
		}
	} else {
		entries, e := os.ReadDir(root)
		if e == nil && len(entries) > 0 {
			return fmt.Errorf("demo directory must be fresh")
		}
		if e = os.MkdirAll(root, 0700); e != nil {
			return e
		}
	}
	fmt.Println("Warrant crash/replacement demo — real worker SIGKILL, local sinks")
	var results []CaseResult
	failed := 0
	for _, spec := range cases {
		result := runCase(exe, root, spec)
		results = append(results, result)
		if !result.Passed {
			failed++
		}
		fmt.Printf("%-27s %5t  %s\n", result.Case, result.Passed, result.Actual)
	}
	if e = saveJSON(filepath.Join(root, "results.json"), results); e != nil {
		return e
	}
	fmt.Printf("Artifacts: %s\n%d/%d cases passed (includes 3 detected mutants)\n", root, len(results)-failed, len(results))
	if failed > 0 {
		return fmt.Errorf("%d cases failed", failed)
	}
	return nil
}
func runCase(exe, root string, spec caseSpec) (result CaseResult) {
	result = CaseResult{Case: spec.name, Expected: spec.expected, Artifact: filepath.Join(root, spec.name)}
	s, e := newScenario(exe, root, spec.name, spec.mode, spec.mutant)
	if e != nil {
		result.Error = e.Error()
		return result
	}
	defer s.close()
	defer func() {
		effects, effectErr := sinkRecords(s.private)
		var outcome Outcome
		outcomeErr := readJSON(filepath.Join(s.dir, "outcome.json"), &outcome)
		observation := map[string]any{"effect_count": len(effects), "effects": effects, "worker_outcome": outcome, "observer": "controller read of sink private records; worker has socket interface only"}
		if effectErr != nil {
			observation["effect_read_error"] = effectErr.Error()
		}
		if outcomeErr != nil {
			observation["outcome_read_error"] = outcomeErr.Error()
		}
		if e := saveJSON(filepath.Join(s.root, "controller-observation.json"), observation); e != nil && result.Error == "" {
			result.Error = e.Error()
		}
		if result.Error == "" {
			result.Passed = true
		}
		_ = saveJSON(filepath.Join(s.root, "result.json"), result)
	}()
	e = exercise(s, spec.name)
	if spec.mutant != "" {
		if e == nil {
			result.Error = "mutant survived independent oracle"
			result.Actual = result.Error
			return result
		}
		if !strings.HasPrefix(e.Error(), "ORACLE: ") {
			result.Error = e.Error()
			result.Actual = "mutant did not reach oracle"
			return result
		}
		result.Actual = "detected: " + strings.TrimPrefix(e.Error(), "ORACLE: ")
		return result
	}
	if e != nil {
		result.Error = e.Error()
		result.Actual = e.Error()
		return result
	}
	result.Actual = spec.expected
	return result
}
func oracle(e error) error {
	if e != nil {
		return fmt.Errorf("ORACLE: %w", e)
	}
	return nil
}

func exercise(s *scenario, name string) error {
	switch name {
	case "replay-is-read-only":
		if e := s.crash("intent", false); e != nil {
			return e
		}
		before, e := os.ReadFile(filepath.Join(s.dir, "journal.ndjson"))
		if e != nil {
			return e
		}
		args := []string{"replay", "-dir", s.dir, "-mode", s.mode}
		s.recordCommand(args)
		c, e := launch(s.exe, args, filepath.Join(s.root, "replay.log"), false)
		if e != nil {
			return e
		}
		if e = c.wait(); e != nil {
			return e
		}
		after, e := os.ReadFile(filepath.Join(s.dir, "journal.ndjson"))
		if e != nil {
			return e
		}
		effects, e := sinkRecords(s.private)
		if e != nil {
			return e
		}
		if string(before) != string(after) || len(effects) != 0 {
			return fmt.Errorf("replay executed or wrote state")
		}
		if _, e = s.run(); e != nil {
			return e
		}
		return auditCompleted(s, "S1", 1)
	case "good-queryable", "good-opaque":
		if _, e := s.run(); e != nil {
			return e
		}
		return auditCompleted(s, "S1", 1)
	case "intent-queryable", "intent-opaque":
		if e := s.crash("intent", false); e != nil {
			return e
		}
		if _, e := s.run(); e != nil {
			return e
		}
		if s.mode == "opaque" {
			return auditUnresolved(s, 0)
		}
		return auditCompleted(s, "S1", 1)
	case "lost-ack-queryable", "lost-ack-opaque", "torn-tail-queryable", "torn-tail-opaque", "mutant-blind-retry":
		if e := s.crash("sink committed; reply withheld", true); e != nil {
			return e
		}
		if strings.HasPrefix(name, "torn-tail") {
			if e := appendRaw(filepath.Join(s.dir, "journal.ndjson"), []byte(`{"seq":`)); e != nil {
				return e
			}
		}
		if _, e := s.run(); e != nil {
			return e
		}
		if strings.HasPrefix(name, "torn-tail") {
			b, e := os.ReadFile(filepath.Join(s.dir, "journal.ndjson.discarded-tail"))
			if e != nil || string(b) != `{"seq":` {
				return fmt.Errorf("torn bytes not preserved")
			}
		}
		if s.mode == "opaque" {
			if name == "mutant-blind-retry" {
				return oracle(auditUnresolved(s, 1))
			}
			return auditUnresolved(s, 1)
		}
		return auditCompleted(s, "S1", 1)
	case "receipt-queryable", "receipt-opaque":
		if e := s.crash("receipt", false); e != nil {
			return e
		}
		before, e := sinkRecords(s.private)
		if e != nil {
			return e
		}
		if _, e = s.run(); e != nil {
			return e
		}
		after, e := sinkRecords(s.private)
		if e != nil {
			return e
		}
		if string(encoded(before)) != string(encoded(after)) {
			return fmt.Errorf("retained effect changed")
		}
		return auditCompleted(s, "S1", 1)
	case "late-terminal", "mutant-old-terminal":
		return lateCase(s)
	case "source-change":
		if _, e := s.run(); e != nil {
			return e
		}
		if e := auditCompleted(s, "S1", 1); e != nil {
			return e
		}
		j, e := loadJournal(s.dir, s.mode)
		if e != nil {
			return e
		}
		oldLedger := j.ledger()
		oldReceipt := *j.Receipt
		s2 := mustSource("S2")
		newSubject := digest(encoded([]string{digest(s2), digest(candidateBytes(s2))}))
		if met, _ := definition(s.mode).RequirementsMet("deliver", newSubject, oldLedger); met {
			return fmt.Errorf("S1 evidence certified S2")
		}
		if e = saveJSON(filepath.Join(s.root, "S2-old-evidence-refusal.json"), map[string]any{"subject": newSubject, "old_evidence": oldLedger, "requirements_met": false}); e != nil {
			return e
		}
		if _, e = s.run("-revision", "S2"); e != nil {
			return e
		}
		if e = auditCompleted(s, "S2", 2); e != nil {
			return e
		}
		effects, e := sinkRecords(s.private)
		if e != nil {
			return e
		}
		if string(encoded(effects[0])) != string(encoded(oldReceipt)) {
			return fmt.Errorf("S1 historical receipt changed")
		}
		return nil
	case "candidate-change", "mutant-stale-candidate":
		if e := s.crash("checked", false); e != nil {
			return e
		}
		path := filepath.Join(s.dir, "candidate-S1.json")
		var cand Candidate
		if e := readJSON(path, &cand); e != nil {
			return e
		}
		cand.Result++
		if e := save(path, encoded(cand)); e != nil {
			return e
		}
		if _, e := s.run(); e != nil {
			return e
		}
		if s.mutant != "" {
			return oracle(auditCompleted(s, "S1", 1))
		}
		effects, e := sinkRecords(s.private)
		if e != nil {
			return e
		}
		var out Outcome
		_ = readJSON(filepath.Join(s.dir, "outcome.json"), &out)
		if len(effects) != 0 || out.Status != "unresolved" || !strings.Contains(out.Reason, "deliver failed") {
			return fmt.Errorf("changed bytes were certified")
		}
		return nil
	case "corrupt-record", "definition-mismatch":
		if e := s.crash("intent", false); e != nil {
			return e
		}
		path := filepath.Join(s.dir, "journal.ndjson")
		before, e := os.ReadFile(path)
		if e != nil {
			return e
		}
		if name == "corrupt-record" {
			e = appendRaw(path, []byte("{broken}\n"))
		} else {
			lines := strings.Split(strings.TrimSuffix(string(before), "\n"), "\n")
			var rs []Record
			for _, line := range lines {
				var r Record
				_ = json.Unmarshal([]byte(line), &r)
				rs = append(rs, r)
			}
			rs[0].Definition = "sha256:wrong-definition"
			var updated []byte
			for i := range rs {
				if i > 0 {
					rs[i].Prev = rs[i-1].Hash
				}
				rs[i].Hash = recordHash(rs[i])
				updated = append(updated, append(encoded(rs[i]), '\n')...)
			}
			e = save(path, updated)
		}
		if e != nil {
			return e
		}
		corrupt, e := os.ReadFile(path)
		if e != nil {
			return e
		}
		out, runErr := s.run()
		if runErr == nil || out.Status != "refused" {
			return fmt.Errorf("invalid journal not refused")
		}
		after, e := os.ReadFile(path)
		if e != nil {
			return e
		}
		effects, e := sinkRecords(s.private)
		if e != nil {
			return e
		}
		if string(after) != string(corrupt) || len(effects) != 0 {
			return fmt.Errorf("refusal changed journal or effects")
		}
		return nil
	case "conflicting-key":
		if _, e := s.run(); e != nil {
			return e
		}
		effects, e := sinkRecords(s.private)
		if e != nil {
			return e
		}
		r := effects[0]
		repeat, e := callSink(s.socket, Request{Action: "put", Mode: s.mode, Receipt: r})
		if e != nil {
			return e
		}
		if string(encoded(repeat)) != string(encoded(r)) {
			return fmt.Errorf("dedup receipt changed")
		}
		conflicting := r
		conflicting.Payload = []byte("different payload")
		conflicting.Input = digest(conflicting.Payload)
		_, e = callSink(s.socket, Request{Action: "put", Mode: s.mode, Receipt: conflicting})
		if e == nil || !strings.Contains(e.Error(), "conflicts") {
			return fmt.Errorf("conflicting key not refused")
		}
		if e = saveJSON(filepath.Join(s.root, "conflict-refusal.json"), map[string]any{"original": r, "conflicting_request": conflicting, "error": e.Error()}); e != nil {
			return e
		}
		return auditCompleted(s, "S1", 1)
	}
	return fmt.Errorf("unknown case")
}
func appendRaw(path string, b []byte) error {
	f, e := os.OpenFile(path, os.O_WRONLY|os.O_APPEND, 0600)
	if e != nil {
		return e
	}
	defer f.Close()
	if _, e = f.Write(b); e != nil {
		return e
	}
	return f.Sync()
}
func lateCase(s *scenario) error {
	if e := s.crash("sink committed; reply withheld", true); e != nil {
		return e
	}
	j, e := loadJournal(s.dir, s.mode)
	if e != nil {
		return e
	}
	oldInc := j.Incarnation
	effects, e := sinkRecords(s.private)
	if e != nil {
		return e
	}
	path := filepath.Join(s.root, "late-terminal.json")
	if e = saveJSON(path, LateTerminal{oldInc, effects[0]}); e != nil {
		return e
	}
	before := len(j.Events)
	c, e := s.start("-terminal", path, "-pause", "fenced")
	if e != nil {
		return e
	}
	defer c.kill()
	if e = c.until("fenced"); e != nil {
		return e
	}
	c.kill()
	j, e = loadJournal(s.dir, s.mode)
	if e != nil {
		return e
	}
	if len(j.Events) != before {
		return oracle(fmt.Errorf("old incarnation advanced journal from %d to %d events", before, len(j.Events)))
	}
	found := false
	for _, r := range j.Records {
		if r.Kind == "rejected" && strings.Contains(r.Reason, oldInc) {
			found = true
		}
	}
	if !found || j.Incarnation == oldInc {
		return oracle(fmt.Errorf("old terminal lacks durable fence reason or fresh incarnation"))
	}
	if _, e = s.run(); e != nil {
		return e
	}
	return auditCompleted(s, "S1", 1)
}
