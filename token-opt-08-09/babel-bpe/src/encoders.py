"""Encoders: one pure function per candidate wire format.

Each format exposes ``encode(payload) -> str`` and
``decode(text, shape, schema) -> canonical`` where:

  * ``shape`` is the payload type ("record" | "table" | "hierarchy" | "graph")
  * ``schema`` carries only structural hints the format legitimately relies on
    out of band (today: the attribute key list for hierarchy leaf nodes, which
    Babel keeps in its spec/legend rather than in every message). It contains
    NO data values, so ``decode`` cannot cheat by echoing the input.

Round-trip contract (asserted in tests):

    decode(encode(p), p["type"], schema(p)) == canonical(p)

``canonical`` is the single normalized structure every decoder must reproduce,
so the round-trip is comparing information content, not byte-for-byte strings.
"""

import json
import re

_INT = re.compile(r"-?\d+")


def coerce(tok):
    """A bare token is an int iff it is all digits (optional sign); else str."""
    return int(tok) if _INT.fullmatch(tok) else tok


# --- canonical normalized forms -------------------------------------------

def canon_node(n):
    return {
        "role": n["role"],
        "name": n["name"],
        "attrs": {k: v for k, v in n["attrs"]},
        "children": [canon_node(c) for c in n["children"]],
    }


def canonical(p):
    t = p["type"]
    if t == "record":
        return {"type": "record", "fields": {k: v for k, v in p["fields"]}}
    if t == "table":
        return {
            "type": "table",
            "columns": list(p["columns"]),
            "rows": [list(r) for r in p["rows"]],
        }
    if t == "hierarchy":
        return {"type": "hierarchy", "root": canon_node(p["root"])}
    if t == "graph":
        return {
            "type": "graph",
            "nodes": list(p["nodes"]),
            "edges": [list(e) for e in p["edges"]],
        }
    raise ValueError(f"unknown payload type: {t}")


def schema(p):
    """Structural hints a decoder may use — attr key list for hierarchies."""
    keys = []
    if p["type"] == "hierarchy":
        def walk(n):
            if n["attrs"] and not keys:
                keys.extend(k for k, _ in n["attrs"])
            for c in n["children"]:
                walk(c)
        walk(p["root"])
    return {"attr_keys": keys}


def applies(fmt, shape):
    """CSV/TSV only make sense for tables."""
    if fmt in ("csv", "tsv"):
        return shape == "table"
    return True


# --- helpers shared by several formats ------------------------------------

def _plain(p):
    """Naive, obvious in-memory structure an agent would hand to a serializer.

    Tables become an array of row-objects (keys repeated per row) — the common,
    verbose way — so the compact formats have something honest to beat.
    """
    t = p["type"]
    if t == "record":
        return {k: v for k, v in p["fields"]}
    if t == "table":
        return [dict(zip(p["columns"], r)) for r in p["rows"]]
    if t == "hierarchy":
        return _plain_node(p["root"])
    if t == "graph":
        return {"nodes": list(p["nodes"]), "edges": [list(e) for e in p["edges"]]}


def _plain_node(n):
    out = {"role": n["role"], "name": n["name"]}
    for k, v in n["attrs"]:
        out[k] = v
    if n["children"]:
        out["children"] = [_plain_node(c) for c in n["children"]]
    return out


def _node_from_plain(o):
    role, name = o["role"], o["name"]
    children = [_node_from_plain(c) for c in o.get("children", [])]
    attrs = {k: v for k, v in o.items() if k not in ("role", "name", "children")}
    return {"role": role, "name": name, "attrs": attrs, "children": children}


# =========================================================================
# JSON (compact and pretty share a decoder)
# =========================================================================

def json_encode(p, indent):
    sep = (",", ":") if indent is None else (",", ": ")
    return json.dumps(_plain(p), separators=sep, indent=indent, ensure_ascii=False)


def json_decode(text, shape, schema):
    o = json.loads(text)
    if shape == "record":
        return {"type": "record", "fields": o}
    if shape == "table":
        cols = list(o[0].keys())
        return {"type": "table", "columns": cols, "rows": [[r[c] for c in cols] for r in o]}
    if shape == "hierarchy":
        return {"type": "hierarchy", "root": _node_from_plain(o)}
    return {"type": "graph", "nodes": o["nodes"], "edges": o["edges"]}


# =========================================================================
# YAML (hand-rolled block subset — no third-party dep)
# =========================================================================

def _yaml_scalar(v):
    return str(v)


def yaml_encode(p):
    t = p["type"]
    if t == "record":
        return "".join(f"{k}: {_yaml_scalar(v)}\n" for k, v in p["fields"])
    if t == "table":
        out = []
        for r in p["rows"]:
            cells = list(zip(p["columns"], r))
            out.append(f"- {cells[0][0]}: {_yaml_scalar(cells[0][1])}\n")
            out += [f"  {k}: {_yaml_scalar(v)}\n" for k, v in cells[1:]]
        return "".join(out)
    if t == "hierarchy":
        return _yaml_hnode(p["root"], 0)
    if t == "graph":
        out = ["nodes:\n"]
        out += [f"- {n}\n" for n in p["nodes"]]
        out.append("edges:\n")
        out += [f"- {s} {d}\n" for s, d in p["edges"]]
        return "".join(out)


def _yaml_hnode(n, depth):
    """Org-chart YAML: node name is the key, so the tree is pure nested maps.
    Child maps and scalar attrs are told apart by value type on decode."""
    pad, pad2 = "  " * depth, "  " * (depth + 1)
    out = [f"{pad}{n['name']}:\n", f"{pad2}role: {n['role']}\n"]
    out += [f"{pad2}{k}: {_yaml_scalar(v)}\n" for k, v in n["attrs"]]
    out += [_yaml_hnode(c, depth + 1) for c in n["children"]]
    return "".join(out)


def _node_from_named(name, m):
    attrs = {k: v for k, v in m.items() if k != "role" and not isinstance(v, dict)}
    children = [_node_from_named(k, v) for k, v in m.items() if isinstance(v, dict)]
    return {"role": m["role"], "name": name, "attrs": attrs, "children": children}


def _yaml_load_map(pairs, i, indent):
    """Load a block map of scalar and nested-map entries; return (dict, i)."""
    d = {}
    while i < len(pairs) and pairs[i][0] == indent:
        _, s = pairs[i]
        key, sep, val = s.partition(": ")
        if sep:
            d[key] = coerce(val)
            i += 1
            continue
        key = s[:-1]  # 'key:' introduces a nested map
        if i + 1 < len(pairs) and pairs[i + 1][0] > indent:
            d[key], i = _yaml_load_map(pairs, i + 1, indent + 2)
        else:
            d[key], i = {}, i + 1
    return d, i


def yaml_decode(text, shape, schema):
    lines = [ln for ln in text.split("\n") if ln.strip()]
    if shape == "record":
        fields = {}
        for ln in lines:
            k, v = ln.split(": ", 1)
            fields[k] = coerce(v)
        return {"type": "record", "fields": fields}
    if shape == "table":
        rows, cur = [], None
        for ln in lines:
            body = ln.lstrip()
            if ln.startswith("- "):
                if cur is not None:
                    rows.append(cur)
                cur = {}
                body = body[2:]
            k, v = body.split(": ", 1)
            cur[k] = coerce(v)
        rows.append(cur)
        cols = list(rows[0].keys())
        return {"type": "table", "columns": cols, "rows": [[r[c] for c in cols] for r in rows]}
    if shape == "hierarchy":
        pairs = [(len(ln) - len(ln.lstrip()), ln.rstrip()) for ln in lines]
        pairs = [(ind, s.strip()) for ind, s in pairs]
        top, _ = _yaml_load_map(pairs, 0, 0)
        name = next(iter(top))
        return {"type": "hierarchy", "root": _node_from_named(name, top[name])}
    # graph
    nodes, edges, mode = [], [], None
    for ln in lines:
        s = ln.strip()
        if s == "nodes:":
            mode = "n"
        elif s == "edges:":
            mode = "e"
        elif mode == "n":
            nodes.append(s[2:])
        else:
            a, b = s[2:].split(" ")
            edges.append([a, b])
    return {"type": "graph", "nodes": nodes, "edges": edges}


# =========================================================================
# CSV / TSV (tables only) — stdlib-style, hand-emitted to avoid quoting drift
# =========================================================================

def _delim_encode(p, sep):
    out = [sep.join(p["columns"])]
    out += [sep.join(str(c) for c in r) for r in p["rows"]]
    return "\n".join(out) + "\n"


def _delim_decode(text, sep):
    lines = [ln for ln in text.split("\n") if ln]
    cols = lines[0].split(sep)
    rows = [[coerce(c) for c in ln.split(sep)] for ln in lines[1:]]
    return {"type": "table", "columns": cols, "rows": rows}


# =========================================================================
# TOON — pinned subset, see README ("TOON r1, this repo's pinned revision")
# =========================================================================

def toon_encode(p):
    t = p["type"]
    if t == "record":
        return "".join(f"{k}: {v}\n" for k, v in p["fields"])
    if t == "table":
        cols = ",".join(p["columns"])
        out = [f"rows[{len(p['rows'])}]{{{cols}}}:\n"]
        out += ["  " + ",".join(str(c) for c in r) + "\n" for r in p["rows"]]
        return "".join(out)
    if t == "hierarchy":
        return _toon_node(p["root"], 0)
    if t == "graph":
        out = [f"nodes[{len(p['nodes'])}]: " + ",".join(p["nodes"]) + "\n"]
        out.append(f"edges[{len(p['edges'])}]{{from,to}}:\n")
        out += ["  " + f"{s},{d}" + "\n" for s, d in p["edges"]]
        return "".join(out)


def _toon_node(n, depth):
    pad = "  " * depth
    parts = [n["role"], n["name"]] + [str(v) for _, v in n["attrs"]]
    line = pad + ",".join(parts)
    if n["children"]:
        out = [line + f" [{len(n['children'])}]:\n"]
        for c in n["children"]:
            out.append(_toon_node(c, depth + 1))
        return "".join(out)
    return line + "\n"


def toon_decode(text, shape, schema):
    lines = [ln for ln in text.split("\n") if ln.strip()]
    if shape == "record":
        fields = {}
        for ln in lines:
            k, v = ln.split(": ", 1)
            fields[k] = coerce(v)
        return {"type": "record", "fields": fields}
    if shape == "table":
        header = lines[0]
        cols = header[header.index("{") + 1: header.index("}")].split(",")
        rows = [[coerce(c) for c in ln.strip().split(",")] for ln in lines[1:]]
        return {"type": "table", "columns": cols, "rows": rows}
    if shape == "hierarchy":
        pairs = [(len(ln) - len(ln.lstrip()), ln.strip()) for ln in lines]
        root, _ = _toon_parse_node(pairs, 0, schema["attr_keys"])
        return {"type": "hierarchy", "root": root}
    # graph
    nline = lines[0]
    nodes = nline.split(": ", 1)[1].split(",")
    edges = [[*ln.strip().split(",")] for ln in lines[2:]]
    return {"type": "graph", "nodes": nodes, "edges": edges}


def _toon_parse_node(pairs, i, attr_keys):
    ind, s = pairs[i]
    body = s
    nchild = 0
    if body.endswith("]:"):
        head, _, tail = body.rpartition(" [")
        nchild = int(tail[:-2])
        body = head
    parts = body.split(",")
    role, name = parts[0], parts[1]
    attrs = {attr_keys[j]: coerce(parts[2 + j]) for j in range(len(parts) - 2)}
    node = {"role": role, "name": name, "attrs": attrs, "children": []}
    i2 = i + 1
    for _ in range(nchild):
        child, i2 = _toon_parse_node(pairs, i2, attr_keys)
        node["children"].append(child)
    return node, i2


# =========================================================================
# Babel v3 (underscore-packed) and babel-nl (the BPE-tuned mutation)
# =========================================================================
# Babel packs multi-word values with '_' and separates fields with spaces;
# that underscore packing is exactly what fragments whole words for the BPE
# tokenizer. babel-nl keeps identical structure but preserves real spaces and
# separates record/table fields with newlines/tabs instead — so the tokenizer
# sees whole words. Same information, same grammar family, different packing.

def _sp(v, pack):
    return str(v).replace(" ", "_") if pack else str(v)


def _unsp(v, pack):
    return v.replace("_", " ") if pack else v


def _babel_encode(p, pack):
    t = p["type"]
    if t == "record":
        if pack:
            body = " ".join(f"{k}:{_sp(v, True)}" for k, v in p["fields"])
            return "R\n" + body + "\n"
        return "R\n" + "".join(f"{k}:{v}\n" for k, v in p["fields"])
    if t == "table":
        if pack:
            head = "[" + " ".join(p["columns"]) + "]"
            out = ["T " + head]
            out += [" ".join(_sp(c, True) for c in r) for r in p["rows"]]
            return "\n".join(out) + "\n"
        head = "\t".join(p["columns"])
        out = ["T", head]
        out += ["\t".join(str(c) for c in r) for r in p["rows"]]
        return "\n".join(out) + "\n"
    if t == "hierarchy":
        out = ["H"]
        _babel_node(p["root"], 0, out, pack)
        return "\n".join(out) + "\n"
    if t == "graph":
        adj = {}
        for s, d in p["edges"]:
            adj.setdefault(s, []).append(d)
        out = ["G"]
        for n in p["nodes"]:
            if n in adj:
                out.append(f"{n}>" + ",".join(adj[n]))
        return "\n".join(out) + "\n"


def _babel_node(n, depth, out, pack):
    prefix = "/" * depth
    parts = [n["role"], n["name"]] + [str(v) for _, v in n["attrs"]]
    out.append(prefix + ":".join(_sp(x, pack) for x in parts))
    for c in n["children"]:
        _babel_node(c, depth + 1, out, pack)


def _babel_decode(text, shape, schema, pack):
    lines = [ln for ln in text.split("\n") if ln.strip()]
    header, body = lines[0], lines[1:]
    if shape == "record":
        fields = {}
        if pack:
            for tok in body[0].split(" "):
                k, v = tok.split(":", 1)
                fields[k] = coerce(_unsp(v, True))
        else:
            for ln in body:
                k, v = ln.split(":", 1)
                fields[k] = coerce(v)
        return {"type": "record", "fields": fields}
    if shape == "table":
        if pack:
            cols = header[header.index("[") + 1: header.index("]")].split(" ")
            rows = [[coerce(_unsp(c, True)) for c in ln.split(" ")] for ln in body]
        else:
            cols = body[0].split("\t")
            rows = [[coerce(c) for c in ln.split("\t")] for ln in body[1:]]
        return {"type": "table", "columns": cols, "rows": rows}
    if shape == "hierarchy":
        keys = schema["attr_keys"]
        stack = []
        root = None
        for ln in body:
            depth = len(ln) - len(ln.lstrip("/"))
            parts = ln.lstrip("/").split(":")
            role = _unsp(parts[0], pack)
            name = _unsp(parts[1], pack)
            attrs = {keys[j]: coerce(_unsp(parts[2 + j], pack)) for j in range(len(parts) - 2)}
            nd = {"role": role, "name": name, "attrs": attrs, "children": []}
            if depth == 0:
                root = nd
            else:
                stack[depth - 1]["children"].append(nd)
            if len(stack) > depth:
                stack[depth:] = [nd]
            else:
                stack.append(nd)
        return {"type": "hierarchy", "root": root}
    # graph
    nodes, edges = [], []
    seen = set()
    for ln in body:
        s, _, tgts = ln.partition(">")
        if s not in seen:
            nodes.append(s)
            seen.add(s)
        for d in tgts.split(","):
            edges.append([s, d])
            if d not in seen:
                # targets may be leaves; record order of first appearance later
                pass
    # recover full node list from schema order is lost for babel-graph; rebuild
    # node set preserving edge-source order then append pure sinks in edge order
    allnodes = []
    seen = set()
    for s, d in edges:
        for x in (s, d):
            if x not in seen:
                allnodes.append(x)
                seen.add(x)
    return {"type": "graph", "nodes": allnodes, "edges": edges}


# =========================================================================
# registry
# =========================================================================

FORMATS = {
    "json-compact": (lambda p: json_encode(p, None), json_decode),
    "json-pretty": (lambda p: json_encode(p, 2), json_decode),
    "yaml": (yaml_encode, yaml_decode),
    "tsv": (lambda p: _delim_encode(p, "\t"), lambda t, s, sc: _delim_decode(t, "\t")),
    "csv": (lambda p: _delim_encode(p, ","), lambda t, s, sc: _delim_decode(t, ",")),
    "toon": (toon_encode, toon_decode),
    "babel": (lambda p: _babel_encode(p, True), lambda t, s, sc: _babel_decode(t, s, sc, True)),
    "babel-nl": (lambda p: _babel_encode(p, False), lambda t, s, sc: _babel_decode(t, s, sc, False)),
}

FORMAT_ORDER = [
    "json-compact", "json-pretty", "yaml", "csv", "tsv", "toon", "babel", "babel-nl",
]


def _norm(c):
    c = json.loads(json.dumps(c))  # deep copy
    if c["type"] == "graph":
        # node/edge ordering is not semantically meaningful for a graph
        c["nodes"] = sorted(c["nodes"])
        c["edges"] = sorted(c["edges"])
    return c


def roundtrip_equal(a, b):
    return _norm(a) == _norm(b)


def encode(fmt, payload):
    return FORMATS[fmt][0](payload)


def decode(fmt, text, shape, sch):
    return FORMATS[fmt][1](text, shape, sch)
