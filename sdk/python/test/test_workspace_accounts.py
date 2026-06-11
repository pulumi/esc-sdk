# coding: utf-8

# Copyright 2024, Pulumi Corporation.  All rights reserved.

from typing import Optional
import unittest
import os

import pulumi_esc_sdk as esc


class TestDefaultCredentials(unittest.TestCase):
    """Tests that default credentials are sourced from environment variables only,
    and never from the Pulumi/ESC CLI credentials file on disk."""
    tokenBefore: Optional[str]
    backendBefore: Optional[str]
    homeBefore: Optional[str]

    def setUp(self) -> None:
        self.tokenBefore = os.getenv("PULUMI_ACCESS_TOKEN")
        self.backendBefore = os.getenv("PULUMI_BACKEND_URL")
        self.homeBefore = os.getenv("PULUMI_HOME")
        os.environ["PULUMI_ACCESS_TOKEN"] = ""
        os.environ["PULUMI_BACKEND_URL"] = ""

    def tearDown(self) -> None:
        os.environ["PULUMI_ACCESS_TOKEN"] = self.tokenBefore or ''
        os.environ["PULUMI_BACKEND_URL"] = self.backendBefore or ''
        os.environ["PULUMI_HOME"] = self.homeBefore or ''

    def test_no_creds(self):
        self.client = esc.esc_client.default_client()
        self.config = self.client.esc_api.api_client.configuration
        self.assertEqual(self.config.host, "https://api.pulumi.com/api/esc")
        self.assertTrue('Authorization' not in self.config.api_key)

    def test_env_vars(self):
        os.environ["PULUMI_ACCESS_TOKEN"] = "pul-fake-token-env"
        os.environ["PULUMI_BACKEND_URL"] = "https://api.moolumi.com"
        self.client = esc.esc_client.default_client()
        self.config = self.client.esc_api.api_client.configuration
        self.assertEqual(self.config.host, "https://api.moolumi.com/api/esc")
        self.assertEqual(self.config.api_key['Authorization'], "pul-fake-token-env")

    def test_env_token_default_backend(self):
        os.environ["PULUMI_ACCESS_TOKEN"] = "pul-fake-token-env"
        self.client = esc.esc_client.default_client()
        self.config = self.client.esc_api.api_client.configuration
        self.assertEqual(self.config.host, "https://api.pulumi.com/api/esc")
        self.assertEqual(self.config.api_key['Authorization'], "pul-fake-token-env")

    def test_cli_credentials_are_ignored(self):
        # Even with a populated Pulumi home, default credentials must not be
        # read from disk; only environment variables are honored.
        os.environ["PULUMI_HOME"] = "/some/pulumi/home"
        self.client = esc.esc_client.default_client()
        self.config = self.client.esc_api.api_client.configuration
        self.assertEqual(self.config.host, "https://api.pulumi.com/api/esc")
        self.assertTrue('Authorization' not in self.config.api_key)


if __name__ == '__main__':
    unittest.main()
