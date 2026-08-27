# Entry 1: hack2-story-capture — interview grandma, get chapters

Read ~/dev/interject/docs/hackathon-codex/README.md first — house rules and
judging apply verbatim. Repo: `~/dev/hack2-story-capture`. You never see the
other 3 entries.

## The bet

People will talk for twenty minutes about things they would never write three
paragraphs on. StoryWorth (~$99/yr) and Remento prove families pay for
exactly this — but both work by EMAILED WRITTEN PROMPTS or one-shot voice
memos, no conversation. A voice interviewer that actually listens, follows
up on the good thread ("wait — what was the boat called?"), and turns the
session into readable chapters is the product those companies would build if
they started today. Buyer: adult children buying for a parent or grandparent.
Store category: Lifestyle / Family. First paid version: $79/yr or $25 per
printed-ready book export.

## What to build

A browser app that runs one interview session and produces one chapter.

- **The deterministic core is the interview spine + the assembler.** A
  session covers ONE life topic (e.g. "how you two met") from a coverage
  checklist in code: the beats a complete chapter needs (time, place, people,
  the complication, the feeling, one sensory detail). Track coverage
  deterministically as transcript accumulates — keyword/entity rules are
  fine, TODO-mark shaky heuristics. Uncovered beats become the follow-up
  queue the interviewer draws from.
- **Realtime voice is the interviewer** — warm, brief, asks ONE question at a
  time, follows up on what was actually said. Browser WebRTC + server-minted
  ephemeral token (see README). The model never decides coverage; it is
  handed the next beat to chase.
- **Chapter assembly is code, not model:** the transcript, cut into the
  covered beats in spine order, speaker labels stripped, teller's words kept
  verbatim with light join-stitching. Provenance rule: every sentence in the
  chapter must exist in the transcript — assert it in a test. (A model may
  OPTIONALLY smooth transitions in a clearly-labeled "polished" variant, but
  the verbatim chapter is the deliverable.)
- End screen: the chapter, the coverage checklist all green, and the
  follow-up questions it still wants — that trio is the demo moment.

## What NOT to build

No accounts, no multi-session memory, no photo uploads, no printing/export
pipelines, no family sharing, no more than ONE topic spine. No emailing. No
long-term storage — session state may die with the tab.

## Canned demo (required)

A bundled fixture transcript of a real-feeling interview (write one: a
grandmother telling how she met her husband, ~2 minutes of turns, including
one tangent and one beat that never gets covered). A "replay session" button
drives it through the live coverage tracker turn by turn, then assembles the
chapter — judge watches beats light up and the chapter appear, no mic, no key.

## The 60-second demo story

"My kids bought my mom StoryWorth. She answered two email prompts and quit.
Watch this instead — [replay runs] — it's tracking what a complete story
needs, see the beats filling in… it just noticed she never said WHERE, so
that's its next question. And here's the chapter — every sentence is her own
words, that's tested, not promised. $99 a year for the email version has
been selling for a decade. This is the conversation version."
