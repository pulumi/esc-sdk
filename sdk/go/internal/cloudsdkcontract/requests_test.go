// Copyright 2026, Pulumi Corporation.  All rights reserved.

package cloudsdkcontract

import (
	"context"
	"net/http"
	"net/url"
	"testing"

	"github.com/pulumi/esc-sdk/sdk/go/internal/httpstub"
	"github.com/pulumi/pulumi-cloud-sdk/go/apiclient"
	"github.com/pulumi/pulumi-cloud-sdk/go/apitype"
	"github.com/stretchr/testify/require"
)

const environmentYAML = "# payments/prod\nvalues:\n  app:\n    name: demo\n    replicas: 3\n"

func TestCheckYAMLSendsRawYAML(t *testing.T) {
	t.Parallel()
	client, rec := newClient(t, httpstub.JSON(http.StatusOK, `{}`))

	_, err := client.CheckYAML_esc(ctx(), org, ptr(true), environmentYAML)
	require.NoError(t, err)

	req := rec.Only(t)
	httpstub.RequireRequest(t, req, http.MethodPost, "/api/esc/environments/acme/yaml/check", url.Values{"showSecrets": {"true"}})
	require.Equal(t, "application/x-yaml", req.Header.Get("Content-Type"))
	require.Equal(t, environmentYAML, req.Body)
}

func TestCheckYAMLOmitsUnsetQueryParameters(t *testing.T) {
	t.Parallel()
	client, rec := newClient(t, httpstub.JSON(http.StatusOK, `{}`))

	_, err := client.CheckYAML_esc(ctx(), org, nil, environmentYAML)
	require.NoError(t, err)

	httpstub.RequireRequest(t, rec.Only(t), http.MethodPost, "/api/esc/environments/acme/yaml/check", nil)
}

func TestUpdateEnvironmentSendsRawYAML(t *testing.T) {
	t.Parallel()
	client, rec := newClient(t, httpstub.JSON(http.StatusOK, `{}`))

	_, err := client.UpdateEnvironment_esc_environments(ctx(), org, project, env, environmentYAML)
	require.NoError(t, err)

	req := rec.Only(t)
	httpstub.RequireRequest(t, req, http.MethodPatch, "/api/esc/environments/acme/payments/prod", nil)
	require.Equal(t, "application/x-yaml", req.Header.Get("Content-Type"))
	require.Equal(t, environmentYAML, req.Body)
}

func TestEnvironmentReadsReturnTheRawYAMLBody(t *testing.T) {
	t.Parallel()
	cases := []struct {
		name string
		path string
		call func(client *apiclient.CloudClient) (*string, error)
	}{
		{"read", "/api/esc/environments/acme/payments/prod", func(c *apiclient.CloudClient) (*string, error) {
			return c.ReadEnvironment_esc_environments(ctx(), org, project, env)
		}},
		{"read at version", "/api/esc/environments/acme/payments/prod/versions/3", func(c *apiclient.CloudClient) (*string, error) {
			return c.ReadEnvironment_esc_environments_versions(ctx(), org, project, env, version)
		}},
		{"decrypt", "/api/esc/environments/acme/payments/prod/decrypt", func(c *apiclient.CloudClient) (*string, error) {
			return c.DecryptEnvironment_esc_environments(ctx(), org, project, env)
		}},
		{"decrypt at version", "/api/esc/environments/acme/payments/prod/versions/3/decrypt", func(c *apiclient.CloudClient) (*string, error) {
			return c.DecryptEnvironment_esc_environments_versions(ctx(), org, project, env, version)
		}},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			client, rec := newClient(t, httpstub.YAML(environmentYAML))

			body, err := tc.call(client)
			require.NoError(t, err)
			require.NotNil(t, body)
			require.Equal(t, environmentYAML, *body)

			httpstub.RequireRequest(t, rec.Only(t), http.MethodGet, tc.path, nil)
		})
	}
}

func TestOpenEnvironmentRequests(t *testing.T) {
	t.Parallel()

	t.Run("latest with duration", func(t *testing.T) {
		t.Parallel()
		client, rec := newClient(t, httpstub.JSON(http.StatusOK, `{"id":"sess-1"}`))

		open, err := client.OpenEnvironment_esc_environments(ctx(), org, project, env, ptr("2h"))
		require.NoError(t, err)
		require.Equal(t, "sess-1", open.ID)

		httpstub.RequireRequest(t, rec.Only(t), http.MethodPost, "/api/esc/environments/acme/payments/prod/open", url.Values{"duration": {"2h"}})
	})

	t.Run("at version without duration", func(t *testing.T) {
		t.Parallel()
		client, rec := newClient(t, httpstub.JSON(http.StatusOK, `{"id":"sess-2"}`))

		open, err := client.OpenEnvironment_esc_environments_versions(ctx(), org, project, env, version, nil)
		require.NoError(t, err)
		require.Equal(t, "sess-2", open.ID)

		httpstub.RequireRequest(t, rec.Only(t), http.MethodPost, "/api/esc/environments/acme/payments/prod/versions/3/open", nil)
	})
}

func TestReadOpenEnvironmentRequests(t *testing.T) {
	t.Parallel()

	t.Run("whole environment", func(t *testing.T) {
		t.Parallel()
		client, rec := newClient(t, httpstub.JSON(http.StatusOK, `{"properties":{}}`))

		_, err := client.ReadOpenEnvironment_esc_environments(ctx(), org, project, env, "sess-1", nil)
		require.NoError(t, err)

		httpstub.RequireRequest(t, rec.Only(t), http.MethodGet, "/api/esc/environments/acme/payments/prod/open/sess-1", nil)
	})

	t.Run("single property", func(t *testing.T) {
		t.Parallel()
		client, rec := newClient(t, httpstub.JSON(http.StatusOK, `{"value":"demo"}`))

		_, err := client.ReadOpenEnvironment_esc_environments(ctx(), org, project, env, "sess-1", ptr("app.name"))
		require.NoError(t, err)

		httpstub.RequireRequest(t, rec.Only(t), http.MethodGet, "/api/esc/environments/acme/payments/prod/open/sess-1", url.Values{"property": {"app.name"}})
	})
}

func TestPaginationQueries(t *testing.T) {
	t.Parallel()

	t.Run("organization environments", func(t *testing.T) {
		t.Parallel()
		client, rec := newClient(t, httpstub.JSON(http.StatusOK, `{"environments":[]}`))

		_, err := client.ListOrgEnvironments_esc(ctx(), org, ptr("page-2"), nil, ptr(50), nil)
		require.NoError(t, err)

		httpstub.RequireRequest(t, rec.Only(t), http.MethodGet, "/api/esc/environments/acme", url.Values{"continuationToken": {"page-2"}, "maxResults": {"50"}})
	})

	t.Run("revisions", func(t *testing.T) {
		t.Parallel()
		client, rec := newClient(t, httpstub.JSON(http.StatusOK, `[]`))

		_, err := client.ListEnvironmentRevisions_esc_environments(ctx(), org, project, env, ptr(3), ptr(2))
		require.NoError(t, err)

		httpstub.RequireRequest(t, rec.Only(t), http.MethodGet, "/api/esc/environments/acme/payments/prod/versions", url.Values{"before": {"3"}, "count": {"2"}})
	})

	t.Run("revision tags", func(t *testing.T) {
		t.Parallel()
		client, rec := newClient(t, httpstub.JSON(http.StatusOK, `{"tags":[],"nextToken":""}`))

		_, err := client.ListRevisionTags_esc_environments_versions(ctx(), org, project, env, ptr("stable"), ptr(10))
		require.NoError(t, err)

		httpstub.RequireRequest(t, rec.Only(t), http.MethodGet, "/api/esc/environments/acme/payments/prod/versions/tags", url.Values{"after": {"stable"}, "count": {"10"}})
	})

	t.Run("environment tags", func(t *testing.T) {
		t.Parallel()
		client, rec := newClient(t, httpstub.JSON(http.StatusOK, `{"tags":{},"nextToken":""}`))

		_, err := client.ListEnvironmentTags_esc_environments(ctx(), org, project, env, ptr(7), ptr(10))
		require.NoError(t, err)

		httpstub.RequireRequest(t, rec.Only(t), http.MethodGet, "/api/esc/environments/acme/payments/prod/tags", url.Values{"after": {"7"}, "count": {"10"}})
	})
}

func TestEnvironmentLifecycleRequests(t *testing.T) {
	t.Parallel()

	t.Run("create", func(t *testing.T) {
		t.Parallel()
		client, rec := newClient(t, httpstub.Empty(http.StatusOK))

		err := client.CreateEnvironment_esc_environments(ctx(), org, apitype.CreateEnvironmentRequest{Project: project, Name: env})
		require.NoError(t, err)

		req := rec.Only(t)
		httpstub.RequireRequest(t, req, http.MethodPost, "/api/esc/environments/acme", nil)
		require.Equal(t, "application/json", req.Header.Get("Content-Type"))
		require.JSONEq(t, `{"project":"payments","name":"prod"}`, req.Body)
	})

	t.Run("clone with every option", func(t *testing.T) {
		t.Parallel()
		client, rec := newClient(t, httpstub.Empty(http.StatusNoContent))

		err := client.CloneEnvironment(ctx(), org, project, env, apitype.CloneEnvironmentRequest{
			Project:                 "staging",
			Name:                    "prod-copy",
			Version:                 3,
			PreserveHistory:         true,
			PreserveAccess:          true,
			PreserveEnvironmentTags: true,
			PreserveRevisionTags:    true,
		})
		require.NoError(t, err)

		req := rec.Only(t)
		httpstub.RequireRequest(t, req, http.MethodPost, "/api/esc/environments/acme/payments/prod/clone", nil)
		require.JSONEq(t, `{"project":"staging","name":"prod-copy","version":3,"preserveHistory":true,"preserveAccess":true,"preserveEnvironmentTags":true,"preserveRevisionTags":true}`, req.Body)
	})

	t.Run("clone with defaults", func(t *testing.T) {
		t.Parallel()
		client, rec := newClient(t, httpstub.Empty(http.StatusNoContent))

		err := client.CloneEnvironment(ctx(), org, project, env, apitype.CloneEnvironmentRequest{Project: "staging", Name: "prod-copy"})
		require.NoError(t, err)

		require.JSONEq(t, `{"project":"staging","name":"prod-copy"}`, rec.Only(t).Body)
	})

	t.Run("delete", func(t *testing.T) {
		t.Parallel()
		client, rec := newClient(t, httpstub.Empty(http.StatusNoContent))

		err := client.DeleteEnvironment_esc_environments(ctx(), org, project, env)
		require.NoError(t, err)

		httpstub.RequireRequest(t, rec.Only(t), http.MethodDelete, "/api/esc/environments/acme/payments/prod", nil)
	})
}

func TestEnvironmentTagRequests(t *testing.T) {
	t.Parallel()
	const tagJSON = `{"name":"team","value":"platform","created":"2026-08-01T10:00:00Z","modified":"2026-08-01T10:00:00Z","editorLogin":"alice","editorName":"Alice Example"}`

	t.Run("create", func(t *testing.T) {
		t.Parallel()
		client, rec := newClient(t, httpstub.JSON(http.StatusOK, tagJSON))

		tag, err := client.CreateEnvironmentTag_esc_environments(ctx(), org, project, env, apitype.CreateEnvironmentTagRequest{Name: "team", Value: "platform"})
		require.NoError(t, err)
		require.Equal(t, "platform", tag.Value)

		req := rec.Only(t)
		httpstub.RequireRequest(t, req, http.MethodPost, "/api/esc/environments/acme/payments/prod/tags", nil)
		require.JSONEq(t, `{"name":"team","value":"platform"}`, req.Body)
	})

	t.Run("get", func(t *testing.T) {
		t.Parallel()
		client, rec := newClient(t, httpstub.JSON(http.StatusOK, tagJSON))

		tag, err := client.GetEnvironmentTag_esc_environments(ctx(), org, project, env, "team")
		require.NoError(t, err)
		require.Equal(t, "team", tag.Name)

		httpstub.RequireRequest(t, rec.Only(t), http.MethodGet, "/api/esc/environments/acme/payments/prod/tags/team", nil)
	})

	t.Run("update", func(t *testing.T) {
		t.Parallel()
		client, rec := newClient(t, httpstub.JSON(http.StatusOK, tagJSON))

		_, err := client.UpdateEnvironmentTag_esc_environments(ctx(), org, project, env, "team", apitype.UpdateEnvironmentTagRequest{
			CurrentTag: apitype.UpdateEnvironmentTagRequestCurrentTag{Value: "platform"},
			NewTag:     apitype.UpdateEnvironmentTagRequestNewTag{Name: "owner", Value: "payments"},
		})
		require.NoError(t, err)

		req := rec.Only(t)
		httpstub.RequireRequest(t, req, http.MethodPatch, "/api/esc/environments/acme/payments/prod/tags/team", nil)
		require.JSONEq(t, `{"currentTag":{"value":"platform"},"newTag":{"name":"owner","value":"payments"}}`, req.Body)
	})

	t.Run("delete", func(t *testing.T) {
		t.Parallel()
		client, rec := newClient(t, httpstub.Empty(http.StatusNoContent))

		err := client.DeleteEnvironmentTag_esc_environments(ctx(), org, project, env, "team")
		require.NoError(t, err)

		httpstub.RequireRequest(t, rec.Only(t), http.MethodDelete, "/api/esc/environments/acme/payments/prod/tags/team", nil)
	})
}

func TestRevisionTagRequests(t *testing.T) {
	t.Parallel()

	t.Run("create", func(t *testing.T) {
		t.Parallel()
		client, rec := newClient(t, httpstub.Empty(http.StatusNoContent))

		err := client.CreateRevisionTag_esc_environments_versions_tags(ctx(), org, project, env, apitype.CreateEnvironmentRevisionTagRequest{Name: "stable", Revision: ptr(3)})
		require.NoError(t, err)

		req := rec.Only(t)
		httpstub.RequireRequest(t, req, http.MethodPost, "/api/esc/environments/acme/payments/prod/versions/tags", nil)
		require.JSONEq(t, `{"name":"stable","revision":3}`, req.Body)
	})

	t.Run("read", func(t *testing.T) {
		t.Parallel()
		client, rec := newClient(t, httpstub.JSON(http.StatusOK, `{"name":"stable","revision":3,"created":"2026-08-01T10:00:00Z","modified":"2026-09-01T10:00:00Z"}`))

		tag, err := client.ReadRevisionTag_esc_environments(ctx(), org, project, env, "stable")
		require.NoError(t, err)
		require.Equal(t, 3, tag.Revision)

		httpstub.RequireRequest(t, rec.Only(t), http.MethodGet, "/api/esc/environments/acme/payments/prod/versions/tags/stable", nil)
	})

	t.Run("update", func(t *testing.T) {
		t.Parallel()
		client, rec := newClient(t, httpstub.Empty(http.StatusNoContent))

		err := client.UpdateRevisionTag_esc_environments(ctx(), org, project, env, "stable", apitype.UpdateEnvironmentRevisionTagRequest{Revision: ptr(4)})
		require.NoError(t, err)

		req := rec.Only(t)
		httpstub.RequireRequest(t, req, http.MethodPatch, "/api/esc/environments/acme/payments/prod/versions/tags/stable", nil)
		require.JSONEq(t, `{"revision":4}`, req.Body)
	})

	t.Run("delete", func(t *testing.T) {
		t.Parallel()
		client, rec := newClient(t, httpstub.Empty(http.StatusNoContent))

		err := client.DeleteRevisionTag_esc_environments(ctx(), org, project, env, "stable")
		require.NoError(t, err)

		httpstub.RequireRequest(t, rec.Only(t), http.MethodDelete, "/api/esc/environments/acme/payments/prod/versions/tags/stable", nil)
	})
}

func TestPathParametersAreEscaped(t *testing.T) {
	t.Parallel()
	client, rec := newClient(t, httpstub.YAML(environmentYAML))

	_, err := client.ReadEnvironment_esc_environments(ctx(), org, "my project", "blue/green")
	require.NoError(t, err)

	require.Equal(t, "/api/esc/environments/acme/my%20project/blue%2Fgreen", rec.Only(t).Path)
}

func TestExtraHeadersReachTheServer(t *testing.T) {
	t.Parallel()
	client, rec := newClient(t, httpstub.YAML(environmentYAML))

	_, err := client.ReadEnvironment_esc_environments(ctx(), org, project, env, http.Header{
		"Authorization":   {"token pul-secret"},
		"User-Agent":      {"esc-sdk/go"},
		"X-Pulumi-Source": {"esc-sdk"},
	})
	require.NoError(t, err)

	header := rec.Only(t).Header
	require.Equal(t, "token pul-secret", header.Get("Authorization"))
	require.Equal(t, "esc-sdk/go", header.Get("User-Agent"))
	require.Equal(t, "esc-sdk", header.Get("X-Pulumi-Source"))
	require.Equal(t, "application/vnd.pulumi+8", header.Get("Accept"))
}

func TestCancelledContextAbortsTheCall(t *testing.T) {
	t.Parallel()
	client, _ := newClient(t, httpstub.YAML(environmentYAML))
	cancelled, cancel := context.WithCancel(context.Background())
	cancel()

	_, err := client.ReadEnvironment_esc_environments(cancelled, org, project, env)
	require.ErrorIs(t, err, context.Canceled)
}
