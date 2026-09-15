# interject — kickoff / handoff

**Status: working POC, verified end-to-end, uncommitted.** Built in one session
(2026-08-07) from a cold start. This doc is the whole context a fresh agent
needs; no prior conversation required.

## What this is

A voice listener that is allowed to interrupt the speaker **3 times per
10-minute window** and must decide — while they are still talking — whether
what it has to say is worth spending one. Rationed speech as forced
calibration: every interruption costs the agent something, so restraint is
structural, not prompted.

Fully local: Go server, vanilla-JS browser UI, `qwen2.5:7b` on Ollama for
classification, browser Web Speech for mic input, `speechSynthesis` for voice
out. **No OpenAI key, no cloud, no Realtime API, no build step.**

This was idea 1 of an eight-idea voice-model brainstorm. The underlying
primitive (shared with several of the other ideas): *deciding WHETHER to speak
in ~200ms when deciding WHAT to say takes seconds.*

## Current state — all verified

- `go build` / `vet` / `gofmt` clean; policy layer table-tested green
  (`internal/budget`, 19 test functions).
- End-to-end verified twice via Playwright against the live model: rehearsal
  monologue reproduces the intended arc deterministically —
  silent through 12 of 14 utterances, spends turn 1 on a real contradiction,
  refuses the same contradiction 5s later (surcharged bar), spends turn 2 only
  on a more expensive `error`.
- Verdict/calibration loop verified over HTTP: `wasted` verdict raised the bar
  (trim 0→7), `worth` lowered it (7→3), bogus verdict → 400.
- UI screenshot in `docs/demo.png`.

**Not committed.** `git` exits 69 on this machine (Xcode license never
accepted). Operator must run `sudo xcodebuild -license` — needs their
password, an agent cannot do it. Until then: `go build -buildvcs=false`
(README documents this; drop the flag once git works). First commit of the
whole tree is the obvious first action once git is alive.

## Run it

```sh
ollama serve                    # operator prefers ollama STOPPED when idle — stop it after
cd voice-08-07/interject && go run -buildvcs=false .
```

Open http://127.0.0.1:8722 → **▶ Run the rehearsal** (~55s, no mic needed).
Mic path needs Chrome/Safari. Typing into the transcript box also works.

## Architecture (the part that must not be undone)

```
main.go              flags, routes, embedded UI
internal/budget      POLICY — the ration. pure Go, table-tested, zero model calls
internal/listen      mechanism — one classification, any OpenAI-shaped endpoint
internal/session     composition — transcript in, decisions out
internal/stream      mechanism — SSE fan-out
web/                 vanilla HTML/CSS/JS
prompts/classify.md  the classifier prompt — reloaded on EVERY call, tune it live
```

**The model is a sensor, not the agent.** It answers only "what kind of thing
was just said" (`error` / `contradiction` / `blindspot` / `circling` / `none`
+ confidence + a ≤14-word line). ALL scarcity — pressure envelope, pacing
threshold, recency surcharge, floor, calibration trim — is deterministic code
in `internal/budget`.

**Why continuous scoring:** classification runs *while the person talks*
(1.3–2.4s warm on qwen via `keep_alive: 30m`), so an opinion is always already
formed when a pause opens; the gap only reads it. Never move classification
into the gap.

## Hard-won lessons (do not re-learn these)

1. **Small models cannot do economic self-restraint.** Asking qwen "would you
   spend one of your only 3 interruptions on this?" collapsed to always-no and
   flattened calibration (real security bug scored 30, pure filler 20). The
   fix was the sensor/policy split above. Same warning applies to any
   bid/floor-control variant: bids must be derived in code, not declared by
   the model.
2. **Classify the newest utterance, not the transcript blob.** Given one blob,
   the model re-reports the most interesting thing *anywhere* in it forever.
   `listen.framed()` separates "earlier, for context" / "already said aloud,
   don't repeat" / "classify ONLY this."
3. **Cooldown-as-veto was wrong; surcharge-that-decays is right.** A fixed
   cooldown swallowed a genuine `error` spotted 8s after a spend. Now recency
   is priced into the threshold (`Recent: 22`, decaying across 25s) so urgency
   can buy in early but nothing merely adequate can. This *removed* a branch
   from `Decide`.
4. **Early contradictions are hallucinated.** `contradiction` claimed with <4
   utterances of context is discounted to blindspot weight (`contextDepth` in
   budget.go) — the model confidently invented one on utterance 2.
5. **An opinion the model cannot phrase is not an opinion** — non-`none` kind
   with empty `line` is zeroed (session.apply).
6. `go build` failing with "error obtaining VCS status: exit 69" is the broken
   git, NOT a code problem. A stale binary from this once silently shipped
   old policy into a test run — always confirm `BUILD_OK` before re-testing.

## Operator context

- **Governing constraint: simplify until it hurts.** Answer findings by
  deleting requirements before building mechanisms. This repo exists partly
  because its predecessor (`roll-call`) went spec-first and
  bloated to 8,002 LOC. **Build first, demo, then decide if it deserves a
  spec.** Do not write a spec, do not add adapter seams, do not widen fields.
- Engineering style: Dave Cheney lineage — no `else`, ≤2 nesting per scope,
  policy vs mechanism, errors as values. See roll-call/CLAUDE.md §Engineering
  principles; this repo follows it and has no CLAUDE.md of its own yet
  (writing a thin one is a reasonable early task).
- Operator reaction so far: "that's cool actually." Liked the solo/N=1 shape.
  Interested in: **education** (tutor allowed 3 corrections per session —
  strongest pivot candidate; the whole domain lives in `prompts/classify.md`,
  so it's a prompt swap, not a code change), small-business voice apps, and a
  general brainstorm that never actually happened.

## Where to go next (in rough order of value)

1. **Commit.** Once the operator runs `sudo xcodebuild -license`: `git add -A`,
   one commit, done. Everything below is blocked on nothing but this is the
   cheapest insurance.
2. **Live-mic session with the operator.** The rehearsal is proven; the real
   product moment is the operator thinking out loud for 10 minutes and judging
   the spent turns. The worth/wasted buttons already feed calibration. This is
   an operator-driven validation — prepare it, don't simulate it.
3. **Education variant** — fork `prompts/classify.md` into a tutor prompt
   (kinds like `misconception` / `arithmetic-slip` / `productive-struggle` /
   `none`; spend only on what generalizes). Zero Go changes. A/B the two
   prompts on the same monologue.
4. **The silence ledger.** The most novel artifact observed: `worth 64, bar is
   84 (+18, just spoke)` — legible, accounted-for restraint. A post-session
   report of what it *chose not to say* might be worth more than the
   interruptions themselves. Small: the data already flows through the
   `reading` SSE event; persist + render it.
5. **Bigger sensor behind the same policy** — qwen misses real things
   (JWT-in-localStorage read as `none`). `-base-url`/`-model` flags already
   accept any OpenAI-shaped endpoint; try a hosted model as the sensor and
   measure what changes. Policy code must not change.

Anti-goals: multi-agent floor control (N× listening time, anti-economic),
plugin/adapter systems, persistence beyond what #4 needs, any spec document.

## Loose ends

- Server may still be running on :8722 (`pkill -f 'interject -addr'`) and
  Ollama may be up (`pkill ollama`) — operator wants Ollama stopped when idle.
- Favicon 404 — cosmetic, one line if it ever matters.
- No tests above the policy layer (deliberate for a POC; `listen.framed` and
  `session.apply` are the two spots that would earn tests first).
