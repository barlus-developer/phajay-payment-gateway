# Changelog

All notable changes to this project are documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.1.0/),
and this project adheres to
[Semantic Versioning](https://semver.org/spec/v2.0.0.html).

## [Unreleased]

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

[Unreleased]: https://github.com/barlus-developer/phajay-payment-gateway/compare/v0.4.0-alpha1...HEAD
[0.4.0-alpha1]: https://github.com/barlus-developer/phajay-payment-gateway/compare/v0.3.0-alpha...v0.4.0-alpha1
[0.3.0-alpha]: https://github.com/barlus-developer/phajay-payment-gateway/compare/v0.2.0-alpha1...v0.3.0-alpha
[0.2.0-alpha1]: https://github.com/barlus-developer/phajay-payment-gateway/compare/v0.2.0-alpha...v0.2.0-alpha1
[0.2.0-alpha]: https://github.com/barlus-developer/phajay-payment-gateway/compare/v0.1.0-alpha...v0.2.0-alpha
[0.1.0-alpha]: https://github.com/barlus-developer/phajay-payment-gateway/releases/tag/v0.1.0-alpha
