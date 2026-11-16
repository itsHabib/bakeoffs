import puppeteer from 'puppeteer-core';
import { assess } from './policy.mjs';

export class InfrastructureError extends Error {}

export async function connect(endpoint) {
  const endpointUrl = new URL(endpoint);
  if (!['127.0.0.1', 'localhost', '[::1]'].includes(endpointUrl.hostname)) throw new InfrastructureError('Only a local Kitesurf endpoint is supported');
  const discovery = new URL('/json/version', endpoint);
  discovery.protocol = endpointUrl.protocol === 'wss:' ? 'https:' : 'http:';
  let identity;
  try {
    const response = await fetch(discovery, { signal: AbortSignal.timeout(4000) });
    if (!response.ok) throw new Error(`HTTP ${response.status}`);
    identity = await response.json();
    if (!JSON.stringify(identity).toLowerCase().includes('kitesurf')) throw new Error('Discovery did not identify Kitesurf');
  } catch (error) { throw new InfrastructureError(`Kitesurf discovery unavailable: ${error.message}`); }
  try {
    const browser = await puppeteer.connect({ browserWSEndpoint: endpoint, protocolTimeout: 15000, defaultViewport: null });
    return { browser, identity };
  } catch (error) { throw new InfrastructureError(`Kitesurf CDP connect failed: ${error.message}`); }
}

export async function observe(browser, url, config) {
  let page, context;
  const errors = [];
  try {
    context = await browser.createBrowserContext();
    page = await context.newPage();
    page.on('pageerror', error => errors.push(error.message));
    await page.goto(url, { waitUntil: 'domcontentloaded', timeout: 15000 });
    const deadline = Date.now() + (config.timeoutMs ?? 5000);
    let ready = false;
    while (Date.now() < deadline) {
      ready = await page.evaluate(selector => !!document.querySelector(selector), config.ready);
      if (ready) break;
      await new Promise(resolve => setTimeout(resolve, 50));
    }
    if (!ready) return { pass: false, reason: 'Consumer never reached ready selector', checks: [], errors };
    const observed = await page.evaluate(checks => checks.map(check => {
      try {
        const nodes = document.querySelectorAll(check.selector);
        return { count: nodes.length, text: Array.from(nodes).map(node => node.textContent.trim()).join('\n'), disabled: nodes.length ? nodes[0].hasAttribute('disabled') : null };
      } catch (error) { return { error: String(error) }; }
    }), config.checks);
    const checks = assess(config.checks, observed);
    return { pass: checks.every(check => check.pass) && errors.length === 0, checks, errors };
  } catch (error) {
    throw new InfrastructureError(`Browser operation failed; candidate is inconclusive: ${error.message}`);
  } finally {
    if (context) await context.close().catch(() => {});
  }
}
