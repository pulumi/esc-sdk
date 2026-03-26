namespace Pulumi.Esc.Extensions.Configuration;

public interface IEscConfigurationReloader
{
    Task ReloadAsync(CancellationToken cancellationToken = default);
}
