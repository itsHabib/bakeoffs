package main

import (
	"encoding/json"
	"fmt"
	"github.com/itsHabib/bakeoffs/agent-substrates-08-21/proofline/internal/proofline"
	"os"
)

func main() {
	if len(os.Args) != 2 || os.Args[1] != "demo" {
		fmt.Fprintln(os.Stderr, "usage: proofline demo")
		os.Exit(2)
	}
	m, err := proofline.Load("fixtures/exact-head-lifecycle-v1.json")
	if err != nil {
		fatal(err)
	}
	h1ID, h2ID := m.InitialAssurance, m.RepairedAssurance
	h1Current, err := m.Graph.Assess(h1ID, "", false)
	if err != nil {
		fatal(err)
	}
	h1AfterH2, err := m.Graph.Assess(h1ID, m.RepairedContract, false)
	if err != nil {
		fatal(err)
	}
	h2Current, err := m.Graph.Assess(h2ID, "", false)
	if err != nil {
		fatal(err)
	}
	mutant, err := m.Graph.Assess(h1ID, m.DriftID, true)
	if err != nil {
		fatal(err)
	}
	production, err := m.Graph.Assess(h1ID, m.DriftID, false)
	if err != nil {
		fatal(err)
	}
	o1, o2 := m.OracleForAssurance[h1ID], m.OracleForAssurance[h2ID]
	fmt.Printf("H1 lineage: %s  oracle=%s  coverage=%s  authority=%s\n", h1Current.Status, o1.Summary, o1.Coverage, o1.MergeAuthority)
	fmt.Printf("H2 created: H1 lineage %s  first stale=edge %s\n", h1AfterH2.Status, h1AfterH2.FirstStaleEdge)
	fmt.Printf("fresh H2 lineage: %s  oracle=%s  coverage=%s  authority=%s\n", h2Current.Status, o2.Summary, o2.Coverage, o2.MergeAuthority)
	fmt.Println("same head, task/policy changed:")
	fmt.Printf("  head-only mutant: %s  <-- planted bug\n", mutant.Status)
	fmt.Printf("  proofline: retracted %s\n", production.Reason)
	join, err := m.ExactSubjectJoin(m.RepairedContract)
	if err != nil {
		fatal(err)
	}
	fmt.Printf("\nexact-subject join: current=%s matched=%d impact_candidates=%d joined_impact=%d impact_explanation=%t\n", join.ExpectedContract, len(join.MatchedNodes), len(join.ImpactCandidates), len(join.AffectedOldNodes), join.ExplainsImpact)
	fmt.Println("proofline impact: first stale edge plus 3 transitive retractions; historical H1 remains queryable")
	lineage, err := m.Graph.Lineage(h1ID)
	if err != nil {
		fatal(err)
	}
	printJSON("\nordered ProoflineV1 subgraph:", lineage)
	fmt.Printf("retraction set: %v\n", production.RetractedNodes)
	fmt.Printf("artifact digest: %s\n", lineage.Digest())
	envelope, err := m.DemoEnvelope()
	if err != nil {
		fatal(err)
	}
	printJSON("common envelope:", envelope)
}
func printJSON(label string, v any) {
	b, err := json.MarshalIndent(v, "", "  ")
	if err != nil {
		fatal(err)
	}
	fmt.Println(label)
	fmt.Println(string(b))
}
func fatal(err error) { fmt.Fprintln(os.Stderr, err); os.Exit(1) }
