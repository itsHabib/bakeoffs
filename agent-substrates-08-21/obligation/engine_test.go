package obligation

import (
	"encoding/json"
	"os"
	"slices"
	"testing"
)

func loadTestFixture(t *testing.T) Fixture {
	t.Helper()
	data, err := os.ReadFile("fixtures/exact-head-lifecycle-v1.json")
	if err != nil {
		t.Fatal(err)
	}
	var fixture Fixture
	if err := json.Unmarshal(data, &fixture); err != nil {
		t.Fatal(err)
	}
	return fixture
}

func fixtureContract(t *testing.T, fixture Fixture, id string) Contract {
	t.Helper()
	for _, contract := range fixture.Contracts {
		if contract.ContractID == id {
			return contract
		}
	}
	t.Fatalf("contract %s missing", id)
	return Contract{}
}

func fixtureAssurance(t *testing.T, fixture Fixture, id string) Assurance {
	t.Helper()
	for _, assurance := range fixture.OracleAssuranceSnapshots {
		if assurance.CaseID == id {
			return assurance
		}
	}
	t.Fatalf("assurance %s missing", id)
	return Assurance{}
}

func requireSingleOpen(t *testing.T, frontier Frontier, kind ObligationKind, claimID, reason string) {
	t.Helper()
	if len(frontier.OpenFrontier) != 1 {
		t.Fatalf("open frontier = %v, want one", frontier.OpenFrontier)
	}
	for _, item := range frontier.Obligations {
		if item.ObligationID != frontier.OpenFrontier[0] {
			continue
		}
		if item.Kind != kind || item.ClaimID != claimID || item.Reason != reason || !item.Mandatory {
			t.Fatalf("open item = %+v", item)
		}
		return
	}
	t.Fatal("open item missing from obligations")
}

func TestDeterministicContractIDsAndOrdering(t *testing.T) {
	fixture := loadTestFixture(t)
	events := []Event{Activate(fixtureContract(t, fixture, "contract-h1"))}
	first, err := Reduce(events, "")
	if err != nil {
		t.Fatal(err)
	}
	second, err := Reduce(events, "")
	if err != nil {
		t.Fatal(err)
	}
	if !slices.Equal(first.OpenFrontier, second.OpenFrontier) {
		t.Fatalf("ids changed: %v != %v", first.OpenFrontier, second.OpenFrontier)
	}
	if len(first.OpenFrontier) != 3 || !slices.IsSorted(first.OpenFrontier) {
		t.Fatalf("frontier = %v", first.OpenFrontier)
	}
	for _, item := range first.Obligations {
		if item.Origin != OriginContract || !item.Mandatory || item.Kind != CollectClaim {
			t.Fatalf("contract item = %+v", item)
		}
	}
}

func TestLateRefutationProductionVersusMutant(t *testing.T) {
	fixture := loadTestFixture(t)
	events := []Event{
		Activate(fixtureContract(t, fixture, "contract-h1")),
		Observe(fixtureAssurance(t, fixture, "aligned-h1")),
	}
	aligned, err := Reduce(events, "")
	if err != nil {
		t.Fatal(err)
	}
	if len(aligned.OpenFrontier) != 0 {
		t.Fatalf("aligned frontier = %v", aligned.OpenFrontier)
	}
	events = append(events, Observe(fixtureAssurance(t, fixture, "refuted-h1")))
	production, err := Reduce(events, "")
	if err != nil {
		t.Fatal(err)
	}
	requireSingleOpen(t, production, ResolveRefutation, "critical-finding-resolved", ReasonReopened)
	mutant, err := Reduce(events, SettledIsTerminal)
	if err != nil {
		t.Fatal(err)
	}
	if len(mutant.OpenFrontier) != 0 {
		t.Fatalf("mutant reopened %v", mutant.OpenFrontier)
	}
	if mutant.EventDigest != production.EventDigest {
		t.Fatal("mutant and production did not consume byte-identical events")
	}
}

func TestSubjectSupersessionAndH2GapLifecycle(t *testing.T) {
	fixture := loadTestFixture(t)
	h1 := fixtureContract(t, fixture, "contract-h1")
	h2 := fixtureContract(t, fixture, "contract-h2")
	events := []Event{
		Activate(h1),
		Observe(fixtureAssurance(t, fixture, "aligned-h1")),
		Observe(fixtureAssurance(t, fixture, "refuted-h1")),
		Activate(h2),
	}
	activated, err := Reduce(events, "")
	if err != nil {
		t.Fatal(err)
	}
	requireSingleOpen(t, activated, RefreshSubject, "subject-refresh", "subject_changed_refresh_required")
	for _, item := range activated.Obligations {
		if item.Subject.ContractID == "contract-h1" && item.State != Superseded {
			t.Fatalf("H1 item not superseded: %+v", item)
		}
	}
	_, err = Reduce(append(slices.Clone(events), Observe(fixtureAssurance(t, fixture, "aligned-h1"))), "")
	if RefusalReason(err) != "assurance_subject_mismatch" {
		t.Fatalf("stale H1 discharge = %v", err)
	}
	events = append(events, Observe(fixtureAssurance(t, fixture, "partial-h2")))
	partial, err := Reduce(events, "")
	if err != nil {
		t.Fatal(err)
	}
	requireSingleOpen(t, partial, CollectClaim, "critical-finding-resolved", "oracle_named_gap_opened")
	events = append(events, Observe(fixtureAssurance(t, fixture, "refreshed-h2")))
	refreshed, err := Reduce(events, "")
	if err != nil {
		t.Fatal(err)
	}
	if len(refreshed.OpenFrontier) != 0 {
		t.Fatalf("refreshed frontier = %v", refreshed.OpenFrontier)
	}
}

func TestH2OmittedOutcomesPreserveContractFloor(t *testing.T) {
	fixture := loadTestFixture(t)
	base := []Event{
		Activate(fixtureContract(t, fixture, "contract-h1")),
		Observe(fixtureAssurance(t, fixture, "aligned-h1")),
		Activate(fixtureContract(t, fixture, "contract-h2")),
	}
	tests := []struct {
		name     string
		outcomes map[string]string
		missing  []string
	}{
		{"zero", map[string]string{}, []string{"block-dominates", "build-and-test", "critical-finding-resolved"}},
		{"one", map[string]string{"build-and-test": "supported"}, []string{"block-dominates", "critical-finding-resolved"}},
		{"partial", map[string]string{"build-and-test": "supported", "block-dominates": "supported"}, []string{"critical-finding-resolved"}},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			snapshot := fixtureAssurance(t, fixture, "refreshed-h2")
			snapshot.ClaimOutcomes, snapshot.Gaps = test.outcomes, nil
			frontier, err := Reduce(append(slices.Clone(base), Observe(snapshot)), "")
			if err != nil {
				t.Fatal(err)
			}
			openClaims := []string{}
			for _, item := range frontier.Obligations {
				if item.State == Open && item.Mandatory {
					if item.Kind != CollectClaim || item.Subject.ContractID != "contract-h2" {
						t.Fatalf("open item = %+v", item)
					}
					openClaims = append(openClaims, item.ClaimID)
				}
			}
			slices.Sort(openClaims)
			if !slices.Equal(openClaims, test.missing) {
				t.Fatalf("open claims = %v, want %v", openClaims, test.missing)
			}
		})
	}
}

func TestOverlayFloorAndReplay(t *testing.T) {
	fixture := loadTestFixture(t)
	base := []Event{Activate(fixtureContract(t, fixture, "contract-h1"))}
	for name, overlay := range map[string]AgentOverlay{
		"delete":          {RemoveClaimIDs: []string{"critical-finding-resolved"}},
		"downgrade":       {DowngradeClaimIDs: []string{"critical-finding-resolved"}},
		"re-add optional": {Add: []OverlayAddition{{ClaimID: "critical-finding-resolved"}}},
	} {
		t.Run(name, func(t *testing.T) {
			_, err := Reduce(append(slices.Clone(base), ApplyOverlay(overlay)), "")
			if RefusalReason(err) != ReasonWeakened {
				t.Fatalf("got %v", err)
			}
		})
	}
	added := append(slices.Clone(base), ApplyOverlay(AgentOverlay{Add: []OverlayAddition{{ClaimID: "timing-diagnostic"}}}))
	first, err := Reduce(added, "")
	if err != nil {
		t.Fatal(err)
	}
	replayed, err := Reduce(added, "")
	if err != nil {
		t.Fatal(err)
	}
	firstJSON, _ := json.Marshal(first)
	replayedJSON, _ := json.Marshal(replayed)
	if string(firstJSON) != string(replayedJSON) {
		t.Fatal("overlay changed during replay")
	}
	found := false
	for _, item := range first.Obligations {
		found = found || item.ClaimID == "timing-diagnostic" && item.Origin == OriginAgent && !item.Mandatory
	}
	if !found {
		t.Fatal("optional diagnostic missing")
	}
}

func TestRawReceiptCannotDischarge(t *testing.T) {
	fixture := loadTestFixture(t)
	base := []Event{Activate(fixtureContract(t, fixture, "contract-h1"))}
	before, err := Reduce(base, "")
	if err != nil {
		t.Fatal(err)
	}
	_, err = Reduce(append(slices.Clone(base), SubmitRawReceipt(fixture.TrustedRunnerReceipts)), "")
	if RefusalReason(err) != "raw_receipt_input_forbidden" {
		t.Fatalf("got %v", err)
	}
	after, err := Reduce(base, "")
	if err != nil {
		t.Fatal(err)
	}
	if before.EventDigest != after.EventDigest || !slices.Equal(before.OpenFrontier, after.OpenFrontier) {
		t.Fatal("rejected raw receipt changed replay state")
	}
}
