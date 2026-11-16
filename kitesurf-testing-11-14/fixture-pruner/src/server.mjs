import http from 'node:http';
import { readFile, realpath } from 'node:fs/promises';
import { resolve, extname, sep } from 'node:path';

export async function startServer({ appDir, apiPath, port = 8891 }) {
  let fixture;
  let apiHits = 0;
  const root = await realpath(appDir);
  const server = http.createServer(async (req, res) => {
    res.setHeader('Cache-Control', 'no-store');
    const url = new URL(req.url, 'http://127.0.0.1');
    if (url.pathname === apiPath) {
      apiHits++;
      res.setHeader('Content-Type', 'application/json');
      res.end(JSON.stringify(fixture));
      return;
    }
    try {
      const relative = decodeURIComponent(url.pathname === '/' ? '/index.html' : url.pathname);
      const file = await realpath(resolve(root, '.' + relative));
      if (!file.startsWith(root + sep)) { res.writeHead(403).end(); return; }
      const body = await readFile(file);
      res.setHeader('Content-Type', ({ '.html': 'text/html', '.js': 'text/javascript', '.css': 'text/css', '.json': 'application/json', '.svg': 'image/svg+xml' })[extname(file)] ?? 'application/octet-stream');
      res.end(body);
    } catch { res.writeHead(404).end('Not found'); }
  });
  await new Promise((accept, reject) => { server.once('error', reject); server.listen(port, '127.0.0.1', accept); });
  return {
    url: `http://127.0.0.1:${server.address().port}`,
    setFixture(value) { fixture = value; apiHits = 0; },
    hits() { return apiHits; },
    close: () => new Promise(resolve => server.close(resolve)),
  };
}
