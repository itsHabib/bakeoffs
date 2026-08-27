# partyline — fleet brevity channel

Monitor an agent fleet by EAR. Six agents (three drivers, two reviewers, a CI
watcher) share one audio channel, and every transmission is ~2 seconds of
fixed phraseology — "roxiq driver, stream one, tests green." — the way pilots
on a common frequency build a picture of the pattern without looking. You
half-listen while doing something else and just know where the fleet is.

## Run

```
npm start
```

Open http://localhost:8765 in Chrome or Safari, turn your sound on, press
**▶ Open channel**. No dependencies, no build step, no mic, no API key, no
model.

## Test

```
npm test
```

76 tests on the policy layer (Node's built-in runner, no dependencies).

## What's real code vs what's browser

Everything that decides anything is deterministic, tested, pure JavaScript:

- [src/sim.js](src/sim.js) — seeded fleet simulator. Realistic delivery state
  machines (dispatch → tests → review → findings → gate → merge) with a forced
  test failure, a stuck agent, and a CI red. Same scenario → byte-identical
  timeline, every run.
- [src/grammar.js](src/grammar.js) — the brevity grammar. Every event type maps
  to a fixed template with constant ordering (callsign, subject, status, need),
  a hard word cap (9 routine / 12 escalation), one unique stressed status word
  per event type, and a "PAN PAN" attention prefix on escalations. Plus the
  prose control group: the same events as natural sentences, for the toggle.
- [src/channel.js](src/channel.js) — the priority policy. Escalations always
  transmit, jump the queue, and are never coalesced or dropped. Routine calls
  coalesce per agent+stream (latest state supersedes) and shed oldest-first
  when the frequency saturates. Priority is code, not vibes.

The browser part ([app.js](app.js)) is only ears and mouth: `speechSynthesis`
with a distinct voice/pitch per agent, squelch clicks between transmissions,
an alert tone before escalations, and the screen — a fleet board showing
ground truth and a transcript showing only what aired. The gap between those
two is the drop policy, visible.

There is **no model anywhere in this repo**. The grammar is templates; the
correctness path is 100% computed. Upgrade path if it ever earns one: nicer
TTS voices, and real driver/CI events instead of the fixture fleet — the
grammar and channel policy stay exactly as they are.

## The two modes

**Brevity** is the product. **Prose** is the control group: the same events
spoken as natural sentences — what every "agents talk to you" demo does. In
prose mode the channel physically cannot keep up: routine traffic gets shed
in bulk, and the audio narrates a reality the fleet left half a minute ago.
The stats row (on air / coalesced / dropped) keeps both modes honest.

See [DEMO.md](DEMO.md) for the 60-second walkthrough.
