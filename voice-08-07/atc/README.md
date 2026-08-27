# hack-atc — radio phraseology trainer (VFR pattern wedge)

A tower controller you can practice with at 11pm. One scenario: VFR pattern
work at a towered field — taxi, takeoff clearance, pattern sequencing,
landing clearance, taxi back. The controller speaks; you read back; **a
deterministic grader scores every required element of your readback live,
as you say it** — present, missing, or *wrong* (read back runway 27 when
you were cleared runway 9 and it catches the swap mid-sentence).

There is no LLM anywhere in this app. Controller lines are templates and the
grader is token tables + regexes (`public/grader.js`), so grading is instant,
free, repeatable, and table-tested.

## Run

```bash
node server.mjs
```

Open http://localhost:8329. No dependencies, no build step. Node 18+.

- **Judge mode (no mic):** press **▶ Fly the script**. A scripted student
  flies the whole pattern and makes three classic readback mistakes —
  dropped callsign, "roger" instead of a readback, wrong runway — and the
  grader catches each one on screen. Runs itself in ~60 seconds.
- **Live mode:** type your readback and hit TRANSMIT (works in any browser),
  or press 🎙 and speak (Chrome/Safari, Web Speech API). Grading updates
  per keystroke / per interim speech result.

Controller audio uses `speechSynthesis` (rate up, pitch down, a static chirp
before each transmission). The app is fully usable muted.

## Tests

```bash
node --test
```

43 table-driven tests over the normalizer and the grading policy — spoken
vs typed forms ("runway niner" ≡ "runway 9", "one two one point niner" ≡
"121.9", "skyhawk one two three alpha bravo" ≡ "Skyhawk 123AB"), element
order independence, wrong-value detection (runway, frequency, pattern
direction, sequence number), roger-only detection, score math, and the
canned script's exact failure points.

## Layout

| file | what |
|---|---|
| `public/grader.js` | normalizer + element grader (pure, shared with tests) |
| `public/scenario.js` | the one scenario: controller lines, readback specs, canned student |
| `public/app.js` | UI, TTS/mic wiring, canned-demo driver |
| `server.mjs` | stdlib static server |
| `test/grader.test.js` | the policy tests |

Phraseology follows FAA AIM ch. 4 conventions from general knowledge; the
one call I'm unsure of is marked `TODO(phraseology)` in `scenario.js`
rather than asserted confidently.
