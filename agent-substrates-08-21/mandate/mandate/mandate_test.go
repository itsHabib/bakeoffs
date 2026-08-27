package mandate

import (
	"crypto/ed25519"
	"os"
	"path/filepath"
	"testing"
	"time"
)

type chainFixture struct {
	now                          time.Time
	keys                         demoKeys
	root, child, widened         SignedMandate
	validRequest, publishRequest SignedRequest
}

func loadChain(t *testing.T) chainFixture {
	t.Helper()
	repoRoot := filepath.Clean("..")
	keys, err := loadKeys(filepath.Join(repoRoot, "fixtures", "keys.json"))
	if err != nil {
		t.Fatal(err)
	}
	result, err := BuildDemo(repoRoot)
	if err != nil {
		t.Fatal(err)
	}
	now := time.Date(2026, 8, 21, 12, 0, 0, 0, time.UTC)
	root, child, request, widened, publish, err := demoChain(result.Envelope.Subject, keys, now)
	if err != nil {
		t.Fatal(err)
	}
	return chainFixture{now, keys, root, child, widened, request, publish}
}

func TestFixedKeyCanonicalStability(t *testing.T) {
	f := loadChain(t)
	payloadDigest, err := Digest(f.root.Payload)
	if err != nil {
		t.Fatal(err)
	}
	const wantPayloadDigest = "sha256:defcb49837b99ccac7825c4db58699d56ed7e0c50b64e326651a99a740fe4f41"
	const wantSignature = "3eb47f7d11af3fde5f782f4dda1b09756cfa1ed2b00253ca790796e3050c2c793a549e3b0aa3310b54075fb379176861bc756b869914693476f9766a86ab440c"
	const wantSignedDigest = "sha256:4cf75d412845855aa3c5bcbd72ce0d87d48d32479cb2a4e4948ecd1110eb3bda"
	signedDigest, err := Digest(f.root)
	if err != nil {
		t.Fatal(err)
	}
	if payloadDigest != wantPayloadDigest || f.root.Signature != wantSignature || signedDigest != wantSignedDigest {
		t.Fatalf("canonical fixture changed\npayload digest: %s\nsignature: %s\nsigned digest: %s", payloadDigest, f.root.Signature, signedDigest)
	}
}

func TestValidChainAndGoldenArtifacts(t *testing.T) {
	f := loadChain(t)
	receipt := VerifyChainAndRequest(f.root, f.child, f.validRequest, f.keys.root.Public().(ed25519.PublicKey), f.now)
	if receipt.Decision != Allow || receipt.ReasonCode != "authorized" {
		t.Fatalf("unexpected receipt: %#v", receipt)
	}
	assertGolden(t, filepath.Join("..", "fixtures", "golden", "allow-receipt.json"), receipt)
	result, err := BuildDemo("..")
	if err != nil {
		t.Fatal(err)
	}
	assertGolden(t, filepath.Join("..", "fixtures", "golden", "demo-envelope.json"), result.Envelope)
	first, _ := EncodeCanonical(result.Envelope)
	secondResult, err := BuildDemo("..")
	if err != nil {
		t.Fatal(err)
	}
	second, _ := EncodeCanonical(secondResult.Envelope)
	if string(first) != string(second) {
		t.Fatal("identical demo runs were not byte-identical")
	}
}

func TestSingleLawMutant(t *testing.T) {
	f := loadChain(t)
	rootKey := f.keys.root.Public().(ed25519.PublicKey)
	mutant := verifyChainAndRequest(f.root, f.widened, f.publishRequest, rootKey, f.now, verifyChildSignaturesOnly)
	production := VerifyChainAndRequest(f.root, f.widened, f.publishRequest, rootKey, f.now)
	if mutant.Decision != Allow || mutant.ReasonCode != "authorized" {
		t.Fatalf("mutant did not expose planted bug: %#v", mutant)
	}
	if production.Decision != Deny || production.ReasonCode != "scope_inflation" {
		t.Fatalf("production did not reject widening: %#v", production)
	}
}

func TestRootSignatureFailsClosed(t *testing.T) {
	f := loadChain(t)
	trusted := f.keys.root.Public().(ed25519.PublicKey)
	if err := VerifyRoot(f.root, trusted, f.now); err != nil {
		t.Fatalf("valid root rejected: %v", err)
	}
	tampered := f.root
	tampered.Signature = "00" + tampered.Signature[2:]
	if err := VerifyRoot(tampered, trusted, f.now); reason(err) != "invalid_root_signature" {
		t.Fatalf("tampered root: %q", reason(err))
	}
	if err := VerifyRoot(f.root, f.keys.delegate.Public().(ed25519.PublicKey), f.now); reason(err) != "untrusted_root" {
		t.Fatalf("wrong trust root: %q", reason(err))
	}
}

func TestChildAttenuationFailures(t *testing.T) {
	f := loadChain(t)
	cases := []struct {
		name string
		edit func(*WorkMandateV1)
		key  ed25519.PrivateKey
		want string
	}{
		{"wrong issuer", func(m *WorkMandateV1) { m.IssuerPublicKey = PublicKeyHex(f.keys.root.Public().(ed25519.PublicKey)) }, f.keys.root, "wrong_child_issuer"},
		{"parent digest", func(m *WorkMandateV1) { m.ParentMandateDigest = FrozenInputDigest }, f.keys.delegate, "parent_digest_mismatch"},
		{"subject", func(m *WorkMandateV1) { m.Subject.HeadSHA = "4444444444444444444444444444444444444444" }, f.keys.delegate, "subject_mismatch"},
		{"artifact", func(m *WorkMandateV1) { m.ArtifactDigests = []string{FrozenInputDigest} }, f.keys.delegate, "artifact_mismatch"},
		{"expiry growth", func(m *WorkMandateV1) { m.ExpiresAt = f.now.Add(2 * time.Hour).Format(time.RFC3339) }, f.keys.delegate, "expiry_widening"},
		{"depth", func(m *WorkMandateV1) { m.RemainingDelegationDepth = 1 }, f.keys.delegate, "depth_not_decremented"},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			payload := f.child.Payload
			tc.edit(&payload)
			child, err := SignMandate(payload, tc.key)
			if err != nil {
				t.Fatal(err)
			}
			if err := VerifyChild(f.root, child, f.now); reason(err) != tc.want {
				t.Fatalf("got %q want %q", reason(err), tc.want)
			}
		})
	}
	if err := VerifyChild(f.root, f.child, f.now.Add(45*time.Minute)); reason(err) != "expired" {
		t.Fatalf("expired child: %q", reason(err))
	}
}

func TestRequestFailures(t *testing.T) {
	f := loadChain(t)
	cases := []struct {
		name string
		edit func(*MandateRequestV1)
		key  ed25519.PrivateKey
		want string
	}{
		{"wrong audience signature", func(*MandateRequestV1) {}, f.keys.delegate, "invalid_request_signature"},
		{"mandate digest", func(r *MandateRequestV1) { r.MandateDigest = FrozenInputDigest }, f.keys.worker, "request_mandate_mismatch"},
		{"subject", func(r *MandateRequestV1) {
			r.Subject.DiffDigest = "sha256:aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa"
		}, f.keys.worker, "request_subject_mismatch"},
		{"action", func(r *MandateRequestV1) { r.Action = InspectCandidate }, f.keys.worker, "action_not_allowed"},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			payload := f.validRequest.Payload
			tc.edit(&payload)
			request, err := SignRequest(payload, tc.key)
			if err != nil {
				t.Fatal(err)
			}
			if err := VerifyRequest(f.child, request, f.now); reason(err) != tc.want {
				t.Fatalf("got %q want %q", reason(err), tc.want)
			}
		})
	}
}

func TestUnknownFieldsAndActionsFailClosed(t *testing.T) {
	var signed SignedMandate
	if err := DecodeStrict([]byte(`{"payload":{},"signature":"00","extra":true}`), &signed); err == nil {
		t.Fatal("unknown field accepted")
	}
	f := loadChain(t)
	payload := f.child.Payload
	payload.AllowedActions = []Action{"delete_repository"}
	if _, err := SignMandate(payload, f.keys.delegate); err == nil {
		t.Fatal("unknown action accepted")
	}
	request := f.validRequest.Payload
	request.Action = "merge_candidate"
	if _, err := SignRequest(request, f.keys.worker); err == nil {
		t.Fatal("unknown request action accepted")
	}
}

func TestAssuranceAloneIsNotAuthority(t *testing.T) {
	result, err := BuildDemo("..")
	if err != nil {
		t.Fatal(err)
	}
	if result.Envelope.OracleAssurance.Summary != "SUPPORTED" || result.Envelope.OracleAssurance.MergeAuthority != "none" {
		t.Fatal("oracle fixture was not copied unchanged")
	}
	if result.WithoutMandate.Decision != Deny || result.WithoutMandate.ReasonCode != "no_authority" {
		t.Fatal("assurance incorrectly granted authority")
	}
}

func assertGolden(t *testing.T, path string, value any) {
	t.Helper()
	want, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	got, err := EncodeCanonical(value)
	if err != nil {
		t.Fatal(err)
	}
	got = append(got, '\n')
	if string(got) != string(want) {
		t.Fatalf("golden mismatch for %s\n got: %s\nwant: %s", path, got, want)
	}
}
