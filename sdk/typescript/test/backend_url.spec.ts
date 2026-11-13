// Copyright 2026, Pulumi Corporation.

import assert from "assert";
import { Configuration, EscApi } from "index";
import { describe, it } from "node:test";

describe("EscApi backend URL", () => {
    const cases: [string, string][] = [
        ["https://api.pulumi.com", "https://api.pulumi.com/api/esc"],
        ["https://api.pulumi.com/", "https://api.pulumi.com/api/esc"],
        ["https://api.pulumi.com/api/esc", "https://api.pulumi.com/api/esc"],
        ["http://localhost:8080", "http://localhost:8080/api/esc"],
        ["https://example.com/pulumi", "https://example.com/pulumi/api/esc"],
        ["https://example.com/pulumi/", "https://example.com/pulumi/api/esc"],
        ["https://example.com:8443/pulumi/api/esc", "https://example.com:8443/pulumi/api/esc"],
    ];
    for (const [input, want] of cases) {
        it(input, () => {
            const config = new Configuration({ basePath: input });
            const api = new EscApi(config);
            assert.equal(api.config.basePath, want);
            assert.equal(config.basePath, input);
        });
    }
});
