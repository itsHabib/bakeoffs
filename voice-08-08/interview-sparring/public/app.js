import {InterviewMachine, runCannedReplay} from "/interview-core.js";

const elements = {
  arena: document.querySelector("#arena"),
  debrief: document.querySelector("#debrief"),
  startLive: document.querySelector("#start-live"),
  startReplay: document.querySelector("#start-replay"),
  disconnect: document.querySelector("#disconnect-live"),
  statusDot: document.querySelector("#status-dot"),
  connectionStatus: document.querySelector("#connection-status"),
  questionCounter: document.querySelector("#question-counter"),
  prompt: document.querySelector("#current-prompt"),
  timer: document.querySelector("#timer"),
  budgetFill: document.querySelector("#budget-fill"),
  reason: document.querySelector("#followup-reason"),
  signalScore: document.querySelector("#signal-score"),
  filler: document.querySelector("#filler-metric"),
  overlap: document.querySelector("#overlap-metric"),
  time: document.querySelector("#time-metric"),
  transcript: document.querySelector("#transcript"),
  error: document.querySelector("#error"),
  averageScore: document.querySelector("#average-score"),
  scorecards: document.querySelector("#scorecards"),
  worstDodge: document.querySelector("#worst-dodge"),
  tomorrow: document.querySelector("#tomorrow-list"),
};

let replayGeneration = 0;
let machine = null;
let peer = null;
let channel = null;
let microphone = null;
let remoteAudio = null;
let answerStartedAt = null;
let answerTimer = null;
let cutoffTimer = null;
let cutoffTriggered = false;
let responsePurpose = "";

elements.startReplay.addEventListener("click", startReplay);
elements.startLive.addEventListener("click", startLive);
elements.disconnect.addEventListener("click", () => disconnectLive("Live round ended."));

function setStatus(state, text) {
  elements.statusDot.dataset.state = state;
  elements.connectionStatus.textContent = text;
}

function clearTranscript() {
  elements.transcript.replaceChildren();
}

function addTranscript(speaker, text, kind = "") {
  const row = document.createElement("div");
  row.className = `transcript-entry ${kind}`.trim();
  const label = document.createElement("strong");
  label.textContent = speaker;
  const copy = document.createElement("p");
  copy.textContent = text;
  row.append(label, copy);
  elements.transcript.append(row);
  elements.transcript.scrollTop = elements.transcript.scrollHeight;
}

function showPrompt(decision) {
  const prompt = decision.type === "followup" ? decision.followUp.prompt : decision.question.prompt;
  elements.prompt.textContent = prompt;
  if (decision.type === "question") {
    elements.questionCounter.textContent = `${decision.number} / ${decision.total}`;
    elements.reason.hidden = true;
  } else {
    elements.reason.hidden = false;
    elements.reason.textContent = `${decision.followUp.label} → FOLLOW-UP SELECTED IN CODE`;
    addTranscript("RULE FIRED", `${decision.followUp.label}: ${prompt}`, "rule");
  }
  resetTimer(decision.question.budgetSeconds);
  addTranscript("INTERVIEWER", prompt);
}

function renderMetrics(metrics) {
  if (!metrics) return;
  for (const [beat, hit] of Object.entries(metrics.star)) {
    document.querySelector(`[data-beat="${beat}"]`).classList.toggle("hit", hit);
  }
  elements.signalScore.textContent = `${metrics.score}/100`;
  elements.filler.textContent = `${(metrics.fillerDensity * 100).toFixed(1)}% · ${metrics.fillerCount}`;
  elements.overlap.textContent = metrics.dodge ? "DODGE" : `${metrics.overlapTerms.length} term${metrics.overlapTerms.length === 1 ? "" : "s"}`;
  elements.time.textContent = `${metrics.durationSeconds}s / ${metrics.budgetSeconds}s`;
  const ratio = Math.min(metrics.budgetRatio, 1.25) / 1.25;
  elements.budgetFill.style.width = `${ratio * 100}%`;
  elements.budgetFill.classList.toggle("over", !metrics.withinBudget);
  elements.timer.textContent = formatTime(Math.ceil(metrics.durationSeconds));
}

function resetMetrics() {
  document.querySelectorAll("[data-beat]").forEach((item) => item.classList.remove("hit"));
  elements.signalScore.textContent = "—";
  elements.filler.textContent = "—";
  elements.overlap.textContent = "—";
  elements.time.textContent = "—";
  elements.reason.hidden = true;
  elements.debrief.hidden = true;
}

function resetTimer(seconds = 45) {
  clearInterval(answerTimer);
  elements.timer.textContent = formatTime(seconds);
  elements.budgetFill.style.width = "0%";
  elements.budgetFill.classList.remove("over");
}

function formatTime(totalSeconds) {
  const value = Math.max(0, Math.round(totalSeconds));
  return `${String(Math.floor(value / 60)).padStart(2, "0")}:${String(value % 60).padStart(2, "0")}`;
}

async function startReplay() {
  await disconnectLive("");
  const generation = ++replayGeneration;
  elements.error.textContent = "";
  elements.startReplay.disabled = true;
  elements.startLive.disabled = true;
  clearTranscript();
  resetMetrics();
  setStatus("working", "Canned interview replay · no mic, no key");
  elements.arena.scrollIntoView({behavior: "smooth", block: "start"});
  try {
    const replay = runCannedReplay();
    for (const event of replay.events) {
      if (generation !== replayGeneration) return;
      renderReplayEvent(event);
      await delay(event.type === "answer" ? 900 : 650);
    }
    if (generation !== replayGeneration) return;
    renderDebrief(replay.debrief);
    setStatus("live", "Replay complete · debrief computed");
    elements.debrief.scrollIntoView({behavior: "smooth", block: "start"});
  } catch (error) {
    elements.error.textContent = error.message;
    setStatus("error", "Replay failed");
  } finally {
    elements.startReplay.disabled = false;
    elements.startLive.disabled = false;
  }
}

function renderReplayEvent(event) {
  if (event.type === "question" || event.type === "followup") {
    showPrompt(event);
    if (event.completed?.metrics) renderMetrics(event.completed.metrics);
    return;
  }
  if (event.type === "answer") {
    addTranscript("CANDIDATE", event.transcript);
    if (event.wasInterrupted) addTranscript("45s CUTOFF", "Answer budget exceeded. The state machine stopped the ramble.", "rule");
    return;
  }
  if (event.metrics) renderMetrics(event.metrics);
}

async function startLive() {
  replayGeneration += 1;
  await disconnectLive("");
  elements.error.textContent = "";
  elements.startLive.disabled = true;
  elements.startReplay.disabled = true;
  clearTranscript();
  resetMetrics();
  setStatus("working", "Requesting microphone…");
  elements.arena.scrollIntoView({behavior: "smooth", block: "start"});
  try {
    if (!navigator.mediaDevices?.getUserMedia || !window.RTCPeerConnection) {
      throw new Error("This browser does not support microphone WebRTC.");
    }
    microphone = await navigator.mediaDevices.getUserMedia({audio: {echoCancellation: true, noiseSuppression: true}});
    setStatus("working", "Minting short-lived voice secret…");
    const tokenResponse = await fetch("/api/realtime/token", {method: "POST"});
    const token = await tokenResponse.json();
    if (!tokenResponse.ok) throw new Error(token.error || "Could not start Realtime voice.");
    machine = new InterviewMachine({questionIds: ["leadership", "conflict", "failure"]});
    await connectRealtime(token);
    elements.disconnect.hidden = false;
    setStatus("live", "Realtime interviewer connected");
  } catch (error) {
    elements.error.textContent = error.name === "NotAllowedError"
      ? "Microphone permission was blocked. Allow it and try again—or use the no-mic replay."
      : error.message;
    setStatus("error", "Live voice did not start");
    await disconnectLive("");
    elements.startLive.disabled = false;
    elements.startReplay.disabled = false;
  }
}

async function connectRealtime(token) {
  peer = new RTCPeerConnection();
  remoteAudio = document.createElement("audio");
  remoteAudio.autoplay = true;
  remoteAudio.playsInline = true;
  remoteAudio.hidden = true;
  document.body.append(remoteAudio);
  peer.ontrack = (event) => {
    remoteAudio.srcObject = event.streams[0];
    void remoteAudio.play().catch(() => {
      elements.error.textContent = "Voice connected, but speaker playback is blocked. Allow sound and retry.";
    });
  };
  peer.addTrack(microphone.getAudioTracks()[0], microphone);
  channel = peer.createDataChannel("oai-events");
  channel.addEventListener("message", handleRealtimeEvent);
  peer.addEventListener("connectionstatechange", () => {
    if (["failed", "disconnected"].includes(peer?.connectionState)) {
      elements.error.textContent = "The live voice connection ended. The canned replay is still available.";
      void disconnectLive("Voice connection ended.");
    }
  });
  const offer = await peer.createOffer();
  await peer.setLocalDescription(offer);
  const response = await fetch(token.calls_url, {
    method: "POST",
    body: offer.sdp,
    headers: {Authorization: `Bearer ${token.value}`, "Content-Type": "application/sdp"},
  });
  if (!response.ok) throw new Error(`Realtime connection failed (${response.status}).`);
  await peer.setRemoteDescription({type: "answer", sdp: await response.text()});
  await waitForChannel(channel, 12_000);
}

function waitForChannel(target, timeoutMilliseconds) {
  if (target.readyState === "open") return Promise.resolve();
  return new Promise((resolve, reject) => {
    const timer = setTimeout(() => reject(new Error("Realtime voice channel timed out.")), timeoutMilliseconds);
    target.addEventListener("open", () => { clearTimeout(timer); resolve(); }, {once: true});
    target.addEventListener("close", () => { clearTimeout(timer); reject(new Error("Realtime voice channel closed.")); }, {once: true});
  });
}

function handleRealtimeEvent(message) {
  const event = JSON.parse(message.data);
  if (event.type === "session.created") {
    const decision = machine.start();
    showPrompt(decision);
    speakPrompt(decision);
    return;
  }
  if (event.type === "input_audio_buffer.speech_started") {
    if (!["question", "followup"].includes(machine?.phase)) return;
    machine.beginAnswer(performance.now());
    answerStartedAt = performance.now();
    cutoffTriggered = false;
    startAnswerClock(machine.currentQuestion.budgetSeconds);
    setStatus("live", "Listening · deterministic timer running");
    return;
  }
  if (event.type === "conversation.item.input_audio_transcription.completed") {
    completeVoiceAnswer(event.transcript || "");
    return;
  }
  if (event.type === "response.done") {
    if (["failed", "incomplete"].includes(event.response?.status)) {
      elements.error.textContent = "The interviewer could not complete that voice turn.";
      return;
    }
    if (cutoffTriggered && microphone) {
      microphone.getAudioTracks().forEach((track) => { track.enabled = true; });
      cutoffTriggered = false;
    }
    if (responsePurpose === "debrief") setStatus("live", "Round complete · debrief ready");
    else setStatus("live", "Your turn");
    return;
  }
  if (event.type === "error") {
    elements.error.textContent = event.error?.message || "Realtime voice reported an error.";
  }
}

function startAnswerClock(budgetSeconds) {
  clearInterval(answerTimer);
  clearTimeout(cutoffTimer);
  answerTimer = setInterval(() => {
    const elapsed = (performance.now() - answerStartedAt) / 1000;
    elements.timer.textContent = formatTime(Math.max(0, budgetSeconds - elapsed));
    elements.budgetFill.style.width = `${Math.min(100, (elapsed / budgetSeconds) * 100)}%`;
  }, 100);
  cutoffTimer = setTimeout(() => {
    cutoffTriggered = true;
    microphone?.getAudioTracks().forEach((track) => { track.enabled = false; });
    elements.budgetFill.classList.add("over");
    elements.timer.textContent = "00:00";
    addTranscript("45s CUTOFF", "Time budget reached. Microphone paused so the state machine can press for the missing action.", "rule");
    setStatus("working", "Time budget reached · cutting in");
  }, budgetSeconds * 1000);
}

function completeVoiceAnswer(transcript) {
  if (machine?.phase !== "answering") return;
  clearInterval(answerTimer);
  clearTimeout(cutoffTimer);
  const elapsed = Math.max(0.1, (performance.now() - answerStartedAt) / 1000);
  const duration = cutoffTriggered ? Math.max(elapsed, machine.currentQuestion.budgetSeconds + 0.1) : elapsed;
  addTranscript("YOU", transcript || "[No transcript detected]");
  const decision = machine.submitAnswer({transcript, durationSeconds: duration, wasInterrupted: cutoffTriggered});
  if (decision.metrics) renderMetrics(decision.metrics);
  if (decision.completed?.metrics) renderMetrics(decision.completed.metrics);
  if (decision.type === "debrief") {
    renderDebrief(decision.debrief);
    responsePurpose = "debrief";
    sendVoiceResponse("Say exactly one sentence: Round complete. Your code-graded debrief is on screen.");
    elements.debrief.scrollIntoView({behavior: "smooth", block: "start"});
    return;
  }
  showPrompt(decision);
  speakPrompt(decision);
}

function speakPrompt(decision) {
  responsePurpose = decision.type;
  const prompt = decision.type === "followup" ? decision.followUp.prompt : decision.question.prompt;
  const instruction = decision.type === "followup"
    ? `The deterministic controller selected ${decision.followUp.label}. Say this exact follow-up, with firm natural emphasis, then stop: ${JSON.stringify(prompt)}`
    : `Ask this exact interview question, naturally and without adding commentary, then stop: ${JSON.stringify(prompt)}`;
  sendVoiceResponse(instruction);
}

function sendVoiceResponse(instructions) {
  if (!channel || channel.readyState !== "open") throw new Error("Realtime control channel is not open.");
  channel.send(JSON.stringify({
    type: "response.create",
    response: {output_modalities: ["audio"], instructions},
  }));
}

async function disconnectLive(message) {
  clearInterval(answerTimer);
  clearTimeout(cutoffTimer);
  channel?.close();
  peer?.close();
  microphone?.getTracks().forEach((track) => track.stop());
  if (remoteAudio) remoteAudio.remove();
  channel = null;
  peer = null;
  microphone = null;
  remoteAudio = null;
  machine = null;
  answerStartedAt = null;
  cutoffTriggered = false;
  elements.disconnect.hidden = true;
  elements.startLive.disabled = false;
  elements.startReplay.disabled = false;
  if (message) setStatus("idle", message);
}

function renderDebrief(debrief) {
  elements.debrief.hidden = false;
  elements.averageScore.textContent = debrief.averageScore;
  elements.scorecards.replaceChildren();
  for (const card of debrief.scorecards) {
    const row = document.createElement("article");
    row.className = "scorecard";
    const title = document.createElement("div");
    const heading = document.createElement("h3");
    heading.textContent = card.prompt;
    const detail = document.createElement("small");
    detail.textContent = `${card.metrics.starCount}/4 STAR · ${card.metrics.fillerCount} fillers · ${card.followUps.length} follow-up${card.followUps.length === 1 ? "" : "s"}`;
    title.append(heading, detail);
    const time = document.createElement("div");
    time.className = "score-time";
    const bar = document.createElement("div");
    bar.className = "bar";
    const fill = document.createElement("span");
    fill.style.width = `${Math.min(100, (card.metrics.durationSeconds / Math.max(card.metrics.durationSeconds, card.metrics.budgetSeconds)) * 100)}%`;
    if (!card.metrics.withinBudget) fill.style.background = "var(--red)";
    const budgetLine = document.createElement("i");
    budgetLine.style.left = `${Math.min(100, (card.metrics.budgetSeconds / Math.max(card.metrics.durationSeconds, card.metrics.budgetSeconds)) * 100)}%`;
    bar.append(fill, budgetLine);
    const labels = document.createElement("div");
    labels.className = "bar-labels";
    labels.innerHTML = `<span>${card.metrics.durationSeconds}s spoken</span><span>${card.metrics.budgetSeconds}s budget</span>`;
    time.append(bar, labels);
    const score = document.createElement("strong");
    score.textContent = card.score;
    row.append(title, time, score);
    elements.scorecards.append(row);
  }
  elements.worstDodge.textContent = debrief.worstDodge
    ? `“${debrief.worstDodge.quote}”`
    : "No answer met the deterministic dodge rule.";
  elements.tomorrow.replaceChildren();
  for (const item of debrief.tomorrow) {
    const row = document.createElement("li");
    row.textContent = `${item.prompt} (${item.score}/100)`;
    elements.tomorrow.append(row);
  }
}

function delay(milliseconds) {
  return new Promise((resolve) => setTimeout(resolve, milliseconds));
}
