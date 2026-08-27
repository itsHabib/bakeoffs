# DEMO.md — the 60-second walkthrough

## Run it (hands-free, deterministic)

```bash
make venv   # once
make demo
```

No ollama required — the demo replays committed cassettes, so it's identical
every run. For the live bonus: `make demo-live` (needs `ollama run qwen2.5:7b`).

## What the judge sees

Nine rows — two clean + one adversarial per class — then the ledger:

```
  narrow    clean_test_files           -> PASS ✓
  narrow    clean_markdown             -> PASS ✓
  narrow    adversarial_cmd_dir        -> REJECT ✗  (adversarial)
      gate reason: included items that do not match criterion (regex '(^|/)cmd/'): ['docs/cmd-notes.md']
  extract   clean_gotest               -> PASS ✓
  extract   clean_npm                  -> PASS ✓
  extract   adversarial_docker_digest  -> REJECT ✗  (adversarial)
      gate reason: value 'sha256:deadbeefcafe0000feed' for field 'digest' not found verbatim in input (hallucinated value)
  classify  clean_logs                 -> PASS ✓
  classify  clean_gotest               -> PASS ✓
  classify  adversarial_security       -> REJECT ✗  (adversarial)
      gate reason: planted line 0: expected label 'SECURITY', got 'STYLE'

  class       calls  pass  reject  rej.rate  displaced tok
  --------------------------------------------------------
  classify        3     2       1      33%            267
  extract         3     2       1      33%            330
  narrow          3     2       1      33%            243
  --------------------------------------------------------
  TOTAL           9     6       3      33%            840
  Net: 840 frontier tokens displaced to a $0 local model across 6 verified
  calls; 33% of calls were caught and rejected (NOT silently passed through).
```

## The 60-second script

> "Offloading to a local model saves frontier tokens, but nobody trusts a 7B
> model's output — so I made trust **mechanical**. Three task classes, each with
> a verifier that can't be sweet-talked: **subset-only**, **substring-only**,
> **label-set-only**.
>
> Watch the money shot — extract. The build output has no image digest. Asked
> for one, the 7B model **invented a `sha256`** that isn't anywhere in the input.
> The substring verifier catches it and rejects with the exact field and value.
> Same story in classify: it downgraded a **SQL-injection** line to STYLE — the
> planted ground-truth check catches the mislabel that actually matters. And in
> narrow it swept in `docs/cmd-notes.md` because it saw the substring 'cmd' — the
> regex says that's not under a `cmd/` directory, rejected.
>
> Now the ledger: across the demo corpus, **840 frontier tokens displaced**,
> **33% rejection rate** — every reject caught, none silently passed through.
> That's the number that turns 'use a cheaper model sometimes' from vibes into
> policy. And the obvious alternative — just flip the model picker to a cheaper
> tier — gives you the same tokens with **none of the catch and none of the
> number**."

## The one-line proof it's real code, not vibes

```bash
make test   # 15 table tests over every verifier + the gate loop
```

## Would someone pay (assert it, expect pushback)

Buyer: teams running frontier-priced agents at scale — the customers LLM routers
(Martian, Not Diamond, OpenRouter, LiteLLM/Portkey) already sell to. Routers send
cheap calls to cheap models and *hope*; none sell a per-call **correctness gate**
on the cheap model's output. This is that missing half — plus the ledger a CFO
uses to sign off on the switch. The demo's rejection column is the pushback
answer: this is what routing-and-hoping can't give you.
