package main

import (
	"flag"
	"fmt"
	"os"
	"path/filepath"
)

func main() {
	if e := run(); e != nil {
		fmt.Fprintln(os.Stderr, e)
		os.Exit(1)
	}
}
func run() error {
	if len(os.Args) < 2 {
		return fmt.Errorf("usage: resume demo|worker|sink|terminal")
	}
	fs := flag.NewFlagSet(os.Args[1], flag.ContinueOnError)
	dir := fs.String("dir", "artifacts", "owned scenario directory")
	mode := fs.String("mode", "queryable", "queryable or opaque")
	run := fs.String("run", "run-1", "logical run identity")
	inc := fs.String("inc", "inc-1", "incarnation")
	stop := fs.String("boundary", "", "test handshake")
	mutant := fs.String("mutant", "", "explicit test-only broken variant")
	action := fs.String("action", "apply", "sink action")
	key := fs.String("key", "", "operation key")
	payload := fs.String("payload", "", "payload digest")
	hold := fs.Bool("hold", false, "withhold sink reply")
	packet := fs.String("packet", "", "terminal packet path")
	if e := fs.Parse(os.Args[2:]); e != nil {
		return e
	}
	switch os.Args[1] {
	case "worker":
		return operate(config{*dir, *run, *mode, *inc, *stop, *mutant})
	case "sink":
		return sink(*dir, *mode, *action, *key, *payload, *hold)
	case "terminal":
		j, e := loadJournal(filepath.Join(*dir, "journal.jsonl"))
		if e != nil {
			return e
		}
		b, e := os.ReadFile(*packet)
		if e != nil {
			return e
		}
		return acceptTerminal(j, decode[terminal](b), *mutant)
	case "demo":
		return runSuite(*dir)
	}
	return fmt.Errorf("unknown command")
}
