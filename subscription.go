package phajay

import (
	"context"
	"fmt"
	"math"
	"net/http"
	"strings"
	"time"
)

const (
	generateSubscriptionQRPath = "/v1/api/subscription/generate-bcel-qr"
	cancelSubscriptionPath     = "/v1/api/subscription/cancel-subscription"
)

// SubscriptionQRRequest describes a BCEL recurring payment authorization.
type SubscriptionQRRequest struct {
	MaxAmount          float64 `json:"maxAmount"`
	SubscriptionDate   string  `json:"subscriptionDate"`
	ResubscriptionDays int     `json:"resubscriptionDays"`
	Description        string  `json:"description"`
}

// SubscriptionQRResponse contains a subscription QR code and bank app link.
// QRCode may be a QR image URL, as shown in the official subscription example.
type SubscriptionQRResponse struct {
	Message       string `json:"message"`
	TransactionID string `json:"transactionId"`
	QRCode        string `json:"qrCode"`
	Link          string `json:"link"`
}

// CancelSubscriptionRequest identifies the bank authorization to cancel.
type CancelSubscriptionRequest struct {
	AuthCode string `json:"authCode"`
}

// CancelSubscriptionResponse contains the result of canceling an authorization.
type CancelSubscriptionResponse struct {
	Message       string `json:"message"`
	TransactionID string `json:"transactionId"`
	AuthCode      string `json:"authCode"`
	Status        string `json:"status"`
	Time          string `json:"time"`
}

// GenerateSubscriptionQR creates a BCEL subscription authorization QR code.
// SubscriptionDate must use YYYY-MM-DD and ResubscriptionDays must be positive.
// Phajay schedules debits after authorization; there is no documented debit API
// or sandbox route. An approved KYC account is required.
func (p *Phajay) GenerateSubscriptionQR(ctx context.Context, request SubscriptionQRRequest) (SubscriptionQRResponse, error) {
	var response SubscriptionQRResponse
	if err := p.requireProduction("subscription QR"); err != nil {
		return response, err
	}
	if math.IsNaN(request.MaxAmount) || math.IsInf(request.MaxAmount, 0) || request.MaxAmount <= 0 {
		return response, fmt.Errorf("phajay: max amount must be finite and greater than zero")
	}
	if _, err := time.Parse("2006-01-02", request.SubscriptionDate); err != nil {
		return response, fmt.Errorf("phajay: subscription date must use YYYY-MM-DD: %w", err)
	}
	if request.ResubscriptionDays <= 0 {
		return response, fmt.Errorf("phajay: resubscription days must be greater than zero")
	}
	if strings.TrimSpace(request.Description) == "" {
		return response, fmt.Errorf("phajay: description is required")
	}
	err := p.do(ctx, http.MethodPost, generateSubscriptionQRPath, request, authSecretKey, "subscription QR", &response)
	return response, err
}

// CancelSubscription cancels an authorization using its AuthCode and the
// secretKey header. No sandbox endpoint is documented.
func (p *Phajay) CancelSubscription(ctx context.Context, request CancelSubscriptionRequest) (CancelSubscriptionResponse, error) {
	var response CancelSubscriptionResponse
	if err := p.requireProduction("cancel subscription"); err != nil {
		return response, err
	}
	if strings.TrimSpace(request.AuthCode) == "" {
		return response, fmt.Errorf("phajay: auth code is required")
	}
	err := p.do(ctx, http.MethodPost, cancelSubscriptionPath, request, authSecretKey, "cancel subscription", &response)
	return response, err
}
