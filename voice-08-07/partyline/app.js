// Playback engine: walks the deterministic timeline in real time, pushes each
// event through the channel policy, and speaks whatever survives. The board
// renders ground truth from ALL events; the transcript shows only what aired —
// the gap between them is the drop policy working.

import { AGENTS, SCENARIOS, generateTimeline } from './src/sim.js';
import { brevity, prose, wordCount, isPriority } from './src/grammar.js';
import { createChannel } from './src/channel.js';

const $ = (id) => document.getElementById(id);
const board = $('board');
const transcript = $('transcript');

let mode = 'brevity';
let run = null; // { timeline, startedAt, nextIdx, channel, ticker, lastDropped }
let speaking = false;

// ---------- audio: squelch clicks, alert tone, speech ----------

let actx = null;
const audio = () => (actx ??= new (window.AudioContext || window.webkitAudioContext)());

function squelch(durMs = 60) {
  return new Promise((res) => {
    const ctx = audio();
    const n = ctx.createBufferSource();
    const len = Math.floor((ctx.sampleRate * durMs) / 1000);
    const buf = ctx.createBuffer(1, len, ctx.sampleRate);
    const d = buf.getChannelData(0);
    for (let i = 0; i < len; i++) d[i] = (Math.random() * 2 - 1) * (1 - i / len);
    n.buffer = buf;
    const bp = ctx.createBiquadFilter();
    bp.type = 'bandpass';
    bp.frequency.value = 2000;
    const g = ctx.createGain();
    g.gain.value = 0.25;
    n.connect(bp).connect(g).connect(ctx.destination);
    n.onended = res;
    n.start();
  });
}

function alertTone() {
  return new Promise((res) => {
    const ctx = audio();
    const g = ctx.createGain();
    g.gain.value = 0.18;
    g.connect(ctx.destination);
    const t0 = ctx.currentTime;
    [0, 0.2].forEach((off, i) => {
      const o = ctx.createOscillator();
      o.type = 'square';
      o.frequency.value = i === 0 ? 880 : 1100;
      o.connect(g);
      o.start(t0 + off);
      o.stop(t0 + off + 0.14);
    });
    setTimeout(res, 420);
  });
}

const PITCH = { roxiq: 1.0, dossier: 0.85, ship: 1.15, rev1: 0.7, rev2: 1.3, ci: 0.6 };
const voiceMap = new Map();

function assignVoices() {
  const en = speechSynthesis
    .getVoices()
    .filter((v) => v.lang && v.lang.toLowerCase().startsWith('en'))
    .sort((a, b) => a.name.localeCompare(b.name));
  AGENTS.forEach((a, i) => {
    const voice = en.length ? en[Math.floor((i * en.length) / AGENTS.length) % en.length] : null;
    voiceMap.set(a.id, { voice, pitch: PITCH[a.id] ?? 1 });
  });
}
if ('speechSynthesis' in window) {
  speechSynthesis.onvoiceschanged = assignVoices;
  assignVoices();
}

function speakPhrase(phrase, agentId, priority) {
  return new Promise((res) => {
    if (!('speechSynthesis' in window)) return res();
    const utt = new SpeechSynthesisUtterance(phrase);
    const v = voiceMap.get(agentId) ?? { voice: null, pitch: 1 };
    if (v.voice) utt.voice = v.voice;
    utt.pitch = Math.min(2, v.pitch + (priority ? 0.15 : 0));
    utt.rate = mode === 'brevity' ? 1.12 : 1.02;
    let done = false;
    const finish = () => {
      if (done) return;
      done = true;
      res();
    };
    utt.onend = finish;
    utt.onerror = finish;
    // Some engines drop onend; never let the channel jam on a lost callback.
    setTimeout(finish, 3000 + wordCount(phrase) * 500);
    speechSynthesis.speak(utt);
  });
}

// ---------- board ----------

const BOARD_LABEL = {
  dispatch: () => 'implementing',
  tests_green: () => 'tests green',
  tests_red: (e) => `tests RED (${e.data.suite})`,
  fix_verified: () => 'rerunning tests',
  pr_open: () => 'PR open · awaiting review',
  review_start: (e) => `reviewing ${e.data.stream}`,
  findings: (e) => `posted ${e.data.count} findings`,
  fixes_pushed: () => 'fixes pushed · re-review',
  review_clean: (e) => `${e.data.stream} clean · released`,
  gate_pass: () => 'gate pass · merging',
  merged: () => 'MERGED ✓',
  stuck: () => 'STUCK · needs human',
  pipeline_green: () => 'main green',
  pipeline_red: (e) => `main RED (${e.data.job})`,
};

const STATE_CLASS = {
  tests_red: 'warn',
  stuck: 'alert',
  pipeline_red: 'alert',
  merged: 'done',
};

function renderBoard() {
  board.innerHTML = '';
  for (const a of AGENTS) {
    const card = document.createElement('div');
    card.className = 'card';
    card.id = `card-${a.id}`;
    card.innerHTML = `
      <div class="who"><span class="cs">${a.callsign}</span><span class="role">${a.role}${a.stream ? ' · ' + a.stream : ''}</span></div>
      <div class="state">standing by</div>`;
    board.appendChild(card);
  }
}

function boardUpdate(e) {
  const card = $(`card-${e.agentId}`);
  if (!card) return;
  card.querySelector('.state').textContent = BOARD_LABEL[e.type](e);
  card.classList.remove('warn', 'alert', 'done');
  const cls = STATE_CLASS[e.type];
  if (cls) card.classList.add(cls);
}

// ---------- transcript ----------

const fmt = (sec) => `${Math.floor(sec / 60)}:${String(Math.floor(sec % 60)).padStart(2, '0')}`;

function addLine(e, phrase, priority) {
  const div = document.createElement('div');
  div.className = `line${priority ? ' pri' : ''}`;
  div.innerHTML = `<span class="t">${fmt(e.t)}</span><span class="cs">${e.callsign}</span><span class="ph"></span>`;
  div.querySelector('.ph').textContent = phrase;
  transcript.appendChild(div);
  transcript.scrollTop = transcript.scrollHeight;
}

function addSysLine(text, cls = 'sys') {
  const div = document.createElement('div');
  div.className = `line ${cls}`;
  div.textContent = text;
  transcript.appendChild(div);
  transcript.scrollTop = transcript.scrollHeight;
}

// ---------- run loop ----------

function setTX(callsign, priority) {
  $('txwho').textContent = callsign ?? '— channel quiet —';
  document.querySelector('.txbar').classList.toggle('pri', Boolean(callsign && priority));
  const dot = $('txdot');
  dot.className = 'dot' + (callsign ? (priority ? ' pri' : ' on') : '');
  document.querySelectorAll('.card.speaking').forEach((c) => c.classList.remove('speaking'));
  if (callsign) {
    const agent = AGENTS.find((a) => a.callsign === callsign);
    if (agent) $(`card-${agent.id}`)?.classList.add('speaking');
  }
}

function refreshStats() {
  if (!run) return;
  $('s-tx').textContent = run.channel.stats.transmitted;
  $('s-co').textContent = run.channel.stats.coalesced;
  $('s-dr').textContent = run.channel.stats.dropped;
  $('s-q').textContent = run.channel.pending;
}

const phraseFor = (e) => (mode === 'brevity' ? brevity(e) : prose(e));

async function transmit(item) {
  speaking = true;
  const e = item.payload;
  const priority = item.priority;
  if (run && run.channel.stats.dropped > run.lastDropped) {
    const n = run.channel.stats.dropped - run.lastDropped;
    run.lastDropped = run.channel.stats.dropped;
    addSysLine(`· ${n} routine call${n > 1 ? 's' : ''} dropped — frequency saturated ·`);
  }
  const phrase = phraseFor(e);
  setTX(e.callsign, priority);
  addLine(e, phrase, priority);
  try {
    if (priority) await alertTone();
    else await squelch();
    await speakPhrase(phrase, e.agentId, priority);
    await squelch(40);
  } finally {
    speaking = false;
    setTX(null);
  }
}

function tick() {
  if (!run) return;
  const elapsed = (performance.now() - run.startedAt) / 1000;
  $('clock').textContent = fmt(Math.min(elapsed, run.timeline.duration));

  const evs = run.timeline.events;
  while (run.nextIdx < evs.length && evs[run.nextIdx].t <= elapsed) {
    const e = evs[run.nextIdx++];
    boardUpdate(e);
    run.channel.enqueue({
      key: `${e.agentId}:${e.data.stream ?? ''}`,
      priority: isPriority(e.type),
      words: wordCount(phraseFor(e)),
      payload: e,
    });
  }

  if (!speaking) {
    const item = run.channel.next();
    if (item) transmit(item).then(refreshStats);
  }
  refreshStats();

  if (run.nextIdx >= evs.length && run.channel.pending === 0 && !speaking) {
    const s = run.channel.stats;
    addSysLine(
      `— channel closed · ${s.transmitted} transmissions · ${s.coalesced} coalesced · ${s.dropped} dropped —`,
      'end',
    );
    stopRun(false);
  }
}

function startRun() {
  stopRun(false);
  const scenario = $('scenario').value;
  run = {
    timeline: generateTimeline(scenario),
    startedAt: performance.now(),
    nextIdx: 0,
    channel: createChannel(),
    lastDropped: 0,
    ticker: setInterval(tick, 100),
  };
  transcript.innerHTML = '';
  renderBoard();
  audio().resume();
  addSysLine(`— channel open · ${SCENARIOS[scenario].label} · ${mode} mode —`);
  $('play').textContent = '■ Close channel';
  $('play').classList.add('stop');
}

function stopRun(announce = true) {
  if (run) {
    clearInterval(run.ticker);
    run = null;
  }
  if ('speechSynthesis' in window) speechSynthesis.cancel();
  speaking = false;
  setTX(null);
  if (announce) addSysLine('— channel closed by operator —', 'end');
  $('play').textContent = '▶ Open channel';
  $('play').classList.remove('stop');
}

// ---------- controls ----------

$('play').addEventListener('click', () => (run ? stopRun() : startRun()));

$('modeseg').addEventListener('click', (ev) => {
  const btn = ev.target.closest('button');
  if (!btn) return;
  mode = btn.dataset.mode;
  document.querySelectorAll('#modeseg button').forEach((b) => b.classList.toggle('on', b === btn));
  if (run) addSysLine(`— switched to ${mode} mode —`);
});

$('scenario').addEventListener('change', () => stopRun(false));

renderBoard();
