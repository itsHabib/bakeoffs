# Entry 1: hack-atc — radio phraseology trainer (VFR pattern wedge)

Read ~/dev/interject/docs/hackathon/README.md first — house rules and judging
apply verbatim. Repo: `~/dev/hack-atc`. You never see the other three entries.

## The bet

Student pilots dread the radio. They rehearse calls alone in parked cars.
Existing trainers are scripted and rigid (pre-LLM) or use live humans
(scheduling + stage fright). An always-available AI tower that runs you
through pattern work and GRADES YOUR READBACK DETERMINISTICALLY is the wedge.

## What to build

A browser app that plays a tower controller for ONE scenario: VFR pattern
work at a single towered field (taxi → takeoff clearance → pattern → landing
clearance → taxi back).

- Controller speaks via `speechSynthesis` (rate up, slight radio-clipped
  phrasing; a `[static]` chirp is charm, not required).
- Student replies by mic (Web Speech) or by typing — typing is the canned
  path and MUST work fully.
- **The grader is the product.** Encode the expected readback for each
  clearance as a grammar/template in code: required elements (callsign,
  runway, instruction), allowed orderings, allowed shorthand. Grade every
  student response: which required elements present, which missing, which
  WRONG (e.g. read back runway 27 when cleared 9). Show the readback
  scored element-by-element, instantly, on screen.
- A session transcript at the end: every exchange, every score, the three
  worst readbacks flagged with what the correct call was.
- Phraseology source: FAA AIM chapter 4 conventions from your own knowledge
  is fine for a demo — mark any call you're unsure of with a TODO rather
  than inventing confidently.

## What NOT to build

No LLM in the loop at all if you can avoid it — the controller's lines can be
templated per scenario state; that is a FEATURE (zero cost, zero latency,
perfectly repeatable). No multiple airports, no IFR, no emergencies, no
voices for other traffic (one line of "traffic on final" flavor text max).
No accounts, no persistence beyond the session.

## Canned demo (required)

A "fly the script" button that runs a whole pattern circuit with a scripted
student who makes 2-3 realistic readback errors (wrong runway, dropped
callsign, "roger" instead of a full readback) so the judge watches the grader
catch them without touching a mic.

## The 60-second demo story

"This is a tower controller you can practice with at 11pm. It just cleared
me to land runway 9 — I read back runway 27 — watch: caught it, element by
element, before I finished the sentence. Every student pilot pays for
ground school; none of them have this."
