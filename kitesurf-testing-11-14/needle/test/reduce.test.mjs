import {test} from 'node:test';
import assert from 'node:assert/strict';
import {validate, without, reduce, classify, repeated} from '../src/reduce.mjs';
const e = (id, requires = []) => ({id, kind: 'click', selector: '#x', requires});
const sample = {url: 'http://127.0.0.1:8892', events: [e('a')], oracle: {actual: {selector: '#a'}, expected: {selector: '#e'}}};
test('reject duplicate IDs, out-of-order dependencies, remote URLs, malformed response and oracle', () => {
  assert.equal(validate(sample), sample);
  for (const change of [{events: [e('a'), e('a')]}, {events: [e('a', ['b']), e('b')]}, {url: 'https://example.com'}, {events: [{id:'a',kind:'respond',url:'/api'}]}, {oracle: {}}]) assert.throws(() => validate({...sample, ...change}));
});
test('deletion closes transitively over dependencies and preserves order', () => {
  assert.deepEqual(without([e('a'), e('noise'), e('b', ['a']), e('c', ['b'])], new Set(['a'])).map(x => x.id), ['noise']);
});
test('preserves a multi-event cause and proves deletion witnesses', async () => {
  const events = [e('n1'), e('start'), e('n2'), e('new'), e('fresh', ['new']), e('old', ['start']), e('n3')];
  const probe = async events => ({status: ['start', 'new', 'fresh', 'old'].every(id => events.some(e => e.id === id)) ? 'fail' : 'pass'});
  const result = await reduce(events, probe);
  assert.deepEqual(result.events.map(e => e.id), ['start', 'new', 'fresh', 'old']);
  assert.equal(result.witnesses.length, 4);
  assert.ok(result.witnesses.every(w => w.status === 'pass'));
});
test('does not count an unrelated mismatch as the original failure', () => {
  assert.equal(classify({actual:'x',expected:'y'}, {actual:'x',expected:'z'}).status, 'different-failure');
  assert.equal(classify({actual:'x',expected:'x'}).status, 'pass');
  assert.equal(classify({actual:'x',expected:'y'}, {actual:'x',expected:'y'}).status, 'fail');
});
test('invalid candidates do not shrink the trace', async () => {
  const result = await reduce([e('a'), e('b')], async () => ({status: 'unresolved'}));
  assert.equal(result.events.length, 2);
  assert.ok(result.witnesses.every(w => w.status === 'unresolved'));
});
test('infrastructure errors propagate rather than becoming interesting failures', async () => {
  await assert.rejects(reduce([e('a')], async () => {throw new Error('engine died');}), /engine died/);
});
test('flaky runs are not accepted and a fresh audit can reject unstable reduction', async () => {
  let n = 0;
  assert.equal((await repeated(async () => ({status: ++n % 2 ? 'fail' : 'pass'}))).status, 'unstable');
  await assert.rejects(reduce([e('a')], async (_, options) => ({status: options?.fresh ? 'fail' : 'pass'})), /Unstable/);
});
test('handles nonmonotonic predicates without claiming a global minimum', async () => {
  const result = await reduce([e('a'), e('b'), e('c')], async events => ({status: events.length === 3 || (events.length === 1 && events[0].id === 'b') ? 'fail' : 'pass'}));
  assert.ok(result.events.length >= 1);
  assert.ok(result.witnesses.every(w => w.status === 'pass'));
});
