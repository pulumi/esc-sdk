using Microsoft.Extensions.Configuration;

namespace Pulumi.Esc.Extensions.Configuration;

public class EscConfigurationProvider : ConfigurationProvider, IEscConfigurationReloader, IDisposable
{
    private readonly EscConfigurationOptions _options;
    private readonly IEscClient _client;

    internal EscConfigurationProvider(EscConfigurationOptions options, IEscClient client)
    {
        _options = options;
        _client = client;
    }

    public override void Load() => LoadAsync().GetAwaiter().GetResult();

    public async Task ReloadAsync(CancellationToken cancellationToken = default)
    {
        await LoadAsync(cancellationToken);
        OnReload();
    }

    private async Task LoadAsync(CancellationToken cancellationToken = default)
    {
        var (sessionId, _) = await _client.OpenEnvironmentAsync(
            _options.Organization, _options.Project, _options.Environment, cancellationToken);

        var (_, values) = await _client.ReadOpenEnvironmentAsync(
            _options.Organization, _options.Project, _options.Environment, sessionId, cancellationToken);

        var data = new Dictionary<string, string?>(StringComparer.OrdinalIgnoreCase);
        if (values is not null)
        {
            foreach (var kvp in values)
                Flatten(kvp.Key, kvp.Value, data);
        }

        Data = data;
    }

    private static void Flatten(string prefix, object? value, Dictionary<string, string?> result)
    {
        switch (value)
        {
            case Dictionary<string, object?> dict:
                foreach (var kvp in dict)
                    Flatten($"{prefix}:{kvp.Key}", kvp.Value, result);
                break;
            case List<object?> list:
                for (var i = 0; i < list.Count; i++)
                    Flatten($"{prefix}:{i}", list[i], result);
                break;
            default:
                result[prefix] = value?.ToString();
                break;
        }
    }

    public void Dispose() => _client.Dispose();
}
