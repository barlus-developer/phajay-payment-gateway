package phajay

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"errors"
	"io"
	"math"
	"net/http"
	"net/http/httptest"
	"reflect"
	"strconv"
	"strings"
	"sync/atomic"
	"testing"
)

func testPtr[T any](value T) *T { return &value }

type endpointTest struct {
	name, method, path, body, response string
	basic                              bool
	want                               any
	call                               func(context.Context, *Phajay) (any, error)
}

func documentedEndpoints() []endpointTest {
	return []endpointTest{
		{
			name: "transaction status", method: http.MethodGet,
			path: "/v1/api/payment/check-transaction/status/txn-123", basic: true,
			response: `{"message":"SUCCESSFULLY","data":{"transactionId":"txn-123","linkCode":null,"orderNo":null,"status":"WAITING","isPaid":false,"amount":100000,"currency":"LAK","description":"Order","paymentMethod":"BCEL","paymentTime":null,"createdAt":"2026-09-28T04:17:38.973Z"}}`,
			want: TransactionStatusResponse{Message: "SUCCESSFULLY", Data: &TransactionStatus{
				TransactionID: "txn-123", Status: "WAITING", Amount: 100000, Currency: "LAK",
				Description: "Order", PaymentMethod: "BCEL", CreatedAt: "2026-09-28T04:17:38.973Z",
			}},
			call: func(ctx context.Context, p *Phajay) (any, error) { return p.CheckTransactionStatus(ctx, "txn-123") },
		},
		{
			name: "request refund", method: http.MethodPost, path: "/v1/api/refund",
			body:     `{"transactionId":"txn-123"}`,
			response: `{"message":"SUCCESSFULLY","data":{"id":"bill-123","status":"REQUESTING","amount":1000,"createdAt":"2024-10-02T11:28:43.625Z","billNo":"LPG710301"}}`,
			want: RefundResponse{Message: "SUCCESSFULLY", Data: Refund{
				ID: "bill-123", Status: "REQUESTING", Amount: 1000, CreatedAt: "2024-10-02T11:28:43.625Z", BillNo: "LPG710301",
			}},
			call: func(ctx context.Context, p *Phajay) (any, error) {
				return p.RequestRefund(ctx, RefundRequest{TransactionID: "txn-123"})
			},
		},
		{
			name: "refund status", method: http.MethodGet, path: "/v1/api/refund/bill-123",
			response: `{"message":"SUCCESSFULLY","data":{"id":"bill-123","status":"REQUESTING","amount":"1000","createdAt":"2024-10-02T11:28:43.625Z","billNo":"LPG710301"}}`,
			want: RefundResponse{Message: "SUCCESSFULLY", Data: Refund{
				ID: "bill-123", Status: "REQUESTING", Amount: 1000, CreatedAt: "2024-10-02T11:28:43.625Z", BillNo: "LPG710301",
			}},
			call: func(ctx context.Context, p *Phajay) (any, error) { return p.CheckRefundStatus(ctx, "bill-123") },
		},
		{
			name: "credit card link", method: http.MethodPost,
			path: "/v1/api/jdb2c2p/payment/payment-link", basic: true,
			body:     `{"amount":1000,"description":"Buy a Product","tag1":"Shop","tag2":"Customer","tag3":"ORDER-0001"}`,
			response: `{"message":"SUCCESSFULLY","paymentUrl":"https://payment.example/card","expirationTime":"2025-08-04T07:29:33","transactionId":"txn-123","status":"WAITING"}`,
			want: CreditCardPaymentResponse{Message: "SUCCESSFULLY", PaymentURL: "https://payment.example/card",
				ExpirationTime: "2025-08-04T07:29:33", TransactionID: "txn-123", Status: "WAITING"},
			call: func(ctx context.Context, p *Phajay) (any, error) {
				return p.CreateCreditCardPaymentLink(ctx, CreditCardPaymentRequest{Amount: 1000, Description: "Buy a Product", Tag1: "Shop", Tag2: "Customer", Tag3: "ORDER-0001"})
			},
		},
		{
			name: "subscription QR", method: http.MethodPost,
			path:     "/v1/api/subscription/generate-bcel-qr",
			body:     `{"maxAmount":1000,"subscriptionDate":"2026-10-04","resubscriptionDays":30,"description":"Monthly plan"}`,
			response: `{"message":"SUCCESSFULLY","transactionId":"sub-123","qrCode":"https://qr.example/image","link":"onepay://qr/subscription"}`,
			want:     SubscriptionQRResponse{Message: "SUCCESSFULLY", TransactionID: "sub-123", QRCode: "https://qr.example/image", Link: "onepay://qr/subscription"},
			call: func(ctx context.Context, p *Phajay) (any, error) {
				return p.GenerateSubscriptionQR(ctx, SubscriptionQRRequest{MaxAmount: 1000, SubscriptionDate: "2026-10-04", ResubscriptionDays: 30, Description: "Monthly plan"})
			},
		},
		{
			name: "cancel subscription", method: http.MethodPost,
			path: "/v1/api/subscription/cancel-subscription", body: `{"authCode":"auth-123"}`,
			response: `{"message":"CANCEL_SUBSCRIPTION_SUCCESSFULLY","transactionId":"sub-123","authCode":"auth-123","status":"CANCEL","time":"2025-02-12 15:38:06"}`,
			want:     CancelSubscriptionResponse{Message: "CANCEL_SUBSCRIPTION_SUCCESSFULLY", TransactionID: "sub-123", AuthCode: "auth-123", Status: "CANCEL", Time: "2025-02-12 15:38:06"},
			call: func(ctx context.Context, p *Phajay) (any, error) {
				return p.CancelSubscription(ctx, CancelSubscriptionRequest{AuthCode: "auth-123"})
			},
		},
	}
}

func TestDocumentedEndpoints(t *testing.T) {
	for _, tt := range documentedEndpoints() {
		t.Run(tt.name, func(t *testing.T) {
			srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				if r.Method != tt.method || r.URL.Path != tt.path {
					t.Errorf("request = %s %s, want %s %s", r.Method, r.URL.Path, tt.method, tt.path)
				}
				if r.Header.Get("Content-Type") != "application/json" {
					t.Error("missing JSON content type")
				}
				if tt.basic {
					want := "Basic " + base64.StdEncoding.EncodeToString([]byte("test-key"))
					if r.Header.Get("Authorization") != want || r.Header.Get("secretKey") != "" {
						t.Error("incorrect Basic authentication")
					}
				} else if r.Header.Get("secretKey") != "test-key" || r.Header.Get("Authorization") != "" {
					t.Error("incorrect secretKey authentication")
				}
				data, err := io.ReadAll(r.Body)
				if err != nil {
					t.Error(err)
				}
				if tt.body == "" {
					if len(data) != 0 {
						t.Error("GET request has a body")
					}
				} else {
					var got, want any
					if err := json.Unmarshal(data, &got); err != nil {
						t.Error(err)
					}
					if err := json.Unmarshal([]byte(tt.body), &want); err != nil {
						t.Error(err)
					}
					if !reflect.DeepEqual(got, want) {
						t.Errorf("body = %s, want %s", data, tt.body)
					}
				}
				_, _ = io.WriteString(w, tt.response)
			}))
			defer srv.Close()
			got, err := tt.call(context.Background(), New("test-key", WithBaseURL(srv.URL+"/")))
			if err != nil {
				t.Fatal(err)
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("response = %#v, want %#v", got, tt.want)
			}
		})
	}
}

func TestEndpointFailures(t *testing.T) {
	for _, endpoint := range documentedEndpoints() {
		for _, failure := range []struct {
			name   string
			status int
			body   string
		}{
			{"unauthorized", http.StatusUnauthorized, `{"message":"UNAUTHORIZED"}`},
			{"server error", http.StatusInternalServerError, `{"message":"FAILED"}`},
			{"invalid JSON", http.StatusOK, `invalid`},
		} {
			t.Run(endpoint.name+"/"+failure.name, func(t *testing.T) {
				srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
					w.WriteHeader(failure.status)
					_, _ = io.WriteString(w, failure.body)
				}))
				defer srv.Close()
				_, err := endpoint.call(context.Background(), New("key", WithBaseURL(srv.URL)))
				if err == nil || !strings.HasPrefix(err.Error(), "phajay: ") {
					t.Fatalf("error = %v", err)
				}
				if failure.status != http.StatusOK {
					var apiErr *APIError
					if !errors.As(err, &apiErr) || apiErr.StatusCode != failure.status || apiErr.Body != failure.body {
						t.Errorf("APIError = %#v", apiErr)
					}
				}
			})
		}
	}
}

func TestEndpointsCancelAndRejectUnsupportedSandbox(t *testing.T) {
	for _, tt := range documentedEndpoints() {
		t.Run(tt.name, func(t *testing.T) {
			var calls atomic.Int32
			srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) { calls.Add(1) }))
			defer srv.Close()
			ctx, cancel := context.WithCancel(context.Background())
			cancel()
			if _, err := tt.call(ctx, New("key", WithBaseURL(srv.URL))); !errors.Is(err, context.Canceled) {
				t.Errorf("canceled error = %v", err)
			}
			if _, err := tt.call(context.Background(), New("key", WithBaseURL(srv.URL), WithSandbox())); err == nil || !strings.Contains(err.Error(), "sandbox") {
				t.Errorf("sandbox error = %v", err)
			}
			if calls.Load() != 0 {
				t.Error("unexpected HTTP request")
			}
		})
	}
}

func TestValidationBeforeHTTP(t *testing.T) {
	var calls atomic.Int32
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) { calls.Add(1) }))
	defer srv.Close()
	p := New("key", WithBaseURL(srv.URL))
	ctx := context.Background()
	for _, amount := range []float64{0, -1, math.NaN(), math.Inf(1), math.Inf(-1)} {
		for _, method := range []string{"link", "QR", "card", "subscription"} {
			t.Run(method+"/amount/"+strconv.FormatFloat(amount, 'g', -1, 64), func(t *testing.T) {
				var err error
				switch method {
				case "link":
					_, err = p.CreatePaymentLink(ctx, PaymentLinkRequest{Amount: amount, Description: "x"})
				case "QR":
					_, err = p.GenerateQR(ctx, BankJDB, PaymentQRRequest{Amount: amount, Description: "x"})
				case "card":
					_, err = p.CreateCreditCardPaymentLink(ctx, CreditCardPaymentRequest{Amount: amount, Description: "x"})
				case "subscription":
					_, err = p.GenerateSubscriptionQR(ctx, SubscriptionQRRequest{MaxAmount: amount, SubscriptionDate: "2026-10-04", ResubscriptionDays: 30, Description: "x"})
				}
				if err == nil {
					t.Error("expected validation error")
				}
			})
		}
	}
	for _, description := range []string{"", " \n\t"} {
		t.Run("description/"+description, func(t *testing.T) {
			if _, err := p.CreatePaymentLink(ctx, PaymentLinkRequest{Amount: 1, Description: description}); err == nil {
				t.Error("link accepted empty description")
			}
			if _, err := p.GenerateQR(ctx, BankJDB, PaymentQRRequest{Amount: 1, Description: description}); err == nil {
				t.Error("QR accepted empty description")
			}
			if _, err := p.CreateCreditCardPaymentLink(ctx, CreditCardPaymentRequest{Amount: 1, Description: description}); err == nil {
				t.Error("card accepted empty description")
			}
			if _, err := p.GenerateSubscriptionQR(ctx, SubscriptionQRRequest{MaxAmount: 1, SubscriptionDate: "2026-10-04", ResubscriptionDays: 30, Description: description}); err == nil {
				t.Error("subscription accepted empty description")
			}
		})
	}
	for _, id := range []string{"", " \t", ".", ".."} {
		t.Run("ID/"+id, func(t *testing.T) {
			if _, err := p.CheckTransactionStatus(ctx, id); err == nil {
				t.Error("status accepted invalid ID")
			}
			if _, err := p.CheckRefundStatus(ctx, id); err == nil {
				t.Error("refund status accepted invalid ID")
			}
		})
	}
	for _, id := range []string{"", " \t"} {
		if _, err := p.RequestRefund(ctx, RefundRequest{TransactionID: id}); err == nil {
			t.Error("refund accepted empty transaction ID")
		}
		if _, err := p.CancelSubscription(ctx, CancelSubscriptionRequest{AuthCode: id}); err == nil {
			t.Error("cancel accepted empty auth code")
		}
	}
	for _, date := range []string{"", "2026-02-30", "04/10/2026", "2026-1-04"} {
		if _, err := p.GenerateSubscriptionQR(ctx, SubscriptionQRRequest{MaxAmount: 1, SubscriptionDate: date, ResubscriptionDays: 30, Description: "x"}); err == nil {
			t.Errorf("subscription accepted date %q", date)
		}
	}
	for _, days := range []int{0, -1} {
		if _, err := p.GenerateSubscriptionQR(ctx, SubscriptionQRRequest{MaxAmount: 1, SubscriptionDate: "2026-10-04", ResubscriptionDays: days, Description: "x"}); err == nil {
			t.Errorf("subscription accepted days %d", days)
		}
	}
	if calls.Load() != 0 {
		t.Fatalf("invalid requests sent %d HTTP requests", calls.Load())
	}
}

func TestTransactionStatusResponseVariants(t *testing.T) {
	for _, tt := range []struct {
		name, payload string
		want          TransactionStatusResponse
	}{
		{"not found", `{"message":"PAYMENT_NOT_FOUND"}`, TransactionStatusResponse{Message: "PAYMENT_NOT_FOUND"}},
		{"paid", `{"message":"SUCCESSFULLY","data":{"transactionId":"paid","linkCode":"link","orderNo":"order","status":"PAYMENT_COMPLETED","isPaid":true,"amount":10,"paymentTime":"2026-10-04T00:00:00Z"}}`, TransactionStatusResponse{Message: "SUCCESSFULLY", Data: &TransactionStatus{TransactionID: "paid", LinkCode: testPtr("link"), OrderNo: testPtr("order"), Status: TransactionStatusCompleted, IsPaid: true, Amount: 10, PaymentTime: testPtr("2026-10-04T00:00:00Z")}}},
	} {
		t.Run(tt.name, func(t *testing.T) {
			srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) { _, _ = io.WriteString(w, tt.payload) }))
			defer srv.Close()
			got, err := New("key", WithBaseURL(srv.URL)).CheckTransactionStatus(context.Background(), "id")
			if err != nil {
				t.Fatal(err)
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("response = %#v, want %#v", got, tt.want)
			}
		})
	}
}

func TestEscapedIdentifiers(t *testing.T) {
	for _, tt := range []struct {
		name, path string
		call       func(context.Context, *Phajay) (any, error)
	}{
		{"transaction", "/v1/api/payment/check-transaction/status/txn%2Fpart%3Fquery%23fragment", func(ctx context.Context, p *Phajay) (any, error) {
			return p.CheckTransactionStatus(ctx, "txn/part?query#fragment")
		}},
		{"refund", "/v1/api/refund/bill%2Fpart%3Fquery%23fragment", func(ctx context.Context, p *Phajay) (any, error) {
			return p.CheckRefundStatus(ctx, "bill/part?query#fragment")
		}},
	} {
		t.Run(tt.name, func(t *testing.T) {
			srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				if r.URL.EscapedPath() != tt.path || r.URL.RawQuery != "" || r.URL.Fragment != "" {
					t.Errorf("unexpected URL %s", r.URL)
				}
				_, _ = io.WriteString(w, `{"message":"SUCCESSFULLY"}`)
			}))
			defer srv.Close()
			if _, err := tt.call(context.Background(), New("key", WithBaseURL(srv.URL))); err != nil {
				t.Fatal(err)
			}
		})
	}
}
