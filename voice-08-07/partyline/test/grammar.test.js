import { test } from 'node:test';
import assert from 'node:assert/strict';
import {
  EVENT_TYPES,
  STATUS_WORDS,
  ROUTINE_WORD_CAP,
  PRIORITY_WORD_CAP,
  brevity,
  prose,
  wordCount,
  isPriority,
} from '../src/grammar.js';

// Worst-case-long sample data per event type (longest callsigns and fields
// the sim can produce), so the caps are tested at their tightest.
const SAMPLES = {
  dispatch: { callsign: 'dossier driver', type: 'dispatch', data: { stream: 'stream three' } },
  tests_green: { callsign: 'dossier driver', type: 'tests_green', data: { stream: 'stream three' } },
  tests_red: {
    callsign: 'dossier driver',
    type: 'tests_red',
    data: { stream: 'stream three', suite: 'integration' },
  },
  fix_verified: { callsign: 'dossier driver', type: 'fix_verified', data: { stream: 'stream three' } },
  pr_open: { callsign: 'dossier driver', type: 'pr_open', data: { stream: 'stream three' } },
  review_start: { callsign: 'review one', type: 'review_start', data: { stream: 'stream three' } },
  findings: {
    callsign: 'review one',
    type: 'findings',
    data: { stream: 'stream three', count: 3, blocking: true },
  },
  fixes_pushed: { callsign: 'dossier driver', type: 'fixes_pushed', data: { stream: 'stream three' } },
  review_clean: { callsign: 'review one', type: 'review_clean', data: { stream: 'stream three' } },
  gate_pass: { callsign: 'dossier driver', type: 'gate_pass', data: { stream: 'stream three' } },
  merged: { callsign: 'dossier driver', type: 'merged', data: { stream: 'stream three' } },
  stuck: {
    callsign: 'dossier driver',
    type: 'stuck',
    data: { stream: 'stream three', reason: 'no reviewer' },
  },
  pipeline_green: { callsign: 'ci watch', type: 'pipeline_green', data: { recovered: true } },
  pipeline_red: { callsign: 'ci watch', type: 'pipeline_red', data: { job: 'deploy' } },
};

test('every event type has a sample and every sample has an event type', () => {
  assert.deepEqual(Object.keys(SAMPLES).sort(), [...EVENT_TYPES].sort());
});

for (const type of EVENT_TYPES) {
  const e = SAMPLES[type];

  test(`brevity[${type}] respects the hard word cap`, () => {
    const cap = isPriority(type) ? PRIORITY_WORD_CAP : ROUTINE_WORD_CAP;
    const phrase = brevity(e);
    assert.ok(
      wordCount(phrase) <= cap,
      `"${phrase}" is ${wordCount(phrase)} words, cap ${cap}`,
    );
  });

  test(`brevity[${type}] leads with the callsign (attention prefix first if priority)`, () => {
    const phrase = brevity(e);
    const expectedStart = isPriority(type) ? `PAN PAN, ${e.callsign}` : e.callsign;
    assert.ok(phrase.startsWith(expectedStart), `"${phrase}" should start with "${expectedStart}"`);
  });

  test(`brevity[${type}] carries its stressed status word`, () => {
    const phrase = brevity(e).toLowerCase();
    assert.ok(phrase.includes(STATUS_WORDS[type].toLowerCase()), `"${phrase}" lacks "${STATUS_WORDS[type]}"`);
  });

  test(`prose[${type}] is at least twice as long as brevity (the whole point)`, () => {
    assert.ok(
      wordCount(prose(e)) >= 2 * wordCount(brevity(e)),
      `prose "${prose(e)}" is not 2x brevity "${brevity(e)}"`,
    );
  });
}

test('status words are unique across event types', () => {
  const words = Object.values(STATUS_WORDS).map((w) => w.toLowerCase());
  assert.equal(new Set(words).size, words.length);
});

test('exactly the two escalation types are priority', () => {
  assert.deepEqual(
    EVENT_TYPES.filter(isPriority).sort(),
    ['pipeline_red', 'stuck'],
  );
});
