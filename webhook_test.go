package phajay

import (
	"encoding/json"
	"errors"
	"net/http"
	"reflect"
	"strings"
	"testing"
)

// The official public guide does not specify a signature header or algorithm.
// Candidate headers below are deliberately hypothetical, not Phajay fixtures.
// Add valid-signature vectors only after obtaining the private specification.
func TestWebhookVerificationPolicy(t *testing.T) {
	for _, tt := range []struct {
		name    string
		config  Config
		headers http.Header
		body    string
		want    error
		opts    []Option
	}{
		{"zero config", Config{}, nil, `{}`, ErrWebhookSignatureKeyMissing, nil},
		{"API key is not signing key", Config{APIKey: "api-key"}, nil, `{}`, ErrWebhookSignatureKeyMissing, nil},
		{"missing key with candidate signature", Config{APIKey: "api-key"}, http.Header{"X-Unspecified-Signature": {"candidate"}}, `{}`, ErrWebhookSignatureKeyMissing, nil},
		{"blank signing key", Config{APIKey: "api-key", WebhookSignatureKey: " \t\n"}, nil, `{}`, ErrWebhookSignatureKeyMissing, nil},
		{"missing signature", Config{WebhookSignatureKey: "signing-key"}, nil, `{}`, ErrWebhookSignatureVerificationUnavailable, nil},
		{"malformed candidate signature", Config{WebhookSignatureKey: "signing-key"}, http.Header{"X-Unspecified-Signature": {"%%%"}}, `{}`, ErrWebhookSignatureVerificationUnavailable, nil},
		{"invalid candidate signature", Config{WebhookSignatureKey: "signing-key"}, http.Header{"X-Unspecified-Signature": {strings.Repeat("0", 64)}}, `{}`, ErrWebhookSignatureVerificationUnavailable, nil},
		{"multiple candidate signatures", Config{WebhookSignatureKey: "signing-key"}, http.Header{"X-Unspecified-Signature": {"one", "two"}}, `{}`, ErrWebhookSignatureVerificationUnavailable, nil},
		{"verification before JSON parsing", Config{WebhookSignatureKey: "signing-key"}, nil, `not JSON`, ErrWebhookSignatureVerificationUnavailable, nil},
		{"sandbox does not bypass", Config{APIKey: "test-key", WebhookSignatureKey: "signing-key"}, nil, `{}`, ErrWebhookSignatureVerificationUnavailable, []Option{WithSandbox()}},
		{"explicit bypass without key or signature", Config{BypassSignatureVerification: true}, nil, `{}`, nil, nil},
		{"explicit bypass with configured signing key", Config{WebhookSignatureKey: "signing-key", BypassSignatureVerification: true}, nil, `{}`, nil, nil},
		{"explicit bypass with blank key and malformed signature", Config{WebhookSignatureKey: " ", BypassSignatureVerification: true}, http.Header{"X-Unspecified-Signature": {"%%%"}}, `{}`, nil, nil},
	} {
		t.Run(tt.name, func(t *testing.T) {
			p := New(tt.config, tt.opts...)
			body := []byte(tt.body)
			if err := p.VerifyWebhookSignature(body, tt.headers); !errors.Is(err, tt.want) {
				t.Fatalf("verification error = %v, want %v", err, tt.want)
			}
			cb := PaymentQRWebhookCallback{TransactionID: "unchanged"}
			before := cb
			if err := p.DecodeWebhook(body, tt.headers, &cb); !errors.Is(err, tt.want) {
				t.Fatalf("decode error = %v, want %v", err, tt.want)
			}
			if tt.want != nil && !reflect.DeepEqual(cb, before) {
				t.Errorf("verification failure modified callback: %#v", cb)
			}
			if string(body) != tt.body {
				t.Error("modified original body")
			}
		})
	}
}

func TestDecodeWebhookBypassCallbackTypes(t *testing.T) {
	p := New(Config{BypassSignatureVerification: true})
	for _, tt := range []struct {
		name, body string
		callback   any
		want       any
	}{
		{"link", ` {"transactionId":"link","txnAmount":12,"refNo":9007199254740993} `,
			&PaymentLinkWebhookCallback{}, &PaymentLinkWebhookCallback{TransactionID: "link", TxnAmount: 12, RefNo: testPtr("9007199254740993")}},
		{"QR", `{"paymentMethod":"BCEL","transactionId":"qr","txnAmount":12,"refNo":9007199254740993,"status":null}`,
			&PaymentQRWebhookCallback{}, &PaymentQRWebhookCallback{PaymentMethod: "BCEL", TransactionID: "qr", TxnAmount: 12, RefNo: testPtr(FlexibleString("9007199254740993"))}},
		{"card", `{"transactionId":"card","creditCardDetails":{"cardNumber":"masked"}}`,
			&CreditCardWebhookCallback{}, &CreditCardWebhookCallback{TransactionID: testPtr("card"), CreditCardDetails: &CreditCardDetails{CardNumber: testPtr("masked")}}},
		{"subscription", `{"transactionId":"sub","status":"SUBSCRIPTION_SUCCESS","paymentTransactionId":"payment"}`,
			&SubscriptionWebhookCallback{}, &SubscriptionWebhookCallback{TransactionID: "sub", Status: SubscriptionStatusSuccess, PaymentTransactionID: testPtr("payment")}},
		{"sandbox link", `{"_id":"sandbox","amount":12,"descriptione":"Order"}`,
			&SandboxPaymentLinkWebhookCallback{}, &SandboxPaymentLinkWebhookCallback{ID: testPtr("sandbox"), Amount: testPtr(float64(12)), Description: testPtr("Order")}},
	} {
		t.Run(tt.name, func(t *testing.T) {
			if err := p.DecodeWebhook([]byte(tt.body), nil, tt.callback); err != nil {
				t.Fatal(err)
			}
			if !reflect.DeepEqual(tt.callback, tt.want) {
				t.Errorf("callback = %#v, want %#v", tt.callback, tt.want)
			}
		})
	}
}

func TestDecodeWebhookJSONErrors(t *testing.T) {
	p := New(Config{BypassSignatureVerification: true})
	for _, body := range []string{"", "{", "{} {}"} {
		t.Run(body, func(t *testing.T) {
			var cb PaymentQRWebhookCallback
			err := p.DecodeWebhook([]byte(body), nil, &cb)
			var syntaxErr *json.SyntaxError
			if !errors.As(err, &syntaxErr) || !strings.HasPrefix(err.Error(), "phajay: ") {
				t.Fatalf("error = %v, want wrapped JSON syntax error", err)
			}
		})
	}
	for _, callback := range []any{nil, (*PaymentQRWebhookCallback)(nil), PaymentQRWebhookCallback{}} {
		err := p.DecodeWebhook([]byte(`{}`), nil, callback)
		var targetErr *json.InvalidUnmarshalError
		if !errors.As(err, &targetErr) {
			t.Fatalf("error = %v, want wrapped invalid target error", err)
		}
	}
}
