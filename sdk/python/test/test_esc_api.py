# coding: utf-8

# Copyright 2024, Pulumi Corporation.  All rights reserved.

import unittest
import os
from datetime import datetime

import pulumi_esc_sdk as esc

PROJECT_NAME = "sdk-python-test"
ENV_PREFIX = "env"


class TestEscApi(unittest.TestCase):
    """EscApi unit test stubs"""

    def setUp(self) -> None:
        self.accessToken = os.getenv("PULUMI_ACCESS_TOKEN")
        self.assertIsNotNone(self.accessToken, "PULUMI_ACCESS_TOKEN must be set")

        self.orgName = os.getenv("PULUMI_ORG")
        self.assertIsNotNone(self.orgName, "PULUMI_ORG must be set")

        configuration = esc.Configuration(access_token=self.accessToken)
        self.client = esc.EscClient(configuration)

        self.remove_all_python_test_envs()

        self.baseEnvName = f"{ENV_PREFIX}-base-{datetime.now().timestamp()}"
        self.client.create_environment(self.orgName, PROJECT_NAME, self.baseEnvName)
        self.envName = None

    def tearDown(self) -> None:
        if self.baseEnvName is not None:
            self.client.delete_environment(self.orgName, PROJECT_NAME, self.baseEnvName)
        if self.envName is not None:
            self.client.delete_environment(self.orgName, PROJECT_NAME, self.envName)

    def test_environment_end_to_end(self) -> None:
        self.envName = f"{ENV_PREFIX}-end-to-end-{datetime.now().timestamp()}"
        self.client.create_environment(self.orgName, PROJECT_NAME, self.envName)

        envs = self.client.list_environments(self.orgName)
        self.assertFindEnv(envs)

        fooReference = "${foo}"
        yaml = f"""
imports:
  - {PROJECT_NAME}/{self.baseEnvName}
values:
  foo: bar
  my_secret:
    fn::secret: "shh! don't tell anyone"
  my_array: [1, 2, 3]
  pulumiConfig:
    foo: {fooReference}
  environmentVariables:
    FOO: {fooReference}
"""
        self.client.update_environment_yaml(self.orgName, PROJECT_NAME, self.envName, yaml)

        env, new_yaml = self.client.get_environment(self.orgName, PROJECT_NAME, self.envName)
        self.assertIsNotNone(env)
        self.assertIsNotNone(new_yaml)

        self.assertEnvDef(env)
        self.assertIsNotNone(env.values.additional_properties["my_secret"])

        decrypted_env, _ = self.client.decrypt_environment(self.orgName, PROJECT_NAME, self.envName)
        self.assertIsNotNone(decrypted_env)
        self.assertEnvDef(decrypted_env)
        self.assertIsNotNone(decrypted_env.values.additional_properties["my_secret"])

        _, values, yaml = self.client.open_and_read_environment(self.orgName, PROJECT_NAME, self.envName)
        self.assertIsNotNone(yaml)

        self.assertEqual(values["foo"], "bar")
        self.assertEqual(values["my_array"], [1, 2, 3])
        self.assertEqual(values["my_secret"], "shh! don't tell anyone")
        self.assertIsNotNone(values["pulumiConfig"])
        self.assertEqual(values["pulumiConfig"]["foo"], "bar")
        self.assertIsNotNone(values["environmentVariables"])
        self.assertEqual(values["environmentVariables"]["FOO"], "bar")

        openInfo = self.client.open_environment(self.orgName, PROJECT_NAME, self.envName)
        self.assertIsNotNone(openInfo)

        v, value = self.client.read_open_environment_property(self.orgName, PROJECT_NAME, self.envName, openInfo.id, "foo")
        self.assertIsNotNone(v)
        self.assertEqual(v.value, "bar")
        self.assertEqual(value, "bar")

        env, _ = self.client.get_environment_at_version(self.orgName, PROJECT_NAME, self.envName, "2")

        env.values.additional_properties["versioned"] = "true"
        self.client.update_environment(self.orgName, PROJECT_NAME, self.envName, env)

        revisions = self.client.list_environment_revisions(self.orgName, PROJECT_NAME, self.envName)
        self.assertIsNotNone(revisions)
        self.assertEqual(len(revisions), 3)

        self.client.create_environment_revision_tag(self.orgName, PROJECT_NAME, self.envName, "testTag", 2)

        _, values, _ = self.client.open_and_read_environment_at_version(self.orgName, PROJECT_NAME, self.envName, "testTag")
        self.assertIsNotNone(values)
        self.assertFalse("versioned" in values)

        tags = self.client.list_environment_revision_tags(self.orgName, PROJECT_NAME, self.envName)
        self.assertIsNotNone(tags)
        self.assertEqual(len(tags.tags), 2)
        self.assertEqual(tags.tags[0].name, "latest")
        self.assertEqual(tags.tags[1].name, "testTag")

        self.client.update_environment_revision_tag(self.orgName, PROJECT_NAME, self.envName, "testTag", 3)

        _, values, _ = self.client.open_and_read_environment_at_version(self.orgName, PROJECT_NAME, self.envName, "testTag")
        self.assertIsNotNone(values)
        self.assertEqual(values["versioned"], "true")

        testTag = self.client.get_environment_revision_tag(self.orgName, PROJECT_NAME, self.envName, "testTag")
        self.assertIsNotNone(testTag)
        self.assertEqual(testTag.revision, 3)

        self.client.delete_environment_revision_tag(self.orgName, PROJECT_NAME, self.envName, "testTag")
        tags = self.client.list_environment_revision_tags(self.orgName, PROJECT_NAME, self.envName)
        self.assertIsNotNone(tags)
        self.assertEqual(len(tags.tags), 1)

        self.client.create_environment_tag(self.orgName, PROJECT_NAME, self.envName, "owner", "esc-sdk-test")

        tags = self.client.list_environment_tags(self.orgName, PROJECT_NAME, self.envName)
        self.assertIsNotNone(tags)
        self.assertEqual(len(tags.tags), 1)
        self.assertEqual(tags.tags["owner"].name, "owner")
        self.assertEqual(tags.tags["owner"].value, "esc-sdk-test")

        self.client.update_environment_tag(self.orgName, PROJECT_NAME, self.envName, "owner", "esc-sdk-test", "new-owner", "esc-sdk-test-updated")

        tag = self.client.get_environment_tag(self.orgName, PROJECT_NAME, self.envName, "new-owner")
        self.assertEqual(tag.name, "new-owner")
        self.assertEqual(tag.value, "esc-sdk-test-updated")

        self.client.delete_environment_tag(self.orgName, PROJECT_NAME, self.envName, "new-owner")
        tags = self.client.list_environment_tags(self.orgName, PROJECT_NAME, self.envName)
        self.assertIsNotNone(tags)
        self.assertEqual(len(tags.tags), 0)

    def test_check_environment_valid(self):
        envDef = esc.EnvironmentDefinition(values=esc.EnvironmentDefinitionValues(additional_properties={"foo": "bar"}))

        diags = self.client.check_environment(self.orgName, envDef)
        self.assertNotEqual(diags, None)
        self.assertEqual(diags.diagnostics, None)

    def test_check_environment_invalid(self):
        envDef = esc.EnvironmentDefinition(
            values=esc.EnvironmentDefinitionValues(
                additional_properties={"foo": "bar"}, pulumi_config={"foo": "${bad_ref}"}))
        diags = self.client.check_environment(self.orgName, envDef)
        self.assertNotEqual(diags, None)
        self.assertNotEqual(diags.diagnostics, None)
        self.assertEqual(len(diags.diagnostics), 1)
        self.assertEqual(diags.diagnostics[0].summary, "unknown property \"bad_ref\"")

    def assertEnvDef(self, env):
        self.assertListEqual(env.imports, [f'{PROJECT_NAME}/{self.baseEnvName}'])
        self.assertEqual(env.values.additional_properties["foo"], "bar")
        self.assertEqual(env.values.additional_properties["my_array"], [1, 2, 3])
        self.assertIsNotNone(env.values.pulumi_config)
        self.assertEqual(env.values.pulumi_config["foo"], "${foo}")
        self.assertIsNotNone(env.values.environment_variables)
        self.assertEqual(env.values.environment_variables["FOO"], "${foo}")

    def assertFindEnv(self, envs):
        self.assertIsNotNone(envs)
        self.assertGreater(len(envs.environments), 0)
        for env in envs.environments:
            if env.name == self.envName:
                return

        self.fail("Environment {envName} not found".format(envName=self.envName))

    def remove_all_python_test_envs(self) -> None:
        continuationToken = None
        while True:
            envs = self.client.list_environments(self.orgName, continuationToken)
            for env in envs.environments:
                if env.project == PROJECT_NAME and env.name.startswith(ENV_PREFIX):
                    self.client.delete_environment(self.orgName, PROJECT_NAME, env.name)

            continuationToken = envs.next_token
            if continuationToken is None or continuationToken == "":
                break


if __name__ == '__main__':
    unittest.main()
