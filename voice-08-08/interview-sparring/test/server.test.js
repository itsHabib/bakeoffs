import assert from "node:assert/strict";
import {once} from "node:events";
import test from "node:test";
import {createAppServer, mintRealtimeSecret, realtimeSessionPayload} from "../server.js";

test("Realtime payload keeps scoring out of the model and enables semantic VAD", () => {
  const payload = realtimeSessionPayload({});
  assert.equal(payload.session.type, "realtime");
  assert.equal(payload.session.model, "gpt-realtime-2.1");
  assert.equal(payload.session.audio.input.turn_detection.type, "semantic_vad");
  assert.equal(payload.session.audio.input.turn_detection.eagerness, "high");
  assert.equal(payload.session.audio.input.turn_detection.create_response, false);
  assert.equal(payload.session.audio.input.turn_detection.interrupt_response, true);
  assert.match(payload.session.instructions, /deterministic state machine owns every question/);
  assert.match(payload.session.instructions, /Never grade/);
});

test("minting sends the permanent key only in server-side authorization", async () => {
  let captured;
  const fetchImpl = async (url, options) => {
    captured = {url, options};
    return new Response(JSON.stringify({value: "ek_short", expires_at: 1234}), {
      status: 200,
      headers: {"Content-Type": "application/json"},
    });
  };
  const secret = await mintRealtimeSecret({OPENAI_API_KEY: "server-only", OPENAI_REALTIME_MODEL: "test-model"}, fetchImpl);
  assert.equal(secret.value, "ek_short");
  assert.equal(captured.options.headers.Authorization, "Bearer server-only");
  assert.equal(captured.options.headers["OpenAI-Safety-Identifier"], "hack2-interview-sparring-local-demo");
  assert.equal(JSON.parse(captured.options.body).session.model, "test-model");
  assert.doesNotMatch(JSON.stringify(secret), /server-only/);
});

test("minting fails closed without an API key", async () => {
  await assert.rejects(() => mintRealtimeSecret({}, async () => assert.fail("fetch must not run")), {
    message: "Realtime voice is not configured on this server.",
    statusCode: 503,
  });
});

test("HTTP server serves app and shared engine with security headers", async (context) => {
  const server = createAppServer({env: {}});
  server.listen(0, "127.0.0.1");
  await once(server, "listening");
  context.after(() => server.close());
  const {port} = server.address();
  const home = await fetch(`http://127.0.0.1:${port}/`);
  assert.equal(home.status, 200);
  assert.match(await home.text(), /Interview Sparring/);
  assert.match(home.headers.get("content-security-policy"), /https:\/\/api\.openai\.com/);
  assert.equal(home.headers.get("permissions-policy"), "microphone=(self)");
  const core = await fetch(`http://127.0.0.1:${port}/interview-core.js`);
  assert.equal(core.status, 200);
  assert.match(await core.text(), /class InterviewMachine/);
});

test("HTTP token route returns a short-lived secret and never the permanent key", async (context) => {
  const fetchImpl = async () => new Response(JSON.stringify({value: "ek_short", expires_at: 99}), {
    status: 200,
    headers: {"Content-Type": "application/json"},
  });
  const server = createAppServer({env: {OPENAI_API_KEY: "permanent-secret"}, fetchImpl});
  server.listen(0, "127.0.0.1");
  await once(server, "listening");
  context.after(() => server.close());
  const {port} = server.address();
  const response = await fetch(`http://127.0.0.1:${port}/api/realtime/token`, {method: "POST"});
  const body = await response.text();
  assert.equal(response.status, 200);
  assert.match(body, /ek_short/);
  assert.doesNotMatch(body, /permanent-secret/);
});

test("HTTP token route is a clean 503 when voice is not configured", async (context) => {
  const server = createAppServer({env: {}});
  server.listen(0, "127.0.0.1");
  await once(server, "listening");
  context.after(() => server.close());
  const {port} = server.address();
  const response = await fetch(`http://127.0.0.1:${port}/api/realtime/token`, {method: "POST"});
  assert.equal(response.status, 503);
  assert.deepEqual(await response.json(), {error: "Realtime voice is not configured on this server."});
});

test("HTTP server rejects traversal and non-GET static methods", async (context) => {
  const server = createAppServer({env: {}});
  server.listen(0, "127.0.0.1");
  await once(server, "listening");
  context.after(() => server.close());
  const {port} = server.address();
  assert.equal((await fetch(`http://127.0.0.1:${port}/..%2Fserver.js`)).status, 404);
  assert.equal((await fetch(`http://127.0.0.1:${port}/styles.css`, {method: "DELETE"})).status, 405);
});
