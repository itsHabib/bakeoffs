import {test} from 'node:test';
import assert from 'node:assert/strict';
import vm from 'node:vm';
import {installMock, applyEvent, observe} from '../src/browser.mjs';
function harness() {
  const listeners = {};
  const input = {value:'new'};
  const result = {getAttribute: () => 'old'};
  const context = vm.createContext({window:{addEventListener:(type, fn)=>{listeners[type]=fn;}}, document:{querySelector:selector=>({'#input':input, '#result':result}[selector] ?? null)}, Event:class {constructor(type){this.type=type;}}});
  vm.runInContext(`(${installMock.toString()})()`,context);
  return {context,listeners,input,result,run:(fn,arg)=>vm.runInContext(`(${fn.toString()})(${JSON.stringify(arg) ?? ''})`,context)};
}
const oracle={actual:{selector:'#result',attribute:'data-value'},expected:{selector:'#input',property:'value'}};
test('unmatched response is unresolved, pending requests prevent a premature oracle', async () => {
  const h=harness();
  assert.equal((await h.run(applyEvent,{kind:'respond',url:'/missing',body:{}})).status,'unresolved');
  vm.runInContext("window.fetch('/pending')",h.context);
  assert.equal(h.run(observe,oracle).status,'unresolved');
});
test('fetch stub resolves repeated URLs in FIFO order without using a network', async () => {
  const h=harness();
  vm.runInContext("window.received=[];window.fetch('/same').then(r=>r.json()).then(x=>window.received.push(x));window.fetch('/same').then(r=>r.json()).then(x=>window.received.push(x));",h.context);
  await h.run(applyEvent,{kind:'respond',url:'/same',body:{n:1}});
  await h.run(applyEvent,{kind:'respond',url:'/same',body:{n:2}});
  await Promise.resolve();
  assert.deepEqual(JSON.parse(JSON.stringify(h.context.window.received)),[{n:1},{n:2}]);
  assert.equal(h.context.window.__needle.pending.length,0);
});
test('page errors remain distinguishable from DOM mismatches', () => {
  const h=harness();
  h.listeners.error({message:'application crashed'});
  assert.equal(h.run(observe,oracle).status,'page-error');
});
test('reads both sides of DOM invariant, missing targets are unresolved', () => {
  const h=harness();
  assert.deepEqual(JSON.parse(JSON.stringify(h.run(observe,oracle))),{actual:'old',expected:'new'});
  assert.equal(h.run(observe,{...oracle,actual:{selector:'#absent'}}).status,'unresolved');
});
