export const TOPIC = "How you two met";

export const STORY_SPINE = Object.freeze([
  {
    id: "time",
    label: "When",
    prompt: "Find when this happened: a year, age, season, or life period.",
  },
  {
    id: "place",
    label: "Where",
    prompt: "Find the physical place where they first met.",
  },
  {
    id: "people",
    label: "Who",
    prompt: "Identify the people involved and how they were connected.",
  },
  {
    id: "complication",
    label: "The turn",
    prompt: "Find the obstacle, surprise, mistake, or turn that made the meeting a story.",
  },
  {
    id: "feeling",
    label: "The feeling",
    prompt: "Find how the teller felt in that moment.",
  },
  {
    id: "sensory",
    label: "One vivid detail",
    prompt: "Find something seen, heard, smelled, tasted, touched, or physically noticed.",
  },
]);

export const FOLLOW_UPS = Object.freeze({
  time: "About when was this — what year or season do you remember?",
  place: "Where were you standing when you first saw each other?",
  people: "Who was there, and who introduced you?",
  complication: "What nearly kept the two of you from connecting?",
  feeling: "What were you feeling before you knew how it would turn out?",
  sensory: "What is one sound, smell, color, or physical detail you still remember?",
});

const month = "january|february|march|april|may|june|july|august|september|october|november|december";
const season = "spring|summer|autumn|fall|winter";
const placeWords = "airport|apartment|bakery|bar|barn|beach|bus stop|cafe|cafeteria|campus|church|club|dance hall|diner|factory|farm|festival|gym|hall|hospital|hotel|house|kitchen|library|office|park|party|restaurant|school|shop|station|store|street|theater|train station|university|wedding";
const feelingWords = "afraid|angry|awkward|calm|delighted|embarrassed|excited|glad|happy|heartbroken|hopeful|jealous|lonely|nervous|proud|relieved|sad|scared|shy|surprised|thrilled|uneasy|worried";
const sensoryWords = "bright|cold|color|dark|faint|felt rough|glittered|glowing|heard|hot|loud|music|noise|perfume|quiet|red|salty|scent|shining|smell|smelled|sound|sounded|sparkled|sweet|taste|tasted|warm|whistle|white";

const RULES = Object.freeze({
  time: [
    /\b(?:18|19|20)\d{2}\b/i,
    new RegExp(`\\b(?:${month}|${season})\\b`, "i"),
    /\b(?:morning|afternoon|evening|night|wartime|after the war|before the war)\b/i,
    /\b(?:i was|we were)\s+(?:\d{1,2}|sixteen|seventeen|eighteen|nineteen|twenty|twenty-one)\b/i,
  ],
  place: [
    new RegExp(`\\b(?:at|in|inside|near|outside|by) (?:a |an |the )?(?:${placeWords})\\b`, "i"),
    /\bwhere we met was\b/i,
  ],
  people: [
    /\b(?:aunt|brother|cousin|dad|father|friend|grandfather|grandmother|husband|mother|neighbor|sister|uncle|wife)\b/i,
    /\b(?:introduced|set us up|came with|brought along)\b/i,
    /\b(?:his|her|their) name was\b/i,
  ],
  complication: [
    /\b(?:but|except|however|instead|only trouble|problem|unfortunately)\b/i,
    /\b(?:almost didn't|almost did not|couldn't|could not|missed|mistake|mixed up|wrong)\b/i,
  ],
  feeling: [
    new RegExp(`\\b(?:felt|feeling|was|were) (?:so |very |a little )?(?:${feelingWords})\\b`, "i"),
    new RegExp(`\\b(?:${feelingWords})\\b`, "i"),
    /\bmy (?:heart|hands|knees|stomach) (?:jumped|raced|shook|fluttered|dropped)\b/i,
  ],
  sensory: [
    new RegExp(`\\b(?:${sensoryWords})\\b`, "i"),
    /\b(?:could see|could hear|could smell|could taste|remember the way)\b/i,
  ],
});

export function splitSentences(text) {
  return (String(text).match(/[^.!?]+[.!?]+|[^.!?]+$/g) || [])
    .map((sentence) => sentence.trim())
    .filter(Boolean);
}

export function matchingBeatIDs(text) {
  return STORY_SPINE
    .filter((beat) => RULES[beat.id].some((rule) => rule.test(text)))
    .map((beat) => beat.id);
}

export function trackCoverage(turns) {
  const tellerTurns = turns.filter((turn) => turn.speaker === "teller");
  const evidence = Object.fromEntries(STORY_SPINE.map((beat) => [beat.id, []]));

  for (const turn of tellerTurns) {
    for (const sentence of splitSentences(turn.text)) {
      for (const beatID of matchingBeatIDs(sentence)) {
        evidence[beatID].push(sentence);
      }
    }
  }

  const beats = STORY_SPINE.map((beat) => ({
    ...beat,
    covered: evidence[beat.id].length > 0,
    evidence: evidence[beat.id],
  }));
  const uncovered = beats.filter((beat) => !beat.covered);

  return {
    beats,
    coveredCount: beats.length - uncovered.length,
    totalCount: beats.length,
    uncovered,
    nextBeat: uncovered[0] || null,
    followUps: uncovered.map((beat) => FOLLOW_UPS[beat.id]),
  };
}

export function assembleChapter(turns) {
  const buckets = Object.fromEntries(STORY_SPINE.map((beat) => [beat.id, []]));
  const used = new Set();

  for (const turn of turns.filter((candidate) => candidate.speaker === "teller")) {
    for (const sentence of splitSentences(turn.text)) {
      if (used.has(sentence)) continue;
      const [primaryBeat] = matchingBeatIDs(sentence);
      if (!primaryBeat) continue;
      buckets[primaryBeat].push(sentence);
      used.add(sentence);
    }
  }

  const sections = STORY_SPINE
    .map((beat) => ({ beat, sentences: buckets[beat.id] }))
    .filter((section) => section.sentences.length > 0);
  const sentences = sections.flatMap((section) => section.sentences);

  return {
    title: "The Summer Daniel Noticed Me",
    sections,
    sentences,
    text: sections.map((section) => section.sentences.join(" ")).join("\n\n"),
  };
}

export function verifyProvenance(chapter, turns) {
  const tellerText = turns
    .filter((turn) => turn.speaker === "teller")
    .map((turn) => turn.text)
    .join("\n");
  const missing = chapter.sentences.filter((sentence) => !tellerText.includes(sentence));
  return { valid: missing.length === 0, missing };
}
