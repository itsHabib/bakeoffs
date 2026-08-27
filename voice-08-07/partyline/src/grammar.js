// The brevity grammar: every fleet event type maps to a fixed phrase template.
// Ordering is constant — callsign, subject, status, need — so a half-attending
// ear can bind who/what/state without parsing. Escalations carry the "PAN PAN"
// attention prefix. Word caps are enforced by tests, not by convention.

export const ROUTINE_WORD_CAP = 9;
export const PRIORITY_WORD_CAP = 12;

const PRIORITY_TYPES = new Set(['stuck', 'pipeline_red']);

export const isPriority = (type) => PRIORITY_TYPES.has(type);

export const wordCount = (phrase) => phrase.trim().split(/\s+/).length;

const NUM = ['zero', 'one', 'two', 'three', 'four', 'five', 'six', 'seven'];
const num = (n) => NUM[n] ?? String(n);

// One stressed status word (or two-word pair) per event type, always in the
// same slot. These are what an ear keys on; tests assert they stay unique.
export const STATUS_WORDS = {
  dispatch: 'dispatched',
  tests_green: 'tests green',
  tests_red: 'tests red',
  fix_verified: 'rerunning',
  pr_open: 'PR open',
  review_start: 'taking',
  findings: 'findings',
  fixes_pushed: 'addressed',
  review_clean: 'clean',
  gate_pass: 'gate pass',
  merged: 'merged',
  stuck: 'stuck',
  pipeline_green: 'main green',
  pipeline_red: 'main red',
};

const BREVITY = {
  dispatch: (e) => `${e.callsign}, ${e.data.stream}, dispatched, implementing.`,
  tests_green: (e) => `${e.callsign}, ${e.data.stream}, tests green.`,
  tests_red: (e) => `${e.callsign}, ${e.data.stream}, tests red, ${e.data.suite} suite, fixing.`,
  fix_verified: (e) => `${e.callsign}, ${e.data.stream}, fix in, rerunning.`,
  pr_open: (e) => `${e.callsign}, ${e.data.stream}, PR open, awaiting review.`,
  review_start: (e) => `${e.callsign}, taking ${e.data.stream}.`,
  findings: (e) =>
    `${e.callsign}, ${e.data.stream}, ${num(e.data.count)} findings, ${e.data.blocking ? 'one blocking' : 'none blocking'}.`,
  fixes_pushed: (e) => `${e.callsign}, ${e.data.stream}, findings addressed, re-review requested.`,
  review_clean: (e) => `${e.callsign}, ${e.data.stream}, clean, released.`,
  gate_pass: (e) => `${e.callsign}, ${e.data.stream}, gate pass, merging.`,
  merged: (e) => `${e.callsign}, ${e.data.stream}, merged, off channel.`,
  stuck: (e) => `PAN PAN, ${e.callsign}, ${e.data.stream}, stuck, ${e.data.reason}, need human.`,
  pipeline_green: (e) => `${e.callsign}, main green${e.data.recovered ? ' again' : ''}.`,
  pipeline_red: (e) => `PAN PAN, ${e.callsign}, main red, ${e.data.job} job failing.`,
};

// The prose control group: the same events as natural sentences. This is what
// every "agents talk to you" demo sounds like, and why they fail — you cannot
// half-listen to it. One toggle, no other comparison mode.
const PROSE = {
  dispatch: (e) =>
    `This is the ${e.callsign} checking in on the channel — I have just been dispatched onto ${e.data.stream} and I am getting started on the implementation right now.`,
  tests_green: (e) =>
    `The ${e.callsign} here with an update on ${e.data.stream} — the full test suite has finished running and every single test is passing cleanly.`,
  tests_red: (e) =>
    `The ${e.callsign} with some bad news on ${e.data.stream} — the ${e.data.suite} suite has come back with failures, so I am digging into what broke and working on a fix.`,
  fix_verified: (e) =>
    `The ${e.callsign} again on ${e.data.stream} — I believe I have identified and fixed the problem, and I am rerunning the whole test suite now to confirm it.`,
  pr_open: (e) =>
    `The ${e.callsign} reporting that ${e.data.stream} now has a pull request open with all the checks attached, and I am waiting for a reviewer to pick it up.`,
  review_start: (e) =>
    `This is ${e.callsign} letting the channel know that I am picking up the pull request on ${e.data.stream} and starting my review pass on it now.`,
  findings: (e) =>
    `${cap(e.callsign)} back with results on ${e.data.stream} — I have posted ${num(e.data.count)} findings on the pull request, and ${e.data.blocking ? 'one of them is blocking, so it must be addressed before this can merge' : 'none of them are blocking, but they should still be looked at'}.`,
  fixes_pushed: (e) =>
    `The ${e.callsign} on ${e.data.stream} — I have gone through every review finding, pushed fixes for all of them, and requested another review pass on the branch.`,
  review_clean: (e) =>
    `${cap(e.callsign)} closing out ${e.data.stream} — the latest revision addresses everything I raised, the review is clean, and I am releasing it onward.`,
  gate_pass: (e) =>
    `The ${e.callsign} on ${e.data.stream} — the merge gate has evaluated the pull request and passed it, so I am proceeding with the merge right now.`,
  merged: (e) =>
    `The ${e.callsign} with the final word on ${e.data.stream} — the pull request has merged to main and I am signing off of the channel.`,
  stuck: (e) =>
    `Attention everyone, this is the ${e.callsign} on ${e.data.stream} — I have been holding for a long time with ${e.data.reason} responding to me, I am completely stuck, and I need a human to step in and take a look.`,
  pipeline_green: (e) =>
    `This is ${e.callsign} with a routine report — the main branch pipeline has completed ${e.data.recovered ? 'a recovery run and is back to' : 'another full run and everything is'} green across the board.`,
  pipeline_red: (e) =>
    `Attention everyone, this is ${e.callsign} with an urgent report — the main branch is red, the ${e.data.job} job is failing, and everything downstream of it is blocked until someone looks at it.`,
};

const cap = (s) => s.charAt(0).toUpperCase() + s.slice(1);

export const EVENT_TYPES = Object.keys(BREVITY);

export const brevity = (e) => BREVITY[e.type](e);
export const prose = (e) => PROSE[e.type](e);
