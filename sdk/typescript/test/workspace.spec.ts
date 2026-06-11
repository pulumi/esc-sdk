// Copyright 2025, Pulumi Corporation.

import assert from "assert";
import { DefaultClient } from "index";
import { after, before, describe, it } from "node:test";

describe("ESC default credentials", async () => {
    let tokenBefore: string | undefined
    let backendBefore: string | undefined
    let homeBefore: string | undefined

    before(async () => {
        tokenBefore = process.env.PULUMI_ACCESS_TOKEN;
        backendBefore = process.env.PULUMI_BACKEND_URL;
        homeBefore = process.env.PULUMI_HOME;
        delete process.env.PULUMI_ACCESS_TOKEN;
        delete process.env.PULUMI_BACKEND_URL;
        delete process.env.PULUMI_HOME;
    });

    after(async () => {
        process.env.PULUMI_ACCESS_TOKEN = tokenBefore;
        process.env.PULUMI_BACKEND_URL = backendBefore;
        process.env.PULUMI_HOME = homeBefore;
    });

    it("test no creds at all", async () => {
        delete process.env.PULUMI_ACCESS_TOKEN;
        delete process.env.PULUMI_BACKEND_URL;
        let client = DefaultClient();
        assert.equal(client.config.basePath, undefined);
        assert.equal(client.config.accessToken, undefined);
    });

    it("test reads credentials from environment variables", async () => {
        process.env.PULUMI_ACCESS_TOKEN = "pul-fake-token-env";
        process.env.PULUMI_BACKEND_URL = "https://api.moolumi.com";
        let client = DefaultClient();
        assert.equal(client.config.basePath, "https://api.moolumi.com/api/esc");
        assert.equal(client.config.accessToken, "pul-fake-token-env");
    });

    it("test CLI credentials on disk are ignored", async () => {
        // Even with a populated Pulumi home, default credentials must not be
        // read from disk; only environment variables are honored.
        delete process.env.PULUMI_ACCESS_TOKEN;
        delete process.env.PULUMI_BACKEND_URL;
        process.env.PULUMI_HOME = "/some/pulumi/home";
        let client = DefaultClient();
        assert.equal(client.config.basePath, undefined);
        assert.equal(client.config.accessToken, undefined);
    });

});
