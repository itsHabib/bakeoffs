# Restartable agent work — three independent builds

Prepared as three independent local builds. See [execution results](RESULTS.md).
Operator scoring remains separate; the common requirements below were fixed
before launch.

**The worker died. What can its replacement safely do?**

One bounded workflow, three implementation approaches. This round tests whether a POC earns its integration cost against an honest simple baseline. The deliverable is a local execution demo with real subprocess interruption and inspectable artifacts. It is not a new agent platform.

## Entries

| Entry | Fresh local repo | Bet |
| --- | --- | --- |
| [Plain Go](plain-go/BRIEF.md) | [`plain-go/`](plain-go) | An explicit state machine is enough and simpler in total. |
| [Warrant](warrant-entry/BRIEF.md) | [`warrant-entry/`](warrant-entry) | Evidence-aware reduction removes enough consumer complexity to earn a package. |
| [Parley](parley/BRIEF.md) | [`parley/`](parley) | Protocol-derived sequencing catches useful mistakes without absorbing evidence and persistence. |

Read only this common packet and the assigned brief. Never read competing implementations, exchange progress, or share code between entries. The operator launches one fresh session per entry. All three get the same substantive correctness requirements; no weaker baseline.

## The one workflow

A deterministic worker script stands in for an agent at a tool boundary. It produces a local candidate artifact for a declared source revision, checks the candidate, demonstrates the checker rejecting a planted defect, and delivers the candidate to a disposable local sink. Completion names the exact source, candidate, evidence, and delivery receipt.

This first round tests execution semantics, not model reasoning. No API calls or credentials are needed. A later adoption trial can put a real agent behind the same boundary.

Two revisions, S1 and S2, have distinct content digests. Each logical run has a run ID. A replacement worker gets a fresh incarnation. Receipts bind to the subject they actually concern; changing to S2 cannot rename S1 work or turn S1 checks into evidence for S2.

The checker must use actual candidate content, and its planted defect must be detected by the same checking path. A stored boolean or a canned string saying “checks passed” is insufficient.

## Two local effect boundaries

Implement both modes independently in each entry, with the same observable contract:

- **Queryable, deduplicated sink.** A local service or subprocess owns durable effect records. A request carries a stable operation key and payload digest. Repeating the same key and payload returns the retained receipt without another effect. The same key with a different payload is refused. Query by operation key is supported. Once commit is acknowledged as durable, a worker-process crash does not erase it.
- **Opaque sink.** A local subprocess performs a non-idempotent effect but provides no supported query or deduplication after the call is lost. A crash before the worker retains the result creates an unresolved outcome. The replacement must park that case with the affected operation identity; it cannot guess completion or automatically repeat the effect.

The stable operation key includes the logical run, step, subject, and relevant input identity. It must survive a worker incarnation change when continuing the same operation. A fresh incarnation fences messages; it does not itself create a new logical outside effect.

The test controller may inspect the sink's records to establish ground truth. The recovering worker may use only the declared sink interface; reading the opaque sink's private files to resolve ambiguity is not allowed.

Scope is one host and process crashes. State the file-write/flush assumptions. Do not claim arbitrary power-loss durability or distributed consensus. A single local sink may serialize its requests; this does not prove multi-host behavior.

## Required cases

| Case | Observable required outcome |
| --- | --- |
| Uninterrupted good run | Valid S1 output, detected negative control, one delivery effect, named completion |
| Crash after intent, before effect | Replacement proceeds or reconciles correctly under the declared sink semantics |
| Queryable effect committed, acknowledgement withheld, worker killed | Replacement obtains the retained receipt; one effect total; valid completion |
| Receipt durable, next step not yet run | Replacement reuses the committed result; one effect total |
| Opaque effect happened, result not retained | Named unresolved outcome; no automatic repeat and no fabricated completion |
| Late terminal from replaced incarnation | Old terminal does not advance the current run; reason retained |
| Source changes to S2 | S1 evidence cannot certify S2; the S2 positive path can eventually complete after fresh work |
| Candidate bytes change after checking | Old checks cannot certify the changed bytes |
| Truncated final journal record | Complete prefix recovered, unresolved effects handled honestly |
| Corrupt complete journal record or definition mismatch | Explicit refusal; no best-effort invented state |
| Repeated sink key, conflicting payload | Refusal; prior effect and receipt retained |

A legitimate S1 effect already committed before a revision change remains historical fact. Do not claim it was rolled back, and do not reuse it as delivery of S2. The late-terminal case concerns run advancement, not retroactive cancellation of an already committed effect.

The positive paths matter: refusing everything cannot pass. Bound retries and waits. There must be a named terminal or unresolved outcome rather than a hung demo.

## Real interruption, controlled timing

Use explicit test handshakes at semantic boundaries: intent durable, sink committed, reply withheld, receipt durable. The test controller terminates only its own worker subprocess and starts a replacement using the saved journal. Do not emulate a crash solely by returning an error in the original process. Do not rely on arbitrary sleeps to hit a race.

Every run leaves the input identities, journal, sink receipts or controller observations, expected outcome, actual outcome, and the commands required to replay it. Use temporary directories owned by the demo. Do not touch live Org, Gate, fleet, credentials, or production state.

## Negative controls for the implementation

Provide deliberately broken variants for at least these cases: blindly retry after a lost acknowledgement, accept an old-incarnation terminal, and certify changed candidate bytes from old evidence. The common cases must detect each defect. Keep mutants out of the normal execution path.

The controller's outcome checks must be independently expressed from the implementation's own success flag. A matching final label without the correct sink effect count and identity is insufficient.

## House rules

1. **One session, one fresh repo.** Use only the assigned path. If it already contains work, report the collision. Read applicable local instructions. Do not change Warrant, Parley, Workbench, Rooms, or the frozen bakeoff archive.
2. **Correctness is computed.** Table-driven checks or executable assertions determine outcomes from artifacts and effects. A model never grades the run. Language and abstraction preference cannot substitute for passing the same cases.
3. **No spec or design doc.** Working implementation, README, DEMO.md, and meaningful tests of the policy and crash behavior. Keep this one workflow small enough for a build session.
4. **Keyless and local.** No external APIs, cloud services, paid dependencies, message brokers, Kubernetes, or toolchain installation. Inspect available runtimes. If a required runtime is absent, report exactly what could not be executed rather than claiming a substitute proves it.
5. **Finish line.** One-command README; a hands-free canned demo; exact 60-second DEMO.md; local build/tests green; retained inspectable results. Keep the build local. A terminal walkthrough is enough; no dashboard is required.

Reuse by the Warrant and Parley entries is deliberate. Pin the inspected source commit, retain its license, and record any local adaptations. Count both the new consumer code and the relevant reused machinery in the maintenance comparison. Do not present a twelve-line wrapper as the entire implementation.

## Judging — operator, after the builds

The proposed research weighting retains all five house rubric slots. Settle changes before launch. Scores are blank; no automatic judge or ranking is supplied.

| Criterion | Weight | What the operator assesses |
| --- | ---: | --- |
| 60-second demo | 30 | Is the crash, recovery decision, and effect count immediately understandable? |
| Would someone pay | 10 | Is there a concrete reliability/debugging need worth engineering time? External demand is unvalidated; invent no customer evidence. |
| Deterministic share | 30 | Do the required cases and planted mutants distinguish correct behavior from convincing false success? |
| Approach necessity | 20 | Does the chosen machinery improve this problem over the simplest adequate alternative? Plain Go can win this criterion. |
| Restraint | 10 | Total maintained complexity, adapter size, dependencies, and operational steps |

Tie-breaker: which would you actually adopt next week? Report cases passed, untested assumptions, code boundaries, and commands as facts. The operator scores entries and decides promotion. A losing implementation may still contribute a useful test fixture, but do not merge the three into a platform.

## Launch prompts

Original launch prompts, retained for reproduction. The operator subsequently authorized their execution in independent fresh-context agents.

> Read `plain-go/BRIEF.md` and build it. You are one of three independent entries; win by demo, not by design. Follow the common rules it links and deliver a runnable local result with an exact 60-second walkthrough.

> Read `warrant-entry/BRIEF.md` and build it. You are one of three independent entries; win by demo, not by design. Follow the common rules it links and deliver a runnable local result with an exact 60-second walkthrough.

> Read `parley/BRIEF.md` and build it. You are one of three independent entries; win by demo, not by design. Follow the common rules it links and deliver a runnable local result with an exact 60-second walkthrough.
