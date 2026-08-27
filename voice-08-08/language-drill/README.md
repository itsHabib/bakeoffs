# Mesa 40

Spanish restaurant speaking reps against one inspectable, 40-phrase pack.
The learner chooses a known phrase, speaks it, and sees a word-by-word diff plus
named traps such as a wrong-gender article or an English carry-over. The model
provides the voice; normal Go code computes every score.

## Run

The canned demo needs no microphone and no API key:

```sh
go run -buildvcs=false .
```

Open <http://127.0.0.1:8740> and select **Run 60-second demo**.

For live OpenAI voice, make the already-configured key available to the server
(source `~/.zshrc` if needed), use the same run command above, then select
**Connect OpenAI voice** and allow microphone access. The standard API key stays
on the server; the browser receives only a short-lived Realtime client secret,
adds its microphone track to the WebRTC peer connection, and exchanges live
audio directly with OpenAI. Semantic turn detection creates replies and permits
interruption.

## What is real

- Forty Spanish restaurant phrases, accepted variants, and at least two named
  trap rules per phrase live in `internal/drill`.
- Scoring normalizes punctuation and common Spanish accents, selects the
  closest accepted variant, computes word-level edit distance, aligns the
  attempt for display, and checks explicit trap rules.
- A pass requires at least `0.82` similarity and no detected trap.
- Eight canned attempts pass through the same HTTP scoring endpoint used by
  typed and recognized speech.
- A ten-phrase in-session queue puts misses back at the end. No persistence or
  fake streak machinery exists.
- Browser speech synthesis is the keyless fallback. When connected,
  `gpt-realtime-2.1` models phrases and speaks short waiter responses. It
  never receives authority to score or advance the drill.

## Why voice belongs

Typing can test recall, but it cannot create the pressure of hearing a waiter,
forming the phrase, and saying it before the moment passes. The interaction is
a speaking rep. The transcript is merely the inspectable sensor output used by
the deterministic matcher.

## Buyer and paid shape

- **Buyer:** an adult preparing for travel or a move who wants to rehearse one
  concrete situation rather than take a general language course.
- **Store category:** Education.
- **First paid version:** $7.99/month for a reviewed scenario pack,
  live scene mode, and session history.

Adults already pay roughly $12/month for pronunciation products such as ELSA.
This smaller price is credible for a reviewed travel scenario whose corrections
are named and inspectable instead of hidden behind an opaque percentage.

## Tests

```sh
go test -buildvcs=false ./...
go vet -buildvcs=false ./...
go build -buildvcs=false ./...
```

The table tests cover all 40 pack entries, accepted variants, punctuation,
accent normalization, near misses, explicit traps, missing words, and a total
miss.

## Boundaries and limitations

- This scores recognized words, not accent quality, phonemes, cadence, or
  “native-likeness.” It makes no claim to do so.
- Browser Web Speech recognition may send microphone audio to the browser
  vendor. Typed and canned modes do not use it.
- Realtime responses can vary. They are presentation only; the drill remains
  correct even if scene mode says something awkward.
- The phrase pack needs review by a qualified native-speaking teacher before
  anyone pays. Regional restaurant vocabulary also varies.
- Accent marks are deliberately normalized for recognition tolerance, so the
  app does not teach Spanish spelling.
- The current live path uses browser Web Speech for the scored transcript and
  OpenAI Realtime for spoken modeling/replies. Replacing the sensor must not
  change the scorer.

## Shape

```text
main.go                 static server, score API, Realtime secret broker
internal/drill/         40-phrase pack, matcher, trap rules, tests
web/                    vanilla interface, browser recognition, WebRTC client
DEMO.md                 exact one-minute walkthrough
```
