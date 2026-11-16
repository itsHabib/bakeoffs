// Normal Chrome is used only to inspect generated report presentation.
// It is never used for fixture reduction or as a Kitesurf fallback.
import puppeteer from 'puppeteer-core';
import { mkdir, writeFile } from 'node:fs/promises';
import { startServer } from '../src/server.mjs';

await mkdir('evidence', { recursive: true });
const server = await startServer({ appDir: 'artifacts', apiPath: '/__unused', port: 0 });
const browser = await puppeteer.launch({
  executablePath: process.env.CHROME_EXECUTABLE ?? '/Applications/Google Chrome.app/Contents/MacOS/Google Chrome',
  headless: true,
});
const results = [];
try {
  for (const workload of ['invoices', 'warehouse']) {
    const page = await browser.newPage();
    for (const width of [1120, 390]) {
      await page.setViewport({ width, height: 820 });
      await page.goto(`${server.url}/${workload}/index.html`);
      const result = await page.evaluate(() => ({ title: document.title, width: innerWidth, scrollWidth: document.documentElement.scrollWidth, controls: document.querySelectorAll('tbody tr').length, links: Array.from(document.querySelectorAll('a')).map(a => a.getAttribute('href')) }));
      if (result.scrollWidth > width) throw new Error(`Horizontal overflow in ${workload} at ${width}`);
      results.push({ workload, ...result });
      await page.screenshot({ path: `evidence/${workload}-report-${width}.png`, fullPage: true });
    }
    for (const href of results.at(-1).links) {
      const response = await fetch(new URL(href, `${server.url}/${workload}/index.html`));
      if (!response.ok) throw new Error(`Broken report artifact link ${href}`);
      await response.json();
    }
    await page.close();
  }
  await writeFile('evidence/report-inspection.json', JSON.stringify({ browser: await browser.version(), results }, null, 2) + '\n');
  console.log(JSON.stringify(results, null, 2));
} finally {
  await browser.close();
  await server.close();
}
