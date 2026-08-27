package branchroom

import (
	"bytes"
	"encoding/json"
	"os"
	"path/filepath"
	"testing"
)

const fixturePath = "fixtures/exact-head-lifecycle-v1.json"

func TestCanonicalCapsuleBytesAndIDAreStable(t *testing.T) {
	result, err := BuildDemo(fixturePath)
	if err != nil {
		t.Fatal(err)
	}
	firstBytes, err := Canonical(result.Capsule)
	if err != nil {
		t.Fatal(err)
	}
	secondBytes, err := Canonical(result.Capsule)
	if err != nil {
		t.Fatal(err)
	}
	if !bytes.Equal(firstBytes, secondBytes) {
		t.Fatal("canonical capsule bytes changed")
	}
	firstID, _ := Digest(result.Capsule)
	secondID, _ := Digest(result.Capsule)
	if firstID != secondID || firstID != result.Experiment.CapsuleDigest {
		t.Fatalf("unstable capsule id: %q %q", firstID, secondID)
	}

	left, _ := Canonical(map[string]any{"b": 2, "a": 1})
	right, _ := Canonical(map[string]any{"a": 1, "b": 2})
	if !bytes.Equal(left, right) || string(left) != `{"a":1,"b":2}` {
		t.Fatalf("map canonicalization is not stable: %s vs %s", left, right)
	}
}

func TestBranchCreationOrderDoesNotAffectIDs(t *testing.T) {
	result := mustDemo(t)
	left, err := ForkBranches(result.Capsule, 12, []string{"control", "counterfactual"})
	if err != nil {
		t.Fatal(err)
	}
	right, err := ForkBranches(result.Capsule, 12, []string{"counterfactual", "control"})
	if err != nil {
		t.Fatal(err)
	}
	leftBytes, _ := Canonical(left)
	rightBytes, _ := Canonical(right)
	if !bytes.Equal(leftBytes, rightBytes) {
		t.Fatalf("branch order changed identities:\n%s\n%s", leftBytes, rightBytes)
	}
	if left[0].Epoch != 13 || left[1].Epoch != 14 {
		t.Fatalf("unexpected deterministic epochs: %+v", left)
	}
}

func TestPublicStartRejectsNonFreshChildEpoch(t *testing.T) {
	capsule, prefix, branches := demoState(t)
	for _, epoch := range []uint64{branches[0].ParentEpoch, branches[0].ParentEpoch - 1} {
		forged := branches[0]
		forged.Epoch = epoch
		if _, err := StartBranch(prefix, capsule, forged); err == nil || err.Error() != "child_epoch_not_fresh" {
			t.Fatalf("StartBranch accepted child epoch %d: %v", epoch, err)
		}
	}
}

func TestBranchesSharePrefixAcceptOwnAndRejectParent(t *testing.T) {
	result := mustDemo(t)
	if result.Experiment.PrefixDigest != result.Capsule.PrefixDigest {
		t.Fatal("experiment and capsule prefix identities differ")
	}
	for _, branch := range result.Experiment.Branches {
		if branch.Branch.CapsuleDigest != result.Experiment.CapsuleDigest {
			t.Fatalf("%s does not retain capsule identity", branch.Branch.Label)
		}
		if !branch.OwnTerminal.Accepted {
			t.Fatalf("%s refused its own terminal: %+v", branch.Branch.Label, branch.OwnTerminal)
		}
		decision := result.Experiment.RejectedTerminal.RejectedBy[branch.Branch.Label]
		if decision.Accepted || decision.Reason != "parent_epoch_terminal" {
			t.Fatalf("%s accepted late parent terminal: %+v", branch.Branch.Label, decision)
		}
	}
	if !result.MutantDecision.Accepted {
		t.Fatalf("single-law mutant did not expose planted bug: %+v", result.MutantDecision)
	}
	if result.MutantEventID != result.Experiment.RejectedTerminal.EventDigests["control"] {
		t.Fatal("mutant and production did not consume the same late event bytes")
	}
	if result.Experiment.FirstDivergence.EventOrdinal != 7 || result.Experiment.FirstDivergence.Sequence != 6 {
		t.Fatalf("wrong golden divergence: %+v", result.Experiment.FirstDivergence)
	}
}

func TestFirstDivergenceIsComputedFromPayloadNotCausalLabels(t *testing.T) {
	capsule, prefix, branches := demoState(t)
	controlState, err := StartBranch(prefix, capsule, branches[0])
	if err != nil {
		t.Fatal(err)
	}
	counterState, err := StartBranch(prefix, capsule, branches[1])
	if err != nil {
		t.Fatal(err)
	}
	control := terminal(branches[0].Label, branches[0].Epoch, controlState, "off", "exit 0")
	samePayload := terminal(branches[1].Label, branches[1].Epoch, counterState, "off", "exit 0")
	if _, err := FirstMeaningfulDivergence(append(prefix, control), append(prefix, samePayload)); err == nil {
		t.Fatal("branch and epoch labels alone were reported as meaningful divergence")
	}
	changed := terminal(branches[1].Label, branches[1].Epoch, counterState, "compile-error-fixture", "compile failure detected")
	got, err := FirstMeaningfulDivergence(append(prefix, control), append(prefix, changed))
	if err != nil {
		t.Fatal(err)
	}
	if got.EventOrdinal != 7 || got.Sequence != 6 || got.Reason != "declared_input_changed" {
		t.Fatalf("wrong computed divergence: %+v", got)
	}
}

func TestCapsuleDigestCommitsToDeclaredExecutionState(t *testing.T) {
	original := mustDemo(t).Capsule
	want, _ := Digest(original)
	tests := map[string]func(*AgentCapsuleV1){
		"instructions": func(c *AgentCapsuleV1) { c.Instructions[0] = "changed" },
		"tool schema":  func(c *AgentCapsuleV1) { c.ToolSchema["name"] = "changed" },
		"harness":      func(c *AgentCapsuleV1) { c.Harness["method_class"] = "changed" },
		"environment":  func(c *AgentCapsuleV1) { c.Environment["network"] = "enabled" },
	}
	for name, mutate := range tests {
		t.Run(name, func(t *testing.T) {
			changed := cloneCapsule(t, original)
			mutate(&changed)
			got, _ := Digest(changed)
			if got == want {
				t.Fatalf("%s change did not alter capsule digest", name)
			}
		})
	}
}

func TestDuplicateAndGappedSequencesFailClosed(t *testing.T) {
	capsule, prefix, branches := demoState(t)
	state, err := StartBranch(prefix, capsule, branches[0])
	if err != nil {
		t.Fatal(err)
	}
	event := terminal(branches[0].Label, branches[0].Epoch, state, "off", "exit 0")
	next, accepted := Reduce(state, event)
	if !accepted.Accepted {
		t.Fatalf("positive terminal refused: %+v", accepted)
	}
	_, duplicate := Reduce(next, event)
	if duplicate.Accepted || duplicate.Reason != "sequence_mismatch" {
		t.Fatalf("duplicate did not fail closed: %+v", duplicate)
	}
	gapped := event
	gapped.Sequence++
	_, gap := Reduce(state, gapped)
	if gap.Accepted || gap.Reason != "sequence_mismatch" {
		t.Fatalf("gap did not fail closed: %+v", gap)
	}

	brokenPrefix := append([]RunEventV1(nil), prefix...)
	brokenPrefix[3].Sequence = 2
	if _, err := PrefixDigest(brokenPrefix); err == nil {
		t.Fatal("duplicate prefix sequence accepted")
	}
	brokenPrefix = append([]RunEventV1(nil), prefix...)
	brokenPrefix[3].Sequence = 4
	if _, err := PrefixDigest(brokenPrefix); err == nil {
		t.Fatal("gapped prefix sequence accepted")
	}
}

func TestTerminalMustMatchEveryCorrelationField(t *testing.T) {
	capsule, prefix, branches := demoState(t)
	state, err := StartBranch(prefix, capsule, branches[0])
	if err != nil {
		t.Fatal(err)
	}
	valid := terminal(branches[0].Label, branches[0].Epoch, state, "off", "exit 0")
	tests := map[string]struct {
		mutate func(*RunEventV1)
		reason string
	}{
		"branch":       {func(e *RunEventV1) { e.Branch = "other" }, "branch_mismatch"},
		"epoch":        {func(e *RunEventV1) { e.Epoch++ }, "epoch_mismatch"},
		"call":         {func(e *RunEventV1) { e.CallID = "call:other" }, "call_id_mismatch"},
		"parent event": {func(e *RunEventV1) { e.ParentEventDigest = "sha256:stale" }, "parent_event_mismatch"},
	}
	for name, test := range tests {
		t.Run(name, func(t *testing.T) {
			candidate := valid
			test.mutate(&candidate)
			_, decision := Reduce(state, candidate)
			if decision.Accepted || decision.Reason != test.reason {
				t.Fatalf("got %+v, want refusal %s", decision, test.reason)
			}
		})
	}
}

func TestNormalizedEnvelopeMatchesGolden(t *testing.T) {
	result, err := ByteStableDemo(fixturePath)
	if err != nil {
		t.Fatal(err)
	}
	got, err := json.MarshalIndent(result.Envelope, "", "  ")
	if err != nil {
		t.Fatal(err)
	}
	got = append(got, '\n')
	want, err := os.ReadFile(filepath.Join("testdata", "normalized-envelope.golden.json"))
	if err != nil {
		t.Fatal(err)
	}
	if !bytes.Equal(got, want) {
		t.Fatalf("normalized envelope changed\n--- got ---\n%s\n--- want ---\n%s", got, want)
	}
}

func mustDemo(t *testing.T) DemoResult {
	t.Helper()
	result, err := BuildDemo(fixturePath)
	if err != nil {
		t.Fatal(err)
	}
	return result
}

func demoState(t *testing.T) (AgentCapsuleV1, []RunEventV1, []BranchV1) {
	t.Helper()
	raw, err := os.ReadFile(fixturePath)
	if err != nil {
		t.Fatal(err)
	}
	var deck fixtureDeck
	if err := json.Unmarshal(raw, &deck); err != nil {
		t.Fatal(err)
	}
	var contract contractIdentity
	if err := json.Unmarshal(deck.Contracts[1], &contract); err != nil {
		t.Fatal(err)
	}
	result := mustDemo(t)
	method, _, err := fixtureMethod(contract)
	if err != nil {
		t.Fatal(err)
	}
	prefix := buildPrefix(raw, deck, contract, method)
	branches, err := ForkBranches(result.Capsule, 12, []string{"control", "counterfactual"})
	if err != nil {
		t.Fatal(err)
	}
	return result.Capsule, prefix, branches
}

func cloneCapsule(t *testing.T, capsule AgentCapsuleV1) AgentCapsuleV1 {
	t.Helper()
	data, err := json.Marshal(capsule)
	if err != nil {
		t.Fatal(err)
	}
	var clone AgentCapsuleV1
	if err := json.Unmarshal(data, &clone); err != nil {
		t.Fatal(err)
	}
	return clone
}
