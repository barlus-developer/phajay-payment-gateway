// Package phajay provides a dependency-free Go client for the Phajay payment
// gateway's payment links, bank QR codes, transaction status, refunds, hosted
// credit card payments, and BCEL subscriptions.
//
// Construct a client with New and configure it with functional options. Use a
// portal test key and WithSandbox for the documented link and bank QR test
// routes. Other operations have no documented sandbox endpoints.
//
// Webhook types decode payloads; signature verification requires the private
// guide and signing key from Phajay support. Official API documentation is at
// https://payment-doc.lailaolab.com/v1.
package phajay
