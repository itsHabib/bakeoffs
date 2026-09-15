# Research brief: bounded-vocabulary voice products (ATC trainer is OUR pick — beat it)

Self-contained prompt for a research agent. No conversation context required.
Produce a written report; do NOT build anything, do NOT write a spec.

Your job is adversarial, in both directions:

1. **Try to kill our pick.** The ATC trainer below is the operator's current
   favorite. Attack it: find the LLM-era competitor that already shipped,
   the unit-economics problem, the reason students churn, the reason the
   obvious idea is still unbuilt (obvious-but-unbuilt usually means someone
   tried). If it survives your best shot, say so; if not, say what killed it.
2. **Field your own candidates.** Generate at least 5 product ideas of your
   own from the same underlying property, and champion the best 2 as if you
   had to pitch them against ours. Do not anchor on our list — the secondary
   sweeps in Q4 are OUR alternates, not yours. Yours should be different.

The underlying property (generate from THIS, not from "voice apps" broadly):
**a bounded vocabulary / constrained phraseology takes the LLM out of the
correctness path.** When the domain speaks a published grammar — ATC, maritime
VHF, military brevity, medical closed-loop readbacks, checklists, dispatch
codes — recognition can be checked deterministically, mishears become
detectable, readback loops catch the rest, and the model is reduced to ears
and mouth. Everyone else is making voice MORE natural; structured domains
went the other way a century ago because lives depended on it. Products that
exploit that inversion are the space. A good candidate has: an existing
population that ALREADY drills or is examined on the protocol, money already
moving (courses, exams, sims, fines for getting it wrong), and grading that
is genuinely deterministic rather than vibes.

## The hypothesis to test

Student pilots will pay a subscription for an AI controller they can practice
radio calls with — realistic phraseology, conversational timing, deterministic
grading of their readbacks, available 24/7 with no human on the other end
judging them.

Why we believe it: radio anxiety ("mic fright") is a named pain point in
flight training; existing paid products validate willingness to pay but
predate realtime voice models (scripted, rigid, or require scheduling live
humans). The technical edge: aviation phraseology is a *published bounded
grammar* (FAA AIM ch. 4-2, Pilot/Controller Glossary, JO 7110.65), so
correctness checking is deterministic code — the LLM only does ears and
mouth, never grading. A wrong readback is detectable without a model.

## Questions to answer

### Market (most important)
1. Competitors: PlaneEnglish ARSim, PilotEdge, VATSIM, plus anything newer
   (search for LLM-era entrants — "AI ATC simulator", "AI radio trainer",
   apps launched 2024–2026). For each: price, what it actually does, review
   sentiment — what do users complain about?
2. Market size: US student-pilot starts per year, active student certificates,
   flight-school count. What does a student already spend on training apps
   (ForeFlight, Sporty's, King Schools)?
3. Distribution: how do training apps reach students? CFI recommendation,
   flight-school bundles, Reddit r/flying, app stores? Is there a B2B motion
   (schools buy seats)?
4. Secondary sweeps (brief, one paragraph each):
   a. ICAO English proficiency prep for non-native pilots — mandatory exam,
      global market. Who serves it today, at what price?
   b. Medical closed-loop / SBAR radio sim for nursing & paramedic programs.
   c. Maritime VHF (SMCP) and military brevity training.

### Technical feasibility
5. Realtime voice API economics: OpenAI Realtime API current pricing — cost
   of a 30-minute practice session? Are there cheaper paths (speech-to-speech
   vs STT→LLM→TTS pipeline with a small model, given the domain is a bounded
   grammar)? Could recognition be LOCAL (grammar-constrained ASR, e.g.
   whisper + grammar post-match) with only the controller VOICE synthetic?
6. Latency realism: real ATC exchanges have sub-second turn-taking. What's
   achievable today browser-side (WebRTC to Realtime API) vs pipeline?
7. Content: is FAA phraseology (AIM, P/CG, JO 7110.65) public domain? (It
   should be — US government work.) What's the effort to encode, say, VFR
   pattern work at a towered field as a grammar + scenario graph?
8. Existing open assets: open-source ATC simulators, phraseology datasets,
   VATSIM's controller procedures docs — anything reusable?

### Product shape
9. What's the thinnest sellable wedge? Candidate: VFR tower pattern work
   only (taxi, takeoff, pattern, landing clearances) with readback grading —
   is that enough for a first paying user, per what students actually
   struggle with?
10. Regulatory: confirm nothing here needs FAA approval (it's practice, not
    loggable training). Any liability angle if a student learns wrong
    phraseology from us?

## Constraints on any eventual build (context, not tasks)

- Personal Mac, solo operator. Browser holds WebRTC; server only mints
  ephemeral tokens (pattern proven in the sibling `roll-call` project).
- House rule: correctness is computed, never model-judged. Grading must be
  deterministic against the grammar. Model = voice only.
- House rule: simplify until it hurts. Build-first, no spec until a demo
  earns one. Related POC: this archived Interject rationed-interruption listener.
  proved the sensor/policy split this would reuse.

## Deliverable

A report (markdown) with, in this order:

1. **The verdict on our pick** — go / no-go on the ATC wedge, with the
   strongest attack you found and whether it survived. Competitor table and
   cost-per-session math included.
2. **Your candidates** — 5+ ideas generated from the underlying property,
   each with one line on buyer, money-already-moving evidence, and why the
   grammar is genuinely bounded. Then your top 2, argued properly.
3. **The head-to-head** — your best idea vs our ATC trainer, compared on:
   proof someone pays today, size of wedge buildable by one person, and how
   much of the correctness path is truly deterministic. Declare a winner.
   You are allowed to conclude ours is better; you are not allowed to
   conclude it by default.
4. One-paragraph verdicts on OUR three secondary markets (Q4).
5. If your winner is buildable: the thinnest demo you'd cut first.

Rank confidence on each claim and cite sources. Flag anything that kills an
idea early and loudly. Disagreement with the operator's pick is a valid and
welcome outcome — a report that just agrees is a wasted run.
