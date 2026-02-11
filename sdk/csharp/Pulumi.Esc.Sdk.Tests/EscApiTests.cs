// Copyright 2024, Pulumi Corporation.  All rights reserved.

using System;
using System.Collections.Generic;
using System.Threading.Tasks;
using Xunit;

namespace Pulumi.Esc.Sdk.Tests
{
    /// <summary>
    /// End-to-end integration tests for the ESC C# SDK.
    /// Requires PULUMI_ACCESS_TOKEN and PULUMI_ORG environment variables to be set.
    /// Mirrors the Go test in sdk/go/api_esc_test.go.
    /// </summary>
    [Trait("Category", "Integration")]
    public class EscApiTests : IAsyncLifetime
    {
        private const string ProjectName = "sdk-csharp-test";
        private const string EnvPrefix = "env-";

        private readonly string _orgName;
        private EscClient _client = null!;

        public EscApiTests()
        {
            _orgName = Environment.GetEnvironmentVariable("PULUMI_ORG")
                ?? throw new InvalidOperationException("PULUMI_ORG must be set");
        }

        public async Task InitializeAsync()
        {
            _client = EscClient.CreateDefault();
            await RemoveAllCSharpTestEnvsAsync();
        }

        public Task DisposeAsync()
        {
            _client.Dispose();
            return Task.CompletedTask;
        }

        [Fact]
        public async Task FullLifecycle_Create_Clone_List_Update_Get_Decrypt_Open_Tags_Delete()
        {
            var timestamp = DateTime.UtcNow.ToString("yyyyMMddHHmmss");
            var baseEnvName = $"base-{timestamp}";
            var envName = $"{EnvPrefix}{timestamp}";
            var cloneProject = $"{ProjectName}-clone";
            var cloneName = $"{envName}-clone";

            try
            {
                // -- Create base environment --
                await _client.CreateEnvironmentAsync(_orgName, ProjectName, baseEnvName);

                var baseYaml = $@"
values:
  base: {baseEnvName}
";
                await _client.UpdateEnvironmentYamlAsync(_orgName, ProjectName, baseEnvName, baseYaml);

                // -- Create and clone environment --
                await _client.CreateEnvironmentAsync(_orgName, ProjectName, envName);
                await _client.CloneEnvironmentAsync(_orgName, ProjectName, envName, cloneProject, cloneName);

                // -- List environments --
                var envs = await _client.ListEnvironmentsAsync(_orgName);
                AssertFindEnvironment(envs, ProjectName, envName);
                AssertFindEnvironment(envs, cloneProject, cloneName);

                // -- Open and read (empty env should have no values) --
                var (_, values) = await _client.OpenAndReadEnvironmentAsync(_orgName, ProjectName, envName);
                Assert.Null(values);

                // -- Update with YAML --
                var yaml = $@"imports:
  - {ProjectName}/{baseEnvName}
values:
  foo: bar
  my_secret:
    fn::secret: ""shh! don't tell anyone""
  my_array: [1, 2, 3]
  pulumiConfig:
    foo: ${{foo}}
  environmentVariables:
    FOO: ${{foo}}
";
                var diags = await _client.UpdateEnvironmentYamlAsync(_orgName, ProjectName, envName, yaml);

                // -- GetEnvironment (parsed from YAML) --
                var envDef = await _client.GetEnvironmentAsync(_orgName, ProjectName, envName);
                Assert.NotNull(envDef);
                Assert.NotNull(envDef!.Imports);
                Assert.Contains($"{ProjectName}/{baseEnvName}", envDef.Imports!);

                // -- DecryptEnvironment --
                var decrypted = await _client.DecryptEnvironmentAsync(_orgName, ProjectName, envName);
                Assert.NotNull(decrypted);

                // -- Open and read (should have resolved values) --
                var (env, resolvedValues) = await _client.OpenAndReadEnvironmentAsync(_orgName, ProjectName, envName);
                Assert.NotNull(resolvedValues);
                Assert.Equal(baseEnvName, resolvedValues!["base"]);
                Assert.Equal("bar", resolvedValues["foo"]);
                Assert.Equal("shh! don't tell anyone", resolvedValues["my_secret"]);

                // -- Read property --
                var (openSessionId, _) = await _client.OpenEnvironmentAsync(_orgName, ProjectName, envName);
                var (propValue, propPrimitive) = await _client.ReadOpenEnvironmentPropertyAsync(
                    _orgName, ProjectName, envName, openSessionId, "pulumiConfig.foo");
                Assert.Equal("bar", propPrimitive);

                // -- GetEnvironmentAtVersion --
                await _client.GetEnvironmentAtVersionAsync(_orgName, ProjectName, envName, "2");

                // -- Revisions --
                var revisions = await _client.ListEnvironmentRevisionsAsync(_orgName, ProjectName, envName);
                Assert.True(revisions.Count >= 2);

                // -- Revision tags --
                await _client.CreateEnvironmentRevisionTagAsync(_orgName, ProjectName, envName, "testTag", 2);

                var revTags = await _client.ListEnvironmentRevisionTagsAsync(_orgName, ProjectName, envName);
                Assert.Equal(2, revTags.Tags!.Count);

                await _client.UpdateEnvironmentRevisionTagAsync(_orgName, ProjectName, envName, "testTag", 2);

                var testTag = await _client.GetEnvironmentRevisionTagAsync(_orgName, ProjectName, envName, "testTag");
                Assert.Equal(2, testTag.Revision);

                await _client.DeleteEnvironmentRevisionTagAsync(_orgName, ProjectName, envName, "testTag");

                revTags = await _client.ListEnvironmentRevisionTagsAsync(_orgName, ProjectName, envName);
                Assert.Single(revTags.Tags!);

                // -- Environment tags --
                await _client.CreateEnvironmentTagAsync(_orgName, ProjectName, envName, "owner", "esc-sdk-test");

                var envTags = await _client.ListEnvironmentTagsAsync(_orgName, ProjectName, envName);
                Assert.NotNull(envTags.Tags);
                Assert.True(envTags.Tags!.ContainsKey("owner"));
                Assert.Equal("owner", envTags.Tags!["owner"].Name);

                await _client.UpdateEnvironmentTagAsync(_orgName, ProjectName, envName,
                    "owner", "esc-sdk-test", "new-owner", "esc-sdk-test-updated");

                var envTag = await _client.GetEnvironmentTagAsync(_orgName, ProjectName, envName, "new-owner");
                Assert.Equal("new-owner", envTag.Name);

                await _client.DeleteEnvironmentTagAsync(_orgName, ProjectName, envName, "new-owner");

                envTags = await _client.ListEnvironmentTagsAsync(_orgName, ProjectName, envName);
                Assert.Empty(envTags.Tags);

                // -- Check environment YAML --
                var checkYaml = @"
values:
  foo: bar
  pulumiConfig:
    foo: ${bad_ref}
";
                var checkResult = await _client.CheckEnvironmentYamlAsync(_orgName, checkYaml);
                Assert.NotNull(checkResult);
                Assert.NotNull(checkResult!.Diagnostics);
                Assert.NotEmpty(checkResult.Diagnostics);
            }
            finally
            {
                // Cleanup — delete environments
                await SafeDelete(_orgName, ProjectName, envName);
                await SafeDelete(_orgName, cloneProject, cloneName);
                await SafeDelete(_orgName, ProjectName, baseEnvName);
            }
        }

        private async Task SafeDelete(string orgName, string projectName, string envName)
        {
            try
            {
                await _client.DeleteEnvironmentAsync(orgName, projectName, envName);
            }
            catch
            {
                // Ignore — cleanup is best-effort
            }
        }

        private static void AssertFindEnvironment(Model.OrgEnvironments envs, string project, string name)
        {
            Assert.Contains(envs.Environments!, e => e.Project == project && e.Name == name);
        }

        private async Task RemoveAllCSharpTestEnvsAsync()
        {
            string? continuationToken = null;
            do
            {
                var envs = await _client.ListEnvironmentsAsync(_orgName, continuationToken);
                foreach (var env in envs.Environments ?? new List<Model.OrgEnvironment>())
                {
                    if (env.Project == ProjectName && env.Name.StartsWith(EnvPrefix))
                    {
                        await SafeDelete(_orgName, ProjectName, env.Name);
                    }
                }
                continuationToken = envs.NextToken;
            } while (!string.IsNullOrEmpty(continuationToken));
        }
    }
}
