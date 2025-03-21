CHANGELOG
=========

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
