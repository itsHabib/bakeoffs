package proofline

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"sort"
)

const SchemaVersion = "proofline/v1"

type Identity struct {
	TaskID       string `json:"task_id"`
	TaskRevision int    `json:"task_revision"`
	Repository   string `json:"repository"`
	BaseSHA      string `json:"base_sha"`
	HeadSHA      string `json:"head_sha"`
	DiffDigest   string `json:"diff_digest"`
	ContractID   string `json:"contract_id"`
	PolicyDigest string `json:"policy_digest"`
	ArtifactID   string `json:"artifact_id"`
}
type Node struct {
	ID            string   `json:"id"`
	Kind          string   `json:"kind"`
	Identity      Identity `json:"identity"`
	PayloadDigest string   `json:"payload_digest"`
}
type Edge struct {
	ID   string `json:"id"`
	Type string `json:"type"`
	From string `json:"from"`
	To   string `json:"to"`
}
type ProoflineV1 struct {
	SchemaVersion string `json:"schema_version"`
	Selection     string `json:"selection"`
	Nodes         []Node `json:"nodes"`
	Edges         []Edge `json:"edges"`
}
type Graph struct {
	nodes map[string]Node
	edges map[string]Edge
}

var nodeKinds = map[string]bool{
	"verification_contract": true,
	"trusted_receipt":       true,
	"oracle_assurance":      true,
	"consumer_observation":  true,
}
var edgeKinds = map[string]bool{
	"evaluates": true, "includes": true, "derived_from": true,
	"supersedes": true, "invalidates": true, "consumed_by": true,
}

func NewGraph(nodes []Node, edges []Edge) (*Graph, error) {
	g := &Graph{nodes: map[string]Node{}, edges: map[string]Edge{}}
	for _, n := range nodes {
		if n.ID == "" || !nodeKinds[n.Kind] || n.PayloadDigest == "" || !complete(n.Identity) {
			return nil, fmt.Errorf("invalid_node_identity:%s", n.ID)
		}
		if prior, ok := g.nodes[n.ID]; ok {
			if prior != n {
				return nil, fmt.Errorf("duplicate_node_conflict:%s", n.ID)
			}
			continue
		}
		g.nodes[n.ID] = n
	}
	for _, e := range edges {
		if e.ID == "" || !edgeKinds[e.Type] {
			return nil, fmt.Errorf("invalid_edge:%s", e.ID)
		}
		if _, ok := g.nodes[e.From]; !ok {
			return nil, fmt.Errorf("dangling_edge:%s", e.ID)
		}
		if _, ok := g.nodes[e.To]; !ok {
			return nil, fmt.Errorf("dangling_edge:%s", e.ID)
		}
		if prior, ok := g.edges[e.ID]; ok {
			if prior != e {
				return nil, fmt.Errorf("duplicate_edge_conflict:%s", e.ID)
			}
			continue
		}
		g.edges[e.ID] = e
	}
	if err := g.validateRelations(); err != nil {
		return nil, err
	}
	if g.cyclic() {
		return nil, errors.New("forbidden_cycle")
	}
	return g, nil
}
func complete(i Identity) bool {
	return i.TaskID != "" && i.TaskRevision > 0 && i.Repository != "" &&
		i.BaseSHA != "" && i.HeadSHA != "" && i.DiffDigest != "" &&
		i.ContractID != "" && i.PolicyDigest != "" && i.ArtifactID != ""
}
func sameContract(a, b Identity) bool {
	a.ArtifactID, b.ArtifactID = "", ""
	return a == b
}
func sameHead(a, b Identity) bool { return a.Repository == b.Repository && a.HeadSHA == b.HeadSHA }
func (g *Graph) validateRelations() error {
	for _, e := range g.edges {
		from, to := g.nodes[e.From], g.nodes[e.To]
		validKinds := false
		switch e.Type {
		case "evaluates":
			validKinds = from.Kind == "verification_contract" && to.Kind == "trusted_receipt"
		case "includes":
			validKinds = from.Kind == "trusted_receipt" && to.Kind == "oracle_assurance"
		case "derived_from":
			validKinds = from.Kind == "verification_contract" && to.Kind == "oracle_assurance"
		case "consumed_by":
			validKinds = (from.Kind == "oracle_assurance" || from.Kind == "consumer_observation") && to.Kind == "consumer_observation"
		case "supersedes":
			validKinds = from.Kind == "verification_contract" && to.Kind == "verification_contract" &&
				from.Identity.Repository == to.Identity.Repository && from.Identity.TaskID == to.Identity.TaskID &&
				!sameContract(from.Identity, to.Identity)
		case "invalidates":
			validKinds = from.Kind == "verification_contract" && to.Kind == "oracle_assurance" &&
				from.Identity.Repository == to.Identity.Repository && from.Identity.TaskID == to.Identity.TaskID &&
				!sameContract(from.Identity, to.Identity)
		}
		if !validKinds {
			return fmt.Errorf("invalid_relation:%s", e.ID)
		}
		if e.Type != "supersedes" && e.Type != "invalidates" && !sameContract(from.Identity, to.Identity) {
			return fmt.Errorf("identity_mismatch:%s", e.ID)
		}
	}
	return nil
}
func (g *Graph) cyclic() bool {
	state := map[string]uint8{}
	var visit func(string) bool
	visit = func(id string) bool {
		if state[id] == 1 {
			return true
		}
		if state[id] == 2 {
			return false
		}
		state[id] = 1
		for _, e := range g.edges {
			if e.From == id && visit(e.To) {
				return true
			}
		}
		state[id] = 2
		return false
	}
	for id := range g.nodes {
		if visit(id) {
			return true
		}
	}
	return false
}
func (g *Graph) Lineage(assuranceID string) (ProoflineV1, error) {
	n, ok := g.nodes[assuranceID]
	if !ok || n.Kind != "oracle_assurance" {
		return ProoflineV1{}, fmt.Errorf("unknown_assurance:%s", assuranceID)
	}
	selected := map[string]bool{assuranceID: true}
	changed := true
	for changed {
		changed = false
		for _, e := range g.edges {
			ancestor := e.Type == "evaluates" || e.Type == "includes" || e.Type == "derived_from" ||
				e.Type == "invalidates" || e.Type == "supersedes"
			descendant := e.Type == "consumed_by"
			if ancestor && selected[e.To] && !selected[e.From] {
				selected[e.From], changed = true, true
			}
			if descendant && selected[e.From] && !selected[e.To] {
				selected[e.To], changed = true, true
			}
		}
	}
	out := ProoflineV1{SchemaVersion: SchemaVersion, Selection: assuranceID}
	for id, n := range g.nodes {
		if selected[id] {
			out.Nodes = append(out.Nodes, n)
		}
	}
	for _, e := range g.edges {
		if selected[e.From] && selected[e.To] {
			out.Edges = append(out.Edges, e)
		}
	}
	sort.Slice(out.Nodes, func(i, j int) bool { return out.Nodes[i].ID < out.Nodes[j].ID })
	sort.Slice(out.Edges, func(i, j int) bool {
		a, b := out.Edges[i], out.Edges[j]
		if a.From != b.From {
			return a.From < b.From
		}
		if a.Type != b.Type {
			return a.Type < b.Type
		}
		if a.To != b.To {
			return a.To < b.To
		}
		return a.ID < b.ID
	})
	return out, nil
}
func (p ProoflineV1) Digest() string { b, _ := json.Marshal(p); return digest(b) }

type Retraction struct {
	Status         string   `json:"status"`
	Reason         string   `json:"reason,omitempty"`
	FirstStaleEdge string   `json:"first_stale_edge,omitempty"`
	RetractedNodes []string `json:"retracted_nodes,omitempty"`
}

func (g *Graph) Assess(assuranceID, changeID string, headOnly bool) (Retraction, error) {
	lineage, err := g.Lineage(assuranceID)
	if err != nil {
		return Retraction{}, err
	}
	var first Edge
	for _, e := range lineage.Edges {
		if e.Type == "derived_from" && e.To == assuranceID {
			first = e
			break
		}
	}
	if first.ID == "" {
		return Retraction{}, errors.New("missing_contract_ancestry")
	}
	if changeID == "" {
		return Retraction{Status: "current"}, nil
	}
	changed, ok := g.nodes[changeID]
	if !ok || changed.Kind != "verification_contract" {
		return Retraction{}, fmt.Errorf("unknown_change:%s", changeID)
	}
	hasInvalidation, hasSupersession := false, false
	for _, e := range lineage.Edges {
		hasInvalidation = hasInvalidation || e.Type == "invalidates" && e.From == changeID && e.To == assuranceID
		hasSupersession = hasSupersession || e.Type == "supersedes" && e.From == first.From && e.To == changeID
	}
	if !hasInvalidation || !hasSupersession {
		return Retraction{Status: "current"}, nil
	}
	contract := g.nodes[first.From]
	matches := sameContract(contract.Identity, changed.Identity)
	if headOnly {
		matches = sameHead(contract.Identity, changed.Identity)
	}
	if matches {
		return Retraction{Status: "current"}, nil
	}
	reason := "contract_identity_mismatch"
	if !sameHead(contract.Identity, changed.Identity) {
		reason = "subject_identity_mismatch"
	}
	ids := g.descendants(assuranceID)
	return Retraction{
		Status:         "historical",
		Reason:         reason,
		FirstStaleEdge: first.From + " -> " + first.To,
		RetractedNodes: ids,
	}, nil
}
func (g *Graph) descendants(root string) []string {
	affected := map[string]bool{root: true}
	changed := true
	for changed {
		changed = false
		for _, e := range g.edges {
			if e.Type == "consumed_by" && affected[e.From] && !affected[e.To] {
				affected[e.To], changed = true, true
			}
		}
	}
	ids := make([]string, 0, len(affected))
	for id := range affected {
		ids = append(ids, id)
	}
	sort.Strings(ids)
	return ids
}
func digest(b []byte) string { h := sha256.Sum256(b); return "sha256:" + hex.EncodeToString(h[:]) }
