# Consumer voice hackathon — 4 entries, independent Codex sessions, operator judges

Four self-contained briefs. Each goes to a FRESH Codex session that has never
seen the others — do not share context between entries. Every entry ends in a
runnable demo the operator can judge in under five minutes, hands-free.

This is the CONSUMER wave: the primary judging criterion is **would someone
download or pay**. Every brief names the buyer, the store category, and what
the first paid version charges for — the build must make that claim credible,
not just possible.

## Entries

| # | slug | bet |
|---|------|-----|
| 1 | `hack2-story-capture` | adult children pay $99/yr TODAY for worse versions of "interview grandma, get chapters" |
| 2 | `hack2-pitch-compressor` | founders/sales reps will pay to be forced into a 20-second conclusion, graded |
| 3 | `hack2-interview-sparring` | job seekers already spend on prep; real conversational pushback is the missing piece |
| 4 | `hack2-language-drill` | bounded phrase sets make pronunciation drilling gradeable in code, not vibes |

Launch one per Codex session with:

> Read ~/dev/interject/docs/hackathon-codex/entry-<slug>.md and build it. You
> are one of 4 independent entries; you win by demo, not by design.

## House rules (same for every entry — fairness is the point)

- **One session, one repo.** Scaffold at `~/dev/<slug>`. Go stdlib or plain
  Node server; vanilla web UI; NO build steps, NO frameworks, NO new deps
  unless unavoidable.
- **Real voice is available this round.** `OPENAI_API_KEY` is exported in
  `~/.zshrc` (run `source ~/.zshrc` if your shell predates it; verify with
  `test -n "$OPENAI_API_KEY" && echo ok` — never print the value). Use the
  OpenAI Realtime API for conversational voice via the proven pattern:
  **the browser holds the WebRTC connection; your server only mints a
  short-lived ephemeral client secret** (reference implementation:
  `~/dev/roll-call/internal/realtime/client.go`, ~125 LOC). Credentials live
  server-side only — never in client code, page source, logs, error
  messages, or the repo. **Form factor: web or iOS, whichever gets to a
  testable demo fastest — operator explicitly doesn't care.** Practical note:
  until the operator accepts the Xcode license, `xcodebuild` fails on this
  machine, so web is the only form that can actually build today; verify
  `xcodebuild -version` works before choosing iOS, and don't burn time on
  simulator setup if it doesn't.
- **Known machine breakage:** every `git` command exits 69 (unaccepted Xcode
  license). `git init` is fine, commits are impossible — do not burn time on
  it. Go builds need `-buildvcs=false`.
- **Correctness is computed, never model-judged.** Any grading, scoring,
  matching, coverage tracking, or budget lives in deterministic, table-tested
  code. The Realtime model is the VOICE — ears, mouth, conversational feel —
  never the judge of correctness. House invariant, every entry.
- **No spec, no design doc.** README + working demo + tests on the policy
  layer only. Simplify until it hurts.
- **Required at the finish line:** (a) README with one command to run,
  (b) **the LIVE Realtime voice session working — this is the demo.** The
  point of this wave is the conversation: real turn-taking, real voice, the
  thing speechSynthesis can't fake. An entry whose Realtime path doesn't
  work has missed the point of the round, however good its grader is.
  `DEMO.md` must script a ~60-second live exchange the judge performs
  themselves, (c) ALSO a canned replay mode (no mic, no key) — fixtures
  driven through the real grading/assembly engine — as the test harness for
  the deterministic core and the fallback if the live demo hits network
  trouble on judging day, (d) `DEMO.md` also asserts buyer, store category,
  and first paid price point plainly, (e) local green: build/vet/tests pass.

## Judging (operator, after all 4 land)

100 points — note the reweighting: paying is primary this wave.

- **35 — would someone pay.** Named buyer, evidence money already moves,
  and does the built thing make the price plausible? Asserted in DEMO.md;
  the judge will push back.
- **25 — the 60-second LIVE demo.** The judge talks to it. Does the
  conversation feel real — turn-taking, interruption, voice quality — and
  does the product moment land in it? Canned replay scores at most half of
  this slot; it exists as harness and fallback, not as the show.
- **15 — deterministic share.** How much of the correctness path is real
  code with tests vs model vibes. Show the test file.
- **15 — voice-necessity.** If this would be just as good as a text/typing
  app, it loses these points. Name that alternative in your README and say
  why voice beats it.
- **10 — restraint.** Small LOC, no seams for futures that don't exist,
  deleted requirements > built mechanisms.

Tie-breaker: which repo would the operator actually open again next week.

## After judging

Winner gets the follow-through: a market-research brief, a live session with
the operator as the user, and — once the Xcode license is accepted — the iOS
port decision (the deterministic core carries unchanged; only the surface
rebuilds). Losers keep their repos as reference — nothing gets ported into a
platform.
