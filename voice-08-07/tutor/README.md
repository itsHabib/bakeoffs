# hack-tutor

A math tutor that is allowed to interrupt three times per session — and has to
decide, while the student is still talking, whether what it noticed is worth
one of them. At the end it hands over a report of everything it chose *not*
to say.

Everything runs locally: a 7B model on Ollama as the ears, the browser's own
speech synthesis as the mouth, and a Go server holding the budget. No API key,
no cloud, no account.

## Run it

```bash
go run -buildvcs=false .
```

(Needs `ollama serve` running with `qwen2.5:7b` pulled.) Open
http://127.0.0.1:8731 and press **▶ Misconception** — a canned student works
two problems out loud for ~80 seconds, no mic needed. The other two buttons
play a self-caught slip and a genuinely stuck student. The mic button works in
Chrome/Safari; typing lines works everywhere.

## The idea

Every tutoring app corrects every slip instantly, which trains students to
stop thinking and wait for the buzzer. Good tutors do the opposite: they sit
on their hands while a student is wrong, because a sign error the student
catches herself is worth more than a correction, and an interruption is only
worth spending on the mistake that *generalizes*. So the ration is the
pedagogy: three interruptions per session, spent like money.

- `misconception` (weight 90) — a wrong rule ((a+b)² = a²+b²) that will poison
  every problem this month. Clears an on-pace bar immediately.
- `stuck` (62) — no forward motion. Has to *persist* to earn a nudge, and a
  stuck claim two lines into a problem is discounted to nothing, because
  that is a student reading, not a student stuck.
- `procedural-slip` (34) — 3×3 written as 6. Sits **below the floor by
  construction**: there is no budget rich enough to interrupt one. The tests
  pin this.
- `productive-struggle` (16) — wrong but moving. Never interrupted; that is
  what learning sounds like.
- `self-correct` (0) — the payoff. It resolves the matching held entry, so
  the report can say "held: you caught it yourself 5 seconds later."

When the tutor does spend a turn, it must speak a **question** — "does that
square apply to both terms?" — never an answer. That is enforced in code
(`budget.Askable`: ends in `?`, ≤16 words), not hoped for in the prompt: an
opinion the model can only phrase as a lecture costs the student nothing.

**The report is the product.** Interruptions are what any app has; the ledger
of held opinions — each with a timestamp, what it would have asked, why the
budget said no, and whether the student caught it themselves — is what a
parent reads to understand in one pass why this thing stays quiet.

## The experiment: transplanting interject

This repo is a one-session test of whether
[interject](../interject)'s rationed-interruption engine transplants from
engineering monologues to tutoring. Verdict: **the engine moves; the work is
all in the sensor.**

What the transplant cost, by layer:

| layer | change |
|---|---|
| `internal/stream` | none — copied verbatim |
| `internal/listen` | two framing strings ("the student", "already asked") |
| `internal/budget` | taxonomy + weights swapped; `contextDepth` guard became the `stuckDepth` guard; one new pure function (`Askable`); floor 44→40, recency 22→25. The envelope / pacing threshold / surcharge / calibration machinery: untouched |
| `internal/session` | the genuinely new code: the held ledger, self-correct resolution, the report (~80 lines) |
| `prompts/classify.md` | rewritten — the domain lives here, as predicted |
| `web/` | UI copy, three rehearsals, the report overlay |

Where the session actually went: prompt-tuning the 7B sensor against live
probes. Three real misfires found and fixed — it read an arithmetic slip as a
high-confidence misconception (spent a turn on it, then held the *real*
misconception behind the surcharge); it missed self-corrections phrased as
"hold on…"; and it hallucinated a misconception from the harmless phrase
"that seems ugly". Every fix was an example added to the prompt, verified by
probing; zero policy code changed. Interject's hard-won lesson — the model is
a sensor, all scarcity in deterministic code — held up under transplant, and
its per-call prompt reload made the tuning loop instant.

## Shape

```
main.go              flags, routes, embedded UI
internal/budget      policy — the ration and the pedagogy. pure, table-tested
internal/listen      mechanism — one classification from any OpenAI-shaped endpoint
internal/session     composition — transcript in, decisions + the held ledger out
internal/stream      mechanism — SSE fan-out
prompts/classify.md  the tutor's ears — reloaded every call, tune it live
web/                 vanilla HTML/CSS/JS, no build step
```

Tests live where the pedagogy lives: `go test ./...` covers the budget layer,
including the invariants that a slip and productive struggle can never be
interrupted at any confidence or budget, that early stuck claims are
discounted, and that non-questions are unspeakable.

## Known rough edges

- `qwen2.5:7b` cannot check arithmetic ("11+7=17" reads as correct). The
  rehearsals use pattern-shaped mistakes it can see; a live session with a
  hosted sensor (`-base-url`/`-model` flags speak the OpenAI wire format)
  would catch more. The policy is the interesting part; the sensor is
  swappable.
- Browser speech recognition sends audio to the vendor. Only the mic path
  does; rehearsals and typing are fully local.
- `-buildvcs=false` is needed on this machine because `git` exits 69 until
  the Xcode license is accepted. Drop the flag once that is done.

## What this is not

No curriculum, no problem generator, no accounts, no progress tracking, no
adaptive anything. One subject, three interruptions, one report.
