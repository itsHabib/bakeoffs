import {
  FOLLOW_UPS,
  STORY_SPINE,
  assembleChapter,
  trackCoverage,
  verifyProvenance,
} from "/lib/story-core.js";
import { interviewerInstruction } from "/lib/realtime-config.js";
import { FIXTURE } from "/fixture.js";

const elements = {
  replayButton: document.querySelector("#replay-button"),
  liveButton: document.querySelector("#live-button"),
  finishButton: document.querySelector("#finish-button"),
  transcript: document.querySelector("#transcript"),
  status: document.querySelector("#session-status"),
  spineList: document.querySelector("#spine-list"),
  nextQuestion: document.querySelector("#next-question"),
  liveControls: document.querySelector("#live-controls"),
  chapterSection: document.querySelector("#chapter-section"),
  chapterText: document.querySelector("#chapter-text"),
  chapterHeading: document.querySelector("#chapter-heading"),
  provenanceBadge: document.querySelector("#provenance-badge"),
  openQuestions: document.querySelector("#open-questions"),
  remoteAudio: document.querySelector("#remote-audio"),
};

let turns = [];
let replayGeneration = 0;
let peerConnection = null;
let dataChannel = null;
let localStream = null;
let assistantDraft = "";

function sleep(milliseconds) {
  return new Promise((resolve) => window.setTimeout(resolve, milliseconds));
}

function setStatus(label, state = "") {
  elements.status.className = `status ${state}`.trim();
  elements.status.lastChild.textContent = label;
}

function renderSpine() {
  const coverage = trackCoverage(turns);
  elements.spineList.replaceChildren(...coverage.beats.map((beat) => {
    const item = document.createElement("li");
    item.className = `spine-item ${beat.covered ? "covered" : ""} ${coverage.nextBeat?.id === beat.id ? "next" : ""}`.trim();
    const check = document.createElement("span");
    check.className = "beat-check";
    check.textContent = "✓";
    const label = document.createElement("span");
    label.textContent = beat.label;
    item.append(check, label);
    return item;
  }));

  if (coverage.nextBeat) {
    elements.nextQuestion.innerHTML = `<span>Next question</span><p></p>`;
    elements.nextQuestion.querySelector("p").textContent = FOLLOW_UPS[coverage.nextBeat.id];
  } else {
    elements.nextQuestion.innerHTML = "<span>Story spine complete</span><p>All six beats are ready for the chapter.</p>";
  }
  return coverage;
}

function renderTurn(turn) {
  elements.transcript.classList.remove("empty");
  elements.transcript.querySelector(".empty-state")?.remove();
  const row = document.createElement("div");
  row.className = `turn ${turn.speaker}`;
  const avatar = document.createElement("span");
  avatar.className = "avatar";
  avatar.textContent = turn.speaker === "teller" ? "M" : "K";
  const content = document.createElement("div");
  const label = document.createElement("p");
  label.className = "turn-label";
  label.textContent = turn.speaker === "teller" ? "Margaret" : "Keepsake";
  const speech = document.createElement("p");
  speech.textContent = turn.text;
  content.append(label, speech);
  row.append(avatar, content);
  elements.transcript.append(row);
  elements.transcript.scrollTop = elements.transcript.scrollHeight;
}

function resetSession() {
  replayGeneration += 1;
  disconnectLive();
  turns = [];
  elements.transcript.className = "transcript empty";
  elements.transcript.innerHTML = `<div class="empty-state"><span class="sound-rings" aria-hidden="true"><i></i><i></i><i></i></span><p>Getting the story ready…</p></div>`;
  elements.chapterSection.classList.add("hidden");
  elements.liveControls.classList.add("hidden");
  renderSpine();
}

function addTurn(turn) {
  const clean = { speaker: turn.speaker, text: turn.text.trim() };
  if (!clean.text) return trackCoverage(turns);
  turns.push(clean);
  renderTurn(clean);
  return renderSpine();
}

function showChapter() {
  const chapter = assembleChapter(turns);
  const provenance = verifyProvenance(chapter, turns);
  const coverage = trackCoverage(turns);
  elements.chapterHeading.textContent = chapter.title;
  elements.chapterText.replaceChildren(...chapter.sections.map((section) => {
    const paragraph = document.createElement("p");
    paragraph.textContent = section.sentences.join(" ");
    return paragraph;
  }));
  elements.provenanceBadge.textContent = provenance.valid
    ? "✓ Every sentence traced to the interview"
    : "Provenance check failed";
  elements.provenanceBadge.style.color = provenance.valid ? "" : "#843922";

  if (coverage.followUps.length) {
    elements.openQuestions.innerHTML = "<h3>One thread still worth following</h3><p></p>";
    elements.openQuestions.querySelector("p").textContent = coverage.followUps.join(" ");
  } else {
    elements.openQuestions.innerHTML = "<h3>The story spine is complete</h3><p>No essential follow-up questions remain.</p>";
  }
  elements.chapterSection.classList.remove("hidden");
  setStatus(`${coverage.coveredCount} of ${coverage.totalCount} beats captured`, coverage.coveredCount === coverage.totalCount ? "complete" : "");
  elements.chapterSection.scrollIntoView({ behavior: "smooth", block: "start" });
}

async function replayFixture() {
  resetSession();
  const generation = replayGeneration;
  elements.replayButton.disabled = true;
  elements.liveButton.disabled = true;
  setStatus("Replaying", "active");

  for (const turn of FIXTURE) {
    if (generation !== replayGeneration) return;
    await sleep(turn.speaker === "interviewer" ? 560 : 720);
    addTurn(turn);
  }

  if (generation !== replayGeneration) return;
  await sleep(650);
  showChapter();
  elements.replayButton.disabled = false;
  elements.liveButton.disabled = false;
}

function sendEvent(event) {
  if (dataChannel?.readyState === "open") {
    dataChannel.send(JSON.stringify(event));
  }
}

function askForBeat(coverage, lastAnswer = "") {
  sendEvent({
    type: "response.create",
    response: {
      output_modalities: ["audio"],
      instructions: interviewerInstruction(coverage.nextBeat, lastAnswer),
    },
  });
}

function handleRealtimeEvent(event) {
  if (event.type === "conversation.item.input_audio_transcription.completed") {
    const transcript = event.transcript?.trim();
    if (!transcript) return;
    const coverage = addTurn({ speaker: "teller", text: transcript });
    setStatus("Thinking of one follow-up", "active");
    askForBeat(coverage, transcript);
    return;
  }

  if (event.type === "response.output_audio_transcript.delta" || event.type === "response.audio_transcript.delta") {
    assistantDraft += event.delta || "";
    return;
  }

  if (event.type === "response.output_audio_transcript.done" || event.type === "response.audio_transcript.done") {
    const transcript = (event.transcript || assistantDraft).trim();
    assistantDraft = "";
    if (transcript) addTurn({ speaker: "interviewer", text: transcript });
    setStatus("Listening", "active");
    return;
  }

  if (event.type === "error") {
    setStatus(event.error?.message || "Live voice hit a problem");
  }
}

async function startLive() {
  resetSession();
  elements.replayButton.disabled = true;
  elements.liveButton.disabled = true;
  setStatus("Connecting securely", "active");

  try {
    // Ask for the microphone before minting a short-lived credential. A person
    // can take longer than the token lifetime to answer the permission prompt.
    localStream = await navigator.mediaDevices.getUserMedia({ audio: true });

    const secretResponse = await fetch("/api/realtime-token", { method: "POST" });
    const secret = await secretResponse.json();
    if (!secretResponse.ok) throw new Error(secret.error || "Could not start the live interview.");

    peerConnection = new RTCPeerConnection();
    peerConnection.ontrack = (event) => {
      [elements.remoteAudio.srcObject] = event.streams;
    };
    peerConnection.onconnectionstatechange = () => {
      if (peerConnection?.connectionState === "failed") setStatus("Voice connection failed");
    };
    for (const track of localStream.getTracks()) peerConnection.addTrack(track, localStream);

    dataChannel = peerConnection.createDataChannel("oai-events");
    dataChannel.addEventListener("message", (message) => {
      try { handleRealtimeEvent(JSON.parse(message.data)); } catch { /* Ignore malformed upstream events. */ }
    });
    dataChannel.addEventListener("open", () => {
      setStatus("Interviewer is joining", "active");
      elements.liveControls.classList.remove("hidden");
      askForBeat(trackCoverage(turns));
    });

    const offer = await peerConnection.createOffer();
    await peerConnection.setLocalDescription(offer);
    const sdpResponse = await fetch("https://api.openai.com/v1/realtime/calls", {
      method: "POST",
      body: offer.sdp,
      headers: {
        Authorization: `Bearer ${secret.value}`,
        "Content-Type": "application/sdp",
      },
    });
    if (!sdpResponse.ok) throw new Error(`Voice connection was declined (${sdpResponse.status}).`);
    await peerConnection.setRemoteDescription({ type: "answer", sdp: await sdpResponse.text() });
  } catch (error) {
    disconnectLive();
    setStatus(error.message || "Could not start live voice");
    elements.transcript.className = "transcript empty";
    elements.transcript.innerHTML = `<div class="empty-state"><p>${escapeHTML(error.message || "Could not start live voice.")}<br>The no-key replay still works.</p></div>`;
  } finally {
    elements.replayButton.disabled = false;
    elements.liveButton.disabled = false;
  }
}

function escapeHTML(value) {
  const element = document.createElement("span");
  element.textContent = value;
  return element.innerHTML;
}

function disconnectLive() {
  if (dataChannel) dataChannel.close();
  if (peerConnection) peerConnection.close();
  if (localStream) localStream.getTracks().forEach((track) => track.stop());
  dataChannel = null;
  peerConnection = null;
  localStream = null;
  elements.remoteAudio.srcObject = null;
}

elements.replayButton.addEventListener("click", replayFixture);
elements.liveButton.addEventListener("click", startLive);
elements.finishButton.addEventListener("click", () => {
  disconnectLive();
  elements.liveControls.classList.add("hidden");
  showChapter();
});
window.addEventListener("beforeunload", disconnectLive);

renderSpine();
