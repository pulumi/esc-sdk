// Copyright 2026, Pulumi Corporation.  All rights reserved.

// Package cloudsdkcontract pins the behaviour the EscClient facade relies on
// from the pinned pulumi-cloud-sdk Go client: the routes each generated method
// sends, how it decodes recorded Pulumi Cloud responses, and what it does on
// errors. Every test goes through the generated client's request construction,
// transport, decoding and error handling against a local stub server.
package cloudsdkcontract

import (
	"context"
	"testing"

	"github.com/pulumi/esc-sdk/sdk/go/internal/httpstub"
	"github.com/pulumi/pulumi-cloud-sdk/go/apiclient"
)

const (
	org     = "acme"
	project = "payments"
	env     = "prod"
	version = "3"
)

func ptr[T any](v T) *T { return &v }

func ctx() context.Context { return context.Background() }

func newClient(t *testing.T, response httpstub.Response) (*apiclient.CloudClient, *httpstub.Recorder) {
	t.Helper()
	baseURL, rec := httpstub.Serve(t, response)
	return &apiclient.CloudClient{BaseURL: baseURL, Executor: httpstub.DefaultExecutor}, rec
}
