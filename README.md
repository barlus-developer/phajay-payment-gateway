# Phajay Payment Gateway SDK for Go

[![Go Version](https://img.shields.io/badge/Go-1.23-00ADD8?logo=go)](https://go.dev/dl/)
[![License](https://img.shields.io/badge/License-MIT-blue.svg)](LICENSE)

An unofficial Go SDK for the [Phajay](https://payment-gateway.phajay.co) payment gateway.

Official API documentation: <https://payment-doc.phajay.co/v1>

Currently supports creating payment links, with more payment methods on the way.

## Installation

```bash
go get github.com/barlus-developer/phajay-payment-gateway
```

## Quick start

```go
package main

import (
	"context"
	"fmt"
	"log"

	"github.com/barlus-developer/phajay-payment-gateway"
)

func main() {
	client := phajay.New("your-api-key")

	resp, err := client.CreatePaymentLink(context.Background(), phajay.PaymentLinkRequest{
		OrderNo:     "ORD-20260828-0001",
		Amount:      1500.50,
		Description: "Order #0001",
		Tag1:        "web",
	})
	if err != nil {
		log.Fatal(err)
	}

	fmt.Println("Redirect your customer to:", resp.RedirectURL)
}
```

## Configuration options

`phajay.New` accepts your API key and optional functional options:

| Option         | Type             | Description                                    | Default                          |
| -------------- | ---------------- | ---------------------------------------------- | -------------------------------- |
| `WithBaseURL`  | `string`         | Override the gateway base URL (useful for tests). | `https://payment-gateway.phajay.co` |
| `WithTimeout`  | `time.Duration`  | HTTP client request timeout.                   | `30 * time.Second`               |
| `WithHTTPClient` | `*http.Client` | Supply a fully custom HTTP client.             | `&http.Client{Timeout: 30s}`     |

```go
client := phajay.New(
	"your-api-key",
	phajay.WithBaseURL("https://sandbox.payment-gateway.phajay.co"),
	phajay.WithTimeout(15*time.Second),
)
```

## Authentication

Your API key is sent as HTTP Basic authentication — the key is base64-encoded and used as the credentials. All requests are sent over HTTPS. Keep your key secret and never commit it to version control.

## API reference

### `New(key string, opts ...Option) *Phajay`

Creates a new Phajay payment gateway client.

### `CreatePaymentLink(ctx context.Context, request PaymentLinkRequest) (PaymentLinkResponse, error)`

Creates a payment link and returns a redirect URL to send the customer to.

#### `PaymentLinkRequest`

| Field         | Type     | Required | Description                  |
| ------------- | -------- | -------- | ---------------------------- |
| `OrderNo`     | `string` | no       | Your reference for the order. |
| `Amount`      | `float64`| **yes**  | Payment amount. Must be greater than zero. |
| `Description` | `string` | **yes**  | Short description of the payment. |
| `Tag1`        | `string` | no       | Optional tag/label.          |
| `Tag2`        | `string` | no       | Optional tag/label.          |
| `Tag3`        | `string` | no       | Optional tag/label.          |

#### `PaymentLinkResponse`

| Field         | Type     | Description                               |
| ------------- | -------- | ----------------------------------------- |
| `Message`     | `string` | Response message from the gateway.        |
| `RedirectURL` | `string` | URL to send the customer to for payment.  |
| `OrderNo`     | `string` | The order number echoed back.             |

### Webhook callbacks

Phajay POSTs a `WebhookCallback` payload to your callback URL when a payment reaches a terminal state. Decode the request body into `WebhookCallback` to inspect the payment result:

```go
func handleWebhook(w http.ResponseWriter, r *http.Request) {
	var cb phajay.WebhookCallback
	if err := json.NewDecoder(r.Body).Decode(&cb); err != nil {
		http.Error(w, "invalid body", http.StatusBadRequest)
		return
	}

	if cb.Status != nil && *cb.Status == phajay.WebhookStatusCompleted {
		// Fulfil the order
	}
}
```

Different banks may return different subsets of the payload. The fields marked **guaranteed** are always present; all other fields are `*string` and are `nil` when the bank did not send them.

#### `WebhookCallback`

| Field            | Type       | Description                              |
| ---------------- | ---------- | ---------------------------------------- |
| `PaymentMethod`  | `string`   | **Guaranteed.** Payment method used.     |
| `LinkCode`       | `string`   | **Guaranteed.** The payment link code.   |
| `TransactionID`  | `string`   | **Guaranteed.** Gateway transaction ID.  |
| `OrderNo`        | `string`   | **Guaranteed.** The order number echoed back. |
| `TxnAmount`      | `float64`  | **Guaranteed.** Transaction amount.      |
| `Message`        | `*string`  | Gateway response message.                |
| `RefNo`          | `*string`  | Gateway reference number.                |
| `ExReferenceNo`  | `*string`  | External reference number.               |
| `MerchantName`   | `*string`  | Merchant name.                           |
| `Memo`           | `*string`  | Raw payment memo.                        |
| `TxnDateTime`    | `*string`  | Transaction date and time.               |
| `BillNumber`     | `*string`  | Bill number.                             |
| `SourceAccount`  | `*string`  | Source account number.                   |
| `SourceName`     | `*string`  | Source account name.                     |
| `SourceCurrency` | `*string`  | Source account currency.                 |
| `PaymentID`      | `*string`  | Gateway payment ID.                      |
| `Status`         | `*string`  | Terminal payment status.                 |
| `Description`    | `*string`  | Order description.                       |
| `Remark`         | `*string`  | Remark.                                  |
| `Tag1`–`Tag6`    | `*string`  | Optional tags.                           |
| `UserID`         | `*string`  | User ID.                                 |
| `SuccessURL`     | `*string`  | URL to redirect the customer to.         |

`WebhookStatusCompleted` (`"PAYMENT_COMPLETED"`) indicates a successful payment.

## Error handling

`CreatePaymentLink` returns an error in the following cases:

- **Validation errors** — `amount` is zero or negative, or `description` is empty. These fail before any request is sent.
- **HTTP errors** — any non-2xx response, including the response body in the error message.
- **Transport/encoding errors** — request construction, network, and JSON decode failures.

```go
resp, err := client.CreatePaymentLink(ctx, phajay.PaymentLinkRequest{Amount: 0, Description: ""})
if err != nil {
	fmt.Println(err) // phajay: amount must be greater than zero
}
```

## Testing

```bash
go test ./...
```

## License

[MIT](LICENSE)
