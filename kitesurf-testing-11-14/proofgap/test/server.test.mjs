import test from 'node:test';
import assert from 'node:assert/strict';
import { readFile } from 'node:fs/promises';
import { fileURLToPath } from 'node:url';
import { startServer } from '../src/server.mjs';
import { html } from '../src/report.mjs';
const config = JSON.parse(await readFile(new URL('../workloads/checkout/workload.json',import.meta.url)));
test('HTTP mock isolates runs, records actual consumption and serves real app', async () => {
 const server = await startServer(config,fileURLToPath(new URL('../workloads/checkout/app',import.meta.url)),0);
 try {
  const first = server.open(); const second = server.open(config.faults[0]);
  assert.equal(first.hits.length,0); assert.equal(second.hits.length,0);
  const response = await fetch(second.url+'api/quote');
  assert.equal((await response.json()).totalCents,49000);
  assert.equal(second.hits[0].mutated,true); assert.equal(second.hits[0].bodySha256.length,64);
  assert.equal(first.hits.length,0);
  assert.equal((await (await fetch(first.url+'api/quote')).json()).totalCents,4900);
  assert.equal(first.hits[0].mutated,false);
  assert.match(await (await fetch(first.url)).text(),/Calculate total/);
  assert.equal((await fetch(first.url+'missing')).status,404);
  assert.equal((await fetch(first.url+'%2Fetc%2Fpasswd')).status,403);
 } finally { await server.close(); }
});
test('HTML report escapes user data', () => {
 const out = html({name:'<script>alert(1)</script>',summary:'safe',results:[],engine:{Browser:'Kitesurf/test'}});
 assert.ok(!out.includes('<script>')); assert.ok(out.includes('&lt;script&gt;'));
});
