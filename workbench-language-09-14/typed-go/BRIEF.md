# Typed host-language API entry

Read [`../RULES.md`](../RULES.md) for the common workload, boundaries,
deliverables and assessment. This entry was built independently without
inspecting the other contenders.

The bet: a small typed API in an existing language offers references, composition and extension without a new configuration language. Choose a locally practical host language. Keep describing desired work separate from executing its effects; make the boundary inspectable, and state clearly whether evaluating user source executes arbitrary code.

Build the six-case common demo, emitting normalized intent and readable plan before apply. Show a third adapter without planner changes. Explain determinism, authoring/tooling advantages, and where ordinary host-language freedom becomes a liability. Avoid a general workflow engine, remote services, generated SDK ecosystem or claims of sandboxing untrusted source.

Finish with runnable code, tests, README, `DEMO.md`, independent correctness review and a draft POC PR. The user judges the approach; report evidence and limits rather than declaring a winner.
