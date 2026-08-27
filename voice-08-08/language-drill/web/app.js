const state = {
  pack: [], fixtures: [], queue: [], current: null, attempts: [], recycled: 0,
  recognition: null, realtime: null, demoRunning: false,
};

const $ = (selector) => document.querySelector(selector);
const sleep = (ms) => new Promise((resolve) => setTimeout(resolve, ms));

async function load() {
  const [packResponse, fixtureResponse] = await Promise.all([fetch('/api/pack'), fetch('/api/canned')]);
  state.pack = await packResponse.json();
  state.fixtures = await fixtureResponse.json();
  setupRecognition();
  bind();
  restart();
}

function bind() {
  $('#attemptForm').addEventListener('submit', (event) => {
    event.preventDefault();
    const attempt = $('#attempt').value.trim();
    if (attempt) scoreAttempt(attempt);
  });
  $('#speak').addEventListener('click', listen);
  $('#hearPhrase').addEventListener('click', hearCurrent);
  $('#next').addEventListener('click', nextPhrase);
  $('#restart').addEventListener('click', restart);
  $('#runDemo').addEventListener('click', runCannedDemo);
  $('#connectVoice').addEventListener('click', connectRealtime);
}

function restart() {
  state.queue = state.pack.slice(0, 10).map((phrase) => phrase.id);
  state.attempts = [];
  state.recycled = 0;
  $('#history').innerHTML = '<p>Your scored attempts will appear here.</p>';
  $('#history').classList.add('empty');
  $('#tomorrow').classList.add('hidden');
  $('#result').classList.add('hidden');
  $('#next').textContent = 'Next phrase';
  $('#next').onclick = nextPhrase;
  showPhrase(state.queue.shift());
  updateMetrics();
}

function showPhrase(id) {
  state.current = state.pack.find((phrase) => phrase.id === id) || state.pack[0];
  const index = state.pack.indexOf(state.current) + 1;
  $('#phraseNumber').textContent = `PHRASE ${String(index).padStart(2, '0')}`;
  $('#category').textContent = categoryFor(index);
  $('#target').textContent = state.current.target;
  $('#english').textContent = state.current.english;
  $('#attempt').value = '';
  $('#attempt').placeholder = state.current.target.replaceAll(/[¿?¡!,]/g, '');
  $('#result').classList.add('hidden');
  $('#recognitionState').textContent = 'ready';
}

function categoryFor(index) {
  if (index <= 5) return 'ARRIVING';
  if (index <= 15) return 'CHOOSING';
  if (index <= 25) return 'ORDERING';
  if (index <= 33) return 'FIXING';
  return 'PAYING';
}

async function scoreAttempt(attempt, label = '') {
  const response = await fetch('/api/score', {
    method: 'POST', headers: {'Content-Type': 'application/json'},
    body: JSON.stringify({phrase_id: state.current.id, attempt}),
  });
  const result = await response.json();
  if (!response.ok) throw new Error(result.error || 'Could not score attempt');
  renderResult(result, label);
  state.attempts.push(result);
  if (!result.passed && !state.demoRunning) {
    state.queue.splice(Math.min(2, state.queue.length), 0, result.phrase_id);
    state.recycled += 1;
  }
  renderHistory();
  updateMetrics();
  if ($('#sceneMode').checked && state.realtime?.ready) waiterReply(attempt, result.passed);
  return result;
}

function renderResult(result, label = '') {
  $('#result').classList.remove('hidden');
  $('#score').textContent = `${Math.round(result.score * 100)}%`;
  $('#resultLabel').textContent = label ? label.toUpperCase() : (result.passed ? 'CLEAN READ' : 'RETRY — SHOWING WHY');
  $('#resultLabel').classList.toggle('retry', !result.passed);
  $('#diff').innerHTML = result.diff.map((word) => {
    const main = word.expected || word.heard;
    let sub = '';
    if (word.status === 'changed') sub = `heard: ${word.heard}`;
    if (word.status === 'missing') sub = 'missing';
    if (word.status === 'extra') sub = 'extra';
    return `<span class="word ${word.status}">${escapeHTML(main)}${sub ? `<small>${escapeHTML(sub)}</small>` : ''}</span>`;
  }).join('');
  $('#traps').innerHTML = (result.detected_traps || []).map((trap) =>
    `<div class="trap"><strong>${escapeHTML(trap.name)}</strong>${escapeHTML(trap.detail)}</div>`
  ).join('');
}

function renderHistory() {
  const history = $('#history');
  history.classList.remove('empty');
  history.innerHTML = state.attempts.slice().reverse().map((attempt) => {
    const phrase = state.pack.find((item) => item.id === attempt.phrase_id);
    return `<div class="history-row ${attempt.passed ? '' : 'fail'}">
      <i>${attempt.passed ? '✓' : '↺'}</i>
      <div><b>${escapeHTML(phrase?.target || attempt.target)}</b><small>${escapeHTML(attempt.attempt)}</small></div>
      <strong>${Math.round(attempt.score * 100)}%</strong>
    </div>`;
  }).join('');
}

function nextPhrase() {
  if (state.demoRunning) return;
  const id = state.queue.shift();
  if (!id || state.attempts.length >= 16) return finishSession();
  showPhrase(id);
  updateMetrics();
}

function finishSession() {
  const misses = [...new Set(state.attempts.filter((attempt) => !attempt.passed).map((attempt) => attempt.phrase_id))];
  const tomorrow = [...misses, ...state.pack.slice(10).map((phrase) => phrase.id)].slice(0, 10);
  $('#tomorrowCount').textContent = `${tomorrow.length} phrases`;
  $('#tomorrow').classList.remove('hidden');
  $('#resultLabel').textContent = 'REP COMPLETE';
  $('#next').textContent = 'Restart rep';
  $('#next').onclick = restart;
}

function updateMetrics() {
  $('#repCount').textContent = `${String(Math.min(state.attempts.length + 1, 10)).padStart(2, '0')} / 10`;
  $('#cleanCount').textContent = state.attempts.filter((attempt) => attempt.passed).length;
  $('#recycleCount').textContent = state.recycled;
}

async function runCannedDemo() {
  if (state.demoRunning) return;
  state.demoRunning = true;
  state.attempts = [];
  state.recycled = 0;
  $('#runDemo').disabled = true;
  $('#runDemo').textContent = 'Demo running…';
  $('#history').classList.remove('empty');
  for (const fixture of state.fixtures) {
    showPhrase(fixture.phrase_id);
    $('#attempt').value = fixture.attempt;
    await scoreAttempt(fixture.attempt, fixture.label);
    await sleep(950);
  }
  $('#resultLabel').textContent = 'DEMO COMPLETE';
  $('#runDemo').textContent = 'Run it again';
  $('#runDemo').disabled = false;
  state.demoRunning = false;
}

function setupRecognition() {
  const Recognition = window.SpeechRecognition || window.webkitSpeechRecognition;
  if (!Recognition) {
    $('#recognitionState').textContent = 'typing only in this browser';
    return;
  }
  state.recognition = new Recognition();
  state.recognition.lang = 'es-ES';
  state.recognition.interimResults = true;
  state.recognition.continuous = false;
  state.recognition.onstart = () => {
    $('#speak').classList.add('listening');
    $('#recognitionState').textContent = 'listening…';
  };
  state.recognition.onresult = (event) => {
    const transcript = Array.from(event.results).map((result) => result[0].transcript).join(' ');
    $('#attempt').value = transcript;
    if (event.results[event.results.length - 1].isFinal) scoreAttempt(transcript);
  };
  state.recognition.onend = () => {
    $('#speak').classList.remove('listening');
    $('#recognitionState').textContent = 'ready';
  };
  state.recognition.onerror = (event) => {
    $('#speak').classList.remove('listening');
    $('#recognitionState').textContent = event.error === 'not-allowed' ? 'microphone blocked' : 'could not hear that';
  };
}

function listen() {
  if (!state.recognition) return $('#attempt').focus();
  try { state.recognition.start(); } catch (_) { state.recognition.stop(); }
}

function hearCurrent() {
  if (state.realtime?.ready) {
    modelSay(`Say exactly this phrase once, with natural Spanish pronunciation and no other words: ${state.current.target}`);
    return;
  }
  const utterance = new SpeechSynthesisUtterance(state.current.target);
  utterance.lang = 'es-ES';
  utterance.rate = .86;
  speechSynthesis.cancel();
  speechSynthesis.speak(utterance);
}

async function connectRealtime() {
  const button = $('#connectVoice');
  button.disabled = true;
  button.textContent = 'Connecting…';
  let microphone;
  let pc;
  try {
    const tokenResponse = await fetch('/api/realtime-token');
    const token = await tokenResponse.json();
    if (!tokenResponse.ok) throw new Error(token.error || 'Could not mint voice session');
    microphone = await navigator.mediaDevices.getUserMedia({audio: true});
    pc = new RTCPeerConnection();
    pc.addTrack(microphone.getAudioTracks()[0], microphone);
    pc.ontrack = (event) => { $('#remoteAudio').srcObject = event.streams[0]; };
    const dc = pc.createDataChannel('oai-events');
    const offer = await pc.createOffer();
    await pc.setLocalDescription(offer);
    const sdpResponse = await fetch('https://api.openai.com/v1/realtime/calls', {
      method: 'POST', body: offer.sdp,
      headers: {Authorization: `Bearer ${token.value}`, 'Content-Type': 'application/sdp'},
    });
    if (!sdpResponse.ok) throw new Error(`Realtime connection returned ${sdpResponse.status}`);
    await pc.setRemoteDescription({type: 'answer', sdp: await sdpResponse.text()});
    await new Promise((resolve, reject) => {
      const timeout = setTimeout(() => reject(new Error('Voice connection timed out')), 10000);
      if (dc.readyState === 'open') {
        clearTimeout(timeout);
        resolve();
        return;
      }
      dc.addEventListener('open', () => { clearTimeout(timeout); resolve(); }, {once: true});
    });
    pc.onconnectionstatechange = () => {
      if (['failed', 'disconnected', 'closed'].includes(pc.connectionState)) {
        $('#voiceState').classList.remove('live');
        $('#voiceState').textContent = `voice ${pc.connectionState}`;
      }
    };
    state.realtime = {pc, dc, microphone, ready: true};
    $('#voiceState').classList.add('live');
    $('#voiceState').innerHTML = '<i></i> OpenAI voice live';
    button.textContent = 'Voice connected';
    modelSay('Say exactly: Bienvenido a Mesa cuarenta. Empezamos.');
  } catch (error) {
    microphone?.getTracks().forEach((track) => track.stop());
    pc?.close();
    button.disabled = false;
    button.textContent = 'Retry OpenAI voice';
    $('#voiceState').textContent = error.message;
  }
}

function modelSay(instruction) {
  if (!state.realtime?.ready) return;
  state.realtime.dc.send(JSON.stringify({
    type: 'conversation.item.create',
    item: {type: 'message', role: 'user', content: [{type: 'input_text', text: instruction}]},
  }));
  state.realtime.dc.send(JSON.stringify({type: 'response.create', response: {output_modalities: ['audio']}}));
}

function waiterReply(attempt, passed) {
  const direction = passed ? 'Accept the order and ask one short restaurant follow-up.' : 'Do not grade. Ask the customer to repeat once.';
  modelSay(`The customer said: ${attempt}. ${direction} Reply in one short Spanish sentence.`);
}

function escapeHTML(value = '') {
  return value.replace(/[&<>'"]/g, (character) => ({'&':'&amp;','<':'&lt;','>':'&gt;',"'":'&#39;','"':'&quot;'}[character]));
}

load().catch((error) => {
  document.body.innerHTML = `<pre>Could not start Mesa 40: ${escapeHTML(error.message)}</pre>`;
});
