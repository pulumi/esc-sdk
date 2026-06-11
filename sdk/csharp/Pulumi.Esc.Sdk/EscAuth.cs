// Copyright 2024, Pulumi Corporation.  All rights reserved.

using System;

namespace Pulumi.Esc.Sdk
{
    /// <summary>
    /// Provides authentication helpers for the Pulumi ESC SDK.
    /// </summary>
    public static class EscAuth
    {
        private const string DefaultPulumiCloudUrl = "https://api.pulumi.com";
        private const string PulumiAccessTokenEnvVar = "PULUMI_ACCESS_TOKEN";
        private const string PulumiBackendUrlEnvVar = "PULUMI_BACKEND_URL";

        /// <summary>
        /// Gets the default access token from the PULUMI_ACCESS_TOKEN environment variable.
        /// </summary>
        /// <returns>The access token string.</returns>
        /// <exception cref="InvalidOperationException">When no access token can be found.</exception>
        public static string GetDefaultAccessToken()
        {
            var envToken = Environment.GetEnvironmentVariable(PulumiAccessTokenEnvVar);
            if (!string.IsNullOrEmpty(envToken))
            {
                return envToken;
            }

            throw new InvalidOperationException(
                "No Pulumi Access Token found. Set the PULUMI_ACCESS_TOKEN environment variable.");
        }

        /// <summary>
        /// Gets the default backend URL from the PULUMI_BACKEND_URL environment variable,
        /// defaulting to https://api.pulumi.com when it is not set.
        /// </summary>
        /// <returns>The backend URL string.</returns>
        public static string GetDefaultBackendUrl()
        {
            var envUrl = Environment.GetEnvironmentVariable(PulumiBackendUrlEnvVar);
            if (!string.IsNullOrEmpty(envUrl))
            {
                return envUrl;
            }

            return DefaultPulumiCloudUrl;
        }

        /// <summary>
        /// Converts a backend URL (e.g. "https://api.pulumi.com") to the ESC API URL.
        /// </summary>
        /// <param name="backendUrl">The Pulumi backend URL.</param>
        /// <returns>The ESC API base URL (e.g. "https://api.pulumi.com/api/esc").</returns>
        public static string GetEscApiUrl(string backendUrl)
        {
            var uri = new Uri(backendUrl.TrimEnd('/'));
            return $"{uri.Scheme}://{uri.Host}{(uri.IsDefaultPort ? "" : $":{uri.Port}")}/api/esc";
        }
    }
}
