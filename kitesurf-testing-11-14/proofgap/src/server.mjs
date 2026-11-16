import http from 'node:http';
import { readFile } from 'node:fs/promises';
import { resolve, sep, extname } from 'node:path';
import { createHash, randomUUID } from 'node:crypto';
import { replacePointer } from './core.mjs';

export async function startServer(config, appRoot, port) {
  const runs = new Map();
  const root = resolve(appRoot);
  const server = http.createServer(async (req, res) => {
    res.setHeader('Cache-Control', 'no-store');
    try {
      const url = new URL(req.url, 'http://127.0.0.1');
      const match = url.pathname.match(/^\/r\/([^/]+)\/(.*)$/);
      if (!match || !runs.has(match[1])) { res.writeHead(404).end('Unknown run'); return; }
      const run = runs.get(match[1]);
      const path = decodeURIComponent(match[2]) || 'index.html';
      const route = config.routes.find(r => r.path === path);
      if (route) {
        const mutated = run.fault?.route === path;
        const body = JSON.stringify(mutated ? replacePointer(route.body, run.fault.pointer, run.fault.value) : route.body);
        run.hits.push({ method: req.method, route: path, mutated, bodySha256: createHash('sha256').update(body).digest('hex') });
        res.setHeader('Content-Type', 'application/json');
        res.end(body);
        return;
      }
      const file = resolve(root, path);
      if (!file.startsWith(root + sep)) { res.writeHead(403).end('Forbidden'); return; }
      const bytes = await readFile(file);
      res.setHeader('Content-Type', ({ '.html': 'text/html', '.js': 'text/javascript', '.css': 'text/css', '.json': 'application/json' })[extname(file)] ?? 'application/octet-stream');
      res.end(bytes);
    } catch (error) { res.writeHead(error.code === 'ENOENT' ? 404 : 500).end('Local fixture request failed'); }
  });
  await new Promise((done, fail) => { server.once('error', fail); server.listen(port, '127.0.0.1', done); });
  return {
    open(fault = null) {
      const token = randomUUID();
      const run = { fault, hits: [], url: `http://127.0.0.1:${server.address().port}/r/${token}/` };
      runs.set(token, run);
      return run;
    },
    close: () => new Promise(resolve => { server.close(resolve); server.closeAllConnections(); }),
  };
}
