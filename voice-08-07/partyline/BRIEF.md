# Entry 3: hack-partyline — fleet brevity channel

Read ~/dev/interject/docs/hackathon/README.md first — house rules and judging
apply verbatim. Repo: `~/dev/hack-partyline`. You never see the other three
entries.

## The bet

Multi-agent voice was written off as anti-economic: N agents speaking prose
= N× listening time, and audio can't be skimmed. Brevity codes invert that.
If every transmission is ~2 seconds of standard phraseology — "roxiq driver,
stream two, tests green, holding for review" — a shared channel becomes
ambient awareness, the way pilots on a common frequency build a picture of
the pattern without looking. You half-listen while doing something else and
just KNOW where the fleet is. Nobody has built the party line.

## What to build

A browser app that renders a simulated agent fleet as an AUDIO channel.

- A fleet simulator in code: 5-8 fake agents (drivers, reviewers, a CI
  watcher) advancing through realistic state machines (dispatch → tests →
  review → findings → merge; with failures, retries, one stuck agent).
  Deterministic, seeded, ~3-4 minutes per run — this is fixture data with
  a clock, not real integration.
- A brevity grammar in code: every event type maps to a fixed phrase
  template — who, what, state, need. Templates enforce a hard word cap.
  The grammar is the deliverable: design it so distinct events are
  distinguishable BY EAR half-attended (leading callsign, stressed status
  word, consistent ordering).
- Speak the channel with `speechSynthesis`: one transmission at a time,
  strict queue, distinct voice or pitch per agent if the API allows,
  squelch-click separators between transmissions.
- **Priority is code, not vibes:** routine traffic (state advances) is
  droppable when the queue backs up — coalesce or skip, like a real
  saturated frequency. Escalations (failure, stuck, needs-human) always
  transmit and use an attention prefix ("PAN PAN, stream three"). Steal
  interject's budget idea if it helps: routine calls rationed, escalations
  exempt.
- Screen shows the transcript scrolling with the audio plus a fleet board,
  but the DEMO CLAIM is that eyes stay elsewhere — the screen exists so the
  judge can verify what their ears told them.

## What NOT to build

No mic, no voice input at all — this entry is output-only. No real ship/
driver integration (fixture fleet only). No prose mode for comparison
beyond ONE toggle that speaks the same events as natural sentences — that
toggle is your best demo weapon (30s of prose chaos vs 30s of brevity).

## Canned demo (required)

The whole thing is canned by nature: seeded fleet run, ~3 minutes. Include
a 45-second "highlight" seed where a failure and a stuck agent break
through routine traffic.

## The 60-second demo story

"Close your eyes. [30s of channel traffic] — okay: which stream is stuck,
which one merged? You knew. Now the same 30 seconds as prose. [chaos] —
that's why agent voice failed until now. Brevity codes are how eight agents
fit in one ear."
