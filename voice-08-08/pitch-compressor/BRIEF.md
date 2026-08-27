# Entry 2: hack2-pitch-compressor — your conclusion in 20 seconds, graded

Read ~/dev/interject/docs/hackathon-codex/README.md first — house rules and
judging apply verbatim. Repo: `~/dev/hack2-pitch-compressor`. You never see
the other 3 entries.

## The bet

Audio is linear time — padding costs the listener, and everyone pads.
Forcing a conclusion into 20 seconds of speech reveals instantly whether you
know what matters; people pay coaches real money to be forced through
exactly this. Buyer: founders prepping investor calls, sales reps prepping
discovery calls, anyone with a promotion case to make. Store category:
Productivity / Business. First paid version: $9.99/mo for unlimited reps +
history.

## What to build

A browser app that runs timed pitch reps and grades them in code.

- **The deterministic grader is the product.** Grade a transcript + timing
  data against rules: hard 20-second wall (what got cut when the wall hit),
  filler-word count ("um", "like", "sort of", "basically" — list in code),
  words-per-second sanity band, hedge detection ("I think", "maybe",
  "hopefully" — list in code), presence of a concrete ASK (rule-based:
  imperative or number in the final quarter — TODO-mark this heuristic
  honestly), and a repeated-content penalty (n-gram overlap between reps).
  Score 0–100 from weights in code. Table-test the whole thing.
- **Rep loop:** you speak against a visible 20-second countdown (Web Speech
  for transcription is acceptable here; note its vendor-audio caveat in the
  README). Wall hits → cut off mid-word, deliberately. Scorecard renders
  instantly: what survived, what the wall ate, every rule's line item.
- **Realtime voice is the listener you pitch AT** — it plays a skeptical but
  fair audience: after your rep it asks the ONE question your pitch left
  open (it gets your transcript + your scorecard gaps as context). It never
  scores; it makes the next rep harder. Browser WebRTC + ephemeral token
  per README.
- Three reps of the same pitch = a session; show the score trend and the
  diff of what you cut between reps. Getting shorter AND scoring higher is
  the product moment.

## What NOT to build

No video, no team features, no pitch templates library, no LLM scoring or
"AI feedback" prose, no history beyond the session, no more than one
duration mode (20s only).

## Canned demo (required)

Three bundled fixture reps of the same startup pitch (write them:
rep 1 rambling and wall-cut mid-sentence, rep 2 tighter but hedged with no
ask, rep 3 clean with a number and an ask) replayed with simulated timing
through the real grader — judge watches three scorecards and the trend line,
no mic, no key.

## The 60-second demo story

"Twenty seconds, one conclusion. Rep one — [replay] — the wall cut him off
before the ask ever came, and the grader shows it: four hedges, zero ask.
Rep three, same pitch, next morning — 19.2 seconds, one number, one ask,
score 84. Nothing here is AI judgment — every point on that scorecard is a
rule you can read. Founders pay coaches $200 an hour to do this with a
stopwatch."
