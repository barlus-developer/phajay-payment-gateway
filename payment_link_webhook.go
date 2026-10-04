package phajay

import (
	"encoding/json"
	"fmt"
)

// PaymentLinkWebhookCallback is the payload Phajay POSTs to your callback URL.
//
// paymentMethod, linkCode, transactionId, orderNo and txnAmount are
// always present. All other fields may be nil — banks return different
// subsets of the payload.
type PaymentLinkWebhookCallback struct {
	// Guaranteed fields
	PaymentMethod string  `json:"paymentMethod"`
	LinkCode      string  `json:"linkCode"`
	TransactionID string  `json:"transactionId"`
	OrderNo       string  `json:"orderNo"`
	TxnAmount     float64 `json:"txnAmount"`

	// Optional fields — nil when the bank did not send them
	Message        *string `json:"message"`
	RefNo          *string `json:"refNo"`
	ExReferenceNo  *string `json:"exReferenceNo"`
	MerchantName   *string `json:"merchantName"`
	Memo           *string `json:"memo"`
	TxnDateTime    *string `json:"txnDateTime"`
	BillNumber     *string `json:"billNumber"`
	SourceAccount  *string `json:"sourceAccount"`
	SourceName     *string `json:"sourceName"`
	SourceCurrency *string `json:"sourceCurrency"`
	PaymentID      *string `json:"paymentId"`
	Status         *string `json:"status"`
	Description    *string `json:"description"`
	Remark         *string `json:"remark"`
	Tag1           *string `json:"tag1"`
	Tag2           *string `json:"tag2"`
	Tag3           *string `json:"tag3"`
	Tag4           *string `json:"tag4"`
	Tag5           *string `json:"tag5"`
	Tag6           *string `json:"tag6"`
	UserID         *string `json:"userId"`
	SuccessURL     *string `json:"successURL"`
}

// UnmarshalJSON accepts refNo as either a string or a number while retaining the
// existing *string field type. Missing and null references remain nil.
func (cb *PaymentLinkWebhookCallback) UnmarshalJSON(data []byte) error {
	type plain PaymentLinkWebhookCallback
	decoded := struct {
		*plain
		RefNo json.RawMessage `json:"refNo"`
	}{plain: (*plain)(cb)}
	if err := json.Unmarshal(data, &decoded); err != nil {
		return fmt.Errorf("phajay: decode payment link webhook: %w", err)
	}
	if len(decoded.RefNo) != 0 {
		var ref *FlexibleString
		if err := json.Unmarshal(decoded.RefNo, &ref); err != nil {
			return fmt.Errorf("phajay: decode payment link webhook refNo: %w", err)
		}
		cb.RefNo = nil
		if ref != nil {
			value := string(*ref)
			cb.RefNo = &value
		}
	}
	return nil
}

// PaymentLinkWebhookStatusCompleted identifies a completed payment link.
const PaymentLinkWebhookStatusCompleted = "PAYMENT_COMPLETED"
