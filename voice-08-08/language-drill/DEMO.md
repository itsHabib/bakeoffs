# Mesa 40 — judge-run live voice demo

## Buyer claim

Adult Spanish learners preparing for travel or a move. Education category.
The first paid version would charge **$7.99/month per reviewed scenario pack**.

## Live exchange (~60 seconds)

1. Start with `go run -buildvcs=false .`, open <http://127.0.0.1:8740>, select
   **Connect OpenAI voice**, allow the microphone, and wait for **OpenAI voice
   live** plus the spoken welcome.
2. Turn on **Waiter scene**. Say to the waiter: “Buenas tardes. Una mesa para
   dos, por favor.” Let the waiter answer naturally.
3. Interrupt its next sentence with: “Perdón, ¿podemos sentarnos afuera?” The
   interruption is intentional: judge the turn-taking, not just the voice.
4. Select **Hear it**, then select **Speak Spanish** and repeat the modeled
   phrase aloud. The Realtime partner hears the judge over the WebRTC microphone
   while browser recognition places its transcript in the attempt field for the
   deterministic scorer.
5. On the opening `buenas tardes` phrase, deliberately use **Speak Spanish** to
   say “buenos días,” then submit exactly what the browser heard. Point to the
   red words and **time-of-day mix-up** trap: the model is the ears and mouth,
   but ordinary Go code owns the grade.
6. Correct it aloud: “Buenas tardes.” Submit the recognized text and end on the
   readable green diff.
7. Close: “This is for adult travelers in Education: $7.99 a month per reviewed
   scenario pack. The live waiter creates speaking pressure; readable rules,
   never the model, decide the rep.”

## Canned no-key fallback

Restart without `OPENAI_API_KEY`, select **Run 60-second demo**, and watch eight
fixtures traverse the same `/api/score` endpoint as live and typed attempts.
Call out the clean pass, wrong-gender article, dropped word, English carry-over,
accepted declared variant, history, and total miss. This is the deterministic
test harness and judging-day fallback; it is not the primary voice demo.
