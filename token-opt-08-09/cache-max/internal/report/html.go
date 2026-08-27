package report

import (
	"fmt"
	"html"
	"io"

	"hack3-cache-max/internal/analyze"
)

// HTML writes a single self-contained static page (inline CSS + data, no deps,
// no network). Optional sugar over the CLI report.
func HTML(w io.Writer, c analyze.Corpus, usingExample bool) {
	fmt.Fprint(w, `<!doctype html><meta charset=utf-8><title>cache-max</title>
<style>
:root{color-scheme:light dark}
body{font:15px/1.5 -apple-system,system-ui,sans-serif;max-width:820px;margin:2rem auto;padding:0 1rem}
h1{font-size:1.4rem;margin-bottom:.2rem} .sub{opacity:.65;margin-top:0}
.grid{display:grid;grid-template-columns:repeat(auto-fit,minmax(150px,1fr));gap:.75rem;margin:1rem 0}
.card{border:1px solid #8884;border-radius:10px;padding:.8rem 1rem}
.card .n{font-size:1.5rem;font-weight:600} .card .l{opacity:.65;font-size:.8rem}
.hl{border:2px solid #3a7;border-radius:10px;padding:1rem;margin:1rem 0;background:#3a71}
table{border-collapse:collapse;width:100%;margin:.5rem 0} th,td{text-align:left;padding:.35rem .6rem;border-bottom:1px solid #8883}
td.num,th.num{text-align:right;font-variant-numeric:tabular-nums}
.warn{border:2px solid #c94;background:#c941;border-radius:8px;padding:.6rem .9rem;margin:.6rem 0}
small{opacity:.6}
</style>
`)
	fmt.Fprintf(w, "<h1>cache-max</h1><p class=sub>prefix-bust accounting over %d Claude Code sessions</p>", len(c.Sessions))
	if usingExample {
		fmt.Fprint(w, `<div class=warn><b>Example prices.</b> pricing.json rates are TODO — dollar figures are illustrative placeholders.</div>`)
	}

	extrap := c.ExtrapolatedSavings(TargetMonthlyTokens)
	cause := "n/a"
	if len(c.Causes) > 0 {
		cause = c.Causes[0].Bucket
	}
	fmt.Fprintf(w, `<div class=hl><b>Headline.</b> %.0f%% hit rate; %d busts cost $%.2f; top cause <code>%s</code>. Extrapolated to %s tok/mo: <b>$%.2f/month</b> for zero product-code change.</div>`,
		100*c.HitRatio(), c.TotalBusts, c.Cost.Savings(), html.EscapeString(cause), commas(TargetMonthlyTokens), extrap)

	fmt.Fprint(w, `<div class=grid>`)
	card(w, fmt.Sprintf("%.1f%%", 100*c.HitRatio()), "cache hit ratio")
	card(w, commas(c.TotalBusts), "prefix busts")
	card(w, fmt.Sprintf("%.2f", c.BustsPerHour()), "busts / hour")
	card(w, fmt.Sprintf("$%.2f", c.Cost.Savings()), "savings (observed)")
	card(w, fmt.Sprintf("$%.0f", extrap), "savings / mo @ 5B")
	fmt.Fprint(w, `</div>`)

	fmt.Fprint(w, `<h3>Bust causes <small>(correlational)</small></h3><table><tr><th>cause</th><th class=num>busts</th><th class=num>re-written tokens</th><th class=num>sessions</th></tr>`)
	for _, row := range c.Causes {
		fmt.Fprintf(w, `<tr><td><code>%s</code></td><td class=num>%d</td><td class=num>%s</td><td class=num>%d</td></tr>`,
			html.EscapeString(row.Bucket), row.Busts, commas(row.Rewritten), row.Sessions)
	}
	fmt.Fprint(w, `</table>`)
	fmt.Fprint(w, `<p><small>Attribution is correlational — the largest new context block at each bust, not a proven cause. Savings reprice each bust's excess cache-creation at the read rate, assuming a stable prefix would have stayed cached across the session TTL.</small></p>`)
}

func card(w io.Writer, n, label string) {
	fmt.Fprintf(w, `<div class=card><div class=n>%s</div><div class=l>%s</div></div>`, n, html.EscapeString(label))
}
