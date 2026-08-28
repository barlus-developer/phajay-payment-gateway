# Phajay Payment Gateway SDK for Go

[![Go Version](https://img.shields.io/badge/Go-1.23-00ADD8?logo=go)](https://go.dev/dl/)
[![Unit Tests](https://img.shields.io/github/actions/workflow/status/barlus-developer/phajay-payment-gateway/test.yml?branch=main&label=unit%20tests)](https://github.com/barlus-developer/phajay-payment-gateway/actions/workflows/test.yml)
[![Go Reference](https://pkg.go.dev/badge/github.com/barlus-developer/phajay-payment-gateway.svg)](https://pkg.go.dev/github.com/barlus-developer/phajay-payment-gateway)
[![License](https://img.shields.io/badge/License-MIT-blue.svg)](LICENSE)

An unofficial Go SDK for the [Phajay](https://payment-gateway.phajay.co) payment
gateway. Official API documentation: <https://payment-doc.phajay.co/v1>.

## Overview

The SDK provides a typed client for the Phajay payment gateway with support for:

- **Payment links** — create a redirect URL for your customer.
- **Bank QR codes** — generate payment QR strings for JDB, LDB, IB, BCEL, STB,
  and M MoneyX.
- **Webhook callbacks** — decode terminal-state notifications for both payment
  links and QR payments.

## Table of contents

- [Installation](#installation)
- [Quick start](#quick-start)
- [Configuration](#configuration)
- [Authentication](#authentication)
- [API reference](#api-reference)
  - [Client](#client)
  - [Payment links](#payment-links)
  - [Payment link webhooks](#payment-link-webhooks)
  - [QR codes](#qr-codes)
  - [QR webhooks](#qr-webhooks)
- [Error handling](#error-handling)
- [Testing](#testing)
- [License](#license)

## Installation

```bash
go get github.com/barlus-developer/phajay-payment-gateway
```

Requires Go 1.23 or later. The SDK has no third-party dependencies.

## Quick start

```go
package main

import (
	"context"
	"fmt"
	"log"

	phajay "github.com/barlus-developer/phajay-payment-gateway"
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

## Configuration

`phajay.New` takes your API key followed by optional functional options:

| Option          | Type            | Description                                        | Default                             |
| --------------- | --------------- | -------------------------------------------------- | ----------------------------------- |
| `WithBaseURL`   | `string`        | Overrides the gateway base URL.                    | `https://payment-gateway.phajay.co` |
| `WithTimeout`   | `time.Duration` | HTTP request timeout.                              | `30 * time.Second`                  |
| `WithHTTPClient`| `*http.Client`  | Supplies a fully custom HTTP client.               | `&http.Client{Timeout: 30s}`        |

```go
client := phajay.New(
	"your-api-key",
	phajay.WithBaseURL("https://sandbox.payment-gateway.phajay.co"),
	phajay.WithTimeout(15*time.Second),
)
```

## Authentication

Each endpoint authenticates using a different scheme:

| Endpoint                     | Scheme                              | Description                                    |
| ---------------------------- | ----------------------------------- | ---------------------------------------------- |
| Payment links                | HTTP Basic authentication           | API key is base64-encoded and sent as the `Authorization` header. |
| QR code generation           | `secretKey` request header          | API key is sent as the `secretKey` header.     |

All requests are sent over HTTPS. Treat your API key as a secret — never commit
it to version control or expose it in client-side code.

## API reference

### Client

`New(key string, opts ...Option) *Phajay` constructs a new client. See
[Configuration](#configuration) for available options.

### Payment links

`CreatePaymentLink(ctx context.Context, request PaymentLinkRequest) (PaymentLinkResponse, error)`
creates a payment link and returns the redirect URL to send your customer to.

**`PaymentLinkRequest`**

| Field         | Type      | Required | Description                          |
| ------------- | --------- | -------- | ------------------------------------ |
| `OrderNo`     | `string`  | no       | Your internal reference for the order. |
| `Amount`      | `float64` | **yes**  | Payment amount. Must be greater than zero. |
| `Description` | `string`  | **yes**  | Short description of the payment.    |
| `Tag1`        | `string`  | no       | Optional label for your reference.   |
| `Tag2`        | `string`  | no       | Optional label for your reference.   |
| `Tag3`        | `string`  | no       | Optional label for your reference.   |

**`PaymentLinkResponse`**

| Field         | Type     | Description                              |
| ------------- | -------- | ---------------------------------------- |
| `Message`     | `string` | Response message from the gateway.       |
| `RedirectURL` | `string` | URL to send the customer to for payment. |
| `OrderNo`     | `string` | The order number echoed back.            |

### Payment link webhooks

When a payment link reaches a terminal state, Phajay POSTs a
`PaymentLinkWebhookCallback` payload to your callback URL. Decode the request
body to inspect the result:

```go
func handlePaymentLinkWebhook(w http.ResponseWriter, r *http.Request) {
	var cb phajay.PaymentLinkWebhookCallback
	if err := json.NewDecoder(r.Body).Decode(&cb); err != nil {
		http.Error(w, "invalid body", http.StatusBadRequest)
		return
	}

	if cb.Status != nil && *cb.Status == phajay.PaymentLinkWebhookStatusCompleted {
		// Fulfil the order
	}
}
```

Banks may return different subsets of the payload. Fields marked
**guaranteed** are always present; all other fields are `*string` and are `nil`
when the bank omits them.

**`PaymentLinkWebhookCallback`**

| Field            | Type      | Description                               |
| ---------------- | --------- | ----------------------------------------- |
| `PaymentMethod`  | `string`  | **Guaranteed.** Payment method used.      |
| `LinkCode`       | `string`  | **Guaranteed.** The payment link code.    |
| `TransactionID`  | `string`  | **Guaranteed.** Gateway transaction ID.   |
| `OrderNo`        | `string`  | **Guaranteed.** The order number echoed back. |
| `TxnAmount`      | `float64` | **Guaranteed.** Transaction amount.       |
| `Message`        | `*string` | Gateway response message.                 |
| `RefNo`          | `*string` | Gateway reference number.                 |
| `ExReferenceNo`  | `*string` | External reference number.                |
| `MerchantName`   | `*string` | Merchant name.                            |
| `Memo`           | `*string` | Raw payment memo.                         |
| `TxnDateTime`    | `*string` | Transaction date and time.                |
| `BillNumber`     | `*string` | Bill number.                              |
| `SourceAccount`  | `*string` | Source account number.                    |
| `SourceName`     | `*string` | Source account name.                      |
| `SourceCurrency` | `*string` | Source account currency.                  |
| `PaymentID`      | `*string` | Gateway payment ID.                       |
| `Status`         | `*string` | Terminal payment status.                  |
| `Description`    | `*string` | Order description.                        |
| `Remark`         | `*string` | Remark.                                   |
| `Tag1`–`Tag6`    | `*string` | Optional tags.                            |
| `UserID`         | `*string` | User ID.                                  |
| `SuccessURL`     | `*string` | URL to redirect the customer to.          |

`PaymentLinkWebhookStatusCompleted` (`"PAYMENT_COMPLETED"`) indicates a
successful payment.

### QR codes

`GenerateQR(ctx context.Context, bank Bank, request PaymentQRRequest) (PaymentQRResponse, error)`
generates a QR code string for a bank so the customer can pay through the
bank's Mobile Banking app.

```go
resp, err := client.GenerateQR(context.Background(), phajay.BankBCEL, phajay.PaymentQRRequest{
	Amount:      1500.50,
	Description: "Order #0001",
})
if err != nil {
	log.Fatal(err)
}

fmt.Println("QR string:", resp.QRCode)
fmt.Println("Deeplink:", resp.Link)
```

**Supported banks**

| Constant        | Bank                                              |
| --------------- | ------------------------------------------------- |
| `BankJDB`       | Joint Development Bank (JDB)                      |
| `BankLDB`       | Lao Development Bank (LDB)                        |
| `BankIB`        | Indochina Bank (IB)                               |
| `BankBCEL`      | Banque Pour Le Commerce Exterieur Lao Public (BCEL) |
| `BankSTB`       | ST Bank Laos (STB)                                |
| `BankMMoneyX`   | M MoneyX                                          |

> **Note:** BCEL does not currently support Thai/Lao characters in
> `Description`.

**`PaymentQRRequest`**

| Field         | Type      | Required | Description                          |
| ------------- | --------- | -------- | ------------------------------------ |
| `Amount`      | `float64` | **yes**  | Amount to be paid. Must be greater than zero. |
| `Description` | `string`  | **yes**  | Payment description.                 |
| `Tag1`        | `string`  | no       | Custom field for your internal reference. |
| `Tag2`        | `string`  | no       | Custom field for your internal reference. |
| `Tag3`        | `string`  | no       | Custom field for your internal reference. |

**`PaymentQRResponse`**

| Field           | Type     | Description                        |
| --------------- | -------- | ---------------------------------- |
| `Message`       | `string` | Response message from the gateway. |
| `TransactionID` | `string` | Gateway transaction ID.            |
| `QRCode`        | `string` | The QR string of the transaction.  |
| `Link`          | `string` | Deeplink to open the bank's app.   |

### QR webhooks

When a QR payment reaches a terminal state, Phajay POSTs a
`PaymentQRWebhookCallback` payload to your callback URL. Decode the request body
to inspect the result:

```go
func handleQRWebhook(w http.ResponseWriter, r *http.Request) {
	var cb phajay.PaymentQRWebhookCallback
	if err := json.NewDecoder(r.Body).Decode(&cb); err != nil {
		http.Error(w, "invalid body", http.StatusBadRequest)
		return
	}

	if cb.Status == phajay.PaymentQRWebhookStatusCompleted {
		// Fulfil the order
	}
}
```

As with payment link webhooks, fields marked **guaranteed** are always present;
all other fields are `*string` and are `nil` when the bank omits them.

**`PaymentQRWebhookCallback`**

| Field            | Type              | Description                             |
| ---------------- | ----------------- | --------------------------------------- |
| `PaymentMethod`  | `string`          | **Guaranteed.** Payment method used.    |
| `TransactionID`  | `string`          | **Guaranteed.** Gateway transaction ID. |
| `BillNumber`     | `string`          | **Guaranteed.** Bill number.            |
| `TxnAmount`      | `float64`         | **Guaranteed.** Transaction amount.     |
| `Status`         | `string`          | **Guaranteed.** Terminal payment status. |
| `Message`        | `*string`         | Gateway response message.               |
| `RefNo`          | `*FlexibleString` | Gateway reference number.               |
| `ExReferenceNo`  | `*string`         | External reference number.              |
| `MerchantName`   | `*string`         | Merchant name.                          |
| `Description`    | `*string`         | Order description.                      |
| `TxnDateTime`    | `*string`         | Transaction date and time.              |
| `SourceAccount`  | `*string`         | Source account number.                  |
| `SourceName`     | `*string`         | Source account name.                    |
| `SourceCurrency` | `*string`         | Source account currency.                |
| `UserID`         | `*string`         | User ID.                                |
| `Tag1`–`Tag6`    | `*string`         | Optional tags.                          |

`PaymentQRWebhookStatusCompleted` (`"PAYMENT_COMPLETED"`) indicates a
successful payment.

**`FlexibleString`**

The `RefNo` field is typed as `*FlexibleString` because Phajay returns it as a
string for some banks and a JSON number for others. `FlexibleString`
normalises both forms to a decimal string, so the value is always accessed as
a string:

```go
if cb.RefNo != nil {
	fmt.Println("Ref no:", string(*cb.RefNo))
}
```

## Error handling

`CreatePaymentLink` and `GenerateQR` return an error in the following cases:

- **Validation errors** — `amount` is zero or negative, or `description` is
  empty. These fail before any request is sent.
- **Unknown bank** — `GenerateQR` returns an error when passed a `Bank` value
  that is not supported.
- **HTTP errors** — any non-2xx response, including the response body in the
  error message.
- **Transport/encoding errors** — request construction, network, and JSON
  decode failures.

```go
resp, err := client.CreatePaymentLink(ctx, phajay.PaymentLinkRequest{Amount: 0, Description: ""})
if err != nil {
	fmt.Println(err) // phajay: amount must be greater than zero
}
```

## Testing

```bash
go test ./... -race -cover
```

The test suite is also run on every push and pull request via the
[`Unit Tests`](.github/workflows/test.yml) workflow.

## License

[MIT](LICENSE)
