# Changelog

All notable changes to this project will be documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.1.0/),
and this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

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
