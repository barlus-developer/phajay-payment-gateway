# Phajay Go SDK documentation

This SDK is a Go library imported by your application. It requires Go 1.23 or
later and uses only the standard library.

## For application developers

1. Read the [integration guide](INTEGRATION.md) to install the SDK, configure a
   client, and choose a payment flow.
2. Use the [API reference](../README.md#api-reference) for method signatures,
   field tables, bank constants, and callback types.
3. Check the [migration notes](../README.md#migration-notes) and
   [changelog](../CHANGELOG.md) before upgrading.

## For contributors

1. Follow [CONTRIBUTING.md](../CONTRIBUTING.md) for local setup, coding rules,
   test expectations, and the pull request process.
2. Read the [architecture guide](ARCHITECTURE.md) to find the source files and
   understand how requests and callbacks are processed.
3. Follow the repository rules in [AGENTS.md](../AGENTS.md).

## Documentation scope

These guides describe the implementation in this checkout. The
[official Phajay v1 documentation](https://payment-doc.lailaolab.com/v1)
defines the gateway contract; this repository implements an unofficial SDK.
The SDK API reference is also available on
[pkg.go.dev](https://pkg.go.dev/github.com/barlus-developer/phajay-payment-gateway),
where the displayed version may differ from this checkout.

Examples marked as fragments assume application variables are already in
scope. The integration guide's quick start is a complete Go program.
