# The 60-second demo

Setup (before the clock starts): `npm start`, open http://localhost:8765 in
Chrome, sound on. Scenario **Highlight · 45 s**, mode **Brevity**. Everything
is canned — no mic, nothing to type.

**0:00 — say this:** "Six agents, one radio channel, two-second calls. Close
your eyes."

**0:03 — press ▶ Open channel.** Let it run. The judge hears: dispatches,
"tests red" on stream two, "PR open," a squelch-clicked rhythm of routine
traffic — then the alert tone and *"PAN PAN, ship driver, stream three,
stuck, no reviewer, need human"* cutting through — then "review two, taking
stream three," "merged, off channel," "main green."

**0:40 — ask:** "Which stream got stuck? Which streams merged?" They know:
stream three was stuck, everything merged. Eyes closed. Then point at the
fleet board — it says what their ears already told them.

**0:45 — flip the toggle to Prose, press ▶ again, say:** "Same 45 seconds,
same events, as sentences — this is every agent-voice demo you've seen."
Fifteen seconds is plenty: two long monologues air, the ticker prints
"7 routine calls dropped — frequency saturated," and the audio falls half a
minute behind the board. Stop it mid-sentence.

**1:00 — close:** "Prose can't fit eight agents in one ear. Brevity codes
can. The grammar, the simulator, and the drop policy are 76 tests of pure
code — the browser only supplies the voice."

## Would someone pay?

**Named buyer:** the operator running a multi-agent dev fleet — concretely,
me, and every Claude Code / Codex / Cursor power user now running 3–10
parallel driver sessions and alt-tabbing dashboards to see which one is
stuck. The eyes-busy monitoring problem is already paid for everywhere else
it occurs: air traffic uses brevity phraseology, dispatch centers pay for
radio consoles, ops teams pay PagerDuty specifically to interrupt them with
prioritized signals instead of logs. Fleet-of-agents tooling is the same
problem one abstraction up, and the vendors themselves (Anthropic's own
fleet UIs, Cursor's background agents) are racing to solve "which of my N
agents needs me" — with more screens. Audio is the only channel that works
while your eyes are on the code you're writing. This is a feature those
products would buy, and the wedge is exactly what's in this repo: a grammar
plus a priority policy, not a platform.

**Voice-necessity:** the entire bet is a property only audio has — ambient,
eyes-free, pre-attentive interruption. On a screen this would be another
dashboard you forget to look at. The prose toggle exists to prove the point
in reverse: voice without a brevity grammar really is worthless, which is
why this hasn't been built.

**Honest limits:** the fleet is a fixture. The upgrade path (real driver
events in, same grammar out) changes the input adapter and nothing else —
grammar, channel policy, and tests carry over untouched.
