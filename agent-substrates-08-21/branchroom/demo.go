package branchroom

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"os"
)

type fixtureDeck struct {
	FixtureID string            `json:"fixture_id"`
	Source    fixtureSource     `json:"source"`
	Contracts []json.RawMessage `json:"contracts"`
	Attacks   fixtureAttacks    `json:"attacks"`
}

type fixtureSource struct {
	TaskID          string `json:"task_id"`
	TaskRevision    int    `json:"task_revision"`
	Repository      string `json:"repository"`
	BaseSHA         string `json:"base_sha"`
	InitialHeadSHA  string `json:"initial_head_sha"`
	RepairedHeadSHA string `json:"repaired_head_sha"`
}

type fixtureAttacks struct {
	LateParent struct {
		ExpectedReason string `json:"expected_reason"`
	} `json:"branchroom_late_parent_terminal"`
}

type contractIdentity struct {
	ContractID string `json:"contract_id"`
	Subject    struct {
		BaseSHA    string `json:"base_sha"`
		HeadSHA    string `json:"head_sha"`
		DiffDigest string `json:"diff_digest"`
	} `json:"subject"`
	Environment map[string]string `json:"environment_constraints"`
	Claims      []struct {
		AcceptedMethodClasses    []string `json:"accepted_method_classes"`
		RequiredNegativeControls []string `json:"required_negative_controls"`
	} `json:"claims"`
}

type BranchResultV1 struct {
	Branch              BranchV1 `json:"branch"`
	BranchDigest        string   `json:"branch_digest"`
	TerminalEventDigest string   `json:"terminal_event_digest"`
	OwnTerminal         Decision `json:"own_terminal"`
}

type DivergenceV1 struct {
	EventOrdinal int    `json:"event_ordinal"`
	Sequence     uint64 `json:"sequence"`
	Reason       string `json:"reason"`
}

type RejectedTerminalV1 struct {
	Epoch         uint64              `json:"epoch"`
	Sequence      uint64              `json:"sequence"`
	CallID        string              `json:"call_id"`
	PayloadDigest string              `json:"payload_digest"`
	EventDigests  map[string]string   `json:"event_digests"`
	RejectedBy    map[string]Decision `json:"rejected_by"`
}

type PerturbationV1 struct {
	Field          string `json:"field"`
	Control        string `json:"control"`
	Counterfactual string `json:"counterfactual"`
}

type ExperimentForkV1 struct {
	SchemaVersion        string             `json:"schema_version"`
	CapsuleDigest        string             `json:"capsule_digest"`
	PrefixDigest         string             `json:"prefix_digest"`
	DeclaredPerturbation PerturbationV1     `json:"declared_perturbation"`
	Branches             []BranchResultV1   `json:"branches"`
	FirstDivergence      DivergenceV1       `json:"first_divergence"`
	RejectedTerminal     RejectedTerminalV1 `json:"rejected_terminal"`
	Limitations          []string           `json:"limitations"`
}

type ExactSubjectV1 struct {
	TaskID       string `json:"task_id"`
	TaskRevision int    `json:"task_revision"`
	Repository   string `json:"repository"`
	BaseSHA      string `json:"base_sha"`
	HeadSHA      string `json:"head_sha"`
	DiffDigest   string `json:"diff_digest"`
}

type NormalizedEnvelopeV1 struct {
	SchemaVersion        string         `json:"schema_version"`
	Entry                string         `json:"entry"`
	CaseID               string         `json:"case_id"`
	InputBundleDigest    string         `json:"input_bundle_digest"`
	ExactSubject         ExactSubjectV1 `json:"exact_subject"`
	NativeStatus         string         `json:"native_status"`
	ReasonCodes          []string       `json:"reason_codes"`
	NativeArtifactDigest string         `json:"native_artifact_digest"`
}

type DemoResult struct {
	Capsule        AgentCapsuleV1
	Experiment     ExperimentForkV1
	ArtifactDigest string
	Envelope       NormalizedEnvelopeV1
	MutantDecision Decision
	MutantEventID  string
}

func BuildDemo(fixturePath string) (DemoResult, error) {
	raw, err := os.ReadFile(fixturePath)
	if err != nil {
		return DemoResult{}, err
	}
	var deck fixtureDeck
	if err := json.Unmarshal(raw, &deck); err != nil {
		return DemoResult{}, err
	}
	if len(deck.Contracts) < 2 {
		return DemoResult{}, fmt.Errorf("fixture requires repaired-head contract")
	}
	var contract contractIdentity
	if err := json.Unmarshal(deck.Contracts[1], &contract); err != nil {
		return DemoResult{}, err
	}
	method, negativeControl, err := fixtureMethod(contract)
	if err != nil {
		return DemoResult{}, err
	}
	canonicalContract, err := CanonicalJSON(deck.Contracts[1])
	if err != nil {
		return DemoResult{}, err
	}
	prefix := buildPrefix(raw, deck, contract, method)
	prefixDigest, err := PrefixDigest(prefix)
	if err != nil {
		return DemoResult{}, err
	}
	capsule := AgentCapsuleV1{
		SchemaVersion:              CapsuleSchema,
		InputBundleDigest:          DigestBytes(raw),
		VerificationContractDigest: DigestBytes(canonicalContract),
		RepositoryTree: RepositoryTreeV1{
			Repository: deck.Source.Repository,
			BaseSHA:    contract.Subject.BaseSHA,
			HeadSHA:    contract.Subject.HeadSHA,
			DiffDigest: contract.Subject.DiffDigest,
		},
		Instructions: []string{
			"execute one trusted fixture verification method",
			"record the negative-control mode as a declared input",
		},
		ToolSchema: map[string]any{
			"name":  "trusted_fixture.verify",
			"input": map[string]any{"negative_control_mode": "string"},
		},
		Harness: map[string]string{
			"runner_class":     "trusted-fixture-runner",
			"method_class":     method,
			"negative_control": negativeControl,
		},
		Environment:            contract.Environment,
		TerminalPrefixSequence: prefix[len(prefix)-1].Sequence,
		PrefixDigest:           prefixDigest,
	}
	branches, err := ForkBranches(capsule, 12, []string{"counterfactual", "control"})
	if err != nil {
		return DemoResult{}, err
	}
	mutantBranches := retainParentEpochMutant(branches)

	results := make([]BranchResultV1, 0, 2)
	rejections := make(map[string]Decision, 2)
	lateEventDigests := make(map[string]string, 2)
	ownEvents := make(map[string]RunEventV1, 2)
	var mutantDecision Decision
	var mutantEventID string
	for index, branch := range branches {
		state, err := StartBranch(prefix, capsule, branch)
		if err != nil {
			return DemoResult{}, err
		}
		mode, observed := "off", "exit 0"
		if branch.Label == "counterfactual" {
			mode, observed = negativeControl, "compile failure detected"
		}
		own := terminal(branch.Label, branch.Epoch, state, mode, observed)
		_, ownDecision := Reduce(state, own)
		late := terminal(branch.Label, branch.ParentEpoch, state, "off", "exit 0")
		_, lateDecision := Reduce(state, late)
		if ownDecision.Accepted == false || lateDecision.Reason != deck.Attacks.LateParent.ExpectedReason {
			return DemoResult{}, fmt.Errorf("branch %s violated the causal epoch law", branch.Label)
		}
		branchDigest, err := Digest(branch)
		if err != nil {
			return DemoResult{}, err
		}
		results = append(results, BranchResultV1{Branch: branch, BranchDigest: branchDigest, TerminalEventDigest: EventDigest(own), OwnTerminal: ownDecision})
		rejections[branch.Label] = lateDecision
		lateEventDigests[branch.Label] = EventDigest(late)
		ownEvents[branch.Label] = own
		if branch.Label == "control" {
			mutantState, err := startRetainParentEpochMutant(prefix, capsule, mutantBranches[index])
			if err != nil {
				return DemoResult{}, err
			}
			_, mutantDecision = Reduce(mutantState, late)
			mutantEventID = EventDigest(late)
		}
	}
	if !mutantDecision.Accepted {
		return DemoResult{}, errors.New("retain-parent-epoch mutant did not accept late parent terminal")
	}
	capsuleDigest, err := Digest(capsule)
	if err != nil {
		return DemoResult{}, err
	}
	controlTrace := append(append([]RunEventV1(nil), prefix...), ownEvents["control"])
	counterfactualTrace := append(append([]RunEventV1(nil), prefix...), ownEvents["counterfactual"])
	firstDivergence, err := FirstMeaningfulDivergence(controlTrace, counterfactualTrace)
	if err != nil {
		return DemoResult{}, err
	}
	experiment := ExperimentForkV1{
		SchemaVersion: ExperimentSchema,
		CapsuleDigest: capsuleDigest,
		PrefixDigest:  prefixDigest,
		DeclaredPerturbation: PerturbationV1{
			Field:          "tool_input.negative_control_mode",
			Control:        "off",
			Counterfactual: negativeControl,
		},
		Branches:        results,
		FirstDivergence: firstDivergence,
		RejectedTerminal: RejectedTerminalV1{
			Epoch: 12, Sequence: 6, CallID: "call:verify-negative-control",
			PayloadDigest: payloadDigest(map[string]string{"negative_control_mode": "off", "observed": "exit 0"}),
			EventDigests:  lateEventDigests,
			RejectedBy:    rejections,
		},
		Limitations: []string{
			"captures declared harness-visible state only; no hidden model or process state",
			"runner-capturable experiment artifact; not evidence, assurance, or authority",
		},
	}
	artifactDigest, err := Digest(experiment)
	if err != nil {
		return DemoResult{}, err
	}
	envelope := NormalizedEnvelopeV1{
		SchemaVersion:     "agent-substrate-bakeoff-envelope/v1",
		Entry:             "hack-branchroom",
		CaseID:            deck.FixtureID + "/late-parent-terminal",
		InputBundleDigest: DigestBytes(raw),
		ExactSubject: ExactSubjectV1{
			TaskID: deck.Source.TaskID, TaskRevision: deck.Source.TaskRevision,
			Repository: deck.Source.Repository, BaseSHA: contract.Subject.BaseSHA,
			HeadSHA: contract.Subject.HeadSHA, DiffDigest: contract.Subject.DiffDigest,
		},
		NativeStatus:         "causal_separation_enforced",
		ReasonCodes:          []string{"own_terminal_correlated", "parent_epoch_terminal"},
		NativeArtifactDigest: artifactDigest,
	}
	return DemoResult{Capsule: capsule, Experiment: experiment, ArtifactDigest: artifactDigest, Envelope: envelope, MutantDecision: mutantDecision, MutantEventID: mutantEventID}, nil
}

func retainParentEpochMutant(branches []BranchV1) []BranchV1 {
	mutant := append([]BranchV1(nil), branches...)
	for index := range mutant {
		mutant[index].Epoch = mutant[index].ParentEpoch
	}
	return mutant
}

func startRetainParentEpochMutant(prefix []RunEventV1, capsule AgentCapsuleV1, branch BranchV1) (ReducerState, error) {
	if branch.Epoch != branch.ParentEpoch {
		return ReducerState{}, errors.New("retain-parent-epoch mutant requires the parent epoch")
	}
	return bindBranch(prefix, capsule, branch)
}

func fixtureMethod(contract contractIdentity) (string, string, error) {
	for _, claim := range contract.Claims {
		if len(claim.AcceptedMethodClasses) > 0 && len(claim.RequiredNegativeControls) > 0 {
			return claim.AcceptedMethodClasses[0], claim.RequiredNegativeControls[0], nil
		}
	}
	return "", "", errors.New("contract has no verification method with a negative control")
}

func FirstMeaningfulDivergence(control, counterfactual []RunEventV1) (DivergenceV1, error) {
	limit := len(control)
	if len(counterfactual) < limit {
		limit = len(counterfactual)
	}
	for index := range limit {
		left, right := control[index], counterfactual[index]
		if left.Sequence != right.Sequence || left.Kind != right.Kind || left.CallID != right.CallID {
			return DivergenceV1{EventOrdinal: index + 1, Sequence: left.Sequence, Reason: "event_shape_changed"}, nil
		}
		if left.PayloadDigest != right.PayloadDigest {
			return DivergenceV1{EventOrdinal: index + 1, Sequence: left.Sequence, Reason: "declared_input_changed"}, nil
		}
	}
	if len(control) != len(counterfactual) {
		return DivergenceV1{EventOrdinal: limit + 1, Sequence: uint64(limit), Reason: "trace_length_changed"}, nil
	}
	return DivergenceV1{}, errors.New("no meaningful divergence")
}

func buildPrefix(raw []byte, deck fixtureDeck, contract contractIdentity, method string) []RunEventV1 {
	items := []struct {
		kind    string
		callID  string
		payload any
	}{
		{"declared_input", "", map[string]string{"input_bundle_digest": DigestBytes(raw)}},
		{"declared_contract", "", map[string]string{"contract_id": contract.ContractID}},
		{"declared_tree", "", map[string]string{"head_sha": contract.Subject.HeadSHA}},
		{"declared_instructions", "", []string{"execute one trusted fixture verification method"}},
		{"declared_runner", "", map[string]string{"runner_class": contract.Environment["runner_class"]}},
		{"tool_call", "call:verify-negative-control", map[string]string{"method_class": method, "negative_control_mode": "declared-at-branch"}},
	}
	events := make([]RunEventV1, 0, len(items))
	parent := ""
	for sequence, item := range items {
		event := RunEventV1{
			SchemaVersion: RunEventSchema, Branch: "root", Epoch: 12,
			Sequence: uint64(sequence), CallID: item.callID, Kind: item.kind,
			PayloadDigest: payloadDigest(item.payload), ParentEventDigest: parent,
		}
		events = append(events, event)
		parent = EventDigest(event)
	}
	return events
}

func terminal(branch string, epoch uint64, state ReducerState, mode, observed string) RunEventV1 {
	return RunEventV1{
		SchemaVersion: RunEventSchema, Branch: branch, Epoch: epoch,
		Sequence: state.nextSequence, CallID: state.pendingCallID, Kind: "tool_terminal",
		PayloadDigest:     payloadDigest(map[string]string{"negative_control_mode": mode, "observed": observed}),
		ParentEventDigest: state.parentEventDigest,
	}
}

func payloadDigest(value any) string {
	digest, err := Digest(value)
	if err != nil {
		panic(err)
	}
	return digest
}

func ByteStableDemo(fixturePath string) (DemoResult, error) {
	first, err := BuildDemo(fixturePath)
	if err != nil {
		return DemoResult{}, err
	}
	second, err := BuildDemo(fixturePath)
	if err != nil {
		return DemoResult{}, err
	}
	firstBytes, _ := Canonical(first.Envelope)
	secondBytes, _ := Canonical(second.Envelope)
	if !bytes.Equal(firstBytes, secondBytes) {
		return DemoResult{}, errors.New("normalized envelope is not byte stable")
	}
	return first, nil
}
