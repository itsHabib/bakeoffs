package main

import (
	"flag"
	"fmt"
	"os"
	"path/filepath"

	w "github.com/itsHabib/warrant"
)

func main() {
	if e := run(os.Args[1:]); e != nil {
		fmt.Fprintln(os.Stderr, e)
		os.Exit(1)
	}
}
func run(args []string) error {
	if len(args) == 0 {
		return fmt.Errorf("usage: resume demo|worker|sink|replay|check")
	}
	f := flag.NewFlagSet(args[0], flag.ContinueOnError)
	dir := f.String("dir", "", "owned state directory")
	socket := f.String("socket", "", "local sink socket")
	mode := f.String("mode", "queryable", "queryable or opaque")
	revision := f.String("revision", "S1", "S1 or S2")
	pause := f.String("pause", "", "controller handshake")
	hold := f.Bool("hold", false, "sink withholds response")
	terminal := f.String("terminal", "", "incoming terminal file")
	mutant := f.String("mutant", "", "test-only defect")
	source := f.String("source", "", "source JSON")
	candidate := f.String("candidate", "", "candidate JSON")
	if e := f.Parse(args[1:]); e != nil {
		return e
	}
	switch args[0] {
	case "worker":
		out, e := worker(WorkerOptions{Dir: *dir, Socket: *socket, Mode: *mode, Revision: *revision, Pause: *pause, Hold: *hold, Terminal: *terminal, Mutant: *mutant})
		if e != nil {
			out.Status = "refused"
			out.Reason = e.Error()
		}
		writeErr := saveJSON(filepath.Join(*dir, "outcome.json"), out)
		fmt.Println(string(encoded(out)))
		if e != nil {
			return e
		}
		return writeErr
	case "sink":
		return serveSink(*dir, *socket, *mode)
	case "replay":
		j, e := loadJournal(*dir, *mode)
		if e != nil {
			return e
		}
		v, n := w.Reduce(definition(*mode), j.Events)
		fmt.Println(string(encoded(struct {
			View               w.View
			Next               w.NextAction
			TruncatedTailBytes int
		}{v, n, len(j.Tail)})))
		return nil
	case "check":
		s, e := os.ReadFile(*source)
		if e != nil {
			return e
		}
		c, e := os.ReadFile(*candidate)
		if e != nil {
			return e
		}
		ev := evidence(s, c)
		fmt.Println(string(encoded(ev)))
		if !check(s, c) {
			return fmt.Errorf("checker rejected candidate")
		}
		return nil
	case "demo":
		exe, e := os.Executable()
		if e != nil {
			return e
		}
		return demo(exe, *dir)
	}
	return fmt.Errorf("unknown command %q", args[0])
}
