import assert from "node:assert/strict";
import test from "node:test";

import { FIXTURE } from "../public/fixture.js";
import {
  STORY_SPINE,
  assembleChapter,
  matchingBeatIDs,
  trackCoverage,
  verifyProvenance,
} from "../lib/story-core.js";

const policyCases = [
  ["time", "It was the summer of 1958.", true],
  ["time", "He wore a narrow tie.", false],
  ["place", "We met at the train station.", true],
  ["place", "I cannot remember the surroundings.", false],
  ["people", "My cousin introduced us.", true],
  ["people", "There was a blue ribbon.", false],
  ["complication", "The problem was that I had the wrong ticket.", true],
  ["complication", "He smiled at me.", false],
  ["feeling", "I was so nervous.", true],
  ["feeling", "I kept a biscuit tin.", false],
  ["sensory", "The radio sounded scratchy.", true],
  ["sensory", "He asked me a question.", false],
];

for (const [beat, sentence, expected] of policyCases) {
  test(`${beat} policy ${expected ? "accepts" : "rejects"}: ${sentence}`, () => {
    assert.equal(matchingBeatIDs(sentence).includes(beat), expected);
  });
}

test("the fixture covers five beats and leaves place for the next follow-up", () => {
  const coverage = trackCoverage(FIXTURE);
  assert.equal(coverage.totalCount, STORY_SPINE.length);
  assert.equal(coverage.coveredCount, 5);
  assert.deepEqual(coverage.uncovered.map((beat) => beat.id), ["place"]);
  assert.equal(coverage.nextBeat.id, "place");
  assert.match(coverage.followUps[0], /Where/);
});

test("coverage accumulates turn by turn without asking the model to judge", () => {
  const partial = [];
  const counts = [];
  for (const turn of FIXTURE) {
    partial.push(turn);
    counts.push(trackCoverage(partial).coveredCount);
  }
  assert.deepEqual(counts, [...counts].sort((a, b) => a - b));
  assert.ok(counts.includes(1));
  assert.ok(counts.includes(3));
  assert.equal(counts.at(-1), 5);
});

test("assembled chapter uses only teller sentences and excludes the tangent", () => {
  const chapter = assembleChapter(FIXTURE);
  assert.ok(chapter.text.includes("summer of 1958"));
  assert.ok(chapter.text.includes("shirt smelled faintly of cedar"));
  assert.ok(!chapter.text.includes("What made the meeting"));
  assert.ok(!chapter.text.includes("movie stub"));
  assert.ok(!chapter.text.includes("biscuit tin"));
});

test("every chapter sentence exists verbatim in teller transcript", () => {
  const chapter = assembleChapter(FIXTURE);
  const provenance = verifyProvenance(chapter, FIXTURE);
  assert.equal(provenance.valid, true);
  assert.deepEqual(provenance.missing, []);

  const tampered = { ...chapter, sentences: [...chapter.sentences, "Daniel wore a green jacket."] };
  assert.deepEqual(verifyProvenance(tampered, FIXTURE), {
    valid: false,
    missing: ["Daniel wore a green jacket."],
  });
});

test("chapter paragraphs follow spine order rather than transcript order", () => {
  const reversed = [...FIXTURE].reverse();
  const chapter = assembleChapter(reversed);
  assert.deepEqual(chapter.sections.map((section) => section.beat.id), [
    "time",
    "people",
    "complication",
    "feeling",
    "sensory",
  ]);
});
