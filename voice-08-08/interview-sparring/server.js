import {createReadStream} from "node:fs";
import {stat} from "node:fs/promises";
import http from "node:http";
import path from "node:path";
import {fileURLToPath, pathToFileURL} from "node:url";

const ROOT = path.dirname(fileURLToPath(import.meta.url));
const PUBLIC_ROOT = path.join(ROOT, "public");
const CORE_PATH = path.join(ROOT, "lib", "interview-core.js");
const MIME_TYPES = new Map([
  [".css", "text/css; charset=utf-8"],
  [".html", "text/html; charset=utf-8"],
  [".js", "text/javascript; charset=utf-8"],
  [".svg", "image/svg+xml"],
]);

export function realtimeSessionPayload(env = process.env) {
  return {
    session: {
      type: "realtime",
      model: env.OPENAI_REALTIME_MODEL || "gpt-realtime-2.1",
      instructions: [
        "You are a concise behavioral interviewer with natural conversational timing.",
        "The browser's deterministic state machine owns every question, follow-up topic, interruption, metric, and score.",
        "Only voice the exact prompt supplied in each response.create instruction. Never grade, diagnose, choose a follow-up, or answer for the candidate.",
        "Speak one brief sentence and stop. Be firm, warm, and direct.",
      ].join(" "),
      output_modalities: ["audio"],
      audio: {
        input: {
          transcription: {model: "gpt-4o-mini-transcribe"},
          turn_detection: {
            type: "semantic_vad",
            eagerness: "high",
            create_response: false,
            interrupt_response: true,
          },
        },
        output: {voice: env.OPENAI_REALTIME_VOICE || "marin", speed: 1.05},
      },
    },
  };
}

export async function mintRealtimeSecret(env = process.env, fetchImpl = fetch) {
  const apiKey = String(env.OPENAI_API_KEY || "").trim();
  if (!apiKey) {
    const error = new Error("Realtime voice is not configured on this server.");
    error.statusCode = 503;
    throw error;
  }
  const endpoint = env.OPENAI_REALTIME_SECRETS_URL || "https://api.openai.com/v1/realtime/client_secrets";
  let response;
  try {
    response = await fetchImpl(endpoint, {
      method: "POST",
      headers: {
        Authorization: `Bearer ${apiKey}`,
        "Content-Type": "application/json",
        "OpenAI-Safety-Identifier": "hack2-interview-sparring-local-demo",
      },
      body: JSON.stringify(realtimeSessionPayload(env)),
      signal: AbortSignal.timeout(20_000),
    });
  } catch {
    const error = new Error("Realtime voice service could not be reached.");
    error.statusCode = 502;
    throw error;
  }
  if (!response.ok) {
    const error = new Error(`Realtime voice service rejected the request (${response.status}).`);
    error.statusCode = 502;
    throw error;
  }
  const secret = await response.json();
  if (!secret?.value) {
    const error = new Error("Realtime voice service returned an invalid secret.");
    error.statusCode = 502;
    throw error;
  }
  return {
    value: secret.value,
    expires_at: secret.expires_at,
    calls_url: env.OPENAI_REALTIME_CALLS_URL || "https://api.openai.com/v1/realtime/calls",
  };
}

export function createAppServer({env = process.env, fetchImpl = fetch} = {}) {
  return http.createServer(async (request, response) => {
    setSecurityHeaders(response);
    try {
      const url = new URL(request.url, "http://localhost");
      if (request.method === "POST" && url.pathname === "/api/realtime/token") {
        const secret = await mintRealtimeSecret(env, fetchImpl);
        return sendJSON(response, 200, secret);
      }
      if (request.method !== "GET" && request.method !== "HEAD") {
        return sendJSON(response, 405, {error: "Method not allowed."});
      }
      let filePath;
      if (url.pathname === "/interview-core.js") {
        filePath = CORE_PATH;
      } else {
        const relative = url.pathname === "/" ? "index.html" : decodeURIComponent(url.pathname.slice(1));
        filePath = path.resolve(PUBLIC_ROOT, relative);
        if (filePath !== PUBLIC_ROOT && !filePath.startsWith(`${PUBLIC_ROOT}${path.sep}`)) {
          return sendJSON(response, 404, {error: "Not found."});
        }
      }
      const info = await stat(filePath);
      if (!info.isFile()) return sendJSON(response, 404, {error: "Not found."});
      response.writeHead(200, {
        "Content-Type": MIME_TYPES.get(path.extname(filePath)) || "application/octet-stream",
        "Content-Length": info.size,
        "Cache-Control": "no-store",
      });
      if (request.method === "HEAD") return response.end();
      createReadStream(filePath).pipe(response);
    } catch (error) {
      const status = error?.code === "ENOENT" ? 404 : error?.statusCode || 500;
      const message = status === 500 ? "Internal server error." : error.message;
      sendJSON(response, status, {error: message});
    }
  });
}

function setSecurityHeaders(response) {
  response.setHeader("Content-Security-Policy", "default-src 'self'; connect-src 'self' https://api.openai.com; img-src 'self' data:; media-src 'self' blob:; script-src 'self'; style-src 'self'; base-uri 'none'; form-action 'none'");
  response.setHeader("Permissions-Policy", "microphone=(self)");
  response.setHeader("Referrer-Policy", "no-referrer");
  response.setHeader("X-Content-Type-Options", "nosniff");
  response.setHeader("X-Frame-Options", "DENY");
}

function sendJSON(response, status, value) {
  if (response.writableEnded) return;
  const body = JSON.stringify(value);
  response.writeHead(status, {"Content-Type": "application/json; charset=utf-8", "Content-Length": Buffer.byteLength(body)});
  response.end(body);
}

if (process.argv[1] && import.meta.url === pathToFileURL(process.argv[1]).href) {
  const port = Number(process.env.PORT || 4173);
  const server = createAppServer();
  server.listen(port, "127.0.0.1", () => {
    console.log(`Interview Sparring is ready at http://127.0.0.1:${port}`);
  });
}
