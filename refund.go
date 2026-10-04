package phajay

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
)

const refundPath = "/v1/api/refund"

// RefundRequest identifies the payment to refund.
type RefundRequest struct {
	TransactionID string `json:"transactionId"`
}

// Refund contains a refund bill. ID is the identifier used by CheckRefundStatus.
type Refund struct {
	ID        string  `json:"id"`
	Status    string  `json:"status"`
	Amount    float64 `json:"amount"`
	CreatedAt string  `json:"createdAt"`
	BillNo    string  `json:"billNo"`
}

// UnmarshalJSON accepts refund amounts as JSON numbers or numeric strings,
// covering both forms described in the official refund documentation.
func (r *Refund) UnmarshalJSON(data []byte) error {
	type plain Refund
	decoded := struct {
		*plain
		Amount json.Number `json:"amount"`
	}{plain: (*plain)(r)}
	if err := json.Unmarshal(data, &decoded); err != nil {
		return fmt.Errorf("phajay: decode refund: %w", err)
	}
	if decoded.Amount != "" {
		amount, err := decoded.Amount.Float64()
		if err != nil {
			return fmt.Errorf("phajay: decode refund amount: %w", err)
		}
		r.Amount = amount
	}
	return nil
}

// RefundResponse contains the gateway message and refund bill.
type RefundResponse struct {
	Message string `json:"message"`
	Data    Refund `json:"data"`
}

// RequestRefund submits a refund request for a payment. A successful request
// starts the refund workflow; it does not mean the refund has been completed.
// A KYC-approved account is required. No sandbox endpoint is documented.
func (p *Phajay) RequestRefund(ctx context.Context, request RefundRequest) (RefundResponse, error) {
	var response RefundResponse
	if err := p.requireProduction("refund"); err != nil {
		return response, err
	}
	if strings.TrimSpace(request.TransactionID) == "" {
		return response, fmt.Errorf("phajay: transaction ID is required")
	}
	err := p.do(ctx, http.MethodPost, refundPath, request, authSecretKey, "refund", &response)
	return response, err
}

// CheckRefundStatus retrieves a refund bill by the Data.ID from RequestRefund,
// using the secretKey header. No sandbox endpoint is documented.
func (p *Phajay) CheckRefundStatus(ctx context.Context, billID string) (RefundResponse, error) {
	var response RefundResponse
	if err := p.requireProduction("refund status"); err != nil {
		return response, err
	}
	id, err := pathID(billID, "refund bill ID")
	if err != nil {
		return response, err
	}
	err = p.do(ctx, http.MethodGet, refundPath+"/"+id, nil, authSecretKey, "refund status", &response)
	return response, err
}
