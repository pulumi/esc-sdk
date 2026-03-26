using Pulumi.Esc.Sdk;
using Pulumi.Esc.Sdk.Model;

namespace Pulumi.Esc.Extensions.Configuration;

internal sealed class EscClientAdapter : IEscClient
{
    private readonly EscClient _client;

    public EscClientAdapter(EscClient client) => _client = client;

    public Task<(string Id, List<EnvironmentDiagnostic>? Diagnostics)> OpenEnvironmentAsync(
        string orgName,
        string projectName,
        string envName,
        CancellationToken cancellationToken = default)
        => _client.OpenEnvironmentAsync(orgName, projectName, envName, cancellationToken);

    public Task<(ModelEnvironment Environment, Dictionary<string, object?>? Values)> ReadOpenEnvironmentAsync(
        string orgName,
        string projectName,
        string envName,
        string openSessionId,
        CancellationToken cancellationToken = default)
        => _client.ReadOpenEnvironmentAsync(orgName, projectName, envName, openSessionId, cancellationToken);

    public void Dispose() => _client.Dispose();
}
