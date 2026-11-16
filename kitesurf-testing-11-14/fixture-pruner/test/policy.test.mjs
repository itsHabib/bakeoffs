import test from 'node:test';
import assert from 'node:assert/strict';
import { assess, checkConfig, paths, pointer, reduceFixture, remove } from '../src/policy.mjs';

test('removes unused response data while preserving the actual contract', async () => {
  const original = { trace: 'secret-ish', customer: { name: 'Ada', email: 'synthetic@example.test' }, amount: 42 };
  const result = await reduceFixture(original, async value => ({ pass: value.customer?.name === 'Ada' && value.amount === 42 }));
  assert.deepEqual(result.fixture, { customer: { name: 'Ada' }, amount: 42 });
  assert.equal(result.complete, true);
  assert.equal(original.customer.email, 'synthetic@example.test');
  for (const path of paths(result.fixture)) {
    const candidate = remove(result.fixture, path);
    assert.equal(candidate.customer?.name === 'Ada' && candidate.amount === 42, false, pointer(path));
  }
});

test('revisits keys after fallback behavior changes; not merely one pass', async () => {
  const result = await reduceFixture({ a: 1, b: 1 }, async value => ({ pass: !('b' in value) || 'a' in value }));
  assert.deepEqual(result.fixture, {});
  assert.equal(result.attempts[0].accepted, false);
});

test('budget never claims minimality', async () => {
  const result = await reduceFixture({ a: 1, b: 2 }, async () => ({ pass: true }), { maxRuns: 1 });
  assert.equal(result.complete, false);
  assert.equal(result.runs, 1);
});

test('preserves array length and order, handles null and escaped pointer keys', async () => {
  const result = await reduceFixture({ records: [{ 'a/b~c': null, keep: 3 }, { keep: 4 }] }, async value => ({ pass: value.records?.length === 2 && value.records[0].keep === 3 && value.records[1].keep === 4 }));
  assert.deepEqual(result.fixture, { records: [{ keep: 3 }, { keep: 4 }] });
  assert.equal(pointer(['records', 0, 'a/b~c']), '/records/0/a~1b~0c');
});

test('infrastructure failures abort instead of becoming retained-field evidence', async () => {
  await assert.rejects(reduceFixture({ a: 1 }, async () => { throw new Error('engine down'); }), /engine down/);
});

test('explicit DOM contract detects broken content, missing rows and enabled controls', () => {
  const checks = [{ selector: '#name', text: 'Ada' }, { selector: '.row', count: 2 }, { selector: '#pay', disabled: true }];
  assert.ok(assess(checks, [{ count: 1, text: 'Ada' }, { count: 2, text: 'a\nb' }, { count: 1, text: 'Pay', disabled: true }]).every(check => check.pass));
  assert.ok(assess(checks, [{ count: 1, text: 'Bob' }, { count: 1, text: 'a' }, { count: 1, text: 'Pay', disabled: false }]).every(check => !check.pass));
});

test('rejects vacuous contracts and invalid assertion types', () => {
  const base = { name: 'Test', appDir: '.', fixture: 'fixture.json', apiPath: '/api', ready: '#ready' };
  assert.throws(() => checkConfig({ ...base, checks: [] }), /explicit DOM check/);
  assert.throws(() => checkConfig({ ...base, checks: [{ selector: '#a' }] }), /assertion/);
  assert.throws(() => checkConfig({ ...base, checks: [{ selector: '#a', count: -1 }] }), /nonnegative/);
  assert.throws(() => checkConfig({ ...base, checks: [{ selector: '#a', disabled: 'false' }] }), /boolean/);
});
