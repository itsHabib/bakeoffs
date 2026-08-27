# Demo

Run `stack run`.

The workflow crashes after dispatch has been committed, resumes from its
journal, and finishes verify and record while the dispatch counter remains one.
It then changes the dispatch step's stable identity and prints the replay
refusal. `stack test` also covers the at-least-once window, reorder, duplicate,
version drift, and changed-input diagnostics.
