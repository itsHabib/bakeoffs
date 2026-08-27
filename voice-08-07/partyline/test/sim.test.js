import { test } from 'node:test';
import assert from 'node:assert/strict';
import { generateTimeline, SCENARIOS, AGENTS } from '../src/sim.js';
import { EVENT_TYPES } from '../src/grammar.js';

const types = (tl, type) => tl.events.filter((e) => e.type === type);

test('timelines are deterministic: same scenario, identical events', () => {
  for (const name of Object.keys(SCENARIOS)) {
    assert.deepEqual(generateTimeline(name), generateTimeline(name));
  }
});

test('events are time-sorted and inside the scenario duration', () => {
  for (const name of Object.keys(SCENARIOS)) {
    const tl = generateTimeline(name);
    let prev = -1;
    for (const e of tl.events) {
      assert.ok(e.t >= 0 && e.t <= tl.duration, `event at ${e.t} outside 0..${tl.duration}`);
      assert.ok(e.t >= prev, 'events out of order');
      prev = e.t;
    }
  }
});

test('the sim only emits event types the grammar can speak', () => {
  const known = new Set(EVENT_TYPES);
  for (const name of Object.keys(SCENARIOS)) {
    for (const e of generateTimeline(name).events) {
      assert.ok(known.has(e.type), `grammar has no template for "${e.type}"`);
    }
  }
});

test('full run: all three drivers merge, and every agent gets on the air', () => {
  const tl = generateTimeline('full');
  assert.equal(types(tl, 'merged').length, 3);
  const speakers = new Set(tl.events.map((e) => e.agentId));
  for (const a of AGENTS) assert.ok(speakers.has(a.id), `${a.callsign} never transmits`);
});

test('full run: contains the fail, the stuck agent, and the CI red', () => {
  const tl = generateTimeline('full');
  assert.ok(types(tl, 'tests_red').length >= 1);
  assert.equal(types(tl, 'stuck').length, 1);
  assert.equal(types(tl, 'pipeline_red').length, 1);
});

test('highlight: a failure and a stuck agent break through inside 45 seconds', () => {
  const tl = generateTimeline('highlight');
  assert.ok(tl.duration <= 45);
  assert.equal(types(tl, 'stuck').length, 1);
  assert.ok(types(tl, 'tests_red').length >= 1);
});

test('highlight: at least one stream merges by the 30-second mark (the quiz answer)', () => {
  const tl = generateTimeline('highlight');
  assert.ok(types(tl, 'merged').some((e) => e.t <= 30));
});

test('highlight: dense enough to saturate a prose channel', () => {
  const tl = generateTimeline('highlight');
  assert.ok(tl.events.length >= 15, `only ${tl.events.length} events`);
});
