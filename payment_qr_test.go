package phajay

import (
	"context"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestGenerateQR(t *testing.T) {
	var (
		gotPath      string
		gotSecretKey string
		gotHeader    string
		gotBody      PaymentQRRequest
	)

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotPath = r.URL.Path
		gotSecretKey = r.Header.Get("secretKey")
		gotHeader = r.Header.Get("Content-Type")

		data, _ := io.ReadAll(r.Body)
		_ = json.Unmarshal(data, &gotBody)

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(`{"message":"SUCCESSFULLY","transactionId":"8cc876b4-a4af-4886-81f1-3890453eb656","qrCode":"0002010102","link":"onepay://qr/0002010102"}`))
	}))
	defer srv.Close()

	p := New("test-key", WithBaseURL(srv.URL))
	resp, err := p.GenerateQR(context.Background(), BankJDB, PaymentQRRequest{
		Amount:      1500.50,
		Description: "Buy Pants",
		Tag1:        "tag-one",
	})
	if err != nil {
		t.Fatalf("GenerateQR() error = %v", err)
	}

	if want := generateJDBQRPath; gotPath != want {
		t.Errorf("path = %q, want %q", gotPath, want)
	}
	if gotSecretKey != "test-key" {
		t.Errorf("secretKey = %q, want %q", gotSecretKey, "test-key")
	}
	if gotHeader != "application/json" {
		t.Errorf("Content-Type = %q, want %q", gotHeader, "application/json")
	}
	if gotBody.Amount != 1500.50 {
		t.Errorf("body amount = %v, want 1500.50", gotBody.Amount)
	}
	if gotBody.Description != "Buy Pants" {
		t.Errorf("body description = %q, want %q", gotBody.Description, "Buy Pants")
	}
	if gotBody.Tag1 != "tag-one" {
		t.Errorf("body tag1 = %q, want %q", gotBody.Tag1, "tag-one")
	}
	if resp.TransactionID != "8cc876b4-a4af-4886-81f1-3890453eb656" {
		t.Errorf("TransactionID = %q, want %q", resp.TransactionID, "8cc876b4-a4af-4886-81f1-3890453eb656")
	}
	if resp.QRCode != "0002010102" {
		t.Errorf("QRCode = %q, want %q", resp.QRCode, "0002010102")
	}
	if resp.Link != "onepay://qr/0002010102" {
		t.Errorf("Link = %q, want %q", resp.Link, "onepay://qr/0002010102")
	}
}

func TestGenerateQRBankPaths(t *testing.T) {
	tests := []struct {
		name string
		bank Bank
		path string
	}{
		{name: "JDB", bank: BankJDB, path: generateJDBQRPath},
		{name: "LDB", bank: BankLDB, path: generateLDBQRPath},
		{name: "IB", bank: BankIB, path: generateIBQRPath},
		{name: "BCEL", bank: BankBCEL, path: generateBCELQRPath},
		{name: "STB", bank: BankSTB, path: generateSTBQRPath},
		{name: "M MoneyX", bank: BankMMoneyX, path: generateMMoneyXQRPath},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var gotPath string
			srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				gotPath = r.URL.Path
				w.Header().Set("Content-Type", "application/json")
				w.WriteHeader(http.StatusOK)
				_, _ = w.Write([]byte(`{"message":"ok"}`))
			}))
			defer srv.Close()

			p := New("test-key", WithBaseURL(srv.URL))
			if _, err := p.GenerateQR(context.Background(), tt.bank, PaymentQRRequest{Amount: 10, Description: "x"}); err != nil {
				t.Fatalf("GenerateQR() error = %v", err)
			}
			if gotPath != tt.path {
				t.Errorf("path = %q, want %q", gotPath, tt.path)
			}
		})
	}
}

func TestGenerateQRValidation(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))
	defer srv.Close()

	p := New("test-key", WithBaseURL(srv.URL))

	tests := []struct {
		name    string
		request PaymentQRRequest
	}{
		{name: "zero amount", request: PaymentQRRequest{Amount: 0, Description: "x"}},
		{name: "empty description", request: PaymentQRRequest{Amount: 10}},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if _, err := p.GenerateQR(context.Background(), BankJDB, tt.request); err == nil {
				t.Error("expected validation error, got nil")
			}
		})
	}
}

func TestGenerateQRUnknownBank(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))
	defer srv.Close()

	p := New("test-key", WithBaseURL(srv.URL))
	_, err := p.GenerateQR(context.Background(), Bank(0), PaymentQRRequest{Amount: 10, Description: "x"})
	if err == nil {
		t.Fatal("expected error, got nil")
	}
	if got, want := err.Error(), "unknown bank"; !strings.Contains(got, want) {
		t.Errorf("error = %q, want to contain %q", got, want)
	}
}

func TestGenerateQRServerError(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
		_, _ = w.Write([]byte(`{"message":"boom"}`))
	}))
	defer srv.Close()

	p := New("test-key", WithBaseURL(srv.URL))
	_, err := p.GenerateQR(context.Background(), BankJDB, PaymentQRRequest{Amount: 10, Description: "x"})
	if err == nil {
		t.Fatal("expected error, got nil")
	}
	if got, want := err.Error(), "500"; !strings.Contains(got, want) {
		t.Errorf("error = %q, want to contain %q", got, want)
	}
}
