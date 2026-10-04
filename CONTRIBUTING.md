# Contributing

Thanks for your interest in contributing to the Phajay Payment Gateway SDK for
Go. This document describes how to report issues and propose changes.

## Getting started

- **Go 1.23** or later is required.
- The SDK has no third-party dependencies, so no `go mod tidy` step beyond the
  standard module setup is needed.

```bash
git clone https://github.com/barlus-developer/phajay-payment-gateway.git
cd phajay-payment-gateway
go test ./... -race -cover
```

## Reporting issues

Before opening an issue, check the existing
[issues](https://github.com/barlus-developer/phajay-payment-gateway/issues) to
avoid duplicates. A good issue includes:

- The SDK and Go version you are using.
- A minimal, reproducible example.
- The expected and actual behaviour.
- Any relevant request/response payloads (with API keys and sensitive data
  redacted).

## Proposing changes

1. Fork the repository and create a feature branch from `main`.
2. Make your changes, keeping them focused on a single concern.
3. Add or update tests to cover the change.
4. Run the full test suite:

   ```bash
   go test ./... -race -cover
   ```

5. Run `gofmt` on the files you changed:

   ```bash
   gofmt -w .
   ```

6. Run `go vet ./...`.
7. Update `README.md`, `CHANGELOG.md`, and this guide when the public API changes.
8. Open a pull request against `main` with a clear description of the problem
   and solution.

## Code style

- Follow standard Go conventions (`gofmt`, `go vet`).
- Keep exported types and functions documented with Go doc comments.
- Write table-driven tests for new behaviour.
- Keep the package flat and dependency-free; preserve the functional-options API.
- Wrap underlying errors with `%w` and prefix SDK errors with `phajay: `.
- Validate inputs before sending requests, including finite positive amounts.

## Gateway API changes

Use the [official v1 documentation](https://payment-doc.lailaolab.com/v1) as the
source for routes, HTTP methods, authentication, field names, and nullability.
Add `httptest` coverage for request bodies, headers, response decoding, failure
responses, cancellation, and validation that sends no HTTP requests. Use literal
documented paths in expectations so a route typo cannot also change the test.

Payment links, credit card links, and transaction status use Basic auth with
base64 of the raw key. QR, refund, and subscription routes use `secretKey`.
Use the shared request helper to keep response closing and error wrapping
consistent. Payment and subscription creation methods must not retry by default.

Sandbox has different paths on the same gateway host. Add a sandbox route only
when documented; unsupported sandbox methods must fail before HTTP instead of
falling through to production. M MoneyX, transaction status, refunds, credit
cards, and subscriptions currently have no documented test routes.

Preserve optional callback fields as pointers. The documented QR guarantees are
`paymentMethod`, `transactionId`, and `txnAmount`; `billNumber` and `status` may
be absent. Link `refNo` accepts numbers and strings without changing its public
`*string` type. Card payloads can contain null nested fields, and subscription
`paymentTransactionId` is present only on debit events. Mark type changes as
breaking and provide migration examples in README and CHANGELOG.

The [webhook verification page](https://payment-doc.lailaolab.com/v1/verify-webhook-signature)
does not publish the signing algorithm or header format. Do not invent a
verifier; implementing one requires the guide and key supplied by Phajay support.

## Commit messages

This repository follows the
[Conventional Commits](https://www.conventionalcommits.org/) style:

- `feat:` for new features.
- `fix:` for bug fixes.
- `docs:` for documentation changes.
- `refactor:` for code changes that do not alter behaviour.
- `ci:` for CI/CD changes.
- Use `!` (e.g. `refactor!:`) or a `BREAKING CHANGE:` footer for
  backwards-incompatible changes.

## Licence

By contributing, you agree that your contributions are licensed under the
[MIT License](LICENSE).
