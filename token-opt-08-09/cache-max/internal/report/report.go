// Package report renders analysis results as plain text. Presentation only —
// it computes nothing billable.
package report

import (
	"fmt"
	"io"
	"path/filepath"
	"sort"

	"hack3-cache-max/internal/analyze"
)

// TargetMonthlyTokens is the portfolio's stated monthly agent-token volume that
// the corpus savings are extrapolated to.
const TargetMonthlyTokens = 5_000_000_000

// Text writes the full corpus report.
func Text(w io.Writer, c analyze.Corpus, usingExample bool, topCauses int) {
	fmt.Fprintln(w, "cache-max — prefix-bust accounting over Claude Code transcripts")
	fmt.Fprintln(w, "================================================================")
	if usingExample {
		fmt.Fprintln(w, "!! EXAMPLE PRICES: pricing.json rates are still TODO — dollar")
		fmt.Fprintln(w, "!! figures use illustrative placeholders. Fill real rates from")
		fmt.Fprintln(w, "!! https://docs.claude.com/en/docs/about-claude/pricing")
	}
	fmt.Fprintln(w)

	fmt.Fprintf(w, "Sessions analyzed      %d\n", len(c.Sessions))
	fmt.Fprintf(w, "Assistant messages     %d\n", assistants(c))
	fmt.Fprintf(w, "Observed session span  %.1f h\n", c.SpanHours)
	fmt.Fprintln(w)

	fmt.Fprintln(w, "-- Cache accounting --------------------------------------------")
	fmt.Fprintf(w, "Cache-read tokens      %s\n", commas(c.SumRead))
	fmt.Fprintf(w, "Cache-write tokens     %s\n", commas(c.SumCreate))
	fmt.Fprintf(w, "Fresh-input tokens     %s\n", commas(c.SumInput))
	fmt.Fprintf(w, "Output tokens          %s\n", commas(c.SumOutput))
	fmt.Fprintf(w, "HIT RATIO              %.1f%%   (read / (read+write+fresh))\n", 100*c.HitRatio())
	fmt.Fprintln(w)

	fmt.Fprintln(w, "-- Prefix busts ------------------------------------------------")
	fmt.Fprintf(w, "Total busts            %d\n", c.TotalBusts)
	fmt.Fprintf(w, "Busts / hour           %.2f\n", c.BustsPerHour())
	fmt.Fprintln(w)

	fmt.Fprintln(w, "-- Bust causes (correlational: largest new context at the bust) ")
	if len(c.Causes) == 0 {
		fmt.Fprintln(w, "  (none)")
	}
	fmt.Fprintf(w, "  %-22s %7s %14s %9s\n", "cause", "busts", "re-written", "sessions")
	for i, row := range c.Causes {
		if topCauses > 0 && i >= topCauses {
			break
		}
		fmt.Fprintf(w, "  %-22s %7d %14s %9d\n", row.Bucket, row.Busts, commas(row.Rewritten), row.Sessions)
	}
	fmt.Fprintln(w)

	fmt.Fprintln(w, "-- What-if: stable prefixes ------------------------------------")
	fmt.Fprintf(w, "Cost as-happened       $%.2f\n", c.Cost.AsHappened)
	fmt.Fprintf(w, "Cost with stable prefix $%.2f\n", c.Cost.Counterfactual)
	fmt.Fprintf(w, "SAVINGS (busts repriced as reads)  $%.2f  (%.1f%% of spend)\n",
		c.Cost.Savings(), pct(c.Cost.Savings(), c.Cost.AsHappened))
	extrap := c.ExtrapolatedSavings(TargetMonthlyTokens)
	fmt.Fprintf(w, "EXTRAPOLATED to %s tok/mo:  $%.2f / month\n", commas(TargetMonthlyTokens), extrap)
	fmt.Fprintln(w)

	headline(w, c)
	fmt.Fprintln(w)
	fmt.Fprintln(w, "Assumptions: (1) a stable prefix would have stayed cache-resident")
	fmt.Fprintln(w, "across the session's TTL; excess above threshold is repriced read.")
	fmt.Fprintln(w, "(2) Attribution is correlational — the largest new context block at")
	fmt.Fprintln(w, "each bust, not a proven cause. (3) Extrapolation assumes the same")
	fmt.Fprintln(w, "cache behavior at the target monthly volume.")
}

func headline(w io.Writer, c analyze.Corpus) {
	cause := "n/a"
	var bustShare, sessShare float64
	if len(c.Causes) > 0 {
		cause = c.Causes[0].Bucket
		if c.TotalBusts > 0 {
			bustShare = 100 * float64(c.Causes[0].Busts) / float64(c.TotalBusts)
		}
		if n := len(c.Sessions); n > 0 {
			sessShare = 100 * float64(c.Causes[0].Sessions) / float64(n)
		}
	}
	fmt.Fprintln(w, "================================================================")
	fmt.Fprintf(w, "HEADLINE: %.0f%% hit rate; %d busts cost $%.2f; top cause %s\n",
		100*c.HitRatio(), c.TotalBusts, c.Cost.Savings(), cause)
	fmt.Fprintf(w, "          (%.0f%% of busts, seen in %.0f%% of sessions).\n", bustShare, sessShare)
	fmt.Fprintf(w, "          At %s tok/mo that's $%.2f/mo for zero product-code change.\n",
		commas(TargetMonthlyTokens), c.ExtrapolatedSavings(TargetMonthlyTokens))
	fmt.Fprintln(w, "================================================================")
}

// Sessions prints a compact per-session table sorted by savings.
func Sessions(w io.Writer, c analyze.Corpus, limit int) {
	type row struct {
		st   analyze.SessionStats
		save float64
	}
	rows := make([]row, 0, len(c.Sessions))
	for _, st := range c.Sessions {
		rows = append(rows, row{st, analyze.SessionCost(st, c.Rates).Savings()})
	}
	sort.Slice(rows, func(i, j int) bool { return rows[i].save > rows[j].save })
	fmt.Fprintf(w, "  %-40s %5s %6s %10s\n", "session", "hit%", "busts", "savings$")
	for i, r := range rows {
		if limit > 0 && i >= limit {
			break
		}
		fmt.Fprintf(w, "  %-40s %4.0f%% %6d %10.2f\n",
			trunc(filepath.Base(r.st.Path), 40), 100*r.st.HitRatio(), len(r.st.Busts), r.save)
	}
}

func assistants(c analyze.Corpus) int {
	n := 0
	for _, s := range c.Sessions {
		n += s.Assistants
	}
	return n
}

func pct(part, whole float64) float64 {
	if whole == 0 {
		return 0
	}
	return 100 * part / whole
}

func commas(n int) string {
	s := fmt.Sprintf("%d", n)
	neg := ""
	if len(s) > 0 && s[0] == '-' {
		neg, s = "-", s[1:]
	}
	var out []byte
	for i, d := range []byte(s) {
		if i > 0 && (len(s)-i)%3 == 0 {
			out = append(out, ',')
		}
		out = append(out, d)
	}
	return neg + string(out)
}

func trunc(s string, n int) string {
	if len(s) <= n {
		return s
	}
	return s[:n-1] + "…"
}
