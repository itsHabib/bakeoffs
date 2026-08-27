# hack3-offload-router

**Displace frontier tokens onto a $0 local model — behind a gate that can't lie.**

Offloading mechanical sub-steps (narrow this file list, extract fields from
noisy output, classify these log lines) to a local 7B model saves frontier
tokens. The reason nobody does it at scale isn't cost — it's *trust*: a 7B model
quietly inventing a value is worse than paying for the frontier call. So this
repo makes trust **mechanical**. Every local result is checked by a
deterministic verifier before it's allowed out, and a ledger proves how many
frontier tokens were displaced.

> The model proposes. A deterministic verifier disposes. Offload without
> verification is just moving errors somewhere cheaper.

## One command

```bash
make venv   # once: creates .venv, installs tiktoken
make demo   # the whole thing, hands-free, deterministic (no ollama needed)
```

`make demo` runs all three task classes against bundled fixtures: two clean runs
that pass the gate and one adversarial run per class where a planted-bad model
output is caught and **rejected with its exact reason** — then prints the
displacement ledger. It replays committed cassettes, so it produces the same
story every time even with ollama down. `make demo-live` runs the same fixtures
against real `qwen2.5:7b`.

## The three task classes

Each class is a triple: a prompt template for `qwen2.5:7b`, a **deterministic
verifier**, and golden fixtures. Each shape was chosen because its answer is
checkable *by construction* — the verifier recomputes or bounds the answer, it
never asks a model to grade a model.

| class | input → output | verifier (the law) |
|---|---|---|
| `narrow` | file paths + criterion → matching subset | output ⊆ input **and** exactly matches the criterion's spot-check regex |
| `extract` | noisy output + schema → structured JSON | schema-valid **and** every value appears **verbatim** in the input (no hallucinated values) |
| `classify` | log lines + fixed label set → line→label | labels ∈ set, all lines covered, **and** planted ground-truth lines scored exactly |

## The gate

```
render prompt → call local model → parse + verify
  ├─ verify ok  → PASS: write output to stdout, ledger it, exit 0
  └─ verify fail → retry ONCE with the failure reason appended
       ├─ ok   → PASS (attempts=2)
       └─ fail → REJECT: reason to stderr, nothing on stdout, exit 2
```

It never silently passes unverified output, and it never calls a frontier model
— this machine is keyless; **`REJECT` is the fallback**, not a bigger model.

## CLI

```bash
# per-call contract:
hack3-offload <narrow|extract|classify> [--replay FILE] [--ledger FILE] < in.json > out.json
hack3-offload report [--ledger FILE]     # displaced tokens + rejection rate per class
hack3-offload demo   [--live]            # the canned walkthrough

# default model source is live ollama; --replay FILE uses a cassette instead.
```

Exit codes: `0` PASS · `2` REJECT · `3` usage/input error · `4` ollama unreachable.

## The displacement ledger

One append-only JSONL record per call (task class, input/output bytes,
tiktoken-estimated tokens a frontier model would have spent, local wall time,
verdict, attempts, reason). `report` totals **displaced tokens** (frontier
tokens a verified local call replaced) and **rejection rate** per class. That
number is the product: it turns "use a cheaper model sometimes" into a policy
you can point at.

## Would someone pay?

**Buyer:** anyone running frontier-priced agents at scale — the same teams
already buying [Martian](https://withmartian.com/),
[Not Diamond](https://www.notdiamond.ai/), OpenRouter's routing, and
LiteLLM/Portkey gateways. Money already moves here: LLM **routing** is a funded
category whose whole pitch is "send the cheap calls to the cheap model." What
those routers *don't* sell is a per-call **correctness gate** on the cheap
model's output — they route on predicted difficulty and hope. This repo is the
missing half: route the mechanical calls down **and** verify each result, so a
buyer can set "offload class X" as policy instead of a gamble. The ledger is the
CFO-facing artifact that justifies the switch.

## Why not just flip the model picker to a cheaper tier? (token-necessity)

That's the obvious alternative, and it loses on the thing that matters: an
un-verified cheaper tier **silently returns wrong answers**. Flipping the picker
saves the same tokens right up until the 7B model invents a digest, mislabels a
SQL-injection line as STYLE, or drops a file from a subset — and nothing tells
you. The demo shows all three being caught. The value here isn't the cheaper
model (ollama is free and already installed); it's the **verifier + ledger** that
make the cheaper model *safe to depend on* and *measurable*. Flipping a picker
gives you neither the catch nor the number.

## Deterministic share

All grading is real code with table tests — the model is phrasing, never the
grader (house invariant). See [`tests/test_verifiers.py`](tests/test_verifiers.py)
(every verifier: happy path, structural violations, and the adversarial shape)
and [`tests/test_gate.py`](tests/test_gate.py) (retry-once-then-reject, ledger
math). Run `make test`.

## Prices & tokenizer honesty

- **Tokenizer:** token counts use tiktoken `o200k_base`, which is a **proxy** for
  the real Claude tokenizer, not the tokenizer itself. Every ledger figure is
  labeled an estimate. Without tiktoken the code falls back to a chars/4
  heuristic so tests run keyless.
- **Prices:** the one editable `$/MTok` table lives in
  [`hack3_offload/prices.py`](hack3_offload/prices.py), values left `TODO` for
  you to fill from <https://docs.claude.com/en/docs/about-claude/pricing>
  (separate columns for fresh input, cache write, cache read, output). Until
  filled, `report` prints tokens only — no invented dollar figures.

## Layout

```
hack3_offload/
  verifiers.py   the law — pure, deterministic, table-tested
  gate.py        policy: verify-or-retry-once-then-reject
  tasks.py       prompt templates (phrasing only)
  model.py       mechanism: ollama (live) | cassette (replay). never frontier.
  ledger.py      the displacement ledger + report
  tokens.py      tiktoken o200k_base proxy (heuristic fallback)
  prices.py      the ONE editable $/MTok table (TODO)
  demo.py        the canned, self-checking walkthrough
fixtures/        per-class inputs (synthesized — no private data)
cassettes/       committed model outputs (clean + planted-bad) for determinism
tests/           table tests on the correctness path
```

Keyless, offline, no frameworks, no build step, no daemon. Go/Node were options;
Python won because the required tokenizer (`tiktoken`) is Python — one language,
no cross-process shelling.
