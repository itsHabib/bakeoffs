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
  events.addEventListener('held', (e) => onHeld(JSON.parse(e.data)));
  events.addEventListener('caught', (e) => onCaught(JSON.parse(e.data)));
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
    $('log').innerHTML = '<p class="empty">Nothing yet. Silence is the default — that\'s the pedagogy.</p>';
    $('report').hidden = true;
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
  $('budgetLabel').textContent = `${s.remaining} left · ${s.held} held`;
  $('allowance').textContent = `${s.total} interruptions`;
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
    ? 'nothing worth a turn'
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

function clearEmpty() {
  const empty = $('log').querySelector('.empty');
  if (empty) empty.remove();
}

function onInterrupt(item) {
  state.marks.push(Date.now());
  const verdict = $('verdictLine');
  verdict.classList.add('live');
  verdict.textContent = `asked: “${item.line}”`;

  clearEmpty();
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
  $('log').prepend(card);

  if ($('speakOut').checked) speak(item.line);
}

// A held opinion gets a dim card keyed by time+line so a later self-correct
// can come back and mark it caught.
function onHeld(item) {
  clearEmpty();
  const card = el('div', 'card held');
  card.dataset.key = heldKey(item);
  const why = el('p', 'why');
  why.textContent = `heard: “${item.heard}”`;
  const line = el('p', 'line');
  line.textContent = `held back: “${item.line}”`;
  const meta = el('div', 'meta');
  meta.append(tag2(item.kind), span(item.reason), span(item.at));
  card.append(why, line, meta);
  $('log').prepend(card);
}

function onCaught(item) {
  const card = $('log').querySelector(`[data-key="${heldKey(item)}"]`);
  if (!card || card.querySelector('.caught')) return;
  const note = el('p', 'caught');
  note.textContent = `✓ student caught it at ${item.caught_at} — no turn needed`;
  card.appendChild(note);
}

const heldKey = (item) => `${item.at}|${item.kind}`;

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
const tag2 = (t) => { const n = el('span', 'tag quiet'); n.textContent = t; return n; };
const fmt = (s) => `${Math.floor(s / 60)}:${String(s % 60).padStart(2, '0')}`;

function setStatus(text, bad) {
  $('status').textContent = text;
  $('status').classList.toggle('bad', !!bad);
}

/* ---------------------------------------------------------------- report */

async function showReport() {
  let report;
  try {
    report = await (await fetch('/api/report')).json();
  } catch (err) {
    setStatus('server unreachable', true);
    return;
  }
  $('reportSub').textContent =
    `${report.elapsed} of working · ${report.spent.length} of ${report.total} turns spent · ${report.held.length} held back`;

  const body = $('reportBody');
  body.innerHTML = '';

  body.appendChild(section('What it said'));
  if (!report.spent.length) body.appendChild(emptyLine('Nothing. It never found a reason worth a turn.'));
  report.spent.forEach((s) => {
    body.appendChild(item(s.at, [
      ['said', `“${s.line}”`],
      ['why', `${s.kind} · worth ${s.pressure} against a ${s.threshold} bar`],
    ]));
  });

  body.appendChild(section('What it chose not to say'));
  if (!report.held.length) body.appendChild(emptyLine('Nothing was held — a quiet session.'));
  report.held.forEach((h) => {
    const rows = [
      ['heard', `heard: “${h.heard}”`],
      ['said', `would have asked: “${h.line}”`],
      ['why', `${h.kind} · held: ${h.reason}`],
    ];
    if (h.kind === 'procedural-slip' || h.kind === 'misconception') {
      rows.push(h.caught_at
        ? ['caught', `✓ student caught it at ${h.caught_at} — the hold paid off`]
        : ['uncaught', `not yet caught — a human tutor might revisit this`]);
    }
    body.appendChild(item(h.at, rows));
  });

  $('report').hidden = false;
}

const section = (title) => { const n = el('div', 'report-section'); n.textContent = title; return n; };
const emptyLine = (text) => { const n = el('p', 'report-empty'); n.textContent = text; return n; };

function item(when, rows) {
  const wrap = el('div', 'report-item');
  const time = el('div', 'when');
  time.textContent = when;
  const what = el('div', 'what');
  rows.forEach(([cls, text]) => {
    const row = el('div', cls);
    row.textContent = text;
    what.appendChild(row);
  });
  wrap.append(time, what);
  return wrap;
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
    setStatus('this browser has no speech recognition — use a rehearsal or type', true);
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

// Three canned students. A: the demo — two slips it rightly sits on (both
// self-caught) and one real misconception it spends a turn on. B: a slip the
// tutor holds through entirely. C: genuinely stuck, one nudge earned by
// persistence.
const REHEARSALS = {
  A: [
    ['Okay. Solve x plus three, quantity squared, equals twenty five.', 4200],
    ['I will expand the left side first. So x plus three, times x plus three.', 4600],
    ['That gives x squared plus three x plus three x plus six.', 5000],
    ['Wait, no. Three times three is nine, not six. So x squared plus six x plus nine.', 5000],
    ['Equals twenty five. Subtract twenty five from both sides. x squared plus six x minus sixteen equals zero.', 5400],
    ['Now factor. Two numbers that multiply to negative sixteen and add to six. Eight and negative two.', 5000],
    ['So it factors as x minus eight, times x plus two.', 4800],
    ['Let me check that. That gives minus eight x plus two x. The signs are backwards. It is x plus eight times x minus two.', 5600],
    ['So x is negative eight, or x is two. Both check out in the original.', 4800],
    ['Next one. Expand x plus y, quantity squared.', 4200],
    ['That one is easy. The square goes onto each piece, so it is x squared plus y squared.', 5200],
    ['So with x equals two and y equals three I get four plus nine, which is thirteen.', 5200],
    ['Hmm. But two plus three is five, and five squared is twenty five. Not thirteen.', 5000],
    ['Wait, that is wrong. I forgot the middle term. It is x squared plus two x y plus y squared, so four plus twelve plus nine. Twenty five.', 5600],
  ],
  B: [
    ['Solve three x plus four x equals twenty one.', 4200],
    ['First I combine the like terms on the left side.', 4400],
    ['Three x plus four x gives twelve x.', 4600],
    ['So twelve x equals twenty one. Hmm, wait. Three plus four is seven, not twelve. It is seven x.', 5200],
    ['Seven x equals twenty one. Divide both sides by seven. So x equals three.', 4800],
    ['Check it. Three times three is nine, plus four times three is twelve. Nine plus twelve is twenty one. Done.', 5200],
  ],
  C: [
    ['Solve x squared minus five x plus six equals zero.', 4200],
    ['Hmm. I do not really know where to start with this one.', 4600],
    ['I could... I do not know. I keep just looking at it.', 4800],
    ['I still do not know what to do. Same as before. I just do not see it.', 5000],
    ['Yeah, I am stuck. Nothing is coming to me.', 4800],
    ['Oh wait. Factor it. Two and three. x minus two, times x minus three.', 5000],
    ['So x is two, or x is three. That was it the whole time.', 4600],
  ],
};

async function rehearse(key, button) {
  if (state.rehearsing) return;
  state.rehearsing = true;
  const label = button.textContent;
  button.textContent = '▶ playing…';
  rehearseButtons().forEach((b) => (b.disabled = b !== button));
  await post('/api/reset');
  await sleep(600);
  for (const [line, pause] of REHEARSALS[key]) {
    if (!state.rehearsing) break;
    await post('/api/hear', { text: line });
    await sleep(pause);
  }
  state.rehearsing = false;
  button.textContent = label;
  rehearseButtons().forEach((b) => (b.disabled = false));
  await sleep(2500); // let the last classification land before the curtain
  showReport();
}

const rehearseButtons = () => [$('rehearseA'), $('rehearseB'), $('rehearseC')];
const sleep = (ms) => new Promise((r) => setTimeout(r, ms));

/* ------------------------------------------------------------------ wire */

$('micBtn').onclick = () => (state.listening ? stopMic() : startMic());
$('rehearseA').onclick = () => rehearse('A', $('rehearseA'));
$('rehearseB').onclick = () => rehearse('B', $('rehearseB'));
$('rehearseC').onclick = () => rehearse('C', $('rehearseC'));
$('resetBtn').onclick = () => post('/api/reset');
$('reportBtn').onclick = showReport;
$('reportClose').onclick = () => ($('report').hidden = true);
$('report').onclick = (e) => { if (e.target === $('report')) $('report').hidden = true; };
$('typeForm').onsubmit = (e) => {
  e.preventDefault();
  const text = $('typeInput').value.trim();
  if (!text) return;
  post('/api/hear', { text });
  $('typeInput').value = '';
};

connect();
draw();
