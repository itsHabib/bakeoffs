#!/usr/bin/env node
import { readFile, mkdir, writeFile, rm } from 'node:fs/promises';
import { resolve, dirname } from 'node:path';
import { execFileSync } from 'node:child_process';
import { createHash } from 'node:crypto';
import { validate, analyze } from './core.mjs';
import { startServer } from './server.mjs';
import { CDP, identity, runBrowser } from './browser.mjs';
import { html, markdown } from './report.mjs';

const [configArg, ...args] = process.argv.slice(2);
if (!configArg || args.some(a => !['--enforce', '--demo-check'].includes(a))) {
  console.error('Usage: node src/cli.mjs workload.json [--enforce | --demo-check]'); process.exit(2);
}
const output = resolve(process.env.PROOFGAP_OUT || 'artifacts');
let server, cdp;
try {
  await mkdir(output, { recursive: true });
  for (const file of ['report.json', 'report.html', 'report.md', 'failure.json']) await rm(resolve(output, file), { force: true });
  const path = resolve(configArg);
  const source = await readFile(path, 'utf8');
  const config = validate(JSON.parse(source));
  const endpoint = process.env.KITESURF_ENDPOINT || 'ws://127.0.0.1:18893/devtools/browser';
  const engine = await identity(endpoint);
  let engineRevision = process.env.KITESURF_REVISION || null;
  if (process.env.KITESURF_SOURCE) engineRevision = execFileSync('git', ['-C', process.env.KITESURF_SOURCE, 'rev-parse', 'HEAD'], { encoding: 'utf8' }).trim();
  server = await startServer(config, resolve(dirname(path), config.app), Number(process.env.PROOFGAP_PORT || 8893));
  cdp = await CDP.connect(endpoint);
  console.log(`${config.name} · ${engine.Browser}`);
  console.log('Running baseline…');
  const baseline = await runBrowser(cdp, config, server.open());
  const trials = [];
  for (const fault of config.faults) {
    if (baseline.error) {
      trials.push({ snapshot: {}, hits: [], elapsedMs: 0, error: 'Not attempted: baseline execution failed' });
      continue;
    }
    console.log(`Injecting ${fault.id}…`);
    trials.push(await runBrowser(cdp, config, server.open(fault)));
  }
  const results = analyze(config, baseline, trials);
  const incomplete = results.some(r => Object.values(r.suites).some(s => ['baseline-error', 'baseline-red', 'execution-error'].includes(s.status)));
  const summary = (incomplete ? 'INCOMPLETE — inspect execution errors. ' : '') + Object.keys(config.suites).map(name => {
    const relevant = results.filter(r => ['detected', 'survived'].includes(r.suites[name].status));
    return `${name}: ${relevant.filter(r => r.suites[name].status === 'survived').length} of ${relevant.length} observed faults missed`;
  }).join(' · ');
  const report = { schemaVersion: 1, name: config.name, createdAt: new Date().toISOString(), configSha256: createHash('sha256').update(source).digest('hex'), endpoint, engine, engineRevision, revisionSource: process.env.KITESURF_SOURCE ? 'local checkout; not cryptographic runtime attestation' : 'caller supplied; not cryptographic runtime attestation', actions: 'DOM click / input events via Runtime.evaluate', baseline, summary, results };
  await mkdir(output, { recursive: true });
  await writeFile(resolve(output, 'report.json'), JSON.stringify(report, null, 2) + '\n');
  await writeFile(resolve(output, 'report.html'), html(report));
  await writeFile(resolve(output, 'report.md'), markdown(report));
  console.log(summary);
  console.log(`Evidence: ${output}/report.json\nReadable report: ${output}/report.html`);
  const statuses = results.flatMap(r => Object.values(r.suites).map(s => s.status));
  if (statuses.some(s => ['baseline-error', 'baseline-red', 'execution-error'].includes(s))) process.exitCode = 2;
  if (args.includes('--enforce') && statuses.some(s => s !== 'detected')) process.exitCode = process.exitCode || 1;
  if (args.includes('--demo-check')) {
    if (!config.expected) throw new Error('No expected demo controls supplied');
    for (const result of results) for (const [name, suite] of Object.entries(result.suites)) {
      if (config.expected[result.fault.id]?.[name] !== suite.status) throw new Error(`Demo control mismatch: ${result.fault.id}/${name}: ${suite.status}`);
    }
    console.log('All declared demo controls verified.');
  }
} catch (error) {
  await mkdir(output, { recursive: true });
  await writeFile(resolve(output, 'failure.json'), JSON.stringify({ error: error.message, time: new Date().toISOString(), verdict: 'incomplete' }, null, 2));
  console.error(`INCOMPLETE: ${error.message}`); process.exitCode = 2;
} finally { cdp?.close(); if (server) await server.close(); }
