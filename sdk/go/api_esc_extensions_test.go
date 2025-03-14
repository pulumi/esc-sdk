// Copyright 2025, Pulumi Corporation.  All rights reserved.

/*
ESC (Environments, Secrets, Config) API

Testing EscAPIService

*/

package esc_sdk

import (
	"os"
	"testing"

	"github.com/stretchr/testify/require"
)

func Test_EscClientLogin(t *testing.T) {
	t.Run("verify default auth context picks up PULUMI_ACCESS_TOKEN variable", func(t *testing.T) {
		beforeTest := os.Getenv("PULUMI_ACCESS_TOKEN")
		os.Setenv("PULUMI_ACCESS_TOKEN", "FAKE_TOKEN")

		authContext, err := NewDefaultAuthContext()
		require.NoError(t, err)

		auth, ok := authContext.Value(ContextAPIKeys).(map[string]APIKey)
		require.True(t, ok)
		token, ok := auth["Authorization"]
		require.True(t, ok)
		require.Equal(t, "FAKE_TOKEN", token.Key)

		os.Setenv("PULUMI_ACCESS_TOKEN", beforeTest)
	})

	t.Run("verify default client picks up PULUMI_BACKEND_URL by default", func(t *testing.T) {
		beforeTest := os.Getenv("PULUMI_BACKEND_URL")
		os.Setenv("PULUMI_BACKEND_URL", "https://api.moolumi.com")

		client, err := NewDefaultClient()
		require.NoError(t, err)

		url, err := client.rawClient.cfg.ServerURL(0, make(map[string]string))
		require.NoError(t, err)
		require.Equal(t, "https://api.moolumi.com/api/esc", url)

		os.Setenv("PULUMI_BACKEND_URL", beforeTest)
	})
}
