package phajay

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestCreatePaymentLink(t *testing.T) {
	var (
		gotPath   string
		gotAuth   string
		gotBody   PaymentLinkRequest
		gotHeader string
	)

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotPath = r.URL.Path
		gotAuth = r.Header.Get("Authorization")
		gotHeader = r.Header.Get("Content-Type")

		data, _ := io.ReadAll(r.Body)
		_ = json.Unmarshal(data, &gotBody)

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(`{"message":"success","redirectURL":"https://pay.phajay.co/x","orderNo":"ORD-001"}`))
	}))
	defer srv.Close()

	p := New("test-key", WithBaseURL(srv.URL))
	resp, err := p.CreatePaymentLink(context.Background(), PaymentLinkRequest{
		OrderNo:     "ORD-001",
		Amount:      1500.50,
		Description: "Test order",
		Tag1:        "tag-one",
	})
	if err != nil {
		t.Fatalf("CreatePaymentLink() error = %v", err)
	}

	if gotPath != createPaymentLinkPath {
		t.Errorf("path = %q, want %q", gotPath, createPaymentLinkPath)
	}
	if want := "Basic " + base64.StdEncoding.EncodeToString([]byte("test-key")); gotAuth != want {
		t.Errorf("Authorization = %q, want %q", gotAuth, want)
	}
	if gotHeader != "application/json" {
		t.Errorf("Content-Type = %q, want %q", gotHeader, "application/json")
	}
	if gotBody.Amount != 1500.50 {
		t.Errorf("body amount = %v, want 1500.50", gotBody.Amount)
	}
	if gotBody.Description != "Test order" {
		t.Errorf("body description = %q, want %q", gotBody.Description, "Test order")
	}
	if resp.RedirectURL != "https://pay.phajay.co/x" {
		t.Errorf("RedirectURL = %q, want %q", resp.RedirectURL, "https://pay.phajay.co/x")
	}
	if resp.OrderNo != "ORD-001" {
		t.Errorf("OrderNo = %q, want %q", resp.OrderNo, "ORD-001")
	}
}

func TestCreatePaymentLinkValidation(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))
	defer srv.Close()

	p := New("test-key", WithBaseURL(srv.URL))

	tests := []struct {
		name    string
		request PaymentLinkRequest
	}{
		{name: "zero amount", request: PaymentLinkRequest{Amount: 0, Description: "x"}},
		{name: "empty description", request: PaymentLinkRequest{Amount: 10}},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if _, err := p.CreatePaymentLink(context.Background(), tt.request); err == nil {
				t.Error("expected validation error, got nil")
			}
		})
	}
}

func TestCreatePaymentLinkServerError(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
		_, _ = w.Write([]byte(`{"message":"boom"}`))
	}))
	defer srv.Close()

	p := New("test-key", WithBaseURL(srv.URL))
	_, err := p.CreatePaymentLink(context.Background(), PaymentLinkRequest{Amount: 10, Description: "x"})
	if err == nil {
		t.Fatal("expected error, got nil")
	}
	if got, want := err.Error(), "500"; !strings.Contains(got, want) {
		t.Errorf("error = %q, want to contain %q", got, want)
	}
}
