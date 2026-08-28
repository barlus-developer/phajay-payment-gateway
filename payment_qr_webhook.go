package phajay

import (
	"bytes"
	"encoding/json"
)

// FlexibleString is a string that also unmarshals from a JSON number,
// normalised to its decimal form. Phajay returns refNo as a string for
// some banks and a number for others.
type FlexibleString string

func (s *FlexibleString) UnmarshalJSON(data []byte) error {
	data = bytes.TrimSpace(data)
	if len(data) == 0 || string(data) == "null" {
		*s = ""
		return nil
	}
	if data[0] == '"' {
		var str string
		if err := json.Unmarshal(data, &str); err != nil {
			return err
		}
		*s = FlexibleString(str)
		return nil
	}
	var num json.Number
	if err := json.Unmarshal(data, &num); err != nil {
		return err
	}
	*s = FlexibleString(num.String())
	return nil
}

// PaymentQRWebhookCallback is the payload Phajay POSTs to your callback URL
// when a QR payment reaches a terminal state.
//
// paymentMethod, transactionId, billNumber, txnAmount and status are
// always present. All other fields may be nil — banks return different
// subsets of the payload.
type PaymentQRWebhookCallback struct {
	// Guaranteed fields
	PaymentMethod string  `json:"paymentMethod"`
	TransactionID string  `json:"transactionId"`
	BillNumber    string  `json:"billNumber"`
	TxnAmount     float64 `json:"txnAmount"`
	Status        string  `json:"status"`

	// Optional fields — nil when the bank did not send them
	Message        *string         `json:"message"`
	RefNo          *FlexibleString `json:"refNo"`
	ExReferenceNo  *string         `json:"exReferenceNo"`
	MerchantName   *string         `json:"merchantName"`
	Description    *string         `json:"description"`
	TxnDateTime    *string         `json:"txnDateTime"`
	SourceAccount  *string         `json:"sourceAccount"`
	SourceName     *string         `json:"sourceName"`
	SourceCurrency *string         `json:"sourceCurrency"`
	UserID         *string         `json:"userId"`
	Tag1           *string         `json:"tag1"`
	Tag2           *string         `json:"tag2"`
	Tag3           *string         `json:"tag3"`
	Tag4           *string         `json:"tag4"`
	Tag5           *string         `json:"tag5"`
	Tag6           *string         `json:"tag6"`
}

const PaymentQRWebhookStatusCompleted = "PAYMENT_COMPLETED"
