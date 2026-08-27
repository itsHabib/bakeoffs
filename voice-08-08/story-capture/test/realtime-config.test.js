import assert from "node:assert/strict";
import test from "node:test";

import {
  REALTIME_MODEL,
  createRealtimeSessionPayload,
  interviewerInstruction,
} from "../lib/realtime-config.js";

test("Realtime session uses browser-safe voice configuration", () => {
  const payload = createRealtimeSessionPayload();
  assert.equal(payload.model, REALTIME_MODEL);
  assert.equal(payload.model, "gpt-realtime-2.1");
  assert.equal(payload.audio.input.transcription.model, "gpt-4o-mini-transcribe");
  assert.equal(payload.audio.input.turn_detection.create_response, false);
  assert.equal(payload.audio.input.turn_detection.interrupt_response, true);
  assert.deepEqual(payload.output_modalities, ["audio"]);
  assert.ok(!JSON.stringify(payload).includes("OPENAI_API_KEY"));
});

test("the deterministic next beat is handed to the interviewer", () => {
  const instruction = interviewerInstruction({
    id: "place",
    label: "Where",
    prompt: "Find the physical place.",
  }, "It was 1958.");
  assert.match(instruction, /next uncovered beat is "Where"/);
  assert.match(instruction, /exactly one short follow-up/i);
  assert.match(instruction, /It was 1958/);
});

test("a complete spine produces a closing instruction, not another question", () => {
  const instruction = interviewerInstruction(null, "The radio was scratchy.");
  assert.match(instruction, /chapter is ready/i);
  assert.match(instruction, /Do not ask another question/);
});
