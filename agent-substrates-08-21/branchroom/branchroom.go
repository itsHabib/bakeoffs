package branchroom

import (
	"bytes"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"sort"
)

const (
	CapsuleSchema    = "agent-capsule/v1"
	BranchSchema     = "branch/v1"
	RunEventSchema   = "run-event/v1"
	ExperimentSchema = "experiment-fork/v1"
)

type RepositoryTreeV1 struct {
	Repository string `json:"repository"`
	BaseSHA    string `json:"base_sha"`
	HeadSHA    string `json:"head_sha"`
	DiffDigest string `json:"diff_digest"`
}

type AgentCapsuleV1 struct {
	SchemaVersion              string            `json:"schema_version"`
	InputBundleDigest          string            `json:"input_bundle_digest"`
	VerificationContractDigest string            `json:"verification_contract_digest"`
	RepositoryTree             RepositoryTreeV1  `json:"repository_tree"`
	Instructions               []string          `json:"instructions"`
	ToolSchema                 map[string]any    `json:"tool_schema"`
	Harness                    map[string]string `json:"harness"`
	Environment                map[string]string `json:"environment"`
	TerminalPrefixSequence     uint64            `json:"terminal_prefix_sequence"`
	PrefixDigest               string            `json:"prefix_digest"`
}

type BranchV1 struct {
	SchemaVersion string `json:"schema_version"`
	CapsuleDigest string `json:"capsule_digest"`
	Label         string `json:"branch_label"`
	ParentEpoch   uint64 `json:"parent_epoch"`
	Epoch         uint64 `json:"child_epoch"`
}

type RunEventV1 struct {
	SchemaVersion     string `json:"schema_version"`
	Branch            string `json:"branch"`
	Epoch             uint64 `json:"epoch"`
	Sequence          uint64 `json:"sequence"`
	CallID            string `json:"call_id"`
	Kind              string `json:"event_kind"`
	PayloadDigest     string `json:"payload_digest"`
	ParentEventDigest string `json:"parent_event_digest"`
}

type ReducerState struct {
	branch            BranchV1
	prefixBranch      string
	prefixEpoch       uint64
	nextSequence      uint64
	pendingCallID     string
	parentEventDigest string
}

type Decision struct {
	Accepted bool   `json:"accepted"`
	Reason   string `json:"reason"`
}

func Canonical(value any) ([]byte, error) {
	return json.Marshal(value)
}

func CanonicalJSON(raw []byte) ([]byte, error) {
	decoder := json.NewDecoder(bytes.NewReader(raw))
	decoder.UseNumber()
	var value any
	if err := decoder.Decode(&value); err != nil {
		return nil, err
	}
	var trailing any
	if err := decoder.Decode(&trailing); !errors.Is(err, io.EOF) {
		return nil, errors.New("canonical JSON requires exactly one value")
	}
	return Canonical(value)
}

func DigestBytes(data []byte) string {
	sum := sha256.Sum256(data)
	return "sha256:" + hex.EncodeToString(sum[:])
}

func Digest(value any) (string, error) {
	data, err := Canonical(value)
	if err != nil {
		return "", err
	}
	return DigestBytes(data), nil
}

func EventDigest(event RunEventV1) string {
	digest, err := Digest(event)
	if err != nil {
		panic(err)
	}
	return digest
}

func PrefixDigest(events []RunEventV1) (string, error) {
	if _, err := replayPrefix(events); err != nil {
		return "", err
	}
	return Digest(events)
}

func replayPrefix(events []RunEventV1) (ReducerState, error) {
	state := ReducerState{}
	for index, event := range events {
		next, decision := Reduce(state, event)
		if !decision.Accepted {
			return ReducerState{}, fmt.Errorf("prefix event %d: %s", index+1, decision.Reason)
		}
		state = next
	}
	if len(events) == 0 || state.pendingCallID == "" {
		return ReducerState{}, errors.New("prefix must end with a pending tool call")
	}
	return state, nil
}

func ForkBranches(capsule AgentCapsuleV1, parentEpoch uint64, labels []string) ([]BranchV1, error) {
	capsuleDigest, err := Digest(capsule)
	if err != nil {
		return nil, err
	}
	ordered := append([]string(nil), labels...)
	sort.Strings(ordered)
	branches := make([]BranchV1, 0, len(ordered))
	for index, label := range ordered {
		if label == "" || (index > 0 && label == ordered[index-1]) {
			return nil, errors.New("branch labels must be non-empty and unique")
		}
		epoch := parentEpoch + uint64(index) + 1
		branches = append(branches, BranchV1{
			SchemaVersion: BranchSchema,
			CapsuleDigest: capsuleDigest,
			Label:         label,
			ParentEpoch:   parentEpoch,
			Epoch:         epoch,
		})
	}
	return branches, nil
}

func StartBranch(prefix []RunEventV1, capsule AgentCapsuleV1, branch BranchV1) (ReducerState, error) {
	if branch.SchemaVersion != BranchSchema || branch.Label == "" {
		return ReducerState{}, errors.New("invalid_branch")
	}
	if branch.Epoch <= branch.ParentEpoch {
		return ReducerState{}, errors.New("child_epoch_not_fresh")
	}
	return bindBranch(prefix, capsule, branch)
}

func bindBranch(prefix []RunEventV1, capsule AgentCapsuleV1, branch BranchV1) (ReducerState, error) {
	state, err := replayPrefix(prefix)
	if err != nil {
		return ReducerState{}, err
	}
	prefixDigest, err := Digest(prefix)
	if err != nil {
		return ReducerState{}, err
	}
	capsuleDigest, err := Digest(capsule)
	if err != nil {
		return ReducerState{}, err
	}
	if prefixDigest != capsule.PrefixDigest || branch.CapsuleDigest != capsuleDigest {
		return ReducerState{}, errors.New("capsule_identity_mismatch")
	}
	if branch.ParentEpoch != state.prefixEpoch {
		return ReducerState{}, errors.New("parent_epoch_mismatch")
	}
	state.branch = branch
	return state, nil
}

func Reduce(state ReducerState, event RunEventV1) (ReducerState, Decision) {
	refuse := func(reason string) (ReducerState, Decision) {
		return state, Decision{Accepted: false, Reason: reason}
	}
	if event.SchemaVersion != RunEventSchema {
		return refuse("schema_mismatch")
	}
	if state.branch.SchemaVersion == "" {
		if event.Sequence != state.nextSequence {
			return refuse("sequence_mismatch")
		}
		if event.ParentEventDigest != state.parentEventDigest {
			return refuse("parent_event_mismatch")
		}
		if state.nextSequence > 0 && (event.Branch != state.prefixBranch || event.Epoch != state.prefixEpoch) {
			return refuse("causal_identity_mismatch")
		}
		if state.pendingCallID != "" {
			return refuse("prefix_already_terminal")
		}
		next := state
		if state.nextSequence == 0 {
			next.prefixBranch = event.Branch
			next.prefixEpoch = event.Epoch
		}
		switch event.Kind {
		case "declared_input", "declared_contract", "declared_tree", "declared_instructions", "declared_runner":
			if event.CallID != "" {
				return refuse("unexpected_call_id")
			}
		case "tool_call":
			if event.CallID == "" {
				return refuse("invalid_tool_call")
			}
			next.pendingCallID = event.CallID
		default:
			return refuse("unknown_kind")
		}
		next.nextSequence++
		next.parentEventDigest = EventDigest(event)
		return next, Decision{Accepted: true, Reason: "prefix_continuity"}
	}
	if event.Kind != "tool_terminal" {
		return refuse("invalid_terminal")
	}
	if event.Epoch != state.branch.Epoch {
		if event.Epoch == state.branch.ParentEpoch {
			return refuse("parent_epoch_terminal")
		}
		return refuse("epoch_mismatch")
	}
	if event.Branch != state.branch.Label {
		return refuse("branch_mismatch")
	}
	if event.Sequence != state.nextSequence {
		return refuse("sequence_mismatch")
	}
	if event.ParentEventDigest != state.parentEventDigest {
		return refuse("parent_event_mismatch")
	}
	if state.pendingCallID == "" || event.CallID != state.pendingCallID {
		return refuse("call_id_mismatch")
	}
	next := state
	next.nextSequence++
	next.pendingCallID = ""
	next.parentEventDigest = EventDigest(event)
	return next, Decision{Accepted: true, Reason: "terminal_correlated"}
}
