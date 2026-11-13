# Changelog

All notable changes to this project will be documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.1.0/),
and this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

## [Unreleased]

## [0.2.0] - 2026-11-13

### Added
- `pairing.failed` / `pairing.succeeded` events: `EventPairingFailed`, `EventPairingSucceeded`, the `PairingFailed` payload (`Method`, `Reason`, `At`), `PairingSucceeded`, the `PairingFailureReason` constants and `PairingFailureReasons()`. An unknown `Reason` still decodes; treat it like `PairingError` (`Reason.Known()` tells them apart)
- `disconnected` event: `Disconnected` payload (`Reason`, always-present `WillReconnect`, `Resume`, `Code`, `Detail`, `ExpireSeconds`, `RetryAfterSeconds`), the `DisconnectReason` constants and `DisconnectReasons()`. An unknown `Reason` still decodes; branch on `WillReconnect` (`Reason.Known()` tells them apart)
- `VerifyWebhook(body, header, secret, VerifyOptions)`: verifies `x-wi-signature-v2` (HMAC of `<t>.<body>`) with replay protection (default tolerance 300 s); falls back to the legacy `x-wi-signature` only when no v2 header is sent
- `receipt` event: `EventReceipt`, `Receipt` and `ReceiptStatus`

### Changed
- `WebhookHandler` verifies with `VerifyWebhook`

## [0.1.1] - 2026-09-03

### Added
- Comprehensive unit tests covering all session methods, webhooks, and options via `httptest`
- Race detector enabled in continuous integration

### Fixed
- Updated Go directive to Go 1.27
- Corrected import paths and module declaration to `github.com/wiApi/go`

## [0.1.0] - 2026-09-03

### Added
- Initial release of the official wi-api Go SDK
- Zero runtime dependencies (stdlib only)
- Complete WhatsApp session management and messaging APIs
- HTTP webhook signature verification and standard `http.Handler` integration
