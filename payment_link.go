package phajay

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
)

const createPaymentLinkPath = "/v1/api/link/payment-link"

// ---- PaymentLink Types ----

type PaymentLinkRequest struct {
	OrderNo     string  `json:"orderNo"`
	Amount      float64 `json:"amount"`      // required
	Description string  `json:"description"` // required
	Tag1        string  `json:"tag1"`
	Tag2        string  `json:"tag2"`
	Tag3        string  `json:"tag3"`
}

type PaymentLinkResponse struct {
	Message     string `json:"message"`
	RedirectURL string `json:"redirectURL"`
	OrderNo     string `json:"orderNo"`
}

// ---- PaymentLink Usecase ----

func (p *Phajay) CreatePaymentLink(ctx context.Context, request PaymentLinkRequest) (PaymentLinkResponse, error) {
	var response PaymentLinkResponse

	if request.Amount <= 0 {
		return response, fmt.Errorf("phajay: amount must be greater than zero")
	}
	if request.Description == "" {
		return response, fmt.Errorf("phajay: description is required")
	}

	body, err := json.Marshal(request)
	if err != nil {
		return response, fmt.Errorf("phajay: marshal payment link request: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, p.baseURL+createPaymentLinkPath, strings.NewReader(string(body)))
	if err != nil {
		return response, fmt.Errorf("phajay: build payment link request: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Basic "+base64.StdEncoding.EncodeToString([]byte(p.key)))

	resp, err := p.client.Do(req)
	if err != nil {
		return response, fmt.Errorf("phajay: send payment link request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		respBody, _ := io.ReadAll(io.LimitReader(resp.Body, 1<<20))
		return response, fmt.Errorf("phajay: payment link request failed with status %d: %s", resp.StatusCode, strings.TrimSpace(string(respBody)))
	}

	if err := json.NewDecoder(resp.Body).Decode(&response); err != nil {
		return response, fmt.Errorf("phajay: decode payment link response: %w", err)
	}

	return response, nil
}
