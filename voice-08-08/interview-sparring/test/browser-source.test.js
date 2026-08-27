import assert from "node:assert/strict";
import {readFile} from "node:fs/promises";
import test from "node:test";

test("live machine exists before Realtime can deliver session.created", async () => {
  const source = await readFile(new URL("../public/app.js", import.meta.url), "utf8");
  const construct = source.indexOf("machine = new InterviewMachine");
  const connect = source.indexOf("await connectRealtime(token)");
  const listen = source.indexOf('channel.addEventListener("message", handleRealtimeEvent)');
  const offer = source.indexOf("const offer = await peer.createOffer()");

  assert.ok(construct >= 0 && connect >= 0 && construct < connect);
  assert.ok(listen >= 0 && offer >= 0 && listen < offer);
});
