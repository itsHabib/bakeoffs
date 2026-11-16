import test from 'node:test';
import assert from 'node:assert/strict';
import http from 'node:http';
import { execFile } from 'node:child_process';
import { mkdtemp, writeFile, readFile, access, rm } from 'node:fs/promises';
import { resolve } from 'node:path';

test('failed engine validation removes stale success and emits incomplete evidence with exit 2', async () => {
 const output = await mkdtemp(resolve('artifacts/cli-test-'));
 const server = http.createServer((req,res) => res.end(JSON.stringify({Browser:'Chrome/test-control'})));
 await new Promise(r => server.listen(0,'127.0.0.1',r));
 try {
  await writeFile(resolve(output,'report.json'),'{"summary":"stale green"}');
  const result = await new Promise(resolveResult => execFile(process.execPath,['src/cli.mjs','workloads/checkout/workload.json'], {
   env:{...process.env,PROOFGAP_OUT:output,KITESURF_ENDPOINT:`ws://127.0.0.1:${server.address().port}/devtools/browser`},
  },(error,stdout,stderr) => resolveResult({error,stdout,stderr})));
  assert.equal(result.error.code,2);
  assert.match(result.stderr,/INCOMPLETE: Expected Kitesurf/);
  assert.equal(JSON.parse(await readFile(resolve(output,'failure.json'))).verdict,'incomplete');
  await assert.rejects(access(resolve(output,'report.json')));
 } finally {
  await new Promise(r => server.close(r));
  await rm(output,{recursive:true,force:true});
 }
});
