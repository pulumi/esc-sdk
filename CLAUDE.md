# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Overview

This is the Pulumi ESC SDK repository, which provides programmatic access to Pulumi ESC (Environments, Secrets, and Configurations) across multiple languages: Go, TypeScript/JavaScript, and Python. The SDKs are auto-generated from OpenAPI specs and include high-level wrapper clients.

## Architecture

The repository follows a multi-language SDK structure:

- **`/sdk/`** - Root directory for all SDK implementations
  - **`swagger.yaml`** - OpenAPI specification that drives SDK generation. This has been hand edited to provide a good code generation.  This is a subset of the overall Pulumi Cloud API.
  - **`templates/`** - Mustache templates for OpenAPI Generator (go/, python/, typescript/)
  - **`go/`** - Go SDK with auto-generated models and API clients
  - **`python/`** - Python SDK with Poetry configuration and high-level wrapper (`esc_client.py`)
  - **`typescript/`** - TypeScript/Node.js SDK with raw API clients and workspace functionality
  - **`test/`** - Test credentials and configuration for integration tests

Each language SDK contains:
- Auto-generated API clients from OpenAPI specs
- High-level wrapper clients providing ergonomic APIs. These are written by hand on top of the generated raw sdk.
- Integration tests requiring Pulumi credentials

## Common Development Commands

### Prerequisites
- OpenAPI Generator CLI
- Go, Node.js, Python depending on SDK you're working on
- For tests: `PULUMI_ACCESS_TOKEN` and `PULUMI_ORG` environment variables

### Building
```bash
# Generate all SDK clients from OpenAPI spec
make generate_sdks

# Build individual SDKs
make build_go
make build_typescript  
make build_python

# Default build (Go only)
make
```

### Testing
```bash
# Test individual SDKs (requires Pulumi credentials)
make test_go
make test_typescript
make test_python

# Go testing with coverage
make test_go_cover
```

### Linting
```bash
# Lint all languages
make lint

# Individual linters
make lint-golang     # golangci-lint in sdk/
make lint-python     # flake8 on esc_client.py and tests
make lint-copyright  # pulumictl copyright check
```

### Code Generation
The SDK clients are generated from `sdk/swagger.yaml` using OpenAPI Generator:
- Go: `make generate_go_client_sdk`
- TypeScript: `make generate_ts_client_sdk` 
- Python: `make generate_python_client_sdk`

## Key Files

- **`Makefile`** - Primary build automation
- **`sdk/swagger.yaml`** - OpenAPI spec (source of truth for API)
- **`sdk/typescript/esc/workspace.ts`** - High-level TypeScript client
- **`sdk/python/pulumi_esc_sdk/esc_client.py`** - High-level Python client
- **`sdk/go/`** - Contains both generated models and custom extensions (`api_esc_extensions.go`)

## Development Notes

- The SDKs provide both raw generated API clients and higher-level wrapper clients
- All changes to API surface should be made via `swagger.yaml` and then regenerated
- Integration tests require valid Pulumi credentials and organization access
- Python SDK uses Poetry for dependency management
- TypeScript SDK uses standard npm/package.json workflow
- Go SDK is designed as a submodule within the repository structure