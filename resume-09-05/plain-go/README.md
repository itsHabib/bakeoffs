# Plain Go: restart the worker, retain the facts

Run the hands-free local demo:

```sh
./demo.sh
```

Requires Go 1.26 and a POSIX host with process kill support. No dependencies, credentials, service setup, installations, or network access are needed. The script builds one binary and runs 15 normal cases plus three deliberately broken variants. Every run creates its own directory under `artifacts/`; the command prints that directory. The timed presentation is in [DEMO.md](DEMO.md).

```sh
go test ./...
go vet ./...
```

Tests build a disposable binary and execute the same crash controller in temporary directories. The demo retains all journals, original source snapshots, candidates, planted defects, evidence, sink effect records, process logs, late-terminal packets, mutation bytes, expected/actual observations, and exact process invocations. `results.json` summarizes outcomes. `REPLAY.txt` records the replay entry point. Individual commands containing `--boundary` wait for the controller; use `./demo.sh` to replay them automatically in fresh directories.

The workflow builds deterministic candidate JSON from source bytes, runs the same content checker on the valid candidate and a planted defect, writes content-addressed evidence, records delivery intent, calls a sink subprocess, retains its receipt, and accepts a terminal bound to the active incarnation. S1 and S2 use distinct source digests. Completion binds the run, source, candidate, evidence, operation key, receipt, and incarnation. Changing candidate bytes is refused even after an otherwise valid receipt exists.

The operation key hashes `[run, "deliver", source digest, candidate digest, evidence digest]`. Incarnation fences terminal messages but does not change that key. The journal binds its original run ID and sink mode, and every complete record carries a sequence, workflow definition, previous hash, and SHA-256 checksum. An incomplete final line is discarded before appending; any corrupt complete record or definition mismatch is refused. Checksums detect accidental corruption, not malicious rewriting.

The queryable sink supports query and key/payload deduplication. Applying the same pair returns its original receipt; a conflicting payload refuses without altering the retained effect. The opaque sink exposes only apply and creates a new non-idempotent effect for each call. Its private files are accessible only to the sink and controller in the program's call graph; recovery never reads them. An opaque intent without a retained receipt remains `UNRESOLVED` with its operation identity, including when the controller knows the crash occurred before the effect. There is no evidence available to the replacement that distinguishes that case from a lost acknowledgement.

At the lost-ack boundary, the sink has flushed an effect and emits `SINK_COMMITTED_REPLY_WITHHELD`, without sending the receipt. The worker forwards a boundary handshake and blocks. The controller sends SIGKILL to that worker and launches a new executable process against the journal. Closing the dead worker's pipe lets the sink exit. There are no timing sleeps to target a crash. All subprocess waits are bounded; a short cleanup poll waits for sink lock removal after the already-observed boundary.

The oracle reconstructs expected candidate JSON separately from the worker's checker and checks retained bytes, digest bindings, terminal incarnation, sink receipts, and effect counts. Normal success labels are never sufficient. Mutants require their specific observable fault to occur: two opaque effects; an old-incarnation completion; or changed candidate bytes with stale certification. Infrastructure failure cannot count as catching a mutant. The mutants are activated only by explicit test flags and are exercised in separate scenario directories.

| Case | Required observed result |
| --- | --- |
| Good run | Valid S1, negative control rejected, one effect, named completion |
| Intent crash, queryable | Query misses, replacement applies once |
| Intent crash, opaque | Unresolved, zero effects, no completion |
| Lost acknowledgement, queryable | Existing receipt recovered, one effect |
| Receipt crash | Retained result reused, one effect |
| Lost acknowledgement, opaque | Unresolved through repeated restarts, one effect |
| Late terminal | Old incarnation rejected with reason; fresh worker completes |
| Source change | S1 history retained; fresh S2 candidate/evidence/effect/completion |
| Candidate change | Refusal without delivery; restored bytes can complete |
| Truncated tail, both modes | Valid prefix recovered; queryable completes, opaque parks |
| Corrupt complete record | Explicit refusal, zero effects |
| Definition mismatch | Explicit refusal, zero effects |
| Sink repeat/conflict | Same receipt on repeat; conflict refuses; one effect |
| Run binding change | Explicit refusal, zero effects |

Scope is one host and worker-process crashes. The controller serializes workers per journal; concurrent live writers are unsupported. Sink requests use a local exclusion directory and refuse contention. Files are written and `fsync`ed before their relevant handshakes; acknowledged writes are assumed to survive a worker crash. Parent directories are not fsynced, there is no arbitrary power-loss guarantee, and sink-process crashes during a file write are not tested. A killed sink could leave a lock requiring operator inspection. This is neither distributed consensus nor a security/authority boundary, and no incarnation is minted from live Org state.

The production-shaped logic is `workflow.go`, `journal.go`, and `sink.go`; `main.go` is CLI wiring. `controller.go` is the executable scenario harness and independent outcome oracle, with the three small worker mutants intentionally included behind flags. `workflow_test.go` exercises subprocess recovery and the real checker. There is no reusable kernel, scheduler, adapter framework, dependency, or copied machinery. Operational steps: run one command; inspect its artifact directory. No external demand has been validated, and this entry makes no adoption or scoring claim.

Recorded validation (September 5, 2026; Go 1.26.5, darwin/arm64): `go test ./...`, `go vet ./...`, and `./demo.sh` passed. The retained demo is `artifacts/run-708358222/`: 15 normal cases passed, and all three mutants produced their specific independently detected defect. A separate read of private effect records and journal completions confirmed the key recovery, revision, and mutant counts. Maintained source is 1,073 Go lines across six files (564 workflow/journal/sink/CLI, 479 executable controller, 30 tests), plus the six-line shell entry point; this includes comments and blank lines. All machinery is counted; there is no reused package.
