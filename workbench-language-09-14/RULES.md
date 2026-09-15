# Workbench language bakeoff

Three independent POCs explore a general declarative execution model: describe intended resources and work, inspect a plan/diff, apply it, and safely repeat or resume. Fleet and Rooms are possible consumers, not the definition of the language.

## What the language must earn

Compare against direct use of the same underlying operations with a small script and existing records. Include a runnable direct-tools baseline for the common workload; do not deliberately handicap it or require it to implement language-specific features. Show whether the declarative approach improves inspection, change handling or recovery enough to justify its extra concepts. A conclusion that the baseline suffices is legitimate; the operator chooses.

The operational foundation owns real effects and observations. A language is an optional client of it, not a replacement supervisor or a duplicate authority/ownership registry. Keep parser/planner state distinct from backend facts. Evaluate interrupted-run recovery and the information a fresh caller needs; disclose when recovery depends on new persistent state introduced by the entry.

## Entries

| Entry | Bet | Archived entry |
|---|---|---|
| HCL | An established configuration language provides useful references and readable diffs without owning a parser | [`hcl/`](hcl) |
| Custom language | A small purpose-built syntax makes intent and lifecycle clearer than general configuration | [`custom-language/`](custom-language) |
| Typed API | An ordinary typed programming language provides composition and tooling with less new machinery | [`typed-go/`](typed-go) |

## Common workload

Use the same small local workflow: manage an input text artifact and run a transformation that produces an output artifact. Add a third, materially different adapter after the initial two, such as a disposable directory resource. Demonstrate that extending the system does not require editing the language parser or planner. The extension mechanism is required; dynamic plugin loading is a choice to justify, not a requirement.

The source must express a dependency/reference between input and transformation. Show source -> normalized intent -> readable proposed changes -> apply -> observed outcome. Distinguish persistent desired resources from one-shot work and retained evidence. Do not assume exactly-once effects or automatic rollback.

Each canned demo must show:

1. Initial plan and apply in a freshly created disposable directory.
2. Unchanged re-apply with no unintended repeated effects.
3. An input change and the resulting downstream plan.
4. External modification between plan and apply, with honest handling of the stale plan.
5. An injected partial failure followed by retry without repeating already completed unrelated effects.
6. The third adapter and the exact extension code it requires.

Declare resource identity, task replay semantics, ownership of files, and which inputs make a result stale. Validate behavior with deterministic executable assertions; readable output alone is not proof. Restrict demo effects to its disposable directory and child processes. Existing Fleet/Rooms integration can be explained or simulated explicitly; do not mutate live Fleet state, start real Rooms infrastructure, or claim simulations prove those integrations.

## Working rules

- One fresh repository per entry. Do not inspect other entries' code, conclusions, or progress, and do not share libraries between entries during the bakeoff.
- Read applicable local instructions. Use local, keyless tooling. No purchases, new paid accounts, cloud resources, global setup changes, or deployment.
- Build a small runnable POC, not a platform or a language-feature checklist. Choose the stack appropriate to the bet; explain necessary dependencies.
- README plus working code, meaningful tests and `DEMO.md` suffice; no separate speculative design document is required.
- Keep human authority and provider-specific boundaries intact. Do not mint grants or weaken checks.
- Finish with a draft POC PR, independent correctness review and concrete disposition of findings. Publish only scrubbed source, not local artifacts, private paths or credentials. Discover repository/remotes and visibility deliberately; do not create a public repository implicitly. If PR publication needs an unavailable repository decision, complete the local artifact and report that exact blocker.
- No merge. Michael assesses the competing POCs and chooses what to pursue.

## Deliverables

One-command canned demo, deterministic tests for the six cases, a 60-second `DEMO.md`, source/compiled-plan/diff examples, a small provider-extension example, and a draft PR with verification and limits. Explain the value to a Workbench operator and how this compares with a plain script or data file. Record actual dependencies and complexity; do not optimize for a line-count target.

## Human assessment

The operator scores after trying the demos: demo clarity 30, practical value / plausible buyer 25, executable correctness 20, need for this representation versus a simpler alternative 15, restraint 10. Do not invent commercial evidence. Scores and winner remain blank until Michael assesses the entries; factual test results can be recorded.

## Launch prompts

Give each entry to a fresh independent agent with only this common README and its own brief. No surrounding conversation or other entry results. Each brief is self-contained through the shared workload above.

- Read [`hcl/BRIEF.md`](hcl/BRIEF.md) and build its POC through a draft PR.
- Read [`custom-language/BRIEF.md`](custom-language/BRIEF.md) and build its POC through a draft PR.
- Read [`typed-go/BRIEF.md`](typed-go/BRIEF.md) and build its POC through a draft PR.
