# coding: utf-8

# Copyright 2024, Pulumi Corporation.  All rights reserved.

from typing import Optional
import unittest
import os

import pulumi_esc_sdk as esc

class TestWorkspaceAccounts(unittest.TestCase):
    """WorkspaceAccounts unit test stubs"""
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

    def test_no_creds_at_all(self):
        os.environ["PULUMI_HOME"] = "/not_real_dir"
        self.client = esc.esc_client.default_client()
        self.config = self.client.esc_api.api_client.configuration
        self.assertEqual(self.config.host, "https://api.pulumi.com/api/esc")
        self.assertTrue('Authorization' not in self.config.api_key)

    def test_just_pulumi_creds(self):
        os.environ["PULUMI_HOME"] = os.path.dirname(os.getcwd()) + "/test/test_pulumi_home"
        self.client = esc.esc_client.default_client()
        self.config = self.client.esc_api.api_client.configuration
        self.assertEqual(self.config.host, "https://api.moolumi.com/api/esc")
        self.assertEqual(self.config.api_key['Authorization'], "pul-fake-token-moo")

    def test_pulumi_creds_with_esc(self):
        os.environ["PULUMI_HOME"] = os.path.dirname(os.getcwd()) + "/test/test_pulumi_home_esc"
        self.client = esc.esc_client.default_client()
        self.config = self.client.esc_api.api_client.configuration
        self.assertEqual(self.config.host, "https://api.boolumi.com/api/esc")
        self.assertEqual(self.config.api_key['Authorization'], "pul-fake-token-boo")

    def test_pulumi_creds_bad_format(self):
        os.environ["PULUMI_HOME"] = os.path.dirname(os.getcwd()) + "/test/test_pulumi_home_bad_format"
        self.client = esc.esc_client.default_client()
        self.config = self.client.esc_api.api_client.configuration
        self.assertEqual(self.config.host, "https://api.pulumi.com/api/esc")
        self.assertTrue('Authorization' not in self.config.api_key)

if __name__ == '__main__':
    unittest.main()
