package phajay

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"sync/atomic"
	"testing"
)

func TestSandboxEndpoints(t *testing.T) {
	for _, tt := range []struct {
		name, path string
		bank       Bank
	}{
		{"link", "/v1/api/test/payment/get-payment-link", 0},
		{"JDB", "/v1/api/test/payment/generate-jdb-qr", BankJDB},
		{"LDB", "/v1/api/test/payment/generate-ldb-qr", BankLDB},
		{"IB", "/v1/api/test/payment/generate-ib-qr", BankIB},
		{"BCEL", "/v1/api/test/payment/generate-bcel-qr", BankBCEL},
		{"STB", "/v1/api/test/payment/generate-stb-qr", BankSTB},
	} {
		t.Run(tt.name, func(t *testing.T) {
			srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				if r.Method != http.MethodPost || r.URL.Path != tt.path {
					t.Errorf("request = %s %s", r.Method, r.URL.Path)
				}
				var body map[string]any
				if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
					t.Error(err)
				}
				if body["amount"] != float64(10) || body["description"] != "Order" || body["orderNo"] != "ORD-1" {
					t.Errorf("body = %#v", body)
				}
				if tt.bank == 0 {
					if want := "Basic " + base64.StdEncoding.EncodeToString([]byte("test-key")); r.Header.Get("Authorization") != want {
						t.Error("incorrect link auth")
					}
					_, _ = io.WriteString(w, `{"message":"SUCCESSFULLY","redirectURL":"https://pay.example/test","orderNo":"ORD-1"}`)
				} else {
					if r.Header.Get("secretKey") != "test-key" {
						t.Error("incorrect QR auth")
					}
					_, _ = io.WriteString(w, `{"message":"SUCCESS","qrCode":"QR","transactionId":"txn","link":"phaJay://qr/test"}`)
				}
			}))
			defer srv.Close()
			p := New("test-key", WithSandbox(), WithBaseURL(srv.URL))
			if tt.bank == 0 {
				got, err := p.CreatePaymentLink(context.Background(), PaymentLinkRequest{OrderNo: "ORD-1", Amount: 10, Description: "Order"})
				if err != nil || got.RedirectURL != "https://pay.example/test" {
					t.Fatalf("response = %#v, error = %v", got, err)
				}
			} else {
				got, err := p.GenerateQR(context.Background(), tt.bank, PaymentQRRequest{OrderNo: "ORD-1", Amount: 10, Description: "Order"})
				if err != nil || got.QRCode != "QR" || got.TransactionID != "txn" {
					t.Fatalf("response = %#v, error = %v", got, err)
				}
			}
		})
	}
}

func TestSandboxUnsupportedBank(t *testing.T) {
	var calls atomic.Int32
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) { calls.Add(1) }))
	defer srv.Close()
	p := New("test-key", WithSandbox(), WithBaseURL(srv.URL))
	for _, bank := range []Bank{BankMMoneyX, Bank(0), Bank(100)} {
		if _, err := p.GenerateQR(context.Background(), bank, PaymentQRRequest{Amount: 10, Description: "Order"}); err == nil {
			t.Errorf("bank %d accepted", bank)
		}
	}
	if calls.Load() != 0 {
		t.Error("unsupported bank sent an HTTP request")
	}
}
