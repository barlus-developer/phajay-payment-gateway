# Changelog

All notable changes to this project are documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.1.0/),
and this project adheres to
[Semantic Versioning](https://semver.org/spec/v2.0.0.html).

## [Unreleased]

## [0.5.1] - 2026-10-11

### Added

- `Config` with API key, optional base URL, separate webhook signing key, and
  explicit development/testing signature bypass (default false).
- Reusable `VerifyWebhookSignature` and `DecodeWebhook` methods, with
  inspectable missing-key and verification-unavailable errors.
- Tests covering configuration, mandatory verification, missing/blank signing
  keys, rejected candidate signatures, sandbox policy, bypass decoding for all
  callback types, and JSON error wrapping.

### Changed

- **Breaking:** `New` accepts `Config` instead of a key string. Migrate
  `New("key", opts...)` to `New(Config{APIKey: "key"}, opts...)`.
  Existing functional options are preserved and applied after Config.
- Client-managed webhook decoding requires verification by default and leaves
  callback data untouched on verification failures. Missing signing configuration
  never enables bypass; sandbox does not bypass verification either.

### Security limitation

- Phajay's public guide withholds the signing algorithm, payload, encoding, and
  header format. Verification fails closed for every request until the private
  specification is available. Valid-signature tests and cryptographic comparison
  are deferred; no signing scheme has been invented. Explicit bypass permits
  unauthenticated decoding only for development and tests.

## [0.5.0-alpha1] - 2026-10-11

### Added

- Developer integration examples, contributor implementation guidance, and an
  architecture guide with request and payment flow diagrams, linked from a
  documentation index.
- Optional unique `OrderNo` on `PaymentQRRequest`.
- `CheckTransactionStatus` with nullable transaction details and status constants.
- `RequestRefund` and `CheckRefundStatus`, accepting numeric and string refund amounts.
- `CreateCreditCardPaymentLink` and typed card webhook payloads.
- BCEL subscriptions via `GenerateSubscriptionQR` and `CancelSubscription`,
  with setup, debit, and cancellation webhook types and status constants.
- `WithSandbox` for documented payment link and JDB/LDB/IB/BCEL/STB QR test paths,
  and a separate `SandboxPaymentLinkWebhookCallback` payload.
- Inspectable `APIError` for non-2xx HTTP responses.
- Go doc comments and tests for new endpoints, validation, routing, authentication,
  optional callback fields, cancellation, and error handling.

### Changed

- **Breaking:** `PaymentQRWebhookCallback.BillNumber` and `.Status` are now
  `*string`, matching the current documentation's optional fields. Check for nil
  before dereferencing; see README migration notes.
- Custom HTTP clients are copied so timeout options no longer mutate the
  supplied client; a nil HTTP client keeps the default.
- Official documentation links now point to `https://payment-doc.lailaolab.com/v1`.

### Fixed

- Payment link callbacks accept `refNo` as a string or number while retaining
  the existing `*string` field type and preserving large integer precision.
- Reject non-finite amounts and whitespace-only descriptions before HTTP calls.
- Correct sandbox configuration to use test paths on the production gateway host.
- Document webhook signature verification as requiring Phajay's private guide
  and signing key; no undocumented verifier is provided.

## [0.4.0-alpha1] - 2026-08-28

### Added

- QR payment webhook callback type (`PaymentQRWebhookCallback`) with
  `FlexibleString` for normalising `refNo` across string and number forms.

## [0.3.0-alpha] - 2026-08-28

### Added

- Bank QR code generation via `GenerateQR` for JDB, LDB, IB, BCEL, STB, and
  M MoneyX.

## [0.2.0-alpha1] - 2026-08-28

### Changed

- Renamed the webhook callback type to `PaymentLinkWebhookCallback`
  (previously `WebhookCallback`). **Breaking change.**

## [0.2.0-alpha] - 2026-08-28

### Added

- Payment link webhook callback type.
- Unit test CI workflow.

## [0.1.0-alpha] - 2026-08-28

### Added

- Payment link creation via `CreatePaymentLink`.

[Unreleased]: https://github.com/barlus-developer/phajay-payment-gateway/compare/v0.5.1...HEAD
[0.5.1]: https://github.com/barlus-developer/phajay-payment-gateway/compare/v0.5.0-alpha1...v0.5.1
[0.5.0-alpha1]: https://github.com/barlus-developer/phajay-payment-gateway/compare/v0.4.0-alpha1...v0.5.0-alpha1
[0.4.0-alpha1]: https://github.com/barlus-developer/phajay-payment-gateway/compare/v0.3.0-alpha...v0.4.0-alpha1
[0.3.0-alpha]: https://github.com/barlus-developer/phajay-payment-gateway/compare/v0.2.0-alpha1...v0.3.0-alpha
[0.2.0-alpha1]: https://github.com/barlus-developer/phajay-payment-gateway/compare/v0.2.0-alpha...v0.2.0-alpha1
[0.2.0-alpha]: https://github.com/barlus-developer/phajay-payment-gateway/compare/v0.1.0-alpha...v0.2.0-alpha
[0.1.0-alpha]: https://github.com/barlus-developer/phajay-payment-gateway/releases/tag/v0.1.0-alpha
