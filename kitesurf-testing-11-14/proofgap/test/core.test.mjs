import test from 'node:test';
import assert from 'node:assert/strict';
import { replacePointer, check, classify, validate, analyze } from '../src/core.mjs';
import { readFile } from 'node:fs/promises';
const config = JSON.parse(await readFile(new URL('../workloads/checkout/workload.json', import.meta.url)));

test('nested and escaped JSON pointers replace only their target', () => {
 const original = {'a/b': {'x~y': [1, 2]}, keep: true};
 assert.deepEqual(replacePointer(original, '/a~1b/x~0y/1', 3), {'a/b': {'x~y': [1,3]}, keep: true});
 assert.equal(original['a/b']['x~y'][1], 2);
});
test('invalid, unchanged and prototype pointers fail visibly', () => {
 for (const pointer of ['bad','/absent','/nested/absent','/__proto__/polluted','/constructor']) assert.throws(() => replacePointer({nested: {}}, pointer, 1));
 assert.throws(() => replacePointer({x: 1}, '/x', 1), /does not change/);
 assert.throws(() => replacePointer({x: null}, '/x/y', 1), /Missing/);
});
test('assertions distinguish missing, empty and exact text', () => {
 assert.equal(check({selector:'#x',op:'exists'}, {'#x':''}).pass, true);
 assert.equal(check({selector:'#x',op:'exists'}, {}).pass, false);
 assert.equal(check({selector:'#x',op:'equals',value:'USD 49.00'}, {'#x':'USD 490.00'}).pass, false);
 assert.equal(check({selector:'#x',op:'includes',value:'USD'}, {'#x':'USD 490.00'}).pass, true);
 assert.throws(() => check({selector:'#x',op:'unknown'}, {}));
});
const baseline = {snapshot: {'#x':'49'}, hits:[], error:null};
const trial = {snapshot: {'#x':'490'}, hits:[{mutated:true}], error:null};
const base = {baseline,trial,witness:'#x',baselineChecks:[{pass:true}],trialChecks:[{pass:true}]};
test('only an exercised, observed fault can survive or be detected', () => {
 assert.equal(classify(base),'survived');
 assert.equal(classify({...base,trialChecks:[{pass:false}]}),'detected');
 assert.equal(classify({...base,trial:{...trial,hits:[]},trialChecks:[{pass:false}]}),'not-exercised');
 assert.equal(classify({...base,trial:{...trial,snapshot:baseline.snapshot},trialChecks:[{pass:false}]}),'no-observed-change');
});
test('broken baseline and browser errors never earn detection credit', () => {
 assert.equal(classify({...base,baseline:{...baseline,error:'navigation'}}),'baseline-error');
 assert.equal(classify({...base,baselineChecks:[{pass:false}]}),'baseline-red');
 assert.equal(classify({...base,trial:{...trial,error:'CDP disconnected'},trialChecks:[{pass:false}]}),'execution-error');
});
test('config validation rejects empty suites and missing routes', () => {
 assert.equal(validate(config),config);
 assert.throws(() => validate({...config,suites:{weak:[]}}));
 assert.throws(() => validate({...config,routes:[]}));
 assert.throws(() => validate({...config,faults:[{...config.faults[0],route:'api/missing'}]}));
 assert.throws(() => validate({...config,steps:[{action:'execute',selector:'#x'}]}));
});
test('same trial reveals a hollow smoke test and strong behavior test', () => {
 const configOne = {...config,faults:[config.faults[0]]};
 const before = {error:null,hits:[],snapshot:{'#status':'Quote ready','#total':'USD 49.00'}};
 const after = {error:null,hits:[{mutated:true}],snapshot:{'#status':'Quote ready','#total':'USD 490.00'}};
 const [r] = analyze(configOne,before,[after]);
 assert.equal(r.suites.smoke.status,'survived');
 assert.equal(r.suites.behavior.status,'detected');
});
