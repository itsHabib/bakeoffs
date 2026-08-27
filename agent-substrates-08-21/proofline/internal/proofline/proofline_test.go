package proofline

import (
	"bytes"
	"encoding/json"
	"os"
	"reflect"
	"strings"
	"testing"
)

const fixturePath = "../../fixtures/exact-head-lifecycle-v1.json"

func loadModel(t *testing.T) *Model {
	t.Helper()
	m, err := Load(fixturePath)
	if err != nil {
		t.Fatal(err)
	}
	return m
}

func TestCanonicalLineagesAndStableDigest(t *testing.T) {
	m := loadModel(t)
	emptyDigest := digest(nil)
	for _, tc := range []struct {
		id           string
		nodes, edges int
	}{
		{m.InitialAssurance, 9, 13}, {m.RepairedAssurance, 8, 10},
	} {
		id := tc.id
		first, err := m.Graph.Lineage(id)
		if err != nil {
			t.Fatal(err)
		}
		second, err := m.Graph.Lineage(id)
		if err != nil {
			t.Fatal(err)
		}
		if len(first.Nodes) != tc.nodes || len(first.Edges) != tc.edges {
			t.Fatalf("%s: got %d nodes and %d edges", id, len(first.Nodes), len(first.Edges))
		}
		if !reflect.DeepEqual(first, second) || first.Digest() != second.Digest() {
			t.Fatalf("%s lineage is not byte-stable", id)
		}
		for i := 1; i < len(first.Nodes); i++ {
			if first.Nodes[i-1].ID >= first.Nodes[i].ID {
				t.Fatalf("nodes not canonical: %s then %s", first.Nodes[i-1].ID, first.Nodes[i].ID)
			}
		}
		for _, n := range first.Nodes {
			if n.PayloadDigest == emptyDigest {
				t.Fatalf("%s payload was not retained", n.ID)
			}
		}
		for i := 1; i < len(first.Edges); i++ {
			left, right := first.Edges[i-1], first.Edges[i]
			leftKey := left.From + "\x00" + left.Type + "\x00" + left.To + "\x00" + left.ID
			rightKey := right.From + "\x00" + right.Type + "\x00" + right.To + "\x00" + right.ID
			if leftKey >= rightKey {
				t.Fatalf("edges not canonical: %s then %s", left.ID, right.ID)
			}
		}
	}
}

func TestHeadChangeRetractsTransitiveDescendantsWithoutDeletion(t *testing.T) {
	m := loadModel(t)
	before, _ := m.Graph.Lineage(m.InitialAssurance)
	r, err := m.Graph.Assess(m.InitialAssurance, m.RepairedContract, false)
	if err != nil {
		t.Fatal(err)
	}
	caseID := m.OracleForAssurance[m.InitialAssurance].CaseID
	want := []string{m.InitialAssurance, "consumer:compliance:" + caseID, "consumer:release:" + caseID}
	if r.Status != "historical" || r.Reason != "subject_identity_mismatch" ||
		r.FirstStaleEdge != m.InitialContract+" -> "+m.InitialAssurance || !reflect.DeepEqual(r.RetractedNodes, want) {
		t.Fatalf("unexpected retraction: %+v", r)
	}
	after, err := m.Graph.Lineage(m.InitialAssurance)
	if err != nil {
		t.Fatal(err)
	}
	if !reflect.DeepEqual(before, after) {
		t.Fatal("historical lineage was deleted or changed")
	}
}

func TestFreshH2IsCurrent(t *testing.T) {
	m := loadModel(t)
	r, err := m.Graph.Assess(m.RepairedAssurance, "", false)
	if err != nil {
		t.Fatal(err)
	}
	if r.Status != "current" || r.Reason != "" {
		t.Fatalf("fresh H2 not current: %+v", r)
	}
}

func TestClosedGraphExercisesEveryRelation(t *testing.T) {
	m := loadModel(t)
	seen := map[string]bool{}
	for _, e := range m.Graph.edges {
		seen[e.Type] = true
	}
	for relation := range edgeKinds {
		if !seen[relation] {
			t.Fatalf("relation %s is not exercised", relation)
		}
	}
}

func TestSameHeadDriftKillsProductionButNotMutant(t *testing.T) {
	m := loadModel(t)
	production, err := m.Graph.Assess(m.InitialAssurance, m.DriftID, false)
	if err != nil {
		t.Fatal(err)
	}
	mutant, err := m.Graph.Assess(m.InitialAssurance, m.DriftID, true)
	if err != nil {
		t.Fatal(err)
	}
	if production.Status != "historical" || production.Reason != "contract_identity_mismatch" {
		t.Fatalf("production retained drift: %+v", production)
	}
	if production.FirstStaleEdge != m.InitialContract+" -> "+m.InitialAssurance {
		t.Fatalf("wrong first stale edge: %+v", production)
	}
	if mutant.Status != "current" {
		t.Fatalf("mutant did not expose planted bug: %+v", mutant)
	}
}

func TestRetractionRequiresRecordedChangeLineage(t *testing.T) {
	m := loadModel(t)
	for _, removed := range []string{
		edge("invalidates", m.DriftID, m.InitialAssurance).ID,
		edge("supersedes", m.InitialContract, m.DriftID).ID,
	} {
		t.Run(removed, func(t *testing.T) {
			g := graphWithoutEdge(t, m.Graph, removed)
			got, err := g.Assess(m.InitialAssurance, m.DriftID, false)
			if err != nil {
				t.Fatal(err)
			}
			if got.Status != "current" {
				t.Fatalf("retracted without %s: %+v", removed, got)
			}
		})
	}
}

func TestLoaderRejectsConflictingDuplicateIDs(t *testing.T) {
	tests := []struct {
		name, field, code string
		value             any
		selectRaw         func(*Fixture) *[]json.RawMessage
	}{
		{"contract", "risk_floor", "duplicate_contract_conflict", "CHANGED", func(f *Fixture) *[]json.RawMessage { return &f.Contracts }},
		{"receipt", "limitations", "duplicate_receipt_conflict", []any{"changed"}, func(f *Fixture) *[]json.RawMessage { return &f.TrustedRunnerReceipts }},
		{"assurance", "summary", "duplicate_assurance_conflict", "CHANGED", func(f *Fixture) *[]json.RawMessage { return &f.OracleSnapshots }},
	}
	original, err := os.ReadFile(fixturePath)
	if err != nil {
		t.Fatal(err)
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var f Fixture
			if err := json.Unmarshal(original, &f); err != nil {
				t.Fatal(err)
			}
			items := tt.selectRaw(&f)
			var duplicate map[string]any
			if err := json.Unmarshal((*items)[0], &duplicate); err != nil {
				t.Fatal(err)
			}
			duplicate[tt.field] = tt.value
			raw, _ := json.Marshal(duplicate)
			*items = append(*items, raw)
			mutated, _ := json.Marshal(f)
			_, err := loadBytes(mutated)
			if err == nil || !strings.Contains(err.Error(), tt.code) {
				t.Fatalf("want %s, got %v", tt.code, err)
			}
		})
	}
}

func TestAttackSelectionFollowsSemanticRecordAndFixtureID(t *testing.T) {
	original, err := os.ReadFile(fixturePath)
	if err != nil {
		t.Fatal(err)
	}
	base, err := loadBytes(original)
	if err != nil {
		t.Fatal(err)
	}
	var document map[string]json.RawMessage
	if err := json.Unmarshal(original, &document); err != nil {
		t.Fatal(err)
	}
	var attacks map[string]json.RawMessage
	if err := json.Unmarshal(document["attacks"], &attacks); err != nil {
		t.Fatal(err)
	}
	attack := attacks[base.AttackID]
	delete(attacks, base.AttackID)
	const renamedID = "renamed_proofline_attack"
	attacks[renamedID] = attack
	document["attacks"], _ = json.Marshal(attacks)
	renamedBytes, _ := json.Marshal(document)
	renamed, err := loadBytes(renamedBytes)
	if err != nil {
		t.Fatal(err)
	}
	envelope, err := renamed.DemoEnvelope()
	if err != nil {
		t.Fatal(err)
	}
	production, _ := renamed.Graph.Assess(renamed.InitialAssurance, renamed.DriftID, false)
	mutant, _ := renamed.Graph.Assess(renamed.InitialAssurance, renamed.DriftID, true)
	baseProduction, _ := base.Graph.Assess(base.InitialAssurance, base.DriftID, false)
	if renamed.AttackID != renamedID || envelope.CaseID != renamedID ||
		!reflect.DeepEqual(production, baseProduction) ||
		production.Status != "historical" || mutant.Status != "current" {
		t.Fatalf("renamed semantic attack did not drive mechanics: model=%s envelope=%s production=%+v mutant=%+v", renamed.AttackID, envelope.CaseID, production, mutant)
	}
	attacks["second_semantic_match"] = attack
	document["attacks"], _ = json.Marshal(attacks)
	ambiguous, _ := json.Marshal(document)
	if _, err := loadBytes(ambiguous); err == nil || err.Error() != "ambiguous_proofline_attack" {
		t.Fatalf("want ambiguous_proofline_attack, got %v", err)
	}
	delete(attacks, renamedID)
	delete(attacks, "second_semantic_match")
	document["attacks"], _ = json.Marshal(attacks)
	missing, _ := json.Marshal(document)
	if _, err := loadBytes(missing); err == nil || err.Error() != "missing_proofline_attack" {
		t.Fatalf("want missing_proofline_attack, got %v", err)
	}
}

func TestGraphRejectsInvalidInput(t *testing.T) {
	id := testIdentity("contract-a", "head-a")
	c := testNode("contract-a", "verification_contract", id)
	rID := id
	rID.ArtifactID = "receipt-a"
	r := testNode("receipt-a", "trusted_receipt", rID)
	aID := id
	aID.ArtifactID = "assurance-a"
	a := testNode("assurance-a", "oracle_assurance", aID)

	tests := []struct {
		name  string
		nodes []Node
		edges []Edge
		code  string
	}{
		{"dangling edge", []Node{c}, []Edge{edge("derived_from", "contract-a", "missing")}, "dangling_edge"},
		{"duplicate conflict", []Node{c, func() Node { n := c; n.PayloadDigest = "sha256:other"; return n }()}, nil, "duplicate_node_conflict"},
		{"missing identity", []Node{func() Node { n := c; n.Identity.PolicyDigest = ""; return n }()}, nil, "invalid_node_identity"},
		{"identity mismatch", []Node{c, func() Node { n := r; n.Identity.PolicyDigest = "sha256:changed"; return n }()}, []Edge{edge("evaluates", c.ID, r.ID)}, "identity_mismatch"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := NewGraph(tt.nodes, tt.edges)
			if err == nil || !strings.Contains(err.Error(), tt.code) {
				t.Fatalf("want %s, got %v", tt.code, err)
			}
		})
	}

	bID := testIdentity("contract-b", "head-b")
	b := testNode("contract-b", "verification_contract", bID)
	_, err := NewGraph([]Node{c, b, r, a}, []Edge{
		edge("supersedes", c.ID, b.ID), edge("supersedes", b.ID, c.ID),
	})
	if err == nil || err.Error() != "forbidden_cycle" {
		t.Fatalf("want forbidden_cycle, got %v", err)
	}
}

func TestOracleFieldsCopiedUnchanged(t *testing.T) {
	m := loadModel(t)
	envelope, err := m.DemoEnvelope()
	if err != nil {
		t.Fatal(err)
	}
	o := m.OracleForAssurance[m.InitialAssurance]
	want := OracleDisplay{o.Summary, o.Coverage, o.Gaps, o.MergeAuthority}
	if !reflect.DeepEqual(envelope.OracleAssurance, want) {
		t.Fatalf("oracle fields changed: got %+v want %+v", envelope.OracleAssurance, want)
	}
}

func TestExactSubjectComparatorCannotSeeOldImpact(t *testing.T) {
	m := loadModel(t)
	got, err := m.ExactSubjectJoin(m.RepairedContract)
	if err != nil {
		t.Fatal(err)
	}
	if len(got.MatchedNodes) != 7 || len(got.ImpactCandidates) != 3 || len(got.AffectedOldNodes) != 0 || got.ExplainsImpact {
		t.Fatalf("unexpected comparator: %+v", got)
	}
	control, err := m.ExactSubjectJoin(m.InitialContract)
	if err != nil {
		t.Fatal(err)
	}
	if len(control.ImpactCandidates) != 0 || !control.ExplainsImpact {
		t.Fatalf("comparator control did not change conclusion: %+v", control)
	}
	without := graphWithoutEdge(t, m.Graph, edge("invalidates", m.RepairedContract, m.InitialAssurance).ID)
	copy := *m
	copy.Graph = without
	control, err = copy.ExactSubjectJoin(m.RepairedContract)
	if err != nil || len(control.ImpactCandidates) != 0 || !control.ExplainsImpact {
		t.Fatalf("comparator ignored removed impact edge: %+v %v", control, err)
	}
}

func TestNormalizedDemoMatchesGolden(t *testing.T) {
	m := loadModel(t)
	envelope, err := m.DemoEnvelope()
	if err != nil {
		t.Fatal(err)
	}
	got, err := json.MarshalIndent(envelope, "", "  ")
	if err != nil {
		t.Fatal(err)
	}
	got = append(got, '\n')
	want, err := os.ReadFile("../../testdata/demo.golden.json")
	if err != nil {
		t.Fatal(err)
	}
	if !bytes.Equal(got, want) {
		t.Fatalf("normalized demo differs from golden\n--- got ---\n%s\n--- want ---\n%s", got, want)
	}
}

func testIdentity(contract, head string) Identity {
	return Identity{
		TaskID: "task", TaskRevision: 1, Repository: "example/repo",
		BaseSHA: "base", HeadSHA: head, DiffDigest: "sha256:diff",
		ContractID: contract, PolicyDigest: "sha256:policy", ArtifactID: contract,
	}
}

func testNode(id, kind string, identity Identity) Node {
	return Node{ID: id, Kind: kind, Identity: identity, PayloadDigest: "sha256:payload"}
}

func graphWithoutEdge(t *testing.T, g *Graph, removed string) *Graph {
	t.Helper()
	var nodes []Node
	var edges []Edge
	for _, n := range g.nodes {
		nodes = append(nodes, n)
	}
	for id, e := range g.edges {
		if id != removed {
			edges = append(edges, e)
		}
	}
	copy, err := NewGraph(nodes, edges)
	if err != nil {
		t.Fatal(err)
	}
	return copy
}
