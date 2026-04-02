# Changelog

All notable changes to this project will be documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.1.0/),
and this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

## [Unreleased]

### Added

- [#65] CLI: `monitor --raw` writes each WebSocket text frame unchanged to stdout (newline-terminated); user hints and slog output stay on stderr.

## [v2.5.0] - 22.03.2026

### Changed

- Bump dependencies
- **Go module path**: The module is now `github.com/lion-and-bear/freeathome/v2` (repository moved to the `lion-and-bear` GitHub organization). Update `import` paths and `go get` accordingly.

## [v2.4.0] - 22.02.2026

### Changed

- Bump dependencies
- **Go Version**: Updated to Go 1.26

## [v2.3.0] - 10.01.2026

### Changed

- Bump dependencies

## [v2.2.0] - 28.11.2025

### Changed

- Bump dependencies

## [v2.1.0] - 26.09.2025

### Fixed

- [#41] Fixed go module major version path

### Changed

- Bump dependencies

## [v2.0.0] - 30.08.2025

### Added

- [#28] **Comprehensive CLI Tool**: Implemented a full-featured command-line interface for all free@home operations
  - Configuration management with interactive and non-interactive modes
  - Support for YAML configuration files and environment variables
  - Data retrieval commands (devicelist, configuration, device, datapoint)
  - Data modification commands (set datapoint)
  - Real-time monitoring with configurable reconnection strategies
  - Flexible output formats (JSON/text) with prettify options
  - TLS configuration options with certificate verification controls
  - Configurable logging levels for debugging and monitoring
- [#33] **Docker Support**: Added Docker image building and multi-architecture support
  - Multi-arch Docker images (linux/amd64, linux/arm64)
  - Automated Docker image builds in CI/CD pipeline
  - Security-focused Docker images with non-root user
- **Enhanced Build System**: Improved Makefile with CLI-specific build targets
  - `make cli-build` for building the CLI binary
  - `make cli-run-local` for running without building
  - `make cli-build-docker` for Docker image building
  - `make cli-build-docker-multiarch` for multi-arch images

### Changed

- **Monitor Application Integration**: Refactored the standalone monitor application to be fully integrated into the CLI tool
- **WebSocket Improvements**: Enhanced WebSocket communication with better error handling and reconnection logic
- **Code Quality**: Improved separation of concerns and testability throughout the CLI implementation
- **Error Handling**: Enhanced error handling and validation across all CLI commands

### Technical Improvements

- **Go Version**: Updated to Go 1.25
- **Dependencies**: Bumped all Go dependencies to latest compatible versions
- **Test Coverage**: Significantly improved test coverage for CLI functionality
- **Integration Tests**: Enhanced integration tests with better timeout handling

## [v1.0.0] - 12.06.2025

### Added

- [#7] Implemented initial API scope:
  - The scope of the [TypeScript free@home API Client](https://github.com/pgerke/freeathome-local-api-client) is covered
  - System Access Point configuration
  - Logger support
  - Websocket communication with keepalive
  - Get configuration
  - Get device list
  - Get device
  - Create virtual device
  - Get and set datapoints
  - Trigger proxy device
  - Set proxy device value
- [#16] Implemented monitor application covering the scope of the [JavaScript free@home Monitor](https://github.com/pgerke/freeathome-monitor)

[Unreleased]: https://github.com/lion-and-bear/freeathome/compare/2.5.0...HEAD
[v2.5.0]: https://github.com/lion-and-bear/freeathome/releases/tag/2.5.0
[v2.4.0]: https://github.com/lion-and-bear/freeathome/releases/tag/2.4.0
[v2.3.0]: https://github.com/lion-and-bear/freeathome/releases/tag/2.3.0
[v2.2.0]: https://github.com/lion-and-bear/freeathome/releases/tag/2.2.0
[v2.1.0]: https://github.com/lion-and-bear/freeathome/releases/tag/2.1.0
[v2.0.0]: https://github.com/lion-and-bear/freeathome/releases/tag/2.0.0
[v1.0.0]: https://github.com/lion-and-bear/freeathome/releases/tag/1.0.0
[#65]: https://github.com/lion-and-bear/freeathome/issues/65
[#41]: https://github.com/lion-and-bear/freeathome/issues/41
[#28]: https://github.com/lion-and-bear/freeathome/issues/28
[#33]: https://github.com/lion-and-bear/freeathome/issues/33
[#7]: https://github.com/lion-and-bear/freeathome/issues/7
[#16]: https://github.com/lion-and-bear/freeathome/issues/16
