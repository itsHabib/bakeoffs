# protocol-compiler

Projects a branching global agent protocol into one local contract per role and
refuses a protocol when a role has branch-dependent obligations but cannot know
which branch occurred.

```powershell
stack run
stack test
```

## Result

The POC's narrow guarantee is useful: global choice ownership and observer
notification become compiler inputs, and projection produces a role-specific
refusal with the exact branch path. The valid fixture models an evidence receipt
moving through collector, assurance kernel, human judgment, and gate.

## Why Haskell

`Protocol` is a closed algebraic program value. Projection and rendering are
total interpreters, and `-Wall -Werror` makes a newly added syntax constructor
break every incomplete interpreter at compile time. An idiomatic Go version
would use an interface/type-switch visitor and runtime exhaustiveness tests; it
can implement the same algorithm, but it does not get the same closed-world
case pressure from the compiler.

The POC does **not** prove deadlock freedom or protocol fidelity at runtime. It
checks its declared projection rule, including choice observability, over the
provided finite protocol value.

## Strongest alternative

Scribble-style tooling or a small Rust enum-based compiler. Rust preserves
exhaustive matching and narrows the Haskell advantage substantially; Haskell's
remaining edge is the ease of treating the protocol as a compositional EDSL
with several pure interpreters.
