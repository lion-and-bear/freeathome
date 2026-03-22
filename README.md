# freeathome

A client library for the BUSCH-JAEGER free@home local API implemented in Golang.

![GitHub go.mod Go version](https://img.shields.io/github/go-mod/go-version/lion-and-bear/freeathome)
![CI](https://img.shields.io/github/actions/workflow/status/lion-and-bear/freeathome/ci.yaml?style=flat-square)
[![codecov](https://codecov.io/gh/lion-and-bear/freeathome/branch/main/graph/badge.svg?token=UJQVXZ5PPM)](https://codecov.io/gh/lion-and-bear/freeathome)
[![Quality Gate Status](https://sonarcloud.io/api/project_badges/measure?project=lion-and-bear_freeathome&metric=alert_status)](https://sonarcloud.io/summary/new_code?id=lion-and-bear_freeathome)
![License](https://img.shields.io/github/license/lion-and-bear/freeathome?style=flat-square)

## Installation

### Library Usage

To use this library in your own Go project, make sure your project is using Go modules. Then, run the following command to add the dependency:

```sh
go get github.com/lion-and-bear/freeathome@latest
```

You can then import the package in your code:

```go
import "github.com/lion-and-bear/freeathome/v2"
```

This will give you access to the public API client and related utilities for interacting with a local free\@home SysAP.

### CLI Tool

The project includes a comprehensive command-line interface (CLI) tool for interacting with free@home systems. The CLI provides a unified interface for all operations including configuration, data retrieval, data modification, and real-time monitoring.

#### Building the CLI

```sh
# Build the CLI tool with proper version and commit information
make cli-build

# Or run it directly without building
make cli-run-local
```

#### Docker Support

The CLI is also available as a Docker image:

```sh
# Build Docker image
make cli-build-docker

# Build multi-architecture Docker image (linux/amd64, linux/arm64)
make cli-build-docker-multiarch

# Run with Docker
docker run --rm ghcr.io/lion-and-bear/freeathome-cli:latest --help
```

#### CLI Commands

##### Configuration

```sh
# Configure your system access point (interactive)
./fh configure

# Configure with command line flags
./fh configure --hostname 192.168.1.100 --username admin --password mypass

# Configure with environment variables
export FREEATHOME_HOSTNAME=192.168.1.100
export FREEATHOME_USERNAME=admin
export FREEATHOME_PASSWORD=mypass
./fh configure

# Show current configuration
./fh configure show
```

##### Data Retrieval

```sh
# Get device list
./fh get devicelist

# Get configuration
./fh get configuration

# Get specific device by serial
./fh get device [serial]

# Get specific datapoint
./fh get datapoint [serial] [channel] [datapoint]

# Output options
./fh get devicelist --output json --prettify
./fh get devicelist --output text
```

##### Data Modification

```sh
# Set datapoint value
./fh set datapoint [serial] [channel] [datapoint] [value]
```

##### Real-time Monitoring

```sh
# Monitor real-time events
./fh monitor

# Monitor with custom settings
./fh monitor --timeout 60 --max-reconnection-attempts 5 --exponential-backoff
```

##### Global Options

All commands support these global options:

```sh
# TLS configuration
--tls                    # Enable TLS (default: true)
--skip-tls-verify       # Skip TLS certificate verification

# Logging
--log-level             # Set log level (debug, info, warn, error)

# Output format (for get commands)
--output                # Output format (json, text)
--prettify              # Prettify JSON output with indentation
```

For more information about available commands and options, run `./fh --help` or `./fh [command] --help`.

## Features

The project initially covers the scope of the [TypeScript free@home API Client](https://github.com/pgerke/freeathome-local-api-client).

- Connect to your B+J System Access Point 2.0 and control it using the local API.
- 100% covered by automated unit tests
- Websocket communication with keepalive
- Get configuration
- Get device list
- Get device
- Create virtual device
- Get and set datapoints
- Trigger proxy device
- Set proxy device value
- Default and custom loggers!

### CLI Tool Features

The CLI tool provides a comprehensive interface for all free@home operations:

- **Configuration Management**: Interactive and non-interactive configuration with support for YAML files and environment variables
- **Data Retrieval**: Get device lists, configurations, individual devices, and datapoints with flexible output formats
- **Data Modification**: Set datapoint values with validation
- **Real-time Monitoring**: WebSocket-based monitoring with configurable reconnection strategies
- **Docker Support**: Multi-architecture Docker images for easy deployment
- **Flexible Output**: JSON and text output formats with prettify options
- **TLS Configuration**: Configurable TLS settings with certificate verification options
- **Logging**: Configurable log levels for debugging and monitoring

## Usage Requirements

- A free@home System Access Point 2.0 running firmware > v3.0
- Local API has to be enabled for the user account to be used

## Documentation

The API documentation is available at https://pkg.go.dev/github.com/lion-and-bear/freeathome.

## I found a bug, what do I do?

I'm happy to hear any feedback regarding the library or it's implementation, be it critizism, praise or rants. Please create a [GitHub issue](https://github.com/lion-and-bear/freeathome/issues) or drop me an [email](mailto:info@philipgerke.com) if you would like to contact me.

I would especially appreciate, if you could report any issues you encounter while using the library. Issues I know about, I can probably fix.

If you want to submit a bug report, please check if the issue you have has already been reported. If you want to contribute additional information to the issue, please add it to the existing issue instead of creating another one. Duplicate issues will take time from bugfixing and thus delay a fix.

While creating a bug report, please make it easy for me to fix it by giving us all the details you have about the issue. Always include the version of the library and a short concise description of the issue. Besides that, there are a few other pieces of information that help tracking down bugs:

- The system environment in which the issue occurred
- Some steps to reproduce the issue, e.g. a code snippet
- The expected behaviour and how the failed failed to meet that expectation
- Anything else you think I might need

## I have a feature request, what do I do?

Please create a [GitHub issue](https://github.com/lion-and-bear/freeathome/issues) or drop me an [email](mailto:info@philipgerke.com)!

## Non-Affiliation Disclaimer

This library is not endorsed by, directly affiliated with, maintained, authorized, or sponsored by Busch-Jaeger Elektro GmbH or ABB Asea Brown Boveri Ltd or . All product and company names are the registered trademarks of their original owners. The use of any trade name or trademark is for identification and reference purposes only and does not imply any association with the trademark holder of their product brand.

## License

The project is subject to the MIT license unless otherwise noted. A copy can be found in the root directory of the project [LICENSE](./LICENSE).

<hr>

Made with ❤️ by [Philip Gerke](https://github.com/pgerke)
