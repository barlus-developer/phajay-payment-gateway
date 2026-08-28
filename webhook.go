package phajay

// WebhookCallback is the payload Phajay POSTs to your callback URL.
//
// paymentMethod, linkCode, transactionId, orderNo and txnAmount are
// always present. All other fields may be nil — banks return different
// subsets of the payload.
type WebhookCallback struct {
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

const WebhookStatusCompleted = "PAYMENT_COMPLETED"
