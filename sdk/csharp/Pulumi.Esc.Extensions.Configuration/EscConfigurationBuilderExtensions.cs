using Microsoft.Extensions.Configuration;

namespace Pulumi.Esc.Extensions.Configuration;

public static class EscConfigurationBuilderExtensions
{
    public static IConfigurationBuilder AddEscConfiguration(
        this IConfigurationBuilder builder,
        string organization,
        string project,
        string environment,
        string? accessToken = null)
    {
        return builder.Add(new EscConfigurationSource(new EscConfigurationOptions
        {
            Organization = organization,
            Project = project,
            Environment = environment,
            AccessToken = accessToken
        }));
    }

    public static IConfigurationBuilder AddEscConfiguration(
        this IConfigurationBuilder builder,
        EscConfigurationOptions options)
    {
        return builder.Add(new EscConfigurationSource(options));
    }
}
