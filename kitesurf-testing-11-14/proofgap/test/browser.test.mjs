import test from 'node:test';
import assert from 'node:assert/strict';
import http from 'node:http';
import { identity, runBrowser } from '../src/browser.mjs';

test('discovery rejects a Chromium endpoint instead of falling back', async () => {
 const server = http.createServer((req,res) => res.end(JSON.stringify({Browser:'Chrome/123'})));
 await new Promise(r => server.listen(0,'127.0.0.1',r));
 try { await assert.rejects(identity(`ws://127.0.0.1:${server.address().port}/devtools/browser`),/Expected Kitesurf/); }
 finally { await new Promise(r => server.close(r)); }
});
test('protocol failure becomes execution evidence, never a false assertion failure', async () => {
 const cdp = {send: async () => {throw new Error('controlled protocol failure');}};
 const result = await runBrowser(cdp,{},{hits:[],url:'http://127.0.0.1/'});
 assert.equal(result.error,'controlled protocol failure');
 assert.deepEqual(result.snapshot,{}); assert.deepEqual(result.hits,[]);
});
test('browser context cleanup failure also makes execution inconclusive', async () => {
 const calls=[];
 const cdp={send:async method => {
   calls.push(method);
   if(method==='Target.createBrowserContext') return {browserContextId:'unit-context'};
   if(method==='Target.createTarget') throw new Error('create failed');
   if(method==='Target.disposeBrowserContext') throw new Error('cleanup failed');
 }};
 const result=await runBrowser(cdp,{},{hits:[],url:'http://127.0.0.1/'});
 assert.match(result.error,/create failed; Context cleanup: cleanup failed/);
 assert.deepEqual(calls,['Target.createBrowserContext','Target.createTarget','Target.disposeBrowserContext']);
});
