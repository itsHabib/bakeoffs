import test from 'node:test';
import assert from 'node:assert/strict';
import { mkdtemp, writeFile, mkdir, symlink, rm } from 'node:fs/promises';
import { tmpdir } from 'node:os';
import { join } from 'node:path';
import { startServer } from '../src/server.mjs';

test('serves the current mock and prevents symlink escape from appDir', async () => {
  const dir = await mkdtemp(join(tmpdir(), 'fixture-pruner-test-'));
  let server;
  try {
    await mkdir(join(dir, 'app'));
    await writeFile(join(dir, 'app/index.html'), '<h1>Local</h1>');
    await writeFile(join(dir, 'private.txt'), 'private');
    await symlink(join(dir, 'private.txt'), join(dir, 'app/escape.txt'));
    server = await startServer({ appDir: join(dir, 'app'), apiPath: '/api/data', port: 0 });
    server.setFixture({ value: 1 });
    assert.deepEqual(await (await fetch(server.url + '/api/data')).json(), { value: 1 });
    assert.equal(server.hits(), 1);
    server.setFixture({});
    assert.equal(server.hits(), 0);
    assert.deepEqual(await (await fetch(server.url + '/api/data')).json(), {});
    assert.equal(await (await fetch(server.url)).text(), '<h1>Local</h1>');
    assert.equal((await fetch(server.url + '/escape.txt')).status, 403);
  } finally {
    if (server) await server.close();
    await rm(dir, { recursive: true, force: true });
  }
});
