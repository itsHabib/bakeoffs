// Command cache-max is an offline analyzer over Claude Code session
// transcripts. It computes per-session cache accounting, detects prefix busts
// mechanically, attributes them to the largest new context block, and prices
// what stable prefixes would have saved. No model, no network, no keys.
package main

import (
	"flag"
	"fmt"
	"os"
	"path/filepath"
	"sort"

	"hack3-cache-max/internal/analyze"
	"hack3-cache-max/internal/pricing"
	"hack3-cache-max/internal/report"
	"hack3-cache-max/internal/transcript"
)

func main() {
	def := analyze.Default()
	home, _ := os.UserHomeDir()
	dir := flag.String("dir", filepath.Join(home, ".claude", "projects"), "directory of *.jsonl transcripts")
	priceFile := flag.String("pricing", "pricing.json", "path to the $/MTok table")
	bustFrac := flag.Float64("bust-frac", def.BustFrac, "bust when cache-creation exceeds this fraction of the previous cache-read")
	minPrev := flag.Int("min-prev-read", def.MinPrevRead, "minimum previous cache-read for a message to be bust-eligible")
	top := flag.Int("top-causes", 10, "how many bust causes to list")
	sessionsN := flag.Int("sessions", 0, "also print the top-N sessions by savings (0 = skip)")
	htmlOut := flag.String("html", "", "also write a static HTML report to this path")
	flag.Parse()

	rates, usingExample, err := pricing.Load(*priceFile)
	if err != nil {
		fmt.Fprintf(os.Stderr, "cache-max: pricing: %v\n", err)
		os.Exit(1)
	}

	files, err := findTranscripts(*dir)
	if err != nil {
		fmt.Fprintf(os.Stderr, "cache-max: %v\n", err)
		os.Exit(1)
	}
	if len(files) == 0 {
		fmt.Fprintf(os.Stderr, "cache-max: no *.jsonl under %s\n", *dir)
		os.Exit(1)
	}

	params := analyze.Params{BustFrac: *bustFrac, MinPrevRead: *minPrev}
	stats := make([]analyze.SessionStats, 0, len(files))
	for _, f := range files {
		s, err := transcript.ParseFile(f)
		if err != nil {
			fmt.Fprintf(os.Stderr, "cache-max: skip %s: %v\n", f, err)
			continue
		}
		if s.Turns == nil {
			continue
		}
		st := analyze.Session(s, params)
		if st.Assistants > 0 {
			stats = append(stats, st)
		}
	}

	corpus := analyze.Aggregate(stats, rates)
	report.Text(os.Stdout, corpus, usingExample, *top)
	if *sessionsN > 0 {
		fmt.Println()
		fmt.Println("-- Sessions by savings -----------------------------------------")
		report.Sessions(os.Stdout, corpus, *sessionsN)
	}
	if *htmlOut != "" {
		if err := writeHTML(*htmlOut, corpus, usingExample); err != nil {
			fmt.Fprintf(os.Stderr, "cache-max: html: %v\n", err)
			os.Exit(1)
		}
		fmt.Printf("\nHTML report written to %s\n", *htmlOut)
	}
}

func findTranscripts(dir string) ([]string, error) {
	var out []string
	err := filepath.WalkDir(dir, func(p string, d os.DirEntry, err error) error {
		if err != nil {
			return nil // tolerate unreadable subtrees
		}
		if !d.IsDir() && filepath.Ext(p) == ".jsonl" {
			out = append(out, p)
		}
		return nil
	})
	sort.Strings(out)
	return out, err
}

func writeHTML(path string, c analyze.Corpus, usingExample bool) error {
	f, err := os.Create(path)
	if err != nil {
		return err
	}
	defer f.Close()
	report.HTML(f, c, usingExample)
	return nil
}
