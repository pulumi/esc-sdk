# CHANGELOG

## 0.14.0

### Bug Fixes

- Fix typo (occured -> occurred) in Python SDK workspace error messages
  [#131](https://github.com/pulumi/esc-sdk/pull/131)

- Replace pagination-dependent `listEnvironments` existence checks with direct `getEnvironment` calls in Go, TypeScript, Python, and C# integration tests to remove a flake when the shared test org grows past one page
  [#134](https://github.com/pulumi/esc-sdk/pull/134)

### Breaking changes

- Default credentials are now sourced exclusively from the `PULUMI_ACCESS_TOKEN` and
  `PULUMI_BACKEND_URL` environment variables. The SDKs no longer fall back to reusing the
  Pulumi/ESC CLI login on disk (`~/.pulumi/credentials.json`). This removes the Go SDK's
  dependency on `github.com/pulumi/esc` (and its transitive `pulumi/pulumi` dependency) and
  removes the credential-file readers from the Python, TypeScript, and C# SDKs. Callers that
  relied on automatic CLI-login pickup must now set `PULUMI_ACCESS_TOKEN` (and optionally
  `PULUMI_BACKEND_URL`) or pass credentials explicitly.
  [#137](https://github.com/pulumi/esc-sdk/pull/137)

## 0.13.0

### Improvements

- Add C# SDK with generated client, hand-written wrappers (EscClient, EscAuth, ValueMapper,
  EnvironmentDefinitionSerializer), and xUnit tests
  [#118](https://github.com/pulumi/esc-sdk/pull/118)

## 0.12.4

### Improvements

- Support proxy environment variables in Python SDK
  [#108](https://github.com/pulumi/esc-sdk/pull/108)

### Bug Fixes

- Fix Python default_client return type annotation
  [#109](https://github.com/pulumi/esc-sdk/pull/109)

## 0.12.3

## 0.12.2

### Bug Fixes

- Drop urllib constraint from Python SDK release
  [#99](https://github.com/pulumi/esc-sdk/pull/99)

## 0.12.1

### Bug Fixes

- Fixing bad import in TS SDK release
  [#88](https://github.com/pulumi/esc-sdk/pull/88)

## 0.12.0

### Improvements

- Adding default authorization methods for parity with CLI
  - All SDK now automatically read in configuration environment variables
  - Go SDK also automatically picks up configuration from CLI Pulumi accounts
    [#76](https://github.com/pulumi/esc-sdk/pull/76)
- Adds support for reading credentials from disk to Python SDK
  [#81](https://github.com/pulumi/esc-sdk/pull/81)
- Adds support for reading credentials from disk to Typescript SDK
  [#86](https://github.com/pulumi/esc-sdk/pull/86)

## 0.11.0

### Improvements

- Add environment clone support
  [#45](https://github.com/pulumi/esc-sdk/pull/45)

### Bug Fixes

- Fix panic when reading invalid environment property
  [#60](https://github.com/pulumi/esc-sdk/pull/60)
