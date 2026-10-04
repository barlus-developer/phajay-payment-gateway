package phajay

import (
	"context"
	"net/http"
)

const checkTransactionStatusPath = "/v1/api/payment/check-transaction/status/"

const (
	// TransactionStatusWaiting indicates a transaction awaiting payment.
	TransactionStatusWaiting = "WAITING"
	// TransactionStatusExpired indicates an expired transaction.
	TransactionStatusExpired = "EXPIRED"
	// TransactionStatusRefunded indicates a refunded transaction.
	TransactionStatusRefunded = "REFUNDED"
	// TransactionStatusDisputed indicates a disputed transaction.
	TransactionStatusDisputed = "DISPUTED"
	// TransactionStatusCompleted indicates a completed payment.
	TransactionStatusCompleted = "PAYMENT_COMPLETED"
)

// TransactionStatus contains the gateway's current view of a transaction.
// Nullable fields remain nil. Timestamps retain their original wire format.
type TransactionStatus struct {
	TransactionID string  `json:"transactionId"`
	LinkCode      *string `json:"linkCode"`
	OrderNo       *string `json:"orderNo"`
	Status        string  `json:"status"`
	IsPaid        bool    `json:"isPaid"`
	Amount        float64 `json:"amount"`
	Currency      string  `json:"currency"`
	Description   string  `json:"description"`
	PaymentMethod string  `json:"paymentMethod"`
	PaymentTime   *string `json:"paymentTime"`
	CreatedAt     string  `json:"createdAt"`
}

// TransactionStatusResponse contains transaction details. Data is nil if the
// gateway returns a message-only response such as PAYMENT_NOT_FOUND.
type TransactionStatusResponse struct {
	Message string             `json:"message"`
	Data    *TransactionStatus `json:"data"`
}

// CheckTransactionStatus retrieves a transaction using Basic authentication.
// Confirm Data is non-nil and both IsPaid and Status indicate completed payment
// before fulfilling an order. The endpoint has no documented sandbox route.
func (p *Phajay) CheckTransactionStatus(ctx context.Context, transactionID string) (TransactionStatusResponse, error) {
	var response TransactionStatusResponse
	if err := p.requireProduction("transaction status"); err != nil {
		return response, err
	}
	id, err := pathID(transactionID, "transaction ID")
	if err != nil {
		return response, err
	}
	err = p.do(ctx, http.MethodGet, checkTransactionStatusPath+id, nil, authBasic, "transaction status", &response)
	return response, err
}
