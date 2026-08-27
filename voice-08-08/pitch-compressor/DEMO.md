# 60-second live demo

## Before the clock

Run:

```sh
cd the pitch-compressor entry && npm start
```

Open <http://127.0.0.1:4173> in current Chrome. Confirm the terminal says
`Realtime listener: configured`. Allow microphone access when Chrome asks.

## Walkthrough

**0–8 seconds**

> Twenty seconds, one conclusion. Pitch Compressor makes you deliver the part
> that matters before your listener's time runs out.

Click **Connect live listener**. Wait for **Live listener ready**, then click
**Start 20-second rep**.

**8–28 seconds — say this live, naturally, without racing**

> Um, I think we basically help sales teams, like, see renewals that might be
> drifting before managers notice. Maybe the dashboard hopefully gives them
> enough time to act, and what I would really like you to do is—

Keep speaking until zero; let the wall interrupt. Do not click **Stop early**.

**28–39 seconds — listen**

The OpenAI Realtime listener asks one skeptical spoken question based on the
frozen transcript and deterministic gaps. Let it finish. Its wording is live,
not scripted; it cannot change the score.

**39–49 seconds**

> The wall cut the ask. The scorecard landed instantly: fillers, hedges, pace,
> ask, repetition, and exactly what survived. Every point is code, not AI
> judgment. The voice only made the pressure—and the unanswered question—real.

Point at **Wall ate** and the six rule lines.

**49–60 seconds**

> Founders pay coaches hundreds an hour to do this with a stopwatch. This is
> for founders and sales reps, in Productivity and Business: $9.99 a month for
> unlimited reps and history.

## Canned fallback and deterministic product moment

Click **Replay 3 canned reps** once. It needs no microphone, no API key, and
runs all three fixture transcripts with simulated timing through the same
`gradeRep` function as live speech:

> Rep one rambles into the wall: four hedges, no ask, 30. Rep two is shorter
> but still soft: two hedges, no ask, 56. Rep three finishes in 19.2 seconds
> with one number and one ask: 84. The 30 → 56 → 84 trend and rep-to-rep cuts
> are the product moment.

## Commercial assertion

- **Buyer:** founders preparing investor calls and sales reps preparing
  discovery calls.
- **Store category:** Productivity / Business.
- **First paid version:** $9.99/month for unlimited reps plus history.
- **Why voice is necessary:** the scarce resource is listener time. Typing
  removes pacing, fillers, delivery pressure, and the hard spoken cutoff.
