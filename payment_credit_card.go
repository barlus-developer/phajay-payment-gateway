package phajay

import (
	"context"
	"net/http"
)

const createCreditCardPaymentLinkPath = "/v1/api/jdb2c2p/payment/payment-link"

// CreditCardPaymentRequest describes a payment through the hosted card page.
type CreditCardPaymentRequest struct {
	Amount      float64 `json:"amount"`
	Description string  `json:"description"`
	Tag1        string  `json:"tag1,omitempty"`
	Tag2        string  `json:"tag2,omitempty"`
	Tag3        string  `json:"tag3,omitempty"`
}

// CreditCardPaymentResponse contains the hosted payment page and transaction.
// ExpirationTime retains the gateway's date-time string, including its timezone
// format (the documentation's example omits the timezone suffix).
type CreditCardPaymentResponse struct {
	Message        string `json:"message"`
	PaymentURL     string `json:"paymentUrl"`
	ExpirationTime string `json:"expirationTime"`
	TransactionID  string `json:"transactionId"`
	Status         string `json:"status"`
}

// CreateCreditCardPaymentLink creates a hosted card payment page using Basic
// authentication. Phajay requires an approved KYC account and 3DS-enabled cards.
// No sandbox endpoint is documented.
func (p *Phajay) CreateCreditCardPaymentLink(ctx context.Context, request CreditCardPaymentRequest) (CreditCardPaymentResponse, error) {
	var response CreditCardPaymentResponse
	if err := p.requireProduction("credit card payment"); err != nil {
		return response, err
	}
	if err := validatePayment(request.Amount, request.Description); err != nil {
		return response, err
	}
	err := p.do(ctx, http.MethodPost, createCreditCardPaymentLinkPath, request, authBasic, "credit card payment", &response)
	return response, err
}
