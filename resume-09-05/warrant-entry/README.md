# Warrant: resume the operation the evidence actually names

Run the hands-free local demo:

```sh
./run.sh
```

It builds the binary, runs 20 deterministic cases with real worker processes, kills workers at explicit handshakes, starts replacements, and prints a retained `/tmp/warrant-demo-…` artifact directory. All effects are disposable local JSON records owned by a separate sink process. No network service, credentials, installations, or live portfolio state is used. Go 1.26 is required; verified here with Go 1.26.5 on macOS arm64. Dependency downloads and automatic toolchain downloads are disabled.

```sh
./run.sh test                 # entry integration/unit tests and pinned upstream tests
./run.sh demo -dir /tmp/my-fresh-warrant-run
./count.sh                   # reproducible physical and non-comment line counts
```

`-dir` must be empty or absent. The demo retains results; the test suite uses automatically removed temporary directories. `DEMO.md` supplies a 60-second walkthrough. `RUN-RESULTS.md` records the verified local execution and code counts.

## What actually runs

One logical run prepares a candidate for S1, checks its bytes, runs the same checker on a planted defect, delivers, and records completion. S1 contains integer 7, S2 contains 11; the candidate must contain the exact source digest and twice that integer. Source and candidate byte digests form the Warrant subject. The control is a separate, actually invalid candidate with its own subject.

Warrant's JSON-compatible definition declares `prepare → check → control → deliver → complete`. `RequirementsMet` demands both support for the current subject and a refutation of a different subject; completion also requires delivery evidence. `Reduce` owns the next step and the honest replay-class decision. Evidence is recorded only after executing the checker or obtaining a sink receipt.

The adapter re-reads candidate bytes before delivery and completion. A source change creates a fresh source epoch in the same logical run, with fresh preparation and evidence. The outer journal retains all S1 events, and any committed S1 effect remains in the sink. S1 delivery is never renamed into S2 delivery. Changing source with unresolved prior work is refused until that prior source is resolved.

The stable delivery key hashes `[logical run, deliver, subject, candidate digest]`. Incarnation is deliberately absent from that key. Each replacement acquires a local exclusive file lock, creates a fresh incarnation, and records it durably. Incoming old-incarnation terminal messages are rejected with a durable reason before they can become Warrant events.

## The two effect contracts

The queryable sink offers `put(key, payload)` and `query(key)`. It checks payload bytes against the supplied digest, fsyncs the effect record before responding, deduplicates an identical key and input, and refuses a conflicting payload while retaining the original receipt. After a lost acknowledgement, Warrant requests recovery and the adapter queries the exact pending key. A missing receipt allows the same operation to be submitted.

The opaque sink offers only a non-idempotent `put`. Query is explicitly refused and repeated calls perform repeated effects. Warrant declares delivery `manual`. After any durable intent without a durable receipt, its replacement parks with the operation key. That includes an intent-only crash where the controller knows no effect occurred: the replacement has no interface that establishes that fact. The controller alone reads opaque private records to verify effect count and identity.

An acknowledged, journaled opaque receipt can be reused after a worker crash. A receipt that only the sink retained cannot be invented by the worker. This is the distinction the lost-ack demo exhibits.

## The added machinery is visible

| File | Responsibility |
| --- | --- |
| `model.go` | Concrete workflow, byte identities, actual checker, receipts |
| `journal.go` | Durable append/load, strict framing, checksums, definitions, source epochs, event pairing |
| `worker.go` | Continuation, local lock, incarnation fence, dispatch, sink reconciliation |
| `sink.go` | Separate serialized service process, durable effects, deduplication/query |
| `controller.go` | Subprocess handshakes/SIGKILL, fixtures, independent effect and identity oracles |
| `main_test.go` | Real-process suite plus framing/checker/mutant controls |

The original `Runner.Run` always creates a fresh journal, so this entry does not use it. The original `ReadJournal` treats any malformed record as a torn tail, so this entry supplies a stricter loader. Neither `Reduce` nor the original runner implements definition-drift enforcement or incarnation fencing; those checks belong explicitly to this adapter. No upstream protocol fields were repurposed. The outer records wrap ordinary Warrant envelopes.

`replay` strictly loads and reduces without executing, claiming an incarnation, repairing a tail, or writing a journal. `worker` executes a continuation. Every case includes a `REPLAY.sh` that invokes the former and a `commands.ndjson` containing the actual process commands used by its controller. Re-running `./run.sh demo` reconstructs the crash schedule; shell replay of the worker alone does not recreate the controller's withheld acknowledgements or SIGKILL.

## Cases and controls

The suite covers uninterrupted runs in both sink modes; intent crashes in both modes; lost acknowledgements in both modes; receipt-durable crashes in both modes; late terminal rejection plus successful continuation; S2 refusing old evidence then completing with fresh work; changed candidate bytes; torn tails in both modes; complete corruption; a validly checksummed definition mismatch; duplicate/conflicting sink keys; and read-only replay.

Three deliberately broken variants are enabled only by the controller's test environment: blindly retrying an unresolved opaque effect; accepting an old-incarnation terminal; and checking requirements against a stale subject. The same scenario oracles must detect each. Failures during setup do not count as detecting a mutant. The observer checks actual integer arithmetic, actual candidate/source hashes, receipt contents/IDs, evidence-to-byte bindings, journal advancement, and private sink effect counts. A worker's `completed` label alone cannot pass.

Each case retains input bytes/identities, `journal-at-crash.ndjson` where relevant, `run/journal.ndjson`, current/per-worker outcomes, process logs and commands, expected/actual results, and `controller-observation.json` with actual sink receipts. A torn tail is retained as `journal.ndjson.discarded-tail`. Sink-private files are under `controller-private/`; the worker is given only its own directory and a Unix socket, and its code never reads that directory. This is an interface discipline on one account, not an OS security sandbox.

## Source pin and license status

The local dependency is Warrant commit
**`33094f0b99a5a200fdcfebc87c28b39efab5ef8f`**. `verify-source.sh` requires
that exact clean checkout at `../warrant` and verifies `SOURCE.sha256` before
every supported build/run. No upstream files were modified or vendored; the
demo does not fetch them.

There is **no LICENSE, COPYING, or NOTICE file in that source commit**. This entry uses the explicitly pinned local-dependency option and does not invent a license or redistribute Warrant source. No local adaptations were made to Warrant. The five reducer/type files are the relevant reused core; the upstream runner is compiled as part of the package but unused. Both are counted, alongside the entire consumer, sink, controller, and tests. Upstream tests are also counted separately.

## Durability and limits

The scope is one host, one serialized sink, and worker-process crashes. The worker writes one checksummed newline-terminated record, checks the write, and fsyncs before acting on it. A replacement validates every complete record, keeps only an unterminated final tail out of the fold, preserves that tail, truncates it under the exclusive lock, and fsyncs before appending. Complete corruption, sequence/hash breaks, unknown record kinds, and definition mismatch refuse resume. The checksum detects accidental damage; it is not authentication against a writer who can recompute it.

Artifacts use fsync-and-rename. This does not claim arbitrary machine/power-loss durability, directory-fsync guarantees, multi-host consensus, or recovery from sink-process crashes during its own write. The service's durable commit precedes its acknowledgement; killing a worker leaves that service and effect record intact. Waits have 15–20 second controller bounds and the continuation has a 16-step bound. No multi-agent scheduler, generic storage API, or model reasoning is demonstrated.
