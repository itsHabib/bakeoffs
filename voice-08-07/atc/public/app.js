import { grade } from './grader.js';
import { SCENARIO, CANNED } from './scenario.js';

const $ = id => document.getElementById(id);
const log = $('log'), elementsEl = $('elements'), phaseEl = $('phase');
const flagEl = $('flag'), sayEl = $('say');
const btnTx = $('btn-tx'), btnMic = $('btn-mic'), btnDemo = $('btn-demo'), btnReset = $('btn-reset');

let idx = -1;            // current exchange
let results = [];        // {exchange, said, result}
let cannedRunning = false;

$('atis').textContent = `${SCENARIO.aircraft} · ${SCENARIO.atis}`;

// ------------------------------------------------------------ radio audio

const audio = { ctx: null };
function chirp() {
  try {
    audio.ctx ??= new (window.AudioContext || window.webkitAudioContext)();
    const ctx = audio.ctx, len = ctx.sampleRate * 0.07;
    const buf = ctx.createBuffer(1, len, ctx.sampleRate);
    const d = buf.getChannelData(0);
    for (let i = 0; i < len; i++) d[i] = (Math.random() * 2 - 1) * (1 - i / len) * 0.25;
    const src = ctx.createBufferSource();
    src.buffer = buf;
    const bp = ctx.createBiquadFilter();
    bp.type = 'bandpass'; bp.frequency.value = 1800; bp.Q.value = 0.8;
    src.connect(bp).connect(ctx.destination);
    src.start();
  } catch { /* audio is charm, not required */ }
}

function speak(text) {
  return new Promise(resolve => {
    if (!('speechSynthesis' in window)) return resolve();
    speechSynthesis.cancel();
    const u = new SpeechSynthesisUtterance(text);
    u.rate = 1.25; u.pitch = 0.85;
    const fallback = setTimeout(resolve, 1000 + text.length * 75);
    u.onend = () => { clearTimeout(fallback); resolve(); };
    chirp();
    speechSynthesis.speak(u);
  });
}

// ------------------------------------------------------------ transcript UI

function addLine(cls, who, text, extra) {
  const div = document.createElement('div');
  div.className = `line ${cls}`;
  if (who) {
    const w = document.createElement('span');
    w.className = 'who'; w.textContent = who;
    div.appendChild(w);
  }
  div.appendChild(document.createTextNode(text));
  if (extra) div.appendChild(extra);
  log.appendChild(div);
  log.scrollTop = log.scrollHeight;
  return div;
}

function gradeStrip(result) {
  const strip = document.createElement('div');
  strip.className = 'grade-strip';
  for (const el of result.elements) {
    if (el.optional && el.status !== 'ok') continue;
    const chip = document.createElement('span');
    chip.className = `chip ${el.status}`;
    chip.textContent =
      el.status === 'ok' ? `✓ ${el.label}` :
      el.status === 'wrong' ? `✗ ${el.label} — heard "${el.heard}"` :
      `∅ ${el.label}`;
    strip.appendChild(chip);
  }
  return strip;
}

// ------------------------------------------------------------ grader panel

function renderPanel(result) {
  const exchange = SCENARIO.exchanges[idx];
  if (!exchange) return;
  phaseEl.textContent = exchange.phase;
  elementsEl.replaceChildren();
  for (const el of exchange.elements) {
    const li = document.createElement('li');
    const graded = result?.elements.find(e => e.key === el.key);
    const status = graded?.status ?? 'pending';
    li.className = `${status}${el.optional ? ' optional' : ''}`;
    const label = document.createElement('span');
    label.textContent = el.label + (el.optional ? ' (optional)' : '');
    const verdict = document.createElement('span');
    verdict.className = 'verdict';
    verdict.textContent =
      status === 'ok' ? '✓ READ BACK' :
      status === 'wrong' ? `✗ WRONG — "${graded.heard}"` :
      status === 'missing' ? '∅ MISSING' : '· · ·';
    li.append(label, verdict);
    elementsEl.appendChild(li);
  }
  const showFlag = result?.ackOnly;
  flagEl.hidden = !showFlag;
  if (showFlag) flagEl.textContent = '⚠ "Roger" is not a readback — say the elements back.';
}

sayEl.addEventListener('input', () => {
  if (idx < 0 || idx >= SCENARIO.exchanges.length) return;
  const text = sayEl.value.trim();
  renderPanel(text ? grade(text, SCENARIO.exchanges[idx]) : null);
});

// ------------------------------------------------------------ flow

async function presentExchange() {
  const exchange = SCENARIO.exchanges[idx];
  if (exchange.preamble) addLine('sys', '', exchange.preamble);
  renderPanel(null);
  addLine('tower', 'MILLBROOK ' + exchange.phase.split(' ')[0], exchange.controller);
  await speak(exchange.tts);
  if (!cannedRunning) { setInputEnabled(true); sayEl.focus(); }
}

function setInputEnabled(on) {
  sayEl.disabled = !on; btnTx.disabled = !on; btnMic.disabled = !on;
}

async function transmit(text, note) {
  if (idx < 0 || idx >= SCENARIO.exchanges.length || !text.trim()) return;
  const exchange = SCENARIO.exchanges[idx];
  const result = grade(text, exchange);
  results.push({ exchange, said: text, result });
  setInputEnabled(false);
  stopMic();

  renderPanel(result);
  const line = addLine('you', 'YOU — SKYHAWK 123AB', text, gradeStrip(result));
  if (result.ackOnly) {
    const n = document.createElement('span');
    n.className = 'note';
    n.textContent = '⚠ "Roger" is not a readback — required elements were not read back.';
    line.appendChild(n);
  } else if (note) {
    const n = document.createElement('span');
    n.className = 'note'; n.textContent = `(scripted student: ${note})`;
    line.appendChild(n);
  }
  sayEl.value = '';

  await new Promise(r => setTimeout(r, cannedRunning ? 1600 : 600));
  idx += 1;
  if (idx < SCENARIO.exchanges.length) return presentExchange();
  finishSession();
}

btnTx.addEventListener('click', () => transmit(sayEl.value));
sayEl.addEventListener('keydown', e => { if (e.key === 'Enter') transmit(sayEl.value); });

// ------------------------------------------------------------ summary

function finishSession() {
  setInputEnabled(false);
  const pct = Math.round(100 * results.reduce((s, r) => s + r.result.score, 0) / results.length);

  const card = document.createElement('div');
  card.className = 'summary';
  card.innerHTML = `<h3>SESSION DEBRIEF</h3><div class="total">${pct}% readback accuracy</div>`;

  const table = document.createElement('table');
  for (const r of results) {
    const tr = document.createElement('tr');
    const pctRow = Math.round(r.result.score * 100);
    const cls = pctRow === 100 ? '' : pctRow >= 60 ? 'mid' : 'bad';
    tr.innerHTML = `<td>${r.exchange.phase}</td>
      <td><span class="bar ${cls}" style="width:${Math.max(6, pctRow * 0.9)}px"></span> ${pctRow}%</td>`;
    table.appendChild(tr);
  }
  card.appendChild(table);

  const worst = [...results].filter(r => r.result.score < 1)
    .sort((a, b) => a.result.score - b.result.score).slice(0, 3);
  if (worst.length) {
    const w = document.createElement('div');
    w.className = 'worst';
    w.innerHTML = '<h4>WORST READBACKS — WHAT THE CALL SHOULD HAVE BEEN</h4>';
    for (const r of worst) {
      const item = document.createElement('div');
      item.className = 'item';
      const bad = r.result.elements.filter(e => !e.optional && e.status !== 'ok')
        .map(e => e.status === 'wrong' ? `${e.label}: heard "${e.heard}"` : `${e.label}: missing`)
        .join(' · ');
      item.innerHTML = `<div>${r.exchange.phase} — <span class="said">"${r.said}"</span></div>
        <div>${bad}</div>
        <div class="fix">✓ Correct: "${r.exchange.solution}"</div>`;
      w.appendChild(item);
    }
    card.appendChild(w);
  }

  const wrap = document.createElement('div');
  wrap.className = 'line';
  wrap.appendChild(card);
  log.appendChild(wrap);
  log.scrollTop = log.scrollHeight;

  btnReset.hidden = false;
  btnDemo.disabled = false;
  cannedRunning = false;
  speak(`Session complete. Readback accuracy ${pct} percent.`);
}

// ------------------------------------------------------------ canned demo

btnDemo.addEventListener('click', async () => {
  reset();
  cannedRunning = true;
  btnDemo.disabled = true;
  addLine('sys', '', '▶ CANNED DEMO — a scripted student flies the pattern and makes three classic readback mistakes. Watch the grader catch each one live. No mic involved.');
  idx = 0;
  await presentExchange();
  for (let i = 0; i < CANNED.length && cannedRunning; i++) {
    await typeInto(CANNED[i].text);
    await new Promise(r => setTimeout(r, 500));
    await transmit(CANNED[i].text, CANNED[i].note);
  }
});

// Types the scripted answer into the live input so the judge watches the
// grader panel flip element-by-element mid-sentence — same code path as a
// human typing or speaking.
async function typeInto(text) {
  sayEl.value = '';
  for (const ch of text) {
    sayEl.value += ch;
    sayEl.dispatchEvent(new Event('input'));
    await new Promise(r => setTimeout(r, 34));
  }
}

// ------------------------------------------------------------ mic (live path)

let rec = null;
const SR = window.SpeechRecognition || window.webkitSpeechRecognition;
if (!SR) { btnMic.title = 'Speech input needs Chrome or Safari — typing works everywhere'; }

btnMic.addEventListener('click', () => {
  if (!SR) return addLine('sys', '', 'Mic mode needs Chrome/Safari (Web Speech API). Typing works everywhere.');
  if (rec) return stopMic();
  rec = new SR();
  rec.interimResults = true;
  rec.continuous = false;
  rec.lang = 'en-US';
  btnMic.classList.add('listening');
  rec.onresult = e => {
    const text = [...e.results].map(r => r[0].transcript).join(' ');
    sayEl.value = text;
    sayEl.dispatchEvent(new Event('input'));   // live grading mid-sentence
    if (e.results[e.results.length - 1].isFinal) transmit(text);
  };
  rec.onend = () => stopMic();
  rec.onerror = () => stopMic();
  rec.start();
});

function stopMic() {
  btnMic.classList.remove('listening');
  if (rec) { try { rec.stop(); } catch { /* already stopped */ } rec = null; }
}

// ------------------------------------------------------------ session start

function reset() {
  speechSynthesis?.cancel();
  cannedRunning = false;
  idx = -1; results = [];
  log.replaceChildren();
  elementsEl.replaceChildren();
  flagEl.hidden = true;
  btnReset.hidden = true;
  btnDemo.disabled = false;
  phaseEl.textContent = '';
  sayEl.value = '';
  addLine('sys', '', SCENARIO.intro);
  addLine('sys', '', 'Press ▶ Fly the script for the hands-free demo, or transmit your own readbacks below (type, or 🎙 in Chrome/Safari).');
  setInputEnabled(false);
}

btnReset.addEventListener('click', () => { reset(); start(); });

async function start() {
  idx = 0;
  await presentExchange();
}

reset();
start();
