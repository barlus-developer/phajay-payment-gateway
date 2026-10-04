# Phajay Payment Gateway SDK for Go

[![Go Version](https://img.shields.io/badge/Go-1.23-00ADD8?logo=go)](https://go.dev/dl/)
[![Unit Tests](https://img.shields.io/github/actions/workflow/status/barlus-developer/phajay-payment-gateway/test.yml?branch=main&label=unit%20tests)](https://github.com/barlus-developer/phajay-payment-gateway/actions/workflows/test.yml)
[![Go Reference](https://pkg.go.dev/badge/github.com/barlus-developer/phajay-payment-gateway.svg)](https://pkg.go.dev/github.com/barlus-developer/phajay-payment-gateway)
[![License](https://img.shields.io/badge/License-MIT-blue.svg)](LICENSE)

An unofficial Go SDK for the [Phajay](https://payment-gateway.phajay.co) payment
gateway. Official API documentation: <https://payment-doc.lailaolab.com/v1>.

## Overview

The SDK provides a typed client for the Phajay payment gateway with support for:

- **Payment links** — create a redirect URL for your customer.
- **Bank QR codes** — generate payment QR strings for JDB, LDB, IB, BCEL, STB,
  and M MoneyX.
- **Transaction status** — confirm payment by transaction ID.
- **Refunds** — request a refund and check its bill status.
- **Credit cards** — create a hosted 3DS payment page and decode callbacks.
- **Subscriptions** — authorize BCEL recurring payments, cancel authorizations,
  and decode setup, debit, and cancellation callbacks.
- **Sandbox** — use the documented test routes for payment links and bank QR.
- **Webhook callbacks** — decode bank, card, subscription, and sandbox link
  notifications.

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
  - [Transaction status](#transaction-status)
  - [Refunds](#refunds)
  - [Credit card payments](#credit-card-payments)
  - [Subscriptions](#subscriptions)
  - [Sandbox callbacks](#sandbox-callbacks)
- [Webhook verification](#webhook-verification)
- [Migration notes](#migration-notes)
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
| `WithHTTPClient`| `*http.Client`  | Copies a custom HTTP client; nil keeps the default. | `&http.Client{Timeout: 30s}`        |
| `WithSandbox`   | none            | Selects documented test paths for links and QR.     | production                         |

```go
client := phajay.New(
	"your-test-key",
	phajay.WithSandbox(),
	phajay.WithTimeout(15*time.Second),
)
```

Options are applied in order. `WithTimeout` after `WithHTTPClient` overrides
that client's timeout without modifying the supplied client. In the reverse
order, the custom client's timeout takes precedence. A zero timeout disables it.
Clients may be shared between goroutines after construction.

Sandbox uses the same `https://payment-gateway.phajay.co` host as production.
`WithSandbox()` changes payment link creation to
`/v1/api/test/payment/get-payment-link` and QR generation to
`/v1/api/test/payment/generate-{bank}-qr`. JDB, LDB, IB, BCEL, and STB are
documented; M MoneyX is unavailable in sandbox. The status, refund, card, and
subscription methods fail before sending a request in sandbox mode because no
test routes are documented. Use the portal test key and scan test QR codes with
the [PhaJay App](https://payment-doc.lailaolab.com/v1/sandbox/paymentQR-sandbox).

## Authentication

Each endpoint authenticates using a different scheme:

| Endpoint                     | Scheme                              | Description                                    |
| ---------------------------- | ----------------------------------- | ---------------------------------------------- |
| Payment links                | HTTP Basic authentication           | API key is base64-encoded and sent as the `Authorization` header. |
| Transaction status, credit card links | HTTP Basic authentication | Base64 of the raw API key. |
| QR codes, refunds, subscriptions | `secretKey` request header | Raw API key. |

The default gateway URL uses HTTPS. Treat your API key as a secret — never commit
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
| `OrderNo`     | `string`  | no       | Your order reference; must be unique when supplied. |
| `Amount`      | `float64` | **yes**  | Payment amount. Must be finite and greater than zero. |
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
	// Verify the raw request signature first using the Phajay support guide.
	var cb phajay.PaymentLinkWebhookCallback
	if err := json.NewDecoder(r.Body).Decode(&cb); err != nil {
		http.Error(w, "invalid body", http.StatusBadRequest)
		return
	}

	if cb.Status != nil && *cb.Status == phajay.PaymentLinkWebhookStatusCompleted {
		// Match your saved order and amount, then fulfil once (idempotently).
	}
	w.WriteHeader(http.StatusOK)
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
| `RefNo`          | `*string` | Reference number; numeric JSON is also accepted.                 |
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
> `Description`. BCEL QR codes can only be paid with the BCEL One app.

**`PaymentQRRequest`**

| Field         | Type      | Required | Description                          |
| ------------- | --------- | -------- | ------------------------------------ |
| `OrderNo`     | `string`  | no       | Recommended unique order reference; omitted when empty. |
| `Amount`      | `float64` | **yes**  | Amount to be paid. Must be finite and greater than zero. |
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
	// Verify the raw request signature first using the Phajay support guide.
	var cb phajay.PaymentQRWebhookCallback
	if err := json.NewDecoder(r.Body).Decode(&cb); err != nil {
		http.Error(w, "invalid body", http.StatusBadRequest)
		return
	}

	if cb.Status != nil && *cb.Status == phajay.PaymentQRWebhookStatusCompleted {
		// Match your saved order and amount, then fulfil once (idempotently).
	}
	w.WriteHeader(http.StatusOK)
}
```

As with payment link webhooks, fields marked **guaranteed** are always present;
all other fields are pointers and are `nil` when omitted or null.

**`PaymentQRWebhookCallback`**

| Field            | Type              | Description                             |
| ---------------- | ----------------- | --------------------------------------- |
| `PaymentMethod`  | `string`          | **Guaranteed.** Payment method used.    |
| `TransactionID`  | `string`          | **Guaranteed.** Gateway transaction ID. |
| `BillNumber`     | `*string`         | Optional bill number.            |
| `TxnAmount`      | `float64`         | **Guaranteed.** Transaction amount.     |
| `Status`         | `*string`         | Optional payment status. |
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
preserves the text of strings and numbers without rounding large integers.
Access the value as a string:

```go
if cb.RefNo != nil {
	fmt.Println("Ref no:", string(*cb.RefNo))
}
```

### Transaction status

`CheckTransactionStatus(ctx context.Context, transactionID string) (TransactionStatusResponse, error)`
uses `GET /v1/api/payment/check-transaction/status/{transactionId}` with Basic
auth. See the [official status API](https://payment-doc.lailaolab.com/v1/check-transaction-status).

```go
resp, err := client.CheckTransactionStatus(ctx, transactionID)
if err != nil {
    return err
}
if resp.Data != nil && resp.Data.IsPaid && resp.Data.Status == phajay.TransactionStatusCompleted {
    // Match transaction ID, order number, amount, and currency to your saved order.
    // Fulfil the order once.
}
```

**`TransactionStatusResponse`**

| Field | Type | Description |
| --- | --- | --- |
| `Message` | `string` | Gateway result, including `PAYMENT_NOT_FOUND`. |
| `Data` | `*TransactionStatus` | Nil for a message-only or null-data response. |

**`TransactionStatus`**

| Field | Type | Description |
| --- | --- | --- |
| `TransactionID` | `string` | Gateway transaction ID. |
| `LinkCode` | `*string` | Nullable payment link identifier. |
| `OrderNo` | `*string` | Nullable platform order reference. |
| `Status` | `string` | Current transaction state. |
| `IsPaid` | `bool` | Gateway payment flag; require true before fulfilment. |
| `Amount` | `float64` | Transaction amount. |
| `Currency` | `string` | Currency code, such as LAK. |
| `Description` | `string` | Payment description. |
| `PaymentMethod` | `string` | Bank or payment method. |
| `PaymentTime` | `*string` | Nullable completion timestamp. |
| `CreatedAt` | `string` | Creation timestamp. |

Status constants are `TransactionStatusWaiting` (`WAITING`),
`TransactionStatusExpired` (`EXPIRED`), `TransactionStatusRefunded` (`REFUNDED`),
`TransactionStatusDisputed` (`DISPUTED`), and `TransactionStatusCompleted`
(`PAYMENT_COMPLETED`). Status fields remain strings to support future values.
Timestamps retain their wire representation.

### Refunds

`RequestRefund(ctx context.Context, request RefundRequest) (RefundResponse, error)`
sends `POST /v1/api/refund`. `CheckRefundStatus(ctx context.Context, billID string)
(RefundResponse, error)` sends `GET /v1/api/refund/{billID}`. Both use the
`secretKey` header. See the [official refund API](https://payment-doc.lailaolab.com/v1/refund).

```go
refund, err := client.RequestRefund(ctx, phajay.RefundRequest{TransactionID: transactionID})
if err != nil {
    return err
}
// A successful request starts the refund workflow. Poll by Data.ID, not BillNo.
status, err := client.CheckRefundStatus(ctx, refund.Data.ID)
if err != nil {
    return err
}
fmt.Println(status.Data.Status)
```

**`RefundRequest`, `RefundResponse`, and `Refund`**

| Type / field | Go type | Description |
| --- | --- | --- |
| `RefundRequest.TransactionID` | `string` | Required payment transaction ID. |
| `RefundResponse.Message` | `string` | Gateway result. |
| `RefundResponse.Data` | `Refund` | Refund bill. |
| `Refund.ID` | `string` | Refund bill ID for status checks. |
| `Refund.Status` | `string` | Refund workflow state, such as `REQUESTING`. |
| `Refund.Amount` | `float64` | Accepts numeric JSON or a numeric string. |
| `Refund.CreatedAt` | `string` | Creation timestamp. |
| `Refund.BillNo` | `string` | Display bill number; differs from ID. |

Refund requests require an approved KYC account. Completion is asynchronous;
a request response alone does not confirm that money has been returned.

### Credit card payments

`CreateCreditCardPaymentLink(ctx context.Context, request CreditCardPaymentRequest)
(CreditCardPaymentResponse, error)` sends
`POST /v1/api/jdb2c2p/payment/payment-link` with Basic auth. Redirect the customer
to `PaymentURL`; card data is entered on the hosted payment page. Phajay requires
an approved KYC account and a 3DS-enabled card. See the
[official card API](https://payment-doc.lailaolab.com/v1/connect-credit-card).

**`CreditCardPaymentRequest`**

| Field | Type | Required | Description |
| --- | --- | --- | --- |
| `Amount` | `float64` | yes | Finite positive payment amount. |
| `Description` | `string` | yes | Non-blank payment description. |
| `Tag1`–`Tag3` | `string` | no | Internal metadata; omitted when empty. |

**`CreditCardPaymentResponse`**

| Field | Type | Description |
| --- | --- | --- |
| `Message` | `string` | Gateway result. |
| `PaymentURL` | `string` | Hosted card payment page (`paymentUrl` on the wire). |
| `ExpirationTime` | `string` | Gateway date-time text; the example has no timezone suffix. |
| `TransactionID` | `string` | Gateway transaction ID. |
| `Status` | `string` | Current payment state. |

**`CreditCardWebhookCallback`**

The card guide does not guarantee a mandatory subset. Every field is optional
and remains nil when missing or null. Acknowledge a verified callback with HTTP
200 and process completed payments idempotently.

| Field | Type | Description |
| --- | --- | --- |
| `Message`, `Status` | `*string` | Gateway result and payment state. |
| `TransactionID`, `LinkCode` | `*string` | Payment identifiers. |
| `PaymentMethod` | `*string` | `CREDIT_CARD` in the documented payload. |
| `UserID`, `Description`, `SuccessURL` | `*string` | Account ID, description, redirect URL. |
| `Tag1`–`Tag6` | `*string` | Optional internal metadata. |
| `TxnAmount` | `*float64` | Payment amount. |
| `Currency`, `CardType` | `*string` | Currency and card brand. |
| `CreditCardDetails` | `*CreditCardDetails` | Masked card details. |
| `TransactionAmount` | `*CreditCardTransactionAmount` | Amount encoding and currency. |

**Nested card fields**

| Type / field | Go type | Description |
| --- | --- | --- |
| `CreditCardDetails.CardHolderName`, `.CardNumber`, `.CardExpiry` | `*string` | Holder name, masked number, masked expiry. |
| `CreditCardDetails.IssuerCode`, `.AuthenticationStatus` | `*string` | Nullable issuer and authentication information. |
| `CreditCardDetails.ECIValue`, `.XIDValue`, `.CAVVValue`, `.EnrolledFlag` | `*string` | Nullable 3DS metadata. |
| `CreditCardTransactionAmount.AmountText` | `*string` | Gateway amount text. |
| `CreditCardTransactionAmount.CurrencyCode` | `*string` | Currency code. |
| `CreditCardTransactionAmount.DecimalPlaces` | `*int` | Number of decimal places. |
| `CreditCardTransactionAmount.Amount` | `*float64` | Numeric amount. |

### Subscriptions

`GenerateSubscriptionQR(ctx context.Context, request SubscriptionQRRequest)
(SubscriptionQRResponse, error)` sends
`POST /v1/api/subscription/generate-bcel-qr`. `CancelSubscription(ctx
context.Context, request CancelSubscriptionRequest) (CancelSubscriptionResponse,
error)` sends `POST /v1/api/subscription/cancel-subscription`. Both authenticate
with `secretKey`. Subscriptions require approved KYC and configured setup and
subscription webhooks. See the [subscription guide](https://payment-doc.lailaolab.com/v1/connect-subscription/generateQR).

```go
resp, err := client.GenerateSubscriptionQR(ctx, phajay.SubscriptionQRRequest{
    MaxAmount:          1000,
    SubscriptionDate:   "2026-10-04",
    ResubscriptionDays: 30,
    Description:        "Monthly plan",
})
if err != nil {
    return err
}
fmt.Println(resp.QRCode)
// Save the authCode received in the subscription webhook for cancellation.
```

**Subscription request and response fields**

| Type / field | Go type | Description |
| --- | --- | --- |
| `SubscriptionQRRequest.MaxAmount` | `float64` | Required finite positive authorization limit. |
| `SubscriptionQRRequest.SubscriptionDate` | `string` | Required valid date in `YYYY-MM-DD` format. |
| `SubscriptionQRRequest.ResubscriptionDays` | `int` | Required positive recurrence interval in days. |
| `SubscriptionQRRequest.Description` | `string` | Required non-blank description. |
| `SubscriptionQRResponse.Message` | `string` | Gateway result. |
| `SubscriptionQRResponse.TransactionID` | `string` | Authorization transaction ID. |
| `SubscriptionQRResponse.QRCode` | `string` | QR value; the official example returns an image URL. |
| `SubscriptionQRResponse.Link` | `string` | Bank app deep link. |
| `CancelSubscriptionRequest.AuthCode` | `string` | Required authorization code received by webhook. |
| `CancelSubscriptionResponse.Message`, `.Status` | `string` | Cancellation result and state. |
| `CancelSubscriptionResponse.TransactionID`, `.AuthCode` | `string` | Authorization identifiers. |
| `CancelSubscriptionResponse.Time` | `string` | Gateway event timestamp. |

Phajay schedules debits automatically; the guide describes debit callbacks,
with no manual debit endpoint. For a future start date, the setup webhook
reports authorization. The guide says this setup event is skipped when the
start date is the current date.

**`SubscriptionWebhookCallback`**

| Field | Type | Description |
| --- | --- | --- |
| `Message`, `Status` | `string` | Shared event result and state. |
| `TransactionID`, `AuthCode` | `string` | Shared authorization identifiers. |
| `Time` | `string` | Shared event timestamp. |
| `PaymentTransactionID` | `*string` | Debit payment ID; nil for setup and cancellation. |

| Constant | Value | Event |
| --- | --- | --- |
| `SubscriptionStatusConnected` | `SUBSCRIPTION_CONNECTED` | Authorization accepted. |
| `SubscriptionStatusSuccess` | `SUBSCRIPTION_SUCCESS` | Recurring debit succeeded. |
| `SubscriptionStatusCanceled` | `CANCEL` | Authorization canceled. |

Configure the portal's subscription setup webhook and subscription webhook,
then acknowledge verified events with HTTP 200. Cancellation through the bank
app arrives on the subscription webhook.

### Sandbox callbacks

`SandboxPaymentLinkWebhookCallback` models the
[documented sandbox link payload](https://payment-doc.lailaolab.com/v1/sandbox/paymentLink-sandbox),
which differs from production. All fields are optional pointers.

| Field | Type | JSON field / description |
| --- | --- | --- |
| `Message`, `Status` | `*string` | `message`, `status`. |
| `ID` | `*string` | `_id` document identifier. |
| `OrderNo`, `LinkCode` | `*string` | `orderNo`, `linkCode`. |
| `Amount` | `*float64` | `amount`, rather than production `txnAmount`. |
| `Currency` | `*string` | `currency`. |
| `Description` | `*string` | `descriptione`, using the spelling in the official table. |
| `User` | `*string` | `user`, rather than production `userId`. |
| `Tag1`–`Tag6` | `*string` | `tag1`–`tag6`. |
| `IsMerchantIDPayment`, `IsAffiliatePayment` | `*bool` | `isMerchantIdPayment`, `isAffiliatePayment`. |
| `CreatedAt`, `UpdatedAt` | `*string` | `createdAt`, `updateAt`, preserving the documented spelling. |

## Webhook verification

The [official verification page](https://payment-doc.lailaolab.com/v1/verify-webhook-signature)
requires contacting Phajay support for the signing method and key; neither is
public. The SDK decodes payloads but does not authenticate incoming requests.
Before using the webhook examples, verify the signature over the original
request body following the guide supplied for your platform. Treat redirects
as navigation; confirm payment using a verified webhook or the authenticated
status API, and match saved order identifiers, amount, and currency.

## Migration notes

`PaymentQRWebhookCallback.BillNumber` and `.Status` changed from `string` to
`*string` because they are not guaranteed by the current callback guide. Check
nil before reading or comparing them:

```go
if cb.Status != nil && *cb.Status == phajay.PaymentQRWebhookStatusCompleted {
    // Process a verified payment.
}
if cb.BillNumber != nil {
    fmt.Println(*cb.BillNumber)
}
```

`PaymentLinkWebhookCallback.RefNo` remains `*string` and now also decodes numeric
references. `WithHTTPClient` copies the client before options are applied, so
`WithTimeout` no longer changes the caller's client. Sandbox configuration now
uses `WithSandbox()` and a test key with the default gateway host.

## Error handling

All client methods return errors for invalid input, HTTP failures, request
construction, transport, and JSON decoding failures. Amounts must be finite
and positive, descriptions must contain non-whitespace text, and required IDs
must be non-empty. Subscription dates must use `YYYY-MM-DD` and recurrence
intervals must be positive. Validation fails before sending any request.

`APIError` exposes non-2xx response details through `errors.As`:

```go
var apiErr *phajay.APIError
if errors.As(err, &apiErr) {
    fmt.Println(apiErr.StatusCode, apiErr.Operation, apiErr.Body)
}
if errors.Is(err, context.DeadlineExceeded) {
    // Handle a timeout.
}
```

| `APIError` field | Type | Description |
| --- | --- | --- |
| `StatusCode` | `int` | Gateway HTTP status. |
| `Body` | `string` | Trimmed error body, limited to 1 MiB. |
| `Operation` | `string` | Operation that failed. |

Underlying errors are wrapped with `%w`. Every SDK error starts with `phajay: `.
Successful HTTP responses are decoded as returned; inspect gateway messages
and data as well. For example, `PAYMENT_NOT_FOUND` can be a message-only
transaction response with `Data == nil`. The SDK does not automatically retry
requests that create payments, refunds, or authorizations.

## Testing

```bash
go test ./... -race -cover
```

The test suite is also run on every push and pull request via the
[`Unit Tests`](.github/workflows/test.yml) workflow.

## License

[MIT](LICENSE)
