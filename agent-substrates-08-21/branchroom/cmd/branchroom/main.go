package main

import (
	"encoding/json"
	"fmt"
	"os"
	"strings"

	branchroom "github.com/itshabib/hack-branchroom"
)

func main() {
	result, err := branchroom.ByteStableDemo("fixtures/exact-head-lifecycle-v1.json")
	if err != nil {
		fmt.Fprintln(os.Stderr, "branchroom:", err)
		os.Exit(1)
	}
	control := result.Experiment.Branches[0]
	counterfactual := result.Experiment.Branches[1]
	production := result.Experiment.RejectedTerminal.RejectedBy["control"]
	fmt.Printf("capsule %s  common-prefix %s  terminal-seq 6\n", result.Experiment.CapsuleDigest, result.Experiment.PrefixDigest)
	fmt.Printf("control epoch %d        own terminal: %s\n", control.Branch.Epoch, verdict(control.OwnTerminal))
	fmt.Printf("counterfactual epoch %d own terminal: %s\n", counterfactual.Branch.Epoch, verdict(counterfactual.OwnTerminal))
	fmt.Printf("mutant epoch 12         late parent terminal: %s  <-- planted bug\n", verdict(result.MutantDecision))
	fmt.Printf("branchroom              late parent terminal: %s %s\n", verdict(production), production.Reason)
	fmt.Printf("same late event: %s\n", result.MutantEventID)
	fmt.Printf("first divergence: event %d\n", result.Experiment.FirstDivergence.EventOrdinal)
	fmt.Println("determinism: 2 normalized runs BYTE-IDENTICAL")
	printJSON("ExperimentForkV1", result.Experiment)
	fmt.Println("artifact digest:", result.ArtifactDigest)
	printJSON("normalized envelope", result.Envelope)
}

func verdict(decision branchroom.Decision) string {
	if decision.Accepted {
		return "ACCEPT"
	}
	return strings.ToUpper("refuse")
}

func printJSON(label string, value any) {
	data, err := json.MarshalIndent(value, "", "  ")
	if err != nil {
		panic(err)
	}
	fmt.Printf("\n%s\n%s\n", label, data)
}
