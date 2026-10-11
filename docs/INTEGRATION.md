# Phajay Go SDK Integration Guide

This guide walks through the main payment flows supported by the Go SDK. Use
the [README API reference](../README.md#api-reference) for full request,
response, and webhook field tables. Check the
[official Phajay v1 documentation](https://payment-doc.lailaolab.com/v1) for
gateway requirements that may change.

The examples target the API in this checkout. Check the
[changelog](../CHANGELOG.md) when using a released version.

## Contents

- [Before you start](#before-you-start)
- [Quick start](#quick-start)
- [Create a client](#create-a-client)
- [Choose a payment flow](#choose-a-payment-flow)
- [Test with the sandbox](#test-with-the-sandbox)
- [Confirm a payment](#confirm-a-payment)
- [Handle webhooks](#handle-webhooks)
- [Request and track a refund](#request-and-track-a-refund)
- [Errors and input rules](#errors-and-input-rules)
- [Troubleshooting](#troubleshooting)

## Before you start

- Go 1.23 or later.
- A Phajay portal account and API key. Production card and subscription APIs
  require an approved KYC account.
- A public HTTPS webhook endpoint if your application receives callbacks.

Install the SDK:

```bash
go get github.com/barlus-developer/phajay-payment-gateway
```

Keep the API key on your server. Do not include it in browser or mobile app
code.

## Quick start

This complete example creates a payment link. Replace the sample order ID with
the unique ID from your order system:

```go
package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"time"

	phajay "github.com/barlus-developer/phajay-payment-gateway"
)

func main() {
	secretKey := os.Getenv("PHAJAY_SECRET_KEY")
	if secretKey == "" {
		log.Fatal("PHAJAY_SECRET_KEY is required")
	}

	client := phajay.New(secretKey, phajay.WithTimeout(15*time.Second))
	orderID := "order-123" // Replace with a unique order ID.
	response, err := client.CreatePaymentLink(context.Background(), phajay.PaymentLinkRequest{
		OrderNo:     orderID,
		Amount:      1500,
		Description: "Order " + orderID,
	})
	if err != nil {
		log.Fatal(err)
	}

	fmt.Println("Redirect the customer to:", response.RedirectURL)
}
```

## Create a client

Use the secret key for production. Set a timeout that fits your application:

```go
client := phajay.New(
	secretKey,
	phajay.WithTimeout(15*time.Second),
)
```

Pass the same client to the parts of your service that call Phajay. The client
can be shared across goroutines after construction. Use `WithHTTPClient` if you
need a custom transport or test client.

The remaining code blocks show method fragments inside application code. They
assume that `ctx`, `client`, and application values such as `orderID` are in
scope. Add the standard library imports used in each fragment to your file.

### Request deadlines

In an HTTP handler, pass `r.Context()` so cancellation can reach the gateway
request. For a background operation, create a deadline and release its timer:

```go
ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
defer cancel()
```

The client defaults to a 30-second HTTP timeout. A context can set a shorter
deadline for one call. `WithTimeout(0)` disables the client timeout, so use a
context deadline when a request must finish within a defined time.

### Custom HTTP clients

```go
httpClient := &http.Client{Timeout: 20 * time.Second}
client := phajay.New(
	secretKey,
	phajay.WithHTTPClient(httpClient),
	phajay.WithTimeout(15*time.Second),
)
```

Options apply in order; here the SDK client uses 15 seconds. Reversing the last
two options makes the supplied client's 20-second timeout take precedence.
`WithHTTPClient` copies the `http.Client` value, leaving its original timeout
unchanged. The transport and other referenced objects remain shared; custom
transports must support concurrent calls. A nil HTTP client keeps the default.
`WithBaseURL` is useful for a local test server; use a trusted destination
because requests include the API key.

## Choose a payment flow

| Flow | SDK method | Authentication | Result |
| --- | --- | --- | --- |
| Bank payment link | `CreatePaymentLink` | Basic auth | Redirect URL for the customer. |
| Bank QR | `GenerateQR` | `secretKey` header | QR data and optional bank deep link. |
| Credit card | `CreateCreditCardPaymentLink` | Basic auth | Hosted 3DS payment URL. |
| BCEL subscription authorization | `GenerateSubscriptionQR` | `secretKey` header | QR code for bank authorization. |

Use a unique `OrderNo` with payment links and QR payments so callback and
transaction records can be matched to your order. Phajay recommends a unique
order number when provided.

Save a pending payment record before calling the gateway. Keep your internal
order ID, expected amount and currency, chosen flow, and any gateway identifiers
returned by the call. Link creation returns `OrderNo` and `RedirectURL`; QR and
card creation also return `TransactionID`. The QR callback type has no `OrderNo`
field, so store the returned transaction ID for correlation. Card requests have
no `OrderNo` field; use the returned transaction ID to identify the payment.

### Payment link

Create a link on your server, then send `RedirectURL` to the customer's browser:

```go
response, err := client.CreatePaymentLink(ctx, phajay.PaymentLinkRequest{
	OrderNo:     orderID,
	Amount:      1500,
	Description: "Order " + orderID,
})
if err != nil {
	return err
}
redirectURL := response.RedirectURL
```

Configure success, cancel, and webhook URLs in the Phajay portal. A browser
redirect reports navigation. Use an authenticated status check or verified
webhook to decide whether to fulfill the order.

### Bank QR

Select a bank and display `QRCode` for the customer to scan. Use `Link` to open
the bank app when it is available:

```go
response, err := client.GenerateQR(ctx, phajay.BankBCEL, phajay.PaymentQRRequest{
	OrderNo:     orderID,
	Amount:      1500,
	Description: "Order " + orderID,
})
if err != nil {
	return err
}
qrValue := response.QRCode
bankAppLink := response.Link
```

BCEL QR codes must be paid in the BCEL One app. The Phajay guide says BCEL QR
descriptions do not support Lao or Thai characters.

### Credit card

`CreateCreditCardPaymentLink` returns a hosted `PaymentURL`. Redirect the
customer there; the payment page handles card entry. The guide requires a
3D Secure enabled card and an approved KYC account.

```go
response, err := client.CreateCreditCardPaymentLink(ctx, phajay.CreditCardPaymentRequest{
	Amount:      1500,
	Description: "Order " + orderID,
	Tag1:        orderID,
})
if err != nil {
	return err
}
paymentURL := response.PaymentURL
transactionID := response.TransactionID
```

Store the transaction ID with the order and check the eventual payment result.
`ExpirationTime` is returned as gateway text; the SDK does not parse its timezone
or expire the payment locally.

### BCEL subscriptions

`GenerateSubscriptionQR` starts a subscription authorization. Give the
customer the returned QR code to scan and approve in the bank app. Store the
`AuthCode` from the setup webhook; pass it to `CancelSubscription` if the
customer later cancels. The gateway schedules recurring debits. The public guide
does not document a manual debit endpoint.

```go
response, err := client.GenerateSubscriptionQR(ctx, phajay.SubscriptionQRRequest{
	MaxAmount:          1000,
	SubscriptionDate:   startDate, // Application-selected date in YYYY-MM-DD form.
	ResubscriptionDays: 30,
	Description:        "Monthly membership",
})
if err != nil {
	return err
}
authorizationTransactionID := response.TransactionID
qrValue := response.QRCode
```

Subscription `QRCode` can be an image URL, as shown in the
[gateway example](https://payment-doc.lailaolab.com/v1/connect-subscription/generateQR).
Handle that form when displaying it. Creating the QR starts authorization;
record successful recurring charges from debit notifications.

Configure both subscription setup and subscription webhooks in the portal.
The [setup guide](https://payment-doc.lailaolab.com/v1/connect-subscription/setup-subscription)
says the setup notification is skipped when the subscription date is the
current date. Handle the authorization code in the applicable subscription
events instead of relying exclusively on the setup notification.

| Callback status | Application action |
| --- | --- |
| `SubscriptionStatusConnected` | Save the authorization transaction and `AuthCode`. |
| `SubscriptionStatusSuccess` | Save the debit's `PaymentTransactionID`, checking for nil, and process the charge once. |
| `SubscriptionStatusCanceled` | Mark the authorization canceled. |

To cancel a saved authorization:

```go
response, err := client.CancelSubscription(ctx, phajay.CancelSubscriptionRequest{
	AuthCode: authCode,
})
if err != nil {
	return err
}
cancellationStatus := response.Status
```

## Test with the sandbox

Use a portal test key and `WithSandbox()`:

```go
client := phajay.New(testKey, phajay.WithSandbox())
```

The sandbox uses the gateway host with separate test paths. It supports payment
links and QR generation for JDB, LDB, IB, BCEL, and STB. Scan sandbox QR codes
with the PhaJay app; a real bank app cannot complete a sandbox payment.

M MoneyX QR, transaction status, refunds, credit card payments, and
subscriptions have no documented sandbox paths. Those methods return an error
before sending a request when the client uses `WithSandbox()`.

## Confirm a payment

Use the transaction ID returned when creating a payment, or received in a
callback, with `CheckTransactionStatus`:

```go
result, err := client.CheckTransactionStatus(ctx, transactionID)
if err != nil {
	return err
}
if result.Data == nil {
	// The gateway may return a message such as PAYMENT_NOT_FOUND.
	return fmt.Errorf("transaction status unavailable: %s", result.Message)
}
if result.Data.IsPaid && result.Data.Status == phajay.TransactionStatusCompleted {
	// Match transaction ID, order number, amount, and currency to the saved order.
	// Fulfill once; callbacks and status checks can both report the same payment.
}
```

Require both `IsPaid` and `PAYMENT_COMPLETED`. Compare the returned identifiers,
amount, and currency with your saved order. Make fulfillment idempotent so a
repeated callback or status check cannot fulfill the same order twice.

Keep orders pending while the gateway reports `WAITING`. Handle `EXPIRED`,
`REFUNDED`, and `DISPUTED` according to your order workflow. Treat unknown status
strings as unrecognized states. A successful HTTP response with `Data == nil`
does not establish whether an order was paid. Status checks require a production
client.

## Handle webhooks

Configure the relevant webhook URL in the Phajay portal. Before decoding or
using callback data, verify the signature over the original request body. The
[public signature guide](https://payment-doc.lailaolab.com/v1/verify-webhook-signature)
does not include the verification method or signing key; request those from
Phajay support. The SDK provides payload types, but it does not verify sender
identity.

Select the type for the flow that produced the callback:

| Flow | Callback type | Completion or event field |
| --- | --- | --- |
| Production payment link | `PaymentLinkWebhookCallback` | Optional `Status`; compare with `PaymentLinkWebhookStatusCompleted`. |
| Bank QR | `PaymentQRWebhookCallback` | Optional `Status`; compare with `PaymentQRWebhookStatusCompleted`. |
| Hosted credit card | `CreditCardWebhookCallback` | Optional `Status` and `TransactionID`; inspect the verified payment result. |
| Subscription setup, debit, cancellation | `SubscriptionWebhookCallback` | `Status`; debit events may include `PaymentTransactionID`. |
| Sandbox payment link | `SandboxPaymentLinkWebhookCallback` | Optional `Status`, with a distinct sandbox field schema. |

The SDK represents optional callback values as pointers. The following handler
factory takes application functions for signature verification and durable
payment processing. Import `context`, `encoding/json`, `io`, and `net/http` in addition
to `phajay`. `verify` must implement Phajay's supplied verification method;
`process` must match the saved transaction and amount and apply the payment
once, including handling already processed events successfully:

```go
func QRWebhookHandler(
	verify func(http.Header, []byte) error,
	process func(context.Context, phajay.PaymentQRWebhookCallback) error,
) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			w.Header().Set("Allow", http.MethodPost)
			http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
			return
		}
		if verify == nil || process == nil {
			http.Error(w, "handler not configured", http.StatusInternalServerError)
			return
		}
		// Choose a body limit that fits your application's callback payloads.
		r.Body = http.MaxBytesReader(w, r.Body, 1<<20)
		rawBody, err := io.ReadAll(r.Body)
		if err != nil {
			http.Error(w, "invalid body", http.StatusBadRequest)
			return
		}
		if err := verify(r.Header, rawBody); err != nil {
			http.Error(w, "invalid signature", http.StatusUnauthorized)
			return
		}
		var callback phajay.PaymentQRWebhookCallback
		if err := json.Unmarshal(rawBody, &callback); err != nil {
			http.Error(w, "invalid body", http.StatusBadRequest)
			return
		}
		if err := process(r.Context(), callback); err != nil {
			http.Error(w, "processing failed", http.StatusInternalServerError)
			return
		}
		w.WriteHeader(http.StatusOK)
	}
}
```

Inside `process`, check `callback.Status != nil` before comparing its value.
Status can be absent; use a production status check when you need to reconcile
the transaction. Record or reconcile other callback outcomes before acknowledging
them. Plain value fields reflect the documented guaranteed subset, but JSON
decoding does not enforce their presence; validate identifiers and amounts in
your application.

Return HTTP 200 after accepting a verified callback. Link `RefNo` remains a
`*string` and accepts JSON numbers and strings. QR `RefNo` is a
`*FlexibleString`; use `string(*callback.RefNo)` after checking for nil. Both
preserve numeric reference text without rounding large integers.

Sandbox payment-link callbacks use a separate
`SandboxPaymentLinkWebhookCallback` type. Their field names differ from
production callbacks. The sandbox guide does not define a separate QR callback
schema; inspect the payload before choosing a callback type.

## Request and track a refund

Use a production client and the payment's transaction ID. Refund operations
require an approved KYC account. A successful request starts a refund workflow:

```go
response, err := client.RequestRefund(ctx, phajay.RefundRequest{
	TransactionID: transactionID,
})
if err != nil {
	return err
}
refundBillID := response.Data.ID
```

Persist the refund bill ID and use it for later checks. The identifier is
`Data.ID`, distinct from the original transaction ID and `Data.BillNo`:

```go
response, err := client.CheckRefundStatus(ctx, refundBillID)
if err != nil {
	return err
}
refundStatus := response.Data.Status
```

Interpret the returned status using the
[refund contract](https://payment-doc.lailaolab.com/v1/refund).
The SDK returns status strings and does not run a polling loop. `Refund.Amount`
accepts both numeric and numeric-string JSON representations.

## Errors and input rules

The client rejects invalid requests before making HTTP calls:

- Payment amounts must be finite and greater than zero.
- Descriptions must contain non-whitespace text.
- Subscription dates must use `YYYY-MM-DD`; recurrence intervals must be
  positive.
- Transaction IDs, refund bill IDs, and subscription authorization codes must
  be present.

For non-2xx gateway responses, use `errors.As` to inspect `APIError`:

```go
var apiErr *phajay.APIError
if errors.As(err, &apiErr) {
	fmt.Println(apiErr.StatusCode, apiErr.Operation, apiErr.Body)
}
```

`APIError.Body` contains up to 1 MiB of trimmed response text. Network errors
remain discoverable with `errors.Is` because the SDK wraps the underlying error.
The SDK does not retry payment creation or refund requests automatically.

A nil error means the HTTP response was 2xx and its JSON was decoded. Inspect
the response message and business state before advancing your order. A timeout
can occur after the gateway accepted a payment or refund; reconcile the saved
identifiers before deciding whether to submit again.

Detect request cancellation or a deadline through wrapping:

```go
if errors.Is(err, context.Canceled) {
	// The caller canceled the operation.
}
if errors.Is(err, context.DeadlineExceeded) {
	// Reconcile the payment before resubmitting a creation request.
}
```

## Troubleshooting

| Symptom | Check |
| --- | --- |
| Amount or description validation error | Use a finite positive amount and a description containing text. |
| Unknown bank | Pass one of the exported `Bank*` constants. |
| No documented sandbox endpoint | Use a supported sandbox flow or a production client for that operation. |
| Gateway rejects authentication | Check the key, selected environment, and any transport that changes headers. The SDK selects endpoint authentication itself. |
| Callback status is nil | Banks can omit it; correlate the transaction and reconcile through the production status API. |
| Transaction details are nil | Inspect `Message` and confirm the transaction ID and environment. |
| Repeated fulfillment | Make the saved payment transition and fulfillment record atomic, keyed by your order and gateway transaction. |
| Subscription QR appears as a URL | Display the image URL form returned by the gateway. |

## Further reading

- [Full SDK API reference](../README.md#api-reference)
- [Architecture](ARCHITECTURE.md)
- [Contributor guide](../CONTRIBUTING.md)
- [Payment link guide](https://payment-doc.lailaolab.com/v1/payment-link)
- [QR payment guide](https://payment-doc.lailaolab.com/v1/connect-payment-QR/generateQR)
- [Transaction status guide](https://payment-doc.lailaolab.com/v1/check-transaction-status)
- [Refund guide](https://payment-doc.lailaolab.com/v1/refund)
- [Credit card guide](https://payment-doc.lailaolab.com/v1/connect-credit-card)
- [Subscription guide](https://payment-doc.lailaolab.com/v1/connect-subscription/generateQR)
