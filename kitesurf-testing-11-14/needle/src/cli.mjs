import {readFile, mkdir, writeFile} from 'node:fs/promises';
import {resolve} from 'node:path';
import {CDP, replay} from './browser.mjs';
import {validate, classify, repeated, reduce} from './reduce.mjs';

const [command, input, output = 'artifacts/run'] = process.argv.slice(2);
if (!['reduce', 'replay'].includes(command) || !input) {
  console.error('Usage: node src/cli.mjs <reduce|replay> scenario.json [output-directory]');
  process.exit(2);
}
let cdp;
try {
  const scenario = validate(JSON.parse(await readFile(input, 'utf8')));
  const endpoint = process.env.KITESURF_ENDPOINT ?? 'ws://127.0.0.1:18892/devtools/browser';
  cdp = await CDP.connect(endpoint);
  const engine = await cdp.send('Browser.getVersion');
  if (!/kitesurf/i.test([engine.product, engine.revision, engine.userAgent].join(' '))) throw new Error(`Refusing non-Kitesurf engine: ${JSON.stringify(engine)}`);
  const provenance = process.env.KITESURF_REVISION ?? 'not-provided';
  const started = new Date().toISOString();
  const baseline = await repeated(async () => classify(await replay(cdp, scenario, scenario.events)), 3);
  if (command === 'replay') {
    console.log(JSON.stringify({engine, revision: provenance, baseline}, null, 2));
    process.exitCode = baseline.status === (process.env.EXPECT_STATUS ?? 'fail') ? 0 : 1;
  } else {
    if (baseline.status !== 'fail') throw new Error(`Input does not stably fail: ${JSON.stringify(baseline)}`);
    const attempts = [];
    const cache = new Map();
    async function probe(events, options = {}) {
      const key = JSON.stringify(events.map(e => e.id));
      if (!options.fresh && cache.has(key)) return cache.get(key);
      const result = await repeated(async () => classify(await replay(cdp, scenario, events), baseline));
      attempts.push({events: events.map(e => e.id), ...result});
      cache.set(key, result);
      process.stdout.write(`  ${String(events.length).padStart(2)} events: ${result.status}\n`);
      return result;
    }
    const reduced = await reduce(scenario.events, probe);
    const final = await probe(reduced.events, {fresh: true});
    if (final.status !== 'fail') throw new Error(`Final replay did not reproduce: ${JSON.stringify(final)}`);
    const report = {name: scenario.name, started, completed: new Date().toISOString(), endpoint, engine, revision: provenance, baseline, final, originalCount: scenario.events.length, reducedCount: reduced.events.length, attempts, witnesses: reduced.witnesses, guarantee: 'Each single event deletion, including dependent closure, was tested twice and did not consistently reproduce the exact mismatch in both runs. Unresolved witnesses do not establish semantic necessity; unstable witnesses are inconclusive.'};
    await mkdir(output, {recursive: true});
    await writeFile(resolve(output, 'minimal.json'), JSON.stringify({...scenario, events: reduced.events}, null, 2) + '\n');
    await writeFile(resolve(output, 'evidence.json'), JSON.stringify(report, null, 2) + '\n');
    await writeFile(resolve(output, 'report.md'), `# ${scenario.name}\n\nReduced ${scenario.events.length} events to ${reduced.events.length}.\n\nExpected \`${baseline.expected}\`; observed \`${baseline.actual}\`.\n\nEngine: Kitesurf (CDP product ${engine.product}; marker ${engine.revision}), source revision ${provenance}.\n\n## Replay\n\n\`node src/cli.mjs replay ${output}/minimal.json\`\n\n## Remaining events\n\n${reduced.events.map(e => `- ${e.id}: ${e.kind}`).join('\n')}\n\n## Deletion witnesses\n\n${reduced.witnesses.map(w => `- Remove ${w.removed}: ${w.status}${w.reason ? ` (${w.reason})` : ''}`).join('\n')}\n\n${report.guarantee}\n`);
    console.log(`\n${scenario.events.length} → ${reduced.events.length} events. Evidence: ${output}/report.md`);
  }
} catch (error) {
  console.error(`Needle stopped: ${error.message}`);
  process.exitCode = 1;
} finally {cdp?.close();}
