# Bidirectional artifacts

A lawful reconciler for generated agent-workflow artifacts. Compact intent owns
task identity and dependency structure; reviewers may override a bounded
execution policy. Regeneration preserves those overrides and refuses edits to
derived structure instead of silently winning or losing them.

The POC intentionally stops at the semantic boundary: typed values in, typed
values out. It does not build a YAML parser, templating platform, or editor.

```powershell
stack test
stack run
```

The test suite executes GetPut, PutGet, and PutPut over an adversarial finite
corpus, then exercises upstream renames, illegal derived edits, and policy
bounds.

## Why Haskell

The mapping is one value containing both directions, and its contract is stated
as executable equations over immutable values. That makes law failures local:
the generator and reconciler cannot quietly evolve as unrelated utilities.

The result is also an honest negative signal. In this bounded POC the kernel is
still two ordinary conversion functions packaged together; the types do not
prevent an incoherent `put`. Haskell makes the laws pleasant to state and test,
but the language does not enforce them.

## Strongest alternative

A Rust pair of exhaustive conversion functions plus property tests would retain
nearly all of the demonstrated value. Real YAML/Markdown artifacts would also
need a concrete-syntax-preserving parser, which may dominate the elegant
bidirectional core. That is why this candidate does not clear the Haskell bar.
