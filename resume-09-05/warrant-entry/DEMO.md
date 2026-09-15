# Exactly 60 seconds

This is a timed presentation script, not a claim that execution deliberately sleeps for 60 seconds. Start in this repository with a terminal visible. The full hands-free run normally takes a few seconds; use the remaining segment time to point at the output/artifacts. The seven intervals total exactly 60 seconds.

| Time | Action and words |
| --- | --- |
| 00–08 (8s) | Run `./run.sh`. “The worker will die at a tool boundary. Warrant determines what its evidence allows next.” |
| 08–18 (10s) | Point to `lost-ack-queryable`. “The sink committed. The controller withheld the reply and killed the real worker. Its replacement queries the same operation and retains one effect.” |
| 18–27 (9s) | Point to `lost-ack-opaque`. “Here the sink cannot answer. The operation stays named and unresolved. It is neither retried nor called complete.” |
| 27–36 (9s) | Point to `source-change` and `candidate-change`. “S1 support cannot certify S2. Fresh S2 work completes. Changing checked candidate bytes is refused before delivery.” |
| 36–44 (8s) | Point to `late-terminal` and `replay-is-read-only`. “Incarnation fencing is the adapter's job. Reading a journal executes nothing; a continuation does.” |
| 44–54 (10s) | Point to the three mutant rows. “The same checks detect two opaque effects, a stale terminal advancing the journal, and a false byte-level certification. These are measured failures.” |
| 54–60 (6s) | Point to the printed artifact path. “The receipts and crash journals are retained. This is one-host process recovery; the adapter and reused core are both counted.” |

After the clock, inspect `results.json`, any case's `controller-observation.json`, and `lost-ack-queryable/journal-at-crash.ndjson`. `sh <case>/REPLAY.sh` demonstrates read-only reduction. `./run.sh test` runs the entry and upstream tests.
