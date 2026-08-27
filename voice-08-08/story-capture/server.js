import { createReadStream, existsSync, statSync } from "node:fs";
import { createServer } from "node:http";
import { extname, join, normalize } from "node:path";
import { fileURLToPath } from "node:url";

import { createRealtimeSessionPayload } from "./lib/realtime-config.js";

const root = fileURLToPath(new URL(".", import.meta.url));
const publicRoot = join(root, "public");
const port = Number.parseInt(process.env.PORT || "4312", 10);

const contentTypes = new Map([
  [".css", "text/css; charset=utf-8"],
  [".html", "text/html; charset=utf-8"],
  [".js", "text/javascript; charset=utf-8"],
  [".json", "application/json; charset=utf-8"],
  [".ico", "image/x-icon"],
]);

function securityHeaders(response) {
  response.setHeader("Content-Security-Policy", [
    "default-src 'self'",
    "connect-src 'self' https://api.openai.com",
    "img-src 'self' data:",
    "media-src 'self' blob:",
    "style-src 'self'",
    "script-src 'self'",
  ].join("; "));
  response.setHeader("Referrer-Policy", "no-referrer");
  response.setHeader("X-Content-Type-Options", "nosniff");
  response.setHeader("X-Frame-Options", "DENY");
}

function sendJSON(response, status, body) {
  securityHeaders(response);
  response.writeHead(status, {
    "Cache-Control": "no-store",
    "Content-Type": "application/json; charset=utf-8",
  });
  response.end(JSON.stringify(body));
}

async function mintRealtimeSecret(response) {
  const apiKey = process.env.OPENAI_API_KEY?.trim();
  if (!apiKey) {
    sendJSON(response, 503, {
      error: "Live interview is not configured. Canned replay is still available.",
    });
    return;
  }

  let upstream;
  try {
    upstream = await fetch("https://api.openai.com/v1/realtime/client_secrets", {
      method: "POST",
      headers: {
        Authorization: `Bearer ${apiKey}`,
        "Content-Type": "application/json",
        "OpenAI-Safety-Identifier": "hack2-story-capture-local-demo",
      },
      body: JSON.stringify({ session: createRealtimeSessionPayload() }),
      signal: AbortSignal.timeout(20_000),
    });
  } catch {
    sendJSON(response, 502, { error: "Could not reach the live voice service." });
    return;
  }

  if (!upstream.ok) {
    sendJSON(response, 502, {
      error: `The live voice service declined the session (${upstream.status}).`,
    });
    return;
  }

  let secret;
  try {
    secret = await upstream.json();
  } catch {
    sendJSON(response, 502, { error: "The live voice service returned an invalid session." });
    return;
  }

  if (!secret?.value) {
    sendJSON(response, 502, { error: "The live voice service returned an empty session." });
    return;
  }

  // Deliberately return only the short-lived browser credential and expiry.
  sendJSON(response, 200, { value: secret.value, expires_at: secret.expires_at });
}

function serveStatic(request, response) {
  const requestPath = request.url === "/" ? "/index.html" : request.url.split("?")[0];
  const normalized = normalize(decodeURIComponent(requestPath)).replace(/^(\.\.(\/|\\|$))+/, "");
  let filePath = join(publicRoot, normalized);

  if (requestPath.startsWith("/lib/")) {
    filePath = join(root, normalized);
  }

  const allowedRoots = [publicRoot, join(root, "lib")];
  const allowed = allowedRoots.some((allowedRoot) => filePath.startsWith(`${allowedRoot}/`));
  if (!allowed || !existsSync(filePath) || !statSync(filePath).isFile()) {
    sendJSON(response, 404, { error: "Not found" });
    return;
  }

  securityHeaders(response);
  response.writeHead(200, {
    "Cache-Control": "no-store",
    "Content-Type": contentTypes.get(extname(filePath)) || "application/octet-stream",
  });
  createReadStream(filePath).pipe(response);
}

const server = createServer(async (request, response) => {
  if (request.method === "POST" && request.url === "/api/realtime-token") {
    await mintRealtimeSecret(response);
    return;
  }
  if (request.method !== "GET" && request.method !== "HEAD") {
    sendJSON(response, 405, { error: "Method not allowed" });
    return;
  }
  serveStatic(request, response);
});

server.listen(port, "127.0.0.1", () => {
  console.log(`Keepsake is ready at http://127.0.0.1:${port}`);
});
