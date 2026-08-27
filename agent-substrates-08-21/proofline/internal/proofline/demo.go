package proofline

import (
	"encoding/json"
	"fmt"
	"os"
	"reflect"
	"sort"
)

type Subject struct {
	TaskID       string `json:"task_id"`
	TaskRevision int    `json:"task_revision"`
	Repository   string `json:"repository"`
	BaseSHA      string `json:"base_sha"`
	HeadSHA      string `json:"head_sha"`
	DiffDigest   string `json:"diff_digest"`
}
type Contract struct {
	ContractID   string  `json:"contract_id"`
	Subject      Subject `json:"subject"`
	PolicyDigest string  `json:"policy_digest"`
}
type Receipt struct {
	ReceiptID   string          `json:"receipt_id"`
	ContractID  string          `json:"contract_id"`
	SubjectHead string          `json:"subject_head"`
	Raw         json.RawMessage `json:"-"`
}
type Oracle struct {
	CaseID           string          `json:"case_id"`
	SubjectHead      string          `json:"subject_head"`
	ContractID       string          `json:"contract_id"`
	Summary          string          `json:"summary"`
	Coverage         string          `json:"coverage"`
	AcceptedReceipts []string        `json:"accepted_receipts"`
	Gaps             []string        `json:"gaps"`
	MergeAuthority   string          `json:"merge_authority"`
	Raw              json.RawMessage `json:"-"`
}
type Fixture struct {
	SchemaVersion string `json:"schema_version"`
	Source        struct {
		InitialHeadSHA  string `json:"initial_head_sha"`
		RepairedHeadSHA string `json:"repaired_head_sha"`
	} `json:"source"`
	Contracts             []json.RawMessage `json:"contracts"`
	TrustedRunnerReceipts []json.RawMessage `json:"trusted_runner_receipts"`
	OracleSnapshots       []json.RawMessage `json:"oracle_assurance_snapshots"`
	Attacks               map[string]struct {
		Mutation       string `json:"mutation"`
		ExpectedReason string `json:"expected_reason"`
	} `json:"attacks"`
}
type Model struct {
	Graph                               *Graph
	InputDigest                         string
	OracleForAssurance                  map[string]Oracle
	InitialContract, RepairedContract   string
	InitialAssurance, RepairedAssurance string
	AttackID, DriftID                   string
}

func Load(path string) (*Model, error) {
	b, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}
	return loadBytes(b)
}
func loadBytes(b []byte) (*Model, error) {
	var f Fixture
	if err := json.Unmarshal(b, &f); err != nil {
		return nil, err
	}
	if f.SchemaVersion != "agent-substrate-bakeoff-fixture/v1" {
		return nil, fmt.Errorf("unsupported_fixture:%s", f.SchemaVersion)
	}
	m := &Model{InputDigest: digest(b), OracleForAssurance: map[string]Oracle{}}
	contracts := map[string]Contract{}
	oracles := map[string]Oracle{}
	contractRaw := map[string]json.RawMessage{}
	for _, raw := range f.Contracts {
		var c Contract
		if err := json.Unmarshal(raw, &c); err != nil {
			return nil, err
		}
		if prior, ok := contractRaw[c.ContractID]; ok && !equalRaw(prior, raw) {
			return nil, fmt.Errorf("duplicate_contract_conflict:%s", c.ContractID)
		}
		if _, ok := contracts[c.ContractID]; ok {
			continue
		}
		contracts[c.ContractID], contractRaw[c.ContractID] = c, raw
	}
	receipts := map[string]Receipt{}
	for _, raw := range f.TrustedRunnerReceipts {
		var r Receipt
		if err := json.Unmarshal(raw, &r); err != nil {
			return nil, err
		}
		r.Raw = raw
		if prior, ok := receipts[r.ReceiptID]; ok && !equalRaw(prior.Raw, raw) {
			return nil, fmt.Errorf("duplicate_receipt_conflict:%s", r.ReceiptID)
		}
		if _, ok := receipts[r.ReceiptID]; ok {
			continue
		}
		receipts[r.ReceiptID] = r
	}
	var ordered []Oracle
	for _, raw := range f.OracleSnapshots {
		var o Oracle
		if err := json.Unmarshal(raw, &o); err != nil {
			return nil, err
		}
		o.Raw = raw
		if prior, ok := oracles[o.CaseID]; ok && !equalRaw(prior.Raw, raw) {
			return nil, fmt.Errorf("duplicate_assurance_conflict:%s", o.CaseID)
		}
		if _, ok := oracles[o.CaseID]; ok {
			continue
		}
		oracles[o.CaseID] = o
		ordered = append(ordered, o)
	}
	byHead := map[string]Contract{}
	for _, c := range contracts {
		byHead[c.Subject.HeadSHA] = c
	}
	initial, initialOK := byHead[f.Source.InitialHeadSHA]
	repaired, repairedOK := byHead[f.Source.RepairedHeadSHA]
	if !initialOK || !repairedOK || repaired.Subject.BaseSHA != initial.Subject.HeadSHA {
		return nil, fmt.Errorf("missing_revision_contract")
	}
	m.InitialContract, m.RepairedContract = initial.ContractID, repaired.ContractID
	var initialOracle, repairedOracle Oracle
	for _, o := range ordered {
		if o.SubjectHead == initial.Subject.HeadSHA && initialOracle.CaseID == "" {
			initialOracle = o
		}
		if o.SubjectHead == repaired.Subject.HeadSHA {
			repairedOracle = o
		}
	}
	if initialOracle.CaseID == "" || repairedOracle.CaseID == "" {
		return nil, fmt.Errorf("missing_lifecycle_assurance")
	}
	selected := []Oracle{initialOracle, repairedOracle}
	m.InitialAssurance = "assurance:" + initialOracle.CaseID
	m.RepairedAssurance = "assurance:" + repairedOracle.CaseID
	var nodes []Node
	var edges []Edge
	for _, o := range selected {
		caseID := o.CaseID
		c, ok := contracts[o.ContractID]
		if !ok {
			return nil, fmt.Errorf("oracle_contract_missing:%s", caseID)
		}
		if o.SubjectHead != c.Subject.HeadSHA {
			return nil, fmt.Errorf("oracle_identity_mismatch:%s", caseID)
		}
		assuranceID := "assurance:" + caseID
		m.OracleForAssurance[assuranceID] = o
		base := identity(c)
		nodes = append(nodes, node(c.ContractID, "verification_contract", base, contractRaw[c.ContractID]))
		for _, receiptID := range o.AcceptedReceipts {
			r, ok := receipts[receiptID]
			if !ok || r.ContractID != c.ContractID || r.SubjectHead != c.Subject.HeadSHA {
				return nil, fmt.Errorf("trusted_receipt_identity_mismatch:%s", receiptID)
			}
			nodes = append(nodes, node(receiptID, "trusted_receipt", base, r.Raw))
			edges = append(edges,
				edge("evaluates", c.ContractID, receiptID),
				edge("includes", receiptID, assuranceID),
			)
		}
		nodes = append(nodes, node(assuranceID, "oracle_assurance", base, o.Raw))
		edges = append(edges, edge("derived_from", c.ContractID, assuranceID))
		releaseID := "consumer:release:" + caseID
		complianceID := "consumer:compliance:" + caseID
		nodes = append(nodes,
			node(releaseID, "consumer_observation", base, []byte("release-observed:"+caseID)),
			node(complianceID, "consumer_observation", base, []byte("compliance-observed:"+caseID)),
		)
		edges = append(edges,
			edge("consumed_by", assuranceID, releaseID),
			edge("consumed_by", releaseID, complianceID),
		)
	}
	h1 := identity(initial)
	h2 := identity(repaired)
	var attackMutation, attackReason string
	for id, candidate := range f.Attacks {
		if candidate.ExpectedReason != "contract_identity_mismatch" || candidate.Mutation == "" {
			continue
		}
		if m.AttackID != "" {
			return nil, fmt.Errorf("ambiguous_proofline_attack")
		}
		m.AttackID, attackMutation, attackReason = id, candidate.Mutation, candidate.ExpectedReason
	}
	if m.AttackID == "" {
		return nil, fmt.Errorf("missing_proofline_attack")
	}
	drift := h1
	drift.TaskRevision++
	drift.ContractID, drift.ArtifactID = initial.ContractID+":drift", initial.ContractID+":drift"
	drift.PolicyDigest = digest([]byte(attackMutation + ":policy-v2"))
	m.DriftID = drift.ContractID
	nodes = append(nodes, Node{
		ID: drift.ContractID, Kind: "verification_contract", Identity: drift,
		PayloadDigest: digest([]byte(attackMutation + ":" + attackReason)),
	})
	edges = append(edges,
		edge("supersedes", h1.ContractID, h2.ContractID),
		edge("invalidates", h2.ContractID, m.InitialAssurance),
		edge("supersedes", h1.ContractID, drift.ContractID),
		edge("invalidates", drift.ContractID, m.InitialAssurance),
	)
	graph, err := NewGraph(nodes, edges)
	if err != nil {
		return nil, err
	}
	m.Graph = graph
	return m, nil
}
func identity(c Contract) Identity {
	return Identity{
		TaskID: c.Subject.TaskID, TaskRevision: c.Subject.TaskRevision,
		Repository: c.Subject.Repository, BaseSHA: c.Subject.BaseSHA,
		HeadSHA: c.Subject.HeadSHA, DiffDigest: c.Subject.DiffDigest,
		ContractID: c.ContractID, PolicyDigest: c.PolicyDigest, ArtifactID: c.ContractID,
	}
}
func node(id, kind string, base Identity, raw []byte) Node {
	base.ArtifactID = id
	return Node{ID: id, Kind: kind, Identity: base, PayloadDigest: digest(raw)}
}
func edge(kind, from, to string) Edge {
	return Edge{ID: kind + ":" + from + ":" + to, Type: kind, From: from, To: to}
}
func equalRaw(a, b json.RawMessage) bool {
	var left, right any
	return json.Unmarshal(a, &left) == nil && json.Unmarshal(b, &right) == nil && reflect.DeepEqual(left, right)
}

type OracleDisplay struct {
	Summary        string   `json:"summary"`
	Coverage       string   `json:"coverage"`
	Gaps           []string `json:"gaps"`
	MergeAuthority string   `json:"merge_authority"`
}
type Envelope struct {
	SchemaVersion        string        `json:"schema_version"`
	Entry                string        `json:"entry"`
	CaseID               string        `json:"case_id"`
	InputBundleDigest    string        `json:"input_bundle_digest"`
	SubjectIdentity      Identity      `json:"subject_identity"`
	NativeStatus         string        `json:"native_status"`
	ReasonCodes          []string      `json:"reason_codes"`
	NativeArtifactDigest string        `json:"native_artifact_digest"`
	FirstStaleEdge       string        `json:"first_stale_edge"`
	RetractedNodes       []string      `json:"retracted_nodes"`
	MutantStatus         string        `json:"mutant_status"`
	OracleAssurance      OracleDisplay `json:"oracle_assurance"`
}

func (m *Model) DemoEnvelope() (Envelope, error) {
	h1, err := m.Graph.Lineage(m.InitialAssurance)
	if err != nil {
		return Envelope{}, err
	}
	production, err := m.Graph.Assess(m.InitialAssurance, m.DriftID, false)
	if err != nil {
		return Envelope{}, err
	}
	mutant, err := m.Graph.Assess(m.InitialAssurance, m.DriftID, true)
	if err != nil {
		return Envelope{}, err
	}
	o := m.OracleForAssurance[m.InitialAssurance]
	drift := m.Graph.nodes[m.DriftID].Identity
	return Envelope{
		SchemaVersion: "proofline-demo/v1", Entry: "hack-proofline",
		CaseID: m.AttackID, InputBundleDigest: m.InputDigest,
		SubjectIdentity: drift, NativeStatus: production.Status,
		ReasonCodes: []string{production.Reason}, NativeArtifactDigest: h1.Digest(),
		FirstStaleEdge: production.FirstStaleEdge, RetractedNodes: production.RetractedNodes,
		MutantStatus:    mutant.Status,
		OracleAssurance: OracleDisplay{Summary: o.Summary, Coverage: o.Coverage, Gaps: o.Gaps, MergeAuthority: o.MergeAuthority},
	}, nil
}

type JoinResult struct {
	ExpectedContract string   `json:"expected_contract"`
	MatchedNodes     []string `json:"matched_nodes"`
	ImpactCandidates []string `json:"impact_candidates"`
	AffectedOldNodes []string `json:"affected_old_nodes"`
	ExplainsImpact   bool     `json:"explains_transitive_impact"`
}

func (m *Model) ExactSubjectJoin(changeID string) (JoinResult, error) {
	changed, ok := m.Graph.nodes[changeID]
	if !ok || changed.Kind != "verification_contract" {
		return JoinResult{}, fmt.Errorf("unknown_change:%s", changeID)
	}
	var matched, candidates, affected []string
	for id, n := range m.Graph.nodes {
		if sameContract(n.Identity, changed.Identity) {
			matched = append(matched, id)
		}
	}
	matchedSet := map[string]bool{}
	for _, id := range matched {
		matchedSet[id] = true
	}
	for _, e := range m.Graph.edges {
		if e.Type != "invalidates" || e.From != changeID {
			continue
		}
		for _, id := range m.Graph.descendants(e.To) {
			candidates = append(candidates, id)
			if matchedSet[id] {
				affected = append(affected, id)
			}
		}
	}
	sort.Strings(matched)
	sort.Strings(candidates)
	sort.Strings(affected)
	return JoinResult{ExpectedContract: changeID, MatchedNodes: matched,
		ImpactCandidates: candidates, AffectedOldNodes: affected,
		ExplainsImpact: len(candidates) == len(affected)}, nil
}
