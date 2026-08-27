// Deterministic readback grader. There is no model anywhere in this file —
// every verdict comes from token tables and regexes, so grading is instant,
// free, and identical on every run. This module is shared verbatim by the
// browser UI and by `node --test`.

export const PHONETIC = {
  alpha: 'a', bravo: 'b', charlie: 'c', delta: 'd', echo: 'e', foxtrot: 'f',
  golf: 'g', hotel: 'h', india: 'i', juliet: 'j', juliett: 'j', kilo: 'k',
  lima: 'l', mike: 'm', november: 'n', oscar: 'o', papa: 'p', quebec: 'q',
  romeo: 'r', sierra: 's', tango: 't', uniform: 'u', victor: 'v',
  whiskey: 'w', xray: 'x', yankee: 'y', zulu: 'z',
};

export const NUMBER_WORDS = {
  zero: '0', one: '1', two: '2', three: '3', four: '4', five: '5',
  six: '6', seven: '7', eight: '8', nine: '9', niner: '9',
};

const ACK_WORDS = new Set(['roger', 'wilco', 'copy', 'that', 'affirm',
  'affirmative', 'ok', 'okay', 'copied', 'got', 'it']);

// "Runway Two Seven!" -> tokens ['runway', '27']. Spoken and typed input
// converge on the same canonical token stream, so accept-tables are written
// once, in canonical form.
export function normalize(text) {
  const raw = String(text).toLowerCase()
    .replace(/[^a-z0-9. ]+/g, ' ')
    .replace(/\.(?!\d)/g, ' ')     // keep dots only inside numbers (121.9)
    .replace(/(?<!\d)\./g, ' ');
  const mapped = raw.split(/\s+/).filter(Boolean).map(t => {
    if (NUMBER_WORDS[t] !== undefined) return NUMBER_WORDS[t];
    if (PHONETIC[t] !== undefined) return PHONETIC[t];
    if (t === 'point' || t === 'decimal') return '.';
    return t;
  });

  // Merge spelled-out runs: '1 2 3 a b' -> '123ab', '1 2 1 . 9' -> '121.9'.
  const isAtom = t => /^\d+$/.test(t) || /^[a-z]$/.test(t);
  const out = [];
  let i = 0;
  while (i < mapped.length) {
    const t = mapped[i];
    if (t === '.') {
      const prev = out[out.length - 1];
      const next = mapped[i + 1];
      if (prev && /\d$/.test(prev) && next && /^\d/.test(next)) {
        out[out.length - 1] = prev + '.' + next;
        i += 2;
        continue;
      }
      i += 1; // stray dot
      continue;
    }
    if (isAtom(t) && isAtom(mapped[i + 1] ?? '')) {
      let run = t;
      let j = i + 1;
      while (j < mapped.length && isAtom(mapped[j])) run += mapped[j++];
      out.push(run);
      i = j;
      continue;
    }
    out.push(t);
    i += 1;
  }
  return out;
}

const pad = s => ` ${s} `;

function containsPhrase(joined, phrase) {
  return pad(joined).includes(pad(phrase));
}

// spec.elements: [{ key, label, accept: [phrases], wrong?: [{re, expected, label}], optional? }]
// Verdict per element: 'ok' | 'wrong' | 'missing'. Wrong beats missing:
// reading back the WRONG runway is a different (worse) failure than silence.
export function grade(responseText, spec) {
  const tokens = normalize(responseText);
  const joined = tokens.join(' ');

  const elements = spec.elements.map(el => {
    for (const w of el.wrong ?? []) {
      const m = pad(joined).match(w.re);
      if (m && m[1] !== w.expected) {
        return { ...pick(el), status: 'wrong', heard: (w.label ?? '') + m[1] };
      }
    }
    if (el.accept.some(p => containsPhrase(joined, p))) {
      return { ...pick(el), status: 'ok' };
    }
    return { ...pick(el), status: 'missing' };
  });

  // 'Roger' is an acknowledgement, not a readback. If everything left after
  // stripping the callsign is ack filler, flag it — clearances and hold-short
  // instructions must be read back element by element.
  let rest = pad(joined);
  const callsignEl = spec.elements.find(e => e.key === 'callsign');
  for (const p of callsignEl?.accept ?? []) rest = rest.replaceAll(pad(p), ' ');
  const restTokens = rest.split(/\s+/).filter(Boolean);
  const ackOnly = restTokens.length > 0 && restTokens.every(t => ACK_WORDS.has(t));

  const required = elements.filter(e => !e.optional);
  const okCount = required.filter(e => e.status === 'ok').length;
  return {
    elements,
    ackOnly,
    joined,
    score: required.length ? okCount / required.length : 1,
    wrongCount: required.filter(e => e.status === 'wrong').length,
  };
}

function pick(el) {
  return { key: el.key, label: el.label, optional: !!el.optional };
}
