import { isDeepStrictEqual } from 'node:util';

export function replacePointer(source, pointer, value) {
  if (!/^\//.test(pointer)) throw new Error('Mutation pointer must start with /');
  const parts = pointer.slice(1).split('/').map(x => x.replace(/~1/g, '/').replace(/~0/g, '~'));
  if (parts.some(x => ['__proto__', 'constructor', 'prototype'].includes(x))) throw new Error('Unsafe pointer');
  const copy = structuredClone(source);
  let current = copy;
  for (const key of parts.slice(0, -1)) {
    if (current === null || typeof current !== 'object' || !Object.hasOwn(current, key)) throw new Error(`Missing pointer: ${pointer}`);
    current = current[key];
  }
  const key = parts.at(-1);
  if (current === null || typeof current !== 'object' || !Object.hasOwn(current, key)) throw new Error(`Missing pointer: ${pointer}`);
  if (isDeepStrictEqual(current[key], value)) throw new Error(`Mutation does not change ${pointer}`);
  current[key] = structuredClone(value);
  return copy;
}

export function check(assertion, snapshot) {
  const actual = snapshot[assertion.selector] ?? null;
  if (assertion.op === 'exists') return { ...assertion, actual, pass: actual !== null };
  if (assertion.op === 'equals') return { ...assertion, actual, pass: actual === assertion.value };
  if (assertion.op === 'includes') return { ...assertion, actual, pass: actual !== null && actual.includes(assertion.value) };
  throw new Error(`Unsupported assertion: ${assertion.op}`);
}

export function classify({ baseline, trial, witness, baselineChecks, trialChecks }) {
  if (baseline.error) return 'baseline-error';
  if (baselineChecks.some(x => !x.pass)) return 'baseline-red';
  if (trial.error) return 'execution-error';
  if (!trial.hits.some(x => x.mutated)) return 'not-exercised';
  if (trial.snapshot[witness] === baseline.snapshot[witness]) return 'no-observed-change';
  return trialChecks.every(x => x.pass) ? 'survived' : 'detected';
}

export function validate(config) {
  const need = (ok, message) => { if (!ok) throw new Error(message); };
  need(typeof config.name === 'string' && typeof config.app === 'string', 'name and app required');
  need(typeof config.ready === 'string', 'ready CSS selector required');
  need(Array.isArray(config.routes) && config.routes.length > 0, 'routes required');
  need(new Set(config.routes.map(r => r.path)).size === config.routes.length, 'Duplicate route');
  for (const route of config.routes) need(/^api\/[\w/.-]+$/.test(route.path) && route.body !== undefined, 'Route must be relative api/... with body');
  need(Array.isArray(config.faults) && config.faults.length > 0, 'faults required');
  need(new Set(config.faults.map(f => f.id)).size === config.faults.length, 'Duplicate fault id');
  for (const fault of config.faults) {
    const route = config.routes.find(r => r.path === fault.route);
    need(route && typeof fault.id === 'string' && typeof fault.witness === 'string', 'Fault requires id, known route and witness selector');
    replacePointer(route.body, fault.pointer, fault.value);
  }
  need(config.suites && Object.keys(config.suites).length > 0, 'suites required');
  for (const [name, assertions] of Object.entries(config.suites)) {
    need(Array.isArray(assertions) && assertions.length > 0, `${name}: nonempty assertions required`);
    for (const a of assertions) need(typeof a.selector === 'string' && ['exists', 'equals', 'includes'].includes(a.op) && (a.op === 'exists' || typeof a.value === 'string'), 'Invalid assertion');
  }
  for (const step of config.steps ?? []) need(['click', 'fill', 'wait'].includes(step.action) && typeof step.selector === 'string' && (step.action !== 'fill' || typeof step.value === 'string'), 'Invalid journey step');
  return config;
}

export function analyze(config, baseline, trials) {
  return config.faults.map((fault, i) => {
    const trial = trials[i];
    const suites = Object.fromEntries(Object.entries(config.suites).map(([name, assertions]) => {
      const baselineChecks = assertions.map(a => check(a, baseline.snapshot));
      const trialChecks = assertions.map(a => check(a, trial.snapshot));
      return [name, { status: classify({ baseline, trial, witness: fault.witness, baselineChecks, trialChecks }), checks: trialChecks, baselineChecks }];
    }));
    return { fault, before: baseline.snapshot[fault.witness], after: trial.snapshot[fault.witness], suites, ...trial };
  });
}
