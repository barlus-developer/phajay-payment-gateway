package phajay

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"strings"
)

var (
	// ErrWebhookSignatureKeyMissing indicates verification is required but no
	// non-blank webhook signing key was configured. The API key is never used.
	ErrWebhookSignatureKeyMissing = errors.New("phajay: webhook signature key is missing")
	// ErrWebhookSignatureVerificationUnavailable indicates the private Phajay
	// signing specification is unavailable, so the webhook cannot be verified.
	ErrWebhookSignatureVerificationUnavailable = errors.New("phajay: webhook signature verification unavailable: obtain the signing specification from Phajay support")
)

// VerifyWebhookSignature is the verification entry point for the original
// request body and HTTP headers.
// Until Phajay's private signing specification is available, this method fails
// closed: it returns ErrWebhookSignatureKeyMissing for an empty or blank key,
// otherwise ErrWebhookSignatureVerificationUnavailable for every request.
// No algorithm, signature encoding, signed payload, or header name is assumed.
// See https://payment-doc.phajay.co/v1/verify-webhook-signature.
//
// With BypassSignatureVerification explicitly enabled, it returns nil without
// inspecting the key, body, or headers. That result does not authenticate the
// webhook; bypass is intended only for development and tests.
func (p *Phajay) VerifyWebhookSignature(rawBody []byte, headers http.Header) error {
	if p.bypassSignatureVerification {
		return nil
	}
	if strings.TrimSpace(p.webhookSignatureKey) == "" {
		return ErrWebhookSignatureKeyMissing
	}
	return ErrWebhookSignatureVerificationUnavailable
}

// DecodeWebhook verifies the original request bytes before decoding a single
// JSON payload into callback, which must be a non-nil pointer (for example,
// *PaymentQRWebhookCallback). Read the request body once and pass its unchanged
// bytes and headers. A verification failure leaves callback untouched.
//
// Verification currently fails closed because Phajay's signing specification
// is private. Explicit bypass allows decoding unauthenticated data only for
// development or tests. Direct json.Unmarshal also does not authenticate data.
func (p *Phajay) DecodeWebhook(rawBody []byte, headers http.Header, callback any) error {
	if err := p.VerifyWebhookSignature(rawBody, headers); err != nil {
		return err
	}
	if err := json.Unmarshal(rawBody, callback); err != nil {
		return fmt.Errorf("phajay: decode webhook: %w", err)
	}
	return nil
}
