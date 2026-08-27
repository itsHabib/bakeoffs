# Demo

Run `stack run`. It shows the authored operation order and full envelope,
refuses a read-only grant before any effect, executes under the exact grant,
then proves that merge requires a separate capability and mutation tier.
