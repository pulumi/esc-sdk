using Pulumi.Esc.Sdk.Model;

namespace Pulumi.Esc.Extensions.Configuration;

internal interface IEscClient : IDisposable
{
    Task<(string Id, List<EnvironmentDiagnostic>? Diagnostics)> OpenEnvironmentAsync(
        string orgName,
        string projectName,
        string envName,
        CancellationToken cancellationToken = default);

    Task<(ModelEnvironment Environment, Dictionary<string, object?>? Values)> ReadOpenEnvironmentAsync(
        string orgName,
        string projectName,
        string envName,
        string openSessionId,
        CancellationToken cancellationToken = default);
}
