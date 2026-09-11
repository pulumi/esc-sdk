### Bug Fixes

- Fix typo (occured -> occurred) in Python SDK workspace error messages
  [#131](https://github.com/pulumi/esc-sdk/pull/131)

- Replace pagination-dependent `listEnvironments` existence checks with direct `getEnvironment` calls in Go, TypeScript, Python, and C# integration tests to remove a flake when the shared test org grows past one page
  [#134](https://github.com/pulumi/esc-sdk/pull/134)

### Breaking changes

- The Go SDK now calls Pulumi Cloud through `github.com/pulumi/pulumi-cloud-sdk/go` instead of a client
  generated from `sdk/swagger.yaml`, and requires Go 1.26.6. The `EscClient` methods keep their names, with
  these changes to the surrounding API:
  - Response types are now the `apitype` types of `pulumi-cloud-sdk`, re-exported under the old names
    (`Value`, `Environment`, `CheckEnvironment`, `OrgEnvironments`, `EnvironmentRevision`, ...). Fields that
    were pointers or `int32` are now values and `int`; `OpenEnvironment.Id` is `OpenEnvironment.ID`
  - `EnvironmentDefinitionValues.EnvironmentVariables` and `Files` are maps instead of pointers to maps
  - Revision and tag counts, `before`, `after` and `revision` parameters are `int`; the `after` cursor of
    `ListEnvironmentTagsPaginated` is an `int`
  - `Configuration` is a plain struct with `BaseURL`, `UserAgent`, `DefaultHeader` and `HTTPClient`;
    `BaseURL` is the scheme and host of the backend without the `/api/esc` path. `NewCustomBackendConfiguration`
    keeps the port of the URL it is given
  - `EscClient.EscAPI` is replaced by `EscClient.Cloud`, the configured `*apiclient.CloudClient`
  - Errors from Pulumi Cloud are `*apiclient.APIError`; `GenericOpenAPIError`, `ContextAPIKeys`, `APIKey`,
    the `Nullable*` helpers and the `Ptr*` helpers are gone. `AccessTokenFromContext` and `WithAccessToken`
    read and set the token `NewAuthContext` stores
  - `GetEnvironmentETag` is gone: the generated client cannot read response headers
  - `CheckEnvironmentYaml` and `UpdateEnvironmentYaml` return the diagnostics of a rejected definition together
    with the HTTP 400 error
  [#143](https://github.com/pulumi/esc-sdk/pull/143)

- Default credentials are now sourced exclusively from the `PULUMI_ACCESS_TOKEN` and
  `PULUMI_BACKEND_URL` environment variables. The SDKs no longer fall back to reusing the
  Pulumi/ESC CLI login on disk (`~/.pulumi/credentials.json`). This removes the Go SDK's
  dependency on `github.com/pulumi/esc` (and its transitive `pulumi/pulumi` dependency) and
  removes the credential-file readers from the Python, TypeScript, and C# SDKs. Callers that
  relied on automatic CLI-login pickup must now set `PULUMI_ACCESS_TOKEN` (and optionally
  `PULUMI_BACKEND_URL`) or pass credentials explicitly.
  [#137](https://github.com/pulumi/esc-sdk/pull/137)
