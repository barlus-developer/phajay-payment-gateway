package phajay

import (
	"context"
	"fmt"
	"net/http"
	"strings"
)

// Bank identifies the target bank for QR generation.
type Bank int

const (
	// BankJDB identifies Joint Development Bank.
	BankJDB Bank = iota + 1
	// BankLDB identifies Lao Development Bank.
	BankLDB
	// BankIB identifies Indochina Bank.
	BankIB
	// BankBCEL identifies BCEL. Its QR codes require the BCEL One app.
	BankBCEL
	// BankSTB identifies ST Bank Laos.
	BankSTB
	// BankMMoneyX identifies M MoneyX. No sandbox route is documented.
	BankMMoneyX
)

const (
	generateJDBQRPath     = "/v1/api/payment/generate-jdb-qr"
	generateLDBQRPath     = "/v1/api/payment/generate-ldb-qr"
	generateIBQRPath      = "/v1/api/payment/generate-ib-qr"
	generateBCELQRPath    = "/v1/api/payment/generate-bcel-qr"
	generateSTBQRPath     = "/v1/api/payment/generate-stb-qr"
	generateMMoneyXQRPath = "/v1/api/payment/generate-m-money-qr"
)

// PaymentQRRequest describes a bank QR payment. OrderNo should be unique when set.
type PaymentQRRequest struct {
	OrderNo     string  `json:"orderNo,omitempty"`
	Amount      float64 `json:"amount"`      // required
	Description string  `json:"description"` // required
	Tag1        string  `json:"tag1"`
	Tag2        string  `json:"tag2"`
	Tag3        string  `json:"tag3"`
}

// PaymentQRResponse contains a QR string and a bank app deep link, when available.
type PaymentQRResponse struct {
	Message       string `json:"message"`
	TransactionID string `json:"transactionId"`
	QRCode        string `json:"qrCode"`
	Link          string `json:"link"`
}

// GenerateQR creates a bank QR payment after validating amount and description.
// BCEL descriptions do not support Lao or Thai characters according to Phajay.
func (p *Phajay) GenerateQR(ctx context.Context, bank Bank, request PaymentQRRequest) (PaymentQRResponse, error) {
	var response PaymentQRResponse
	path, err := bankPath(bank)
	if err != nil {
		return response, err
	}
	if err := validatePayment(request.Amount, request.Description); err != nil {
		return response, err
	}
	if p.sandbox {
		if bank == BankMMoneyX {
			return response, fmt.Errorf("phajay: M MoneyX has no documented sandbox endpoint")
		}
		path = strings.Replace(path, "/v1/api/payment/", "/v1/api/test/payment/", 1)
	}
	err = p.do(ctx, http.MethodPost, path, request, authSecretKey, "payment qr", &response)
	return response, err
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
