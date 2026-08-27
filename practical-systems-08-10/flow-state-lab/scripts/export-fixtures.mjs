import { existsSync, mkdirSync, readFileSync, rmSync, writeFileSync } from 'node:fs';
import { join } from 'node:path';
import { projectRoot, runQuint, withApalache } from './quint-runner.mjs';

const cases = [
  {
    slug: 'duplicate-save-after-lost-receipt',
    module: 'DuplicateSaveBug',
    invariant: 'uploadIsDeduplicated',
    summary: 'A blind retry repeats a remote payload save after the first response was lost.',
  },
  {
    slug: 'retained-command-without-recovery',
    module: 'RetainedWorkflowBug',
    invariant: 'retainedCommandIsRecoverable',
    summary: 'A retained durable command blocks retries without exposing resume or release.',
  },
  {
    slug: 'local-completion-before-receipt',
    module: 'EarlyLocalCompletionBug',
    invariant: 'localCompletionHasReceipt',
    summary: 'Local task completion outruns durable evidence that the remote save succeeded.',
  },
  {
    slug: 'invalidation-orphans-remote-payload',
    module: 'InvalidateAfterRemoteSaveBug',
    invariant: 'remoteWorkIsNotOrphaned',
    summary: 'Local invalidation abandons a remote payload without reconciliation or compensation.',
  },
];

const rawDir = join(projectRoot, 'tmp', 'itf');
const fixtureDir = join(projectRoot, 'fixtures', 'generated');
mkdirSync(rawDir, { recursive: true });
mkdirSync(fixtureDir, { recursive: true });

function normalizeValue(value) {
  if (value && typeof value === 'object') {
    if (typeof value.tag === 'string') return value.tag;
    if (typeof value['#bigint'] === 'string') return Number(value['#bigint']);
  }
  return value;
}

function normalizeState(state) {
  return {
    taskStatus: normalizeValue(state.taskStatus),
    epochStatus: normalizeValue(state.epochStatus),
    flowStatus: normalizeValue(state.flowStatus),
    commandStatus: normalizeValue(state.commandStatus),
    snapshotReady: state.snapshotReady,
    receiptRecorded: state.receiptRecorded,
    recoveryAvailable: state.recoveryAvailable,
    flowSaveCount: normalizeValue(state.flowSaveCount),
  };
}

function inferAction(previous, current) {
  if (!previous) return 'init';
  if (previous.taskStatus === 'NotStarted' && current.taskStatus === 'Started') return 'start-task';
  if (!previous.snapshotReady && current.snapshotReady) return 'stage-snapshot';
  if (previous.flowStatus === 'NoPayload' && current.flowStatus === 'WriteOpen') return 'open-flow-write';
  if (previous.flowStatus === 'WriteOpen' && current.flowStatus === 'PayloadSaved') return 'save-flow-payload';
  if (current.flowSaveCount > previous.flowSaveCount && previous.flowStatus === 'PayloadSaved') {
    return 'retry-without-reconciliation';
  }
  if (!previous.receiptRecorded && current.receiptRecorded) return 'record-upload-receipt';
  if (previous.taskStatus !== 'TaskComplete' && current.taskStatus === 'TaskComplete') {
    return current.receiptRecorded ? 'complete-local-task' : 'complete-before-receipt';
  }
  if (previous.commandStatus === 'NoCommand' && current.commandStatus === 'CommandRunning') {
    return 'begin-complete-command';
  }
  if (previous.commandStatus === 'CommandRunning' && current.commandStatus === 'CommandRetained') {
    return current.recoveryAvailable ? 'retain-with-recovery' : 'retain-without-recovery';
  }
  if (previous.epochStatus === 'EpochOpen' && current.epochStatus === 'EpochInvalidated') {
    return previous.flowStatus === 'NoPayload' ? 'invalidate-local-epoch' : 'invalidate-after-remote-save';
  }
  return 'state-transition';
}

await withApalache(({ serverEndpoint }) => {
  for (const scenario of cases) {
    const rawPath = join(rawDir, `${scenario.slug}.itf.json`);
    rmSync(rawPath, { force: true });

    const result = runQuint([
      'verify',
      'model/KnownFailures.qnt',
      `--main=${scenario.module}`,
      '--step=stepBug',
      `--invariant=${scenario.invariant}`,
      '--max-steps=12',
      `--server-endpoint=${serverEndpoint}`,
      `--out-itf=${rawPath}`,
      '--verbosity=0',
    ], 1);

    if (result.stderr.trim() !== 'error: found a counterexample') {
      throw new Error(
        `Quint failed without proving the expected ${scenario.invariant} counterexample:\n${result.stderr}`,
      );
    }
    if (!existsSync(rawPath)) {
      throw new Error(`Quint reported a counterexample but did not write ${rawPath}.`);
    }

    const itf = JSON.parse(readFileSync(rawPath, 'utf8'));
    const states = itf.states.map(normalizeState);
    const trace = states.map((state, index) => ({
      step: index,
      action: inferAction(states[index - 1], state),
      state,
    }));
    const fixture = {
      schemaVersion: 1,
      scenario: scenario.slug,
      summary: scenario.summary,
      source: {
        specification: 'model/KnownFailures.qnt',
        module: scenario.module,
        violatedInvariant: scenario.invariant,
      },
      trace,
    };

    writeFileSync(
      join(fixtureDir, `${scenario.slug}.json`),
      `${JSON.stringify(fixture, null, 2)}\n`,
    );
    process.stdout.write(`[fixture] ${scenario.slug}: ${trace.length} states\n`);
  }
});
