import test from "node:test";
import assert from "node:assert/strict";

import {
  WALL_MS,
  applyWall,
  countPhrases,
  detectAsk,
  diffCuts,
  gradeRep,
  ngramOverlap,
  words,
} from "./grader.mjs";

test("word normalization is punctuation-safe", () => {
  assert.deepEqual(words("Book a 30-minute pilot — don't wait."), ["book", "a", "30-minute", "pilot", "don't", "wait"]);
});

test("phrase counting handles single and multi-word rules", () => {
  const cases = [
    ["Um, this is basically ready", ["um", "basically"], 2],
    ["Sort of useful, sort of vague", ["sort of"], 2],
    ["A likely outcome is not the filler like", ["like"], 1],
    ["No verbal lint here", ["um", "like", "sort of"], 0],
  ];
  for (const [text, phrases, expected] of cases) {
    assert.equal(countPhrases(text, phrases), expected, text);
  }
});

test("hard wall table includes an intentional mid-word cut", () => {
  const cases = [
    { durationMs: 19_200, hit: false, eaten: "Nothing — you stopped before the wall." },
    { durationMs: WALL_MS, hit: true, eaten: "Audio after 20.0s" },
    { durationMs: 26_000, hit: true, eaten: null },
  ];
  for (const example of cases) {
    const result = applyWall("alpha bravo charlie delta echo foxtrot golf hotel", example.durationMs, example.durationMs === WALL_MS);
    assert.equal(result.wallHit, example.hit);
    if (example.eaten) assert.match(result.eaten, new RegExp(example.eaten.replace(".", "\\.")));
    if (example.durationMs > WALL_MS) {
      assert.match(result.survived, /—$/);
      assert.doesNotMatch(result.survived.slice(0, -1), /\s$/);
      assert.notEqual(result.eaten[0], " ");
    }
  }
});

test("ask heuristic only considers an imperative or number in final quarter", () => {
  const cases = [
    ["We saved 12 renewals but the conclusion stays vague and drifts away", false],
    ["We stop churn for growing teams book a pilot", true],
    ["The pilot proves our retention claim across 12 accounts", true],
    ["We stop churn for growing teams and it works", false],
  ];
  for (const [text, expected] of cases) assert.equal(detectAsk(text).present, expected, text);
});

test("bigram overlap is based on current content", () => {
  assert.equal(ngramOverlap("alpha beta gamma", "alpha beta delta"), 0.5);
  assert.equal(ngramOverlap("alpha beta", ""), 0);
});

test("diff reports words removed between reps in reading order", () => {
  assert.deepEqual(diffCuts("we really help sales teams today", "we help teams"), ["really", "sales", "today"]);
});

test("grade table exercises every rule and clamps the score", () => {
  const cases = [
    {
      name: "clean timed ask",
      input: {
        transcript: "We flag renewal risk before accounts drift so managers act today. Our pilot saved three customers last quarter. The evidence is clear. Book a pilot this week with your 10 highest-risk accounts.",
        durationMs: 19_200,
      },
      expected: { wall: 0, fillers: 0, pace: 0, hedges: 0, ask: 0, repeat: 0, score: 100 },
    },
    {
      name: "wall filler hedge missing ask and slow pace",
      input: {
        transcript: "Um I think basically maybe hopefully like sort of this pitch keeps wandering without a concrete request and eventually it falls past the deadline into words nobody hears",
        durationMs: 31_000,
      },
      expected: { wall: 18, fillers: 16, pace: 12, hedges: 12, ask: 20, repeat: 0, score: 22 },
    },
    {
      name: "repeat penalty is deterministic",
      input: {
        transcript: "We flag renewal risk before accounts drift so managers act today. Run a pilot with five accounts this week. Book now.",
        durationMs: 12_000,
        previousTranscript: "We flag renewal risk before accounts drift so managers act today. Run a pilot with five accounts next week. Book now.",
      },
      expected: { repeat: 16, score: 84 },
    },
  ];

  for (const example of cases) {
    const result = gradeRep(example.input);
    for (const [key, expected] of Object.entries(example.expected)) {
      if (key === "score") assert.equal(result.score, expected, example.name);
      else assert.equal(result.penalties[key], expected, `${example.name}: ${key}`);
    }
  }

  const bottom = gradeRep({
    transcript: "Um like basically sort of I think maybe hopefully um like basically sort of I think maybe hopefully um like basically sort of I think maybe hopefully um like basically sort of I think maybe hopefully",
    durationMs: 40_000,
    previousTranscript: "Um like basically sort of I think maybe hopefully um like basically sort of I think maybe hopefully um like basically sort of I think maybe hopefully um like basically sort of I think maybe hopefully",
  });
  assert.equal(bottom.score, 0);
});

test("invalid grade inputs fail loudly", () => {
  assert.throws(() => gradeRep({ transcript: null, durationMs: 10 }), /transcript/);
  assert.throws(() => gradeRep({ transcript: "words", durationMs: 0 }), /durationMs/);
});
