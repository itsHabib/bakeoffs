# HCL entry

Read [`../RULES.md`](../RULES.md) for the common workload, boundaries,
deliverables and assessment. This entry was built independently without
inspecting the other contenders.

The bet: an established configuration language can describe resources, dependencies and repeatable work clearly, while a small compiler and adapter interface supply execution semantics. Use actual HCL parsing rather than an HCL-looking hand parser. Keep the supported subset deliberate and provide useful errors for unsupported constructs.

Build the six-case common demo, showing authored source, normalized representation, plan and apply. Show a third adapter without parser/planner edits. Explain which HCL capabilities earn their cost and how the design handles state and one-shot work. Avoid copying Terraform's provider distribution, remote state services, module registry, cloud accounts or compatibility surface.

Finish with runnable code, tests, README, `DEMO.md`, independent correctness review and a draft POC PR. The user judges the approach; report evidence and limits rather than declaring a winner.
