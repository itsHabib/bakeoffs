# Pitch Compressor

Twenty seconds, one conclusion. Pitch Compressor runs three spoken reps of the
same business pitch and grades each transcript with visible, deterministic
rules. The OpenAI Realtime model is only the skeptical listener: it asks one
question after a live rep and never awards or removes points.

## Run it

Node 20+ is the only requirement. There are no package dependencies or build
steps.

```sh
npm start
```

Open <http://127.0.0.1:4173>. Click **Replay 3 canned reps** for the complete
hands-free canned demo. It needs no microphone and still works if
`OPENAI_API_KEY` is absent.

For the live exchange, use Chrome, click **Connect live listener**, then start
the 20-second rep. The local server exchanges
`OPENAI_API_KEY` for a short-lived browser credential. The long-lived key is
never sent to or stored by the browser.

## What is real code

[`grader.mjs`](./grader.mjs) starts from 100 and applies these
fixed rules:

| Rule | Fixed deduction |
|---|---:|
| Speech reaching or crossing the 20-second wall | 18 |
| Filler phrase (`um`, `like`, `sort of`, `basically`) | 4 each, capped at 20 |
| Pace outside 1.5–3.5 words/second | 12 |
| Hedge (`I think`, `maybe`, `hopefully`) | 4 each, capped at 16 |
| No concrete ask in the final quarter | 20 |
| At least 25% repeated bigram overlap with an earlier rep | 16 |

The canned grader maps simulated duration onto the transcript at exactly 20
seconds and deliberately cuts inside the crossing word. Chrome's live Web
Speech API does not expose word timing, so live mode preserves finalized speech
and labels the browser's still-interim fragment as what the wall ate.

The concrete-ask detector accepts a number or an ask verb in the final quarter.
The source has an explicit `TODO`: treating an ask verb as an imperative is a
blunt heuristic and can produce false positives. It is inspectable and
repeatable rather than disguised as model judgment.

The fixture session at the top of [`app.mjs`](./app.mjs) is graded by
the exact same function as live speech:

- Rep 1: rambling, four hedges, cut off before its ask — **30**.
- Rep 2: shorter but hedged, with no ask — **56**.
- Rep 3: 19.2 seconds, no filler or hedges, number and ask — **84**.

## Voice architecture

Live transcription uses the browser's Web Speech API. Depending on browser and
platform, the browser vendor may transmit microphone audio to its own speech
service. That audio path is outside this app's control; do not use live mode for
confidential pitches. Canned mode uses neither microphone nor vendor speech.

The optional listener uses browser WebRTC directly with the OpenAI Realtime
API. The server only mints a short-lived client secret. Turn detection has
automatic responses disabled, so the listener remains silent during the pitch.
After deterministic grading, the browser sends the frozen transcript and rule
gaps and asks for exactly one spoken question. The listener cannot edit the
scorecard.

Typing would be easier technically, but it would erase the product constraint:
spoken audio costs the listener real linear time. The countdown, involuntary
fillers, speaking pace, and deliberate cutoff are the exercise.

## Checks

```sh
npm run check
```

The tests table-check wall splitting, repeated filler and hedge counts, pace
boundaries, the final-quarter ask heuristic, n-gram repetition, transcript
diffing, empty input, scoring weights, and clamping. Server tests prove the API
key stays upstream, only an ephemeral credential reaches the browser, missing
configuration fails closed, and static serving is allowlisted. The check also
syntax-checks the server and browser app.

## Scope and limitations

- Exactly three reps per in-memory session; refreshing clears everything.
- Exactly one duration: 20 seconds.
- No accounts, templates, video, team features, analytics, or saved history.
- Web Speech recognition quality and availability vary by browser. Chrome is
  the supported live-demo path.
- A live mid-word cut is inferred from the current interim Web Speech fragment;
  the API does not provide timestamps for that partial fragment.
- Realtime voice requires network access, microphone permission, and a valid
  server-side key. The entire graded demo does not.
- The ask rule recognizes surface forms, not grammatical intent.

## Buyer and paid shape

The first buyer is a founder preparing investor calls or a sales representative
preparing discovery calls. Store category: **Productivity / Business**. First
paid version: **$9.99/month for unlimited reps plus history**. This demo omits
history deliberately; it proves the repeat-and-improve loop before building the
subscription surface.
