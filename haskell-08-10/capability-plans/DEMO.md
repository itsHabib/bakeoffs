# Demo

Run `stack run`.

The demo prints the complete envelope and operation order of a
read/check/comment plan, refuses it under a read-only grant, and executes it
under the exact grant. It then adds merge and shows that comment authority is
insufficient because merge has its own capability and mutation class.
