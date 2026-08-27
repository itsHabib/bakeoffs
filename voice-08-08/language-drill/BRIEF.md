# Entry 4: hack2-language-drill — pronunciation reps against a bounded phrase set

Read ~/dev/interject/docs/hackathon-codex/README.md first — house rules and
judging apply verbatim. Repo: `~/dev/hack2-language-drill`. You never see the
other 3 entries.

## The bet

Language apps are a proven-payer market (Duolingo, Pimsleur, ELSA — ELSA
charges ~$12/mo for pronunciation alone), but speaking practice is either
absent or scored by an opaque model. A BOUNDED phrase set flips it: when the
target is one of 40 known phrases, checking what you said against what you
meant to say is deterministic string work, the gap is explainable
word-by-word, and progress is a number you can trust. Buyer: adult learners
prepping for travel or a move (pick ONE language and ONE scenario — e.g.
Spanish, "ordering and asking in a restaurant"). Store category: Education.
First paid version: $7.99/mo per scenario pack.

## What to build

A browser app that drills one 40-phrase scenario pack.

- **The deterministic core is the matcher/scorer.** The pack lives in code:
  each phrase with target text, acceptable variants, and the 2-3 known
  learner traps (dropped article, wrong gender, anglicized word — encode
  per-phrase). Recognize the attempt (Web Speech with `lang` set to the
  target language; note its vendor-audio caveat in the README), then score
  in code: normalized edit distance per word, trap detection by rule,
  pass/retry threshold. The word-level diff — green/red per word, trap
  named — is the screen. Table-test the matcher hard: accents, variants,
  near-misses. TODO-mark anything about the target language you are not
  sure of rather than inventing it.
- **Realtime voice is the drill partner:** it speaks the phrase natively
  first (say-after-me), and in "scene mode" plays the waiter — sticking to
  utterances the pack expects, so the learner's replies stay inside the
  bounded set. Browser WebRTC + ephemeral token per README. The model never
  scores; it models.
- **Session loop:** 10 phrases per rep; misses recycle with spaced
  repetition (simple in-session queue, not a persistence system). End
  screen: per-phrase history, which traps you keep hitting, tomorrow's 10.

## What NOT to build

No second language, no second scenario, no grammar lessons, no vocabulary
flashcards, no model-based pronunciation scoring ("87% native-like" vibes),
no accounts, no streaks, no gamification.

## Canned demo (required)

Bundled fixture attempts (write ~8: a clean pass, a dropped article, a
gender trap, an anglicized word, a total miss) replayed through the real
matcher — judge watches the word-level diffs and trap callouts render, no
mic, no key.

## The 60-second demo story

"Say 'la cuenta, por favor' — [replay shows attempt] — he said 'el cuenta'.
Watch the diff: one word red, and it NAMES the trap — gender article, his
third time this session. No AI grade pretending to measure his accent; the
target was known, the check is string math you can read. ELSA charges $12 a
month for a black box. This one shows its work."
