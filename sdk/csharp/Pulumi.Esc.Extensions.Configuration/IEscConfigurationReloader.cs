// Copyright 2026, Pulumi Corporation.  All rights reserved.

namespace Pulumi.Esc.Extensions.Configuration;

public interface IEscConfigurationReloader
{
    Task ReloadAsync(CancellationToken cancellationToken = default);
}
