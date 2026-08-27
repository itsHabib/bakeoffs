// One scenario: VFR pattern work at a single towered field (fictional
// "Millbrook"), runway 9. Controller lines are templates — no model, so they
// are free, instant, and identical every run. Phraseology follows FAA AIM
// ch. 4 / JO 7110.65 conventions from general knowledge; calls I'm less sure
// of carry a TODO rather than fake confidence.

// Callsign: Skyhawk 123AB. Full and abbreviated (after initial contact)
// forms are both accepted, spelled out or compact.
const CALLSIGN = {
  key: 'callsign',
  label: 'Callsign',
  accept: ['skyhawk 123ab', 'skyhawk n123ab', 'skyhawk 3ab', '123ab', 'n123ab', '3ab'],
};

const RUNWAY_WRONG = [
  { re: /runway (\d+)/, expected: '9', label: 'runway ' },
  { re: /cleared to land (\d+)/, expected: '9', label: 'runway ' },
  { re: /cleared for takeoff (\d+)/, expected: '9', label: 'runway ' },
];

const RUNWAY_9 = {
  key: 'runway',
  label: 'Runway 9',
  accept: ['runway 9', 'runway 09', '9 cleared', 'cleared to land 9', 'cleared for takeoff 9'],
  wrong: RUNWAY_WRONG,
};

export const SCENARIO = {
  field: 'Millbrook Tower (fictional Class D)',
  aircraft: 'Cessna 172 — Skyhawk 123AB',
  atis: 'Information Alpha: wind 090 at 8, altimeter 30.02, runway 9 in use.',
  intro: 'You are at the ramp with information Alpha, calling for taxi. Read back every instruction the controller gives you — the grader scores each required element as you go.',
  exchanges: [
    {
      id: 'taxi-out',
      phase: 'GROUND — TAXI',
      controller: 'Skyhawk 123AB, Millbrook Ground, runway 9, taxi via Alpha, hold short of runway 9.',
      tts: 'Skyhawk one two three alpha bravo, Millbrook Ground, runway niner, taxi via alpha, hold short of runway niner.',
      solution: 'Runway 9, taxi via Alpha, hold short runway 9, Skyhawk 123AB.',
      elements: [
        RUNWAY_9,
        { key: 'route', label: 'Taxi via Alpha', accept: ['via a', 'taxi a'] },
        { key: 'holdshort', label: 'Hold short runway 9',
          accept: ['hold short of runway 9', 'hold short runway 9', 'hold short 9', 'hold short'],
          wrong: [{ re: /hold short (?:of )?(?:runway )?(\d+)/, expected: '9', label: 'hold short ' }] },
        CALLSIGN,
      ],
    },
    {
      id: 'takeoff',
      phase: 'TOWER — TAKEOFF',
      controller: 'Skyhawk 123AB, Millbrook Tower, runway 9, wind 090 at 8, cleared for takeoff, make left closed traffic.',
      tts: 'Skyhawk one two three alpha bravo, Millbrook Tower, runway niner, wind zero niner zero at eight, cleared for takeoff, make left closed traffic.',
      solution: 'Runway 9, cleared for takeoff, left closed traffic, Skyhawk 123AB.',
      elements: [
        RUNWAY_9,
        { key: 'clearance', label: 'Cleared for takeoff',
          accept: ['cleared for takeoff', 'cleared takeoff'],
          wrong: [{ re: / (cleared to land)/, expected: 'cleared for takeoff', label: '' }] },
        { key: 'traffic', label: 'Left closed traffic',
          accept: ['left closed traffic', 'left traffic', 'closed traffic left'],
          wrong: [{ re: / (right) (?:closed )?traffic/, expected: 'left', label: '' }] },
        CALLSIGN,
      ],
    },
    {
      id: 'sequence',
      phase: 'TOWER — PATTERN',
      controller: 'Skyhawk 3AB, number 2 following a Cessna on short final, report midfield left downwind.',
      tts: 'Skyhawk three alpha bravo, number two, following a Cessna on short final, report midfield left downwind.',
      solution: 'Number 2, report midfield left downwind, Skyhawk 3AB.',
      elements: [
        { key: 'sequence', label: 'Number 2',
          accept: ['number 2'],
          wrong: [{ re: /number (\d+)/, expected: '2', label: 'number ' }] },
        { key: 'report', label: 'Report midfield downwind',
          accept: ['report midfield left downwind', 'report midfield downwind', 'report left downwind', 'report downwind'] },
        { key: 'traffic', label: 'Traffic in sight', optional: true,
          accept: ['traffic in sight', 'looking for traffic', 'looking'] },
        CALLSIGN,
      ],
    },
    {
      id: 'landing',
      phase: 'TOWER — LANDING',
      preamble: 'You report: "Millbrook Tower, Skyhawk 3AB, midfield left downwind, runway 9."',
      controller: 'Skyhawk 3AB, runway 9, cleared to land, wind 090 at 8.',
      tts: 'Skyhawk three alpha bravo, runway niner, cleared to land, wind zero niner zero at eight.',
      solution: 'Runway 9, cleared to land, Skyhawk 3AB.',
      elements: [
        RUNWAY_9,
        { key: 'clearance', label: 'Cleared to land',
          accept: ['cleared to land'],
          wrong: [{ re: / (cleared for takeoff)/, expected: 'cleared to land', label: '' }] },
        CALLSIGN,
      ],
    },
    {
      id: 'taxi-in',
      phase: 'GROUND — TAXI IN',
      // TODO(phraseology): verify "monitor ground" vs "contact ground" for a
      // Class D exit instruction; using "contact" here.
      controller: 'Skyhawk 3AB, turn right at Bravo, taxi to the ramp, contact ground on 121.9.',
      tts: 'Skyhawk three alpha bravo, turn right at bravo, taxi to the ramp, contact ground on one two one point niner.',
      solution: 'Right at Bravo, taxi to the ramp, ground on 121.9, Skyhawk 3AB.',
      elements: [
        { key: 'exit', label: 'Right at Bravo',
          accept: ['right at b', 'right on b', 'right b'],
          wrong: [{ re: / (left) (?:at|on) [a-z]\b/, expected: 'right', label: '' }] },
        { key: 'ramp', label: 'Taxi to the ramp', accept: ['to the ramp', 'taxi ramp', 'the ramp'] },
        { key: 'freq', label: 'Ground 121.9',
          accept: ['121.9', 'point 9'],
          wrong: [{ re: /ground (?:on )?(1\d\d\.\d)/, expected: '121.9', label: '' }] },
        CALLSIGN,
      ],
    },
  ],
};

// The scripted student for the canned no-mic demo: three realistic readback
// failures (dropped callsign, "roger" instead of a readback, wrong runway)
// so a judge watches the grader catch each one without touching a mic.
export const CANNED = [
  { text: 'Runway 9, taxi via Alpha, hold short runway 9', note: 'drops the callsign' },
  { text: 'Runway 9, cleared for takeoff, left closed traffic, Skyhawk 123AB', note: 'clean readback' },
  { text: 'Roger', note: '"roger" is not a readback' },
  { text: 'Cleared to land runway 27, Skyhawk 3AB', note: 'reads back the WRONG runway' },
  { text: 'Right at Bravo, taxi to the ramp, ground on 121.9, Skyhawk 3AB', note: 'clean readback' },
];
