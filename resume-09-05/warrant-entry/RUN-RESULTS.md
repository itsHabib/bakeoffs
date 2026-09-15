# Verified local result — 2026-09-05

`./run.sh` passed **20/20 cases**, including three deliberately defective variants detected by independent oracles. `./run.sh test` passed the entry suite and the pinned Warrant suite. `GOPROXY=off GOTOOLCHAIN=local go vet ./...`, `./verify-source.sh`, and `git diff --check` passed.

The final demo artifacts are retained at `/tmp/warrant-demo-1339524518`, with a copy at `.artifacts/verified-demo/` in this repository. The copied commands preserve the original execution paths. The three mutant observations were:

- Blind retry: **2 effects**, expected 1 unresolved effect.
- Old terminal accepted: **8 → 9 Warrant events** before legitimate continuation.
- Stale candidate: actual candidate arithmetic invalid for S1 despite reported completion.

The other 17 cases cover every required common case, both sink modes where applicable, and read-only replay. Queryable lost-ack recovery finishes with one effect. Opaque lost-ack recovery remains unresolved with one effect; opaque intent-only recovery remains unresolved with zero effects. The controller verifies identities and actual sink records, not only result labels.

## Reproduce

```sh
cd resume-09-05/warrant-entry
./run.sh
./run.sh test
GOPROXY=off GOTOOLCHAIN=local go vet ./...
sh /tmp/warrant-demo-1339524518/lost-ack-queryable/REPLAY.sh
./count.sh
```

The entry integration suite builds a real executable in a test-owned temporary directory and starts real sink/worker subprocesses. It uses explicit durable-intent, reply-withheld, and durable-receipt handshakes, then SIGKILL and a replacement. Additional tests cover malformed middle records, empty complete records, valid-JSON damage, sequence corruption, unknown kinds, torn-tail repair, actual candidate arithmetic, and test-only mutant isolation.

## Code footprint

Counts from `./count.sh`. Physical lines include comments and whitespace. The second metric excludes blank lines and whole-line Go `//` comments; it is a reproducible size metric, not a language parser or complexity score.

| Boundary | Physical lines | Nonblank, non-line-comment lines |
| --- | ---: | ---: |
| New workflow, journal, worker, CLI | 728 | 701 |
| New local sink | 163 | 157 |
| New crash controller and independent oracles | 743 | 729 |
| New test entry points and unit tests | 122 | 116 |
| **All new Go** | **1,756** | **1,703** |
| Reused Warrant reducer/type core | 864 | 588 |
| Warrant runner compiled but unused | 179 | 118 |
| **New Go plus whole imported Warrant package** | **2,799** | **2,409** |
| Upstream tests also executed | 825 | 629 |
| **Including upstream test source** | **3,624** | **3,038** |

Three shell helpers add 41 physical lines. Documentation, `go.mod`, hash manifests, generated artifacts, and upstream JSON fixtures are outside the Go source metric. There are no external downloaded libraries; the local Warrant package is the only dependency. Its exact unmodified commit is `33094f0b99a5a200fdcfebc87c28b39efab5ef8f`, with per-file hashes in `SOURCE.sha256`. No upstream license file exists at that commit, so this uses the pinned local-dependency route and vendors no upstream source.

The durability, fencing, epoch, and sink bookkeeping remain substantial adapter responsibilities. The original runner and permissive reader are not used. Limits remain one-host worker-process crashes, no power-loss or multi-host guarantees, and no sink-write-crash recovery claim. No live Org/Gate/fleet state was accessed, no canonical source changed, and nothing was pushed or published.
