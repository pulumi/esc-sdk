// Copyright 2024, Pulumi Corporation.  All rights reserved.

using System.Collections.Generic;
using Pulumi.Esc.Sdk.Client;
using Pulumi.Esc.Sdk.Model;
using Xunit;

namespace Pulumi.Esc.Sdk.Tests
{
    /// <summary>
    /// Unit tests for <see cref="EnvironmentDefinitionSerializer"/>.
    /// </summary>
    public class EnvironmentDefinitionSerializerTests
    {
        [Fact]
        public void Deserialize_Null_ReturnsNull()
        {
            Assert.Null(EnvironmentDefinitionSerializer.Deserialize(null));
        }

        [Fact]
        public void Deserialize_Empty_ReturnsNull()
        {
            Assert.Null(EnvironmentDefinitionSerializer.Deserialize(""));
            Assert.Null(EnvironmentDefinitionSerializer.Deserialize("   "));
        }

        [Fact]
        public void Deserialize_ValuesOnly()
        {
            var yaml = @"
values:
  environmentVariables:
    FOO: bar
    BAZ: qux
  pulumiConfig:
    aws:region: us-west-2
";

            var result = EnvironmentDefinitionSerializer.Deserialize(yaml);
            Assert.NotNull(result);
            Assert.NotNull(result!.Values);

            var envVars = result.Values!.EnvironmentVariables;
            Assert.NotNull(envVars);
            Assert.Equal("bar", envVars!["FOO"]);
            Assert.Equal("qux", envVars["BAZ"]);

            var pulumiConfig = result.Values.PulumiConfig;
            Assert.NotNull(pulumiConfig);
            Assert.Equal("us-west-2", pulumiConfig!["aws:region"]);
        }

        [Fact]
        public void Deserialize_WithImports()
        {
            var yaml = @"
imports:
  - myproject/base-env
  - myproject/secrets
values:
  environmentVariables:
    APP_ENV: production
";

            var result = EnvironmentDefinitionSerializer.Deserialize(yaml);
            Assert.NotNull(result);

            Assert.NotNull(result!.Imports);
            Assert.Equal(2, result.Imports!.Count);
            Assert.Equal("myproject/base-env", result.Imports[0]);
            Assert.Equal("myproject/secrets", result.Imports[1]);

            Assert.NotNull(result.Values?.EnvironmentVariables);
            Assert.Equal("production", result.Values!.EnvironmentVariables!["APP_ENV"]);
        }

        [Fact]
        public void Deserialize_WithFiles()
        {
            var yaml = @"
values:
  files:
    KUBECONFIG: contents-here
";

            var result = EnvironmentDefinitionSerializer.Deserialize(yaml);
            Assert.NotNull(result);
            Assert.NotNull(result!.Values?.Files);
            Assert.Equal("contents-here", result.Values!.Files!["KUBECONFIG"]);
        }

        [Fact]
        public void Deserialize_EmptyValues()
        {
            var yaml = @"
values: {}
";

            var result = EnvironmentDefinitionSerializer.Deserialize(yaml);
            // The YAML parser may return an empty dict for values: {}
            // which won't match the Dictionary<object, object> type check
            Assert.NotNull(result);
        }

        [Fact]
        public void Serialize_ProducesValidJson()
        {
            var definition = new EnvironmentDefinition(
                imports: new Option<List<string>?>(new List<string> { "project/base" }),
                values: new Option<EnvironmentDefinitionValues?>(new EnvironmentDefinitionValues(
                    environmentVariables: new Option<Dictionary<string, string>?>(new Dictionary<string, string>
                    {
                        ["FOO"] = "bar",
                    }),
                    files: default,
                    pulumiConfig: default
                ))
            );

            var json = EnvironmentDefinitionSerializer.Serialize(definition);
            Assert.Contains("\"imports\"", json);
            Assert.Contains("\"project/base\"", json);
            Assert.Contains("\"FOO\"", json);
            Assert.Contains("\"bar\"", json);
        }

        [Fact]
        public void SerializeToYaml_ProducesValidYaml()
        {
            var definition = new EnvironmentDefinition(
                imports: new Option<List<string>?>(new List<string> { "project/base" }),
                values: new Option<EnvironmentDefinitionValues?>(new EnvironmentDefinitionValues(
                    environmentVariables: new Option<Dictionary<string, string>?>(new Dictionary<string, string>
                    {
                        ["FOO"] = "bar",
                    }),
                    files: default,
                    pulumiConfig: default
                ))
            );

            var yaml = EnvironmentDefinitionSerializer.SerializeToYaml(definition);
            Assert.Contains("imports:", yaml);
            Assert.Contains("project/base", yaml);
            Assert.Contains("environmentVariables:", yaml);
            Assert.Contains("FOO: bar", yaml);
        }

        [Fact]
        public void RoundTrip_DeserializeThenSerializeToYaml()
        {
            var originalYaml = @"
imports:
  - myproject/base
values:
  environmentVariables:
    DB_HOST: localhost
    DB_PORT: ""5432""
  pulumiConfig:
    aws:region: us-east-1
";

            var definition = EnvironmentDefinitionSerializer.Deserialize(originalYaml);
            Assert.NotNull(definition);

            var roundTripped = EnvironmentDefinitionSerializer.SerializeToYaml(definition!);
            Assert.Contains("imports:", roundTripped);
            Assert.Contains("myproject/base", roundTripped);
            Assert.Contains("environmentVariables:", roundTripped);
            Assert.Contains("DB_HOST: localhost", roundTripped);
            Assert.Contains("DB_PORT:", roundTripped);
            Assert.Contains("pulumiConfig:", roundTripped);
            Assert.Contains("aws:region: us-east-1", roundTripped);
        }
    }
}
