// Copyright 2026, Pulumi Corporation.  All rights reserved.

package cloudsdkcontract

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"testing"

	"github.com/pulumi/esc-sdk/sdk/go/internal/httpstub"
	"github.com/pulumi/pulumi-cloud-sdk/go/apiclient"
	"github.com/pulumi/pulumi-cloud-sdk/go/apitype"
	"github.com/stretchr/testify/require"
)

func requireAPIError(t *testing.T, err error, status int) *apiclient.APIError {
	t.Helper()
	require.Error(t, err)
	var apiErr *apiclient.APIError
	require.ErrorAs(t, err, &apiErr)
	require.Equal(t, status, apiErr.HTTPStatusCode())
	return apiErr
}

func TestDiagnosticsSurviveAnHTTP400(t *testing.T) {
	t.Parallel()
	body := httpstub.Fixture(t, "check_diagnostics_400.json")

	t.Run("check", func(t *testing.T) {
		t.Parallel()
		client, _ := newClient(t, httpstub.JSON(http.StatusBadRequest, body))

		result, err := client.CheckYAML_esc(ctx(), org, nil, environmentYAML)
		require.Nil(t, result)
		apiErr := requireAPIError(t, err, http.StatusBadRequest)

		var diagnostics apitype.EnvironmentDiagnosticsResponse
		require.NoError(t, json.Unmarshal([]byte(apiErr.ResponseMessage()), &diagnostics), "the error must carry the response body for the facade to recover diagnostics")
		require.Len(t, diagnostics.Diagnostics, 1)
		require.Equal(t, `unknown property "bad_ref"`, diagnostics.Diagnostics[0].Summary)
		require.Equal(t, "values.ref", diagnostics.Diagnostics[0].Path)
		require.Equal(t, 3, diagnostics.Diagnostics[0].Range.Begin.Line)
	})

	t.Run("update", func(t *testing.T) {
		t.Parallel()
		client, _ := newClient(t, httpstub.JSON(http.StatusBadRequest, `{"diagnostics":[{"summary":"unknown property \"bad_ref\""}]}`))

		result, err := client.UpdateEnvironment_esc_environments(ctx(), org, project, env, environmentYAML)
		require.Nil(t, result)
		apiErr := requireAPIError(t, err, http.StatusBadRequest)

		var diagnostics apitype.EnvironmentDiagnosticsResponse
		require.NoError(t, json.Unmarshal([]byte(apiErr.ResponseMessage()), &diagnostics))
		require.Equal(t, `unknown property "bad_ref"`, diagnostics.Diagnostics[0].Summary)
	})
}

func TestOrdinaryAPIErrors(t *testing.T) {
	t.Parallel()

	t.Run("structured error body", func(t *testing.T) {
		t.Parallel()
		client, _ := newClient(t, httpstub.JSON(http.StatusNotFound, `{"code":404,"message":"environment 'acme/payments/prod' not found"}`))

		_, err := client.ReadEnvironment_esc_environments(ctx(), org, project, env)
		apiErr := requireAPIError(t, err, http.StatusNotFound)
		require.True(t, apiErr.IsNotFound())
		require.Equal(t, "environment 'acme/payments/prod' not found", apiErr.ResponseMessage())
	})

	t.Run("plain text error body", func(t *testing.T) {
		t.Parallel()
		client, _ := newClient(t, httpstub.Response{Status: http.StatusBadGateway, ContentType: "text/plain", Body: "upstream timed out\n"})

		_, err := client.ReadEnvironment_esc_environments(ctx(), org, project, env)
		apiErr := requireAPIError(t, err, http.StatusBadGateway)
		require.Equal(t, "upstream timed out", apiErr.ResponseMessage())
	})

	t.Run("body-less method", func(t *testing.T) {
		t.Parallel()
		client, _ := newClient(t, httpstub.JSON(http.StatusConflict, `{"code":409,"message":"environment already exists"}`))

		err := client.CreateEnvironment_esc_environments(ctx(), org, apitype.CreateEnvironmentRequest{Project: project, Name: env})
		apiErr := requireAPIError(t, err, http.StatusConflict)
		require.True(t, apiErr.IsConflict())
	})
}

func TestEmptySuccessfulResponses(t *testing.T) {
	t.Parallel()

	t.Run("204 on a body-less method", func(t *testing.T) {
		t.Parallel()
		client, _ := newClient(t, httpstub.Empty(http.StatusNoContent))

		require.NoError(t, client.DeleteEnvironment_esc_environments(ctx(), org, project, env))
		require.NoError(t, client.CreateRevisionTag_esc_environments_versions_tags(ctx(), org, project, env, apitype.CreateEnvironmentRevisionTagRequest{Name: "stable"}))
	})

	t.Run("200 with an empty JSON body on a body-less method", func(t *testing.T) {
		t.Parallel()
		client, _ := newClient(t, httpstub.JSON(http.StatusOK, `null`))

		require.NoError(t, client.CreateEnvironment_esc_environments(ctx(), org, apitype.CreateEnvironmentRequest{Project: project, Name: env}))
	})

	t.Run("200 with an empty diagnostics object", func(t *testing.T) {
		t.Parallel()
		client, _ := newClient(t, httpstub.JSON(http.StatusOK, `{}`))

		result, err := client.UpdateEnvironment_esc_environments(ctx(), org, project, env, environmentYAML)
		require.NoError(t, err)
		require.Empty(t, result.Diagnostics)
	})
}

func callWithoutPanic(t *testing.T, call func() error) (err error) {
	t.Helper()
	defer func() {
		if recovered := recover(); recovered != nil {
			err = fmt.Errorf("call panicked: %v", recovered)
			require.Fail(t, "an invalid base URL must surface as an error, not a panic, see https://github.com/pulumi/pulumi-cloud-sdk/issues/6", "%v", recovered)
		}
	}()
	return call()
}

func TestInvalidBaseURLIsAnError(t *testing.T) {
	t.Parallel()
	client := &apiclient.CloudClient{BaseURL: "://not-a-url", Executor: httpstub.DefaultExecutor}

	cases := map[string]func() error{
		"read": func() error {
			_, err := client.ReadEnvironment_esc_environments(ctx(), org, project, env)
			return err
		},
		"check YAML": func() error {
			_, err := client.CheckYAML_esc(ctx(), org, nil, environmentYAML)
			return err
		},
		"update YAML": func() error {
			_, err := client.UpdateEnvironment_esc_environments(ctx(), org, project, env, environmentYAML)
			return err
		},
	}
	for name, call := range cases {
		t.Run(name, func(t *testing.T) {
			t.Parallel()
			err := callWithoutPanic(t, call)
			require.Error(t, err)
			var apiErr *apiclient.APIError
			require.False(t, errors.As(err, &apiErr), "a client-side failure is not an API error")
		})
	}
}
