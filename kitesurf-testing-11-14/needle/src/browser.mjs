export class CDP {
  constructor(ws) {
    this.ws = ws;
    this.next = 0;
    this.pending = new Map();
    ws.addEventListener('message', event => {
      const message = JSON.parse(String(event.data));
      const entry = this.pending.get(message.id);
      if (!entry) return;
      clearTimeout(entry.timer);
      this.pending.delete(message.id);
      if (message.error) entry.reject(new Error(`CDP: ${JSON.stringify(message.error)}`));
      else entry.resolve(message.result);
    });
    ws.addEventListener('close', () => {
      for (const entry of this.pending.values()) {clearTimeout(entry.timer); entry.reject(new Error('Kitesurf connection closed'));}
      this.pending.clear();
    });
  }
  static async connect(endpoint) {
    const url = new URL(endpoint);
    if (url.protocol !== 'ws:' || !['localhost', '127.0.0.1', '[::1]'].includes(url.hostname)) throw new Error('Only a local ws:// engine is supported');
    const ws = new WebSocket(endpoint);
    await new Promise((resolve, reject) => {
      const timer = setTimeout(() => {ws.close(); reject(new Error('Engine connection timed out'));}, 10000);
      ws.addEventListener('open', () => {clearTimeout(timer); resolve();}, {once: true});
      ws.addEventListener('error', () => {clearTimeout(timer); reject(new Error(`Cannot reach Kitesurf at ${endpoint}`));}, {once: true});
    });
    return new CDP(ws);
  }
  send(method, params = {}, sessionId) {
    const id = ++this.next;
    return new Promise((resolve, reject) => {
      const timer = setTimeout(() => {this.pending.delete(id); reject(new Error(`CDP timeout: ${method}`));}, 20000);
      this.pending.set(id, {resolve, reject, timer});
      this.ws.send(JSON.stringify({id, method, params, ...(sessionId ? {sessionId} : {})}));
    });
  }
  close() {this.ws.close();}
}

// Narrow fetch stub: controlled JSON/text responses to requests issued after installation.
export function installMock() {
  window.__needle = {pending: [], errors: []};
  window.addEventListener('error', event => window.__needle.errors.push(String(event.message)));
  window.addEventListener('unhandledrejection', event => window.__needle.errors.push(String(event.reason)));
  window.fetch = function(input) {
    const url = typeof input === 'string' ? input : input.url;
    return new Promise(resolve => window.__needle.pending.push({url, resolve}));
  };
  return true;
}

export async function applyEvent(event) {
  if (event.kind === 'respond') {
    const index = window.__needle.pending.findIndex(request => request.url === event.url);
    if (index < 0) return {status: 'unresolved', reason: `No pending fetch for ${event.url}`};
    const request = window.__needle.pending.splice(index, 1)[0];
    request.resolve({ok: true, status: 200, json: () => Promise.resolve(event.body), text: () => Promise.resolve(JSON.stringify(event.body))});
    await Promise.resolve();
    await Promise.resolve();
    return {status: 'applied'};
  }
  const element = document.querySelector(event.selector);
  if (!element) return {status: 'unresolved', reason: `Missing action target ${event.selector}`};
  if (event.kind === 'fill') {
    element.value = event.value;
    element.dispatchEvent(new Event('input', {bubbles: true}));
  } else element.click();
  await Promise.resolve();
  return {status: 'applied'};
}

export function observe(oracle) {
  if (window.__needle.errors.length) return {status: 'page-error', errors: window.__needle.errors};
  if (window.__needle.pending.length) return {status: 'unresolved', reason: `${window.__needle.pending.length} fetch request(s) remain undelivered`};
  function read(ref) {
    const element = document.querySelector(ref.selector);
    if (!element) return null;
    const value = ref.attribute ? element.getAttribute(ref.attribute) : ref.property ? element[ref.property] : element.textContent;
    return value === null || value === undefined ? null : String(value);
  }
  const actual = read(oracle.actual), expected = read(oracle.expected);
  if (actual === null || expected === null) return {status: 'unresolved', reason: 'Oracle target or value missing'};
  return {actual, expected};
}

export async function evaluate(cdp, session, fn, argument) {
  const result = await cdp.send('Runtime.evaluate', {
    expression: `(${fn.toString()})(${JSON.stringify(argument) ?? ''})`,
    returnByValue: true, awaitPromise: true,
  }, session);
  if (result.exceptionDetails) throw new Error(`Page execution failed: ${JSON.stringify(result.exceptionDetails)}`);
  return result.result.value;
}

export async function replay(cdp, scenario, events) {
  const {targetId} = await cdp.send('Target.createTarget', {url: 'about:blank'});
  try {
    const {sessionId} = await cdp.send('Target.attachToTarget', {targetId, flatten: true});
    await cdp.send('Page.enable', {}, sessionId);
    const navigation = await cdp.send('Page.navigate', {url: scenario.url}, sessionId);
    if (navigation.errorText) throw new Error(`Navigation failed: ${navigation.errorText}`);
    // Navigation completion differs by engine. Wait on an explicit page marker.
    let ready = false;
    for (let attempt = 0; attempt < 60; attempt++) {
      ready = await evaluate(cdp, sessionId, selector => !!document.querySelector(selector), scenario.readySelector ?? scenario.oracle.expected.selector);
      if (ready) break;
      await new Promise(resolve => setTimeout(resolve, 50));
    }
    if (!ready) throw new Error('Page readiness marker never appeared');
    await evaluate(cdp, sessionId, installMock);
    for (const event of events) {
      const result = await evaluate(cdp, sessionId, applyEvent, event);
      if (result.status !== 'applied') return result;
    }
    const observation = await evaluate(cdp, sessionId, observe, scenario.oracle);
    if (observation.status === 'page-error') throw new Error(`Page error: ${JSON.stringify(observation.errors)}`);
    return observation;
  } finally {
    await cdp.send('Target.closeTarget', {targetId});
  }
}
