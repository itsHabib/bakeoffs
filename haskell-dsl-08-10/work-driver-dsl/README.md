# Work-driver DSL

A small embedded language for authoring project tasks, dependencies, file
scopes, intended parallelism, validation, and landing. It compiles to maximum
safe batches, explains forced serialization, and reports a critical path.

Task references are opaque values created inside the project builder: missing
dependencies and forward-reference cycles are not representable through the
public DSL. Landing still receives a dynamic ancestor check because task kind is
project data, not a type index.

```powershell
stack test
stack run
```

The strongest alternative is the current manifest plus a TypeScript or Go graph
compiler. Haskell's do-notation and opaque references make authoring unusually
pleasant, but the candidate loses if the notation obscures rather than clarifies
computed parallelism.
