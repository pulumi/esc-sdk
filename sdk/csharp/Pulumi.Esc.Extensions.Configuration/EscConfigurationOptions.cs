// Copyright 2026, Pulumi Corporation.  All rights reserved.

namespace Pulumi.Esc.Extensions.Configuration;

public class EscConfigurationOptions
{
    public required string Organization { get; set; }
    public required string Project { get; set; }
    public required string Environment { get; set; }
    public string? AccessToken { get; set; }
}
