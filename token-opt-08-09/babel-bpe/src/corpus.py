"""The bake-off corpus: ~10 structured payloads spanning the shapes agents
actually exchange.

Every payload is a plain in-memory value built from four structural types:

    record     an ordered list of (key, value) scalar fields
    table      {columns: [...], rows: [[...], ...]}
    hierarchy  a tree of {role, name, attrs, children} nodes
    graph      {nodes: [...], edges: [(src, dst), ...]}  (directed)

The six shape *families* the brief names map onto these four types:
"flat records" -> record, "multi-row tables" -> table, "org-chart
hierarchies" -> hierarchy, "graphs" -> graph, "constraint sets" ->
table (id/rule rows), "verification reports" -> table (check/status rows).

Scalar values are only `str` or `int`. That is deliberate: on parse, a token
that matches ^-?\\d+$ is an int, everything else is a string. No booleans, no
bare-integer strings — so every format round-trips (encode -> parse -> equal)
without a schema or a type-guessing heuristic that could disagree between
formats. All data values are freshly synthesized: nothing here is memorized
text from the 2025 experiment.
"""


def record(fields):
    return {"type": "record", "fields": list(fields)}


def table(columns, rows):
    return {"type": "table", "columns": list(columns), "rows": [list(r) for r in rows]}


def node(role, name, attrs=None, children=None):
    return {
        "role": role,
        "name": name,
        "attrs": list(attrs or []),
        "children": list(children or []),
    }


def hierarchy(root):
    return {"type": "hierarchy", "root": root}


def graph(nodes, edges):
    return {"type": "graph", "nodes": list(nodes), "edges": [tuple(e) for e in edges]}


# --- record payloads -------------------------------------------------------

DEPLOY_RECORD = record([
    ("service", "billing-gateway"),
    ("region", "eu-west-2"),
    ("version", "4.7.1"),
    ("replicas", 6),
    ("commit", "d41ab9c"),
    ("rollout", "2026-05-14"),
])

INCIDENT_RECORD = record([
    ("incident", "INC-90214"),
    ("severity", "sev2"),
    ("owner", "Priya Raman"),
    ("component", "search-indexer"),
    ("open_minutes", 47),
    ("detected", "2026-06-02"),
])


# --- table payloads --------------------------------------------------------

FLIGHT_TABLE = table(
    ["flight", "origin", "destination", "departure"],
    [
        ["NZ 118", "AKL", "SYD", "07:20"],
        ["QF 442", "MEL", "BNE", "09:55"],
        ["EK 409", "SIN", "DXB", "12:10"],
        ["LX 038", "ZRH", "NRT", "13:40"],
        ["AC 856", "YYZ", "LHR", "18:05"],
    ],
)

METRIC_TABLE = table(
    ["service", "rps", "p99_ms", "error_pct"],
    [
        ["checkout", 1840, 210, 3],
        ["catalog", 5120, 95, 1],
        ["recommend", 760, 430, 12],
        ["payments", 990, 180, 2],
    ],
)

VERIFY_REPORT = table(
    ["check", "status", "detail"],
    [
        ["schema-lint", "pass", "0 findings"],
        ["type-check", "pass", "0 errors"],
        ["unit-tests", "fail", "2 of 314 failed"],
        ["race-detector", "pass", "clean"],
        ["boundary-law", "warn", "1 cross-import flagged"],
    ],
)

CONSTRAINT_SET = table(
    ["id", "rule"],
    [
        ["C1", "Ada seats in an even chair"],
        ["C2", "Bo sits immediately left of Cal"],
        ["C3", "Dee avoids chair 4"],
        ["C4", "Cal outranks Ada by chair"],
    ],
)


# --- hierarchy payloads ----------------------------------------------------

ORG_CHART = hierarchy(node(
    "CEO", "Mara Ostrowski", children=[
        node("VP Engineering", "Theo Vance", children=[
            node("Lead Platform", "Suki Ito", [("headcount", 9)]),
            node("Lead Data", "Owen Brandt", [("headcount", 5)]),
            node("Lead Mobile", "Nadia Rex", [("headcount", 7)]),
        ]),
        node("VP Product", "Lucia Menon", children=[
            node("PM Growth", "Ravi Anand", [("headcount", 4)]),
            node("PM Trust", "Elin Sato", [("headcount", 6)]),
        ]),
        node("VP Revenue", "Dario Fuchs", children=[
            node("Dir Enterprise", "Bea Cole", [("headcount", 11)]),
            node("Dir Partner", "Ken Awad", [("headcount", 8)]),
        ]),
    ],
))

TAXONOMY = hierarchy(node(
    "root", "catalog", children=[
        node("category", "hardware", children=[
            node("item", "router", [("sku", 3301)]),
            node("item", "switch", [("sku", 3302)]),
        ]),
        node("category", "software", children=[
            node("item", "firewall", [("sku", 5210)]),
            node("item", "monitor", [("sku", 5211)]),
        ]),
    ],
))


# --- graph payloads --------------------------------------------------------

NETWORK_GRAPH = graph(
    ["web", "api", "auth", "cache", "db", "queue"],
    [
        ("web", "api"),
        ("web", "auth"),
        ("api", "cache"),
        ("api", "db"),
        ("auth", "db"),
        ("cache", "db"),
        ("api", "queue"),
        ("queue", "db"),
    ],
)

SERVICE_GRAPH = graph(
    ["ingest", "parse", "score", "store", "notify"],
    [
        ("ingest", "parse"),
        ("parse", "score"),
        ("score", "store"),
        ("score", "notify"),
        ("store", "notify"),
    ],
)


CORPUS = [
    ("deploy_record", DEPLOY_RECORD),
    ("incident_record", INCIDENT_RECORD),
    ("flight_table", FLIGHT_TABLE),
    ("metric_table", METRIC_TABLE),
    ("verify_report", VERIFY_REPORT),
    ("constraint_set", CONSTRAINT_SET),
    ("org_chart", ORG_CHART),
    ("taxonomy", TAXONOMY),
    ("network_graph", NETWORK_GRAPH),
    ("service_graph", SERVICE_GRAPH),
]
