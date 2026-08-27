package main

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"

	"example.com/hack-mandate/mandate"
)

func main() {
	if len(os.Args) != 2 || os.Args[1] != "demo" {
		fmt.Fprintln(os.Stderr, "usage: mandate demo")
		os.Exit(2)
	}
	root, err := filepath.Abs(".")
	if err != nil {
		fatal(err)
	}
	result, err := mandate.BuildDemo(root)
	if err != nil {
		fatal(err)
	}

	gaps, _ := json.Marshal(result.Envelope.OracleAssurance.Gaps)
	fmt.Printf("oracle H2: summary=%s  coverage=%q  gaps=%s  merge_authority=%s\n", result.Envelope.OracleAssurance.Summary, result.Envelope.OracleAssurance.Coverage, gaps, result.Envelope.OracleAssurance.MergeAuthority)
	fmt.Printf("without mandate: %s %s\n", upper(result.WithoutMandate.Decision), result.WithoutMandate.ReasonCode)
	fmt.Println("root -> child review_candidate: signatures valid")
	fmt.Printf("audience-signed review request: %s receipt %s\n", upper(result.ValidRequest.Decision), mandate.ReceiptDigest(result.ValidRequest))
	fmt.Println("validly signed child adds publish_candidate:")
	fmt.Printf("  signatures-only mutant: %s  <-- planted bug\n", upper(result.SignaturesOnly.Decision))
	fmt.Printf("  mandate: %s %s\n", upper(result.Production.Decision), result.Production.ReasonCode)
	fmt.Printf("standard capability + receipt: %s\n", result.CapabilityComparator)
	printJSON("MandateDecisionReceiptV1", result.ValidRequest)
	fmt.Printf("receipt digest: %s\n", mandate.ReceiptDigest(result.ValidRequest))
	printJSON("common envelope", result.Envelope)
}

func printJSON(label string, value any) {
	b, err := json.MarshalIndent(value, "", "  ")
	if err != nil {
		fatal(err)
	}
	fmt.Printf("%s:\n%s\n", label, b)
}

func upper(decision mandate.Decision) string {
	if decision == mandate.Allow {
		return "ALLOW"
	}
	return "DENY"
}

func fatal(err error) {
	fmt.Fprintln(os.Stderr, err)
	os.Exit(1)
}
