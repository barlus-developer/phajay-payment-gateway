package phajay

import (
	"encoding/json"
	"testing"
)

func TestPaymentQRWebhookCallbackUnmarshal(t *testing.T) {
	payload := `{
    "billNumber": "81f8b0a6-9277-45e5-a6d2-8c1bd248bedf",
    "description": "short is good",
    "exReferenceNo": "YQ75K0RG6AXG",
    "merchantName": "Lailao Payment",
    "message": "SUCCESS",
    "paymentMethod": "BCEL",
    "refNo": 580960503,
    "sourceAccount": "202509232428318",
    "sourceCurrency": "LAK",
    "sourceName": "ເຢີກື ເຊ່ຍເຊີ",
    "status": "PAYMENT_COMPLETED",
    "tag1": "",
    "tag2": "",
    "tag3": "",
    "tag4": "",
    "tag5": "",
    "tag6": "",
    "transactionId": "81f8b0a6-9277-45e5-a6d2-8c1bd248bedf",
    "txnAmount": 1,
    "txnDateTime": "23/09/2025 11:37:55",
    "userId": "683ae8e8912ff3f68e2faa78"
  }`

	var cb PaymentQRWebhookCallback
	if err := json.Unmarshal([]byte(payload), &cb); err != nil {
		t.Fatalf("json.Unmarshal() error = %v", err)
	}

	if cb.PaymentMethod != "BCEL" {
		t.Errorf("PaymentMethod = %q, want %q", cb.PaymentMethod, "BCEL")
	}
	if cb.TransactionID != "81f8b0a6-9277-45e5-a6d2-8c1bd248bedf" {
		t.Errorf("TransactionID = %q, want %q", cb.TransactionID, "81f8b0a6-9277-45e5-a6d2-8c1bd248bedf")
	}
	if cb.BillNumber != "81f8b0a6-9277-45e5-a6d2-8c1bd248bedf" {
		t.Errorf("BillNumber = %q, want %q", cb.BillNumber, "81f8b0a6-9277-45e5-a6d2-8c1bd248bedf")
	}
	if cb.TxnAmount != 1 {
		t.Errorf("TxnAmount = %v, want %v", cb.TxnAmount, 1)
	}
	if cb.Status != PaymentQRWebhookStatusCompleted {
		t.Errorf("Status = %q, want %q", cb.Status, PaymentQRWebhookStatusCompleted)
	}

	if cb.RefNo == nil {
		t.Fatal("RefNo = nil, want \"580960503\"")
	}
	if got, want := string(*cb.RefNo), "580960503"; got != want {
		t.Errorf("RefNo = %q, want %q", got, want)
	}

	tests := []struct {
		name string
		got  *string
		want string
	}{
		{name: "Message", got: cb.Message, want: "SUCCESS"},
		{name: "ExReferenceNo", got: cb.ExReferenceNo, want: "YQ75K0RG6AXG"},
		{name: "MerchantName", got: cb.MerchantName, want: "Lailao Payment"},
		{name: "Description", got: cb.Description, want: "short is good"},
		{name: "TxnDateTime", got: cb.TxnDateTime, want: "23/09/2025 11:37:55"},
		{name: "SourceAccount", got: cb.SourceAccount, want: "202509232428318"},
		{name: "SourceCurrency", got: cb.SourceCurrency, want: "LAK"},
		{name: "SourceName", got: cb.SourceName, want: "ເຢີກື ເຊ່ຍເຊີ"},
		{name: "UserID", got: cb.UserID, want: "683ae8e8912ff3f68e2faa78"},
		{name: "Tag1", got: cb.Tag1, want: ""},
		{name: "Tag2", got: cb.Tag2, want: ""},
		{name: "Tag3", got: cb.Tag3, want: ""},
		{name: "Tag4", got: cb.Tag4, want: ""},
		{name: "Tag5", got: cb.Tag5, want: ""},
		{name: "Tag6", got: cb.Tag6, want: ""},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.got == nil {
				t.Fatalf("%s = nil, want %q", tt.name, tt.want)
			}
			if *tt.got != tt.want {
				t.Errorf("%s = %q, want %q", tt.name, *tt.got, tt.want)
			}
		})
	}
}

func TestPaymentQRWebhookCallbackMinimalPayload(t *testing.T) {
	payload := `{
  "paymentMethod": "BCEL",
  "transactionId": "81f8b0a6-9277-45e5-a6d2-8c1bd248bedf",
  "billNumber": "81f8b0a6-9277-45e5-a6d2-8c1bd248bedf",
  "txnAmount": 1,
  "status": "PAYMENT_COMPLETED"
}`

	var cb PaymentQRWebhookCallback
	if err := json.Unmarshal([]byte(payload), &cb); err != nil {
		t.Fatalf("json.Unmarshal() error = %v", err)
	}

	if cb.PaymentMethod != "BCEL" {
		t.Errorf("PaymentMethod = %q, want %q", cb.PaymentMethod, "BCEL")
	}
	if cb.TransactionID != "81f8b0a6-9277-45e5-a6d2-8c1bd248bedf" {
		t.Errorf("TransactionID = %q, want %q", cb.TransactionID, "81f8b0a6-9277-45e5-a6d2-8c1bd248bedf")
	}
	if cb.BillNumber != "81f8b0a6-9277-45e5-a6d2-8c1bd248bedf" {
		t.Errorf("BillNumber = %q, want %q", cb.BillNumber, "81f8b0a6-9277-45e5-a6d2-8c1bd248bedf")
	}
	if cb.TxnAmount != 1 {
		t.Errorf("TxnAmount = %v, want %v", cb.TxnAmount, 1)
	}
	if cb.Status != PaymentQRWebhookStatusCompleted {
		t.Errorf("Status = %q, want %q", cb.Status, PaymentQRWebhookStatusCompleted)
	}

	if cb.Message != nil {
		t.Errorf("Message = %q, want nil", *cb.Message)
	}
	if cb.RefNo != nil {
		t.Errorf("RefNo = %q, want nil", *cb.RefNo)
	}
	if cb.ExReferenceNo != nil {
		t.Errorf("ExReferenceNo = %q, want nil", *cb.ExReferenceNo)
	}
	if cb.MerchantName != nil {
		t.Errorf("MerchantName = %q, want nil", *cb.MerchantName)
	}
	if cb.Description != nil {
		t.Errorf("Description = %q, want nil", *cb.Description)
	}
	if cb.TxnDateTime != nil {
		t.Errorf("TxnDateTime = %q, want nil", *cb.TxnDateTime)
	}
	if cb.SourceAccount != nil {
		t.Errorf("SourceAccount = %q, want nil", *cb.SourceAccount)
	}
	if cb.SourceName != nil {
		t.Errorf("SourceName = %q, want nil", *cb.SourceName)
	}
	if cb.SourceCurrency != nil {
		t.Errorf("SourceCurrency = %q, want nil", *cb.SourceCurrency)
	}
	if cb.UserID != nil {
		t.Errorf("UserID = %q, want nil", *cb.UserID)
	}
	if cb.Tag1 != nil {
		t.Errorf("Tag1 = %q, want nil", *cb.Tag1)
	}
	if cb.Tag2 != nil {
		t.Errorf("Tag2 = %q, want nil", *cb.Tag2)
	}
}

func TestFlexibleStringUnmarshal(t *testing.T) {
	tests := []struct {
		name string
		data string
		want string
	}{
		{name: "number", data: `580960503`, want: "580960503"},
		{name: "string", data: `"001LNMI395488650608"`, want: "001LNMI395488650608"},
		{name: "null", data: `null`, want: ""},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var s FlexibleString
			if err := json.Unmarshal([]byte(tt.data), &s); err != nil {
				t.Fatalf("json.Unmarshal() error = %v", err)
			}
			if got, want := string(s), tt.want; got != want {
				t.Errorf("FlexibleString = %q, want %q", got, want)
			}
		})
	}
}
