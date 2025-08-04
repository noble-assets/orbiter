# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this
repository.

## Common Commands

### Building

- `make build` - Build the simd binary (builds in simapp directory)
- `make all` - Full build pipeline: protobuf generation, formatting, linting, license check, unit
  tests, and build

### Testing

- `make test-unit` - Run unit tests for keeper, controllers, and types packages
- `make test-unit-viz` - Run unit tests with coverage visualization and HTML reports
- `make test-e2e` - Run end-to-end tests (requires Docker image build first)
- `go test -v ./keeper/...` - Run tests for a specific package
- `go test -v ./controllers/...` - Run controller tests
- `go test -v ./types/...` - Run type tests

### Code Quality

- `make format` - Run Go formatters using golangci-lint
- `make lint` - Run linter using golangci-lint configuration
- `make license` - Add license headers to Go files

### Protobuf

- `make proto-all` - Full protobuf pipeline (format, lint, generate)
- `make proto-gen` - Generate Go code from protobuf files using Docker
- `make proto-format` - Format protobuf files using buf
- `make proto-lint` - Lint protobuf files using buf

### Local Development

- `make local-image` - Build local Docker image using heighliner
- `./local.sh` - Local development script

## Architecture

### Core Components

The Orbiter module is a Cosmos SDK module that implements cross-chain functionality using a
component-based architecture:

- **Keeper** (`keeper/keeper.go`): Main module keeper managing four core components
- **ActionComponent**: Handles pre-execution actions (like fee payments)
- **OrbitComponent**: Manages cross-chain forwarding operations
- **DispatcherComponent**: Orchestrates payload processing and execution
- **AdapterComponent**: Interfaces with external protocols

### Key Directories

- `keeper/components/` - Core business logic components
- `controllers/` - Protocol-specific implementations (actions, orbits)
- `types/` - Type definitions, interfaces, and protobuf-generated code
- `types/interfaces/` - Interface definitions for components and controllers
- `entrypoint/` - IBC middleware integration
- `proto/` - Protobuf schema definitions
- `e2e/` - End-to-end integration tests
- `simapp/` - Simulation application for testing
- `testutil/` - Test utilities and mocks

### Supported Protocols

- **CCTP (Circle Cross-Chain Transfer Protocol)** - Production ready
- **IBC** - In development (TODO)
- **Hyperlane** - In development (TODO)

### Actions & Orbits

- **Actions**: Pre-execution operations (currently supports fee payments)
- **Orbits**: Cross-chain forwarding operations via bridge protocols
- Both use a controller pattern with routers for extensibility

### Payload Structure

The module processes JSON payloads containing:

- `pre_actions[]` - List of actions to execute before forwarding
- `orbit` - Forwarding operation specification with protocol-specific attributes

## Integration Notes

- IBC memo field integration for payload delivery
- Protobuf-based type system with interface registry
- Component-based architecture allows easy extension of protocols
- Comprehensive test coverage with unit and e2e tests

## Development

- Uses Go 1.24.4
- Cosmos SDK v0.50.13
- IBC-Go v8.3.2
- Circle CCTP integration
- Buf for protobuf tooling
- Docker required for protobuf generation and e2e tests

## Tests

This section describes how tests should be written for the Orbiter module.

Unit tests, should be written using the `testCases` pattern. The name of every test case, should be:

- `success - <DESCRIPTION>`: when the test is evaluating a successful case.
- `error - <DESCRIPTION>`: when the test is evaluating that an error is returned from the tested
  function.

When the test does not have any error to test, the prefix can be omitted.

To check that a function returns an error, the test case structure should have a field named
`expError` of type string, and the check against the error should be
`require.ErrContains(t, tC.expError, err)`.
