# Interview Sparring

A browser-based behavioral interviewer that interrupts rambling answers and
presses the exact STAR beat the candidate skipped. OpenAI Realtime supplies the
ears, voice, and conversational timing. A deterministic state machine owns the
questions, follow-up topic, time budget, metrics, score, and debrief.

## Run it

```sh
npm start
```

Open <http://127.0.0.1:4173>. **Watch 3-question replay** is the complete canned
demo: no microphone, API key, install, or network request is required.

For live voice, make `OPENAI_API_KEY` available to the server (source
`~/.zshrc` if needed), then use the same run command above.

The browser receives only a short-lived Realtime client secret. The permanent
key is used solely by the local server and is never returned, logged, embedded
in the page, or saved in this repository.

## What is real

- Nine classic behavioral questions and question-specific follow-up angles are
  frozen in `lib/interview-core.js`.
- `InterviewMachine` enforces `question → answer → 0–2 follow-ups → next`.
- Every primary answer gets a hard 45-second budget.
- Word count, filler density, STAR coverage, question-term overlap, dodge
  detection, and the score are ordinary JavaScript functions.
- The debrief contains per-question scorecards, a time-budget graph, the worst
  dodge quoted verbatim, and the two lowest-scoring questions to rerun.
- The three-answer fixture—solid, interrupted ramble, dodge plus recovery—runs
  through that same state machine and metric code.

The STAR detector is deliberately readable lexical policy. Its `TODO` marks
the honest limitation: the cues need validation against a labelled transcript
corpus before the numeric score should be presented as coaching truth. The app
does not ask a model to conceal that uncertainty with a confident grade.

## Voice boundary

Live mode uses browser WebRTC and server-minted ephemeral credentials with
`gpt-realtime-2.1`. Input transcription is `gpt-4o-mini-transcribe`.
Semantic VAD uses high eagerness, automatic response creation is disabled, and
interruptions are enabled. The browser sends `response.create` only after the
state machine supplies an exact banked question or follow-up angle.

The text alternative is another interview questionnaire. Voice is better here
because the product moment is temporal: speaking for 45 seconds under pressure,
getting cut off, and hearing a direct follow-up before a rehearsed answer can be
silently rewritten. Typing removes the failure mode the product is designed to
train.

## Checks

```sh
npm run check
```

There are no runtime dependencies or build step. The check runs syntax
validation plus policy, fixture, transition, HTTP security, and token-boundary
tests.

## Product claim

- Buyer: an active job seeker preparing for behavioral interviews.
- Store category: Education / Careers.
- First paid version: **$14.99/month during the job-search spike**.
- Scope intentionally excluded: resume parsing, job-description tailoring,
  technical interviews, video, marketplaces, model scoring, accounts, and
  persistence.
