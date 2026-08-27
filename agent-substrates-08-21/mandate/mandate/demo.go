package mandate

import (
	"crypto/ed25519"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"
)

const FrozenInputDigest = "sha256:ec742e0129e44d0529bbf5051125b33cfa6bea3c665463062c5da0f1f8947402"

type DemoResult struct {
	WithoutMandate       MandateDecisionReceiptV1 `json:"without_mandate"`
	ValidRequest         MandateDecisionReceiptV1 `json:"valid_request"`
	SignaturesOnly       MandateDecisionReceiptV1 `json:"signatures_only_mutant"`
	Production           MandateDecisionReceiptV1 `json:"production"`
	CapabilityComparator string                   `json:"standard_capability_comparator"`
	Envelope             DemoEnvelope             `json:"envelope"`
}

type fixtureKeys struct {
	Warning      string `json:"warning"`
	RootSeed     string `json:"root_seed"`
	DelegateSeed string `json:"delegate_seed"`
	WorkerSeed   string `json:"worker_seed"`
}

type frozenDeck struct {
	Contracts []struct {
		ContractID string  `json:"contract_id"`
		Subject    Subject `json:"subject"`
	} `json:"contracts"`
	OracleAssuranceSnapshots []struct {
		CaseID         string   `json:"case_id"`
		SubjectHead    string   `json:"subject_head"`
		ContractID     string   `json:"contract_id"`
		Summary        string   `json:"summary"`
		Coverage       string   `json:"coverage"`
		Gaps           []string `json:"gaps"`
		MergeAuthority string   `json:"merge_authority"`
	} `json:"oracle_assurance_snapshots"`
}

func BuildDemo(repoRoot string) (DemoResult, error) {
	deckBytes, err := os.ReadFile(filepath.Join(repoRoot, "fixtures", "exact-head-lifecycle-v1.json"))
	if err != nil {
		return DemoResult{}, err
	}
	if digest := DigestBytes(deckBytes); digest != FrozenInputDigest {
		return DemoResult{}, fmt.Errorf("frozen input digest mismatch: %s", digest)
	}
	var deck frozenDeck
	if err := json.Unmarshal(deckBytes, &deck); err != nil {
		return DemoResult{}, err
	}
	subject, oracle, err := selectH2(deck)
	if err != nil {
		return DemoResult{}, err
	}
	keys, err := loadKeys(filepath.Join(repoRoot, "fixtures", "keys.json"))
	if err != nil {
		return DemoResult{}, err
	}

	now := time.Date(2026, 8, 21, 12, 0, 0, 0, time.UTC)
	root, child, validRequest, widenedChild, publishRequest, err := demoChain(subject, keys, now)
	if err != nil {
		return DemoResult{}, err
	}

	without := DenyWithoutMandate(subject, ReviewCandidate)
	valid := VerifyChainAndRequest(root, child, validRequest, keys.root.Public().(ed25519.PublicKey), now)
	mutant := verifyChainAndRequest(root, widenedChild, publishRequest, keys.root.Public().(ed25519.PublicKey), now, verifyChildSignaturesOnly)
	production := VerifyChainAndRequest(root, widenedChild, publishRequest, keys.root.Public().(ed25519.PublicKey), now)
	if without.ReasonCode != "no_authority" || valid.Decision != Allow || mutant.Decision != Allow || production.ReasonCode != "scope_inflation" {
		return DemoResult{}, errors.New("demo invariant failed")
	}

	artifact := struct {
		WithoutMandate string `json:"without_mandate"`
		Valid          string `json:"valid"`
		Mutant         string `json:"mutant"`
		Production     string `json:"production"`
	}{ReceiptDigest(without), ReceiptDigest(valid), ReceiptDigest(mutant), ReceiptDigest(production)}
	artifactDigest, _ := Digest(artifact)
	lines, err := SourceLines(repoRoot)
	if err != nil {
		return DemoResult{}, err
	}
	envelope := DemoEnvelope{
		SchemaVersion:        EnvelopeSchema,
		Entry:                "hack-mandate",
		CaseID:               "h2-one-hop-attenuation",
		InputBundleDigest:    FrozenInputDigest,
		Subject:              subject,
		NativeStatus:         "attenuation_enforced",
		ReasonCodes:          []string{"no_authority", "authorized", "scope_inflation"},
		NativeArtifactDigest: artifactDigest,
		OracleAssurance:      oracle,
		SourceLines:          lines,
	}
	return DemoResult{
		WithoutMandate:       without,
		ValidRequest:         valid,
		SignaturesOnly:       mutant,
		Production:           production,
		CapabilityComparator: "MATCH: a standard caveated capability plus audit receipt can enforce the same law; custom-format necessity is not established",
		Envelope:             envelope,
	}, nil
}

func verifyChildSignaturesOnly(parent, child SignedMandate, now time.Time) error {
	return verifyChild(parent, child, now, false)
}

type demoKeys struct {
	root, delegate, worker ed25519.PrivateKey
}

func loadKeys(path string) (demoKeys, error) {
	b, err := os.ReadFile(path)
	if err != nil {
		return demoKeys{}, err
	}
	var f fixtureKeys
	if err := DecodeStrict(b, &f); err != nil {
		return demoKeys{}, err
	}
	if f.Warning != "SYNTHETIC TEST KEYS ONLY" {
		return demoKeys{}, errors.New("fixture key warning missing")
	}
	parse := func(s string) (ed25519.PrivateKey, error) {
		seed, err := hex.DecodeString(s)
		if err != nil || len(seed) != ed25519.SeedSize {
			return nil, errors.New("invalid fixture seed")
		}
		return ed25519.NewKeyFromSeed(seed), nil
	}
	root, err := parse(f.RootSeed)
	if err != nil {
		return demoKeys{}, err
	}
	delegate, err := parse(f.DelegateSeed)
	if err != nil {
		return demoKeys{}, err
	}
	worker, err := parse(f.WorkerSeed)
	return demoKeys{root: root, delegate: delegate, worker: worker}, err
}

func demoChain(subject Subject, keys demoKeys, now time.Time) (SignedMandate, SignedMandate, SignedRequest, SignedMandate, SignedRequest, error) {
	rootPayload := WorkMandateV1{
		SchemaVersion: MandateSchema, MandateID: "mandate-root-h2",
		IssuerPublicKey: PublicKeyHex(keys.root.Public().(ed25519.PublicKey)), AudiencePublicKey: PublicKeyHex(keys.delegate.Public().(ed25519.PublicKey)),
		Subject: subject, AllowedActions: []Action{InspectCandidate, ReviewCandidate},
		IssuedAt: now.Add(-time.Hour).Format(time.RFC3339), ExpiresAt: now.Add(time.Hour).Format(time.RFC3339), RemainingDelegationDepth: 1,
	}
	root, err := SignMandate(rootPayload, keys.root)
	if err != nil {
		return SignedMandate{}, SignedMandate{}, SignedRequest{}, SignedMandate{}, SignedRequest{}, err
	}
	rootDigest, _ := Digest(root)
	childPayload := WorkMandateV1{
		SchemaVersion: MandateSchema, MandateID: "mandate-child-review-h2",
		IssuerPublicKey: PublicKeyHex(keys.delegate.Public().(ed25519.PublicKey)), AudiencePublicKey: PublicKeyHex(keys.worker.Public().(ed25519.PublicKey)), ParentMandateDigest: rootDigest,
		Subject: subject, AllowedActions: []Action{ReviewCandidate},
		IssuedAt: now.Add(-30 * time.Minute).Format(time.RFC3339), ExpiresAt: now.Add(30 * time.Minute).Format(time.RFC3339), RemainingDelegationDepth: 0,
	}
	child, err := SignMandate(childPayload, keys.delegate)
	if err != nil {
		return SignedMandate{}, SignedMandate{}, SignedRequest{}, SignedMandate{}, SignedRequest{}, err
	}
	childDigest, _ := Digest(child)
	validRequest, err := SignRequest(MandateRequestV1{
		SchemaVersion: RequestSchema, RequestID: "request-review-h2", MandateDigest: childDigest,
		Subject: subject, Action: ReviewCandidate, RequestedAt: now.Format(time.RFC3339),
	}, keys.worker)
	if err != nil {
		return SignedMandate{}, SignedMandate{}, SignedRequest{}, SignedMandate{}, SignedRequest{}, err
	}
	widenedPayload := childPayload
	widenedPayload.MandateID = "mandate-child-widened-h2"
	widenedPayload.AllowedActions = []Action{PublishCandidate, ReviewCandidate}
	widenedChild, err := SignMandate(widenedPayload, keys.delegate)
	if err != nil {
		return SignedMandate{}, SignedMandate{}, SignedRequest{}, SignedMandate{}, SignedRequest{}, err
	}
	widenedDigest, _ := Digest(widenedChild)
	publishRequest, err := SignRequest(MandateRequestV1{
		SchemaVersion: RequestSchema, RequestID: "request-publish-h2", MandateDigest: widenedDigest,
		Subject: subject, Action: PublishCandidate, RequestedAt: now.Format(time.RFC3339),
	}, keys.worker)
	return root, child, validRequest, widenedChild, publishRequest, err
}

func selectH2(deck frozenDeck) (Subject, OracleAssurance, error) {
	var subject Subject
	for _, contract := range deck.Contracts {
		if contract.ContractID == "contract-h2" {
			subject = contract.Subject
		}
	}
	for _, snapshot := range deck.OracleAssuranceSnapshots {
		if snapshot.CaseID == "refreshed-h2" {
			oracle := OracleAssurance(snapshot)
			if oracle.MergeAuthority != "none" || oracle.Summary != "SUPPORTED" || oracle.Coverage != "3/3 required claims supported" || oracle.SubjectHead != subject.HeadSHA {
				return Subject{}, OracleAssurance{}, errors.New("unexpected H2 oracle snapshot")
			}
			return subject, oracle, nil
		}
	}
	return Subject{}, OracleAssurance{}, errors.New("contract-h2/refreshed-h2 missing")
}

func SourceLines(repoRoot string) (int, error) {
	total := 0
	err := filepath.WalkDir(repoRoot, func(path string, entry os.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if entry.IsDir() && (entry.Name() == ".git" || entry.Name() == "fixtures") {
			return filepath.SkipDir
		}
		if entry.IsDir() || !strings.HasSuffix(path, ".go") || strings.HasSuffix(path, "_test.go") {
			return nil
		}
		b, err := os.ReadFile(path)
		if err != nil {
			return err
		}
		for _, line := range strings.Split(string(b), "\n") {
			trimmed := strings.TrimSpace(line)
			if trimmed != "" && !strings.HasPrefix(trimmed, "//") {
				total++
			}
		}
		return nil
	})
	return total, err
}
