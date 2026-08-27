# bakeoffs — the hackathon archive

A frozen shelf of self-contained build-hackathon entries, grouped by
**category** and tagged by date. Each entry was built in its own fresh,
independent session that never saw the others, under a shared rulebook, then
scored and retired here.

> **This is an archive, not software.** Entries are frozen experiments — most
> are a few hours old and were never meant to survive. Nothing here is
> maintained, nothing takes issues, and you should not build on one in place.
> If an entry earned a future it was *promoted out* into its own repository;
> those are listed under **Promotion** below and are where the maintained code
> lives.

Published because the interesting part is the shape of the exercise — a
rulebook, independent sessions, a scorecard, and a written kill decision — not
the code any single entry produced. The scorecards include the entries that
lost and why.

## Why categories, not rounds

They happened on consecutive days, but the order is an accident of the
calendar, not a ranking. So they're filed by **what they are**
(`voice`, `token-opt`), with a `MM-DD` date tag to keep distinct waves apart.
A future voice bake-off is just another `voice-<date>` folder — no renumbering.

## The map

### `voice-08-07/` — voice apps, keyless idea bake-off (Aug 7 2026)
Judged on whether the *bet* lands. No API keys: browser `speechSynthesis` /
Web Speech + local Ollama. Birthplace of the house invariant —
*correctness is computed in table-tested code; the model is only ears/mouth.*

| entry | the bet | status |
|---|---|---|
| [atc](voice-08-07/atc) | ATC radio-phraseology trainer — a deterministic grader scores every element of your pattern-work readback live | built |
| [readback](voice-08-07/readback) | irreversible ops actions gated by a spoken readback (voice as a merge/confirm gate) | built |
| [partyline](voice-08-07/partyline) | monitor an agent fleet *by ear* via 2-second brevity calls | built |
| [tutor](voice-08-07/tutor) | interject's rationed-interruption engine transplanted to tutoring | built |

### `voice-08-08/` — consumer voice, would-someone-pay wave (Aug 8 2026)
Same domain, higher bar: real `OpenAI Realtime` conversational voice (browser
holds WebRTC, server mints ephemeral tokens), and the rubric is reweighted so
*would someone pay* is primary (35 pts). Built by fresh **Codex** sessions.

| entry | the bet | status |
|---|---|---|
| [story-capture](voice-08-08/story-capture) | "Keepsake" — interview grandma, get chapters; adult children pay $99/yr today for worse versions | built |
| [pitch-compressor](voice-08-08/pitch-compressor) | force founders/reps into a graded 20-second conclusion | built |
| [interview-sparring](voice-08-08/interview-sparring) | real conversational pushback for job-seeker prep | built |
| [language-drill](voice-08-08/language-drill) | "Mesa 40" — bounded phrase sets make pronunciation drilling gradeable in code, not vibes | built |

### `token-opt-08-09/` — cut agent token spend (Aug 9 2026)
Different genre: internal tooling, not consumer apps. Goal: get monthly agent
token spend from ~5B toward ~3B. Operates on your own session transcripts.
Grew out of debunking the `babel-protocol` 88%-compression headline (a
whitespace-token artifact).

| entry | the bet | status |
|---|---|---|
| [spend-audit](token-opt-08-09/spend-audit) | dollars-first breakdown of where 5B tokens actually go | built |
| [babel-bpe](token-opt-08-09/babel-bpe) | rerun the Babel bake-off in the *right unit* (real BPE) vs honest baselines; winner becomes the house wire format | built |
| [context-diet](token-opt-08-09/context-diet) | tool results are the fattest slice — mechanical hygiene rules recoup tokens with zero behavior change | built |
| [offload-router](token-opt-08-09/offload-router) | a deterministic quality gate that makes local-model offload trustworthy for mechanical sub-steps | built |
| [cache-max](token-opt-08-09/cache-max) | cache misses are silent dollars — find prefix-busting events, price what stability would save | built |

### `haskell-08-10/` — Haskell-shaped agentic semantic kernels (Aug 10 2026)

Five dependency-light POCs tested whether Haskell was essential to one narrow
agent-development responsibility rather than merely a pleasant implementation
language. All use GHC 9.10.3 with `-Wall -Werror`; the scored decision is in
[`SCORECARD.md`](haskell-08-10/SCORECARD.md).

| entry | the bet | status |
|---|---|---|
| [protocol-compiler](haskell-08-10/protocol-compiler) | project global evidence protocols into local role obligations | graduated → [parley](https://github.com/itsHabib/parley) |
| [provenance-datalog](haskell-08-10/provenance-datalog) | compute exact explanations and selective retraction with every derived fact | reference |
| [bidirectional-artifacts](haskell-08-10/bidirectional-artifacts) | preserve legal human overrides while regenerating derived workflow structure | rejected — misses Haskell floor |
| [capability-plans](haskell-08-10/capability-plans) | preflight the complete authority envelope of the same typed plan that executes | runner-up |
| [durable-workflows](haskell-08-10/durable-workflows) | replay journaled agent activities without duplicating committed effects | winner — promote semantic bet |

### `haskell-dsl-08-10/` — delightful executable languages for agent work (Aug 10 2026)

Six Haskell DSLs were judged on authoring joy as well as static and operational
leverage. The wave includes four general agentic languages, one compiler for the
portfolio's existing work-driver flow, and one entry grounded in day-job domain
rules that is **withheld from this public archive** — its scores and judging
notes are kept so the round still reads honestly. See the scored [`SCORECARD.md`](haskell-dsl-08-10/SCORECARD.md).

| entry | the language | status |
|---|---|---|
| [assurance-dsl](haskell-dsl-08-10/assurance-dsl) | exact-head evidence policy with proof-path explanations | reference |
| [delegation-dsl](haskell-dsl-08-10/delegation-dsl) | artifact contracts compiled into agent-local briefs | rejected — misses graduation floors |
| [capability-dsl](haskell-dsl-08-10/capability-dsl) | typed executable plans with complete static authority preflight | reference |
| [recovery-dsl](haskell-dsl-08-10/recovery-dsl) | durable typed activities with executable retry semantics | reference |
| *withheld — see note* | day-job domain rules as decisions, hydration manifests, and portable IR | runner-up |
| [work-driver-dsl](haskell-dsl-08-10/work-driver-dsl) | project intent compiled into safe work-driver batches | winner — promote compiler bet |

### `practical-systems-08-10/` — small languages against real systems work (Aug 10 2026)

Four deliberately infrastructure-light projects: three practical Gleam tools and
one Quint model grounded in a manufacturing workflow handoff. The shared bar is
usefulness outside the demo and a language choice that changes the design.

| entry | language | the bet | status |
|---|---|---|---|
| [mcp-contract-lab](practical-systems-08-10/mcp-contract-lab) | Gleam | turn MCP declarations and recorded exchanges into executable compatibility fixtures | queued |
| [streaming-fixture-lab](practical-systems-08-10/streaming-fixture-lab) | Gleam | capture, redact, and replay large event streams without loading them whole | queued |
| [flow-state-lab](practical-systems-08-10/flow-state-lab) | Quint | model-check a CAM-to-factory handoff and export failure traces as ordinary regression fixtures | built |
| [durable-pipeline-kernel](practical-systems-08-10/durable-pipeline-kernel) | Gleam | resume evidence-gated shipping and maintenance loops without rebuilding a workflow platform | graduated → repair-loop-kernel (not published) |

## Folder anatomy

Every category folder is self-judging and self-describing:

- `RULES.md` — that wave's house rules + the 100-point rubric (the original pack README).
- `SCORECARD.md` — a blank scoring grid built from that rubric; fill it to judge.
- `<entry>/BRIEF.md` — the exact spec the entry was built from (travels with the code).
- `<entry>/` — the entry itself: `README.md` (one command to run), `DEMO.md` (60-second hands-free walkthrough), the app, and policy-layer tests. Regenerable dirs (`.git`, `node_modules`, `.venv`, `__pycache__`) were dropped on archival.

## Judging

Point an agent (or yourself) at a single category folder — it has everything
needed: the entries, the rubric in `RULES.md`, and a blank `SCORECARD.md` to
fill. Judging is per-wave because each wave has its own rubric weights.

## Promotion

The one live path out of the archive. When an entry earns a future:

1. Copy its folder out to `~/dev/<name>`, `git init`, and it becomes a real repo.
2. Its `BRIEF.md` is the starting spec; its tests come along.
3. Flip its **status** here to `graduated → <repo url>` so the shelf records where it went.

Status legend: `built` (archived, unjudged) · `reference` (judged, kept for
parts) · `graduated → url` (promoted to its own repo).

## Provenance

The `RULES.md`/`BRIEF.md` docs originated as hackathon "packs" produced by the
`/hackathon` skill, and lived at `interject/docs/hackathon`,
`interject/docs/hackathon-codex`, and `workbench/docs/hackathon-token-opt`
before being consolidated here.
