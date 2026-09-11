// Copyright 2026, Pulumi Corporation.  All rights reserved.

package cloudsdkcontract

import (
	"encoding/json"
	"net/http"
	"testing"
	"time"

	"github.com/pulumi/esc-sdk/sdk/go/internal/httpstub"
	"github.com/pulumi/pulumi-cloud-sdk/go/apitype"
	"github.com/stretchr/testify/require"
)

func checkYAML(t *testing.T, fixtureName string) (*apitype.EnvironmentResponse, error) {
	t.Helper()
	client, _ := newClient(t, httpstub.JSON(http.StatusOK, httpstub.Fixture(t, fixtureName)))
	return client.CheckYAML_esc(ctx(), org, nil, environmentYAML)
}

func fixtureSchemaJSON(t *testing.T, fixtureName string) string {
	t.Helper()
	var envelope struct {
		Schema json.RawMessage `json:"schema"`
	}
	require.NoError(t, json.Unmarshal([]byte(httpstub.Fixture(t, fixtureName)), &envelope))
	return string(envelope.Schema)
}

func TestCheckResponseWithObjectSchemas(t *testing.T) {
	t.Parallel()
	result, err := checkYAML(t, "check_contract_schema.json")
	require.NoError(t, err)

	app := result.Properties["app"]
	require.Equal(t, "<yaml>", app.Trace.Def.Environment)
	require.Equal(t, apitype.EscPos{Line: 2, Column: 3, Byte: 10}, app.Trace.Def.Begin)
	nested, ok := app.Value.(map[string]any)
	require.True(t, ok, "nested values arrive as a JSON object tree")
	require.Equal(t, "demo", nested["name"].(map[string]any)["value"])
	require.Equal(t, float64(3), nested["replicas"].(map[string]any)["value"])

	password := result.Properties["password"]
	require.True(t, password.Secret)
	require.Equal(t, "[secret]", password.Value)

	appVariable := result.Properties["environmentVariables"].Value.(map[string]any)["APP_NAME"].(map[string]any)
	require.Equal(t, "demo", appVariable["trace"].(map[string]any)["base"].(map[string]any)["value"])

	require.Equal(t, "object", result.Schema.Type)
	replicas := result.Schema.Properties["app"].Properties["replicas"]
	require.Equal(t, "number", replicas.Type)
	require.Equal(t, float64(3), replicas.Const)
	require.Equal(t, json.Number("1"), replicas.Minimum)
	require.Equal(t, json.Number("10.5"), replicas.Maximum)
	require.Equal(t, "string", result.Schema.Properties["environmentVariables"].AdditionalProperties.Type)
	require.True(t, result.Schema.Properties["password"].Secret)

	require.Equal(t, "demo", result.Exprs["app"].Object["name"].Literal)
	require.Equal(t, "fn::secret", result.Exprs["password"].Builtin.Name)
	require.Len(t, result.Exprs["environmentVariables"].Object["APP_NAME"].Interpolate, 1)

	require.Equal(t, "alice", result.ExecutionContext.Properties["pulumi"].Value.(map[string]any)["user"].(map[string]any)["value"].(map[string]any)["login"].(map[string]any)["value"])
	require.Equal(t, map[string]int{"fn::secret": 1}, result.EnvironmentFunctionSummary.FuncCounts)
	require.Empty(t, result.Diagnostics)
}

func TestSchemaRoundTripsPreserveMeaning(t *testing.T) {
	t.Parallel()
	cases := []struct {
		name    string
		fixture string
	}{
		{"object schemas", "check_contract_schema.json"},
		{"array literals encode items as false", "check_literal_schemas.json"},
		{"top-level schema encoded as true", "check_boolean_schema.json"},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			result, err := checkYAML(t, tc.fixture)
			require.NoError(t, err, "boolean-form schemas must decode, see %s", "https://github.com/pulumi/pulumi-cloud-sdk/issues/6")
			require.NotNil(t, result.Schema)

			encoded, err := json.Marshal(result.Schema)
			require.NoError(t, err)
			require.JSONEq(t, fixtureSchemaJSON(t, tc.fixture), string(encoded))
		})
	}
}

func TestCheckResponseFromAnOlderService(t *testing.T) {
	t.Parallel()
	result, err := checkYAML(t, "check_legacy_service.json")
	require.NoError(t, err)

	require.Equal(t, "hello", result.Properties["greeting"].Value)
	require.True(t, result.Properties["token"].Unknown)
	require.Nil(t, result.ExecutionContext)
	require.Empty(t, result.EnvironmentFunctionSummary.FuncCounts)

	require.Len(t, result.Diagnostics, 2)
	require.Equal(t, `unknown property "bad_ref"`, result.Diagnostics[0].Summary)
	require.Equal(t, 6, result.Diagnostics[0].Range.Begin.Line)
	require.Empty(t, result.Diagnostics[0].Path)
	require.Nil(t, result.Diagnostics[1].Range)
}

func TestValueFlagsAndTypedTraces(t *testing.T) {
	t.Parallel()
	var value apitype.EscValue
	require.NoError(t, json.Unmarshal([]byte(`{
		"value": "demo",
		"secret": true,
		"unknown": true,
		"trace": {
			"def": {"environment": "payments/prod", "begin": {"line": 8, "column": 15, "byte": 141}, "end": {"line": 8, "column": 26, "byte": 152}},
			"base": {"value": "demo", "trace": {"def": {"environment": "shared/base", "begin": {"line": 3, "column": 11, "byte": 26}, "end": {"line": 3, "column": 15, "byte": 30}}}}
		}
	}`), &value))

	require.True(t, value.Secret)
	require.True(t, value.Unknown)
	require.Equal(t, "payments/prod", value.Trace.Def.Environment)
	require.NotNil(t, value.Trace.Base)
	require.Equal(t, "demo", value.Trace.Base.Value)
	require.Equal(t, "shared/base", value.Trace.Base.Trace.Def.Environment)
	require.Nil(t, value.Trace.Base.Trace.Base)
}

func TestAbsentOptionalFieldsDecodeToZeroValues(t *testing.T) {
	t.Parallel()

	t.Run("check", func(t *testing.T) {
		t.Parallel()
		client, _ := newClient(t, httpstub.JSON(http.StatusOK, `{}`))

		result, err := client.CheckYAML_esc(ctx(), org, nil, environmentYAML)
		require.NoError(t, err)
		require.Nil(t, result.EscEnvironment, "the embedded environment stays nil, so promoted fields must not be read before checking it, see %s", "https://github.com/pulumi/pulumi-cloud-sdk/issues/6")
		require.Nil(t, result.Diagnostics)
	})

	t.Run("update", func(t *testing.T) {
		t.Parallel()
		client, _ := newClient(t, httpstub.JSON(http.StatusOK, `{}`))

		result, err := client.UpdateEnvironment_esc_environments(ctx(), org, project, env, environmentYAML)
		require.NoError(t, err)
		require.Nil(t, result.Diagnostics)
	})

	t.Run("open", func(t *testing.T) {
		t.Parallel()
		client, _ := newClient(t, httpstub.JSON(http.StatusOK, `{"id":"sess-1"}`))

		result, err := client.OpenEnvironment_esc_environments(ctx(), org, project, env, nil)
		require.NoError(t, err)
		require.Equal(t, "sess-1", result.ID)
		require.Nil(t, result.Diagnostics)
	})
}

func TestReadOpenEnvironmentResponses(t *testing.T) {
	t.Parallel()

	t.Run("whole environment", func(t *testing.T) {
		t.Parallel()
		client, _ := newClient(t, httpstub.JSON(http.StatusOK, httpstub.Fixture(t, "open_read_full.json")))

		result, err := client.ReadOpenEnvironment_esc_environments(ctx(), org, project, env, "sess-1", nil)
		require.NoError(t, err)

		environment := (*result).(map[string]any)
		properties := environment["properties"].(map[string]any)
		password := properties["password"].(map[string]any)
		require.Equal(t, "hunter2", password["value"])
		require.Equal(t, true, password["secret"])
		region := properties["pulumiConfig"].(map[string]any)["value"].(map[string]any)["aws:region"].(map[string]any)
		require.Equal(t, "eu-west-1", region["value"])
		require.Equal(t, false, environment["schema"].(map[string]any)["properties"].(map[string]any)["regions"].(map[string]any)["items"])
	})

	t.Run("single property", func(t *testing.T) {
		t.Parallel()
		client, _ := newClient(t, httpstub.JSON(http.StatusOK, httpstub.Fixture(t, "open_read_property.json")))

		result, err := client.ReadOpenEnvironment_esc_environments(ctx(), org, project, env, "sess-1", ptr("pulumiConfig"))
		require.NoError(t, err)

		property := (*result).(map[string]any)
		require.Equal(t, "payments/prod", property["trace"].(map[string]any)["def"].(map[string]any)["environment"])
		require.Equal(t, "eu-west-1", property["value"].(map[string]any)["aws:region"].(map[string]any)["value"])
	})
}

func TestRevisionsResponse(t *testing.T) {
	t.Parallel()
	client, _ := newClient(t, httpstub.JSON(http.StatusOK, httpstub.Fixture(t, "revisions.json")))

	result, err := client.ListEnvironmentRevisions_esc_environments(ctx(), org, project, env, nil, nil)
	require.NoError(t, err)

	revisions := *result
	require.Len(t, revisions, 3)
	require.Equal(t, 3, revisions[0].Number)
	require.Equal(t, time.Date(2026, 9, 1, 10, 0, 0, 0, time.UTC), revisions[0].Created)
	require.Equal(t, []string{"latest", "stable"}, revisions[0].Tags)
	require.Equal(t, "cr-42", revisions[0].SourceChangeRequest.ID)
	require.Nil(t, revisions[0].Retracted)
	require.Equal(t, 3, revisions[1].Retracted.Replacement)
	require.Equal(t, "leaked secret", revisions[1].Retracted.Reason)
	require.Empty(t, revisions[2].CreatorLogin)
	require.Nil(t, revisions[2].Tags)
}

func TestListResponses(t *testing.T) {
	t.Parallel()

	t.Run("environments page", func(t *testing.T) {
		t.Parallel()
		client, _ := newClient(t, httpstub.JSON(http.StatusOK, httpstub.Fixture(t, "environments_page.json")))

		result, err := client.ListOrgEnvironments_esc(ctx(), org, nil, nil, nil, nil)
		require.NoError(t, err)
		require.Len(t, result.Environments, 2)
		require.Equal(t, "payments", result.Environments[0].Project)
		require.Equal(t, "prod", result.Environments[0].Name)
		require.Equal(t, "alice", result.Environments[0].OwnedBy.GitHubLogin)
		require.Equal(t, map[string]string{"team": "platform"}, result.Environments[0].Tags)
		require.True(t, result.Environments[0].Settings.DeletionProtected)
		require.Equal(t, 4, result.Environments[0].ReferrerMetadata.StackReferrerCount)
		require.NotNil(t, result.NextToken)
		require.Equal(t, "eyJvZmZzZXQiOjJ9", *result.NextToken)
	})

	t.Run("environments from an older service", func(t *testing.T) {
		t.Parallel()
		client, _ := newClient(t, httpstub.JSON(http.StatusOK, httpstub.Fixture(t, "environments_legacy_service.json")))

		result, err := client.ListOrgEnvironments_esc(ctx(), org, nil, nil, nil, nil)
		require.NoError(t, err)
		require.Len(t, result.Environments, 1)
		require.Empty(t, result.Environments[0].Project)
		require.Nil(t, result.NextToken)
	})

	t.Run("environment tags page", func(t *testing.T) {
		t.Parallel()
		client, _ := newClient(t, httpstub.JSON(http.StatusOK, httpstub.Fixture(t, "environment_tags_page.json")))

		result, err := client.ListEnvironmentTags_esc_environments(ctx(), org, project, env, nil, nil)
		require.NoError(t, err)
		require.Len(t, result.Tags, 2)
		require.Equal(t, "platform", result.Tags["team"].Value)
		require.Equal(t, time.Date(2026, 8, 15, 10, 0, 0, 0, time.UTC), result.Tags["team"].Modified)
		require.Equal(t, "2", result.NextToken)
	})

	t.Run("revision tags page", func(t *testing.T) {
		t.Parallel()
		client, _ := newClient(t, httpstub.JSON(http.StatusOK, httpstub.Fixture(t, "revision_tags_page.json")))

		result, err := client.ListRevisionTags_esc_environments_versions(ctx(), org, project, env, nil, nil)
		require.NoError(t, err)
		require.Len(t, result.Tags, 2)
		require.Equal(t, "latest", result.Tags[0].Name)
		require.Equal(t, 3, result.Tags[0].Revision)
		require.Empty(t, result.Tags[1].EditorLogin)
		require.Equal(t, "stable", result.NextToken)
	})
}
