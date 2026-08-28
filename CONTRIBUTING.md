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
go test ./...
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

6. Open a pull request against `main` with a clear description of the problem
   and solution.

## Code style

- Follow standard Go conventions (`gofmt`, `go vet`).
- Keep exported types and functions documented with Go doc comments.
- Write table-driven tests for new behaviour.

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
