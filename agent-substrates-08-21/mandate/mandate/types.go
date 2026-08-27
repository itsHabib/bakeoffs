package mandate

import (
	"bytes"
	"crypto/ed25519"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"slices"
	"strings"
	"time"
)

const MandateSchema, RequestSchema = "work-mandate/v1", "mandate-request/v1"
const ReceiptSchema, EnvelopeSchema = "mandate-decision-receipt/v1", "agent-substrate-demo/v1"

type Action string

const (
	InspectCandidate Action = "inspect_candidate"
	ReviewCandidate  Action = "review_candidate"
	PublishCandidate Action = "publish_candidate"
)

type Subject struct {
	TaskID       string `json:"task_id"`
	TaskRevision int    `json:"task_revision"`
	Repository   string `json:"repository"`
	BaseSHA      string `json:"base_sha"`
	HeadSHA      string `json:"head_sha"`
	DiffDigest   string `json:"diff_digest"`
}

type WorkMandateV1 struct {
	SchemaVersion            string   `json:"schema_version"`
	MandateID                string   `json:"mandate_id"`
	IssuerPublicKey          string   `json:"issuer_public_key"`
	AudiencePublicKey        string   `json:"audience_public_key"`
	ParentMandateDigest      string   `json:"parent_mandate_digest,omitempty"`
	Subject                  Subject  `json:"subject"`
	AllowedActions           []Action `json:"allowed_actions"`
	IssuedAt                 string   `json:"issued_at"`
	ExpiresAt                string   `json:"expires_at"`
	RemainingDelegationDepth int      `json:"remaining_delegation_depth"`
	ArtifactDigests          []string `json:"artifact_digests,omitempty"`
}

type SignedMandate struct {
	Payload   WorkMandateV1 `json:"payload"`
	Signature string        `json:"signature"`
}

type MandateRequestV1 struct {
	SchemaVersion string  `json:"schema_version"`
	RequestID     string  `json:"request_id"`
	MandateDigest string  `json:"mandate_digest"`
	Subject       Subject `json:"subject"`
	Action        Action  `json:"action"`
	RequestedAt   string  `json:"requested_at"`
}

type SignedRequest struct {
	Payload   MandateRequestV1 `json:"payload"`
	Signature string           `json:"signature"`
}

type Decision string

const Allow, Deny Decision = "allow", "deny"

type MandateDecisionReceiptV1 struct {
	SchemaVersion        string   `json:"schema_version"`
	ReceiptID            string   `json:"receipt_id"`
	VerifiedChainDigests []string `json:"verified_chain_digests"`
	RequestDigest        string   `json:"request_digest"`
	Subject              Subject  `json:"subject"`
	Action               Action   `json:"action"`
	Decision             Decision `json:"decision"`
	ReasonCode           string   `json:"reason_code"`
}

type OracleAssurance struct {
	CaseID         string   `json:"case_id"`
	SubjectHead    string   `json:"subject_head"`
	ContractID     string   `json:"contract_id"`
	Summary        string   `json:"summary"`
	Coverage       string   `json:"coverage"`
	Gaps           []string `json:"gaps"`
	MergeAuthority string   `json:"merge_authority"`
}

type DemoEnvelope struct {
	SchemaVersion        string          `json:"schema_version"`
	Entry                string          `json:"entry"`
	CaseID               string          `json:"case_id"`
	InputBundleDigest    string          `json:"input_bundle_digest"`
	Subject              Subject         `json:"subject"`
	NativeStatus         string          `json:"native_status"`
	ReasonCodes          []string        `json:"reason_codes"`
	NativeArtifactDigest string          `json:"native_artifact_digest"`
	OracleAssurance      OracleAssurance `json:"oracle_assurance"`
	SourceLines          int             `json:"source_lines"`
}

func EncodeCanonical(v any) ([]byte, error) {
	b, err := json.Marshal(v)
	if err != nil {
		return nil, fmt.Errorf("canonical JSON: %w", err)
	}
	return b, nil
}

func DecodeStrict(data []byte, dst any) error {
	dec := json.NewDecoder(bytes.NewReader(data))
	dec.DisallowUnknownFields()
	if err := dec.Decode(dst); err != nil {
		return fmt.Errorf("strict JSON: %w", err)
	}
	if err := dec.Decode(&struct{}{}); !errors.Is(err, io.EOF) {
		return errors.New("strict JSON: trailing data")
	}
	return nil
}

func DigestBytes(b []byte) string {
	sum := sha256.Sum256(b)
	return "sha256:" + hex.EncodeToString(sum[:])
}

func Digest(v any) (string, error) {
	b, err := EncodeCanonical(v)
	if err != nil {
		return "", err
	}
	return DigestBytes(b), nil
}

func SignMandate(payload WorkMandateV1, key ed25519.PrivateKey) (SignedMandate, error) {
	if err := validateMandateShape(payload); err != nil {
		return SignedMandate{}, err
	}
	b, err := EncodeCanonical(payload)
	if err != nil {
		return SignedMandate{}, err
	}
	return SignedMandate{Payload: payload, Signature: hex.EncodeToString(ed25519.Sign(key, b))}, nil
}

func SignRequest(payload MandateRequestV1, key ed25519.PrivateKey) (SignedRequest, error) {
	if err := validateRequestShape(payload); err != nil {
		return SignedRequest{}, err
	}
	b, err := EncodeCanonical(payload)
	if err != nil {
		return SignedRequest{}, err
	}
	return SignedRequest{Payload: payload, Signature: hex.EncodeToString(ed25519.Sign(key, b))}, nil
}

func PublicKeyHex(key ed25519.PublicKey) string { return hex.EncodeToString(key) }

func ParsePublicKey(s string) (ed25519.PublicKey, error) {
	b, err := hex.DecodeString(s)
	if err != nil || len(b) != ed25519.PublicKeySize {
		return nil, errors.New("invalid public key")
	}
	return ed25519.PublicKey(b), nil
}

func parseTimestamp(s string) (time.Time, error) {
	t, err := time.Parse(time.RFC3339, s)
	if err != nil || t.Format(time.RFC3339) != s {
		return time.Time{}, errors.New("timestamp must be canonical RFC3339 UTC")
	}
	return t, nil
}

func validateSubject(s Subject) error {
	if s.TaskID == "" || s.TaskRevision < 1 || s.Repository == "" || s.BaseSHA == "" || s.HeadSHA == "" || s.DiffDigest == "" {
		return errors.New("incomplete subject")
	}
	return nil
}

func validateMandateShape(m WorkMandateV1) error {
	if m.SchemaVersion != MandateSchema || m.MandateID == "" || m.IssuerPublicKey == "" || m.AudiencePublicKey == "" {
		return errors.New("invalid mandate identity")
	}
	if _, err := ParsePublicKey(m.IssuerPublicKey); err != nil {
		return err
	}
	if _, err := ParsePublicKey(m.AudiencePublicKey); err != nil {
		return err
	}
	if err := validateSubject(m.Subject); err != nil {
		return err
	}
	if len(m.AllowedActions) == 0 || !slices.IsSorted(m.AllowedActions) {
		return errors.New("actions must be non-empty and sorted")
	}
	for i, action := range m.AllowedActions {
		if action != InspectCandidate && action != ReviewCandidate && action != PublishCandidate {
			return errors.New("unknown action")
		}
		if i > 0 && action == m.AllowedActions[i-1] {
			return errors.New("duplicate action")
		}
	}
	issued, err := parseTimestamp(m.IssuedAt)
	if err != nil {
		return err
	}
	expires, err := parseTimestamp(m.ExpiresAt)
	if err != nil || !expires.After(issued) {
		return errors.New("invalid validity interval")
	}
	if m.RemainingDelegationDepth < 0 {
		return errors.New("invalid delegation depth")
	}
	if !slices.IsSorted(m.ArtifactDigests) {
		return errors.New("artifact digests must be sorted")
	}
	for i, digest := range m.ArtifactDigests {
		if !strings.HasPrefix(digest, "sha256:") || len(digest) != 71 {
			return errors.New("invalid artifact digest")
		}
		if i > 0 && digest == m.ArtifactDigests[i-1] {
			return errors.New("duplicate artifact digest")
		}
	}
	return nil
}

func validateRequestShape(r MandateRequestV1) error {
	if r.SchemaVersion != RequestSchema || r.RequestID == "" || r.MandateDigest == "" {
		return errors.New("invalid request identity")
	}
	if err := validateSubject(r.Subject); err != nil {
		return err
	}
	if r.Action != InspectCandidate && r.Action != ReviewCandidate && r.Action != PublishCandidate {
		return errors.New("unknown action")
	}
	_, err := parseTimestamp(r.RequestedAt)
	return err
}

func verifySignature(publicHex string, payload any, signatureHex string) bool {
	key, err := ParsePublicKey(publicHex)
	if err != nil {
		return false
	}
	sig, err := hex.DecodeString(signatureHex)
	if err != nil || len(sig) != ed25519.SignatureSize {
		return false
	}
	b, err := EncodeCanonical(payload)
	return err == nil && ed25519.Verify(key, b, sig)
}
