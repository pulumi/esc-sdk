// Copyright 2025, Pulumi Corporation.

import assert from "assert";
import { DefaultClient } from "index";
import { after, before, describe, it } from "node:test";
import path from "path";

describe("ESC", async () => {
    let tokenBefore: string | undefined
    let backendBefore: string | undefined
    let homeBefore: string | undefined

    before(async () => {
        tokenBefore = process.env.PULUMI_ACCESS_TOKEN;
        backendBefore = process.env.PULUMI_BACKEND_URL;
        homeBefore = process.env.PULUMI_HOME;
        delete process.env.PULUMI_ACCESS_TOKEN;
        delete process.env.PULUMI_BACKEND_URL;
    });

    after(async () => {
        process.env.PULUMI_ACCESS_TOKEN = tokenBefore;
        process.env.PULUMI_BACKEND_URL = backendBefore;
        process.env.PULUMI_HOME = homeBefore;
    });

    it("test no creds at all", async () => {
        process.env.PULUMI_HOME = "/not_real_dir";
        let client = DefaultClient();
        assert.equal(client.config.basePath, undefined);
        assert.equal(client.config.accessToken, undefined);
    });

    it("test just pulumi creds", async () => {
        process.env.PULUMI_HOME = path.dirname(process.cwd()) + "/test/test_pulumi_home";
        let client = DefaultClient();
        assert.equal(client.config.basePath, "https://api.moolumi.com/api/esc");
        assert.equal(client.config.accessToken, "pul-fake-token-moo");
    });

    it("test pulumi creds with esc", async () => {
        process.env.PULUMI_HOME = path.dirname(process.cwd()) + "/test/test_pulumi_home_esc";
        let client = DefaultClient();
        assert.equal(client.config.basePath, "https://api.boolumi.com/api/esc");
        assert.equal(client.config.accessToken, "pul-fake-token-boo");
    });

    it("test pulumi creds bad format", async () => {
        process.env.PULUMI_HOME = path.dirname(process.cwd()) + "/test/test_pulumi_home_bad_format";
        let client = DefaultClient();
        assert.equal(client.config.basePath, undefined);
        assert.equal(client.config.accessToken, undefined);
    });

});
