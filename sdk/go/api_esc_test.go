// Copyright 2024, Pulumi Corporation.  All rights reserved.

/*
ESC (Environments, Secrets, Config) API

Testing EscAPIService

*/

// Code generated by OpenAPI Generator (https://openapi-generator.tech);

package esc_sdk

import (
	"context"
	"os"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
)

const ENV_PREFIX = "sdk-go-test-"

func Test_EscClient(t *testing.T) {
	accessToken := os.Getenv("PULUMI_ACCESS_TOKEN")
	require.NotEmpty(t, accessToken, "PULUMI_ACCESS_TOKEN must be set")
	orgName := os.Getenv("PULUMI_ORG")
	require.NotEmpty(t, orgName, "PULUMI_ORG must be set")
	configuration := NewConfiguration()
	apiClient := NewClient(configuration)
	auth := context.WithValue(
		context.Background(),
		ContextAPIKeys,
		map[string]APIKey{
			"Authorization": {Key: accessToken, Prefix: "token"},
		},
	)

	removeAllGoTestEnvs(t, apiClient, auth, orgName)

	baseEnvName := ENV_PREFIX + "base-" + time.Now().Format("20060102150405")
	err := apiClient.CreateEnvironment(auth, orgName, baseEnvName)
	require.Nil(t, err)
	t.Cleanup(func() {
		err := apiClient.DeleteEnvironment(auth, orgName, baseEnvName)
		require.Nil(t, err)
	})

	baseEnv := &EnvironmentDefinition{
		Values: &EnvironmentDefinitionValues{
			AdditionalProperties: map[string]any{
				"base": baseEnvName,
			},
		},
	}

	_, err = apiClient.UpdateEnvironment(auth, orgName, baseEnvName, baseEnv)
	require.Nil(t, err)

	t.Run("should create, list, update, get, decrypt, open and delete an environment", func(t *testing.T) {
		envName := ENV_PREFIX + time.Now().Format("20060102150405")
		err := apiClient.CreateEnvironment(auth, orgName, envName)
		require.Nil(t, err)
		t.Cleanup(func() {
			err := apiClient.DeleteEnvironment(auth, orgName, envName)
			require.Nil(t, err)
		})

		envs, err := apiClient.ListEnvironments(auth, orgName, nil)
		require.Nil(t, err)

		requireFindEnvironment(t, envs, envName)

		yaml := "imports:\n  - " + baseEnvName + "\n" + `
values:
  foo: bar
  my_secret:
    fn::secret: "shh! don't tell anyone"
  my_array: [1, 2, 3]
  pulumiConfig:
    foo: ${foo}
  environmentVariables:
    FOO: ${foo}
`

		diags, err := apiClient.UpdateEnvironmentYaml(auth, orgName, envName, yaml)
		require.Nil(t, err)
		require.NotNil(t, diags)
		require.Len(t, diags.Diagnostics, 0)

		env, newYaml, err := apiClient.GetEnvironment(auth, orgName, envName)
		require.Nil(t, err)
		require.NotNil(t, newYaml)

		assertEnvDef(t, env, baseEnvName)

		require.NotNil(t, env.Values.AdditionalProperties["my_secret"].(map[string]any))

		decryptEnv, _, err := apiClient.DecryptEnvironment(auth, orgName, envName)
		require.Nil(t, err)

		assertEnvDef(t, decryptEnv, baseEnvName)

		mySecret, ok := decryptEnv.Values.AdditionalProperties["my_secret"].(map[string]any)
		require.True(t, ok)
		require.Equal(t, "shh! don't tell anyone", mySecret["fn::secret"])

		_, values, err := apiClient.OpenAndReadEnvironment(auth, orgName, envName)
		require.Nil(t, err)

		require.Equal(t, baseEnvName, values["base"])
		require.Equal(t, "bar", values["foo"])
		require.Equal(t, []any{1.0, 2.0, 3.0}, values["my_array"])
		require.Equal(t, "shh! don't tell anyone", values["my_secret"])
		pulumiConfig, ok := values["pulumiConfig"].(map[string]any)
		require.True(t, ok)
		require.Equal(t, "bar", pulumiConfig["foo"])
		environmentVariables, ok := values["environmentVariables"].(map[string]any)
		require.True(t, ok)
		require.Equal(t, "bar", environmentVariables["FOO"])

		openInfo, err := apiClient.OpenEnvironment(auth, orgName, envName)
		require.Nil(t, err)

		v, value, err := apiClient.ReadEnvironmentProperty(auth, orgName, envName, openInfo.Id, "pulumiConfig.foo")
		require.Nil(t, err)
		require.Equal(t, "bar", v.Value)
		require.Equal(t, "bar", value)
	})

	t.Run("check environment definition valid", func(t *testing.T) {
		env := &EnvironmentDefinition{
			Values: &EnvironmentDefinitionValues{
				AdditionalProperties: map[string]any{
					"foo": "bar",
				},
			},
		}

		diags, err := apiClient.CheckEnvironment(auth, orgName, env)
		require.Nil(t, err)
		require.NotNil(t, diags)
		require.Len(t, diags.Diagnostics, 0)
	})

	t.Run("check environment yaml invalid", func(t *testing.T) {
		yaml := `
values:
  foo: bar
  pulumiConfig:
    foo: ${bad_ref}
`
		diags, err := apiClient.CheckEnvironmentYaml(auth, orgName, yaml)
		require.Error(t, err, "400 Bad Request")
		require.NotNil(t, diags)
		require.Len(t, diags.Diagnostics, 1)
		require.Equal(t, "unknown property \"bad_ref\"", diags.Diagnostics[0].Summary)

	})
}

func assertEnvDef(t *testing.T, env *EnvironmentDefinition, baseEnvName string) {
	require.Len(t, env.Imports, 1)
	require.Equal(t, baseEnvName, env.Imports[0])
	require.Equal(t, "bar", env.Values.AdditionalProperties["foo"])
	require.Equal(t, []any{1.0, 2.0, 3.0}, env.Values.AdditionalProperties["my_array"])

	require.Equal(t, "${foo}", env.Values.PulumiConfig["foo"])
	require.NotNil(t, env.Values.EnvironmentVariables)
	envVariables := *env.Values.EnvironmentVariables
	require.Equal(t, "${foo}", envVariables["FOO"])
}

func requireFindEnvironment(t *testing.T, envs *OrgEnvironments, envName string) {
	found := false
	for _, env := range envs.Environments {
		if env.Name == envName {
			found = true
		}
	}

	require.True(t, found)
}

func removeAllGoTestEnvs(t *testing.T, apiClient *EscClient, auth context.Context, orgName string) {
	var continuationToken *string
	for {
		envs, err := apiClient.ListEnvironments(auth, orgName, continuationToken)
		require.Nil(t, err)

		if len(envs.Environments) == 0 {
			break
		}
		for _, env := range envs.Environments {
			if strings.HasPrefix(env.Name, ENV_PREFIX) {
				err := apiClient.DeleteEnvironment(auth, orgName, env.Name)
				require.Nil(t, err)
			}
		}

		continuationToken = envs.NextToken
		if continuationToken == nil || *continuationToken == "" {
			break
		}
	}
}
