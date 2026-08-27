# Keepsake

Keepsake is a browser voice interviewer that helps a parent or grandparent tell one complete family story, then arranges the teller's exact words into a chapter.

The demo covers one topic: **How you two met**. Its deterministic story spine tracks six beats in code: when, where, who, the turn, the feeling, and one vivid sensory detail. The Realtime model supplies warm ears and a voice, but it never decides whether a beat is covered and never writes the chapter.

## Run it

```sh
npm start
```

Open <http://127.0.0.1:4312>. For judging, press **Start a live interview**,
allow microphone access, and use the one-minute exchange in `DEMO.md`.

`OPENAI_API_KEY` must be available to the server (source `~/.zshrc` if needed)
for live voice. The browser holds the WebRTC
connection. The Node server only mints a short-lived Realtime client secret;
the standard API key is never placed in client code, page source, local
storage, logs, or this repository.

**Replay the 2-minute story** is the no-key fallback and deterministic harness.
It needs no microphone, API key, account, or network request, and drives the
same coverage tracker and chapter assembler used by live transcripts.

## Verify it

```sh
npm run check
```

The policy tests cover every story-beat rule, monotonic turn-by-turn coverage, next-question selection, spine-order assembly, tangent exclusion, and the provenance invariant: **every sentence in the verbatim chapter must already exist in a teller turn**.

## Why voice is necessary

The obvious alternative is another emailed prompt or blank writing box. That asks the family member to compose, edit, and sustain momentum alone. Voice lets someone simply remember out loud while an interviewer notices the missing shape and asks one small follow-up. The product is the conversation, not speech-to-text attached to a form.

## Buyer and price

- Buyer: an adult child buying for a parent or grandparent
- Store category: Lifestyle / Family
- First paid version: **$79/year**, or **$25 per printed-ready book export**
- Existing spending signal: StoryWorth's roughly $99/year prompt-based product and Remento's voice-memory product show that families already buy help capturing life stories.

## Deliberate limits

- One topic spine and one tab-lived session
- No accounts, persistence, photos, exports, email, sharing, or multi-session memory
- Coverage rules use intentionally small keyword/entity heuristics. They are deterministic and visible, but accents, paraphrases, and unfamiliar place descriptions can be missed. Those heuristics are marked by their compact rule table in `lib/story-core.js`; a real product would validate them against consented transcripts.
- Chapter assembly retains only sentences that map to a covered beat. It does not model-polish or paraphrase them.
- Live Realtime depends on microphone permission, network quality, a configured server key, and availability of `gpt-realtime-2.1` plus `gpt-4o-mini-transcribe`.

See [DEMO.md](./DEMO.md) for the exact one-minute walkthrough.
