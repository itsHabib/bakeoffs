# Resume Parley

**The worker dies after a delivery commits. Its replacement either obtains the
retained receipt or names the unresolved operation.** Parley audits the resulting
worker/checker/recorder conversation and identifies the first ordering deviation.

```sh
cd resume-09-05/parley
make demo
```

No installation, network, credentials, models, live control-plane state, or service
setup. Requires Python 3, GHC, and make already on the host. Verified here with
Python 3.14.6 and GHC 9.14.1 on macOS. The demo runs hands-free in several seconds;
[DEMO.md](DEMO.md) supplies the exact sixty-second spoken walkthrough. `make test`
shows every assertion result, upstream compiler test, and source count.

## What ran

The entry has 20 computed cases and 27 unchanged upstream compiler tests. The
controller launches a real worker subprocess, waits for a JSON handshake, sends
SIGKILL to that PID only, waits for exit, and launches a replacement. There are
no timing sleeps. Normal subprocess calls and test handshakes have bounded waits.
Only an explicitly paused worker waits for a controller instruction.

S1 squares `[2,3,5]`; S2 squares `[7,11]`. The source documents and their exact
digests are retained. The same byte-reading checker validates each candidate and
rejects a candidate with its first square incremented by one. Positive evidence
binds the source digest, exact candidate bytes, and checker version. Completion
names the run, subject, candidate, evidence, incarnation, and full sink receipt.

| Boundary or control | Computed result |
| --- | --- |
| Good run, either sink | One effect and valid completion |
| Intent durable, worker killed | Queryable continues; opaque parks, zero effects |
| Commit durable, reply withheld, worker killed | Queryable retrieves original receipt, one effect; opaque parks with one effect |
| Receipt durable, worker killed | Both modes reuse it and complete with one effect |
| Late terminal from incarnation 1 after replacement | Reason and original message retained; no completion until valid replacement |
| Switch S1 to S2 | Same-incarnation S1 terminal refused; fresh S2 checks and distinct effect complete; S1 effect unchanged |
| Change checked candidate bytes | Refused before delivery |
| Incomplete final journal record after lost reply | Prefix retained; queryable reconciles, opaque parks |
| Corrupt complete record or definition mismatch | Refused before journal mutation or effect |
| Repeat sink operation key and then conflict | Same receipt on retry; different payload refused; original effect retained |
| Plant a checked/negative event order swap | Independent effect/content assertions pass; Parley reports event 2 as the deviation |
| Blind retry, old terminal, stale-byte evidence mutants | All produce false completion; independent assertions catch each |

The three mutant traces are **protocol-complete**. This is a deliberate witness
that correct interaction order says nothing about effect multiplicity, current
incarnation, or evidence freshness. The ordering-only control shows a useful
incremental diagnostic: it detects an interaction reorder absent from the
independently expressed effect/identity assertions. It does not demonstrate
that Parley is necessary, nor that it prevented any effects.

## Execution boundaries

- `src/common.py` owns the journal, exact identities, checker, and binding checks.
  Every complete JSONL record has a sequence, event ID, previous-record hash, and
  checksum. Recovery validates the entire complete prefix and definition before
  truncating an incomplete suffix; the discarded suffix is saved. The definition
  digest covers the protocol bytes and explicit workflow/checker/schema versions.
  It does not automatically hash every implementation source edit.
- `src/worker.py` is an ordinary sequential executor. A local advisory file lock
  serializes workers and message acceptance. Every replacement durably increments
  the incarnation; old terminal messages cannot advance the run. All three roles
  are logical boundaries in this deterministic script, not separate agent services.
- `src/sink.py` is a separate process that owns the effect database. A transaction
  commits the payload and receipt together. Queryable mode serializes writes,
  deduplicates exact operation key/payload pairs, refuses conflicts, and exposes
  receipt lookup. Opaque mode appends on every call and explicitly refuses queries.
  The worker never reads the private database; only the test controller inspects
  it to assert actual effect count and identity independently of success labels.
- `src/audit.py` converts accepted journal facts into a Parley trace, grouped by
  source digest within a run. Each normalized event has a sidecar containing the
  **entire original record**, including ID, sequence, hash, subject, and incarnation.
  Non-protocol facts have explicit exclusion reasons; unknown kinds refuse
  normalization. The first deviating prefix points back to the precise raw event.

The stable operation key includes logical run, delivery step, source, candidate,
evidence, and payload content. Incarnation is deliberately absent. Restart cannot
create a new logical effect key; changing S1 to S2 creates a different operation.
An opaque intent with no receipt stays unresolved even if the controller knows
the worker died before dispatch: the replacement has no supported evidence of
that distinction. It never reads the controller's observations or retries it.

## Parley provenance and scope

Pinned Parley source at
`ef7cce0b7d5f197c0dfb1bd629724c3bfd98850a`. The 11 vendored files were compared
byte-for-byte with `git show` at that commit. [SOURCE-LOCK.json](SOURCE-LOCK.json)
records every SHA-256; `python3 tools.py` verifies the files on each build.
The upstream MIT license is retained at `vendor/parley/LICENSE`. There are no
changes to reused source or the canonical repository. The compiler's existing
CLI requires all six included Haskell modules; the two protocol fixtures support
its unchanged tests.

**Executed:** Haskell parser, compiler/projection, observer, and upstream tests;
Python persistence, checker, sink, recovery, normalization, controller, and mutants.
The entry protocol compiles to three local contracts in `artifacts/contracts/`.
Those contracts document role obligations; the Python executor does not interpret
them. The observer replays the global protocol after execution.

**Read but not executed:** Gleam pure enforcement core. **Not executed or reused:**
Gleam bus/runtime and Lean proofs. No general compiler/runtime correctness theorem,
proof for this new protocol, or enforcement guarantee is claimed.

Maintenance counts include the complete ordinary executor and normalization
adapter, not just the nine-line protocol. `python3 tools.py` reports physical
lines including blanks/comments for each boundary, all reused Haskell, tests,
fixtures, and build files. Documentation and license text are excluded.

## Artifacts and replay

Each run makes and retains a fresh system temporary directory. Its path is printed
and saved in `artifacts/latest-run.json`. `make demo` also retains the complete
test/build log in `artifacts/demo-test.log`. Temporary output is local and may be
removed by normal OS cleanup; no durable archival service is implied.

Every case has `expected.json`, `actual.json`, `commands.txt`, worker stdout and
results, input identities in `journal.jsonl`, candidate/negative bytes, sink
database (when invoked), independent `controller-effects.json`, and normalized
trace/identity maps under `audit/`. Corrupt-journal cases refuse before audit.
`ordering-deviation/original-journal.jsonl` preserves the legal history before
the deliberately reordered synthetic history is produced with a fresh valid
checksum chain. The sink repeat/conflict requests are retained as JSON fixtures.

```sh
make test                         # full reproduction, fresh temporary directory
python3 tools.py                   # provenance and maintenance counts
python3 tests/run.py               # rerun after the compiler has been built
python3 src/audit.py CASE_PATH     # inspect a retained case's protocol trace
```

`commands.txt` captures each worker invocation; its SIGKILL comments name the
handshake the controller waits for. `python3 tests/run.py` reproduces those
interruptions automatically. Re-running a worker command against retained state
continues that state and can add journal facts; use the full controller for a
fresh reproduction. Opaque reconciliation remains parked on repeated restarts.

## Limits

This is one-host **worker-process crash** behavior. Python writes flush and fsync;
atomic candidate replacement precedes journal references; SQLite uses a FULL
synchronous local transaction before its commit handshake. The directory entry
itself is not fsynced. Arbitrary power loss, disk loss, sink crashes, multi-host
concurrency, distributed consensus, malicious log rewriting, authentication,
revocation of historical effects, and external API behavior are outside this
claim. Hash chains detect tested corruption, not a writer that recomputes them.
The local lock assumes cooperating processes. There is no adoption recommendation,
external demand evidence, score, or publication.
