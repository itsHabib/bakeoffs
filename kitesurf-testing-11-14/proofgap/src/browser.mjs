export class CDP {
  constructor(ws) {
    this.ws = ws; this.next = 1; this.pending = new Map();
    ws.addEventListener('message', event => {
      const data = JSON.parse(event.data);
      const p = this.pending.get(data.id);
      if (!p) return;
      clearTimeout(p.timer); this.pending.delete(data.id);
      data.error ? p.reject(new Error(JSON.stringify(data.error))) : p.resolve(data.result);
    });
    ws.addEventListener('close', () => {
      for (const p of this.pending.values()) { clearTimeout(p.timer); p.reject(new Error('CDP disconnected')); }
      this.pending.clear();
    });
  }
  static async connect(endpoint) {
    const ws = new WebSocket(endpoint);
    await new Promise((resolve, reject) => {
      const timer = setTimeout(() => { ws.close(); reject(new Error('CDP connect timeout')); }, 10000);
      ws.addEventListener('open', () => { clearTimeout(timer); resolve(); }, { once: true });
      ws.addEventListener('error', () => { clearTimeout(timer); reject(new Error('CDP connection failed')); }, { once: true });
    });
    return new CDP(ws);
  }
  send(method, params = {}, sessionId) {
    if (this.ws.readyState !== WebSocket.OPEN) return Promise.reject(new Error('CDP connection is not open'));
    const id = this.next++;
    return new Promise((resolve, reject) => {
      const timer = setTimeout(() => { this.pending.delete(id); reject(new Error(`CDP timeout: ${method}`)); }, 30000);
      this.pending.set(id, { resolve, reject, timer });
      this.ws.send(JSON.stringify({ id, method, params, ...(sessionId ? { sessionId } : {}) }));
    });
  }
  close() { this.ws.close(); }
}

export async function identity(endpoint) {
  const url = new URL(endpoint); url.protocol = url.protocol === 'wss:' ? 'https:' : 'http:'; url.pathname = '/json/version';
  const response = await fetch(url, { signal: AbortSignal.timeout(10000) });
  if (!response.ok) throw new Error(`Engine discovery HTTP ${response.status}`);
  const data = await response.json();
  if (!/^Kitesurf\//i.test(data.Browser ?? '')) throw new Error(`Expected Kitesurf, received ${data.Browser}`);
  return data;
}

export async function runBrowser(cdp, config, run) {
  let browserContextId;
  const result = { snapshot: {}, hits: run.hits, elapsedMs: 0, error: null };
  const start = Date.now();
  try {
    ({ browserContextId } = await cdp.send('Target.createBrowserContext'));
    const { targetId } = await cdp.send('Target.createTarget', { url: 'about:blank', browserContextId });
    const { sessionId } = await cdp.send('Target.attachToTarget', { targetId, flatten: true });
    const evaluate = async expression => {
      const r = await cdp.send('Runtime.evaluate', { expression, returnByValue: true, awaitPromise: true }, sessionId);
      if (r.exceptionDetails) throw new Error(`Page evaluation: ${JSON.stringify(r.exceptionDetails)}`);
      return r.result?.value;
    };
    const waitFor = async selector => {
      const until = Date.now() + (config.timeoutMs ?? 12000);
      do {
        if (await evaluate(`!!document.querySelector(${JSON.stringify(selector)})`)) return;
        await new Promise(r => setTimeout(r, 100));
      } while (Date.now() < until);
      throw new Error(`Readiness timeout: ${selector}`);
    };
    await cdp.send('Page.enable', {}, sessionId);
    const navigation = await cdp.send('Page.navigate', { url: run.url }, sessionId);
    if (navigation.errorText) throw new Error(`Navigation: ${navigation.errorText}`);
    for (const step of config.steps ?? []) {
      await waitFor(step.selector);
      if (step.action === 'wait') continue;
      await evaluate(`(() => {const el=document.querySelector(${JSON.stringify(step.selector)});${step.action === 'click' ? 'el.click();' : `el.value=${JSON.stringify(step.value)}; el.dispatchEvent(new Event('input',{bubbles:true})); el.dispatchEvent(new Event('change',{bubbles:true}));`}return true;})()`);
    }
    await waitFor(config.ready);
    const selectors = [...new Set([...config.faults.map(f => f.witness), ...Object.values(config.suites).flat().map(a => a.selector)])];
    result.snapshot = await evaluate(`(() => {const out={}; for(const s of ${JSON.stringify(selectors)}) {const e=document.querySelector(s);out[s]=e?e.textContent.trim():null;}return out;})()`);
    if (!result.snapshot || typeof result.snapshot !== 'object') throw new Error('Engine returned no DOM snapshot');
  } catch (error) { result.error = error.message; }
  finally {
    result.elapsedMs = Date.now() - start;
    if (browserContextId) {
      try { await cdp.send('Target.disposeBrowserContext', { browserContextId }); }
      catch (error) { result.error = [result.error, `Context cleanup: ${error.message}`].filter(Boolean).join('; '); }
    }
  }
  return result;
}
