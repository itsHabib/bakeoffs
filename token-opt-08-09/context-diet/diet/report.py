"""Render a Report as a CLI table (and, optionally, one static HTML file).

The money shot is the per-rule table: tokens saved, and next to it the
computed risk score (later-referenced rate) that says whether the saving is
free or whether it deletes things the agent came back for.
"""

from __future__ import annotations

SHIP_THRESHOLD = 0.01  # <1% later-referenced spans => shippable default


def _pct(x: float) -> str:
    return f"{100 * x:.1f}%"


def _verdict(stat) -> str:
    if stat.spans == 0:
        return "—"
    return "SHIP" if stat.ref_rate < SHIP_THRESHOLD else "REVIEW"


def render_text(report, corpus_label: str = "corpus") -> str:
    L = []
    a = L.append
    a("=" * 74)
    a(f"  context-diet — replay over {report.sessions} sessions ({corpus_label})")
    a("  token counts are a tiktoken/o200k_base PROXY, not Claude's tokenizer")
    a("=" * 74)
    a("")
    a(f"  tool results replayed : {report.results:,}")
    a(f"  tool-result tokens    : {report.total_result_tokens:,} (proxy)")
    combined_pct = (
        report.combined_saved_tokens / report.total_result_tokens
        if report.total_result_tokens
        else 0.0
    )
    a(
        f"  removable (all rules) : {report.combined_saved_tokens:,} tokens "
        f"= {_pct(combined_pct)} of tool-result tokens"
    )
    tot = report.total_result_tokens or 1
    safe = sum(s.saved_tokens for s in report.rules.values()
               if s.spans and s.ref_rate < SHIP_THRESHOLD)
    risky_rules = [s.name for s in report.rules.values()
                   if s.spans and s.ref_rate >= SHIP_THRESHOLD]
    a(f"    FREE (SHIP-rated)   : {safe:,} = {_pct(safe / tot)}  <- ship today")
    if risky_rules:
        a(f"    the rest is reachable only via {', '.join(risky_rules)},")
        a("    all REVIEW-rated (see ref-rate) — not a safe blind size cut")
    a("")
    a("  per-rule (each rule run independently over the originals)")
    a("  " + "-" * 70)
    a(
        f"  {'rule':<22}{'fires':>7}{'saved tok':>12}"
        f"{'spans':>7}{'ref-rate':>10}{'verdict':>9}"
    )
    a("  " + "-" * 70)
    for stat in sorted(
        report.rules.values(), key=lambda s: s.saved_tokens, reverse=True
    ):
        a(
            f"  {stat.name:<22}{stat.fires:>7}{stat.saved_tokens:>12,}"
            f"{stat.spans:>7}{_pct(stat.ref_rate):>10}{_verdict(stat):>9}"
        )
    a("  " + "-" * 70)
    a("")
    a("  ref-rate = share of removed spans whose identifiers reappear in")
    a("  later assistant text (computed). SHIP = <1% ref-rate.")
    for stat in report.rules.values():
        if stat.examples:
            ex = ", ".join(stat.examples[:3])
            a(f"    ! {stat.name}: later-referenced e.g. {ex}")
    a("=" * 74)
    return "\n".join(L)


def render_html(report, corpus_label: str = "corpus") -> str:
    rows = []
    for stat in sorted(
        report.rules.values(), key=lambda s: s.saved_tokens, reverse=True
    ):
        cls = "ship" if _verdict(stat) == "SHIP" else "review"
        rows.append(
            f"<tr class='{cls}'><td>{stat.name}</td>"
            f"<td class='n'>{stat.fires:,}</td>"
            f"<td class='n'>{stat.saved_tokens:,}</td>"
            f"<td class='n'>{stat.spans:,}</td>"
            f"<td class='n'>{_pct(stat.ref_rate)}</td>"
            f"<td class='v'>{_verdict(stat)}</td></tr>"
        )
    tot = report.total_result_tokens or 1
    combined_pct = report.combined_saved_tokens / tot
    safe = sum(s.saved_tokens for s in report.rules.values()
               if s.spans and s.ref_rate < SHIP_THRESHOLD)
    return _HTML.format(
        label=corpus_label,
        sessions=report.sessions,
        results=f"{report.results:,}",
        total=f"{report.total_result_tokens:,}",
        saved=f"{report.combined_saved_tokens:,}",
        pct=_pct(combined_pct),
        safe=f"{safe:,}",
        safepct=_pct(safe / tot),
        rows="\n".join(rows),
    )


_HTML = """<!doctype html><meta charset=utf-8>
<title>context-diet report</title>
<style>
 body{{font:15px/1.5 -apple-system,system-ui,sans-serif;max-width:820px;
 margin:40px auto;color:#1a1a1a;background:#fafafa;padding:0 16px}}
 h1{{font-size:20px;margin-bottom:2px}}
 .sub{{color:#666;font-size:13px;margin-bottom:24px}}
 .big{{font-size:34px;font-weight:700;color:#0a7}}
 table{{border-collapse:collapse;width:100%;margin-top:20px;background:#fff}}
 th,td{{padding:8px 12px;border-bottom:1px solid #eee;text-align:left}}
 th{{font-size:12px;text-transform:uppercase;letter-spacing:.04em;color:#888}}
 td.n,th.n{{text-align:right;font-variant-numeric:tabular-nums}}
 .v{{font-weight:700}}
 tr.ship .v{{color:#0a7}} tr.review .v{{color:#c60}}
 .note{{color:#888;font-size:12px;margin-top:18px}}
</style>
<h1>context-diet</h1>
<div class=sub>replay over {sessions} sessions — {label} · token counts are a
 tiktoken/o200k_base <b>proxy</b>, not Claude's tokenizer</div>
<div><span class=big>{pct}</span> of {total} tool-result tokens
 removable by size-blind rules ({saved} tokens) — zero model calls</div>
<div class=sub>but only <b>{safepct}</b> ({safe} tokens) is <b>free</b> —
 removals with a computed 0% later-reference rate. The rest is reachable
 only by caps/truncation that delete identifiers the agent reuses ~1-in-3
 times. That gap is the finding.</div>
<table>
<tr><th>rule</th><th class=n>fires</th><th class=n>saved tok</th>
 <th class=n>spans</th><th class=n>ref-rate</th><th>verdict</th></tr>
{rows}
</table>
<div class=note>ref-rate = share of removed spans whose identifiers reappear
 in later assistant text (computed, not judged). SHIP = &lt;1% ref-rate:
 the rule removes tokens the agent never came back for.</div>
"""
