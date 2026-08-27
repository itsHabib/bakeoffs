#!/usr/bin/env python3
"""hack3-spend-audit — the invoice nobody could show you.

Parse ~/.claude/projects/**/*.jsonl, itemize token spend by category / project /
session / day / tool / model, price it from an editable table, and render a
one-file HTML dashboard + a plain-text CLI report.

No model. No network. Token counts come straight from each assistant message's
`message.usage`; the only estimate is tool-result *volume* (bytes -> tokens),
clearly labeled. See README.md.
"""

from __future__ import annotations

import argparse
import glob
import json
import os
import sys
import tomllib
from collections import defaultdict
from dataclasses import dataclass, field

# ---------------------------------------------------------------------------
# Vocabulary: the four (well, five) things a token can be.
# The wire fields collapse into these categories. Order is the stack order.
# ---------------------------------------------------------------------------
CATEGORIES = [
    ("input", "Fresh input"),
    ("cache_read", "Cache read"),
    ("cache_write_5m", "Cache write (5m)"),
    ("cache_write_1h", "Cache write (1h)"),
    ("output", "Output"),
]
CAT_KEYS = [k for k, _ in CATEGORIES]
CAT_LABEL = dict(CATEGORIES)


# ---------------------------------------------------------------------------
# Records (mechanism produces these; policy consumes them). Pure data.
# ---------------------------------------------------------------------------
@dataclass
class UsageEvent:
    """One assistant message's billable usage, already split by category."""

    day: str
    project: str
    session: str
    model: str
    tokens: dict[str, int]  # category key -> count


@dataclass
class ToolVolume:
    """Estimated token volume of a tool_result, attributed to its tool name."""

    tool: str
    est_tokens: int


@dataclass
class ParseStats:
    files: int = 0
    lines: int = 0
    usage_msgs: int = 0
    dup_msgs: int = 0
    tool_results: int = 0
    orphan_tool_results: int = 0


# ---------------------------------------------------------------------------
# Tokenizer for tool-result volume only. tiktoken if present, else chars/4.
# ---------------------------------------------------------------------------
class Estimator:
    def __init__(self) -> None:
        self.enc = None
        self.method = "chars/4 (install tiktoken for o200k_base)"
        try:
            import tiktoken  # noqa: PLC0415 — optional, loaded lazily

            self.enc = tiktoken.get_encoding("o200k_base")
            self.method = "tiktoken o200k_base (proxy for the Claude tokenizer)"
        except Exception:  # noqa: BLE001 — any failure means fall back cleanly
            pass

    def count(self, text: str) -> int:
        if self.enc is not None:
            return len(self.enc.encode(text, disallowed_special=()))
        return (len(text) + 3) // 4


def _result_text(content) -> str:
    """Flatten a tool_result's content (str | list of blocks) to text."""
    if isinstance(content, str):
        return content
    if isinstance(content, list):
        parts = []
        for block in content:
            if isinstance(block, dict) and isinstance(block.get("text"), str):
                parts.append(block["text"])
            else:
                parts.append(json.dumps(block, ensure_ascii=False))
        return "".join(parts)
    return json.dumps(content, ensure_ascii=False)


# ---------------------------------------------------------------------------
# MECHANISM: read transcripts -> (UsageEvent[], ToolVolume[], stats).
# Knows the file layout and nothing about pricing or presentation.
# ---------------------------------------------------------------------------
def _project_of(cwd: str | None, path: str) -> str:
    """Collapse a cwd to its repo root so worktrees roll into one project."""
    base = cwd or os.path.dirname(path)
    marker = "/.claude/worktrees/"
    if marker in base:
        base = base.split(marker, 1)[0]
    return base


def _usage_from(msg: dict) -> dict[str, int]:
    u = msg.get("usage") or {}
    cc = u.get("cache_creation") or {}
    return {
        "input": int(u.get("input_tokens") or 0),
        "cache_read": int(u.get("cache_read_input_tokens") or 0),
        "cache_write_5m": int(cc.get("ephemeral_5m_input_tokens") or 0),
        "cache_write_1h": int(cc.get("ephemeral_1h_input_tokens") or 0),
        "output": int(u.get("output_tokens") or 0),
    }


def parse(paths: list[str], est: Estimator):
    usage: list[UsageEvent] = []
    volumes: list[ToolVolume] = []
    stats = ParseStats()
    # A message id can recur across streamed partial lines: input/cache stay
    # fixed while output_tokens grows from a stub to its final value. Keep the
    # richest line per id (max total tokens) — never the first, never summed.
    by_id: dict[str, int] = {}  # message id -> index into `usage`
    tool_names: dict[str, str] = {}  # tool_use id -> tool name

    for path in paths:
        stats.files += 1
        session = os.path.splitext(os.path.basename(path))[0]
        try:
            fh = open(path, encoding="utf-8")
        except OSError:
            continue
        with fh:
            for line in fh:
                line = line.strip()
                if not line:
                    continue
                stats.lines += 1
                try:
                    rec = json.loads(line)
                except json.JSONDecodeError:
                    continue
                msg = rec.get("message")
                if not isinstance(msg, dict):
                    continue

                # First pass over content: learn tool_use ids -> names, and
                # attribute tool_result volume to the matching tool name.
                for block in msg.get("content") or []:
                    if not isinstance(block, dict):
                        continue
                    if block.get("type") == "tool_use":
                        tool_names[block.get("id", "")] = block.get("name", "?")
                    elif block.get("type") == "tool_result":
                        stats.tool_results += 1
                        tuid = block.get("tool_use_id", "")
                        name = tool_names.get(tuid)
                        if name is None:
                            name = "(unknown tool)"
                            stats.orphan_tool_results += 1
                        text = _result_text(block.get("content"))
                        volumes.append(ToolVolume(name, est.count(text)))

                # Billable usage lives on assistant messages only.
                if msg.get("role") != "assistant" or not msg.get("usage"):
                    continue
                tokens = _usage_from(msg)
                if not any(tokens.values()):
                    continue
                ts = rec.get("timestamp") or ""
                event = UsageEvent(
                    day=ts[:10] or "unknown",
                    project=_project_of(rec.get("cwd"), path),
                    session=session,
                    model=msg.get("model") or "unknown",
                    tokens=tokens,
                )
                mid = msg.get("id")
                if mid and mid in by_id:
                    stats.dup_msgs += 1
                    prev = usage[by_id[mid]]
                    if sum(tokens.values()) > sum(prev.tokens.values()):
                        usage[by_id[mid]] = event
                    continue
                if mid:
                    by_id[mid] = len(usage)
                stats.usage_msgs += 1
                usage.append(event)
    return usage, volumes, stats


# ---------------------------------------------------------------------------
# POLICY: aggregate + price. Pure functions over records. Fully table-tested.
# ---------------------------------------------------------------------------
@dataclass
class Slice:
    """A group's token totals by category, plus dollars once priced."""

    tokens: dict[str, int] = field(default_factory=lambda: {k: 0 for k in CAT_KEYS})
    dollars: dict[str, float] = field(default_factory=lambda: {k: 0.0 for k in CAT_KEYS})

    def add(self, tokens: dict[str, int]) -> None:
        for k in CAT_KEYS:
            self.tokens[k] += tokens.get(k, 0)

    @property
    def total_tokens(self) -> int:
        return sum(self.tokens.values())

    @property
    def total_dollars(self) -> float:
        return sum(self.dollars.values())


@dataclass
class Aggregates:
    by_category: Slice
    by_day: dict[str, Slice]
    by_project: dict[str, Slice]
    by_session: dict[str, Slice]
    by_model: dict[str, Slice]
    tool_volume: dict[str, int]  # tool name -> est tokens


def _grouped(usage: list[UsageEvent], key) -> dict[str, Slice]:
    out: dict[str, Slice] = defaultdict(Slice)
    for e in usage:
        out[key(e)].add(e.tokens)
    return dict(out)


def aggregate(usage: list[UsageEvent], volumes: list[ToolVolume]) -> Aggregates:
    total = Slice()
    for e in usage:
        total.add(e.tokens)
    tv: dict[str, int] = defaultdict(int)
    for v in volumes:
        tv[v.tool] += v.est_tokens
    return Aggregates(
        by_category=total,
        by_day=_grouped(usage, lambda e: e.day),
        by_project=_grouped(usage, lambda e: e.project),
        by_session=_grouped(usage, lambda e: e.session),
        by_model=_grouped(usage, lambda e: e.model),
        tool_volume=dict(tv),
    )


def load_prices(path: str) -> dict:
    with open(path, "rb") as fh:
        data = tomllib.load(fh)
    return data.get("models", {}) | {"__default__": data.get("default", {})}


def _rate(prices: dict, model: str, cat: str) -> float:
    row = prices.get(model) or prices.get("__default__") or {}
    return float(row.get(cat) or 0.0)


def price(agg: Aggregates, usage: list[UsageEvent], prices: dict) -> Aggregates:
    """Assign dollars per category using per-model rates ($/MTok)."""
    # Category / day / project / session dollars are model-blind at the group
    # level, so price them by re-walking usage with each event's own model.
    def price_group(groups: dict[str, Slice], key) -> None:
        acc: dict[tuple[str, str], float] = defaultdict(float)
        for e in usage:
            for cat in CAT_KEYS:
                acc[(key(e), cat)] += e.tokens[cat] * _rate(prices, e.model, cat)
        for gkey, cat in acc:
            groups[gkey].dollars[cat] = acc[(gkey, cat)] / 1_000_000

    # by_category dollars: sum across all events per (model,cat).
    for e in usage:
        for cat in CAT_KEYS:
            agg.by_category.dollars[cat] += e.tokens[cat] * _rate(prices, e.model, cat)
    for cat in CAT_KEYS:
        agg.by_category.dollars[cat] /= 1_000_000

    price_group(agg.by_day, lambda e: e.day)
    price_group(agg.by_project, lambda e: e.project)
    price_group(agg.by_session, lambda e: e.session)
    for m, sl in agg.by_model.items():
        for cat in CAT_KEYS:
            sl.dollars[cat] = sl.tokens[cat] * _rate(prices, m, cat) / 1_000_000
    return agg


def prices_filled(prices: dict) -> bool:
    for model, row in prices.items():
        if any(float(row.get(c) or 0.0) > 0 for c in CAT_KEYS):
            return True
    return False


# ---------------------------------------------------------------------------
# Presentation (mechanism): CLI text + HTML. No decisions here.
# ---------------------------------------------------------------------------
def _fmt_tokens(n: int) -> str:
    if n >= 1_000_000_000:
        return f"{n / 1e9:.2f}B"
    if n >= 1_000_000:
        return f"{n / 1e6:.1f}M"
    if n >= 1_000:
        return f"{n / 1e3:.1f}K"
    return str(n)


def _fmt_dollars(d: float) -> str:
    return f"${d:,.2f}"


def _top(groups: dict[str, Slice], n: int, by_dollars: bool):
    key = (lambda s: s.total_dollars) if by_dollars else (lambda s: s.total_tokens)
    return sorted(groups.items(), key=lambda kv: key(kv[1]), reverse=True)[:n]


def render_cli(agg: Aggregates, stats: ParseStats, est: Estimator, priced: bool) -> str:
    L = []
    total_tok = agg.by_category.total_tokens
    total_usd = agg.by_category.total_dollars
    L.append("=" * 64)
    L.append("  TOKEN SPEND AUDIT — the invoice nobody could show you")
    L.append("=" * 64)
    L.append(
        f"  {stats.files} sessions · {stats.usage_msgs} billed messages "
        f"· {stats.dup_msgs} dup-skipped"
    )
    headline = f"  TOTAL: {_fmt_tokens(total_tok)} tokens"
    if priced:
        headline += f"  =  {_fmt_dollars(total_usd)}"
    L.append(headline)
    L.append("")

    L.append("  BY CATEGORY")
    for cat in CAT_KEYS:
        tok = agg.by_category.tokens[cat]
        pct = 100 * tok / total_tok if total_tok else 0
        row = f"    {CAT_LABEL[cat]:<18} {_fmt_tokens(tok):>8}  {pct:5.1f}%"
        if priced:
            row += f"  {_fmt_dollars(agg.by_category.dollars[cat]):>12}"
        L.append(row)
    L.append("")

    def table(title, groups, shorten=lambda k: k):
        L.append(f"  TOP 10 {title}")
        for name, sl in _top(groups, 10, priced):
            metric = _fmt_dollars(sl.total_dollars) if priced else _fmt_tokens(sl.total_tokens)
            L.append(f"    {metric:>12}  {shorten(name)}")
        L.append("")

    table("PROJECTS", agg.by_project, lambda k: k.replace(os.path.expanduser("~"), "~"))
    table("SESSIONS", agg.by_session, lambda k: k[:8])
    table("MODELS", agg.by_model)

    L.append(f"  TOP 10 TOOLS by result volume — est. tokens [{est.method}]")
    for name, tok in sorted(agg.tool_volume.items(), key=lambda kv: kv[1], reverse=True)[:10]:
        L.append(f"    {_fmt_tokens(tok):>8}  {name}")
    L.append("")

    if priced:
        L.append("  THE ANSWER — biggest dollar slices (what must shrink for 5B->3B):")
        for cat in sorted(CAT_KEYS, key=lambda c: agg.by_category.dollars[c], reverse=True)[:3]:
            share = 100 * agg.by_category.dollars[cat] / total_usd if total_usd else 0
            L.append(f"    {CAT_LABEL[cat]:<18} {_fmt_dollars(agg.by_category.dollars[cat])}  ({share:.0f}% of spend)")
    else:
        L.append("  (!) Prices are unfilled — showing tokens only.")
        L.append("      Fill prices.toml from https://docs.claude.com/en/docs/about-claude/pricing")
    L.append("=" * 64)
    return "\n".join(L)


def render_html(agg: Aggregates, stats: ParseStats, est: Estimator, priced: bool) -> str:
    payload = {
        "priced": priced,
        "estMethod": est.method,
        "stats": vars(stats),
        "categories": [{"key": k, "label": CAT_LABEL[k]} for k in CAT_KEYS],
        "byCategory": {
            "tokens": agg.by_category.tokens,
            "dollars": agg.by_category.dollars,
        },
        "byDay": {
            d: {"tokens": s.tokens, "dollars": s.dollars, "total": s.total_tokens, "usd": s.total_dollars}
            for d, s in sorted(agg.by_day.items())
        },
        "topProjects": [
            {"name": n.replace(os.path.expanduser("~"), "~"), "tokens": s.total_tokens, "usd": s.total_dollars}
            for n, s in _top(agg.by_project, 10, priced)
        ],
        "topSessions": [
            {"name": n[:8], "tokens": s.total_tokens, "usd": s.total_dollars}
            for n, s in _top(agg.by_session, 10, priced)
        ],
        "topModels": [
            {"name": n, "tokens": s.total_tokens, "usd": s.total_dollars}
            for n, s in _top(agg.by_model, 10, priced)
        ],
        "topTools": [
            {"name": n, "tokens": t}
            for n, t in sorted(agg.tool_volume.items(), key=lambda kv: kv[1], reverse=True)[:10]
        ],
    }
    # Injected into a <script type="application/json"> raw-text element: HTML
    # entities are NOT decoded there, so we must NOT html-escape. The only
    # breakout risk is the literal sequence "</" (e.g. "</script>"); "<\/" is
    # an equivalent, safe form inside JSON.
    blob = json.dumps(payload, ensure_ascii=False).replace("</", "<\\/")
    return HTML_TEMPLATE.replace("__DATA__", blob)


# ---------------------------------------------------------------------------
# CLI
# ---------------------------------------------------------------------------
def collect_paths(root: str) -> list[str]:
    return sorted(glob.glob(os.path.join(root, "**", "*.jsonl"), recursive=True))


def main(argv: list[str] | None = None) -> int:
    ap = argparse.ArgumentParser(description="Itemize agent token spend from transcripts.")
    ap.add_argument("--root", default=os.path.expanduser("~/.claude/projects"),
                    help="transcript root (dir of **/*.jsonl)")
    ap.add_argument("--prices", default=os.path.join(os.path.dirname(__file__), "prices.toml"))
    ap.add_argument("--out", default="out", help="output dir for report.html / report.txt")
    ap.add_argument("--open", action="store_true", help="open the HTML dashboard when done")
    args = ap.parse_args(argv)

    paths = collect_paths(args.root)
    if not paths:
        print(f"no *.jsonl under {args.root}", file=sys.stderr)
        return 2

    est = Estimator()
    usage, volumes, stats = parse(paths, est)
    prices = load_prices(args.prices)
    priced = prices_filled(prices)
    agg = aggregate(usage, volumes)
    if priced:
        price(agg, usage, prices)

    os.makedirs(args.out, exist_ok=True)
    txt = render_cli(agg, stats, est, priced)
    print(txt)
    with open(os.path.join(args.out, "report.txt"), "w", encoding="utf-8") as fh:
        fh.write(txt + "\n")
    html_path = os.path.join(args.out, "report.html")
    with open(html_path, "w", encoding="utf-8") as fh:
        fh.write(render_html(agg, stats, est, priced))
    print(f"\nwrote {html_path}")
    if args.open:
        import subprocess

        opener = "open" if sys.platform == "darwin" else "xdg-open"
        subprocess.run([opener, html_path], check=False)
    return 0


HTML_TEMPLATE = r"""<!doctype html>
<html lang="en"><head><meta charset="utf-8">
<meta name="viewport" content="width=device-width,initial-scale=1">
<title>Token Spend Audit</title>
<style>
:root{--bg:#0f1115;--card:#181b22;--fg:#e7e9ee;--mut:#9aa3b2;--line:#262b35;
 --c0:#5aa9e6;--c1:#7fd1b9;--c2:#f6c667;--c3:#e88b6a;--c4:#c792ea;}
@media (prefers-color-scheme:light){:root{--bg:#f6f7f9;--card:#fff;--fg:#1a1d24;--mut:#5b6472;--line:#e5e8ee;}}
*{box-sizing:border-box}body{margin:0;background:var(--bg);color:var(--fg);
 font:15px/1.5 -apple-system,BlinkMacSystemFont,"Segoe UI",Roboto,sans-serif}
.wrap{max-width:1080px;margin:0 auto;padding:28px 20px 60px}
h1{font-size:22px;margin:0 0 2px}.sub{color:var(--mut);font-size:13px;margin-bottom:22px}
.head{display:flex;flex-wrap:wrap;gap:16px;align-items:baseline;margin-bottom:8px}
.big{font-size:40px;font-weight:700;letter-spacing:-.5px}
.big .usd{color:var(--c1)}
.grid{display:grid;grid-template-columns:1fr 1fr;gap:16px;margin-top:20px}
@media(max-width:760px){.grid{grid-template-columns:1fr}}
.card{background:var(--card);border:1px solid var(--line);border-radius:12px;padding:16px}
.card h2{font-size:13px;text-transform:uppercase;letter-spacing:.06em;color:var(--mut);margin:0 0 12px}
.full{grid-column:1/-1}
table{width:100%;border-collapse:collapse;font-size:13px}
td,th{padding:5px 6px;text-align:left;border-bottom:1px solid var(--line)}
th{color:var(--mut);font-weight:600}td.n{text-align:right;font-variant-numeric:tabular-nums;white-space:nowrap}
.legend{display:flex;flex-wrap:wrap;gap:10px;font-size:12px;color:var(--mut);margin-top:10px}
.legend span{display:inline-flex;align-items:center;gap:5px}
.sw{width:11px;height:11px;border-radius:3px;display:inline-block}
.bar{display:flex;height:26px;border-radius:6px;overflow:hidden;margin:4px 0}
.bar>div{height:100%}
.note{color:var(--mut);font-size:12px;margin-top:10px}
.warn{background:#3a2a12;color:#f6c667;border:1px solid #5a4118;padding:10px 12px;border-radius:8px;margin-bottom:18px;font-size:13px}
svg{display:block;width:100%;height:auto;overflow:visible}
.day rect:hover{opacity:.8}
tspan,text{fill:var(--mut);font-size:10px}
.answer{background:linear-gradient(90deg,rgba(90,169,230,.12),transparent);border-left:3px solid var(--c0);padding:12px 14px;border-radius:8px;margin-top:6px;font-size:14px}
.answer b{color:var(--fg)}
</style></head><body><div class="wrap">
<h1>Token Spend Audit <span style="color:var(--mut);font-weight:400">— the invoice nobody could show you</span></h1>
<div class="sub" id="sub"></div>
<div id="warn"></div>
<div class="head"><div class="big" id="big"></div></div>
<div id="answer"></div>
<div class="grid">
  <div class="card"><h2>By category</h2><div id="donut"></div><div class="legend" id="leg"></div></div>
  <div class="card"><h2>Top projects</h2><div id="tProjects"></div></div>
  <div class="card full"><h2>Spend by day</h2><div id="days"></div><div class="legend" id="leg2"></div></div>
  <div class="card"><h2>Top sessions</h2><div id="tSessions"></div></div>
  <div class="card"><h2>By model</h2><div id="tModels"></div></div>
  <div class="card full"><h2>Top tools by result volume <span style="text-transform:none;font-weight:400" id="estm"></span></h2><div id="tTools"></div></div>
</div>
<div class="note">Token counts are exact (from <code>message.usage</code>). Tool-result volume is an estimate. All data stays on this machine.</div>
</div>
<script id="data" type="application/json">__DATA__</script>
<script>
const D=JSON.parse(document.getElementById('data').textContent);
const CC=['var(--c0)','var(--c1)','var(--c2)','var(--c3)','var(--c4)'];
const CATS=D.categories, priced=D.priced;
const fmtT=n=>n>=1e9?(n/1e9).toFixed(2)+'B':n>=1e6?(n/1e6).toFixed(1)+'M':n>=1e3?(n/1e3).toFixed(1)+'K':''+n;
const fmt$=n=>'$'+n.toLocaleString(undefined,{minimumFractionDigits:2,maximumFractionDigits:2});
const metric=o=>priced?fmt$(o.usd):fmtT(o.tokens);
const catColor={}; CATS.forEach((c,i)=>catColor[c.key]=CC[i%CC.length]);

const totT=Object.values(D.byCategory.tokens).reduce((a,b)=>a+b,0);
const totU=Object.values(D.byCategory.dollars).reduce((a,b)=>a+b,0);
const dayKeys=Object.keys(D.byDay).sort();
const span=dayKeys.length?` · ${dayKeys[0]} → ${dayKeys[dayKeys.length-1]} (${dayKeys.length} days)`:'';
document.getElementById('sub').textContent=
  `${D.stats.files} sessions · ${D.stats.usage_msgs} billed messages · ${D.stats.dup_msgs} duplicates skipped${span}`;
document.getElementById('big').innerHTML=
  fmtT(totT)+' tokens'+(priced?' &nbsp;=&nbsp; <span class="usd">'+fmt$(totU)+'</span>':'');
document.getElementById('estm').textContent='('+D.estMethod+')';
if(!priced){document.getElementById('warn').innerHTML=
  '<div class="warn">Prices are unfilled — showing tokens only. Edit <code>prices.toml</code> from the Claude pricing page to see dollars.</div>';}

// donut by category
(function(){
  const R=70,r=44,cx=90,cy=90,C=2*Math.PI*R;
  let off=0,segs='';
  CATS.forEach(c=>{const v=D.byCategory.tokens[c.key];const frac=totT?v/totT:0;
    const len=frac*C;
    segs+=`<circle cx="${cx}" cy="${cy}" r="${R}" fill="none" stroke="${catColor[c.key]}" stroke-width="${R-r}" stroke-dasharray="${len} ${C-len}" stroke-dashoffset="${-off}" transform="rotate(-90 ${cx} ${cy})"><title>${c.label}: ${fmtT(v)}</title></circle>`;
    off+=len;});
  const top=CATS.map(c=>[c,D.byCategory.tokens[c.key]]).sort((a,b)=>b[1]-a[1])[0];
  document.getElementById('donut').innerHTML=
   `<svg viewBox="0 0 180 180" style="max-width:200px;margin:0 auto">${segs}
     <text x="90" y="86" text-anchor="middle" style="fill:var(--fg);font-size:15px;font-weight:700">${((top[1]/totT)*100||0).toFixed(0)}%</text>
     <text x="90" y="102" text-anchor="middle">${top[0].label}</text></svg>`;
  document.getElementById('leg').innerHTML=CATS.map(c=>{
    const v=D.byCategory.tokens[c.key],pct=totT?(100*v/totT).toFixed(0):0;
    return `<span><span class="sw" style="background:${catColor[c.key]}"></span>${c.label} ${pct}%</span>`;}).join('');
})();

// stacked bars by day
(function(){
  const days=Object.entries(D.byDay); if(!days.length){document.getElementById('days').textContent='no data';return;}
  const max=Math.max(...days.map(([,v])=>v.total));
  const W=Math.max(600,days.length*34),H=200,pad=24,bw=Math.min(26,(W-2*pad)/days.length-4);
  let bars='';
  days.forEach(([d,v],i)=>{
    const x=pad+i*((W-2*pad)/days.length);let y=H-pad;
    CATS.forEach(c=>{const t=v.tokens[c.key];const h=max?(t/max)*(H-2*pad):0;y-=h;
      bars+=`<rect x="${x}" y="${y}" width="${bw}" height="${h}" fill="${catColor[c.key]}"><title>${d} · ${c.label}: ${fmtT(t)}</title></rect>`;});
    if(i%Math.ceil(days.length/8)===0)
      bars+=`<text x="${x+bw/2}" y="${H-pad+12}" text-anchor="middle">${d.slice(5)}</text>`;
  });
  document.getElementById('days').innerHTML=`<svg class="day" viewBox="0 0 ${W} ${H}">${bars}</svg>`;
  document.getElementById('leg2').innerHTML=CATS.map(c=>
    `<span><span class="sw" style="background:${catColor[c.key]}"></span>${c.label}</span>`).join('');
})();

function tbl(id,rows,head){
  const el=document.getElementById(id);
  if(!rows.length){el.textContent='no data';return;}
  el.innerHTML='<table><tr><th>'+head+'</th><th style="text-align:right">'+(priced?'$':'tokens')+'</th></tr>'+
   rows.map(r=>`<tr><td>${r.n}</td><td class="n">${r.m}</td></tr>`).join('')+'</table>';
}
tbl('tProjects',D.topProjects.map(o=>({n:o.name,m:metric(o)})),'project');
tbl('tSessions',D.topSessions.map(o=>({n:o.name,m:metric(o)})),'session');
tbl('tModels',D.topModels.map(o=>({n:o.name,m:metric(o)})),'model');
tbl('tTools',D.topTools.map(o=>({n:o.name,m:fmtT(o.tokens)})),'tool');

// the answer strip — and the token%-vs-dollar% contrast, which IS the insight.
if(priced){
  const ranked=CATS.map(c=>({c,usd:D.byCategory.dollars[c.key],tok:D.byCategory.tokens[c.key]})).sort((a,b)=>b.usd-a.usd);
  const top=ranked[0];
  const usdPct=(top.usd/totU)*100||0, tokPct=(top.tok/totT)*100||0;
  document.getElementById('answer').innerHTML=
   `<div class="answer">The biggest dollar slice is <b>${top.c.label}</b> at <b>${fmt$(top.usd)}</b> — `+
   `<b>${usdPct.toFixed(0)}% of spend</b> (vs ${tokPct.toFixed(0)}% of <i>tokens</i> — the donut). `+
   `That gap is the whole point: writes and output cost far more per token, so the dollar bill and the token count rank differently. `+
   `Shrink the top dollar slices for 5B→3B — not whatever's biggest by count.</div>`;
}
</script></body></html>"""


if __name__ == "__main__":
    raise SystemExit(main())
