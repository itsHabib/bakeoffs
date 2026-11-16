import { readFile, mkdir, writeFile } from 'node:fs/promises';
import { resolve, dirname, join } from 'node:path';
import { createHash } from 'node:crypto';
import { fileURLToPath } from 'node:url';
import { checkConfig, reduceFixture, paths, remove, pointer } from './policy.mjs';
import { startServer } from './server.mjs';
import { connect, observe, InfrastructureError } from './browser.mjs';
import { reportHtml } from './report.mjs';

export async function run(configPath, outputDir, options = {}) {
  const configFile = resolve(configPath);
  const config = checkConfig(JSON.parse(await readFile(configFile, 'utf8')));
  const base = dirname(configFile);
  const fixture = JSON.parse(await readFile(resolve(base, config.fixture), 'utf8'));
  const endpoint = options.endpoint ?? process.env.KITESURF_ENDPOINT ?? 'ws://127.0.0.1:18891/devtools/browser';
  const { browser, identity } = await connect(endpoint);
  let server;
  try {
    try {
      server = await startServer({ appDir: resolve(base, config.appDir), apiPath: config.apiPath, port: options.port ?? Number(process.env.PORT ?? 8891) });
    } catch (error) { throw new InfrastructureError(`Local fixture host failed: ${error.message}`); }
    let counter = 0;
    const evaluate = async (value, label) => {
      server.setFixture(value);
      const observation = await observe(browser, `${server.url}${config.entry ?? '/'}?pruneRun=${++counter}`, config);
      if (!server.hits()) throw new InfrastructureError(`Consumer did not request ${config.apiPath}; check fixture wiring`);
      process.stdout.write(`${observation.pass ? 'PASS' : 'FAIL'} ${label}\n`);
      return observation;
    };
    const baseline = await evaluate(fixture, 'original response');
    if (!baseline.pass) {
      const error = new Error('Baseline contract failed; no reduced fixture was produced');
      error.detail = baseline;
      throw error;
    }
    const repeat = await evaluate(fixture, 'baseline repeat');
    if (JSON.stringify(baseline) !== JSON.stringify(repeat)) throw new Error('Baseline was unstable across two fresh pages');
    const result = await reduceFixture(fixture, evaluate, { maxRuns: config.maxRuns ?? 100 });
    const finalCheck = await evaluate(result.fixture, 'final fixture');
    if (!finalCheck.pass) throw new Error('Final verification failed; fixture rejected');
    const out = resolve(outputDir);
    await mkdir(join(out, 'controls'), { recursive: true });
    const controls = [];
    if (result.complete) {
      // Reuse exact final-pass observations, then export independently replayable
      // controls. Every final key was tested in the final complete pass.
      const finalPaths = paths(result.fixture);
      const finalAttempts = result.attempts.slice(-finalPaths.length);
      for (const [i, path] of finalPaths.entries()) {
        const observation = finalAttempts[i];
        if (!observation || observation.path !== pointer(path) || observation.pass) throw new Error('Final control evidence does not match fixture');
        const file = `controls/${String(i + 1).padStart(2, '0')}.json`;
        await writeFile(join(out, file), JSON.stringify(remove(result.fixture, path), null, 2) + '\n');
        controls.push({ ...observation, file });
      }
    }
    const report = {
      name: config.name, timestamp: new Date().toISOString(), engine: identity,
      engineRevision: process.env.KITESURF_REVISION ?? 'unreported (set KITESURF_REVISION)',
      sourceSha256: createHash('sha256').update(JSON.stringify(fixture)).digest('hex'),
      contract: config.checks, checkCount: config.checks.length,
      beforeBytes: Buffer.byteLength(JSON.stringify(fixture)), afterBytes: Buffer.byteLength(JSON.stringify(result.fixture)),
      ...result, baseline, finalCheck, controls,
    };
    await writeFile(join(out, 'fixture.min.json'), JSON.stringify(result.fixture, null, 2) + '\n');
    await writeFile(join(out, 'report.json'), JSON.stringify(report, null, 2) + '\n');
    await writeFile(join(out, 'index.html'), reportHtml(report));
    console.log(`\n${config.name}: ${report.beforeBytes} → ${report.afterBytes} bytes; ${controls.length} failing controls.\nReport: ${join(out, 'index.html')}`);
    return report;
  } finally {
    await browser.disconnect();
    if (server) await server.close();
  }
}

if (process.argv[1] && resolve(process.argv[1]) === fileURLToPath(import.meta.url)) {
  const [config, out = 'artifacts/run'] = process.argv.slice(2);
  if (!config) { console.error('Usage: node src/cli.mjs <scenario.json> [output-dir]'); process.exitCode = 2; }
  else await run(config, out).catch(error => {
    console.error(`${error instanceof InfrastructureError ? 'INFRASTRUCTURE' : 'CONTRACT'}: ${error.message}`);
    if (error.detail) console.error(JSON.stringify(error.detail, null, 2));
    process.exitCode = error instanceof InfrastructureError ? 3 : 2;
  });
}
