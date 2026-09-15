package main

import (
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"os"
	"path/filepath"
	"syscall"

	w "github.com/itsHabib/warrant"
)

type WorkerOptions struct {
	Dir, Socket, Mode, Revision, Pause, Terminal string
	Hold                                         bool
	Mutant                                       string
}
type LateTerminal struct {
	Incarnation string  `json:"incarnation"`
	Receipt     Receipt `json:"receipt"`
}

func handshake(want, point string) {
	if want == point {
		fmt.Println(point)
		var one [1]byte
		_, _ = os.Stdin.Read(one[:])
	}
}
func (o WorkerOptions) candidate() string {
	return filepath.Join(o.Dir, "candidate-"+o.Revision+".json")
}
func worker(o WorkerOptions) (out Outcome, err error) {
	out.Status = "refused"
	if o.Mode != "queryable" && o.Mode != "opaque" {
		return out, fmt.Errorf("unknown mode")
	}
	if o.Mutant != "" && os.Getenv("BAKEOFF_MUTANTS") != "enabled" {
		return out, fmt.Errorf("mutants require test-only enablement")
	}
	if e := os.MkdirAll(o.Dir, 0700); e != nil {
		return out, e
	}
	lock, e := os.OpenFile(filepath.Join(o.Dir, "worker.lock"), os.O_CREATE|os.O_RDWR, 0600)
	if e != nil {
		return out, e
	}
	defer lock.Close()
	if e = syscall.Flock(int(lock.Fd()), syscall.LOCK_EX|syscall.LOCK_NB); e != nil {
		return out, fmt.Errorf("another worker owns journal: %w", e)
	}
	defer syscall.Flock(int(lock.Fd()), syscall.LOCK_UN)
	p := definition(o.Mode)
	if refusals := p.Validate(); len(refusals) > 0 {
		return out, fmt.Errorf("definition invalid: %v", refusals)
	}
	j, e := loadJournal(o.Dir, o.Mode)
	if e != nil {
		return out, e
	}
	if e = j.repairTail(); e != nil {
		return out, e
	}
	token := make([]byte, 16)
	if _, e = rand.Read(token); e != nil {
		return out, e
	}
	inc := hex.EncodeToString(token)
	if len(j.Records) == 0 {
		j.Mode = o.Mode
		if e = j.append(Record{Kind: "header", Run: "run-" + inc[:16], Mode: o.Mode, Definition: p.Digest}); e != nil {
			return out, e
		}
	}
	if e = j.append(Record{Kind: "incarnation", Incarnation: inc}); e != nil {
		return out, e
	}
	source, e := sourceBytes(o.Revision)
	if e != nil {
		return out, e
	}
	if j.Revision != o.Revision {
		if e = j.append(Record{Kind: "source", Revision: o.Revision, Source: digest(source)}); e != nil {
			return out, e
		}
	}
	if e = save(filepath.Join(o.Dir, "source-"+o.Revision+".json"), source); e != nil {
		return out, e
	}
	if len(j.Events) == 0 {
		if e = j.event(w.Event{Type: w.RunStarted, RunID: j.Run, PipelineName: p.Name, PipelineVersion: p.Version, PipelineDigest: p.Digest, Subject: j.Source}, nil); e != nil {
			return out, e
		}
	}
	if o.Terminal != "" {
		if e = receiveTerminal(j, o); e != nil {
			return out, e
		}
		handshake(o.Pause, "fenced")
	}
	for tick := 0; tick < 16; tick++ {
		view, next := w.Reduce(p, j.Events)
		out = Outcome{Status: "unresolved", Run: j.Run, Incarnation: j.Incarnation, Revision: j.Revision, Source: j.Source, Subject: view.Subject, Evidence: j.ledger(), Receipt: j.Receipt, View: view, Next: next, Operation: next.EffectID}
		if next.Action == w.AwaitDecision {
			if o.Mutant == "blind-retry" && j.Events[len(j.Events)-1].Event.Step == "deliver" {
				next = w.NextAction{Action: w.Recover, Step: "deliver", Attempt: 1, EffectID: j.Events[len(j.Events)-1].Event.EffectID}
			} else {
				out.Reason = next.Question
				return out, nil
			}
		}
		if next.Action == w.NothingToDo {
			candidate, e := os.ReadFile(o.candidate())
			if e != nil {
				return out, e
			}
			out.Candidate = digest(candidate)
			if o.Mutant != "stale-candidate" && identity(j.Source, candidate) != view.Subject {
				return out, fmt.Errorf("completed candidate identity changed")
			}
			out.Status = "completed"
			return out, nil
		}
		if next.Action != w.Execute && next.Action != w.Recover {
			return out, fmt.Errorf("unknown action %q", next.Action)
		}
		if next.Step == "deliver" {
			handshake(o.Pause, "checked")
		}
		var spec w.StepSpec
		for _, s := range p.Steps {
			if s.Name == next.Step {
				spec = s
			}
		}
		key := next.EffectID
		if next.Action == w.Execute {
			input := view.Subject
			if next.Step == "deliver" {
				candidate, e := os.ReadFile(o.candidate())
				if e != nil {
					return out, e
				}
				input = digest(candidate)
			}
			key = operation(j.Run, next.Step, view.Subject, input)
			if e = j.event(w.Event{Type: w.EffectPrepared, Step: next.Step, Attempt: next.Attempt, EffectID: key, ReplayClass: spec.ReplayClass, InputDigest: input}, nil); e != nil {
				return out, e
			}
		}
		if next.Step == "deliver" {
			handshake(o.Pause, "intent")
		}
		gateSubject := view.Subject
		if next.Step == "deliver" || next.Step == "complete" {
			candidate, e := os.ReadFile(o.candidate())
			if e != nil {
				return out, e
			}
			if o.Mutant != "stale-candidate" {
				gateSubject = identity(j.Source, candidate)
			}
		}
		if met, unmet := p.RequirementsMet(next.Step, gateSubject, j.ledger()); !met {
			if e = j.event(w.Event{Type: w.StepFailed, Step: next.Step, Attempt: next.Attempt, EffectID: key, Subject: gateSubject, Reason: fmt.Sprint(unmet)}, nil); e != nil {
				return out, e
			}
			continue
		}
		result := w.Event{Type: w.StepSucceeded, Step: next.Step, Attempt: next.Attempt, EffectID: key, PreviousSubject: view.Subject, Subject: view.Subject}
		var receipt *Receipt
		switch next.Step {
		case "prepare":
			candidate := candidateBytes(source)
			if e = save(o.candidate(), candidate); e != nil {
				return out, e
			}
			result.Subject = identity(j.Source, candidate)
		case "check":
			candidate, e := os.ReadFile(o.candidate())
			if e != nil {
				return out, e
			}
			ev := evidence(source, candidate)
			result.Evidence = []w.Evidence{ev}
			if ev.Subject != view.Subject || ev.Verdict != w.Supported {
				result.Type = w.StepFailed
				result.Reason = "actual candidate checker rejected"
			}
		case "control":
			var bad Candidate
			_ = readJSON(o.candidate(), &bad)
			bad.Result++
			b := encoded(bad)
			if e = save(filepath.Join(o.Dir, "control-"+o.Revision+".json"), b); e != nil {
				return out, e
			}
			ev := evidence(source, b)
			result.Evidence = []w.Evidence{ev}
			if ev.Verdict != w.Refuted {
				result.Type = w.StepFailed
				result.Reason = "planted defect escaped checker"
			}
		case "deliver":
			candidate, e := os.ReadFile(o.candidate())
			if e != nil {
				return out, e
			}
			request := Receipt{Key: key, Run: j.Run, Step: "deliver", Source: j.Source, Subject: view.Subject, Input: digest(candidate), Payload: candidate}
			if next.Action == w.Recover && o.Mode == "queryable" {
				receipt, e = callSink(o.Socket, Request{Action: "query", Mode: o.Mode, Receipt: Receipt{Key: key}})
				if e != nil {
					return out, e
				}
			}
			if receipt == nil {
				receipt, e = callSink(o.Socket, Request{Action: "put", Mode: o.Mode, Receipt: request, Hold: o.Hold})
				if e != nil {
					return out, e
				}
			}
			if receipt.Key != key || receipt.Input != request.Input || receipt.Subject != request.Subject || receipt.Source != j.Source || receipt.Run != j.Run || receipt.Step != "deliver" || digest(receipt.Payload) != request.Input {
				return out, fmt.Errorf("receipt identity mismatch")
			}
			result.Evidence = []w.Evidence{deliveryEvidence(*receipt)}
		case "complete":
			if j.Receipt == nil {
				return out, fmt.Errorf("missing receipt")
			}
		default:
			return out, fmt.Errorf("unknown step")
		}
		if e = j.event(result, receipt); e != nil {
			return out, e
		}
		if next.Step == "deliver" {
			handshake(o.Pause, "receipt")
		}
	}
	return out, fmt.Errorf("bounded continuation exhausted")
}
func deliveryEvidence(r Receipt) w.Evidence {
	return w.Evidence{Claim: "delivered", Subject: r.Subject, Verdict: w.Supported, Recipe: w.Recipe{Tool: "local-sink", Version: "1", Command: "query operation " + r.Key, InputDigest: r.Input, Seed: r.ID}}
}
func receiveTerminal(j *Journal, o WorkerOptions) error {
	var terminal LateTerminal
	if e := readJSON(o.Terminal, &terminal); e != nil {
		return e
	}
	if terminal.Incarnation != j.Incarnation && o.Mutant != "old-terminal" {
		return j.append(Record{Kind: "rejected", Incarnation: terminal.Incarnation, Reason: "terminal from replaced incarnation " + terminal.Incarnation + "; current " + j.Incarnation})
	}
	last := j.Events[len(j.Events)-1].Event
	r := terminal.Receipt
	if last.Type != w.EffectPrepared || last.Step != "deliver" || last.EffectID != r.Key || r.Source != j.Source {
		return fmt.Errorf("terminal does not match pending operation")
	}
	return j.event(w.Event{Type: w.StepSucceeded, Step: "deliver", Attempt: last.Attempt, EffectID: r.Key, Subject: r.Subject, Evidence: []w.Evidence{deliveryEvidence(r)}}, &r)
}
