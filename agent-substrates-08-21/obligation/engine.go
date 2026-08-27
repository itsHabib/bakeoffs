package obligation

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"slices"
	"strconv"
	"strings"
)

const (
	SchemaFrontier = "obligation-frontier/v1"
	ReasonReopened = "mandatory_obligation_reopened"
	ReasonWeakened = "mandatory_contract_weakened"
)

type Subject struct {
	TaskID       string `json:"task_id"`
	TaskRevision int    `json:"task_revision"`
	Repository   string `json:"repository"`
	BaseSHA      string `json:"base_sha"`
	HeadSHA      string `json:"head_sha"`
	DiffDigest   string `json:"diff_digest"`
}

type Claim struct {
	ClaimID  string `json:"claim_id"`
	Required bool   `json:"required"`
}

type ObligationPolicy struct {
	MandatoryClaimIDs        []string `json:"mandatory_claim_ids"`
	AgentsMayAdd             bool     `json:"agents_may_add"`
	AgentsMayRemoveDowngrade bool     `json:"agents_may_remove_or_downgrade"`
}

type Contract struct {
	SchemaVersion    string           `json:"schema_version"`
	ContractID       string           `json:"contract_id"`
	Subject          Subject          `json:"subject"`
	PolicyDigest     string           `json:"policy_digest"`
	Claims           []Claim          `json:"claims"`
	ObligationPolicy ObligationPolicy `json:"obligation_policy"`
}

type Assurance struct {
	CaseID           string            `json:"case_id"`
	SubjectHead      string            `json:"subject_head"`
	ContractID       string            `json:"contract_id"`
	Summary          string            `json:"summary"`
	Coverage         string            `json:"coverage"`
	ClaimOutcomes    map[string]string `json:"claim_outcomes"`
	AcceptedReceipts json.RawMessage   `json:"accepted_receipts"`
	RejectedReceipts json.RawMessage   `json:"rejected_receipts"`
	Gaps             []string          `json:"gaps"`
	MergeAuthority   string            `json:"merge_authority"`
}

type Fixture struct {
	SchemaVersion            string          `json:"schema_version"`
	FixtureID                string          `json:"fixture_id"`
	Contracts                []Contract      `json:"contracts"`
	TrustedRunnerReceipts    json.RawMessage `json:"trusted_runner_receipts"`
	OracleAssuranceSnapshots []Assurance     `json:"oracle_assurance_snapshots"`
}

type GoalIdentity struct {
	TaskID       string `json:"task_id"`
	TaskRevision int    `json:"task_revision"`
	ContractID   string `json:"contract_id"`
	Repository   string `json:"repository"`
	BaseSHA      string `json:"base_sha"`
	HeadSHA      string `json:"head_sha"`
	DiffDigest   string `json:"diff_digest"`
	PolicyDigest string `json:"policy_digest"`
}

func (c Contract) Goal() GoalIdentity {
	return GoalIdentity{
		TaskID: c.Subject.TaskID, TaskRevision: c.Subject.TaskRevision,
		ContractID: c.ContractID, Repository: c.Subject.Repository,
		BaseSHA: c.Subject.BaseSHA, HeadSHA: c.Subject.HeadSHA,
		DiffDigest: c.Subject.DiffDigest, PolicyDigest: c.PolicyDigest,
	}
}

type ObligationKind string

const (
	CollectClaim      ObligationKind = "collect_claim"
	ResolveRefutation ObligationKind = "resolve_refutation"
	RefreshSubject    ObligationKind = "refresh_subject"
)

type Origin string

const (
	OriginContract Origin = "contract"
	OriginDerived  Origin = "derived"
	OriginAgent    Origin = "agent"
)

type State string

const (
	Open       State = "open"
	Discharged State = "discharged"
	Reopened   State = "reopened"
	Superseded State = "superseded"
)

type Transition struct {
	Sequence int    `json:"sequence"`
	State    State  `json:"state"`
	Reason   string `json:"reason"`
}

type Obligation struct {
	ObligationID    string         `json:"obligation_id"`
	Kind            ObligationKind `json:"kind"`
	ClaimID         string         `json:"claim_id"`
	Subject         GoalIdentity   `json:"subject"`
	Origin          Origin         `json:"origin"`
	Mandatory       bool           `json:"mandatory"`
	State           State          `json:"state"`
	Reason          string         `json:"reason"`
	PrerequisiteIDs []string       `json:"prerequisite_ids"`
	Transitions     []Transition   `json:"transitions"`
}

type OverlayAddition struct {
	ClaimID   string `json:"claim_id"`
	Mandatory bool   `json:"mandatory"`
}

type AgentOverlay struct {
	RemoveClaimIDs    []string          `json:"remove_claim_ids,omitempty"`
	DowngradeClaimIDs []string          `json:"downgrade_claim_ids,omitempty"`
	Add               []OverlayAddition `json:"add,omitempty"`
}

type Event struct {
	Type       string          `json:"type"`
	Contract   *Contract       `json:"contract,omitempty"`
	Assurance  *Assurance      `json:"assurance,omitempty"`
	Overlay    *AgentOverlay   `json:"overlay,omitempty"`
	RawReceipt json.RawMessage `json:"raw_receipt,omitempty"`
}

func Activate(c Contract) Event         { return Event{Type: "contract_activated", Contract: &c} }
func Observe(a Assurance) Event         { return Event{Type: "assurance_observed", Assurance: &a} }
func ApplyOverlay(o AgentOverlay) Event { return Event{Type: "agent_overlay_applied", Overlay: &o} }
func SubmitRawReceipt(r json.RawMessage) Event {
	return Event{Type: "raw_receipt_submitted", RawReceipt: r}
}

type Mutant string

const SettledIsTerminal Mutant = "settled-is-terminal"

type Refusal struct{ Reason string }

func (r *Refusal) Error() string { return r.Reason }

func RefusalReason(err error) string {
	var refusal *Refusal
	if errors.As(err, &refusal) {
		return refusal.Reason
	}
	return ""
}

type reducerState struct {
	current   *Contract
	items     map[string]*Obligation
	settled   map[string]bool
	sequence  int
	refreshID string
}

type Frontier struct {
	SchemaVersion string       `json:"schema_version"`
	Goal          GoalIdentity `json:"goal"`
	OpenFrontier  []string     `json:"open_frontier"`
	Obligations   []Obligation `json:"obligations"`
	EventDigest   string       `json:"event_digest"`
}

func Reduce(events []Event, mutant Mutant) (Frontier, error) {
	s := reducerState{items: map[string]*Obligation{}, settled: map[string]bool{}}
	for i, event := range events {
		s.sequence = i + 1
		if err := s.apply(event, mutant); err != nil {
			return Frontier{}, err
		}
	}
	return s.frontier(events), nil
}

func (s *reducerState) apply(event Event, mutant Mutant) error {
	switch event.Type {
	case "contract_activated":
		if event.Contract == nil {
			return &Refusal{Reason: "missing_contract"}
		}
		return s.activate(*event.Contract)
	case "assurance_observed":
		if event.Assurance == nil {
			return &Refusal{Reason: "missing_assurance"}
		}
		return s.observe(*event.Assurance, mutant)
	case "agent_overlay_applied":
		if event.Overlay == nil {
			return &Refusal{Reason: "missing_overlay"}
		}
		return s.overlay(*event.Overlay)
	case "raw_receipt_submitted":
		return &Refusal{Reason: "raw_receipt_input_forbidden"}
	default:
		return &Refusal{Reason: "unknown_event_type"}
	}
}

func (s *reducerState) activate(contract Contract) error {
	mandatory := map[string]bool{}
	for _, id := range contract.ObligationPolicy.MandatoryClaimIDs {
		mandatory[id] = true
	}
	for id := range mandatory {
		found := false
		for _, claim := range contract.Claims {
			if claim.ClaimID == id && claim.Required {
				found = true
			}
		}
		if !found {
			return &Refusal{Reason: "invalid_contract_floor"}
		}
	}

	first := s.current == nil
	if !first && s.current.ContractID == contract.ContractID {
		return &Refusal{Reason: "contract_already_active"}
	}
	if !first {
		for _, item := range s.items {
			if item.Subject != contract.Goal() && item.State != Superseded {
				s.transition(item, Superseded, "subject_superseded")
			}
		}
	}
	s.current = &contract
	if first {
		ids := slices.Clone(contract.ObligationPolicy.MandatoryClaimIDs)
		slices.Sort(ids)
		for _, claimID := range ids {
			s.add(CollectClaim, claimID, contract.Goal(), OriginContract, true, Open, "mandatory_contract_slot_opened", nil)
		}
		return nil
	}
	s.refreshID = s.add(RefreshSubject, "subject-refresh", contract.Goal(), OriginDerived, true, Open, "subject_changed_refresh_required", nil)
	return nil
}

func (s *reducerState) observe(a Assurance, mutant Mutant) error {
	if s.current == nil {
		return &Refusal{Reason: "contract_not_active"}
	}
	if a.ContractID != s.current.ContractID || a.SubjectHead != s.current.Subject.HeadSHA {
		return &Refusal{Reason: "assurance_subject_mismatch"}
	}
	goalKey := identityKey(s.current.Goal())
	if mutant == SettledIsTerminal && s.settled[goalKey] {
		return nil
	}
	if s.refreshID != "" {
		if item := s.items[s.refreshID]; item != nil && item.State == Open {
			s.transition(item, Discharged, "exact_subject_assurance_observed")
		}
	}
	for _, claimID := range s.current.ObligationPolicy.MandatoryClaimIDs {
		if _, present := a.ClaimOutcomes[claimID]; !present {
			s.ensureOpen(CollectClaim, claimID, OriginDerived, Open, "mandatory_oracle_outcome_missing")
		}
	}

	claimIDs := make([]string, 0, len(a.ClaimOutcomes))
	for claimID := range a.ClaimOutcomes {
		claimIDs = append(claimIDs, claimID)
	}
	slices.Sort(claimIDs)
	for _, claimID := range claimIDs {
		outcome := a.ClaimOutcomes[claimID]
		switch outcome {
		case "supported":
			s.dischargeClaim(claimID)
		case "insufficient":
			if !gapNamesClaim(a.Gaps, claimID) {
				return &Refusal{Reason: "oracle_gap_not_named"}
			}
			s.ensureOpen(CollectClaim, claimID, OriginDerived, Open, "oracle_named_gap_opened")
		case "refuted":
			if !gapNamesClaim(a.Gaps, claimID) {
				return &Refusal{Reason: "oracle_gap_not_named"}
			}
			state, reason := Open, "oracle_refutation_opened"
			if s.settled[goalKey] {
				state, reason = Reopened, ReasonReopened
			}
			s.ensureOpen(ResolveRefutation, claimID, OriginDerived, state, reason)
		default:
			return &Refusal{Reason: "unknown_oracle_outcome"}
		}
	}
	if !s.hasOpenMandatory() {
		s.settled[goalKey] = true
	}
	return nil
}

func (s *reducerState) overlay(o AgentOverlay) error {
	if s.current == nil {
		return &Refusal{Reason: "contract_not_active"}
	}
	mandatory := map[string]bool{}
	for _, id := range s.current.ObligationPolicy.MandatoryClaimIDs {
		mandatory[id] = true
	}
	for _, id := range append(slices.Clone(o.RemoveClaimIDs), o.DowngradeClaimIDs...) {
		if mandatory[id] {
			return &Refusal{Reason: ReasonWeakened}
		}
	}
	if len(o.Add) > 0 && !s.current.ObligationPolicy.AgentsMayAdd {
		return &Refusal{Reason: "agent_additions_forbidden"}
	}
	for _, add := range o.Add {
		if mandatory[add.ClaimID] && !add.Mandatory {
			return &Refusal{Reason: ReasonWeakened}
		}
		if add.ClaimID == "" {
			return &Refusal{Reason: "empty_agent_claim"}
		}
		s.ensureOpen(CollectClaim, add.ClaimID, OriginAgent, Open, "agent_overlay_added")
		item := s.items[obligationID(CollectClaim, add.ClaimID, s.current.Goal())]
		item.Mandatory = item.Mandatory || add.Mandatory
	}
	return nil
}

func (s *reducerState) ensureOpen(kind ObligationKind, claimID string, origin Origin, state State, reason string) {
	id := obligationID(kind, claimID, s.current.Goal())
	if item := s.items[id]; item != nil {
		if item.State == Discharged {
			s.transition(item, state, reason)
		}
		return
	}
	prerequisites := []string(nil)
	if s.refreshID != "" && kind != RefreshSubject {
		prerequisites = []string{s.refreshID}
	}
	mandatory := false
	for _, required := range s.current.ObligationPolicy.MandatoryClaimIDs {
		mandatory = mandatory || required == claimID
	}
	s.add(kind, claimID, s.current.Goal(), origin, mandatory, state, reason, prerequisites)
}

func (s *reducerState) dischargeClaim(claimID string) {
	for _, item := range s.items {
		if item.Subject == s.current.Goal() && item.ClaimID == claimID && (item.State == Open || item.State == Reopened) {
			s.transition(item, Discharged, "oracle_claim_supported")
		}
	}
}

func (s *reducerState) hasOpenMandatory() bool {
	for _, item := range s.items {
		if item.Mandatory && (item.State == Open || item.State == Reopened) {
			return true
		}
	}
	return false
}

func (s *reducerState) add(kind ObligationKind, claimID string, subject GoalIdentity, origin Origin, mandatory bool, state State, reason string, prerequisites []string) string {
	id := obligationID(kind, claimID, subject)
	item := &Obligation{
		ObligationID: id, Kind: kind, ClaimID: claimID, Subject: subject,
		Origin: origin, Mandatory: mandatory, State: state, Reason: reason,
		PrerequisiteIDs: slices.Clone(prerequisites),
		Transitions:     []Transition{{Sequence: s.sequence, State: state, Reason: reason}},
	}
	s.items[id] = item
	return id
}

func (s *reducerState) transition(item *Obligation, state State, reason string) {
	item.State, item.Reason = state, reason
	item.Transitions = append(item.Transitions, Transition{Sequence: s.sequence, State: state, Reason: reason})
}

func (s *reducerState) frontier(events []Event) Frontier {
	items := make([]Obligation, 0, len(s.items))
	open := []string{}
	for _, item := range s.items {
		copy := *item
		copy.PrerequisiteIDs = slices.Clone(item.PrerequisiteIDs)
		copy.Transitions = slices.Clone(item.Transitions)
		items = append(items, copy)
		if item.State == Open || item.State == Reopened {
			open = append(open, item.ObligationID)
		}
	}
	slices.SortFunc(items, func(a, b Obligation) int { return strings.Compare(a.ObligationID, b.ObligationID) })
	slices.Sort(open)
	goal := GoalIdentity{}
	if s.current != nil {
		goal = s.current.Goal()
	}
	return Frontier{SchemaVersion: SchemaFrontier, Goal: goal, OpenFrontier: open, Obligations: items, EventDigest: digestJSON(events)}
}

func obligationID(kind ObligationKind, claimID string, subject GoalIdentity) string {
	value := strings.Join([]string{string(kind), claimID, identityKey(subject)}, "|")
	sum := sha256.Sum256([]byte(value))
	return "obl_" + hex.EncodeToString(sum[:10])
}

func identityKey(g GoalIdentity) string {
	return strings.Join([]string{g.TaskID, strconv.Itoa(g.TaskRevision), g.ContractID, g.Repository, g.BaseSHA, g.HeadSHA, g.DiffDigest, g.PolicyDigest}, "|")
}

func gapNamesClaim(gaps []string, claimID string) bool {
	for _, gap := range gaps {
		if strings.HasPrefix(gap, claimID+" ") {
			return true
		}
	}
	return false
}

func digestJSON(value any) string {
	data, err := json.Marshal(value)
	if err != nil {
		panic(fmt.Sprintf("marshal deterministic artifact: %v", err))
	}
	sum := sha256.Sum256(data)
	return "sha256:" + hex.EncodeToString(sum[:])
}

func DigestBytes(data []byte) string {
	sum := sha256.Sum256(data)
	return "sha256:" + hex.EncodeToString(sum[:])
}

func ArtifactDigest(frontier Frontier) string { return digestJSON(frontier) }
