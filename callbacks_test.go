package phajay

import (
	"encoding/json"
	"reflect"
	"testing"
)

func TestPaymentLinkWebhookReferences(t *testing.T) {
	for _, tt := range []struct {
		name, payload string
		want          *string
		invalid       bool
	}{
		{"string", `{"refNo":"000123"}`, testPtr("000123"), false},
		{"number", `{"refNo":580960503}`, testPtr("580960503"), false},
		{"large number", `{"refNo":9007199254740993}`, testPtr("9007199254740993"), false},
		{"missing", `{}`, nil, false},
		{"null", `{"refNo":null}`, nil, false},
		{"boolean", `{"refNo":true}`, nil, true},
		{"object", `{"refNo":{}}`, nil, true},
		{"invalid payload", `[]`, nil, true},
	} {
		t.Run(tt.name, func(t *testing.T) {
			var cb PaymentLinkWebhookCallback
			err := json.Unmarshal([]byte(tt.payload), &cb)
			if (err != nil) != tt.invalid {
				t.Fatalf("error = %v", err)
			}
			if !tt.invalid && !reflect.DeepEqual(cb.RefNo, tt.want) {
				t.Errorf("RefNo = %v, want %v", cb.RefNo, tt.want)
			}
		})
	}
}

func TestQRWebhookOptionalFields(t *testing.T) {
	for _, payload := range []string{
		`{"paymentMethod":"BCEL","transactionId":"txn","txnAmount":10}`,
		`{"paymentMethod":"BCEL","transactionId":"txn","txnAmount":10,"billNumber":null,"status":null}`,
	} {
		var cb PaymentQRWebhookCallback
		if err := json.Unmarshal([]byte(payload), &cb); err != nil {
			t.Fatal(err)
		}
		if cb.BillNumber != nil || cb.Status != nil {
			t.Fatalf("optional fields = %#v", cb)
		}
		if cb.PaymentMethod != "BCEL" || cb.TransactionID != "txn" || cb.TxnAmount != 10 {
			t.Fatalf("guaranteed fields = %#v", cb)
		}
	}
}

func TestCreditCardWebhook(t *testing.T) {
	payload := `{"message":"SUCCESS","transactionId":"card-txn","linkCode":"card-txn","paymentMethod":"CREDIT_CARD","status":"PAYMENT_COMPLETED","tag1":null,"txnAmount":1,"currency":"USD","cardType":"CC-VI","creditCardDetails":{"issuerCode":null,"cardHolderName":"Test Customer","cardNumber":"412639XXXXXX2372","cardExpiry":"XXXX","authenticationStatus":null,"eciValue":null,"xidValue":null,"cavvValue":null,"enrolledFlag":null},"transactionAmount":{"amountText":"000000000100","currencyCode":"USD","decimalPlaces":2,"amount":1}}`
	var cb CreditCardWebhookCallback
	if err := json.Unmarshal([]byte(payload), &cb); err != nil {
		t.Fatal(err)
	}
	if cb.TransactionID == nil || *cb.TransactionID != "card-txn" || cb.Status == nil || *cb.Status != TransactionStatusCompleted {
		t.Fatalf("callback = %#v", cb)
	}
	if cb.Tag1 != nil || cb.Tag2 != nil || cb.SuccessURL != nil {
		t.Error("missing or null fields should be nil")
	}
	if cb.CreditCardDetails == nil || cb.CreditCardDetails.CardNumber == nil || *cb.CreditCardDetails.CardNumber != "412639XXXXXX2372" || cb.CreditCardDetails.IssuerCode != nil {
		t.Fatalf("card details = %#v", cb.CreditCardDetails)
	}
	if cb.TransactionAmount == nil || cb.TransactionAmount.Amount == nil || *cb.TransactionAmount.Amount != 1 || cb.TransactionAmount.DecimalPlaces == nil || *cb.TransactionAmount.DecimalPlaces != 2 {
		t.Fatalf("amount = %#v", cb.TransactionAmount)
	}
	for _, payload := range []string{`{}`, `{"creditCardDetails":null,"transactionAmount":null}`} {
		var empty CreditCardWebhookCallback
		if err := json.Unmarshal([]byte(payload), &empty); err != nil {
			t.Fatal(err)
		}
		if !reflect.DeepEqual(empty, CreditCardWebhookCallback{}) {
			t.Errorf("empty callback = %#v", empty)
		}
	}
}

func TestSubscriptionWebhookEvents(t *testing.T) {
	for _, tt := range []struct {
		name, payload, status string
		payment               *string
	}{
		{"setup", `{"message":"SUBSCRIPTION_CONNECTED_SUCCESSFULLY","status":"SUBSCRIPTION_CONNECTED","transactionId":"sub","authCode":"auth","time":"2025-02-12 14:54:45"}`, SubscriptionStatusConnected, nil},
		{"debit", `{"message":"SUBSCRIPTION_SUCCESSFULLY","status":"SUBSCRIPTION_SUCCESS","transactionId":"sub","authCode":"auth","time":"2025-02-12 14:54:45","paymentTransactionId":"payment"}`, SubscriptionStatusSuccess, testPtr("payment")},
		{"cancel", `{"message":"CANCEL_SUBSCRIPTION_SUCCESSFULLY","status":"CANCEL","transactionId":"sub","authCode":"auth","time":"2025-02-12 14:54:45"}`, SubscriptionStatusCanceled, nil},
	} {
		t.Run(tt.name, func(t *testing.T) {
			var cb SubscriptionWebhookCallback
			if err := json.Unmarshal([]byte(tt.payload), &cb); err != nil {
				t.Fatal(err)
			}
			if cb.Status != tt.status || cb.TransactionID != "sub" || cb.AuthCode != "auth" || cb.Time != "2025-02-12 14:54:45" || !reflect.DeepEqual(cb.PaymentTransactionID, tt.payment) {
				t.Fatalf("callback = %#v", cb)
			}
		})
	}
}

func TestSandboxPaymentLinkWebhook(t *testing.T) {
	var cb SandboxPaymentLinkWebhookCallback
	if err := json.Unmarshal([]byte(`{"_id":"test-id","orderNo":"ORD-1","linkCode":"link","amount":10,"descriptione":"Order","status":"PAYMENT_COMPLETED","user":"user","isMerchantIdPayment":false,"updateAt":"2026-10-04T00:00:00Z"}`), &cb); err != nil {
		t.Fatal(err)
	}
	if cb.ID == nil || *cb.ID != "test-id" || cb.Amount == nil || *cb.Amount != 10 || cb.Description == nil || *cb.Description != "Order" {
		t.Fatalf("callback = %#v", cb)
	}
	if cb.IsMerchantIDPayment == nil || *cb.IsMerchantIDPayment || cb.IsAffiliatePayment != nil || cb.Tag1 != nil {
		t.Error("optional fields lost false or missing distinction")
	}
}

func TestRefundAmountVariants(t *testing.T) {
	for _, tt := range []struct {
		name, payload string
		want          float64
		invalid       bool
	}{
		{"number", `{"amount":1000.5}`, 1000.5, false},
		{"string", `{"amount":"1000.5"}`, 1000.5, false},
		{"invalid number", `{"amount":"abc"}`, 0, true},
		{"overflow", `{"amount":1e400}`, 0, true},
		{"boolean", `{"amount":true}`, 0, true},
	} {
		t.Run(tt.name, func(t *testing.T) {
			var refund Refund
			err := json.Unmarshal([]byte(tt.payload), &refund)
			if (err != nil) != tt.invalid {
				t.Fatalf("error = %v", err)
			}
			if !tt.invalid && refund.Amount != tt.want {
				t.Errorf("Amount = %v, want %v", refund.Amount, tt.want)
			}
		})
	}
}
