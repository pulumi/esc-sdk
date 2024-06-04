
# Copyright 2024, Pulumi Corporation.  All rights reserved.

from pulumi_esc_sdk.exceptions import ApiException
import pulumi_esc_sdk.models as models
import pulumi_esc_sdk.api as api
import pulumi_esc_sdk.configuration as configuration
import pulumi_esc_sdk.api_client as api_client
from pydantic import StrictBytes, StrictInt
from typing import Mapping, Any, List
import inspect
import yaml


class EscClient:
    esc_api: api.EscApi = None
    """EscClient

    :param esc_api: EscApi (required)
    """
    def __init__(self, configuration: configuration.Configuration) -> None:
        self.esc_api = api.EscApi(api_client.ApiClient(configuration))

    def list_environments(self, org_name: str, continuation_token: str = None) -> models.OrgEnvironments:
        return self.esc_api.list_environments(org_name, continuation_token)

    def get_environment(self, org_name: str, env_name: str) -> tuple[models.EnvironmentDefinition, StrictBytes]:
        response = self.esc_api.get_environment_with_http_info(org_name, env_name)
        return response.data, response.raw_data

    def get_environment_at_version(
            self, org_name: str, env_name: str, version: str) -> tuple[models.EnvironmentDefinition, StrictBytes]:
        response = self.esc_api.get_environment_at_version_with_http_info(org_name, env_name, version)
        return response.data, response.raw_data

    def open_environment(self, org_name: str, env_name: str) -> models.Environment:
        return self.esc_api.open_environment(org_name, env_name)

    def open_environment_at_version(self, org_name: str, env_name: str, version: str) -> models.Environment:
        return self.esc_api.open_environment_at_version(org_name, env_name, version)

    def read_open_environment(
            self, org_name: str, env_name: str, open_session_id: str) -> tuple[models.Environment, Mapping[str, Any], str]:
        response = self.esc_api.read_open_environment_with_http_info(org_name, env_name, open_session_id)
        values = convertEnvPropertiesToValues(response.data.properties)
        return response.data, values, response.raw_data.decode('utf-8')

    def open_and_read_environment(self, org_name: str, env_name: str) -> tuple[models.Environment, Mapping[str, any], str]:
        openEnv = self.open_environment(org_name, env_name)
        return self.read_open_environment(org_name, env_name, openEnv.id)

    def open_and_read_environment_at_version(
            self, org_name: str, env_name: str, version: str) -> tuple[models.Environment, Mapping[str, any], str]:
        openEnv = self.open_environment_at_version(org_name, env_name, version)
        return self.read_open_environment(org_name, env_name, openEnv.id)

    def read_open_environment_property(
            self, org_name: str, env_name: str, open_session_id: str, property_name: str) -> tuple[models.Value, Any]:
        v = self.esc_api.read_open_environment_property(org_name, env_name, open_session_id, property_name)
        return v, convertPropertyToValue(v.value)

    def create_environment(self, org_name: str, env_name: str) -> models.Environment:
        return self.esc_api.create_environment(org_name, env_name)

    def update_environment_yaml(self, org_name: str, env_name: str, yaml_body: str) -> models.EnvironmentDiagnostics:
        return self.esc_api.update_environment_yaml(org_name, env_name, yaml_body)

    def update_environment(self, org_name: str, env_name: str, env: models.EnvironmentDefinition) -> models.Environment:
        envData = env.to_dict()
        yaml_body = yaml.dump(envData)
        return self.update_environment_yaml(org_name, env_name, yaml_body)

    def delete_environment(self, org_name: str, env_name: str) -> None:
        self.esc_api.delete_environment(org_name, env_name)

    def check_environment_yaml(self, org_name: str, yaml_body: str) -> models.CheckEnvironment:
        try:
            response = self.esc_api.check_environment_yaml_with_http_info(org_name, yaml_body)
            return response.data
        except ApiException as e:
            return e.data

    def check_environment(self, org_name: str, env: models.EnvironmentDefinition) -> models.CheckEnvironment:
        yaml_body = yaml.safe_dump(env.to_dict())
        return self.check_environment_yaml(org_name, yaml_body)

    def decrypt_environment(self, org_name: str, env_name: str) -> tuple[models.EnvironmentDefinition, str]:
        response = self.esc_api.decrypt_environment_with_http_info(org_name, env_name)
        return response.data, response.raw_data.decode('utf-8')

    def list_environment_revisions(
            self,
            org_name: str,
            env_name: str,
            before: StrictInt | None = None,
            count: StrictInt | None = None
            ) -> List[models.EnvironmentRevision]:
        return self.esc_api.list_environment_revisions(org_name, env_name, before, count)

    def list_environment_revision_tags(
            self,
            org_name: str,
            env_name: str,
            before: str | None = None,
            count: StrictInt | None = None
            ) -> models.EnvironmentRevisionTags:
        return self.esc_api.list_environment_revision_tags(org_name, env_name, before, count)

    def create_environment_revision_tag(
            self, org_name: str, env_name: str, tag_name: str, revision: StrictInt) -> models.EnvironmentRevisionTag:
        update_tag = models.UpdateEnvironmentRevisionTag(revision=revision)
        return self.esc_api.create_environment_revision_tag(org_name, env_name, tag_name, update_tag)

    def update_environment_revision_tag(
            self, org_name: str, env_name: str, tag_name: str, revision: StrictInt) -> models.EnvironmentRevisionTag:
        update_tag = models.UpdateEnvironmentRevisionTag(revision=revision)
        return self.esc_api.update_environment_revision_tag(org_name, env_name, tag_name, update_tag)

    def get_environment_revision_tag(self, org_name: str, env_name: str, tag_name: str) -> models.EnvironmentRevisionTag:
        return self.esc_api.get_environment_revision_tag(org_name, env_name, tag_name)

    def delete_environment_revision_tag(self, org_name: str, env_name: str, tag_name: str) -> None:
        self.esc_api.delete_environment_revision_tag(org_name, env_name, tag_name)


def convertEnvPropertiesToValues(env: Mapping[str, models.Value]) -> Any:
    if env is None:
        return env

    values = {}
    for key in env:
        value = env[key]

        values[key] = convertPropertyToValue(value.value)

    return values


def convertPropertyToValue(property: Any) -> Any:
    if property is None:
        return property

    value = property
    if isinstance(property, dict) and "value" in property:
        value = convertPropertyToValue(property["value"])
        return value

    if value is None:
        return value

    if type(value) is list:
        result = []
        for item in value:
            result.append(convertPropertyToValue(item))
        return result

    if isObject(value):
        result = {}
        for key in value:
            result[key] = convertPropertyToValue(value[key])
        return result

    return value


def isObject(obj):
    return inspect.isclass(obj) or isinstance(obj, dict)
