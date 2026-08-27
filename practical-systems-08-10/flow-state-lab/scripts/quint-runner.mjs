import { spawn, spawnSync } from 'node:child_process';
import { createConnection } from 'node:net';
import { join, resolve } from 'node:path';
import { createRequire } from 'node:module';
import { fileURLToPath } from 'node:url';

const projectRoot = resolve(fileURLToPath(new URL('..', import.meta.url)));
const quintCli = join(
  projectRoot,
  'node_modules',
  '@informalsystems',
  'quint',
  'dist',
  'src',
  'cli.js',
);
const require = createRequire(import.meta.url);
const quintApalache = require(join(
  projectRoot,
  'node_modules',
  '@informalsystems',
  'quint',
  'dist',
  'src',
  'apalache.js',
));

const serverEndpoint = 'localhost:8822';

function canConnect(port = 8822, host = '127.0.0.1') {
  return new Promise((resolveConnection) => {
    const socket = createConnection({ port, host });
    socket.setTimeout(250);
    socket.once('connect', () => {
      socket.destroy();
      resolveConnection(true);
    });
    const fail = () => {
      socket.destroy();
      resolveConnection(false);
    };
    socket.once('error', fail);
    socket.once('timeout', fail);
  });
}

async function waitForServer(child, attempts = 80) {
  let spawnError = null;
  child.once('error', (error) => {
    spawnError = error;
  });
  for (let attempt = 0; attempt < attempts; attempt += 1) {
    if (await canConnect()) return;
    if (spawnError) throw spawnError;
    if (child.exitCode !== null) {
      throw new Error(`Apalache server exited before accepting connections (${child.exitCode}).`);
    }
    await new Promise((resolveWait) => setTimeout(resolveWait, 250));
  }
  throw new Error('Timed out waiting for the local Apalache server on port 8822.');
}

async function startWindowsServer() {
  // Quint's downloader works on Windows; only its direct spawn of the resulting
  // .bat file is broken. Reuse the pinned CLI's downloader, then launch that
  // exact executable through cmd.exe.
  const fetched = await quintApalache.fetchApalache(
    quintApalache.DEFAULT_APALACHE_VERSION_TAG,
    0,
  );
  if (fetched.isLeft()) {
    throw new Error(`Failed to install Apalache: ${fetched.value.explanation}`);
  }
  const executable = fetched.value;

  const command = `"${executable}" server --port=8822`;
  const child = spawn('cmd.exe', ['/d', '/s', '/c', command], {
    stdio: 'ignore',
    windowsHide: true,
  });
  try {
    await waitForServer(child);
  } catch (error) {
    stopWindowsServer(child);
    throw error;
  }
  return child;
}

function stopWindowsServer(child) {
  if (!child || child.exitCode !== null) return;
  spawnSync('taskkill.exe', ['/pid', String(child.pid), '/T', '/F'], {
    stdio: 'ignore',
    windowsHide: true,
  });
}

export async function withApalache(callback) {
  let child = null;
  const existingServer = await canConnect();

  if (process.platform === 'win32' && !existingServer) {
    child = await startWindowsServer();
  }

  try {
    return await callback({ serverEndpoint });
  } finally {
    if (process.platform === 'win32') stopWindowsServer(child);
  }
}

export function runQuint(args, expectedStatus = 0) {
  const result = spawnSync(process.execPath, [quintCli, ...args], {
    cwd: projectRoot,
    encoding: 'utf8',
    windowsHide: true,
  });

  if (result.error) throw result.error;
  if (result.status !== expectedStatus) {
    throw new Error(
      [
        `Quint exited ${result.status}; expected ${expectedStatus}.`,
        result.stdout,
        result.stderr,
      ].filter(Boolean).join('\n'),
    );
  }

  return result;
}

export { projectRoot };
