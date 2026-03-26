using Microsoft.Extensions.Configuration;
using Microsoft.Extensions.DependencyInjection;

namespace Pulumi.Esc.Extensions.Configuration;

public static class EscServiceCollectionExtensions
{
    public static IServiceCollection AddEscConfigurationReloader(
        this IServiceCollection services,
        IConfiguration configuration)
    {
        var root = (IConfigurationRoot)configuration;
        var provider = root.Providers.OfType<EscConfigurationProvider>().First();
        return services.AddSingleton<IEscConfigurationReloader>(provider);
    }
}
