# The 60-second demo

Setup (before the clock starts): `ollama serve` running, then
`go run -buildvcs=false .` in this directory, open http://127.0.0.1:8731.
Sound on.

## Script

**[0:00]** Press **▶ Misconception**. Say:

> "This is a math tutor that's only allowed to interrupt three times per
> session. Rationing is the pedagogy — watch it spend."

**[0:10]** The student botches 3×3 while expanding (x+3)². Point at the
right panel:

> "There's a real mistake — and the tutor is sitting on it. See: *held back:
> 'What is three times three?' — below the floor, never worth a turn.*"

**[0:20]** The green line appears:

> "…and she just caught it herself. That's the point. An app that buzzes
> every slip trains kids to wait for the buzzer."

**[0:40]** A sign error in the factoring goes by; the tutor holds again.
Say nothing — let the silence do it.

**[1:00]** The student says (x+y)² = x²+y² and the tutor finally speaks,
out loud: *"Does squaring a binomial always distribute to both terms?"*

> "THAT one it paid for — a wrong rule she'd carry into every problem this
> month. And notice it asked a question. It's not allowed to give answers —
> that's enforced in code, not in the prompt."

**[1:20]** The session report pops up on its own: *1 of 3 turns spent*,
every hold timestamped with its reason, the slips marked
"✓ student caught it — the hold paid off."

> "This report is the product. A parent reads what it *didn't* say and gets
> it instantly. Tutoring apps interrupt forty times an hour; good teachers
> don't."

## Would someone pay?

**The buyer is a parent already paying for restraint.** US parents spend
$10B+/yr on tutoring; Kumon alone has ~4M students at $150–200/mo, and what
those franchises actually sell is a human who makes the kid do the work.
Meanwhile Photomath (100M+ downloads) and every AI tutor since sell the
opposite — instant answers — and parents know it: "the app does my kid's
homework" is the #1 objection to AI in homework help. This is the first
homework tool whose pitch to the parent is *what it refused to say*, with a
receipt. The end-of-session silence report is the paywall artifact: free to
practice, pay to see the ledger. Secondary buyer: tutoring chains
white-labeling the report as proof-of-pedagogy.

**Why voice.** Thinking out loud is how tutors diagnose — typed working is
already self-censored. The mistakes here are caught *mid-stream*, in the
seconds between a student saying a wrong thing and building on it, and the
correction lands as a spoken question while her pencil is still moving. A
text version is a worse Photomath; the audio version is a patient human.

## Fallbacks

- Rehearsal stalls / Ollama down: `ollama serve`, then reset session and
  re-press the button. The classifier warms on first call (~2s each after).
- Want the other stories: **▶ Self-caught slip** (~35s, tutor spends
  nothing, report shows the hold paying off) and **▶ Stuck** (~40s, nudge
  earned by persistence — first stuck claim is discounted by the depth
  guard, visible in the report).
- No sound: the spoken line also appears as the top card in "Spent & held".
