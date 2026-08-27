import assert from "node:assert/strict";
import test from "node:test";
import {
  analyzeAnswer,
  CANNED_FIXTURE,
  InterviewMachine,
  QUESTION_BANK,
  runCannedReplay,
  tokenize,
} from "../lib/interview-core.js";

const leadership = QUESTION_BANK.find((question) => question.id === "leadership");
const conflict = QUESTION_BANK.find((question) => question.id === "conflict");
const failure = QUESTION_BANK.find((question) => question.id === "failure");

test("question bank has nine unique, bounded behavioral questions", () => {
  assert.equal(QUESTION_BANK.length, 9);
  assert.equal(new Set(QUESTION_BANK.map((question) => question.id)).size, 9);
  for (const question of QUESTION_BANK) {
    assert.equal(question.budgetSeconds, 45);
    assert.match(question.prompt, /[?.]$/);
    assert.ok(question.keywords.length >= 4);
    assert.deepEqual(Object.keys(question.followUps).sort(), ["action", "context", "dodge", "result"]);
  }
});

test("tokenizer normalizes punctuation and whitespace", () => {
  assert.deepEqual(tokenize("  I BUILT—then, tested!  "), ["i", "built", "then", "tested"]);
});

test("solid answer detects all STAR beats without a dodge", () => {
  const metrics = analyzeAnswer(
    leadership,
    "The situation was an uncertain project. My task was a safe launch. I personally created the plan. As a result, we reduced failures by 40%.",
    32,
  );
  assert.deepEqual(metrics.star, {situation: true, task: true, action: true, result: true});
  assert.equal(metrics.dodge, false);
  assert.equal(metrics.withinBudget, true);
  assert.equal(metrics.score, 100);
});

test("dodge uses zero question-content overlap, not model judgment", () => {
  const metrics = analyzeAnswer(
    failure,
    "I am excited about this role because the company mission fits my background and I enjoy collaborative teams.",
    18,
  );
  assert.equal(metrics.overlapTerms.length, 0);
  assert.equal(metrics.dodge, true);
  assert.equal(metrics.score, 20);
});

test("a short silence is not mislabeled as a confident dodge", () => {
  const metrics = analyzeAnswer(failure, "I don't know.", 2);
  assert.equal(metrics.dodge, false);
  assert.equal(metrics.words, 3);
});

test("filler phrases are counted once each and normalized by words", () => {
  const metrics = analyzeAnswer(conflict, "Um, I had a conflict and, you know, I basically asked my coworker to meet.", 12);
  assert.equal(metrics.fillerCount, 3);
  assert.equal(metrics.words, 15);
  assert.equal(metrics.fillerDensity, 0.2);
});

test("time budget flips only after 45 seconds", () => {
  assert.equal(analyzeAnswer(conflict, "I resolved a conflict with a coworker.", 45).withinBudget, true);
  assert.equal(analyzeAnswer(conflict, "I resolved a conflict with a coworker.", 45.1).withinBudget, false);
});

test("machine refuses invalid transitions", () => {
  const machine = new InterviewMachine({questionIds: ["leadership"]});
  assert.throws(() => machine.beginAnswer(), /Cannot begin/);
  machine.start();
  assert.throws(() => machine.start(), /already started/);
  machine.beginAnswer(0);
  assert.throws(() => machine.beginAnswer(), /Cannot begin/);
});

test("time interruption has first follow-up priority", () => {
  const machine = new InterviewMachine({questionIds: ["conflict"]});
  machine.start();
  machine.beginAnswer(0);
  const decision = machine.submitAnswer({
    transcript: "The situation was a conflict with a coworker and my task was to resolve the disagreement.",
    durationSeconds: 52,
    wasInterrupted: true,
  });
  assert.equal(decision.type, "followup");
  assert.equal(decision.followUp.reason, "time");
  assert.match(decision.followUp.prompt, /stop you there/i);
});

test("dodge selects a question-specific press", () => {
  const machine = new InterviewMachine({questionIds: ["failure"]});
  machine.start();
  machine.beginAnswer(0);
  const decision = machine.submitAnswer({
    transcript: "Your mission is inspiring and the role seems to fit my background very well.",
    durationSeconds: 12,
  });
  assert.equal(decision.type, "followup");
  assert.equal(decision.followUp.reason, "dodge");
  assert.match(decision.followUp.prompt, /one real failure/i);
});

test("missing personal action is pressed when relevance is present", () => {
  const machine = new InterviewMachine({questionIds: ["leadership"]});
  machine.start();
  machine.beginAnswer(0);
  const decision = machine.submitAnswer({
    transcript: "The situation was an uncertain project. My task was to help the team, and the project eventually shipped.",
    durationSeconds: 24,
  });
  assert.equal(decision.followUp.reason, "action");
  assert.match(decision.followUp.prompt, /personally/i);
});

test("machine allows at most two follow-ups then advances", () => {
  const machine = new InterviewMachine({questionIds: ["failure", "leadership"]});
  machine.start();
  for (let index = 0; index < 3; index += 1) {
    machine.beginAnswer(0);
    const decision = machine.submitAnswer({
      transcript: "This role and company mission fit my background and collaborative style extremely well.",
      durationSeconds: 10,
    });
    if (index < 2) assert.equal(decision.type, "followup");
    else {
      assert.equal(decision.type, "question");
      assert.equal(decision.question.id, "leadership");
    }
  }
  assert.equal(machine.records[0].followUps.length, 2);
});

test("follow-up transcript is combined into the final question scorecard", () => {
  const machine = new InterviewMachine({questionIds: ["leadership"]});
  machine.start();
  machine.beginAnswer(0);
  let decision = machine.submitAnswer({
    transcript: "The situation was an uncertain project and my task was to create a launch plan.",
    durationSeconds: 15,
  });
  assert.equal(decision.followUp.reason, "action");
  machine.beginAnswer(0);
  decision = machine.submitAnswer({
    transcript: "I personally created the checklist. As a result, we reduced failures by 40%.",
    durationSeconds: 10,
  });
  assert.equal(decision.type, "debrief");
  assert.equal(decision.debrief.scorecards[0].metrics.starCount, 4);
  assert.match(decision.debrief.scorecards[0].transcript, /checklist/);
});

test("canned replay uses exactly three fixtures through the real machine", () => {
  const replay = runCannedReplay();
  assert.equal(CANNED_FIXTURE.length, 3);
  assert.equal(replay.debrief.scorecards.length, 3);
  assert.equal(replay.events.filter((event) => event.type === "question").length, 3);
  assert.equal(replay.events.at(-1).type, "debrief");
});

test("canned replay contains solid, interrupted, and dodged primary answers", () => {
  const replay = runCannedReplay();
  const [solid, ramble, dodge] = replay.debrief.scorecards;
  assert.equal(solid.metrics.starCount, 4);
  assert.equal(ramble.followUps[0].reason, "time");
  assert.ok(ramble.metrics.durationSeconds > ramble.metrics.budgetSeconds);
  assert.equal(replay.debrief.worstDodge.quote, CANNED_FIXTURE[2].transcript);
  assert.equal(dodge.followUps[0].reason, "dodge");
});

test("debrief picks exactly two lowest-scoring questions to rerun", () => {
  const {debrief} = runCannedReplay();
  assert.equal(debrief.tomorrow.length, 2);
  assert.ok(debrief.tomorrow[0].score <= debrief.tomorrow[1].score);
  assert.ok(debrief.tomorrow.every((item) => debrief.scorecards.some((card) => card.questionId === item.questionId)));
});
