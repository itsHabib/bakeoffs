# Demo

Run `stack run`. The same crash window executes a deduplicated dispatch once and
an at-least-once notification twice, with explicit trace events. A committed
result is reused, and a changed stable step ID refuses replay.
