# 60-second story

**0–10 seconds:** “A customer types `ca`, then `cats`. The cats response finishes first. The late ca response overwrites it. Our incident log has eleven events. Which ones matter?”

**10–35 seconds:** Run `./demo.sh` (first prepare local Kitesurf as described in README). Watch Needle try shorter event sequences. `pass` means the declared invariant holds; `different-failure` means the symptom changed; `unresolved` means the shortened schedule cannot finish validly.

**35–50 seconds:** Open `artifacts/search/report.md` and `minimal.json`. The goal is a short replay that retains both requests and their reversed responses, with unrelated UI and warmup activity removed. Show the exact actual/expected values and the deletion witnesses. The second workload applies the same tool to seat quantities and server quotes.

**50–60 seconds:** Show `fixed-control.json`: the same original schedule passes when the page ignores stale responses. Hand the reduced JSON to the developer fixing the bug. It is executable input, not a model's explanation of what probably happened.

The timed story describes what to show; actual runtime is recorded in RESULTS and is not promised to fit 60 seconds.
