# AGENTS.md

Guidance for AI agents and contributors working in this repository.

## Project

An unofficial Go SDK for the [Phajay](https://payment-gateway.phajay.co)
payment gateway. Module path: `github.com/barlus-developer/phajay-payment-gateway`,
package name `phajay`. Go 1.23 or later, no third-party dependencies.

## Commands

- Run tests (always with the race detector and coverage):

  ```bash
  go test ./... -race -cover
  ```

- Format code:

  ```bash
  gofmt -w .
  ```

- Vet:

  ```bash
  go vet ./...
  ```

There is no linter configuration beyond the defaults.

## Conventions

- **Package layout** — flat; all source files live in the repository root.
- **Error messages** — prefix with `phajay: ` and wrap underlying errors with
  `%w` (e.g. `fmt.Errorf("phajay: send request: %w", err)`).
- **Client options** — use the functional-options pattern
  (`Option func(*Phajay)`, `With*` functions applied in `New`).
- **Validation** — fail before sending any HTTP request (amount/description
  checks in `CreatePaymentLink` and `GenerateQR`).
- **Auth** — payment links use HTTP Basic auth (base64 API key); QR generation
  uses the `secretKey` header.
- **Webhook fields** — guaranteed fields are plain values; optional fields are
  pointers (`*string`, `*FlexibleString`) that stay `nil` when absent.
- **Doc comments** — exported types, functions, and constants require Go doc
  comments.
- **Tests** — table-driven, no external fixtures; use `httptest` for HTTP
  clients.

## Commit messages

Follow [Conventional Commits](https://www.conventionalcommits.org/): `feat:`,
`fix:`, `docs:`, `refactor:`, `ci:`. Mark breaking changes with `!`
(e.g. `refactor!:`) or a `BREAKING CHANGE:` footer.

## Documentation

Update `README.md` (API reference and field tables), `CHANGELOG.md`, and
`CONTRIBUTING.md` when public API changes.
