// Copyright 2025, Pulumi Corporation.  All rights reserved.

package esc_sdk

import (
	"net/url"
	"testing"

	"github.com/stretchr/testify/require"
)

func Test_EscClientLogin(t *testing.T) {
	t.Run("default auth context picks up PULUMI_ACCESS_TOKEN", func(t *testing.T) {
		t.Setenv("PULUMI_ACCESS_TOKEN", "FAKE_TOKEN")

		authContext, err := NewDefaultAuthContext()
		require.NoError(t, err)

		token, ok := AccessTokenFromContext(authContext)
		require.True(t, ok)
		require.Equal(t, "FAKE_TOKEN", token)
	})

	t.Run("default auth context fails without PULUMI_ACCESS_TOKEN", func(t *testing.T) {
		t.Setenv("PULUMI_ACCESS_TOKEN", "")

		_, err := NewDefaultAuthContext()
		require.Error(t, err)
	})

	t.Run("default client picks up PULUMI_BACKEND_URL", func(t *testing.T) {
		t.Setenv("PULUMI_BACKEND_URL", "https://api.moolumi.com")

		client, err := NewDefaultClient()
		require.NoError(t, err)
		require.Equal(t, "https://api.moolumi.com", client.Cloud.BaseURL)
	})

	t.Run("default client falls back to Pulumi Cloud", func(t *testing.T) {
		t.Setenv("PULUMI_BACKEND_URL", "")

		client, err := NewDefaultClient()
		require.NoError(t, err)
		require.Equal(t, DefaultPulumiAPIURL, client.Cloud.BaseURL)
	})
}

func Test_NewCustomBackendConfiguration(t *testing.T) {
	t.Run("keeps the port and drops the path", func(t *testing.T) {
		backend, err := url.Parse("http://localhost:8080/api/esc")
		require.NoError(t, err)

		cfg, err := NewCustomBackendConfiguration(*backend)
		require.NoError(t, err)
		require.Equal(t, "http://localhost:8080", cfg.BaseURL)
	})

	t.Run("rejects a URL without a host", func(t *testing.T) {
		backend, err := url.Parse("localhost:8080")
		require.NoError(t, err)

		_, err = NewCustomBackendConfiguration(*backend)
		require.Error(t, err)
	})
}
