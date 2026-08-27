package obligation

import (
	"encoding/json"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"slices"
	"strings"
)

type AssuranceEcho struct {
	CaseID         string   `json:"case_id"`
	Summary        string   `json:"summary"`
	Coverage       string   `json:"coverage"`
	Gaps           []string `json:"gaps"`
	MergeAuthority string   `json:"merge_authority"`
}

type CommonEnvelope struct {
	SchemaVersion            string          `json:"schema_version"`
	Entry                    string          `json:"entry"`
	CaseID                   string          `json:"case_id"`
	InputBundleDigest        string          `json:"input_bundle_digest"`
	Subject                  GoalIdentity    `json:"subject"`
	NativeStatus             string          `json:"native_status"`
	ReasonCodes              []string        `json:"reason_codes"`
	NativeArtifactDigest     string          `json:"native_artifact_digest"`
	DisplayedOracleAssurance []AssuranceEcho `json:"displayed_oracle_assurance"`
}

type Comparator struct {
	Name                    string `json:"name"`
	SettledBeforeLate       bool   `json:"settled_before_late"`
	LateNode                string `json:"late_node"`
	ExplicitReopenHook      bool   `json:"explicit_reopen_hook"`
	ContractDerivedIdentity bool   `json:"contract_derived_identity"`
	MonotoneOverlayFloor    bool   `json:"monotone_overlay_floor"`
	ReplayableFrontier      bool   `json:"replayable_frontier"`
	SameNativeArtifact      bool   `json:"same_native_artifact"`
}

type DemoOutput struct {
	Artifact        Frontier       `json:"obligation_frontier_v1"`
	ArtifactDigest  string         `json:"artifact_digest"`
	CommonEnvelope  CommonEnvelope `json:"common_envelope"`
	Comparator      Comparator     `json:"scheduler_comparator"`
	Refusal         string         `json:"overlay_refusal"`
	RawReceipt      string         `json:"raw_receipt_refusal"`
	SourceLineCount int            `json:"source_line_count"`
}

type DemoStages struct {
	Aligned    Frontier
	MutantLate Frontier
	NativeLate Frontier
	Partial    Frontier
	Final      Frontier
	Output     DemoOutput
}

type schedulerNode struct {
	Name    string `json:"name"`
	ClaimID string `json:"claim_id"`
	State   string `json:"state"`
}

func runScheduler(contract Contract, snapshots ...Assurance) ([]schedulerNode, bool) {
	nodes := []schedulerNode{}
	for _, claim := range contract.ObligationPolicy.MandatoryClaimIDs {
		nodes = append(nodes, schedulerNode{Name: "collect-" + claim, ClaimID: claim, State: "open"})
	}
	reopened, settled := false, false
	for _, snapshot := range snapshots {
		wasSettled := settled
		for claim, outcome := range snapshot.ClaimOutcomes {
			for i := range nodes {
				if nodes[i].ClaimID == claim && outcome == "supported" {
					nodes[i].State = "done"
				}
			}
			if outcome == "refuted" {
				nodes = append(nodes, schedulerNode{Name: "repair-" + claim, ClaimID: claim, State: "open"})
			}
		}
		settled = true
		for _, node := range nodes {
			settled = settled && node.State != "open"
		}
		reopened = reopened || wasSettled && !settled
	}
	slices.SortFunc(nodes, func(a, b schedulerNode) int { return strings.Compare(a.Name, b.Name) })
	return nodes, reopened
}

func compareScheduler(contract Contract, aligned, refuted Assurance, native Frontier) Comparator {
	nodes, reopened := runScheduler(contract, aligned, refuted)
	replayed, _ := runScheduler(contract, aligned, refuted)
	open := []schedulerNode{}
	for _, node := range nodes {
		if node.State == "open" {
			open = append(open, node)
		}
	}
	lateNode := ""
	if len(open) == 1 {
		lateNode = open[0].Name
	}
	afterDelete := slices.DeleteFunc(slices.Clone(nodes), func(node schedulerNode) bool { return node.ClaimID == "critical-finding-resolved" })
	nativeOpen := []Obligation{}
	for _, item := range native.Obligations {
		if slices.Contains(native.OpenFrontier, item.ObligationID) {
			nativeOpen = append(nativeOpen, item)
		}
	}
	identity := strings.Contains(lateNode, contract.ContractID) && strings.Contains(lateNode, contract.Subject.HeadSHA)
	return Comparator{
		Name: "named-node-scheduler-with-hooks", SettledBeforeLate: reopened, LateNode: lateNode,
		ExplicitReopenHook: reopened, ContractDerivedIdentity: identity,
		MonotoneOverlayFloor: len(afterDelete) == len(nodes),
		ReplayableFrontier:   digestJSON(nodes) == digestJSON(replayed),
		SameNativeArtifact:   digestJSON(open) == digestJSON(nativeOpen),
	}
}

func RunDemo(root string) (DemoStages, error) {
	fixtureBytes, err := os.ReadFile(filepath.Join(root, "fixtures", "exact-head-lifecycle-v1.json"))
	if err != nil {
		return DemoStages{}, err
	}
	var fixture Fixture
	if err := json.Unmarshal(fixtureBytes, &fixture); err != nil {
		return DemoStages{}, err
	}
	contract := func(id string) (Contract, error) {
		for _, value := range fixture.Contracts {
			if value.ContractID == id {
				return value, nil
			}
		}
		return Contract{}, fmt.Errorf("contract %s missing", id)
	}
	assurance := func(id string) (Assurance, error) {
		for _, value := range fixture.OracleAssuranceSnapshots {
			if value.CaseID == id {
				return value, nil
			}
		}
		return Assurance{}, fmt.Errorf("assurance %s missing", id)
	}
	h1, err := contract("contract-h1")
	if err != nil {
		return DemoStages{}, err
	}
	h2, err := contract("contract-h2")
	if err != nil {
		return DemoStages{}, err
	}
	aligned, err := assurance("aligned-h1")
	if err != nil {
		return DemoStages{}, err
	}
	refuted, err := assurance("refuted-h1")
	if err != nil {
		return DemoStages{}, err
	}
	partial, err := assurance("partial-h2")
	if err != nil {
		return DemoStages{}, err
	}
	refreshed, err := assurance("refreshed-h2")
	if err != nil {
		return DemoStages{}, err
	}

	h1Events := []Event{Activate(h1), Observe(aligned)}
	alignedFrontier, err := Reduce(h1Events, "")
	if err != nil {
		return DemoStages{}, err
	}
	lateEvents := append(slices.Clone(h1Events), Observe(refuted))
	mutantFrontier, err := Reduce(lateEvents, SettledIsTerminal)
	if err != nil {
		return DemoStages{}, err
	}
	nativeLate, err := Reduce(lateEvents, "")
	if err != nil {
		return DemoStages{}, err
	}

	partialEvents := append(slices.Clone(lateEvents), Activate(h2), Observe(partial))
	partialFrontier, err := Reduce(partialEvents, "")
	if err != nil {
		return DemoStages{}, err
	}
	_, overlayErr := Reduce(append(slices.Clone(partialEvents), ApplyOverlay(AgentOverlay{RemoveClaimIDs: []string{"critical-finding-resolved"}})), "")
	if RefusalReason(overlayErr) != ReasonWeakened {
		return DemoStages{}, fmt.Errorf("weakening overlay: got %v", overlayErr)
	}
	optionalEvents := append(slices.Clone(partialEvents), ApplyOverlay(AgentOverlay{Add: []OverlayAddition{{ClaimID: "timing-diagnostic", Mandatory: false}}}))
	optionalFirst, err := Reduce(optionalEvents, "")
	if err != nil {
		return DemoStages{}, err
	}
	optionalReplay, err := Reduce(optionalEvents, "")
	if err != nil || digestJSON(optionalFirst) != digestJSON(optionalReplay) {
		return DemoStages{}, fmt.Errorf("optional overlay did not survive replay")
	}
	_, rawErr := Reduce(append(slices.Clone(partialEvents), SubmitRawReceipt(fixture.TrustedRunnerReceipts)), "")
	if RefusalReason(rawErr) != "raw_receipt_input_forbidden" {
		return DemoStages{}, fmt.Errorf("raw receipt: got %v", rawErr)
	}

	finalEvents := append(slices.Clone(partialEvents), Observe(refreshed))
	finalFrontier, err := Reduce(finalEvents, "")
	if err != nil {
		return DemoStages{}, err
	}
	artifactDigest := ArtifactDigest(finalFrontier)
	echoes := make([]AssuranceEcho, 0, 4)
	for _, a := range []Assurance{aligned, refuted, partial, refreshed} {
		echoes = append(echoes, AssuranceEcho{CaseID: a.CaseID, Summary: a.Summary, Coverage: a.Coverage, Gaps: slices.Clone(a.Gaps), MergeAuthority: a.MergeAuthority})
	}
	lineCount, err := SourceLineCount(root)
	if err != nil {
		return DemoStages{}, err
	}
	output := DemoOutput{
		Artifact: finalFrontier, ArtifactDigest: artifactDigest,
		CommonEnvelope: CommonEnvelope{
			SchemaVersion: "agent-substrate-demo/v1", Entry: "hack-obligation", CaseID: "obligation-lifecycle",
			InputBundleDigest: DigestBytes(fixtureBytes), Subject: h2.Goal(), NativeStatus: "evidence_frontier_empty",
			ReasonCodes:          []string{ReasonReopened, "subject_superseded", "oracle_named_gap_opened", ReasonWeakened, "raw_receipt_input_forbidden"},
			NativeArtifactDigest: artifactDigest, DisplayedOracleAssurance: echoes,
		},
		Comparator: compareScheduler(h1, aligned, refuted, nativeLate),
		Refusal:    ReasonWeakened, RawReceipt: "raw_receipt_input_forbidden", SourceLineCount: lineCount,
	}
	return DemoStages{Aligned: alignedFrontier, MutantLate: mutantFrontier, NativeLate: nativeLate, Partial: partialFrontier, Final: finalFrontier, Output: output}, nil
}

func SourceLineCount(root string) (int, error) {
	count := 0
	err := filepath.WalkDir(root, func(path string, entry fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if entry.IsDir() && entry.Name() == ".git" {
			return filepath.SkipDir
		}
		if entry.IsDir() || !strings.HasSuffix(path, ".go") || strings.HasSuffix(path, "_test.go") {
			return nil
		}
		data, err := os.ReadFile(path)
		if err != nil {
			return err
		}
		for _, line := range strings.Split(string(data), "\n") {
			if strings.TrimSpace(line) != "" {
				count++
			}
		}
		return nil
	})
	return count, err
}

func Pretty(value any) ([]byte, error) {
	data, err := json.MarshalIndent(value, "", "  ")
	if err != nil {
		return nil, err
	}
	return append(data, '\n'), nil
}
