import { test } from 'node:test';
import assert from 'node:assert/strict';
import { createChannel } from '../src/channel.js';

const mk = (key, words, priority = false) => ({ key, words, priority });

const drain = (ch) => {
  const out = [];
  for (let it = ch.next(); it; it = ch.next()) out.push(it);
  return out;
};

test('routine traffic transmits FIFO', () => {
  const ch = createChannel();
  ch.enqueue(mk('a', 5));
  ch.enqueue(mk('b', 5));
  ch.enqueue(mk('c', 5));
  assert.deepEqual(drain(ch).map((it) => it.key), ['a', 'b', 'c']);
});

test('priority jumps ahead of routine but stays FIFO among priorities', () => {
  const ch = createChannel();
  ch.enqueue(mk('r1', 5));
  ch.enqueue(mk('r2', 5));
  ch.enqueue(mk('p1', 10, true));
  ch.enqueue(mk('p2', 10, true));
  assert.deepEqual(drain(ch).map((it) => it.key), ['p1', 'p2', 'r1', 'r2']);
});

test('a newer routine call with the same key supersedes the queued one', () => {
  const ch = createChannel();
  ch.enqueue({ ...mk('other', 5), label: 'other' });
  ch.enqueue({ ...mk('agent:stream', 5), label: 'stale' });
  ch.enqueue({ ...mk('agent:stream', 6), label: 'fresh' });
  const out = drain(ch);
  assert.deepEqual(out.map((it) => it.label), ['other', 'fresh']);
  assert.equal(ch.stats.coalesced, 1);
});

test('different keys never coalesce', () => {
  const ch = createChannel();
  ch.enqueue(mk('a:one', 5));
  ch.enqueue(mk('a:two', 5));
  assert.equal(ch.pending, 2);
  assert.equal(ch.stats.coalesced, 0);
});

test('saturation drops the oldest routine call first', () => {
  // Each 10-word item costs 10*0.5+1 = 6s; budget 13s holds two queued items.
  const ch = createChannel({ maxBacklogSec: 13, secPerWord: 0.5, overheadSec: 1 });
  ch.enqueue({ ...mk('a', 10), label: 'oldest' });
  ch.enqueue({ ...mk('b', 10), label: 'middle' });
  ch.enqueue({ ...mk('c', 10), label: 'newest' });
  const out = drain(ch);
  assert.deepEqual(out.map((it) => it.label), ['middle', 'newest']);
  assert.equal(ch.stats.dropped, 1);
});

test('the newest routine call survives even when it alone exceeds the budget', () => {
  const ch = createChannel({ maxBacklogSec: 5, secPerWord: 1, overheadSec: 1 });
  ch.enqueue(mk('prose-monologue', 30));
  assert.equal(ch.pending, 1);
  assert.equal(ch.stats.dropped, 0);
});

test('escalations are never dropped, even under total saturation', () => {
  const ch = createChannel({ maxBacklogSec: 5, secPerWord: 0.5, overheadSec: 1 });
  for (let i = 0; i < 10; i++) ch.enqueue(mk(`p${i}`, 12, true));
  for (let i = 0; i < 10; i++) ch.enqueue(mk(`r${i}`, 8));
  const out = drain(ch);
  const priorities = out.filter((it) => it.priority);
  assert.equal(priorities.length, 10, 'every escalation must transmit');
  assert.deepEqual(priorities.map((it) => it.key), Array.from({ length: 10 }, (_, i) => `p${i}`));
  assert.ok(ch.stats.dropped > 0, 'routine traffic should have been shed');
});

test('escalations are never coalesced, even with the same key', () => {
  const ch = createChannel();
  ch.enqueue(mk('agent:stream', 10, true));
  ch.enqueue(mk('agent:stream', 10, true));
  assert.equal(ch.pending, 2);
  assert.equal(ch.stats.coalesced, 0);
});

test('transmitted count tracks what actually aired', () => {
  const ch = createChannel();
  ch.enqueue(mk('a', 5));
  ch.enqueue(mk('b', 5));
  drain(ch);
  assert.equal(ch.stats.transmitted, 2);
});
