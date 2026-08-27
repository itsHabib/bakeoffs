# Entry 4: hack3-offload-router — displace frontier tokens behind a gate that can't lie

Read ~/dev/workbench/docs/hackathon-token-opt/README.md first — house rules
and judging apply verbatim. Repo: `~/dev/hack3-offload-router`. You never
see the other 4 entries.

## The bet

A meaningful share of frontier tokens goes to mechanical sub-steps — narrow
this file list, extract structure from noisy output, classify these log
lines — that a 7B local model does for $0. The operator already has an
`/offload` skill built on this instinct; what it lacks is the thing that
makes offload *trustworthy at scale*: a deterministic quality gate that
verifies each local result and a ledger that proves how many frontier tokens
were displaced. The gate is the product — offload without verification is
just moving errors somewhere cheaper. Buyer: anyone running frontier-priced
agents who'd route 20% of calls to a $0 model if they could trust the
results mechanically.

## What to build

A CLI (`hack3-offload <task-class> < input.json > output.json`) with three
task classes, each defined by a triple: prompt template for `qwen2.5:7b`, a
**deterministic verifier**, and a golden-test suite.

1. **Task classes** (pick shapes with checkable answers):
   - `narrow`: given a JSON list of file paths + a criterion, return the
     matching subset. Verifier: output ⊆ input, valid JSON, every excluded
     item fails an operator-visible spot-check regex where the criterion is
     syntactic (TODO: pick 2-3 syntactic criteria for the canned demo).
   - `extract`: given noisy command/build output + a target schema, return
     structured JSON. Verifier: schema-validates AND every extracted value
     appears verbatim as a substring of the input (no hallucinated values).
   - `classify`: given log lines + a fixed label set, return line→label.
     Verifier: labels ∈ label set, all lines covered, plus a seeded
     subset of lines with known labels mixed in — planted ground truth
     scored exactly.
2. **Gate behavior:** verifier failure → retry once with the failure reason
   appended; second failure → exit nonzero with `REJECT` + reason. Never
   silently pass through unverified model output; never call a frontier
   model (keyless).
3. **Displacement ledger:** per call, append a JSONL record: task class,
   input/output size, tiktoken-estimated tokens a frontier model would have
   consumed for the same call (labeled as an estimate), local wall time,
   gate verdict. A `report` subcommand totals displaced tokens and rejection
   rates per class.

The moment the demo turns on: the gate catches the local model inventing a
value that isn't in the input, rejects it with the exact reason, and the
ledger still shows net displacement was worth it.

## What NOT to build

- No daemon, no server, no queue — a CLI invoked per call.
- No frontier fallback path (keyless machine; `REJECT` is the fallback).
- No prompt-optimization loop, no fine-tuning, no model management.
- No integration with the existing `/offload` skill or `local` CLI — steal
  the mechanism idea, not the architecture; this repo stands alone.
- No more than three task classes. Three, tested, beats six sketched.

## Canned demo (required)

`make demo` runs all three classes against bundled fixture inputs: two clean
runs that pass the gate, plus one adversarial fixture per class where a
planted-bad model output (replayed from a committed cassette, so the demo is
deterministic even if ollama drifts) gets caught and rejected with its
reason printed. Then `report` prints the displacement ledger. Live-model
mode behind a flag when ollama is up. Tests: table-tests on every verifier,
including the adversarial cases.

## The 60-second demo story

"Offloading to a local model saves frontier tokens but nobody trusts a 7B
model's output — so I made trust mechanical. Three task classes, each with
a verifier that can't be sweet-talked: subset-only, substring-only,
label-set-only. Watch it catch the local model hallucinating a value...
rejected, with the reason. Now the ledger: across the demo corpus, [X]
frontier tokens displaced, [Y]% rejection rate. That's the number that turns
'use a cheaper model sometimes' from vibes into policy."
