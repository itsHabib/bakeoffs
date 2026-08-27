# Validation

Validated with the pinned container
`ghcr.io/gleam-lang/gleam:v1.18.1-erlang-alpine@sha256:7c82e4a284b7c05c26eac34db497ea0e63ce7cb04bd019d966d70338eb172b68`.
The final check ran from a clean `git archive`, with no working-tree build cache.

The one-command check runs:

1. dependency resolution from `manifest.toml`;
2. `gleam format --check src test`;
3. `gleam check`;
4. `gleam test`;
5. the journal-backed shipping demo;
6. the fresh-process evidence reproduction.

Latest result: **12 tests passed**, followed by a finished attempt-two shipping
view for `rev-b` and a separately reproduced `supported` E2E verdict.

The suite covers:

- a failed E2E loop from `rev-a` back to `ship`, then success on `rev-b`;
- refusal to reuse `rev-a` validation for `rev-b`;
- stable-key recovery for an uncommitted deduplicated effect;
- explicit reconciliation for an uncommitted at-least-once effect;
- identical uninterrupted and journal-replayed projections;
- pipeline digest drift refusal;
- torn-tail truncation;
- refusal to append an invalid transition;
- execution of the maintenance topology with the same kernel;
- independent evidence reproduction and changed-input refusal;
- replay-manifest codec round trips.
