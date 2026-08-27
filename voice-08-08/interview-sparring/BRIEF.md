# Entry 3: hack2-interview-sparring — a practice interviewer that pushes back

Read ~/dev/interject/docs/hackathon-codex/README.md first — house rules and
judging apply verbatim. Repo: `~/dev/hack2-interview-sparring`. You never see
the other 3 entries.

## The bet

Job seekers already spend real money on prep — coaching sessions, mock
interview marketplaces, question banks — and the thing none of the cheap
options provide is a partner with real conversational timing that interrupts
a rambling answer and follows up on the weak part. Sub-300ms turn-taking is
THE feature here: pushback only lands when it arrives like a human's would.
Buyer: active job seekers (behavioral interviews first — biggest pool).
Store category: Education / Careers. First paid version: $14.99/mo during a
job search — pain is spiky and people pay during spikes.

## What to build

A browser app that runs one mock behavioral interview round and debriefs it.

- **The deterministic core is the session structure + the metrics.** A
  question bank in code (8-10 classic behavioral questions with per-question
  follow-up angles). A session state machine: question → answer → 0-2
  follow-ups → next, with a hard per-answer time budget. Answer metrics
  computed in code from transcript + timing: answer length vs budget,
  filler density, a STAR-coverage checklist (situation/task/action/result
  detected by rule — TODO-mark the heuristics, keep them readable),
  question-dodge detection (answer shares almost no content words with the
  question — n-gram overlap rule).
- **Realtime voice is the interviewer.** It asks the bank's question, and
  when your metrics show a dodge or a missing STAR beat, the state machine
  hands it that gap as the follow-up angle to press — "you told me the
  situation; what did YOU do?" It may barge in when the time budget blows
  (semantic VAD interrupt — see roll-call's session config for the knobs).
  The model phrases; the state machine decides when and about what.
- **The debrief screen is the paid product:** per-question scorecards, your
  worst dodge quoted back verbatim, time-budget graph, and the two questions
  to re-run tomorrow. All computed.

## What NOT to build

No resume parsing, no job-description tailoring, no technical/coding
interviews, no video, no hiring-manager marketplace, no scoring by the
model, no persistence beyond the session.

## Canned demo (required)

A bundled fixture interview (write it: 3 questions, one solid answer, one
rambling answer that blows the budget and gets barged, one dodge that draws
the follow-up) replayed through the real state machine and metrics — judge
watches the follow-up get chosen for the right reason and the debrief
render, no mic, no key.

## The 60-second demo story

"Ask anyone how interview prep goes: you rehearse answers alone and they
evaporate under the first follow-up. Watch — [replay] — he never said what
HE did, just what the team did. The machine caught it — n-gram rule, not AI
vibes — and fed the interviewer exactly that press. Here's the debrief: his
worst dodge, quoted. Job seekers pay $150 for one human mock interview.
This is $15 a month and it never gets tired."
