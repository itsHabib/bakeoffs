# Voice hackathon — four entries, independent sessions, operator judges

Four self-contained briefs. Each goes to a FRESH session that has never seen
the others — do not share context between entries. Every entry ends in a
runnable demo the operator can judge in under five minutes.

## Entries

| # | slug | bet |
|---|------|-----|
| 1 | `hack-atc` | people pay to drill a bounded radio grammar (ATC trainer wedge) |
| 2 | `hack-readback` | irreversible ops actions gated by spoken readback (voice merge gate) |
| 3 | `hack-partyline` | an agent fleet you monitor by EAR via 2-second brevity calls |
| 4 | `hack-tutor` | interject's rationed-interruption engine transplants to tutoring in one session |

Launch one per session with:

> Read ~/dev/interject/docs/hackathon/entry-<slug>.md and build it. You are
> one of four independent hackathon entries; you win by demo, not by design.

## House rules (same for every entry — fairness is the point)

- **One session, one repo.** Scaffold at `~/dev/<slug>`. Go stdlib bias or
  plain Node; vanilla web UI; NO build steps, NO frameworks, NO new deps
  unless unavoidable.
- **Keyless and local.** No OPENAI_API_KEY exists on this machine. Voice out =
  browser `speechSynthesis`. Voice in (if needed) = browser Web Speech
  (Chrome/Safari). Local model = Ollama `qwen2.5:7b` via
  `http://localhost:11434/v1` (start `ollama serve`, stop it when done —
  operator preference). If your idea "needs" a better model, your demo must
  work without one and merely note the upgrade path.
- **Git is broken on this machine** (Xcode license; every git call exits 69).
  `git init` if you like but you cannot commit. Go builds need
  `-buildvcs=false`. Do not burn time on this.
- **Correctness is computed, never model-judged.** Any grading, matching,
  gating, or budget lives in deterministic, table-tested code. The model (if
  you use one at all) is ears, mouth, or phrasing only. This is the house
  invariant every entry inherits.
- **No spec, no design doc.** README + working demo + tests on the policy
  layer only. Simplify until it hurts.
- **Steal mechanism, not architecture:** ~/dev/interject (budget/session/SSE
  patterns, speech in/out wiring, prompts-reloaded-per-call trick),
  ~/dev/roll-call/internal/voice/matcher.go (closed-set phrase matching).
- **Required at the finish line:** (a) README with one command to run,
  (b) a scripted/canned demo mode that needs NO mic so the judge can run it
  hands-free, (c) `DEMO.md` — the exact 60-second walkthrough you'd perform,
  (d) local green: build/vet/tests pass.

## Judging (operator, after all four land)

100 points:

- **30 — the 60-second demo.** Does a person watching get it, and does it
  land? Canned mode counts; a live-mic moment that works is a bonus.
- **25 — would someone pay.** Named buyer, evidence money already moves in
  the space. Assert it in DEMO.md; the judge will push back.
- **20 — deterministic share.** How much of the correctness path is real
  code with tests vs model vibes. Show the test file.
- **15 — voice-necessity.** If this would be just as good as a text app, it
  loses these points. The entry must exploit something only audio has.
- **10 — restraint.** Small LOC, no seams for futures that don't exist,
  deleted requirements > built mechanisms.

Tie-breaker: which repo would the operator actually open again next week.

## After judging

Winner gets the follow-through: real research brief (entry 1 already has one
at docs/research/atc-trainer-brief.md), live-mic operator session, and a
decision on whether it earns a spec. Losers keep their repos as reference —
nothing gets ported into a platform.

## Round 2 (operator-decided 2026-08-07): iOS + real voice

Round 1 (the four web entries above) is the IDEA bake-off — cheap, keyless,
judged on whether the bet lands. Round 2 re-runs the winner(s), pivoted or
adapted, in the shape the operator actually wants:

- **Native iOS app**, run/verified in the simulator.
- **OpenAI Realtime voice** — real conversational audio, not speechSynthesis.
  Key: exported in ~/.zshrc (sessions that predate the export must
  `source ~/.zshrc`). Server mints ephemeral tokens; client holds the
  audio (pattern: ~/dev/roll-call/internal/realtime/client.go). Credentials
  never in client code, logs, or repos.
- **The bar moves from "demo lands" to "someone would download or pay".**
  This is the operator's primary criterion for round 2. Every round-2 brief
  must name the buyer, the store category, and what the first paid version
  charges for. An entry that is impressive but has no plausible downloader
  fails, regardless of craft.
- Prereq the operator must clear first: `sudo xcodebuild -license`
  (unblocks xcodebuild AND git on this machine).
- Launch mechanics: the /hackathon skill (prep → launch → judge), one
  fresh session per entry, same independence rule.

Deterministic cores (grammars, graders, budgets) port from round 1
unchanged; round 2 rebuilds the surface, not the correctness layer.
