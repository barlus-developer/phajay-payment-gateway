package phajay

import (
	"encoding/json"
	"testing"
)

func TestWebhookCallbackUnmarshal(t *testing.T) {
	payload := `{
  "message": "SUCCESS",
  "refNo": "001LNMI395488650608",
  "exReferenceNo": "BONEPBNOAIWN6AFO",
  "merchantName": "PhaJay Payment",
  "memo": "LMPS QR PAYMENT|LNMI|BCEL|1600120000000006170001|LAK|PhaJay Payment|ZAzprpRJomfSW0dofOUC0GwcI|PAYMENT|15:59:50",
  "txnDateTime": "2025-09-12 15:59:50",
  "txnAmount": 1,
  "billNumber": "ZAzprpRJomfSW0dofOUC0GwcI",
  "sourceAccount": "138880037",
  "sourceName": "JDB Yes for Domistrict Bank LAP NET ",
  "sourceCurrency": "LAK",
  "paymentId": "68c3dc6c464ee95aafeb9319",
  "linkCode": "1d180ad6-efca-49c5-be71-80b2e2095414",
  "transactionId": "1d180ad6-efca-49c5-be71-80b2e2095414",
  "paymentMethod": "JDB",
  "status": "PAYMENT_COMPLETED",
  "description": "Buy a product",
  "remark": "",
  "tag1": "6868a691f914536d6d731e63",
  "tag2": "BB SHOP",
  "tag3": "",
  "tag4": "",
  "tag5": "",
  "tag6": "",
  "userId": "66923e41ea9468588820b046",
  "orderNo": "ORDER1757666411760",
  "successURL": "http://localhost:5174/payment/success?linkCode=1d180ad6-efca-49c5-be71-80b2e2095414&amount=1&description=Buy a product&orderNo=ORDER1757666411760"
}`

	var cb WebhookCallback
	if err := json.Unmarshal([]byte(payload), &cb); err != nil {
		t.Fatalf("json.Unmarshal() error = %v", err)
	}

	if cb.PaymentMethod != "JDB" {
		t.Errorf("PaymentMethod = %q, want %q", cb.PaymentMethod, "JDB")
	}
	if cb.LinkCode != "1d180ad6-efca-49c5-be71-80b2e2095414" {
		t.Errorf("LinkCode = %q, want %q", cb.LinkCode, "1d180ad6-efca-49c5-be71-80b2e2095414")
	}
	if cb.TransactionID != "1d180ad6-efca-49c5-be71-80b2e2095414" {
		t.Errorf("TransactionID = %q, want %q", cb.TransactionID, "1d180ad6-efca-49c5-be71-80b2e2095414")
	}
	if cb.OrderNo != "ORDER1757666411760" {
		t.Errorf("OrderNo = %q, want %q", cb.OrderNo, "ORDER1757666411760")
	}
	if cb.TxnAmount != 1 {
		t.Errorf("TxnAmount = %v, want %v", cb.TxnAmount, 1)
	}

	tests := []struct {
		name string
		got  *string
		want string
	}{
		{name: "Message", got: cb.Message, want: "SUCCESS"},
		{name: "RefNo", got: cb.RefNo, want: "001LNMI395488650608"},
		{name: "ExReferenceNo", got: cb.ExReferenceNo, want: "BONEPBNOAIWN6AFO"},
		{name: "MerchantName", got: cb.MerchantName, want: "PhaJay Payment"},
		{name: "TxnDateTime", got: cb.TxnDateTime, want: "2025-09-12 15:59:50"},
		{name: "BillNumber", got: cb.BillNumber, want: "ZAzprpRJomfSW0dofOUC0GwcI"},
		{name: "SourceAccount", got: cb.SourceAccount, want: "138880037"},
		{name: "SourceCurrency", got: cb.SourceCurrency, want: "LAK"},
		{name: "PaymentID", got: cb.PaymentID, want: "68c3dc6c464ee95aafeb9319"},
		{name: "Status", got: cb.Status, want: WebhookStatusCompleted},
		{name: "Description", got: cb.Description, want: "Buy a product"},
		{name: "Remark", got: cb.Remark, want: ""},
		{name: "Tag1", got: cb.Tag1, want: "6868a691f914536d6d731e63"},
		{name: "Tag2", got: cb.Tag2, want: "BB SHOP"},
		{name: "Tag3", got: cb.Tag3, want: ""},
		{name: "UserID", got: cb.UserID, want: "66923e41ea9468588820b046"},
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

func TestWebhookCallbackMinimalPayload(t *testing.T) {
	payload := `{
  "paymentMethod": "BCEL",
  "linkCode": "1d180ad6-efca-49c5-be71-80b2e2095414",
  "transactionId": "1d180ad6-efca-49c5-be71-80b2e2095414",
  "orderNo": "ORDER1757666411760",
  "txnAmount": 1500.50
}`

	var cb WebhookCallback
	if err := json.Unmarshal([]byte(payload), &cb); err != nil {
		t.Fatalf("json.Unmarshal() error = %v", err)
	}

	if cb.PaymentMethod != "BCEL" {
		t.Errorf("PaymentMethod = %q, want %q", cb.PaymentMethod, "BCEL")
	}
	if cb.LinkCode != "1d180ad6-efca-49c5-be71-80b2e2095414" {
		t.Errorf("LinkCode = %q, want %q", cb.LinkCode, "1d180ad6-efca-49c5-be71-80b2e2095414")
	}
	if cb.TransactionID != "1d180ad6-efca-49c5-be71-80b2e2095414" {
		t.Errorf("TransactionID = %q, want %q", cb.TransactionID, "1d180ad6-efca-49c5-be71-80b2e2095414")
	}
	if cb.OrderNo != "ORDER1757666411760" {
		t.Errorf("OrderNo = %q, want %q", cb.OrderNo, "ORDER1757666411760")
	}
	if cb.TxnAmount != 1500.50 {
		t.Errorf("TxnAmount = %v, want %v", cb.TxnAmount, 1500.50)
	}

	if cb.Message != nil {
		t.Errorf("Message = %q, want nil", *cb.Message)
	}
	if cb.Status != nil {
		t.Errorf("Status = %q, want nil", *cb.Status)
	}
	if cb.RefNo != nil {
		t.Errorf("RefNo = %q, want nil", *cb.RefNo)
	}
	if cb.MerchantName != nil {
		t.Errorf("MerchantName = %q, want nil", *cb.MerchantName)
	}
	if cb.Tag1 != nil {
		t.Errorf("Tag1 = %q, want nil", *cb.Tag1)
	}
	if cb.SuccessURL != nil {
		t.Errorf("SuccessURL = %q, want nil", *cb.SuccessURL)
	}
}
