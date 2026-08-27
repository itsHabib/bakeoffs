import assert from 'node:assert/strict';
import { readFileSync } from 'node:fs';
import { join } from 'node:path';
import test from 'node:test';
import { projectRoot } from '../scripts/quint-runner.mjs';

function fixture(slug) {
  return JSON.parse(readFileSync(
    join(projectRoot, 'fixtures', 'generated', `${slug}.json`),
    'utf8',
  ));
}

test('duplicate remote save fixture preserves the lost-receipt counterexample', () => {
  const result = fixture('duplicate-save-after-lost-receipt');
  const final = result.trace.at(-1);

  assert.equal(result.source.violatedInvariant, 'uploadIsDeduplicated');
  assert.equal(final.action, 'retry-without-reconciliation');
  assert.equal(final.state.flowStatus, 'PayloadSaved');
  assert.equal(final.state.receiptRecorded, false);
  assert.equal(final.state.flowSaveCount, 2);
});

test('retained command fixture preserves the unrecoverable workflow counterexample', () => {
  const result = fixture('retained-command-without-recovery');
  const final = result.trace.at(-1);

  assert.equal(result.source.violatedInvariant, 'retainedCommandIsRecoverable');
  assert.equal(final.action, 'retain-without-recovery');
  assert.equal(final.state.commandStatus, 'CommandRetained');
  assert.equal(final.state.recoveryAvailable, false);
});

test('early completion fixture preserves local/remote divergence', () => {
  const result = fixture('local-completion-before-receipt');
  const final = result.trace.at(-1);

  assert.equal(result.source.violatedInvariant, 'localCompletionHasReceipt');
  assert.equal(final.action, 'complete-before-receipt');
  assert.equal(final.state.taskStatus, 'TaskComplete');
  assert.equal(final.state.receiptRecorded, false);
});

test('invalidation fixture preserves the orphaned remote effect', () => {
  const result = fixture('invalidation-orphans-remote-payload');
  const final = result.trace.at(-1);

  assert.equal(result.source.violatedInvariant, 'remoteWorkIsNotOrphaned');
  assert.equal(final.action, 'invalidate-after-remote-save');
  assert.equal(final.state.epochStatus, 'EpochInvalidated');
  assert.equal(final.state.flowStatus, 'PayloadSaved');
});
