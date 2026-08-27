# Scorecard — haskell-dsl-08-10

> **One entry is withheld from this public archive.** It was built against
> day-job domain rules and named an employer product, so it is not mine to
> publish. Its scores, validation result and judging notes stay in place —
> a scorecard quietly edited down to five entries would misrepresent the
> round.

| entry | Haskell /25 | Authoring /20 | Utility /20 | Leverage /15 | Hard case /10 | Restraint /10 | **Total /100** | promote case | kill case |
|---|---:|---:|---:|---:|---:|---:|---:|---|---|
| assurance-dsl | 21 | 18 | 18 | 14 | 9 | 9 | **89** | compact policy yields demands, decisions, and proof-path explanations | a TypeScript AST can preserve most semantics and is easier to integrate |
| delegation-dsl | 16 | 16 | 17 | 12 | 8 | 8 | role briefs and dependency batches come from one artifact contract | bounded syntax is still a pleasant manifest; branch semantics would expand scope |
| capability-dsl | 23 | 17 | 17 | 15 | 10 | 9 | heterogeneous typed results coexist with complete static preflight | useful agents often need result-dependent operations that hide the future envelope |
| recovery-dsl | 23 | 17 | 18 | 14 | 10 | 8 | retry declarations materially change the same crash window's behavior | persistence and adapters risk turning the DSL into another workflow engine |
| *withheld — see note* | 22 | 18 | 19 | 15 | 10 | 8 | executable rule semantics and snapshot requirements cannot drift | existing TypeScript/C# engines make a Haskell IR boundary expensive to adopt |
| work-driver-dsl | 21 | 19 | 20 | 15 | 10 | 9 | **94** | opaque references and do-notation compile daily project intent into safe batches | file scopes are declarations, not truth; Haskell authoring must beat the current prep flow |

## Verdict

- Winner: **work-driver-dsl — 94/100**
- Promote to: a standalone Haskell plan compiler, working name **braid**, whose
  output is consumed by the existing work-driver flow.
- Why: it is the language the operator is most likely to author next week. The
  POC makes task references safe by construction, computes maximum compatible
  batches, explains requested parallelism that was serialized by file overlap,
  reports a critical path, and refuses landing without a validation ancestor.
  Its boundary is unusually clean: compile project intent; never dispatch work.
- Runner-up: **the withheld entry — 92/100**
- Why it lost: it solves a concrete and expensive day-job parity problem better
  than the other candidates, but adoption requires an IR boundary inside two
  existing production engines. The winning work-driver compiler can provide
  value immediately as a front-end to an already-used tool.
- Negative results: **delegation-dsl scores 77 and fails both the 78 total and
  18-point Haskell floors.** Its local briefs are useful, but the POC confirms
  that the bounded language is still a dependency manifest with constructors.

## Ranking

1. `work-driver-dsl` — promote the compiler boundary.
2. *withheld* — strongest work-specific candidate; kept for an authorized
   day-job experiment rather than a personal flagship.
3. `capability-dsl` — excellent bounded sublanguage for pre-authorized actions.
4. `recovery-dsl` — semantically deep, but operational surface grows quickly.
5. `assurance-dsl` — useful evidence vocabulary, likely best embedded elsewhere.
6. `delegation-dsl` — rejected as insufficiently different from configuration.

Do not merge the runner-up mechanisms into the winner. `braid` should not gain
capabilities, replay, assurance policy, or agent handoff semantics until a real
work-driver plan demonstrates that the compiler cannot do its one job without
them.
