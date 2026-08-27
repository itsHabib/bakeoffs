export const WALL_MS = 20_000;

export const RULES = Object.freeze({
  fillers: ["um", "like", "sort of", "basically"],
  hedges: ["i think", "maybe", "hopefully"],
  imperatives: [
    "approve", "book", "buy", "call", "choose", "email", "give", "introduce",
    "invest", "join", "schedule", "send", "sign", "start", "try", "visit",
  ],
  pace: { min: 1.5, max: 3.5 },
  weights: {
    wall: 18,
    fillerEach: 4,
    fillerMax: 20,
    pace: 12,
    hedgeEach: 4,
    hedgeMax: 16,
    ask: 20,
    repeatMax: 16,
  },
});

export function words(text) {
  return (text.toLowerCase().match(/[a-z0-9]+(?:['’-][a-z0-9]+)*/g) ?? [])
    .map((word) => word.replace("’", "'"));
}

export function countPhrases(text, phrases) {
  const haystack = words(text);
  return phrases.reduce((total, phrase) => {
    const needle = words(phrase);
    let count = 0;
    for (let index = 0; index <= haystack.length - needle.length; index += 1) {
      if (needle.every((word, offset) => haystack[index + offset] === word)) count += 1;
    }
    return total + count;
  }, 0);
}

export function applyWall(transcript, durationMs, transcriptIsSurvived = false, cutFragment = "") {
  const clean = transcript.trim().replace(/\s+/g, " ");
  if (durationMs < WALL_MS) {
    return { survived: clean, eaten: "Nothing — you stopped before the wall.", wallHit: false };
  }

  if (transcriptIsSurvived || durationMs === WALL_MS) {
    return {
      survived: clean,
      eaten: cutFragment.trim()
        ? `${cutFragment.trim().replace(/[.!?]+$/, "")}—`
        : "Audio after 20.0s — the browser stopped recognition before another word finalized.",
      wallHit: true,
    };
  }

  let cut = Math.floor(clean.length * (WALL_MS / durationMs));
  cut = Math.max(1, Math.min(clean.length - 1, cut));

  // Simulated timing deliberately lands inside a word, matching the product's hard stop.
  if (/\s/.test(clean[cut])) {
    while (cut < clean.length - 1 && /\s/.test(clean[cut])) cut += 1;
    const end = clean.indexOf(" ", cut);
    const wordEnd = end === -1 ? clean.length : end;
    cut += Math.max(1, Math.floor((wordEnd - cut) / 2));
  }
  if (/\s/.test(clean[cut - 1])) cut += 1;

  return {
    survived: `${clean.slice(0, cut).trimEnd()}—`,
    eaten: clean.slice(cut),
    wallHit: true,
  };
}

export function detectAsk(text) {
  const tokens = words(text);
  if (tokens.length === 0) return { present: false, evidence: "No final quarter." };
  const finalQuarter = tokens.slice(Math.floor(tokens.length * 0.75));
  // TODO: This deliberately blunt heuristic treats a listed verb in the final quarter as
  // imperative. It is inspectable and repeatable, but it can produce false positives.
  const evidence = finalQuarter.find((word) => /^\d/.test(word) || RULES.imperatives.includes(word));
  return {
    present: Boolean(evidence),
    evidence: evidence ? `“${evidence}” in the final quarter` : "No imperative or number in the final quarter.",
  };
}

export function ngramOverlap(currentText, previousText, size = 2) {
  const make = (text) => {
    const tokens = words(text);
    const grams = [];
    for (let index = 0; index <= tokens.length - size; index += 1) {
      grams.push(tokens.slice(index, index + size).join(" "));
    }
    return grams;
  };
  const current = make(currentText);
  if (current.length === 0 || !previousText) return 0;
  const previous = new Set(make(previousText));
  return current.filter((gram) => previous.has(gram)).length / current.length;
}

export function diffCuts(previousText, currentText) {
  const before = words(previousText);
  const after = words(currentText);
  const rows = Array.from({ length: before.length + 1 }, () => Array(after.length + 1).fill(0));
  for (let left = before.length - 1; left >= 0; left -= 1) {
    for (let right = after.length - 1; right >= 0; right -= 1) {
      rows[left][right] = before[left] === after[right]
        ? rows[left + 1][right + 1] + 1
        : Math.max(rows[left + 1][right], rows[left][right + 1]);
    }
  }

  const removed = [];
  let left = 0;
  let right = 0;
  while (left < before.length) {
    if (right < after.length && before[left] === after[right]) {
      left += 1;
      right += 1;
    } else if (right < after.length && rows[left][right + 1] > rows[left + 1][right]) {
      right += 1;
    } else {
      removed.push(before[left]);
      left += 1;
    }
  }
  return removed;
}

export function gradeRep({ transcript, durationMs, previousTranscript = "", transcriptIsSurvived = false, cutFragment = "" }) {
  if (typeof transcript !== "string") throw new TypeError("transcript must be a string");
  if (!Number.isFinite(durationMs) || durationMs <= 0) throw new TypeError("durationMs must be positive");

  const wall = applyWall(transcript, durationMs, transcriptIsSurvived, cutFragment);
  const tokenCount = words(wall.survived).length;
  const measuredSeconds = Math.min(durationMs, WALL_MS) / 1000;
  const wps = tokenCount / measuredSeconds;
  const fillerCount = countPhrases(wall.survived, RULES.fillers);
  const hedgeCount = countPhrases(wall.survived, RULES.hedges);
  const ask = detectAsk(wall.survived);
  const overlap = ngramOverlap(wall.survived, previousTranscript);
  const paceOk = wps >= RULES.pace.min && wps <= RULES.pace.max;

  const penalties = {
    wall: wall.wallHit ? RULES.weights.wall : 0,
    fillers: Math.min(RULES.weights.fillerMax, fillerCount * RULES.weights.fillerEach),
    pace: paceOk ? 0 : RULES.weights.pace,
    hedges: Math.min(RULES.weights.hedgeMax, hedgeCount * RULES.weights.hedgeEach),
    ask: ask.present ? 0 : RULES.weights.ask,
    repeat: previousTranscript && overlap >= 0.25 ? RULES.weights.repeatMax : 0,
  };
  const score = Math.max(0, 100 - Object.values(penalties).reduce((sum, value) => sum + value, 0));

  return {
    score,
    durationMs,
    measuredSeconds,
    survived: wall.survived,
    eaten: wall.eaten,
    wallHit: wall.wallHit,
    tokenCount,
    fillerCount,
    hedgeCount,
    wps,
    paceOk,
    ask,
    overlap,
    penalties,
    cutFromPrevious: previousTranscript ? diffCuts(previousTranscript, wall.survived) : [],
  };
}
