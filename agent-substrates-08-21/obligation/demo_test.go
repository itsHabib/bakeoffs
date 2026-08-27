package obligation

import (
	"bytes"
	"os"
	"testing"
)

func TestNormalizedDemoMatchesGoldens(t *testing.T) {
	first, err := RunDemo(".")
	if err != nil {
		t.Fatal(err)
	}
	second, err := RunDemo(".")
	if err != nil {
		t.Fatal(err)
	}
	firstJSON, err := Pretty(first.Output)
	if err != nil {
		t.Fatal(err)
	}
	secondJSON, err := Pretty(second.Output)
	if err != nil {
		t.Fatal(err)
	}
	if !bytes.Equal(firstJSON, secondJSON) {
		t.Fatal("two normalized demo runs differ")
	}
	demoGolden, err := os.ReadFile("testdata/demo.golden.json")
	if err != nil {
		t.Fatal(err)
	}
	if !bytes.Equal(firstJSON, demoGolden) {
		t.Fatal("normalized demo differs from testdata/demo.golden.json")
	}
	frontierJSON, err := Pretty(first.Final)
	if err != nil {
		t.Fatal(err)
	}
	frontierGolden, err := os.ReadFile("testdata/frontier.golden.json")
	if err != nil {
		t.Fatal(err)
	}
	if !bytes.Equal(frontierJSON, frontierGolden) {
		t.Fatal("replayed frontier differs from testdata/frontier.golden.json")
	}
}

func TestOracleEchoIsUnchanged(t *testing.T) {
	fixture := loadTestFixture(t)
	stages, err := RunDemo(".")
	if err != nil {
		t.Fatal(err)
	}
	echoes := stages.Output.CommonEnvelope.DisplayedOracleAssurance
	if len(echoes) != 4 {
		t.Fatalf("echoes = %d", len(echoes))
	}
	for _, echo := range echoes {
		original := fixtureAssurance(t, fixture, echo.CaseID)
		if echo.Summary != original.Summary || echo.Coverage != original.Coverage ||
			echo.MergeAuthority != original.MergeAuthority ||
			!slicesEqual(echo.Gaps, original.Gaps) {
			t.Fatalf("snapshot %s changed in display", echo.CaseID)
		}
	}
}

func TestSchedulerComparatorExercisesReopenHook(t *testing.T) {
	fixture := loadTestFixture(t)
	contract := fixtureContract(t, fixture, "contract-h1")
	aligned := fixtureAssurance(t, fixture, "aligned-h1")
	refuted := fixtureAssurance(t, fixture, "refuted-h1")
	native, err := Reduce([]Event{Activate(contract), Observe(aligned), Observe(refuted)}, "")
	if err != nil {
		t.Fatal(err)
	}
	late := compareScheduler(contract, aligned, refuted, native)
	if !late.SettledBeforeLate || !late.ExplicitReopenHook ||
		late.LateNode != "repair-critical-finding-resolved" ||
		late.ContractDerivedIdentity || late.MonotoneOverlayFloor ||
		!late.ReplayableFrontier || late.SameNativeArtifact {
		t.Fatalf("late comparator = %+v", late)
	}
	nodes, reopened := runScheduler(contract, aligned)
	for _, node := range nodes {
		if node.State == "open" || reopened {
			t.Fatalf("aligned scheduler did not settle: %+v", nodes)
		}
	}
}

func slicesEqual(a, b []string) bool {
	if len(a) != len(b) {
		return false
	}
	for i := range a {
		if a[i] != b[i] {
			return false
		}
	}
	return true
}
