import test from "node:test";
import assert from "node:assert/strict";

import { createAppServer } from "./server.mjs";

async function withServer(options, run) {
  const server = createAppServer(options);
  await new Promise((resolve, reject) => {
    server.once("error", reject);
    server.listen(0, "127.0.0.1", resolve);
  });
  const { port } = server.address();
  try {
    await run(`http://127.0.0.1:${port}`);
  } finally {
    await new Promise((resolve, reject) => server.close((error) => error ? reject(error) : resolve()));
  }
}

test("health reports only whether Realtime is configured", async () => {
  await withServer({ apiKey: "server-secret" }, async (baseURL) => {
    const response = await fetch(`${baseURL}/api/health`);
    assert.equal(response.status, 200);
    const body = await response.text();
    assert.deepEqual(JSON.parse(body), { ok: true, realtime_configured: true });
    assert.doesNotMatch(body, /server-secret/);
  });
});

test("token mint keeps the API key upstream and returns only an ephemeral credential", async () => {
  let upstreamRequest;
  const fetchImpl = async (url, options) => {
    upstreamRequest = { url, options };
    return new Response(JSON.stringify({ value: "ephemeral-client-secret", expires_at: 1_900_000_000 }), {
      status: 200,
      headers: { "Content-Type": "application/json" },
    });
  };

  await withServer({ apiKey: "server-secret", fetchImpl }, async (baseURL) => {
    const response = await fetch(`${baseURL}/api/realtime-token`, { method: "POST" });
    const body = await response.json();
    assert.equal(response.status, 200);
    assert.equal(body.value, "ephemeral-client-secret");
    assert.equal(body.expires_at, 1_900_000_000);
    assert.equal(body.calls_url, "https://api.openai.com/v1/realtime/calls");
    assert.equal(body.model, "gpt-realtime-2.1-mini");
    assert.doesNotMatch(JSON.stringify(body), /server-secret/);
  });

  assert.equal(upstreamRequest.url, "https://api.openai.com/v1/realtime/client_secrets");
  assert.equal(upstreamRequest.options.headers.Authorization, "Bearer server-secret");
  const session = JSON.parse(upstreamRequest.options.body).session;
  assert.equal(session.type, "realtime");
  assert.equal(session.model, "gpt-realtime-2.1-mini");
  assert.deepEqual(session.output_modalities, ["audio"]);
  assert.equal(session.audio.input.turn_detection, null);
  assert.equal(session.audio.output.voice, "marin");
  assert.match(session.instructions, /never score/i);
});

test("token endpoint fails closed without a key or valid upstream secret", async () => {
  await withServer({ apiKey: "" }, async (baseURL) => {
    const response = await fetch(`${baseURL}/api/realtime-token`, { method: "POST" });
    assert.equal(response.status, 503);
    assert.match((await response.json()).error, /not configured/);
  });

  await withServer({
    apiKey: "server-secret",
    fetchImpl: async () => new Response(JSON.stringify({ expires_at: 1_900_000_000 }), { status: 200 }),
  }, async (baseURL) => {
    const response = await fetch(`${baseURL}/api/realtime-token`, { method: "POST" });
    assert.equal(response.status, 502);
    assert.match((await response.json()).error, /no short-lived/);
  });
});

test("token endpoint rejects non-POST requests and static serving is allowlisted", async () => {
  await withServer({ apiKey: "server-secret" }, async (baseURL) => {
    assert.equal((await fetch(`${baseURL}/api/realtime-token`)).status, 405);
    assert.equal((await fetch(`${baseURL}/grader.mjs`)).status, 200);
    assert.equal((await fetch(`${baseURL}/package.json`)).status, 404);
  });
});
