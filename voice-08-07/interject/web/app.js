'use strict';

const $ = (id) => document.getElementById(id);
const WINDOW_MS = 90_000;

const state = {
  samples: [],      // {t, pressure, threshold, floor}
  marks: [],        // timestamps of spent turns
  latest: null,
  listening: false,
  rehearsing: false,
};

/* ---------------------------------------------------------------- server */

async function post(path, body) {
  try {
    await fetch(path, {
      method: 'POST',
      headers: { 'Content-Type': 'application/json' },
      body: JSON.stringify(body || {}),
    });
  } catch (err) {
    setStatus('server unreachable', true);
  }
}

function connect() {
  const events = new EventSource('/api/events');
  events.addEventListener('state', (e) => onState(JSON.parse(e.data)));
  events.addEventListener('heard', (e) => addHeard(JSON.parse(e.data).text));
  events.addEventListener('interrupt', (e) => onInterrupt(JSON.parse(e.data)));
  events.addEventListener('reading', (e) => onReading(JSON.parse(e.data)));
  events.addEventListener('thinking', (e) => {
    const busy = JSON.parse(e.data).active;
    $('thinkingDot').textContent = busy ? 'reading…' : 'idle';
    $('thinkingDot').classList.toggle('busy', busy);
  });
  events.addEventListener('fault', (e) => setStatus(JSON.parse(e.data).message, true));
  events.addEventListener('reset', () => {
    state.samples = [];
    state.marks = [];
    $('transcript').innerHTML = '';
    $('log').innerHTML = '<p class="empty">Nothing yet. Silence is the default — that\'s the point.</p>';
  });
  events.onopen = () => setStatus('connected');
  events.onerror = () => setStatus('reconnecting…', true);
}

/* ----------------------------------------------------------------- state */

function onState(s) {
  state.latest = s;
  const now = Date.now();
  state.samples.push({ t: now, pressure: s.pressure, threshold: s.threshold, floor: s.floor });
  state.samples = state.samples.filter((p) => now - p.t < WINDOW_MS);
  state.marks = state.marks.filter((t) => now - t < WINDOW_MS);

  $('pressureValue').textContent = Math.round(s.pressure);
  $('thresholdValue').textContent = Math.round(s.threshold);
  $('budgetLabel').textContent = `${s.remaining} left`;
  $('allowance').textContent = `${s.total} times`;
  $('resetLabel').textContent = s.resets_in > 0 ? `returns in ${fmt(s.resets_in)}` : 'window full';
  $('modelLine').textContent = `${s.model} · bar moves with what's left`;

  const pips = $('pips');
  if (pips.children.length !== s.total) {
    pips.innerHTML = '';
    for (let i = 0; i < s.total; i++) pips.appendChild(el('div', 'pip'));
  }
  [...pips.children].forEach((pip, i) => pip.classList.toggle('spent', i >= s.remaining));
}

function onReading(r) {
  const verdict = $('verdictLine');
  if (r.decision.speak) return;                       // the interrupt card says it better
  verdict.classList.remove('live');
  verdict.textContent = r.kind === 'none'
    ? 'nothing worth saying'
    : `held: ${r.kind} — ${r.decision.reason}`;
}

/* ------------------------------------------------------------------ view */

function addHeard(text) {
  const box = $('transcript');
  [...box.children].forEach((p) => p.classList.remove('fresh'));
  const p = el('p', 'fresh');
  p.textContent = text;
  box.appendChild(p);
  box.scrollTop = box.scrollHeight;
}

function onInterrupt(item) {
  state.marks.push(Date.now());
  const verdict = $('verdictLine');
  verdict.classList.add('live');
  verdict.textContent = `spoke: “${item.line}”`;

  const log = $('log');
  const empty = log.querySelector('.empty');
  if (empty) empty.remove();

  const card = el('div', 'card');
  const line = el('p', 'line');
  line.textContent = `“${item.line}”`;
  const meta = el('div', 'meta');
  meta.append(
    tag(item.kind), span(`${item.conf} confidence`),
    span(`worth ${item.pressure} against a ${item.threshold} bar`), span(item.at),
  );
  const judge = el('div', 'judge');
  judge.append(
    judgeBtn('worth', 'worth it', item.id, card),
    judgeBtn('wasted', 'not worth it', item.id, card),
  );
  card.append(line, meta, judge);
  log.prepend(card);

  if ($('speakOut').checked) speak(item.line);
}

function judgeBtn(verdict, label, id, card) {
  const button = el('button', verdict);
  button.textContent = label;
  button.onclick = () => {
    post('/api/verdict', { id, verdict });
    const note = el('div', 'judged');
    note.textContent = verdict === 'worth'
      ? 'noted — the bar drops a little'
      : 'noted — the bar rises for everything after this';
    card.querySelector('.judge').replaceWith(note);
  };
  return button;
}

const el = (tagName, cls) => { const n = document.createElement(tagName); if (cls) n.className = cls; return n; };
const span = (t) => { const n = el('span'); n.textContent = t; return n; };
const tag = (t) => { const n = el('span', 'tag'); n.textContent = t; return n; };
const fmt = (s) => `${Math.floor(s / 60)}:${String(s % 60).padStart(2, '0')}`;

function setStatus(text, bad) {
  $('status').textContent = text;
  $('status').classList.toggle('bad', !!bad);
}

/* ----------------------------------------------------------------- chart */

const canvas = $('chart');
const ctx = canvas.getContext('2d');

function draw() {
  const ratio = window.devicePixelRatio || 1;
  const w = canvas.clientWidth;
  const h = canvas.clientHeight;
  if (canvas.width !== w * ratio || canvas.height !== h * ratio) {
    canvas.width = w * ratio;
    canvas.height = h * ratio;
  }
  ctx.setTransform(ratio, 0, 0, ratio, 0, 0);
  ctx.clearRect(0, 0, w, h);

  const now = Date.now();
  const x = (t) => w - ((now - t) / WINDOW_MS) * w;
  const y = (v) => h - (v / 100) * (h - 8) - 4;

  ctx.strokeStyle = '#1b212c';
  ctx.lineWidth = 1;
  for (let v = 25; v <= 100; v += 25) {
    ctx.beginPath();
    ctx.moveTo(0, y(v));
    ctx.lineTo(w, y(v));
    ctx.stroke();
  }

  const pts = state.samples;
  if (pts.length > 1) {
    trace(pts, x, y, 'floor', '#3a4356', 1, [3, 4]);
    trace(pts, x, y, 'threshold', '#6ea8ff', 2, [6, 4]);

    ctx.beginPath();
    ctx.moveTo(x(pts[0].t), y(0));
    pts.forEach((p) => ctx.lineTo(x(p.t), y(p.pressure)));
    ctx.lineTo(x(pts[pts.length - 1].t), y(0));
    ctx.closePath();
    const fill = ctx.createLinearGradient(0, 0, 0, h);
    fill.addColorStop(0, 'rgba(255,107,74,.28)');
    fill.addColorStop(1, 'rgba(255,107,74,0)');
    ctx.fillStyle = fill;
    ctx.fill();

    trace(pts, x, y, 'pressure', '#ff6b4a', 2.2, []);
  }

  // Spent turns are drawn faint and dashed so they read as annotations on the
  // trace rather than as pressure that spiked to the ceiling.
  state.marks.forEach((t) => {
    ctx.strokeStyle = 'rgba(255,107,74,.3)';
    ctx.lineWidth = 1;
    ctx.setLineDash([2, 4]);
    ctx.beginPath();
    ctx.moveTo(x(t), 12);
    ctx.lineTo(x(t), h - 2);
    ctx.stroke();
    ctx.setLineDash([]);
    ctx.fillStyle = '#ff6b4a';
    ctx.beginPath();
    ctx.arc(x(t), 7, 3.5, 0, Math.PI * 2);
    ctx.fill();
  });

  requestAnimationFrame(draw);
}

function trace(pts, x, y, key, color, width, dash) {
  ctx.beginPath();
  ctx.setLineDash(dash);
  ctx.strokeStyle = color;
  ctx.lineWidth = width;
  pts.forEach((p, i) => (i ? ctx.lineTo(x(p.t), y(p[key])) : ctx.moveTo(x(p.t), y(p[key]))));
  ctx.stroke();
  ctx.setLineDash([]);
}

/* ------------------------------------------------------------------ mic */

const Recognition = window.SpeechRecognition || window.webkitSpeechRecognition;
let recognition = null;
let holdMic = false;

function startMic() {
  if (!Recognition) {
    setStatus('this browser has no speech recognition — use rehearsal or type', true);
    return;
  }
  recognition = new Recognition();
  recognition.continuous = true;
  recognition.interimResults = true;
  recognition.lang = 'en-US';

  recognition.onresult = (event) => {
    for (let i = event.resultIndex; i < event.results.length; i++) {
      const result = event.results[i];
      if (result.isFinal) post('/api/hear', { text: result[0].transcript });
    }
  };
  recognition.onerror = (event) => {
    if (event.error === 'no-speech' || event.error === 'aborted') return;
    setStatus(`mic: ${event.error}`, true);
  };
  recognition.onend = () => {
    if (state.listening && !holdMic) recognition.start();
  };

  recognition.start();
  state.listening = true;
  $('micBtn').textContent = '⏸ Stop listening';
  $('micBtn').classList.add('live');
  setStatus('listening');
}

function stopMic() {
  state.listening = false;
  if (recognition) recognition.stop();
  $('micBtn').textContent = '🎙 Start listening';
  $('micBtn').classList.remove('live');
  setStatus('stopped');
}

/* --------------------------------------------------------------- speech */

function speak(line) {
  if (!window.speechSynthesis) return;
  holdMic = true;
  if (recognition && state.listening) recognition.stop();   // don't let it hear itself

  const utterance = new SpeechSynthesisUtterance(line);
  utterance.rate = 1.05;
  const release = () => {
    holdMic = false;
    if (state.listening && recognition) {
      try { recognition.start(); } catch (_) { /* already running */ }
    }
  };
  utterance.onend = release;
  utterance.onerror = release;
  window.speechSynthesis.speak(utterance);
}

/* ------------------------------------------------------------ rehearsal */

// A monologue with two things worth saying buried in ordinary thinking. The
// first is cheap and lands; the bar rises behind it; only the more expensive
// one gets through after that. Everything else is met with silence.
const REHEARSAL = [
  ['Okay. The job queue thing. I want this shipped today.', 3200],
  ['Rule for this sprint is no new dependencies. We already have too much surface area to patch.', 4200],
  ['So workers pull a job off the queue, do the work, mark it done. Nothing exotic.', 4000],
  ['The queue table already has a status column, so I can just reuse that.', 3800],
  ['I will add an index on status and created at, keeps the poll cheap.', 4000],
  ['Um, and for the retry backoff I will just pull in that library. It is only two hundred lines.', 4400],
  ['It saves me writing the jitter logic myself.', 3600],
  ['Then the dashboard reads straight off the same table. No extra service.', 4200],
  ['I will put the worker count in the config file. Start with four.', 3800],
  ['Logging, I will just use the standard logger. Nothing fancy.', 3600],
  ['Oh and the poll interval, maybe two seconds? Five feels too slow.', 4000],
  ['And if a write fails halfway through the batch, we just retry the whole batch. It is fine.', 4600],
  ['Since it is all going to the same table anyway.', 3600],
  ['I think that is basically it. Maybe an hour of work, then I will write the tests.', 4200],
];

async function rehearse() {
  if (state.rehearsing) return;
  state.rehearsing = true;
  $('rehearseBtn').textContent = '▶ rehearsing…';
  for (const [line, pause] of REHEARSAL) {
    if (!state.rehearsing) break;
    await post('/api/hear', { text: line });
    await sleep(pause);
  }
  state.rehearsing = false;
  $('rehearseBtn').textContent = '▶ Run the rehearsal';
}

const sleep = (ms) => new Promise((r) => setTimeout(r, ms));

/* ------------------------------------------------------------------ wire */

$('micBtn').onclick = () => (state.listening ? stopMic() : startMic());
$('rehearseBtn').onclick = rehearse;
$('resetBtn').onclick = () => post('/api/reset');
$('typeForm').onsubmit = (e) => {
  e.preventDefault();
  const text = $('typeInput').value.trim();
  if (!text) return;
  post('/api/hear', { text });
  $('typeInput').value = '';
};

connect();
draw();
