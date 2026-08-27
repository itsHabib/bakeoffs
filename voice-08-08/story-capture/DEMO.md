# 60-second live demo

This is a human exchange, not a click-through. Start the server with the command
in `README.md`, open the page, press **Start a live interview**, allow the
microphone, and answer Keepsake with the lines below. Pause after each answer so
semantic turn detection can hand the next uncovered beat to the interviewer.
The interviewer chooses its own brief wording, so its questions will not be
word-for-word identical.

## The exchange

**Keepsake:** asks when the meeting happened.

**You:** “It was the summer of 1958, just after my nineteenth birthday.”

**Keepsake:** follows up for the missing place.

**You:** “We were at a dance hall near my sister’s apartment.”

**Keepsake:** follows up for the people involved.

**You:** “My cousin Ruth introduced me to Daniel Mercer.”

**Keepsake:** follows up for the complication or turn.

**You:** “The problem was Ruth had promised him to my roommate, but he kept asking me questions instead.”

**Keepsake:** follows up for the feeling.

**You:** “I was so nervous that my hands shook.”

**Keepsake:** follows up for one sensory detail.

**You:** “His shirt smelled faintly of cedar, and the radio sounded scratchy behind him.”

Keepsake should thank you and say the chapter is ready. Press **Finish & make
the chapter**. Point to all six green beats, the empty essential-follow-up
queue, and the green provenance badge. Every chapter sentence is copied from a
teller turn and arranged by the deterministic spine; that is asserted in the
tests, not entrusted to the model.

## The one-breath pitch

“My kids bought my mom StoryWorth. She answered two email prompts and quit.
This listens and keeps the conversation moving. The buyer is an adult child
buying for a parent or grandparent; the store category is Lifestyle / Family;
the first paid version is **$79 per year** or **$25 per printed-ready book
export**. Families already pay about $99 a year for the prompt version. This is
the conversation version.”

## No-key fallback

If the network or microphone fails on judging day, press **Replay the 2-minute
story**. The fixture runs turn-by-turn through the real tracker and assembler,
leaves **Where** uncovered, shows its next follow-up, and produces the
provenance-safe chapter. Replay is the deterministic harness; live voice is the
product demo.
