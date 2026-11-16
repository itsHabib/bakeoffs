export function validate(scenario) {
  if (!scenario || typeof scenario.url !== 'string' || !Array.isArray(scenario.events) || !scenario.events.length) throw new Error('A URL and nonempty events array are required');
  const url = new URL(scenario.url);
  if (!['localhost', '127.0.0.1', '[::1]'].includes(url.hostname)) throw new Error('Needle only runs against local URLs');
  const seen = new Set();
  for (const event of scenario.events) {
    if (typeof event.id !== 'string' || !event.id || seen.has(event.id)) throw new Error('Every event needs a unique string id');
    if (!['fill', 'click', 'respond'].includes(event.kind)) throw new Error(`Unknown event kind: ${event.kind}`);
    if (!Array.isArray(event.requires ?? [])) throw new Error('requires must be an array');
    if ((event.requires ?? []).some(id => !seen.has(id))) throw new Error(`Dependencies must precede ${event.id}`);
    if (event.kind === 'respond' && (typeof event.url !== 'string' || !Object.hasOwn(event, 'body'))) throw new Error('respond needs an exact URL and JSON body');
    if (event.kind !== 'respond' && typeof event.selector !== 'string') throw new Error('DOM actions need a selector');
    if (event.kind === 'fill' && typeof event.value !== 'string') throw new Error('fill needs a string value');
    seen.add(event.id);
  }
  for (const side of ['actual', 'expected']) {
    const ref = scenario.oracle?.[side];
    if (!ref || typeof ref.selector !== 'string' || !ref.selector || (ref.property && ref.attribute) || (ref.property !== undefined && typeof ref.property !== 'string') || (ref.attribute !== undefined && typeof ref.attribute !== 'string')) throw new Error(`Invalid ${side} DOM reference`);
  }
  return scenario;
}

// Removing a cause removes its declared dependants. Order is always preserved.
export function without(events, removed) {
  const kept = new Set();
  return events.filter(event => {
    if (removed.has(event.id) || (event.requires ?? []).some(id => !kept.has(id))) return false;
    kept.add(event.id);
    return true;
  });
}

export async function reduce(events, probe) {
  let current = [...events];
  let granularity = 2;
  while (current.length) {
    const size = Math.ceil(current.length / granularity);
    let changed = false;
    for (let start = 0; start < current.length; start += size) {
      const trial = without(current, new Set(current.slice(start, start + size).map(event => event.id)));
      const result = await probe(trial);
      if (result.status !== 'fail') continue;
      current = trial;
      granularity = Math.max(2, granularity - 1);
      changed = true;
      break;
    }
    if (changed) continue;
    if (granularity >= current.length) break;
    granularity = Math.min(current.length, granularity * 2);
  }
  // Independent final witnesses: one deletion (plus dependent closure) per event.
  const witnesses = [];
  for (const event of current) {
    const candidate = without(current, new Set([event.id]));
    const result = await probe(candidate, {fresh: true});
    if (result.status === 'fail') throw new Error(`Unstable reduction: ${event.id} is still removable`);
    witnesses.push({removed: event.id, retained: candidate.map(e => e.id), ...result});
  }
  return {events: current, witnesses};
}

export function classify(observation, signature) {
  if (observation.status === 'unresolved') return observation;
  if (typeof observation.actual !== 'string' || typeof observation.expected !== 'string') throw new Error('Invalid oracle observation');
  if (observation.actual === observation.expected) return {status: 'pass', ...observation};
  if (signature && (observation.actual !== signature.actual || observation.expected !== signature.expected)) return {status: 'different-failure', ...observation};
  return {status: 'fail', ...observation};
}

export async function repeated(run, count = 2) {
  const results = [];
  for (let i = 0; i < count; i++) results.push(await run());
  if (results.some(result => JSON.stringify(result) !== JSON.stringify(results[0]))) return {status: 'unstable', observations: results};
  return results[0];
}
