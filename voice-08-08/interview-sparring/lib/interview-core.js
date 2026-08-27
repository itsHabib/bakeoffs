const STOP_WORDS = new Set([
  "a", "about", "an", "and", "are", "as", "at", "be", "by", "did", "do", "for", "from",
  "had", "have", "how", "i", "in", "is", "it", "me", "of", "on", "or", "that",
  "the", "this", "to", "was", "were", "what", "when", "where", "with", "you", "your",
]);

const FILLERS = ["um", "uh", "erm", "like", "you know", "sort of", "kind of", "basically"];

export const QUESTION_BANK = Object.freeze([
  question(
    "leadership",
    "Tell me about a time you led a project through uncertainty.",
    ["led", "lead", "project", "uncertainty", "uncertain", "decision"],
    {
      action: "What did you personally decide or do—not the team?",
      result: "What changed because of your leadership?",
      context: "What made the situation uncertain, and what were you responsible for?",
      dodge: "Give me one specific project you led. What happened?",
    },
  ),
  question(
    "conflict",
    "Tell me about a conflict with a coworker and how you resolved it.",
    ["conflict", "coworker", "colleague", "disagreement", "resolved", "resolve"],
    {
      action: "What did you personally say or do to resolve the conflict?",
      result: "How did the working relationship or outcome change?",
      context: "What was the actual disagreement and what was at stake?",
      dodge: "Name one real disagreement with a coworker. What happened next?",
    },
  ),
  question(
    "failure",
    "Tell me about a professional failure and what you learned from it.",
    ["failure", "failed", "mistake", "learned", "lesson", "professional"],
    {
      action: "What did you do once you realized it was going wrong?",
      result: "What measurable change did you make afterward?",
      context: "What specifically failed, and what was your responsibility?",
      dodge: "Pick one real failure, not a strength in disguise. What happened?",
    },
  ),
  question(
    "ambiguity",
    "Describe a time you had to make progress with ambiguous requirements.",
    ["ambiguous", "ambiguity", "requirements", "unclear", "progress", "scope"],
    {
      action: "Which uncertainty did you resolve first, and how?",
      result: "What shipped or improved because of that choice?",
      context: "What was unclear, and what outcome were you accountable for?",
      dodge: "Give me one specific case where the requirements were unclear.",
    },
  ),
  question(
    "influence",
    "Tell me about a time you influenced a decision without formal authority.",
    ["influenced", "influence", "decision", "authority", "persuaded", "stakeholder"],
    {
      action: "How did you personally change the decision-maker's mind?",
      result: "What decision was made, and what happened afterward?",
      context: "Who held the authority, and why did the decision matter?",
      dodge: "Name one decision you influenced without owning it.",
    },
  ),
  question(
    "feedback",
    "Tell me about difficult feedback you received and how you responded.",
    ["feedback", "difficult", "received", "responded", "criticism", "changed"],
    {
      action: "What behavior did you personally change after hearing it?",
      result: "How did you know the change worked?",
      context: "What was the feedback, and why was it hard to hear?",
      dodge: "Quote the difficult feedback as plainly as you can.",
    },
  ),
  question(
    "priorities",
    "Describe a time you had to choose between competing priorities.",
    ["choose", "competing", "priorities", "priority", "tradeoff", "deadline"],
    {
      action: "What tradeoff did you personally make, and why?",
      result: "What was the consequence of that prioritization?",
      context: "Which priorities were in conflict and who depended on them?",
      dodge: "Give me one concrete priority tradeoff you owned.",
    },
  ),
  question(
    "accomplishment",
    "What professional accomplishment are you most proud of, and why?",
    ["accomplishment", "proud", "professional", "achieved", "impact", "built"],
    {
      action: "Which part of that accomplishment was uniquely yours?",
      result: "What evidence shows that it mattered?",
      context: "What problem made this accomplishment worth pursuing?",
      dodge: "Choose one accomplishment and walk me through it.",
    },
  ),
  question(
    "deadline",
    "Tell me about a time you missed a deadline or had to recover one.",
    ["missed", "deadline", "late", "recover", "schedule", "delivery"],
    {
      action: "What did you personally do when the deadline was at risk?",
      result: "When did you finally deliver, and what changed afterward?",
      context: "Why was the deadline at risk, and what did you own?",
      dodge: "Name a specific deadline that slipped or nearly slipped.",
    },
  ),
]);

function question(id, prompt, keywords, followUps, budgetSeconds = 45) {
  return Object.freeze({id, prompt, keywords, followUps: Object.freeze(followUps), budgetSeconds});
}

export function tokenize(text) {
  return String(text || "")
    .toLowerCase()
    .replaceAll(/[^a-z0-9' ]/g, " ")
    .split(/\s+/)
    .filter(Boolean);
}

function contentTokens(text) {
  return tokenize(text).filter((token) => token.length > 2 && !STOP_WORDS.has(token));
}

function hasAny(text, patterns) {
  return patterns.some((pattern) => pattern.test(text));
}

export function analyzeAnswer(question, transcript, durationSeconds) {
  const cleanTranscript = String(transcript || "").trim();
  const lower = cleanTranscript.toLowerCase();
  const tokens = tokenize(cleanTranscript);
  const words = tokens.length;
  const expectedWords = Math.round(question.budgetSeconds * (140 / 60));
  const duration = Math.max(0, Number(durationSeconds) || 0);
  const fillerCount = FILLERS.reduce((total, filler) => {
    const escaped = filler.replaceAll(" ", "\\s+");
    return total + (lower.match(new RegExp(`\\b${escaped}\\b`, "g")) || []).length;
  }, 0);

  // TODO: These intentionally readable phrase rules are a baseline; validate them
  // against anonymized interview transcripts before treating STAR as coaching truth.
  const star = {
    situation: hasAny(lower, [
      /\b(the )?situation\b/, /\bcontext\b/, /\bwhen (i|we)\b/, /\bat the time\b/,
      /\bthe (project|team|client)\b/,
    ]),
    task: hasAny(lower, [
      /\b(my|the) (task|goal|responsibility|objective)\b/, /\bi (needed|had) to\b/,
      /\bi was responsible\b/, /\bwe needed to\b/,
    ]),
    action: hasAny(lower, [
      /\bmy action\b/, /\bi (personally )?(decided|created|built|led|changed|proposed|met|asked|wrote|analyzed|organized|negotiated|escalated|implemented|owned)\b/,
      /\bi (then |personally )?(did|made|worked|spoke|scheduled|prioritized|reduced|resolved|tested|launched)\b/,
    ]),
    result: hasAny(lower, [
      /\b(the )?result\b/, /\bas a result\b/, /\b(the|this) outcome\b/, /\bwe (increased|decreased|reduced|grew|saved|delivered|shipped)\b/,
      /\b(improved|increased|decreased|reduced|grew|saved) by \d+/, /\b\d+%\b/, /\bultimately\b/,
    ]),
  };

  const questionTerms = new Set([
    ...contentTokens(question.prompt),
    ...question.keywords.map((word) => word.toLowerCase()),
  ]);
  const answerTerms = new Set(contentTokens(cleanTranscript));
  const overlapTerms = [...questionTerms].filter((term) => answerTerms.has(term));
  const overlapRatio = questionTerms.size ? overlapTerms.length / questionTerms.size : 0;
  const dodge = words >= 8 && overlapTerms.length === 0;
  const starCount = Object.values(star).filter(Boolean).length;
  const fillerDensity = words ? fillerCount / words : 0;
  const withinBudget = duration <= question.budgetSeconds;
  const budgetRatio = question.budgetSeconds ? duration / question.budgetSeconds : 0;
  const score = clamp(
    (starCount * 15) +
      (dodge ? 0 : 20) +
      (withinBudget ? 15 : budgetRatio <= 1.25 ? 8 : 0) +
      (fillerDensity <= 0.04 ? 5 : fillerDensity <= 0.08 ? 3 : 0),
    0,
    100,
  );

  return Object.freeze({
    words,
    expectedWords,
    durationSeconds: round(duration, 1),
    budgetSeconds: question.budgetSeconds,
    budgetRatio: round(budgetRatio, 2),
    withinBudget,
    fillerCount,
    fillerDensity: round(fillerDensity, 3),
    star: Object.freeze(star),
    starCount,
    overlapTerms: Object.freeze(overlapTerms),
    overlapRatio: round(overlapRatio, 3),
    dodge,
    score,
  });
}

function clamp(value, low, high) {
  return Math.min(high, Math.max(low, value));
}

function round(value, digits) {
  const factor = 10 ** digits;
  return Math.round(value * factor) / factor;
}

function chooseFollowUp(question, metrics, usedReasons, wasInterrupted) {
  if (wasInterrupted && !usedReasons.has("time")) {
    return {
      reason: "time",
      label: "TIME BUDGET",
      prompt: `I'm going to stop you there. In one sentence: ${question.followUps.action}`,
    };
  }
  if (metrics.dodge && !usedReasons.has("dodge")) {
    return {reason: "dodge", label: "QUESTION DODGE", prompt: question.followUps.dodge};
  }
  if (!metrics.star.action && !usedReasons.has("action")) {
    return {reason: "action", label: "MISSING ACTION", prompt: question.followUps.action};
  }
  if (!metrics.star.result && !usedReasons.has("result")) {
    return {reason: "result", label: "MISSING RESULT", prompt: question.followUps.result};
  }
  if ((!metrics.star.situation || !metrics.star.task) && !usedReasons.has("context")) {
    return {reason: "context", label: "MISSING CONTEXT", prompt: question.followUps.context};
  }
  if (metrics.fillerDensity > 0.08 && !usedReasons.has("clarity")) {
    return {
      reason: "clarity",
      label: "FILLER DENSITY",
      prompt: "Try that once more in two direct sentences, without throat-clearing.",
    };
  }
  return null;
}

export class InterviewMachine {
  constructor({questionIds = ["leadership", "conflict", "failure"]} = {}) {
    const bank = new Map(QUESTION_BANK.map((entry) => [entry.id, entry]));
    if (!Array.isArray(questionIds) || questionIds.length === 0) {
      throw new Error("At least one question is required.");
    }
    this.questions = questionIds.map((id) => {
      const found = bank.get(id);
      if (!found) throw new Error(`Unknown question: ${id}`);
      return found;
    });
    this.reset();
  }

  reset() {
    this.phase = "ready";
    this.index = -1;
    this.currentQuestion = null;
    this.answerStartedAt = null;
    this.records = [];
  }

  start() {
    if (this.phase !== "ready") throw new Error("Interview already started.");
    return this.#advanceQuestion();
  }

  beginAnswer(nowMilliseconds = Date.now()) {
    if (!["question", "followup"].includes(this.phase)) {
      throw new Error(`Cannot begin an answer during ${this.phase}.`);
    }
    this.answerStartedAt = nowMilliseconds;
    this.phase = "answering";
    return {type: "answer_started", questionId: this.currentQuestion.id};
  }

  submitAnswer({transcript, durationSeconds, wasInterrupted = false}, nowMilliseconds = Date.now()) {
    if (this.phase !== "answering") throw new Error(`Cannot submit an answer during ${this.phase}.`);
    const duration = durationSeconds ?? Math.max(0, (nowMilliseconds - this.answerStartedAt) / 1000);
    const record = this.records.at(-1);
    const attempt = {
      transcript: String(transcript || "").trim(),
      durationSeconds: round(duration, 1),
      wasInterrupted: Boolean(wasInterrupted),
      kind: record.attempts.length ? "followup" : "primary",
    };
    record.attempts.push(attempt);
    record.totalDurationSeconds = round(record.attempts.reduce((sum, item) => sum + item.durationSeconds, 0), 1);
    record.transcript = record.attempts.map((item) => item.transcript).filter(Boolean).join(" ");
    record.metrics = analyzeAnswer(this.currentQuestion, record.transcript, record.totalDurationSeconds);
    attempt.metrics = analyzeAnswer(this.currentQuestion, attempt.transcript, attempt.durationSeconds);
    this.answerStartedAt = null;

    const followUp = record.followUps.length < 2
      ? chooseFollowUp(this.currentQuestion, record.metrics, new Set(record.followUps.map((item) => item.reason)), wasInterrupted)
      : null;
    if (followUp) {
      record.followUps.push(followUp);
      this.phase = "followup";
      return {
        type: "followup",
        question: this.currentQuestion,
        followUp,
        attempt,
        metrics: record.metrics,
      };
    }
    if (this.index + 1 < this.questions.length) {
      const completed = {attempt, metrics: record.metrics};
      const next = this.#advanceQuestion();
      return {...next, completed};
    }
    this.phase = "debrief";
    return {type: "debrief", attempt, debrief: this.debrief()};
  }

  debrief() {
    if (this.phase !== "debrief") throw new Error("Debrief is not ready.");
    const scorecards = this.records.map((record) => ({
      questionId: record.question.id,
      prompt: record.question.prompt,
      score: record.metrics.score,
      metrics: record.metrics,
      followUps: record.followUps,
      transcript: record.transcript,
      primaryTranscript: record.attempts[0]?.transcript || "",
      primaryMetrics: record.attempts[0]?.metrics || null,
    }));
    const dodgeCandidates = scorecards
      .filter((card) => card.primaryMetrics?.dodge)
      .sort((a, b) => a.primaryMetrics.overlapRatio - b.primaryMetrics.overlapRatio);
    const worstDodge = dodgeCandidates[0]
      ? {
          question: dodgeCandidates[0].prompt,
          quote: dodgeCandidates[0].primaryTranscript,
          overlapRatio: dodgeCandidates[0].primaryMetrics.overlapRatio,
        }
      : null;
    const tomorrow = [...scorecards]
      .sort((a, b) => a.score - b.score || a.questionId.localeCompare(b.questionId))
      .slice(0, Math.min(2, scorecards.length))
      .map((card) => ({questionId: card.questionId, prompt: card.prompt, score: card.score}));
    return Object.freeze({
      scorecards: Object.freeze(scorecards),
      worstDodge: worstDodge ? Object.freeze(worstDodge) : null,
      tomorrow: Object.freeze(tomorrow),
      averageScore: scorecards.length
        ? Math.round(scorecards.reduce((sum, card) => sum + card.score, 0) / scorecards.length)
        : 0,
    });
  }

  #advanceQuestion() {
    this.index += 1;
    this.currentQuestion = this.questions[this.index];
    this.phase = "question";
    this.records.push({
      question: this.currentQuestion,
      attempts: [],
      followUps: [],
      transcript: "",
      totalDurationSeconds: 0,
      metrics: null,
    });
    return {
      type: "question",
      number: this.index + 1,
      total: this.questions.length,
      question: this.currentQuestion,
    };
  }
}

export const CANNED_FIXTURE = Object.freeze([
  Object.freeze({
    questionId: "leadership",
    transcript: "The situation was an uncertain billing migration with three teams. My task was to establish a safe launch plan. I personally created a rollback checklist, led a daily risk review, and asked each owner to prove their cutover step. As a result, we shipped on schedule with no customer-visible failures and reduced support tickets by 28%.",
    durationSeconds: 31,
  }),
  Object.freeze({
    questionId: "conflict",
    transcript: "The situation was a conflict with a coworker about a launch date, and, um, basically the team had a lot of history and everyone had opinions. My task was to resolve the disagreement, but, you know, we kept circling through meetings and sort of revisiting every old decision. Like, people brought up the roadmap, and then we discussed staffing, and then there were concerns from design, and I wanted everyone to feel heard. We held more meetings, and I explained that collaboration matters, and, um, there were several documents and many possible schedules. The coworker felt strongly about quality, which I also value, and the whole team really worked hard for quite a while while the deadline kept getting closer.",
    durationSeconds: 58,
    wasInterrupted: true,
    followUpAnswers: Object.freeze([
      Object.freeze({
        transcript: "My action was to propose a two-day scope cut and get written agreement from my coworker. The result was an on-time launch, and our working relationship improved.",
        durationSeconds: 14,
      }),
    ]),
  }),
  Object.freeze({
    questionId: "failure",
    transcript: "I'm excited about this role because your company has a strong mission, and my background is a great fit. I enjoy collaborative teams and fast-moving products.",
    durationSeconds: 17,
    followUpAnswers: Object.freeze([
      Object.freeze({
        transcript: "The situation was a release I owned that failed after I skipped a load test. My task was recovery. I personally rolled it back, added the missing test, and told the client what happened. The result was service restored in twelve minutes, and I learned to make load evidence a release requirement.",
        durationSeconds: 29,
      }),
    ]),
  }),
]);

export function runCannedReplay() {
  const machine = new InterviewMachine({questionIds: CANNED_FIXTURE.map((item) => item.questionId)});
  const events = [];
  let decision = machine.start();
  events.push(decision);
  for (const fixture of CANNED_FIXTURE) {
    if (decision.question?.id !== fixture.questionId) {
      throw new Error(`Fixture order mismatch at ${fixture.questionId}.`);
    }
    machine.beginAnswer(0);
    events.push({type: "answer", question: decision.question, transcript: fixture.transcript, durationSeconds: fixture.durationSeconds, wasInterrupted: Boolean(fixture.wasInterrupted)});
    decision = machine.submitAnswer(fixture, fixture.durationSeconds * 1000);
    events.push(decision);
    for (const answer of fixture.followUpAnswers || []) {
      if (decision.type !== "followup") break;
      machine.beginAnswer(0);
      events.push({type: "answer", question: decision.question, transcript: answer.transcript, durationSeconds: answer.durationSeconds, wasInterrupted: false, followup: true});
      decision = machine.submitAnswer(answer, answer.durationSeconds * 1000);
      events.push(decision);
    }
  }
  if (decision.type !== "debrief") throw new Error("Canned fixture did not reach debrief.");
  return Object.freeze({events: Object.freeze(events), debrief: decision.debrief});
}
