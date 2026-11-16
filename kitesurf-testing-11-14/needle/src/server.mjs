import {createServer} from 'node:http';
import {readFile} from 'node:fs/promises';
const port = Number(process.env.PORT ?? 8892);
const pages = new Map([['/', 'search.html'], ['/search', 'search.html'], ['/cart', 'cart.html']]);
const server = createServer(async (request, response) => {
  const url = new URL(request.url, 'http://localhost');
  try {
    if (url.pathname === '/health') {response.end('needle-demo'); return;}
    if (url.pathname === '/api/search') {
      const query = url.searchParams.get('q');
      await new Promise(resolve => setTimeout(resolve, query.length <= 2 ? 400 : 30));
      response.setHeader('content-type', 'application/json'); response.end(JSON.stringify({label: 'Matches for ' + query})); return;
    }
    if (url.pathname === '/api/quote') {
      const quantity = Number(url.searchParams.get('quantity'));
      await new Promise(resolve => setTimeout(resolve, quantity <= 2 ? 400 : 30));
      response.setHeader('content-type', 'application/json'); response.end(JSON.stringify({total: quantity * 20})); return;
    }
    if (!pages.has(url.pathname)) {response.writeHead(404); response.end('Not found'); return;}
    response.setHeader('content-type', 'text/html; charset=utf-8');
    response.end(await readFile(new URL('../examples/' + pages.get(url.pathname), import.meta.url)));
  } catch (error) {response.writeHead(500); response.end('Demo server error'); console.error(error);}
});
server.listen(port, '127.0.0.1', () => console.log(`Needle demo at http://127.0.0.1:${port}`));
