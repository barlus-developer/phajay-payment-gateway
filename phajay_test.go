package phajay

import (
	"context"
	"errors"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"
)

func TestClientConfig(t *testing.T) {
	for _, tt := range []struct {
		name    string
		config  Config
		opts    []Option
		baseURL string
	}{
		{"zero config", Config{}, nil, defaultBaseURL},
		{"API and signing keys", Config{APIKey: "api-key", WebhookSignatureKey: "signing-key"}, nil, defaultBaseURL},
		{"custom base URL", Config{BaseURL: "https://example.com"}, nil, "https://example.com"},
		{"option overrides config URL", Config{BaseURL: "https://example.com"}, []Option{WithBaseURL("https://override.example.com")}, "https://override.example.com"},
		{"explicit bypass", Config{BypassSignatureVerification: true}, nil, defaultBaseURL},
	} {
		t.Run(tt.name, func(t *testing.T) {
			p := New(tt.config, tt.opts...)
			if p.key != tt.config.APIKey || p.webhookSignatureKey != tt.config.WebhookSignatureKey ||
				p.bypassSignatureVerification != tt.config.BypassSignatureVerification || p.baseURL != tt.baseURL {
				t.Fatal("client configuration was not applied correctly")
			}
		})
	}
}

func TestClientOptions(t *testing.T) {
	original := &http.Client{Timeout: time.Second}
	for _, tt := range []struct {
		name    string
		opts    []Option
		timeout time.Duration
		sandbox bool
	}{
		{"defaults", nil, 30 * time.Second, false},
		{"custom timeout", []Option{WithTimeout(5 * time.Second)}, 5 * time.Second, false},
		{"custom client", []Option{WithHTTPClient(original)}, time.Second, false},
		{"timeout after client", []Option{WithHTTPClient(original), WithTimeout(5 * time.Second)}, 5 * time.Second, false},
		{"client after timeout", []Option{WithTimeout(5 * time.Second), WithHTTPClient(original)}, time.Second, false},
		{"nil client", []Option{WithHTTPClient(nil)}, 30 * time.Second, false},
		{"sandbox", []Option{WithSandbox()}, 30 * time.Second, true},
	} {
		t.Run(tt.name, func(t *testing.T) {
			p := New(Config{APIKey: "key"}, tt.opts...)
			if p.client.Timeout != tt.timeout || p.sandbox != tt.sandbox || p.baseURL != "https://payment-gateway.phajay.co" {
				t.Fatalf("client = %#v", p)
			}
			if original.Timeout != time.Second {
				t.Error("option modified supplied client")
			}
		})
	}
}

type testTransport func(*http.Request) (*http.Response, error)

func (f testTransport) RoundTrip(r *http.Request) (*http.Response, error) { return f(r) }

func TestClientWrapsUnderlyingErrors(t *testing.T) {
	sentinel := errors.New("connection failed")
	p := New(Config{APIKey: "key"}, WithHTTPClient(&http.Client{Transport: testTransport(func(r *http.Request) (*http.Response, error) { return nil, sentinel })}))
	if _, err := p.CreatePaymentLink(context.Background(), PaymentLinkRequest{Amount: 1, Description: "x"}); !errors.Is(err, sentinel) {
		t.Errorf("error = %v", err)
	}
	p = New(Config{APIKey: "key"}, WithBaseURL("://invalid"))
	if _, err := p.GenerateQR(context.Background(), BankJDB, PaymentQRRequest{Amount: 1, Description: "x"}); err == nil || !strings.Contains(err.Error(), "build") {
		t.Errorf("error = %v", err)
	}
}

func TestAPIErrorBodyLimit(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusBadRequest)
		_, _ = io.WriteString(w, strings.Repeat("x", (1<<20)+100))
	}))
	defer srv.Close()
	_, err := New(Config{APIKey: "key"}, WithBaseURL(srv.URL)).CreatePaymentLink(context.Background(), PaymentLinkRequest{Amount: 1, Description: "x"})
	var apiErr *APIError
	if !errors.As(err, &apiErr) || len(apiErr.Body) != 1<<20 || apiErr.StatusCode != 400 {
		t.Fatalf("error = %v", err)
	}
}
