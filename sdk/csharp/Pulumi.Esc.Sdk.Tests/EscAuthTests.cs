// Copyright 2024, Pulumi Corporation.  All rights reserved.

using System;
using Xunit;

namespace Pulumi.Esc.Sdk.Tests
{
    /// <summary>
    /// Unit tests for <see cref="EscAuth"/> credential resolution.
    /// Default credentials are sourced from environment variables only, never
    /// from the Pulumi/ESC CLI credentials file on disk.
    /// </summary>
    public class EscAuthTests : IDisposable
    {
        private readonly string? _tokenBefore;
        private readonly string? _backendBefore;
        private readonly string? _homeBefore;

        public EscAuthTests()
        {
            _tokenBefore = Environment.GetEnvironmentVariable("PULUMI_ACCESS_TOKEN");
            _backendBefore = Environment.GetEnvironmentVariable("PULUMI_BACKEND_URL");
            _homeBefore = Environment.GetEnvironmentVariable("PULUMI_HOME");

            Environment.SetEnvironmentVariable("PULUMI_ACCESS_TOKEN", "");
            Environment.SetEnvironmentVariable("PULUMI_BACKEND_URL", "");
        }

        public void Dispose()
        {
            Environment.SetEnvironmentVariable("PULUMI_ACCESS_TOKEN", _tokenBefore ?? "");
            Environment.SetEnvironmentVariable("PULUMI_BACKEND_URL", _backendBefore ?? "");
            Environment.SetEnvironmentVariable("PULUMI_HOME", _homeBefore ?? "");
        }

        [Fact]
        public void NoCreds_ThrowsAndDefaultsUrl()
        {
            Assert.Throws<InvalidOperationException>(() => EscAuth.GetDefaultAccessToken());
            Assert.Equal("https://api.pulumi.com", EscAuth.GetDefaultBackendUrl());
        }

        [Fact]
        public void ReadsCredentialsFromEnvVars()
        {
            Environment.SetEnvironmentVariable("PULUMI_ACCESS_TOKEN", "env-token-123");
            Environment.SetEnvironmentVariable("PULUMI_BACKEND_URL", "https://custom.backend.com");

            Assert.Equal("env-token-123", EscAuth.GetDefaultAccessToken());
            Assert.Equal("https://custom.backend.com", EscAuth.GetDefaultBackendUrl());
        }

        [Fact]
        public void EnvToken_WithDefaultBackend()
        {
            Environment.SetEnvironmentVariable("PULUMI_ACCESS_TOKEN", "env-token-123");

            Assert.Equal("env-token-123", EscAuth.GetDefaultAccessToken());
            Assert.Equal("https://api.pulumi.com", EscAuth.GetDefaultBackendUrl());
        }

        [Fact]
        public void CliCredentialsOnDisk_AreIgnored()
        {
            // Even with a populated Pulumi home, default credentials must not be
            // read from disk; only environment variables are honored.
            Environment.SetEnvironmentVariable("PULUMI_HOME", "/some/pulumi/home");

            Assert.Throws<InvalidOperationException>(() => EscAuth.GetDefaultAccessToken());
            Assert.Equal("https://api.pulumi.com", EscAuth.GetDefaultBackendUrl());
        }

        [Theory]
        [InlineData("https://api.pulumi.com", "https://api.pulumi.com/api/esc")]
        [InlineData("https://api.pulumi.com/", "https://api.pulumi.com/api/esc")]
        [InlineData("https://api.pulumi.com/api/esc", "https://api.pulumi.com/api/esc")]
        [InlineData("http://localhost:8080", "http://localhost:8080/api/esc")]
        [InlineData("https://example.com/pulumi", "https://example.com/pulumi/api/esc")]
        [InlineData("https://example.com/pulumi/", "https://example.com/pulumi/api/esc")]
        [InlineData("https://example.com:8443/pulumi/api/esc", "https://example.com:8443/pulumi/api/esc")]
        [InlineData("https://example.com/team%2Fname", "https://example.com/team%2Fname/api/esc")]
        public void GetEscApiUrl_ConvertsBackendUrl(string backendUrl, string expected)
        {
            Assert.Equal(expected, EscAuth.GetEscApiUrl(backendUrl));
        }
    }
}
