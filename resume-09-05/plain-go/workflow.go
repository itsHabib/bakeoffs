package main

import (
	"bytes"
	"fmt"
	"os"
	"path/filepath"
)

type subject struct{ Run, Source, Candidate, Evidence, Key string }
type evidence struct {
	Source, Candidate, Defect string
	NegativeError             string
}
type candidate struct{ Source, Body string }
type receipt struct{ Key, Payload, ID, Mode string }
type terminal struct {
	Incarnation string
	Subject     subject
	Receipt     receipt
}
type binding struct{ Run, Mode string }
type config struct{ Dir, Run, Mode, Incarnation, Boundary, Mutant string }

func candidateBytes(source []byte) []byte {
	return encode(candidate{digest(source), "artifact for " + string(source)})
}
func check(source, artifact []byte) error {
	if !bytes.Equal(candidateBytes(source), artifact) {
		return fmt.Errorf("candidate content differs from deterministic source transformation")
	}
	return nil
}
func boundary(c config, name string) error {
	if c.Boundary != name {
		return nil
	}
	fmt.Println("BOUNDARY " + name)
	var one [1]byte
	_, e := os.Stdin.Read(one[:])
	return e
}
func operate(c config) error {
	if c.Run == "" || c.Incarnation == "" || (c.Mode != "queryable" && c.Mode != "opaque") {
		return fmt.Errorf("invalid run, incarnation or sink mode")
	}
	j, e := loadJournal(filepath.Join(c.Dir, "journal.jsonl"))
	if e != nil {
		return e
	}
	b := binding{c.Run, c.Mode}
	if len(j.Records) == 0 {
		if e = j.add("binding", b); e != nil {
			return e
		}
	}
	if j.Records[0].Kind != "binding" || decode[binding](j.Records[0].Data) != b {
		return fmt.Errorf("REFUSED_RUN_BINDING")
	}
	if e = j.add("active", c.Incarnation); e != nil {
		return e
	}
	if e = boundary(c, "activated"); e != nil {
		return e
	}
	src, e := os.ReadFile(filepath.Join(c.Dir, "source.txt"))
	if e != nil {
		return e
	}
	sd := digest(src)
	base := filepath.Join(c.Dir, "subjects", sd)
	if e = os.MkdirAll(base, 0700); e != nil {
		return e
	}
	if e = durable(filepath.Join(base, "source.txt"), src); e != nil {
		return e
	}
	s := subject{Run: c.Run, Source: sd}
	var ev evidence
	var got receipt
	hasIntent := false
	completed := false
	for _, r := range j.Records {
		switch r.Kind {
		case "checked":
			x := decode[subject](r.Data)
			if x.Source == sd {
				s = x
			}
		case "intent":
			x := decode[subject](r.Data)
			if x.Source == sd {
				hasIntent = true
			}
		case "receipt":
			x := decode[terminal](r.Data)
			if x.Subject.Source == sd {
				got = x.Receipt
			}
		case "completed":
			x := decode[terminal](r.Data)
			if x.Subject.Source == sd {
				completed = true
			}
		}
	}
	artifactPath := filepath.Join(base, "candidate.json")
	if s.Evidence == "" {
		artifact := candidateBytes(src)
		bad := append(append([]byte{}, artifact...), []byte(" defect")...)
		if e = check(src, artifact); e != nil {
			return e
		}
		neg := check(src, bad)
		if neg == nil {
			return fmt.Errorf("negative control escaped checker")
		}
		ev = evidence{Source: sd, Candidate: digest(artifact), Defect: digest(bad), NegativeError: neg.Error()}
		s.Candidate = ev.Candidate
		s.Evidence = digest(encode(ev))
		s.Key = digest(encode([]string{c.Run, "deliver", sd, s.Candidate, s.Evidence}))
		for path, data := range map[string][]byte{"candidate.json": artifact, "defect.json": bad, "evidence.json": encode(ev)} {
			if e = durable(filepath.Join(base, path), data); e != nil {
				return e
			}
		}
		if e = j.add("checked", s); e != nil {
			return e
		}
	}
	evBytes, e := os.ReadFile(filepath.Join(base, "evidence.json"))
	if e != nil {
		return e
	}
	ev = decode[evidence](evBytes)
	artifact, e := os.ReadFile(artifactPath)
	if e != nil {
		return e
	}
	if digest(evBytes) != s.Evidence || ev.Source != sd || ev.Candidate != s.Candidate {
		return fmt.Errorf("REFUSED_EVIDENCE_IDENTITY")
	}
	if c.Mutant != "stale-evidence" && (digest(artifact) != s.Candidate || check(src, artifact) != nil) {
		return fmt.Errorf("REFUSED_CANDIDATE_CHANGED")
	}
	if completed {
		fmt.Println("COMPLETED_ALREADY " + s.Key)
		return nil
	}
	if e = boundary(c, "checked"); e != nil {
		return e
	}
	if !hasIntent {
		if e = j.add("intent", s); e != nil {
			return e
		}
		if e = boundary(c, "intent_durable"); e != nil {
			return e
		}
	}
	if got.ID == "" && hasIntent {
		if c.Mode == "opaque" && c.Mutant != "blind-retry" {
			if e = j.add("unresolved", s); e != nil {
				return e
			}
			fmt.Println("UNRESOLVED " + s.Key)
			return nil
		}
		if c.Mode == "queryable" {
			got, e = callSink(c, "query", s, false)
			if e != nil {
				return e
			}
		}
	}
	if got.ID == "" {
		got, e = callSink(c, "apply", s, c.Boundary == "sink_committed")
		if e != nil {
			return e
		}
	}
	if got.Key != s.Key || got.Payload != s.Candidate || got.Mode != c.Mode {
		return fmt.Errorf("REFUSED_RECEIPT_IDENTITY")
	}
	t := terminal{c.Incarnation, s, got}
	if e = j.add("receipt", t); e != nil {
		return e
	}
	if e = boundary(c, "receipt_durable"); e != nil {
		return e
	}
	if e = acceptTerminal(j, t, c.Mutant); e != nil {
		return e
	}
	fmt.Println("COMPLETED " + s.Key)
	return nil
}
func acceptTerminal(j *journal, t terminal, mutant string) error {
	active := ""
	var checked subject
	var stored receipt
	for _, r := range j.Records {
		switch r.Kind {
		case "active":
			active = decode[string](r.Data)
		case "checked":
			checked = decode[subject](r.Data)
		case "receipt":
			x := decode[terminal](r.Data)
			if x.Subject == t.Subject {
				stored = x.Receipt
			}
		}
	}
	if t.Incarnation != active && mutant != "old-terminal" {
		return j.add("rejected_terminal", map[string]string{"reason": "STALE_INCARNATION", "received": t.Incarnation, "active": active})
	}
	if checked != t.Subject || stored != t.Receipt || stored.ID == "" {
		return fmt.Errorf("REFUSED_TERMINAL_IDENTITY")
	}
	root := filepath.Dir(j.path)
	source, e := os.ReadFile(filepath.Join(root, "source.txt"))
	if e != nil {
		return e
	}
	if digest(source) != t.Subject.Source {
		return fmt.Errorf("REFUSED_TERMINAL_SOURCE")
	}
	artifact, e := os.ReadFile(filepath.Join(root, "subjects", t.Subject.Source, "candidate.json"))
	if e != nil {
		return e
	}
	if mutant != "stale-evidence" && (digest(artifact) != t.Subject.Candidate || check(source, artifact) != nil) {
		return fmt.Errorf("REFUSED_CANDIDATE_CHANGED")
	}
	if t.Receipt.Key != t.Subject.Key || t.Receipt.Payload != t.Subject.Candidate {
		return fmt.Errorf("REFUSED_TERMINAL_RECEIPT")
	}
	return j.add("completed", t)
}
