import { test } from 'node:test';
import assert from 'node:assert/strict';
import { normalize, grade } from '../public/grader.js';
import { SCENARIO, CANNED } from '../public/scenario.js';

const ex = id => SCENARIO.exchanges.find(e => e.id === id);
const statusOf = (result, key) => result.elements.find(e => e.key === key)?.status;

// ---------------------------------------------------------------- normalize

const NORMALIZE_CASES = [
  ['Runway Two Seven', 'runway 27'],
  ['runway niner', 'runway 9'],
  ['Skyhawk one two three alpha bravo', 'skyhawk 123ab'],
  ['skyhawk 123 alpha bravo', 'skyhawk 123ab'],
  ['Skyhawk 123AB', 'skyhawk 123ab'],
  ['three alpha bravo', '3ab'],
  ['one two one point niner', '121.9'],
  ['ground on 121.9', 'ground on 121.9'],
  ['wind zero niner zero at eight', 'wind 090 at 8'],
  ['Hold short of Runway 9!', 'hold short of runway 9'],
  ['taxi via Alpha.', 'taxi via a'],
  ['CLEARED FOR TAKEOFF', 'cleared for takeoff'],
  ['number two, following a Cessna', 'number 2 following a cessna'],
];

for (const [input, expected] of NORMALIZE_CASES) {
  test(`normalize: "${input}" -> "${expected}"`, () => {
    assert.equal(normalize(input).join(' '), expected);
  });
}

// -------------------------------------------------- table-driven grading
// Each case: [exchange id, student says, expected per-element statuses, opts]

const GRADE_CASES = [
  // --- taxi out ---
  ['taxi-out', 'Runway 9, taxi via Alpha, hold short runway 9, Skyhawk 123AB',
    { runway: 'ok', route: 'ok', holdshort: 'ok', callsign: 'ok' }],
  ['taxi-out', 'runway niner taxi via alpha hold short of runway niner skyhawk one two three alpha bravo',
    { runway: 'ok', route: 'ok', holdshort: 'ok', callsign: 'ok' }], // spoken form
  ['taxi-out', 'Runway 9, taxi via Alpha, hold short runway 9',
    { runway: 'ok', route: 'ok', holdshort: 'ok', callsign: 'missing' }], // dropped callsign
  ['taxi-out', 'Runway 9, taxi via Alpha, Skyhawk 123AB',
    { runway: 'ok', route: 'ok', holdshort: 'missing', callsign: 'ok' }], // dropped hold short
  ['taxi-out', 'Runway 9, taxi via Alpha, hold short runway 27, Skyhawk 123AB',
    { holdshort: 'wrong' }], // held short of the wrong runway
  ['taxi-out', 'Hold short runway 9, taxi via Alpha, runway 9, Skyhawk 3AB',
    { runway: 'ok', route: 'ok', holdshort: 'ok', callsign: 'ok' }], // order does not matter

  // --- takeoff ---
  ['takeoff', 'Runway 9, cleared for takeoff, left closed traffic, Skyhawk 123AB',
    { runway: 'ok', clearance: 'ok', traffic: 'ok', callsign: 'ok' }],
  ['takeoff', 'Cleared for takeoff runway 9, left traffic, 3AB',
    { runway: 'ok', clearance: 'ok', traffic: 'ok', callsign: 'ok' }], // shorthand
  ['takeoff', 'Runway 9, cleared for takeoff, right traffic, Skyhawk 123AB',
    { traffic: 'wrong' }], // wrong pattern direction
  ['takeoff', 'Runway 9, cleared to land, left closed traffic, Skyhawk 123AB',
    { clearance: 'wrong' }], // wrong clearance type

  // --- sequence ---
  ['sequence', 'Number 2, report midfield left downwind, Skyhawk 3AB',
    { sequence: 'ok', report: 'ok', callsign: 'ok' }],
  ['sequence', 'Number 2, traffic in sight, report downwind, Skyhawk 3AB',
    { sequence: 'ok', report: 'ok', traffic: 'ok', callsign: 'ok' }],
  ['sequence', 'Roger',
    { sequence: 'missing', report: 'missing', callsign: 'missing' }, { ackOnly: true }],
  ['sequence', 'Roger, Skyhawk 3AB',
    { sequence: 'missing', report: 'missing', callsign: 'ok' }, { ackOnly: true }],
  ['sequence', 'Number 3, report midfield downwind, Skyhawk 3AB',
    { sequence: 'wrong' }],

  // --- landing: the money case ---
  ['landing', 'Runway 9, cleared to land, Skyhawk 3AB',
    { runway: 'ok', clearance: 'ok', callsign: 'ok' }],
  ['landing', 'Cleared to land runway 27, Skyhawk 3AB',
    { runway: 'wrong', clearance: 'ok', callsign: 'ok' }], // cleared 9, read back 27
  ['landing', 'cleared to land runway two seven skyhawk three alpha bravo',
    { runway: 'wrong' }], // same failure, spoken form
  ['landing', 'Cleared to land, Skyhawk 3AB',
    { runway: 'missing', clearance: 'ok', callsign: 'ok' }],

  // --- taxi in ---
  ['taxi-in', 'Right at Bravo, taxi to the ramp, ground on 121.9, Skyhawk 3AB',
    { exit: 'ok', ramp: 'ok', freq: 'ok', callsign: 'ok' }],
  ['taxi-in', 'right at bravo to the ramp ground on one two one point niner 3AB',
    { exit: 'ok', ramp: 'ok', freq: 'ok', callsign: 'ok' }], // spoken frequency
  ['taxi-in', 'Left at Bravo, taxi to the ramp, ground on 121.9, Skyhawk 3AB',
    { exit: 'wrong' }],
  ['taxi-in', 'Right at Bravo, taxi to the ramp, ground on 121.7, Skyhawk 3AB',
    { freq: 'wrong' }],
];

for (const [id, said, expected, opts = {}] of GRADE_CASES) {
  test(`grade ${id}: "${said}"`, () => {
    const result = grade(said, ex(id));
    for (const [key, status] of Object.entries(expected)) {
      assert.equal(statusOf(result, key), status,
        `element "${key}" should be ${status} (got ${statusOf(result, key)})`);
    }
    if (opts.ackOnly !== undefined) assert.equal(result.ackOnly, opts.ackOnly);
  });
}

// ------------------------------------------------------------ score math

test('perfect readback scores 1.0; optional elements are not counted', () => {
  const r = grade('Number 2, report midfield left downwind, Skyhawk 3AB', ex('sequence'));
  assert.equal(r.score, 1);
});

test('wrong runway drags the score and is counted as wrong, not missing', () => {
  const r = grade('Cleared to land runway 27, Skyhawk 3AB', ex('landing'));
  assert.equal(r.wrongCount, 1);
  assert.ok(r.score < 1);
});

test('empty transmission: everything missing, score 0', () => {
  const r = grade('', ex('landing'));
  assert.equal(r.score, 0);
  assert.ok(r.elements.every(e => e.status === 'missing'));
});

// ------------------------------------------------- canned demo integrity
// The scripted student must fail exactly where DEMO.md says it fails.

test('canned script: exchange 1 drops the callsign, rest ok', () => {
  const r = grade(CANNED[0].text, SCENARIO.exchanges[0]);
  assert.equal(statusOf(r, 'callsign'), 'missing');
  assert.equal(r.elements.filter(e => e.status === 'ok').length, 3);
});

test('canned script: exchange 3 is roger-only', () => {
  const r = grade(CANNED[2].text, SCENARIO.exchanges[2]);
  assert.equal(r.ackOnly, true);
  assert.equal(r.score, 0);
});

test('canned script: exchange 4 reads back the wrong runway', () => {
  const r = grade(CANNED[3].text, SCENARIO.exchanges[3]);
  assert.equal(statusOf(r, 'runway'), 'wrong');
});

test('canned script: clean exchanges grade 1.0', () => {
  for (const i of [1, 4]) {
    const r = grade(CANNED[i].text, SCENARIO.exchanges[i]);
    assert.equal(r.score, 1, `canned exchange ${i} should be perfect`);
  }
});
