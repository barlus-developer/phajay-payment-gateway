# Contributing

Thanks for your interest in contributing to the Phajay Payment Gateway SDK for
Go. This document describes how to report issues and propose changes.

Read the [architecture guide](docs/ARCHITECTURE.md) to understand the request
helper, endpoint files, and callback models. The
[developer integration guide](docs/INTEGRATION.md) shows how applications use
the SDK. Repository conventions are recorded in [AGENTS.md](AGENTS.md).

## Getting started

- **Go 1.23** or later is required.
- The SDK has no third-party dependencies, so no `go mod tidy` step beyond the
  standard module setup is needed.
- Git and a Go installation that supports the race detector are required for
  the test command. Tests use local `httptest` servers; no Phajay credentials or
  live gateway access are required.

```bash
git clone https://github.com/barlus-developer/phajay-payment-gateway.git
cd phajay-payment-gateway
go test ./... -race -cover
```

This repository builds a library. Check compilation and preview package docs
locally with:

```bash
go build ./...
go doc .
```

## Find the code to change

| Change | Start here |
| --- | --- |
| Client configuration | `phajay.go`, `phajay_test.go` |
| Shared HTTP behavior or errors | `request.go`, `phajay_test.go`, `endpoints_test.go` |
| Payment links or bank QR | `payment_link.go`, `payment_qr.go`, and their matching tests |
| Status, refunds, cards, subscriptions | `transaction.go`, `refund.go`, `payment_credit_card.go`, `subscription.go`, `endpoints_test.go` |
| Webhook verification policy | `webhook.go`, `webhook_test.go` |
| Callback fields or decoding | The corresponding `*_webhook.go`, `callbacks_test.go`, and `*_webhook_test.go` |
| Sandbox routing | `payment_link.go`, `payment_qr.go`, `request.go`, `sandbox_test.go` |
| User documentation | `README.md`, `docs/INTEGRATION.md`, `docs/ARCHITECTURE.md`, `docs/README.md` |

See the [architecture source map](docs/ARCHITECTURE.md#source-map) for individual
file responsibilities.

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
   Update the integration examples and architecture guide when behavior or
   implementation boundaries change.
8. Open a pull request against `main` with a clear description of the problem
   and solution.

## Code style

- Follow standard Go conventions (`gofmt`, `go vet`).
- Keep exported types and functions documented with Go doc comments.
- Document exported constants and methods too; describe constraints, errors,
  and optional fields that a caller needs to handle.
- Write table-driven tests for new behaviour.
- Keep the package flat and dependency-free; use `New(Config, opts...)` and
  preserve functional options after Config.
- Wrap underlying errors with `%w` and prefix SDK errors with `phajay: `.
- Validate inputs before sending requests, including finite positive amounts.

## Implement an endpoint

1. Confirm the route, HTTP method, authentication, field names, response shape,
   and environment support against the official documentation. Include the
   source link in the pull request so reviewers can assess the contract.
2. Add request and response types in the root file for the feature. Keep JSON
   tags consistent with the wire format, including documented unusual spellings.
   Use pointers for optional callback fields and preserve timestamp text.
3. Add a method with `ctx context.Context` as its first parameter and a typed
   response plus `error`. Return a zero response on local validation failure.
4. Validate inputs before calling `p.do`. For production-only operations, call
   `requireProduction`; use `pathID` for identifiers placed in lookup paths.
5. Call `p.do` with the documented method, path, authentication enum, operation
   name, and response pointer. Keep transport logic in the shared helper.
6. Add or update tests for observable behavior and update the public docs. A new
   bank also needs its `Bank` constant, `bankPath` mapping, and explicit sandbox
   behavior.

Use `New(Config{APIKey: "test-key", BaseURL: server.URL})` or
`WithBaseURL(server.URL)` to direct the client to an `httptest` server.
Use `WithHTTPClient` for transport-specific behavior. Avoid real payment or
refund requests in the unit suite.

## Test expectations

Use table-driven cases and inline payloads. For an endpoint change, cover the
behavior that can break an integration:

| Concern | What the test should establish |
| --- | --- |
| Routing and auth | Exact literal path, method, authentication header, and content type. |
| Request body | Field names, values, and omission of fields tagged `omitempty`. |
| Success decoding | Typed response fields, nullable data, and documented representation variants. |
| Validation | Invalid inputs return an error and send zero HTTP requests. |
| Environment | Supported sandbox path or local rejection for an unsupported operation. |
| Failure | Non-2xx responses expose `APIError`; underlying failures remain inspectable. |
| Cancellation | Caller cancellation reaches the request and remains discoverable through wrapping. |
| Callbacks | Missing/null optional fields remain nil; reference numbers retain precision. |

Always run tests with both the race detector and coverage:

```bash
gofmt -w .
go vet ./...
go test ./... -race -cover
```

The existing GitHub Actions workflow runs the same test command on Go 1.23 for
pull requests and pushes to `main` or `master`. Formatting and vet are local
contributor checks. Keep changes compatible with the minimum Go version in
`go.mod`.

## Documentation changes

Use the [documentation index](docs/README.md) to choose the right guide. Keep
method signatures and field tables in the README API reference, application
examples in the integration guide, and implementation explanations in the
architecture guide. Use relative links for repository files.

Label fragments with their assumed variables and imports. Complete programs
must include package and import declarations. Use placeholder credentials and
application-generated identifiers. Webhook examples must require verification
before processing callback data. Use `DecodeWebhook` with original bytes and
headers to enforce the client policy. Verification currently fails closed:
missing/blank signing keys return `ErrWebhookSignatureKeyMissing`, and configured
keys return `ErrWebhookSignatureVerificationUnavailable`. Explicit bypass is
only for development or tests and never authenticates an event.

For public API changes, update the README field tables, this contributor guide,
and `CHANGELOG.md` under `Unreleased`. Include migration examples and a breaking
change marker when an exported field type or other contract changes. For a
documentation-only change, review links and examples against the source; the
SDK behavior is unchanged.

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
signing scheme; cryptographic verification requires the private guide supplied
by Phajay support. Preserve the fail-closed policy until that guide is available.
Do not infer a signing key from `APIKey`, guess the signature header, or enable
bypass when configuration is missing. Once the specification is available,
implement its exact payload and encoding, reject missing/malformed/invalid
signatures, use constant-time comparison where applicable, and add official
valid-signature vectors plus tampering tests. Development bypass tests must
exercise every callback type without claiming sender authentication.

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
