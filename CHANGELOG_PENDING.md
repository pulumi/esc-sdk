### Improvements

- Add C# SDK with generated client, hand-written wrappers (EscClient, EscAuth, ValueMapper,
  EnvironmentDefinitionSerializer), and xUnit tests
  [#118](https://github.com/pulumi/esc-sdk/pull/118)

- Support proxy environment variables in Python SDK
  [#108](https://github.com/pulumi/esc-sdk/pull/108)

### Bug Fixes

- Fix Python default_client return type annotation
  [#109](https://github.com/pulumi/esc-sdk/pull/109)

- Replace pagination-dependent `listEnvironments` existence checks with direct `getEnvironment` calls in Go, TypeScript, Python, and C# integration tests to remove a flake when the shared test org grows past one page
  [#134](https://github.com/pulumi/esc-sdk/pull/134)

### Breaking changes
