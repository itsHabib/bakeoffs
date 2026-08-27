# Demo — protocol compiler

Run:

```powershell
stack run
stack test
```

The first half prints local obligations for producer, collector, kernel, human,
and gate from one global evidence-assurance protocol.

The second half removes `gate` from the outer choice's observer set. Gate must
act only on the accepted branch, so it cannot distinguish accepted from
malformed. The compiler refuses the protocol at `gate` instead of emitting an
ambiguous local contract.

The tests also pin self-message refusal and the lawful case where an uninvolved
role has identical obligations in both branches and therefore needs no
notification.
