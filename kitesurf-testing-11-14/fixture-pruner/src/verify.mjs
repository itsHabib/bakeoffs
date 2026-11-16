import { readFile } from 'node:fs/promises';
import { resolve, dirname } from 'node:path';
import { checkConfig } from './policy.mjs';
import { startServer } from './server.mjs';
import { connect, observe, InfrastructureError } from './browser.mjs';

const [scenario, fixtureFile] = process.argv.slice(2);
let browser, server;
try {
  if (!scenario || !fixtureFile) throw new Error('Usage: node src/verify.mjs <scenario.json> <fixture.json>');
  const config = checkConfig(JSON.parse(await readFile(scenario, 'utf8')));
  const fixture = JSON.parse(await readFile(fixtureFile, 'utf8'));
  ({ browser } = await connect(process.env.KITESURF_ENDPOINT ?? 'ws://127.0.0.1:18891/devtools/browser'));
  try {
    server = await startServer({ appDir: resolve(dirname(scenario), config.appDir), apiPath: config.apiPath, port: Number(process.env.PORT ?? 8891) });
  } catch (error) { throw new InfrastructureError(`Local fixture host failed: ${error.message}`); }
  server.setFixture(fixture);
  const result = await observe(browser, server.url + (config.entry ?? '/'), config);
  if (!server.hits()) throw new InfrastructureError('Consumer did not request the configured API');
  console.log(JSON.stringify(result, null, 2));
  if (!result.pass) process.exitCode = 2;
} catch (error) {
  console.error(`${error instanceof InfrastructureError ? 'INFRASTRUCTURE' : 'CONTRACT'}: ${error.message}`);
  process.exitCode = error instanceof InfrastructureError ? 3 : 2;
} finally {
  if (browser) await browser.disconnect();
  if (server) await server.close();
}
