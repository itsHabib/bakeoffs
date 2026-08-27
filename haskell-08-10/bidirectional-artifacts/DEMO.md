# Demo

Run `stack run`.

The executable generates a three-step Ship-like workflow, applies a legal
model/concurrency/retry override, and renames the upstream project. The printed
artifact shows new derived task IDs with the human policy intact. It then edits
a derived dependency and prints the structured refusal.

Run `stack test` for the three round-trip laws and boundary cases.
