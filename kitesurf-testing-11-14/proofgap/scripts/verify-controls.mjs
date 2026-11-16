// Real-engine control campaigns. Never used to manufacture campaign verdicts.
import { readFile, writeFile, mkdir } from 'node:fs/promises';
import { resolve } from 'node:path';
import { spawnSync } from 'node:child_process';
import assert from 'node:assert/strict';
const original = JSON.parse(await readFile('workloads/checkout/workload.json','utf8'));
const strong = { ...original, app: resolve('workloads/checkout/app'), faults: original.faults.slice(0,2), suites: {behavior: original.suites.behavior} };
const controls = [
 {id:'strong-only', config:strong, code:0, status:'detected'},
 {id:'baseline-red', config:{...strong,faults:strong.faults.slice(0,1),suites:{behavior:[{selector:'#total',op:'equals',value:'USD 1.00'}]}},code:2,status:'baseline-red'},
 {id:'readiness-error',config:{...strong,faults:strong.faults.slice(0,1),ready:'#deliberately-missing',timeoutMs:300},code:2,status:'baseline-error'},
];
await mkdir('artifacts/controls',{recursive:true});
for (const control of controls) {
 const path=resolve(`artifacts/controls/${control.id}.json`);
 const output=resolve(`artifacts/controls/${control.id}`);
 await writeFile(path,JSON.stringify(control.config,null,2));
 const result=spawnSync(process.execPath,['src/cli.mjs',path,'--enforce'],{encoding:'utf8',env:{...process.env,PROOFGAP_OUT:output}});
 assert.equal(result.status,control.code,`${control.id}: ${result.stdout}\n${result.stderr}`);
 const report=JSON.parse(await readFile(resolve(output,'report.json'),'utf8'));
 assert.ok(report.results.every(r => r.suites.behavior.status===control.status),control.id);
 console.log(`Verified live control ${control.id}: exit ${result.status}, ${control.status}`);
}
