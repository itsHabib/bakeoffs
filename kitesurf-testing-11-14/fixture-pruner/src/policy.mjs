// Policy is independent of HTTP, CDP, and filesystem access.
export function paths(value, prefix = []) {
  if (value === null || typeof value !== 'object') return [];
  // Array length/order are part of the scenario; only object keys are removed.
  if (Array.isArray(value)) return value.flatMap((item, i) => paths(item, [...prefix, i]));
  return Object.keys(value).flatMap(key => [[...prefix, key], ...paths(value[key], [...prefix, key])]);
}

export function pointer(path) {
  return '/' + path.map(key => String(key).replaceAll('~', '~0').replaceAll('/', '~1')).join('/');
}

export function remove(value, path) {
  const result = structuredClone(value);
  let parent = result;
  for (const key of path.slice(0, -1)) parent = parent[key];
  delete parent[path.at(-1)];
  return result;
}

export function assess(checks, observed) {
  return checks.map((check, i) => {
    const actual = observed[i];
    const failures = [];
    if (!actual || actual.error) failures.push(actual?.error || 'Observation missing');
    else {
      if (actual.count !== (check.count ?? 1)) failures.push(`Expected ${check.count ?? 1} element(s), got ${actual.count}`);
      if ('text' in check && actual.text !== check.text) failures.push(`Expected text ${JSON.stringify(check.text)}, got ${JSON.stringify(actual.text)}`);
      if ('includes' in check && !actual.text.includes(check.includes)) failures.push(`Text does not include ${JSON.stringify(check.includes)}`);
      if ('disabled' in check && actual.disabled !== check.disabled) failures.push(`Expected disabled=${check.disabled}, got ${actual.disabled}`);
    }
    return { ...check, actual, pass: failures.length === 0, failures };
  });
}

export async function reduceFixture(original, evaluate, { maxRuns = 100 } = {}) {
  let fixture = structuredClone(original);
  let runs = 0;
  const attempts = [];
  // Restart after each accepted deletion: previously essential keys may become
  // removable when a fallback branch changes. Final pass establishes 1-minimality.
  let changed = true;
  while (changed) {
    changed = false;
    for (const path of paths(fixture)) {
      if (runs >= maxRuns) return { fixture, attempts, complete: false, runs };
      const candidate = remove(fixture, path);
      const observation = await evaluate(candidate, pointer(path));
      runs++;
      attempts.push({ path: pointer(path), accepted: observation.pass, ...observation });
      if (!observation.pass) continue;
      fixture = candidate;
      changed = true;
      break;
    }
  }
  return { fixture, attempts, complete: true, runs };
}

export function checkConfig(config) {
  if (!config || typeof config !== 'object') throw new Error('Config must be an object');
  for (const key of ['name', 'appDir', 'fixture', 'apiPath', 'ready']) {
    if (typeof config[key] !== 'string' || !config[key]) throw new Error(`Config requires ${key}`);
  }
  if (!config.apiPath.startsWith('/') || config.apiPath.includes('?')) throw new Error('apiPath must be an absolute URL path without query');
  if (config.entry !== undefined && (typeof config.entry !== 'string' || !config.entry.startsWith('/') || config.entry.includes('?') || config.entry.includes('#'))) throw new Error('entry must be an absolute URL path without query or fragment');
  if (config.timeoutMs !== undefined && (!Number.isInteger(config.timeoutMs) || config.timeoutMs < 1)) throw new Error('timeoutMs must be a positive integer');
  if (!Array.isArray(config.checks) || !config.checks.length) throw new Error('At least one explicit DOM check is required');
  for (const check of config.checks) {
    if (typeof check.selector !== 'string' || !check.selector) throw new Error('Each check needs a selector');
    if (!['text', 'includes', 'count', 'disabled'].some(key => key in check)) throw new Error('Each check needs an assertion');
    if ('text' in check && typeof check.text !== 'string') throw new Error('text must be a string');
    if ('includes' in check && typeof check.includes !== 'string') throw new Error('includes must be a string');
    if ('count' in check && (!Number.isInteger(check.count) || check.count < 0)) throw new Error('count must be a nonnegative integer');
    if ('disabled' in check && typeof check.disabled !== 'boolean') throw new Error('disabled must be boolean');
  }
  if (config.maxRuns !== undefined && (!Number.isInteger(config.maxRuns) || config.maxRuns < 1)) throw new Error('maxRuns must be a positive integer');
  return config;
}
