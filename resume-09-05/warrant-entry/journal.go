package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"path/filepath"

	w "github.com/itsHabib/warrant"
)

// Outer records own durability, definitions, source epochs and incarnation fences.
// Warrant events keep their upstream meanings and their own per-epoch sequence.
type Record struct {
	Seq         int         `json:"seq"`
	Prev        string      `json:"prev"`
	Hash        string      `json:"hash"`
	Kind        string      `json:"kind"`
	Run         string      `json:"run,omitempty"`
	Mode        string      `json:"mode,omitempty"`
	Definition  string      `json:"definition,omitempty"`
	Incarnation string      `json:"incarnation,omitempty"`
	Revision    string      `json:"revision,omitempty"`
	Source      string      `json:"source,omitempty"`
	Event       *w.Envelope `json:"event,omitempty"`
	Receipt     *Receipt    `json:"receipt,omitempty"`
	Reason      string      `json:"reason,omitempty"`
}
type Journal struct {
	Path                                     string
	Records                                  []Record
	Events                                   []w.Envelope
	Run, Mode, Incarnation, Revision, Source string
	Receipt                                  *Receipt
	Tail                                     []byte
}

func recordHash(r Record) string { r.Hash = ""; return digest(encoded(r)) }

func loadJournal(dir, mode string) (*Journal, error) {
	j := &Journal{Path: filepath.Join(dir, "journal.ndjson")}
	data, e := os.ReadFile(j.Path)
	if os.IsNotExist(e) {
		return j, nil
	}
	if e != nil {
		return nil, e
	}
	end := bytes.LastIndexByte(data, '\n') + 1
	j.Tail = append([]byte{}, data[end:]...)
	lines := bytes.Split(data[:end], []byte{'\n'})
	for i, line := range lines {
		if i == len(lines)-1 && len(line) == 0 {
			continue
		}
		if len(line) == 0 {
			return nil, fmt.Errorf("empty complete journal record")
		}
		var r Record
		d := json.NewDecoder(bytes.NewReader(line))
		d.DisallowUnknownFields()
		if e = d.Decode(&r); e != nil {
			return nil, fmt.Errorf("corrupt complete journal record: %w", e)
		}
		var extra any
		if d.Decode(&extra) != io.EOF {
			return nil, fmt.Errorf("trailing data in complete record")
		}
		prev := ""
		if len(j.Records) > 0 {
			prev = j.Records[len(j.Records)-1].Hash
		}
		if r.Seq != len(j.Records)+1 || r.Prev != prev || r.Hash != recordHash(r) {
			return nil, fmt.Errorf("journal integrity refusal at record %d", r.Seq)
		}
		if e = j.apply(r, mode); e != nil {
			return nil, e
		}
		j.Records = append(j.Records, r)
	}
	if len(j.Records) == 0 && len(data) > 0 {
		return nil, fmt.Errorf("journal has no durable header")
	}
	return j, nil
}
func (j *Journal) apply(r Record, mode string) error {
	switch r.Kind {
	case "header":
		if len(j.Records) != 0 || r.Run == "" || r.Mode != mode || r.Definition != definition(mode).Digest {
			return fmt.Errorf("definition/header mismatch")
		}
		j.Run = r.Run
		j.Mode = r.Mode
	case "incarnation":
		if j.Run == "" || r.Incarnation == "" {
			return fmt.Errorf("invalid incarnation")
		}
		j.Incarnation = r.Incarnation
	case "source":
		source, e := sourceBytes(r.Revision)
		if e != nil || r.Source != digest(source) {
			return fmt.Errorf("source identity mismatch")
		}
		if len(j.Events) > 0 {
			_, n := w.Reduce(definition(mode), j.Events)
			if n.Action != w.NothingToDo {
				return fmt.Errorf("unresolved previous source; resume it first")
			}
		}
		j.Revision = r.Revision
		j.Source = r.Source
		j.Events = nil
		j.Receipt = nil
	case "event":
		if r.Event == nil || r.Incarnation != j.Incarnation || j.Source == "" {
			return fmt.Errorf("event identity refusal")
		}
		env := *r.Event
		ev := env.Event
		p := definition(mode)
		if env.Sequence != len(j.Events)+1 || env.Version != 1 {
			return fmt.Errorf("event sequence refusal")
		}
		if ev.Type == w.RunStarted {
			if len(j.Events) != 0 || ev.RunID != j.Run || ev.Subject != j.Source || ev.PipelineDigest != p.Digest || ev.PipelineName != p.Name || ev.PipelineVersion != p.Version {
				return fmt.Errorf("run definition mismatch")
			}
		} else {
			if len(j.Events) == 0 {
				return fmt.Errorf("event before start")
			}
			_, n := w.Reduce(p, j.Events)
			switch ev.Type {
			case w.EffectPrepared:
				if n.Action != w.Execute || n.Step != ev.Step || n.Attempt != ev.Attempt || ev.EffectID == "" {
					return fmt.Errorf("invalid prepare transition")
				}
				for _, s := range p.Steps {
					if s.Name == ev.Step && s.ReplayClass != ev.ReplayClass {
						return fmt.Errorf("replay class mismatch")
					}
				}
			case w.StepSucceeded, w.StepFailed:
				last := j.Events[len(j.Events)-1].Event
				if last.Type != w.EffectPrepared || last.EffectID != ev.EffectID || last.Step != ev.Step || last.Attempt != ev.Attempt {
					return fmt.Errorf("terminal without matching intent")
				}
			default:
				return fmt.Errorf("unknown event type %q", ev.Type)
			}
		}
		j.Events = append(j.Events, env)
		if r.Receipt != nil {
			j.Receipt = r.Receipt
		}
	case "rejected", "observation":
		if r.Reason == "" {
			return fmt.Errorf("unnamed observation")
		}
	default:
		return fmt.Errorf("unknown outer record %q", r.Kind)
	}
	return nil
}
func (j *Journal) append(r Record) error {
	r.Seq = len(j.Records) + 1
	if len(j.Records) > 0 {
		r.Prev = j.Records[len(j.Records)-1].Hash
	}
	r.Hash = recordHash(r)
	// Validate on a copy before writing. No state advancement before fsync.
	trial := *j
	trial.Events = append([]w.Envelope{}, j.Events...)
	if e := trial.apply(r, j.Mode); e != nil {
		return e
	}
	f, e := os.OpenFile(j.Path, os.O_CREATE|os.O_APPEND|os.O_WRONLY, 0600)
	if e != nil {
		return e
	}
	b := append(encoded(r), '\n')
	n, e := f.Write(b)
	if e == nil && n != len(b) {
		e = io.ErrShortWrite
	}
	if e == nil {
		e = f.Sync()
	}
	closeErr := f.Close()
	if e == nil {
		e = closeErr
	}
	if e != nil {
		return e
	}
	*j = trial
	j.Records = append(j.Records, r)
	return nil
}
func (j *Journal) repairTail() error {
	if len(j.Tail) == 0 {
		return nil
	}
	if e := save(j.Path+".discarded-tail", j.Tail); e != nil {
		return e
	}
	data, e := os.ReadFile(j.Path)
	if e != nil {
		return e
	}
	f, e := os.OpenFile(j.Path, os.O_WRONLY, 0600)
	if e != nil {
		return e
	}
	if e = f.Truncate(int64(len(data) - len(j.Tail))); e == nil {
		e = f.Sync()
	}
	f.Close()
	j.Tail = nil
	return e
}
func (j *Journal) event(ev w.Event, receipt *Receipt) error {
	return j.append(Record{Kind: "event", Incarnation: j.Incarnation, Event: &w.Envelope{Version: 1, Sequence: len(j.Events) + 1, Event: ev}, Receipt: receipt})
}
func (j *Journal) ledger() []w.Evidence {
	var out []w.Evidence
	for _, e := range j.Events {
		out = append(out, e.Event.Evidence...)
	}
	return out
}
