package main

import (
	"fmt"
	"os"

	obligation "hack-obligation"
)

func main() {
	stages, err := obligation.RunDemo(".")
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
	fmt.Print(`aligned H1 frontier: empty
late REFUTED snapshot arrives:
  settled-is-terminal mutant: empty  <-- planted bug
  obligation engine: OPEN resolve_refutation(critical-finding-resolved)
repair -> H2 partial: OPEN collect_claim(critical-finding-resolved)
agent remove mandatory claim: REFUSE mandatory_contract_weakened
refreshed H2 frontier: empty  (evidence only; no authority)
named-node scheduler + reopen hook: OPEN repair-critical-finding-resolved
  native artifact parity: no (no contract-derived identity or overlay floor)`)
	fmt.Println("\n\nObligationFrontierV1 + common envelope:")
	data, _ := obligation.Pretty(stages.Output)
	_, _ = os.Stdout.Write(data)
}
