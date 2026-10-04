package phajay

import (
	"context"
	"net/http"
)

const createPaymentLinkPath = "/v1/api/link/payment-link"
const createSandboxPaymentLinkPath = "/v1/api/test/payment/get-payment-link"

// PaymentLinkRequest describes a payment. OrderNo should be unique when set.
type PaymentLinkRequest struct {
	OrderNo     string  `json:"orderNo"`
	Amount      float64 `json:"amount"`      // required
	Description string  `json:"description"` // required
	Tag1        string  `json:"tag1"`
	Tag2        string  `json:"tag2"`
	Tag3        string  `json:"tag3"`
}

// PaymentLinkResponse contains the URL where the customer completes payment.
type PaymentLinkResponse struct {
	Message     string `json:"message"`
	RedirectURL string `json:"redirectURL"`
	OrderNo     string `json:"orderNo"`
}

// CreatePaymentLink creates a payment link after validating amount and description.
func (p *Phajay) CreatePaymentLink(ctx context.Context, request PaymentLinkRequest) (PaymentLinkResponse, error) {
	var response PaymentLinkResponse
	if err := validatePayment(request.Amount, request.Description); err != nil {
		return response, err
	}
	path := createPaymentLinkPath
	if p.sandbox {
		path = createSandboxPaymentLinkPath
	}
	err := p.do(ctx, http.MethodPost, path, request, authBasic, "payment link", &response)
	return response, err
}
