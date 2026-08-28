package phajay

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
)

// Bank identifies the target bank for QR generation.
type Bank int

const (
	BankJDB     Bank = iota + 1 // Joint Development Bank
	BankLDB                     // Lao Development Bank
	BankIB                      // Indochina Bank
	BankBCEL                    // Banque Pour Le Commerce Exterieur Lao Public
	BankSTB                     // ST Bank Laos
	BankMMoneyX                 // M MoneyX
)

const (
	generateJDBQRPath     = "/v1/api/payment/generate-jdb-qr"
	generateLDBQRPath     = "/v1/api/payment/generate-ldb-qr"
	generateIBQRPath      = "/v1/api/payment/generate-ib-qr"
	generateBCELQRPath    = "/v1/api/payment/generate-bcel-qr"
	generateSTBQRPath     = "/v1/api/payment/generate-stb-qr"
	generateMMoneyXQRPath = "/v1/api/payment/generate-m-money-qr"
)

// ---- PaymentQR Types ----

type PaymentQRRequest struct {
	Amount      float64 `json:"amount"`      // required
	Description string  `json:"description"` // required
	Tag1        string  `json:"tag1"`
	Tag2        string  `json:"tag2"`
	Tag3        string  `json:"tag3"`
}

type PaymentQRResponse struct {
	Message       string `json:"message"`
	TransactionID string `json:"transactionId"`
	QRCode        string `json:"qrCode"`
	Link          string `json:"link"`
}

// ---- PaymentQR Usecase ----

func (p *Phajay) GenerateQR(ctx context.Context, bank Bank, request PaymentQRRequest) (PaymentQRResponse, error) {
	var response PaymentQRResponse

	path, err := bankPath(bank)
	if err != nil {
		return response, err
	}

	if request.Amount <= 0 {
		return response, fmt.Errorf("phajay: amount must be greater than zero")
	}
	if request.Description == "" {
		return response, fmt.Errorf("phajay: description is required")
	}

	body, err := json.Marshal(request)
	if err != nil {
		return response, fmt.Errorf("phajay: marshal payment qr request: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, p.baseURL+path, strings.NewReader(string(body)))
	if err != nil {
		return response, fmt.Errorf("phajay: build payment qr request: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("secretKey", p.key)

	resp, err := p.client.Do(req)
	if err != nil {
		return response, fmt.Errorf("phajay: send payment qr request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		respBody, _ := io.ReadAll(io.LimitReader(resp.Body, 1<<20))
		return response, fmt.Errorf("phajay: payment qr request failed with status %d: %s", resp.StatusCode, strings.TrimSpace(string(respBody)))
	}

	if err := json.NewDecoder(resp.Body).Decode(&response); err != nil {
		return response, fmt.Errorf("phajay: decode payment qr response: %w", err)
	}

	return response, nil
}

func bankPath(bank Bank) (string, error) {
	switch bank {
	case BankJDB:
		return generateJDBQRPath, nil
	case BankLDB:
		return generateLDBQRPath, nil
	case BankIB:
		return generateIBQRPath, nil
	case BankBCEL:
		return generateBCELQRPath, nil
	case BankSTB:
		return generateSTBQRPath, nil
	case BankMMoneyX:
		return generateMMoneyXQRPath, nil
	default:
		return "", fmt.Errorf("phajay: unknown bank %d", bank)
	}
}
