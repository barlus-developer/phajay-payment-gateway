# Phajay Payment Gateway

[![Go Version](https://img.shields.io/badge/Go-1.23-00ADD8?logo=go)](https://go.dev/dl/)
[![License](https://img.shields.io/badge/License-MIT-blue.svg)](LICENSE)

A Go client library for the [Phajay](https://payment-gateway.phajay.co) payment gateway.

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

`phajay.New` accepts a key and optional functional options:

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

The API key is sent as HTTP Basic authentication — the key is base64-encoded and used as the credentials. All requests are sent over HTTPS.

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
