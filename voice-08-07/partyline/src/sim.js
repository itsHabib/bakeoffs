// The fleet simulator: fixture data with a clock. Six agents advance through
// realistic delivery state machines (dispatch -> tests -> review -> findings ->
// gate -> merge) with one forced test failure, one stuck agent, and (full run)
// one CI red. Fully deterministic: same scenario -> identical timeline.

import { mulberry32, u } from './prng.js';

export const AGENTS = [
  { id: 'roxiq', callsign: 'roxiq driver', role: 'driver', stream: 'stream one' },
  { id: 'dossier', callsign: 'dossier driver', role: 'driver', stream: 'stream two' },
  { id: 'ship', callsign: 'ship driver', role: 'driver', stream: 'stream three' },
  { id: 'rev1', callsign: 'review one', role: 'reviewer' },
  { id: 'rev2', callsign: 'review two', role: 'reviewer' },
  { id: 'ci', callsign: 'ci watch', role: 'watcher' },
];

const byId = Object.fromEntries(AGENTS.map((a) => [a.id, a]));

export const SCENARIOS = {
  full: {
    label: 'Full run · ~3.5 min',
    seed: 0xc0ffee,
    duration: 220,
    pace: 2.2,
    drivers: {
      roxiq: { startAt: 3, findings: { count: 3, blocking: true }, reviewer: 'rev1' },
      dossier: { startAt: 12, fail: 'integration', reviewer: 'rev2' },
      ship: { startAt: 22, stuck: true, reviewer: 'rev2' },
    },
    ci: { start: 30, redAt: 110, redJob: 'deploy' },
  },
  highlight: {
    label: 'Highlight · 45 s',
    seed: 45,
    duration: 45,
    pace: 0.35,
    drivers: {
      roxiq: { startAt: 1, reviewer: 'rev1' },
      dossier: { startAt: 4, fail: 'unit', reviewer: 'rev1' },
      ship: { startAt: 7, stuck: true, reviewer: 'rev2' },
    },
    ci: { start: 10 },
  },
};

const round1 = (t) => Math.round(t * 10) / 10;

function driverArc(rng, drv, cfg, pace) {
  const out = [];
  let t = cfg.startAt;
  const rev = byId[cfg.reviewer];
  const push = (agent, type, data = {}) =>
    out.push({
      t: round1(t),
      agentId: agent.id,
      callsign: agent.callsign,
      role: agent.role,
      type,
      data: { stream: drv.stream, ...data },
    });

  push(drv, 'dispatch');
  t += u(rng, 16, 26) * pace;
  if (cfg.fail) {
    push(drv, 'tests_red', { suite: cfg.fail });
    t += u(rng, 10, 16) * pace;
    push(drv, 'fix_verified');
    t += u(rng, 6, 10) * pace;
  }
  push(drv, 'tests_green');
  t += u(rng, 2, 4) * pace;
  push(drv, 'pr_open');

  if (cfg.stuck) {
    t += u(rng, 22, 28) * pace;
    push(drv, 'stuck', { reason: 'no reviewer' });
    t += u(rng, 5, 8) * pace;
    push(rev, 'review_start');
    t += u(rng, 7, 10) * pace;
    push(rev, 'review_clean');
  } else {
    t += u(rng, 5, 9) * pace;
    push(rev, 'review_start');
    t += u(rng, 8, 14) * pace;
    if (cfg.findings) {
      push(rev, 'findings', cfg.findings);
      t += u(rng, 10, 16) * pace;
      push(drv, 'fixes_pushed');
      t += u(rng, 6, 10) * pace;
    }
    push(rev, 'review_clean');
  }
  t += u(rng, 3, 6) * pace;
  push(drv, 'gate_pass');
  t += u(rng, 3, 5) * pace;
  push(drv, 'merged');
  return out;
}

function ciArc(rng, cfg, pace, duration) {
  const ci = byId.ci;
  const out = [];
  const push = (t, type, data = {}) =>
    out.push({ t: round1(t), agentId: ci.id, callsign: ci.callsign, role: ci.role, type, data });

  let t = cfg.start;
  const greens = [];
  while (t < duration) {
    greens.push(t);
    t += u(rng, 28, 40) * pace;
  }
  for (const g of greens) {
    // A routine "main green" right next to the red would lie to the ear.
    if (cfg.redAt != null && Math.abs(g - cfg.redAt) < 30 * pace) continue;
    push(g, 'pipeline_green');
  }
  if (cfg.redAt != null) {
    push(cfg.redAt, 'pipeline_red', { job: cfg.redJob });
    push(cfg.redAt + u(rng, 12, 18) * pace, 'pipeline_green', { recovered: true });
  }
  return out;
}

export function generateTimeline(name) {
  const sc = SCENARIOS[name];
  if (!sc) throw new Error(`unknown scenario: ${name}`);
  const rng = mulberry32(sc.seed);
  const events = [];
  for (const id of ['roxiq', 'dossier', 'ship']) {
    events.push(...driverArc(rng, byId[id], sc.drivers[id], sc.pace));
  }
  events.push(...ciArc(rng, sc.ci, sc.pace, sc.duration));
  const kept = events.filter((e) => e.t <= sc.duration);
  kept.forEach((e, i) => (e.seq = i));
  kept.sort((a, b) => a.t - b.t || a.seq - b.seq);
  return { name, duration: sc.duration, events: kept };
}
