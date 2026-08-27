"""Model sources: live ollama, or a committed cassette for deterministic demos.

The gate only needs one thing from a model: ``generate(prompt) -> raw_text``.
Two implementations satisfy that:

  - OllamaModel  : POST to a local ollama server running qwen2.5:7b, JSON mode.
  - CassetteModel: replay a committed sequence of raw responses (attempt 1, 2).

Cassettes make ``make demo`` fully deterministic and hands-free even if ollama
is down or the model drifts. NEVER a frontier model — this machine is keyless;
REJECT is the fallback, not a bigger model.
"""

from __future__ import annotations

import json
import urllib.error
import urllib.request

DEFAULT_OLLAMA_URL = "http://localhost:11434/api/generate"
DEFAULT_MODEL = "qwen2.5:7b"


class OllamaModel:
    """Calls a local ollama server. The only network this project ever touches."""

    def __init__(self, model: str = DEFAULT_MODEL, url: str = DEFAULT_OLLAMA_URL,
                 timeout: float = 120.0) -> None:
        self.model = model
        self.url = url
        self.timeout = timeout

    def generate(self, prompt: str) -> str:
        payload = json.dumps({
            "model": self.model,
            "prompt": prompt,
            "stream": False,
            "format": "json",
            "options": {"temperature": 0},
        }).encode()
        req = urllib.request.Request(
            self.url, data=payload, headers={"Content-Type": "application/json"}
        )
        try:
            with urllib.request.urlopen(req, timeout=self.timeout) as resp:
                body = json.loads(resp.read().decode())
        except urllib.error.URLError as e:  # server down / unreachable
            raise ModelUnavailable(f"ollama unreachable at {self.url}: {e}") from e
        return body.get("response", "")

    def label(self) -> str:
        return f"ollama:{self.model}"


class CassetteModel:
    """Replays a recorded sequence of raw model responses, one per attempt."""

    def __init__(self, responses: list[str], name: str = "cassette") -> None:
        if not responses:
            raise ValueError("cassette must have at least one response")
        self._responses = responses
        self._i = 0
        self._name = name

    def generate(self, prompt: str) -> str:  # noqa: ARG002 - prompt ignored by design
        resp = self._responses[min(self._i, len(self._responses) - 1)]
        self._i += 1
        return resp

    def label(self) -> str:
        return f"cassette:{self._name}"

    @classmethod
    def from_file(cls, path: str) -> "CassetteModel":
        with open(path) as f:
            data = json.load(f)
        return cls(data["responses"], name=data.get("name", path))


class ModelUnavailable(RuntimeError):
    """Raised when a live model source cannot be reached."""
