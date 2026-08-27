package mandate

import (
	"crypto/ed25519"
	"fmt"
	"slices"
	"time"
)

type failure struct{ code string }

func (f failure) Error() string { return f.code }

func fail(code string) error { return failure{code: code} }

func reason(err error) string {
	if f, ok := err.(failure); ok {
		return f.code
	}
	return "invalid_format"
}

func VerifyRoot(root SignedMandate, trustedRoot ed25519.PublicKey, now time.Time) error {
	m := root.Payload
	if err := validateMandateShape(m); err != nil {
		return fail("invalid_format")
	}
	if m.IssuerPublicKey != PublicKeyHex(trustedRoot) || m.ParentMandateDigest != "" {
		return fail("untrusted_root")
	}
	if !verifySignature(m.IssuerPublicKey, m, root.Signature) {
		return fail("invalid_root_signature")
	}
	if m.RemainingDelegationDepth < 1 {
		return fail("depth_exhausted")
	}
	return verifyTime(m.IssuedAt, m.ExpiresAt, now)
}

func VerifyChild(parent, child SignedMandate, now time.Time) error {
	return verifyChild(parent, child, now, true)
}

func verifyChild(parent, child SignedMandate, now time.Time, enforceActionSubset bool) error {
	p, c := parent.Payload, child.Payload
	if err := validateMandateShape(c); err != nil {
		return fail("invalid_format")
	}
	if !verifySignature(c.IssuerPublicKey, c, child.Signature) {
		return fail("invalid_child_signature")
	}
	if c.IssuerPublicKey != p.AudiencePublicKey {
		return fail("wrong_child_issuer")
	}
	parentDigest, err := Digest(parent)
	if err != nil || c.ParentMandateDigest != parentDigest {
		return fail("parent_digest_mismatch")
	}
	if c.Subject != p.Subject {
		return fail("subject_mismatch")
	}
	if !slices.Equal(c.ArtifactDigests, p.ArtifactDigests) {
		return fail("artifact_mismatch")
	}
	if enforceActionSubset && !isSubset(c.AllowedActions, p.AllowedActions) {
		return fail("scope_inflation")
	}
	pIssued, _ := parseTimestamp(p.IssuedAt)
	pExpires, _ := parseTimestamp(p.ExpiresAt)
	cIssued, _ := parseTimestamp(c.IssuedAt)
	cExpires, _ := parseTimestamp(c.ExpiresAt)
	if cIssued.Before(pIssued) || cExpires.After(pExpires) {
		return fail("expiry_widening")
	}
	if c.RemainingDelegationDepth != p.RemainingDelegationDepth-1 {
		return fail("depth_not_decremented")
	}
	return verifyTime(c.IssuedAt, c.ExpiresAt, now)
}

func VerifyRequest(child SignedMandate, request SignedRequest, now time.Time) error {
	r := request.Payload
	if err := validateRequestShape(r); err != nil {
		return fail("invalid_format")
	}
	if !verifySignature(child.Payload.AudiencePublicKey, r, request.Signature) {
		return fail("invalid_request_signature")
	}
	childDigest, err := Digest(child)
	if err != nil || r.MandateDigest != childDigest {
		return fail("request_mandate_mismatch")
	}
	if r.Subject != child.Payload.Subject {
		return fail("request_subject_mismatch")
	}
	if !slices.Contains(child.Payload.AllowedActions, r.Action) {
		return fail("action_not_allowed")
	}
	requested, _ := parseTimestamp(r.RequestedAt)
	if !requested.Equal(now) {
		return fail("request_time_mismatch")
	}
	return verifyTime(child.Payload.IssuedAt, child.Payload.ExpiresAt, requested)
}

func VerifyChainAndRequest(root, child SignedMandate, request SignedRequest, trustedRoot ed25519.PublicKey, now time.Time) MandateDecisionReceiptV1 {
	return verifyChainAndRequest(root, child, request, trustedRoot, now, VerifyChild)
}

func verifyChainAndRequest(root, child SignedMandate, request SignedRequest, trustedRoot ed25519.PublicKey, now time.Time, verifyDelegation func(SignedMandate, SignedMandate, time.Time) error) MandateDecisionReceiptV1 {
	chainDigests := make([]string, 0, 2)
	rootDigest, _ := Digest(root)
	childDigest, _ := Digest(child)
	requestDigest, _ := Digest(request)

	decision, code := Deny, "invalid_format"
	if err := VerifyRoot(root, trustedRoot, now); err != nil {
		code = reason(err)
	} else {
		chainDigests = append(chainDigests, rootDigest)
		if err := verifyDelegation(root, child, now); err != nil {
			code = reason(err)
		} else {
			chainDigests = append(chainDigests, childDigest)
			if err := VerifyRequest(child, request, now); err != nil {
				code = reason(err)
			} else {
				decision, code = Allow, "authorized"
			}
		}
	}

	receipt := MandateDecisionReceiptV1{
		SchemaVersion:        ReceiptSchema,
		VerifiedChainDigests: chainDigests,
		RequestDigest:        requestDigest,
		Subject:              request.Payload.Subject,
		Action:               request.Payload.Action,
		Decision:             decision,
		ReasonCode:           code,
	}
	identity, _ := Digest(struct {
		Chain   []string `json:"chain"`
		Request string   `json:"request"`
		Code    string   `json:"code"`
	}{chainDigests, requestDigest, code})
	receipt.ReceiptID = "receipt-" + identity[len("sha256:"):len("sha256:")+16]
	return receipt
}

func DenyWithoutMandate(subject Subject, action Action) MandateDecisionReceiptV1 {
	receipt := MandateDecisionReceiptV1{
		SchemaVersion: ReceiptSchema,
		RequestDigest: "sha256:0000000000000000000000000000000000000000000000000000000000000000",
		Subject:       subject,
		Action:        action,
		Decision:      Deny,
		ReasonCode:    "no_authority",
	}
	identity, _ := Digest(receipt)
	receipt.ReceiptID = "receipt-" + identity[len("sha256:"):len("sha256:")+16]
	return receipt
}

func verifyTime(issuedAt, expiresAt string, now time.Time) error {
	issued, err := parseTimestamp(issuedAt)
	if err != nil {
		return fail("invalid_format")
	}
	expires, err := parseTimestamp(expiresAt)
	if err != nil {
		return fail("invalid_format")
	}
	if now.Before(issued) {
		return fail("not_yet_valid")
	}
	if !now.Before(expires) {
		return fail("expired")
	}
	return nil
}

func isSubset(child, parent []Action) bool {
	for _, action := range child {
		if !slices.Contains(parent, action) {
			return false
		}
	}
	return true
}

func ReceiptDigest(receipt MandateDecisionReceiptV1) string {
	digest, err := Digest(receipt)
	if err != nil {
		panic(fmt.Sprintf("digest receipt: %v", err))
	}
	return digest
}
