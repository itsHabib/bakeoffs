package main

import (
	"os/exec"
	"path/filepath"
	"testing"
)

func TestRealProcessCrashCasesAndMutants(t *testing.T) {
	binary := filepath.Join(t.TempDir(), "resume")
	cmd := exec.Command("go", "build", "-o", binary, ".")
	if b, e := cmd.CombinedOutput(); e != nil {
		t.Fatalf("build: %v %s", e, b)
	}
	controllerBinary = binary
	if e := runSuite(t.TempDir()); e != nil {
		t.Fatal(e)
	}
}
func TestCheckerUsesContent(t *testing.T) {
	src := []byte("S1")
	if e := check(src, candidateBytes(src)); e != nil {
		t.Fatal(e)
	}
	for _, bad := range [][]byte{[]byte("checks passed"), candidateBytes([]byte("S2")), []byte(`{"Source":"S1","Body":"defect"}`)} {
		if check(src, bad) == nil {
			t.Fatal("planted defect accepted")
		}
	}
}
