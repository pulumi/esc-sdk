using Microsoft.Extensions.Configuration;
using Pulumi.Esc.Sdk;

namespace Pulumi.Esc.Extensions.Configuration;

public class EscConfigurationSource : IConfigurationSource
{
    private readonly EscConfigurationOptions _options;

    public IEscConfigurationReloader? Reloader { get; private set; }

    public EscConfigurationSource(EscConfigurationOptions options)
    {
        _options = options;
    }

    public IConfigurationProvider Build(IConfigurationBuilder builder)
    {
        var escClient = _options.AccessToken is not null
            ? EscClient.Create(_options.AccessToken)
            : EscClient.CreateDefault();
        var provider = new EscConfigurationProvider(_options, new EscClientAdapter(escClient));
        Reloader = provider;
        return provider;
    }
}
