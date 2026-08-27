"""In-context spec / legend cost per format.

A format only costs its spec once per conversation; after that every payload
rides for free. So the fair unit is *amortized*: spec_tokens + N * payload
tokens. JSON / YAML / CSV / TSV are assumed universally known to the model and
carry no spec (0). TOON, Babel, and babel-nl must teach themselves in context,
so their legend tokens are measured with the same tokenizer as the payloads.

These strings are the literal legend an agent would paste above the payloads.
Keep them minimal — the whole point of the 2025 finding is that spec bloat is
what sinks an invented format at low N.
"""

# JSON / YAML / CSV / TSV: no legend, the model already knows them.
NONE = ""

# TOON r1 (this repo's pinned revision) — compact legend.
TOON = (
    "TOON r1: indentation = nesting. `key: value` scalar. "
    "`key[N]{f1,f2}:` then N indented comma rows = table. "
    "`key[N]: a,b,c` = inline list. Hierarchy node line = "
    "`role,name[,attr...]` and ` [K]:` marks K indented children."
)

# Babel v3 legend, verbatim from the 2025 spec (reference/README.md), trimmed
# to the grammar an agent actually needs to read these six shapes.
BABEL = (
    "Babel v3. Line1=HEADER, lines2+=payload. "
    "Headers: R=record T=table H=hierarchy G=graph. "
    "`:`=key:value. Fields space-separated; `_`=space inside a value. "
    "R: one line of key:value pairs. "
    "T `[f1 f2]` schema then positional space-separated rows. "
    "H: `/` per depth, node=`role:name[:attr]`, child after parent. "
    "G: `src>t1,t2` adjacency, one line per source."
)

# babel-nl: same grammar family, but the packing note changes — words stay
# whole (real spaces), fields are newline/tab separated instead of space+`_`.
BABEL_NL = (
    "babel-nl. Line1=HEADER, lines2+=payload. "
    "Headers: R=record T=table H=hierarchy G=graph. "
    "R: one `key:value` per line (values keep spaces). "
    "T: tab-separated header line then tab-separated rows. "
    "H: `/` per depth, node=`role:name[:attr]` (values keep spaces). "
    "G: `src>t1,t2` adjacency, one line per source."
)

SPEC = {
    "json-compact": NONE,
    "json-pretty": NONE,
    "yaml": NONE,
    "csv": NONE,
    "tsv": NONE,
    "toon": TOON,
    "babel": BABEL,
    "babel-nl": BABEL_NL,
}
