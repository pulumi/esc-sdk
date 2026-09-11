// Copyright 2026, Pulumi Corporation.  All rights reserved.

package esc_sdk

import (
	"context"
	"net/http"
	"net/url"
	"testing"

	"github.com/pulumi/esc-sdk/sdk/go/internal/httpstub"
	"github.com/pulumi/pulumi-cloud-sdk/go/apiclient"
	"github.com/stretchr/testify/require"
)

const (
	testOrg     = "acme"
	testProject = "payments"
	testEnv     = "prod"
)

const definitionYAML = `# payments/prod
imports:
  - shared/base
values:
  app:
    name: demo
    replicas: 3
  password:
    fn::secret: hunter2
  pulumiConfig:
    aws:region: ${app.region}
  environmentVariables:
    APP_NAME: ${app.name}
  files:
    APP_CONFIG: "name=demo"
`

func clientFor(t *testing.T, baseURL string) *EscClient {
	t.Helper()
	backend, err := url.Parse(baseURL)
	require.NoError(t, err)
	cfg, err := NewCustomBackendConfiguration(*backend)
	require.NoError(t, err)
	return NewClient(cfg)
}

func stubClient(t *testing.T, response httpstub.Response) (*EscClient, *httpstub.Recorder) {
	t.Helper()
	baseURL, rec := httpstub.Serve(t, response)
	return clientFor(t, baseURL), rec
}

func TestRequestsCarryAuthenticationAndIdentityHeaders(t *testing.T) {
	client, rec := stubClient(t, httpstub.YAML(definitionYAML))
	client.cfg.AddDefaultHeader("X-Custom", "yes")

	_, _, err := client.GetEnvironment(NewAuthContext("pul-secret"), testOrg, testProject, testEnv)
	require.NoError(t, err)

	header := rec.Only(t).Header
	require.Equal(t, "token pul-secret", header.Get("Authorization"))
	require.Equal(t, "esc-sdk", header.Get("User-Agent"))
	require.Equal(t, "esc-sdk", header.Get("X-Pulumi-Source"))
	require.Equal(t, "yes", header.Get("X-Custom"))
}

func TestUnauthenticatedContextSendsNoAuthorization(t *testing.T) {
	client, rec := stubClient(t, httpstub.YAML(definitionYAML))

	_, _, err := client.GetEnvironment(context.Background(), testOrg, testProject, testEnv)
	require.NoError(t, err)

	require.Empty(t, rec.Only(t).Header.Values("Authorization"))
}

func TestGetEnvironmentParsesTheDefinitionAndKeepsTheYAML(t *testing.T) {
	client, rec := stubClient(t, httpstub.YAML(definitionYAML))

	definition, raw, err := client.GetEnvironment(context.Background(), testOrg, testProject, testEnv)
	require.NoError(t, err)
	httpstub.RequireRequest(t, rec.Only(t), http.MethodGet, "/api/esc/environments/acme/payments/prod", nil)

	require.Equal(t, definitionYAML, raw)
	require.Equal(t, []string{"shared/base"}, definition.Imports)
	require.Equal(t, map[string]any{"name": "demo", "replicas": float64(3)}, definition.Values.AdditionalProperties["app"])
	require.Equal(t, map[string]any{"fn::secret": "hunter2"}, definition.Values.AdditionalProperties["password"])
	require.Equal(t, map[string]any{"aws:region": "${app.region}"}, definition.Values.PulumiConfig)
	require.Equal(t, map[string]string{"APP_NAME": "${app.name}"}, definition.Values.EnvironmentVariables)
	require.Equal(t, map[string]string{"APP_CONFIG": "name=demo"}, definition.Values.Files)
	require.NotContains(t, definition.Values.AdditionalProperties, "pulumiConfig")
}

func TestGetEnvironmentAtVersionAndDecryptUseTheirRoutes(t *testing.T) {
	cases := map[string]struct {
		path string
		call func(client *EscClient) error
	}{
		"at version": {"/api/esc/environments/acme/payments/prod/versions/3", func(c *EscClient) error {
			_, _, err := c.GetEnvironmentAtVersion(context.Background(), testOrg, testProject, testEnv, "3")
			return err
		}},
		"decrypt": {"/api/esc/environments/acme/payments/prod/decrypt", func(c *EscClient) error {
			_, _, err := c.DecryptEnvironment(context.Background(), testOrg, testProject, testEnv)
			return err
		}},
	}
	for name, tc := range cases {
		t.Run(name, func(t *testing.T) {
			client, rec := stubClient(t, httpstub.YAML(definitionYAML))
			require.NoError(t, tc.call(client))
			httpstub.RequireRequest(t, rec.Only(t), http.MethodGet, tc.path, nil)
		})
	}
}

func TestMarshalEnvironmentDefinitionRoundTrips(t *testing.T) {
	definition := &EnvironmentDefinition{
		Imports: []string{"shared/base"},
		Values: &EnvironmentDefinitionValues{
			PulumiConfig:         map[string]any{"aws:region": "eu-west-1"},
			EnvironmentVariables: map[string]string{"APP_NAME": "demo"},
			AdditionalProperties: map[string]any{"app": map[string]any{"name": "demo"}},
		},
	}

	encoded, err := MarshalEnvironmentDefinition(definition)
	require.NoError(t, err)
	require.Equal(t, "imports:\n- shared/base\nvalues:\n  app:\n    name: demo\n  environmentVariables:\n    APP_NAME: demo\n  pulumiConfig:\n    aws:region: eu-west-1\n", encoded)

	client, rec := stubClient(t, httpstub.JSON(http.StatusOK, `{}`))
	_, err = client.UpdateEnvironment(context.Background(), testOrg, testProject, testEnv, definition)
	require.NoError(t, err)
	req := rec.Only(t)
	httpstub.RequireRequest(t, req, http.MethodPatch, "/api/esc/environments/acme/payments/prod", nil)
	require.Equal(t, "application/x-yaml", req.Header.Get("Content-Type"))
	require.Equal(t, encoded, req.Body)
}

func TestCheckEnvironmentYaml(t *testing.T) {
	t.Run("returns the evaluated environment", func(t *testing.T) {
		client, rec := stubClient(t, httpstub.JSON(http.StatusOK, httpstub.Fixture(t, "check_contract_schema.json")))

		result, err := client.CheckEnvironmentYaml(context.Background(), testOrg, definitionYAML)
		require.NoError(t, err)
		req := rec.Only(t)
		httpstub.RequireRequest(t, req, http.MethodPost, "/api/esc/environments/acme/yaml/check", nil)
		require.Equal(t, definitionYAML, req.Body)

		require.Empty(t, result.Diagnostics)
		require.Equal(t, "[secret]", result.Properties["password"].Value)
		require.True(t, result.Properties["password"].Secret)
		require.Equal(t, "object", result.Schema.Type)
	})

	t.Run("decodes schemas the evaluator encodes as booleans", func(t *testing.T) {
		client, _ := stubClient(t, httpstub.JSON(http.StatusOK, httpstub.Fixture(t, "check_boolean_schema.json")))

		result, err := client.CheckEnvironmentYaml(context.Background(), testOrg, definitionYAML)
		require.NoError(t, err, "boolean-form schemas must decode, see %s", "https://github.com/pulumi/pulumi-cloud-sdk/issues/6")
		require.Equal(t, "eu-west-1", result.Properties["region"].Value)
	})

	t.Run("an empty response is safe to read", func(t *testing.T) {
		client, _ := stubClient(t, httpstub.JSON(http.StatusOK, `{}`))

		result, err := client.CheckEnvironmentYaml(context.Background(), testOrg, definitionYAML)
		require.NoError(t, err)
		require.Empty(t, result.Properties)
		require.Empty(t, result.Diagnostics)
	})

	t.Run("returns diagnostics with the error on a rejected definition", func(t *testing.T) {
		client, _ := stubClient(t, httpstub.JSON(http.StatusBadRequest, httpstub.Fixture(t, "check_diagnostics_400.json")))

		result, err := client.CheckEnvironmentYaml(context.Background(), testOrg, definitionYAML)
		var apiErr *apiclient.APIError
		require.ErrorAs(t, err, &apiErr)
		require.Equal(t, http.StatusBadRequest, apiErr.HTTPStatusCode())
		require.NotNil(t, result)
		require.Len(t, result.Diagnostics, 1)
		require.Equal(t, `unknown property "bad_ref"`, result.Diagnostics[0].Summary)
		require.Equal(t, "values.ref", result.Diagnostics[0].Path)
		require.Equal(t, 3, result.Diagnostics[0].Range.Begin.Line)
		require.Empty(t, result.Properties)
	})

	t.Run("returns no result on other errors", func(t *testing.T) {
		client, _ := stubClient(t, httpstub.JSON(http.StatusUnauthorized, `{"code":401,"message":"unauthorized"}`))

		result, err := client.CheckEnvironmentYaml(context.Background(), testOrg, definitionYAML)
		require.Error(t, err)
		require.Nil(t, result)
	})
}

func TestUpdateEnvironmentYaml(t *testing.T) {
	t.Run("returns the diagnostics of an accepted definition", func(t *testing.T) {
		client, _ := stubClient(t, httpstub.JSON(http.StatusOK, `{"diagnostics":[{"summary":"unused import","severity":"warning"}]}`))

		diagnostics, err := client.UpdateEnvironmentYaml(context.Background(), testOrg, testProject, testEnv, definitionYAML)
		require.NoError(t, err)
		require.Len(t, diagnostics.Diagnostics, 1)
		require.Equal(t, "unused import", diagnostics.Diagnostics[0].Summary)
	})

	t.Run("returns diagnostics with the error on a rejected definition", func(t *testing.T) {
		client, _ := stubClient(t, httpstub.JSON(http.StatusBadRequest, `{"diagnostics":[{"summary":"unknown property \"bad_ref\""}]}`))

		diagnostics, err := client.UpdateEnvironmentYaml(context.Background(), testOrg, testProject, testEnv, definitionYAML)
		require.Error(t, err)
		require.NotNil(t, diagnostics)
		require.Equal(t, `unknown property "bad_ref"`, diagnostics.Diagnostics[0].Summary)
	})
}

func TestOpenAndReadEnvironment(t *testing.T) {
	baseURL, rec := httpstub.ServeRoutes(t, map[string]httpstub.Response{
		"POST /api/esc/environments/acme/payments/prod/open":       httpstub.JSON(http.StatusOK, `{"id":"sess-1"}`),
		"GET /api/esc/environments/acme/payments/prod/open/sess-1": httpstub.JSON(http.StatusOK, httpstub.Fixture(t, "open_read_full.json")),
	})
	client := clientFor(t, baseURL)

	env, values, err := client.OpenAndReadEnvironment(context.Background(), testOrg, testProject, testEnv)
	require.NoError(t, err, "array literals carry items: false, see %s", "https://github.com/pulumi/pulumi-cloud-sdk/issues/6")
	require.Len(t, rec.All(), 2)

	require.Equal(t, map[string]any{
		"app":          map[string]any{"name": "demo", "replicas": float64(3)},
		"password":     "hunter2",
		"pulumiConfig": map[string]any{"aws:region": "eu-west-1"},
		"regions":      []any{"eu-west-1", "us-east-1"},
	}, values)

	require.True(t, env.Properties["password"].Secret)
	require.Equal(t, "payments/prod", env.Properties["app"].Trace.Def.Environment)
	require.Equal(t, "object", env.Schema.Type)
}

func TestOpenAndReadEnvironmentAtVersion(t *testing.T) {
	baseURL, _ := httpstub.ServeRoutes(t, map[string]httpstub.Response{
		"POST /api/esc/environments/acme/payments/prod/versions/stable/open": httpstub.JSON(http.StatusOK, `{"id":"sess-2"}`),
		"GET /api/esc/environments/acme/payments/prod/open/sess-2":           httpstub.JSON(http.StatusOK, `{"properties":{"region":{"value":"eu-west-1","trace":{"def":{"environment":"payments/prod","begin":{"line":1,"column":1,"byte":0},"end":{"line":1,"column":1,"byte":0}}}}}}`),
	})
	client := clientFor(t, baseURL)

	_, values, err := client.OpenAndReadEnvironmentAtVersion(context.Background(), testOrg, testProject, testEnv, "stable")
	require.NoError(t, err)
	require.Equal(t, map[string]any{"region": "eu-west-1"}, values)
}

func TestReadOpenEnvironmentWithoutProperties(t *testing.T) {
	client, _ := stubClient(t, httpstub.JSON(http.StatusOK, `{}`))

	env, values, err := client.ReadOpenEnvironment(context.Background(), testOrg, testProject, testEnv, "sess-1")
	require.NoError(t, err)
	require.NotNil(t, env)
	require.Nil(t, values)
}

func TestReadEnvironmentProperty(t *testing.T) {
	client, rec := stubClient(t, httpstub.JSON(http.StatusOK, httpstub.Fixture(t, "open_read_property.json")))

	property, value, err := client.ReadEnvironmentProperty(context.Background(), testOrg, testProject, testEnv, "sess-1", "pulumiConfig")
	require.NoError(t, err)
	httpstub.RequireRequest(t, rec.Only(t), http.MethodGet, "/api/esc/environments/acme/payments/prod/open/sess-1", url.Values{"property": {"pulumiConfig"}})

	require.Equal(t, map[string]any{"aws:region": "eu-west-1"}, value)
	require.Equal(t, "payments/prod", property.Trace.Def.Environment)
	require.Equal(t, 7, property.Trace.Def.Begin.Line)
}

func TestListEnvironmentsForwardsTheContinuationToken(t *testing.T) {
	client, rec := stubClient(t, httpstub.JSON(http.StatusOK, httpstub.Fixture(t, "environments_page.json")))
	token := "page-2"

	page, err := client.ListEnvironments(context.Background(), testOrg, &token)
	require.NoError(t, err)
	httpstub.RequireRequest(t, rec.Only(t), http.MethodGet, "/api/esc/environments/acme", url.Values{"continuationToken": {"page-2"}})
	require.Len(t, page.Environments, 2)
	require.Equal(t, "eyJvZmZzZXQiOjJ9", *page.NextToken)
}

func TestRevisionAndTagCalls(t *testing.T) {
	t.Run("paginated revisions", func(t *testing.T) {
		client, rec := stubClient(t, httpstub.JSON(http.StatusOK, httpstub.Fixture(t, "revisions.json")))

		revisions, err := client.ListEnvironmentRevisionsPaginated(context.Background(), testOrg, testProject, testEnv, 3, 2)
		require.NoError(t, err)
		httpstub.RequireRequest(t, rec.Only(t), http.MethodGet, "/api/esc/environments/acme/payments/prod/versions", url.Values{"before": {"3"}, "count": {"2"}})
		require.Len(t, revisions, 3)
		require.Equal(t, 3, revisions[1].Retracted.Replacement)
	})

	t.Run("create revision tag", func(t *testing.T) {
		client, rec := stubClient(t, httpstub.Empty(http.StatusNoContent))

		require.NoError(t, client.CreateEnvironmentRevisionTag(context.Background(), testOrg, testProject, testEnv, "stable", 3))
		req := rec.Only(t)
		httpstub.RequireRequest(t, req, http.MethodPost, "/api/esc/environments/acme/payments/prod/versions/tags", nil)
		require.JSONEq(t, `{"name":"stable","revision":3}`, req.Body)
	})

	t.Run("update revision tag", func(t *testing.T) {
		client, rec := stubClient(t, httpstub.Empty(http.StatusNoContent))

		require.NoError(t, client.UpdateEnvironmentRevisionTag(context.Background(), testOrg, testProject, testEnv, "stable", 4))
		req := rec.Only(t)
		httpstub.RequireRequest(t, req, http.MethodPatch, "/api/esc/environments/acme/payments/prod/versions/tags/stable", nil)
		require.JSONEq(t, `{"revision":4}`, req.Body)
	})

	t.Run("paginated environment tags", func(t *testing.T) {
		client, rec := stubClient(t, httpstub.JSON(http.StatusOK, httpstub.Fixture(t, "environment_tags_page.json")))

		tags, err := client.ListEnvironmentTagsPaginated(context.Background(), testOrg, testProject, testEnv, 2, 10)
		require.NoError(t, err)
		httpstub.RequireRequest(t, rec.Only(t), http.MethodGet, "/api/esc/environments/acme/payments/prod/tags", url.Values{"after": {"2"}, "count": {"10"}})
		require.Equal(t, "platform", tags.Tags["team"].Value)
	})

	t.Run("update environment tag", func(t *testing.T) {
		client, rec := stubClient(t, httpstub.JSON(http.StatusOK, `{"name":"owner","value":"payments","created":"2026-08-01T10:00:00Z","modified":"2026-08-01T10:00:00Z","editorLogin":"alice","editorName":"Alice"}`))

		tag, err := client.UpdateEnvironmentTag(context.Background(), testOrg, testProject, testEnv, "team", "platform", "owner", "payments")
		require.NoError(t, err)
		require.Equal(t, "owner", tag.Name)
		req := rec.Only(t)
		httpstub.RequireRequest(t, req, http.MethodPatch, "/api/esc/environments/acme/payments/prod/tags/team", nil)
		require.JSONEq(t, `{"currentTag":{"value":"platform"},"newTag":{"name":"owner","value":"payments"}}`, req.Body)
	})
}

func TestCloneEnvironmentSendsTheOptions(t *testing.T) {
	client, rec := stubClient(t, httpstub.Empty(http.StatusNoContent))

	err := client.CloneEnvironment(context.Background(), testOrg, testProject, testEnv, "staging", "prod-copy", &CloneEnvironmentOptions{PreserveHistory: true, PreserveRevisionTags: true})
	require.NoError(t, err)
	req := rec.Only(t)
	httpstub.RequireRequest(t, req, http.MethodPost, "/api/esc/environments/acme/payments/prod/clone", nil)
	require.JSONEq(t, `{"project":"staging","name":"prod-copy","preserveHistory":true,"preserveRevisionTags":true}`, req.Body)

	require.NoError(t, client.CloneEnvironment(context.Background(), testOrg, testProject, testEnv, "staging", "prod-copy", nil))
}

func TestCancelledContextAbortsTheCall(t *testing.T) {
	client, _ := stubClient(t, httpstub.YAML(definitionYAML))
	ctx, cancel := context.WithCancel(NewAuthContext("pul-secret"))
	cancel()

	_, _, err := client.GetEnvironment(ctx, testOrg, testProject, testEnv)
	require.ErrorIs(t, err, context.Canceled)
}
