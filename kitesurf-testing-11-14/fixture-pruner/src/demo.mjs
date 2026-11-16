import { run } from './cli.mjs';
import { InfrastructureError } from './browser.mjs';
try {
  for (const workload of ['invoices', 'warehouse']) {
    await run(`examples/${workload}/scenario.json`, `artifacts/${workload}`);
  }
} catch (error) {
  console.error(`${error instanceof InfrastructureError ? 'INFRASTRUCTURE' : 'CONTRACT'}: ${error.message}`);
  if (error.detail) console.error(JSON.stringify(error.detail, null, 2));
  process.exitCode = error instanceof InfrastructureError ? 3 : 2;
}
