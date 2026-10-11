// Package phajay provides a dependency-free Go client for the Phajay payment
// gateway's payment links, bank QR codes, transaction status, refunds, hosted
// credit card payments, and BCEL subscriptions.
//
// Construct a client with New(Config{APIKey: "..."}) and optional functional
// options. Use a portal test key and WithSandbox for the documented link and
// bank QR test routes. Other operations have no documented sandbox endpoints.
//
// DecodeWebhook requires verification by default and fails closed until the
// private signing specification from Phajay support is available. Explicit
// bypass is for development and tests only. Webhook types and direct JSON
// decoding do not authenticate the sender. Official API documentation is at
// https://payment-doc.lailaolab.com/v1.
package phajay
