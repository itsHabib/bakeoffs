# DEMO.md — the 60-second walkthrough

Setup (before the clock starts): `node server.mjs`, open
http://localhost:8329, sound on.

## The script

**0:00 — one line of framing.**
"Student pilots dread the radio more than the flying. They rehearse calls
alone in parked cars. This is a tower controller you can practice with at
11pm — and it grades your readback with *code*, not vibes."

**0:05 — press ▶ Fly the script.** (No mic. It runs itself.)

**0:05–0:20 — taxi + takeoff.** Ground issues the taxi clearance out loud.
Watch the right panel: as the student's readback appears, elements flip
green one by one — runway, route, hold-short — and **Callsign goes amber:
MISSING**. "Four required elements. It caught the dropped callsign
instantly. Next call is clean — all green, 100%."

**0:20–0:30 — the 'roger' trap.** Tower gives a sequencing instruction; the
student answers "Roger." Every element flags missing and the panel says
*"Roger" is not a readback*. "This is the habit every CFI beats out of
students. The app does it for free."

**0:35 — the money moment.** Tower: "runway 9, cleared to land." Student
reads back **runway 27**. The runway element flips red — ✗ WRONG, heard
"runway 27" — *before the sentence is finished*. "That readback error, in
the real world, is how runway incursions start. Caught it mid-sentence,
deterministically."

**0:45–0:60 — the debrief.** Session ends: overall accuracy, per-call bars,
and the three worst readbacks each shown with what the correct call was.
"Every exchange graded, every miss explained. And if you have Chrome and a
mic —" (press 🎙, read one call back live) "— same grader, same speed."

## Would someone pay?

**Buyer: student pilots (and their CFIs).** ~60k active student pilot
certificates in the US at any time; a private certificate runs $12–17k, and
radio anxiety is one of the most-cited reasons students stall or quit.
Money already moves exactly here: PlaneEnglish ARSim charges ~$60/yr for
scripted radio drills (no free-form grading); PilotEdge charges $20–35/mo
for live humans on scheduled sessions; Sporty's/King sell radio courses that
are pure video. An always-available tower that *grades your actual readback
element-by-element* sits in the gap between the script app and the live
human, at the script app's price point.

**Why voice.** Radio work is a spoken-cadence skill — the failure mode being
trained (dropping an element while your mouth is busy, transposing digits
under time pressure) does not exist in a text box. The controller speaking
at ATC pace and grading you *while you answer* is the product; text mode is
just the accessibility/demo path.

**Deterministic share.** The entire correctness path — normalization,
element matching, wrong-value detection, scoring, the debrief — is
`public/grader.js` + `public/scenario.js`, covered by 43 table tests
(`node --test`). The model count in this repo is zero.
