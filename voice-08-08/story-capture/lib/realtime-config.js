export const REALTIME_MODEL = "gpt-realtime-2.1";

const BASE_INSTRUCTIONS = `You are a warm family-story interviewer helping someone remember how they met their spouse. Keep your voice gentle, curious, and unhurried. Ask exactly ONE brief question at a time. Never summarize, grade, invent, or correct the teller. The application computes coverage in code and will tell you the single story beat to pursue. Follow that beat. If useful, refer to one concrete detail the teller just shared. Do not ask compound questions.`;

export function createRealtimeSessionPayload() {
  return {
    type: "realtime",
    model: REALTIME_MODEL,
    instructions: BASE_INSTRUCTIONS,
    output_modalities: ["audio"],
    audio: {
      input: {
        transcription: { model: "gpt-4o-mini-transcribe" },
        turn_detection: {
          type: "semantic_vad",
          create_response: false,
          interrupt_response: true,
        },
      },
      output: {
        voice: "marin",
        speed: 0.96,
      },
    },
  };
}

export function interviewerInstruction(beat, lastAnswer = "") {
  const context = lastAnswer ? ` The teller just said: ${JSON.stringify(lastAnswer)}.` : "";
  if (!beat) {
    return `All required story beats are covered.${context} Thank the teller warmly in one sentence and say their chapter is ready. Do not ask another question.`;
  }
  return `The deterministic tracker says the next uncovered beat is ${JSON.stringify(beat.label)}. Its required angle is: ${beat.prompt}${context} Ask exactly one short follow-up that pursues only that beat.`;
}
