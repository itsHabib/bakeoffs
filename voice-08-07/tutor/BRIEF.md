# Entry 4: hack-tutor — a tutor allowed three corrections per session

Read ~/dev/interject/docs/hackathon/README.md first — house rules and judging
apply verbatim. Repo: `~/dev/hack-tutor`. You never see the other three
entries.

## The bet

Good tutors shut up and let the student be wrong for a minute; software
never does. Every tutoring app corrects every slip instantly, which trains
students to stop thinking and wait for the buzzer. A tutor RATIONED to
three interruptions per session must spend them on the misconception that
generalizes — not the arithmetic slip the student will catch themselves.
Rationing is the pedagogy.

This entry doubles as an engine test: ~/dev/interject proved a rationed-
interruption engine on engineering monologues (READ ITS README AND
docs/kickoff.md FIRST — the sensor/policy split, the hard-won lessons about
small-model classification, and the continuous-scoring trick are all
transferable). Your question: does the engine transplant to a new domain in
one session? Fork/copy its code freely — this is the one entry where
copying an architecture is allowed, because measuring the transplant cost
is the experiment.

## What to build

A browser app where a student works a math problem OUT LOUD (mic or typed
lines) and a tutor with a 3-interruption budget listens.

- New classifier taxonomy in prompts/ (local qwen2.5:7b as the sensor):
  something like `misconception` (wrong mental model — will poison future
  problems), `procedural-slip` (sign error, dropped term — cheap, usually
  self-caught), `stuck` (silence/circling — might deserve a nudge),
  `productive-struggle` (wrong but progressing — NEVER interrupt), `none`.
  Weights must encode the pedagogy: misconception >> stuck >> slip.
- Budget policy stays deterministic code (interject's internal/budget
  pattern; retune, don't rebuild).
- Tutor speaks its interruption (speechSynthesis) as a QUESTION, not an
  answer ("what does the exponent apply to?") — enforce with a cap +
  question-mark check in code, not prompt hope.
- **The end-of-session report is the product moment:** everything it chose
  NOT to say, with timestamps and reasons ("saw sign slip at 1:42 — held:
  you caught it at 2:10"). A parent/teacher reads that and understands the
  product instantly.
- Two or three canned "student" monologues (rehearsal mode, like
  interject's): one with a real misconception buried mid-stream (e.g.
  (a+b)² = a²+b² while expanding), one where the student self-corrects a
  slip the tutor rightly sat on, one where they're genuinely stuck.

## What NOT to build

No curriculum, no problem generator, no accounts, no progress tracking, no
"adaptive learning" anything. One subject (algebra) — the taxonomy, not the
content, is what's being tested. Don't improve interject along the way;
copy what you need and diverge.

## Canned demo (required)

Rehearsal mode: the misconception monologue, ~60-90s. Judge watches the
tutor sit through two tempting slips and spend a turn only on the
misconception — then sees the silence report.

## The 60-second demo story

"Watch it NOT correct this sign error... the student just caught it
herself — that's the point. Now watch it spend one of its three turns:
'does that square apply to both terms?' — THAT one she'd have carried into
every problem this month. Here's the report of everything it held back.
Tutoring apps interrupt 40 times an hour; good teachers don't."
