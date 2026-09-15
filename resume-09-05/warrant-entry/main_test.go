package main

import (
	"bytes"
	"os"
	"os/exec"
	"path/filepath"
	"testing"
)

var testExecutable string

func TestMain(m *testing.M) {
	dir, e := os.MkdirTemp("/tmp", "warrant-test-bin-")
	if e != nil {
		panic(e)
	}
	testExecutable = filepath.Join(dir, "resume")
	cmd := exec.Command("go", "build", "-o", testExecutable, ".")
	cmd.Env = append(os.Environ(), "GOPROXY=off", "GOTOOLCHAIN=local")
	if b, e := cmd.CombinedOutput(); e != nil {
		os.RemoveAll(dir)
		panic(string(b) + e.Error())
	}
	code := m.Run()
	os.RemoveAll(dir)
	os.Exit(code)
}
func TestCommonCasesWithRealWorkerCrashes(t *testing.T) {
	root := t.TempDir()
	for _, spec := range cases {
		t.Run(spec.name, func(t *testing.T) {
			result := runCase(testExecutable, root, spec)
			if !result.Passed {
				t.Fatalf("%s: %s", result.Case, result.Error)
			}
			t.Log(result.Actual)
		})
	}
}

func TestStrictJournalFramingAndIntegrity(t *testing.T) {
	for _, kind := range []string{"malformed-middle", "empty-record", "valid-json-bit-change", "sequence", "unknown-kind", "torn-tail"} {
		t.Run(kind, func(t *testing.T) {
			dir := t.TempDir()
			j := &Journal{Path: filepath.Join(dir, "journal.ndjson"), Mode: "queryable"}
			if e := j.append(Record{Kind: "header", Run: "test", Mode: j.Mode, Definition: definition(j.Mode).Digest}); e != nil {
				t.Fatal(e)
			}
			if e := j.append(Record{Kind: "incarnation", Incarnation: "first"}); e != nil {
				t.Fatal(e)
			}
			data, e := os.ReadFile(j.Path)
			if e != nil {
				t.Fatal(e)
			}
			switch kind {
			case "empty-record":
				data = append(data, '\n')
			case "malformed-middle":
				data = append([]byte("{invalid}\n"), data...)
			case "valid-json-bit-change":
				data = bytes.Replace(data, []byte(`"first"`), []byte(`"other"`), 1)
			case "sequence":
				data = bytes.Replace(data, []byte(`"seq":2`), []byte(`"seq":9`), 1)
			case "unknown-kind":
				r := j.Records[1]
				r.Kind = "magic"
				r.Hash = recordHash(r)
				data = append(append(encoded(j.Records[0]), '\n'), append(encoded(r), '\n')...)
			case "torn-tail":
				data = append(data, []byte(`{"seq":3`)...)
			}
			if e = save(j.Path, data); e != nil {
				t.Fatal(e)
			}
			loaded, e := loadJournal(dir, "queryable")
			if kind != "torn-tail" {
				if e == nil {
					t.Fatal("corruption accepted")
				}
				return
			}
			if e != nil || len(loaded.Records) != 2 || string(loaded.Tail) != `{"seq":3` {
				t.Fatalf("bad prefix recovery: %+v %v", loaded, e)
			}
			if e = loaded.repairTail(); e != nil {
				t.Fatal(e)
			}
			if e = loaded.append(Record{Kind: "incarnation", Incarnation: "replacement"}); e != nil {
				t.Fatal(e)
			}
			again, e := loadJournal(dir, "queryable")
			if e != nil || len(again.Records) != 3 {
				t.Fatalf("cannot append after torn tail: %v", e)
			}
		})
	}
}

func TestMutantsUnavailableByDefault(t *testing.T) {
	t.Setenv("BAKEOFF_MUTANTS", "")
	_, e := worker(WorkerOptions{Dir: t.TempDir(), Mode: "queryable", Revision: "S1", Mutant: "blind-retry"})
	if e == nil {
		t.Fatal("test-only mutant entered normal execution path")
	}
}

func TestCheckerUsesActualCandidateContent(t *testing.T) {
	source := mustSource("S1")
	good := candidateBytes(source)
	if !check(source, good) {
		t.Fatal("positive candidate failed")
	}
	bad := bytes.Replace(good, []byte(`"result":14`), []byte(`"result":15`), 1)
	if check(source, bad) {
		t.Fatal("same checker accepted planted defect")
	}
	if check(mustSource("S2"), good) {
		t.Fatal("wrong source accepted")
	}
}
