// Copyright 2026, Pulumi Corporation.  All rights reserved.

package esc_sdk

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"

	"github.com/pulumi/pulumi-cloud-sdk/go/apiclient"
	"github.com/pulumi/pulumi-cloud-sdk/go/apitype"
	"gopkg.in/ghodss/yaml.v1"
)

func (c *EscClient) ListEnvironments(ctx context.Context, org string, continuationToken *string) (*OrgEnvironments, error) {
	return c.Cloud.ListOrgEnvironments_esc(ctx, org, continuationToken, nil, nil, nil)
}

// GetEnvironment returns the parsed definition and the YAML it was parsed from.
func (c *EscClient) GetEnvironment(ctx context.Context, org, projectName, envName string) (*EnvironmentDefinition, string, error) {
	return parseDefinition(c.Cloud.ReadEnvironment_esc_environments(ctx, org, projectName, envName))
}

func (c *EscClient) GetEnvironmentAtVersion(ctx context.Context, org, projectName, envName, version string) (*EnvironmentDefinition, string, error) {
	return parseDefinition(c.Cloud.ReadEnvironment_esc_environments_versions(ctx, org, projectName, envName, version))
}

// DecryptEnvironment returns the definition with secrets in plain text, parsed
// and as YAML.
func (c *EscClient) DecryptEnvironment(ctx context.Context, org, projectName, envName string) (*EnvironmentDefinition, string, error) {
	return parseDefinition(c.Cloud.DecryptEnvironment_esc_environments(ctx, org, projectName, envName))
}

func parseDefinition(definitionYAML *string, err error) (*EnvironmentDefinition, string, error) {
	if err != nil {
		return nil, "", err
	}
	var definition EnvironmentDefinition
	if err := yaml.Unmarshal([]byte(*definitionYAML), &definition); err != nil {
		return nil, "", fmt.Errorf("parsing environment definition: %w", err)
	}
	return &definition, *definitionYAML, nil
}

func (c *EscClient) OpenEnvironment(ctx context.Context, org, projectName, envName string) (*OpenEnvironment, error) {
	return c.Cloud.OpenEnvironment_esc_environments(ctx, org, projectName, envName, nil)
}

func (c *EscClient) OpenEnvironmentAtVersion(ctx context.Context, org, projectName, envName, version string) (*OpenEnvironment, error) {
	return c.Cloud.OpenEnvironment_esc_environments_versions(ctx, org, projectName, envName, version, nil)
}

// ReadOpenEnvironment returns the opened environment and its values without
// the trace envelopes.
func (c *EscClient) ReadOpenEnvironment(ctx context.Context, org, projectName, envName, openEnvID string) (*Environment, map[string]any, error) {
	// The generated method decodes into *any, which loses number precision on
	// the way to a typed value, so decode the captured body instead
	// (https://github.com/pulumi/pulumi-cloud-sdk/issues/6).
	ctx, capture := withResponseCapture(ctx)
	if _, err := c.Cloud.ReadOpenEnvironment_esc_environments(ctx, org, projectName, envName, openEnvID, nil); err != nil {
		return nil, nil, err
	}
	var env Environment
	if err := json.Unmarshal(capture.body, &env); err != nil {
		return nil, nil, fmt.Errorf("decoding opened environment: %w", err)
	}
	if env.Properties == nil {
		return &env, nil, nil
	}
	values := make(map[string]any, len(env.Properties))
	for key, property := range env.Properties {
		values[key] = plainValue(property.Value)
	}
	return &env, values, nil
}

func (c *EscClient) OpenAndReadEnvironment(ctx context.Context, org, projectName, envName string) (*Environment, map[string]any, error) {
	open, err := c.OpenEnvironment(ctx, org, projectName, envName)
	if err != nil {
		return nil, nil, err
	}
	return c.ReadOpenEnvironment(ctx, org, projectName, envName, open.ID)
}

func (c *EscClient) OpenAndReadEnvironmentAtVersion(ctx context.Context, org, projectName, envName, version string) (*Environment, map[string]any, error) {
	open, err := c.OpenEnvironmentAtVersion(ctx, org, projectName, envName, version)
	if err != nil {
		return nil, nil, err
	}
	return c.ReadOpenEnvironment(ctx, org, projectName, envName, open.ID)
}

// ReadEnvironmentProperty returns one property of an opened environment, both
// with its trace and as the plain value.
func (c *EscClient) ReadEnvironmentProperty(ctx context.Context, org, projectName, envName, openEnvID, propPath string) (*Value, any, error) {
	ctx, capture := withResponseCapture(ctx)
	if _, err := c.Cloud.ReadOpenEnvironment_esc_environments(ctx, org, projectName, envName, openEnvID, &propPath); err != nil {
		return nil, nil, err
	}
	var property Value
	if err := json.Unmarshal(capture.body, &property); err != nil {
		return nil, nil, fmt.Errorf("decoding property %q: %w", propPath, err)
	}
	return &property, plainValue(property.Value), nil
}

func (c *EscClient) CreateEnvironment(ctx context.Context, org, projectName, envName string) error {
	return c.Cloud.CreateEnvironment_esc_environments(ctx, org, apitype.CreateEnvironmentRequest{Project: projectName, Name: envName})
}

type CloneEnvironmentOptions struct {
	PreserveHistory         bool
	PreserveAccess          bool
	PreserveEnvironmentTags bool
	PreserveRevisionTags    bool
}

func (c *EscClient) CloneEnvironment(ctx context.Context, org, srcProjectName, srcEnvName, destProjectName, destEnvName string, options *CloneEnvironmentOptions) error {
	if options == nil {
		options = &CloneEnvironmentOptions{}
	}
	return c.Cloud.CloneEnvironment(ctx, org, srcProjectName, srcEnvName, apitype.CloneEnvironmentRequest{
		Project:                 destProjectName,
		Name:                    destEnvName,
		PreserveHistory:         options.PreserveHistory,
		PreserveAccess:          options.PreserveAccess,
		PreserveEnvironmentTags: options.PreserveEnvironmentTags,
		PreserveRevisionTags:    options.PreserveRevisionTags,
	})
}

// UpdateEnvironmentYaml replaces the definition. On a rejected definition the
// diagnostics are returned together with the error.
func (c *EscClient) UpdateEnvironmentYaml(ctx context.Context, org, projectName, envName, definitionYAML string) (*EnvironmentDiagnostics, error) {
	response, err := c.Cloud.UpdateEnvironment_esc_environments(ctx, org, projectName, envName, definitionYAML)
	if err != nil {
		return diagnosticsFromError(err), err
	}
	return &response.EnvironmentDiagnosticsResponse, nil
}

func (c *EscClient) UpdateEnvironment(ctx context.Context, org, projectName, envName string, env *EnvironmentDefinition) (*EnvironmentDiagnostics, error) {
	definitionYAML, err := MarshalEnvironmentDefinition(env)
	if err != nil {
		return nil, err
	}
	return c.UpdateEnvironmentYaml(ctx, org, projectName, envName, definitionYAML)
}

func (c *EscClient) DeleteEnvironment(ctx context.Context, org, projectName, envName string) error {
	return c.Cloud.DeleteEnvironment_esc_environments(ctx, org, projectName, envName)
}

func (c *EscClient) CheckEnvironment(ctx context.Context, org string, env *EnvironmentDefinition) (*CheckEnvironment, error) {
	definitionYAML, err := MarshalEnvironmentDefinition(env)
	if err != nil {
		return nil, err
	}
	return c.CheckEnvironmentYaml(ctx, org, definitionYAML)
}

// CheckEnvironmentYaml evaluates a definition without saving it. On a rejected
// definition the check result with its diagnostics is returned together with
// the error.
func (c *EscClient) CheckEnvironmentYaml(ctx context.Context, org, definitionYAML string) (*CheckEnvironment, error) {
	result, err := c.Cloud.CheckYAML_esc(ctx, org, nil, definitionYAML)
	if err != nil {
		return checkResultFromError(err), err
	}
	// An empty body leaves the embedded environment nil
	// (https://github.com/pulumi/pulumi-cloud-sdk/issues/6).
	if result.EscEnvironment == nil {
		result.EscEnvironment = &Environment{}
	}
	return result, nil
}

func rejectedDefinitionBody(err error) ([]byte, bool) {
	var apiErr *apiclient.APIError
	if !errors.As(err, &apiErr) || apiErr.HTTPStatusCode() != http.StatusBadRequest {
		return nil, false
	}
	return []byte(apiErr.ResponseMessage()), true
}

func diagnosticsFromError(err error) *EnvironmentDiagnostics {
	body, ok := rejectedDefinitionBody(err)
	if !ok {
		return nil
	}
	var diagnostics EnvironmentDiagnostics
	if json.Unmarshal(body, &diagnostics) != nil {
		return nil
	}
	return &diagnostics
}

func checkResultFromError(err error) *CheckEnvironment {
	body, ok := rejectedDefinitionBody(err)
	if !ok {
		return nil
	}
	var result CheckEnvironment
	if json.Unmarshal(body, &result) == nil {
		if result.EscEnvironment == nil {
			result.EscEnvironment = &Environment{}
		}
		return &result
	}
	// The full response cannot decode while the SDK rejects boolean-form
	// schemas (https://github.com/pulumi/pulumi-cloud-sdk/issues/6), and the
	// diagnostics are what a caller needs from a rejected definition.
	var diagnostics EnvironmentDiagnostics
	if json.Unmarshal(body, &diagnostics) != nil {
		return nil
	}
	return &CheckEnvironment{EscEnvironment: &Environment{}, Diagnostics: diagnostics.Diagnostics}
}

func (c *EscClient) ListEnvironmentRevisions(ctx context.Context, org, projectName, envName string) ([]EnvironmentRevision, error) {
	return derefSlice(c.Cloud.ListEnvironmentRevisions_esc_environments(ctx, org, projectName, envName, nil, nil))
}

func (c *EscClient) ListEnvironmentRevisionsPaginated(ctx context.Context, org, projectName, envName string, before, count int) ([]EnvironmentRevision, error) {
	return derefSlice(c.Cloud.ListEnvironmentRevisions_esc_environments(ctx, org, projectName, envName, &before, &count))
}

func derefSlice[T any](items *[]T, err error) ([]T, error) {
	if err != nil {
		return nil, err
	}
	return *items, nil
}

func (c *EscClient) ListEnvironmentRevisionTags(ctx context.Context, org, projectName, envName string) (*EnvironmentRevisionTags, error) {
	return c.Cloud.ListRevisionTags_esc_environments_versions(ctx, org, projectName, envName, nil, nil)
}

func (c *EscClient) ListEnvironmentRevisionTagsPaginated(ctx context.Context, org, projectName, envName string, after string, count int) (*EnvironmentRevisionTags, error) {
	return c.Cloud.ListRevisionTags_esc_environments_versions(ctx, org, projectName, envName, &after, &count)
}

func (c *EscClient) GetEnvironmentRevisionTag(ctx context.Context, org, projectName, envName, tagName string) (*EnvironmentRevisionTag, error) {
	return c.Cloud.ReadRevisionTag_esc_environments(ctx, org, projectName, envName, tagName)
}

func (c *EscClient) CreateEnvironmentRevisionTag(ctx context.Context, org, projectName, envName, tagName string, revision int) error {
	return c.Cloud.CreateRevisionTag_esc_environments_versions_tags(ctx, org, projectName, envName, apitype.CreateEnvironmentRevisionTagRequest{Name: tagName, Revision: &revision})
}

func (c *EscClient) UpdateEnvironmentRevisionTag(ctx context.Context, org, projectName, envName, tagName string, revision int) error {
	return c.Cloud.UpdateRevisionTag_esc_environments(ctx, org, projectName, envName, tagName, apitype.UpdateEnvironmentRevisionTagRequest{Revision: &revision})
}

func (c *EscClient) DeleteEnvironmentRevisionTag(ctx context.Context, org, projectName, envName, tagName string) error {
	return c.Cloud.DeleteRevisionTag_esc_environments(ctx, org, projectName, envName, tagName)
}

func (c *EscClient) ListEnvironmentTags(ctx context.Context, org, projectName, envName string) (*ListEnvironmentTags, error) {
	return c.Cloud.ListEnvironmentTags_esc_environments(ctx, org, projectName, envName, nil, nil)
}

func (c *EscClient) ListEnvironmentTagsPaginated(ctx context.Context, org, projectName, envName string, after, count int) (*ListEnvironmentTags, error) {
	return c.Cloud.ListEnvironmentTags_esc_environments(ctx, org, projectName, envName, &after, &count)
}

func (c *EscClient) GetEnvironmentTag(ctx context.Context, org, projectName, envName, tagName string) (*EnvironmentTag, error) {
	return c.Cloud.GetEnvironmentTag_esc_environments(ctx, org, projectName, envName, tagName)
}

func (c *EscClient) CreateEnvironmentTag(ctx context.Context, org, projectName, envName, tagName, tagValue string) (*EnvironmentTag, error) {
	return c.Cloud.CreateEnvironmentTag_esc_environments(ctx, org, projectName, envName, apitype.CreateEnvironmentTagRequest{Name: tagName, Value: tagValue})
}

func (c *EscClient) UpdateEnvironmentTag(ctx context.Context, org, projectName, envName, tagName, currentTagValue, newTagName, newTagValue string) (*EnvironmentTag, error) {
	return c.Cloud.UpdateEnvironmentTag_esc_environments(ctx, org, projectName, envName, tagName, apitype.UpdateEnvironmentTagRequest{
		CurrentTag: apitype.UpdateEnvironmentTagRequestCurrentTag{Value: currentTagValue},
		NewTag:     apitype.UpdateEnvironmentTagRequestNewTag{Name: newTagName, Value: newTagValue},
	})
}

func (c *EscClient) DeleteEnvironmentTag(ctx context.Context, org, projectName, envName, tagName string) error {
	return c.Cloud.DeleteEnvironmentTag_esc_environments(ctx, org, projectName, envName, tagName)
}

func MarshalEnvironmentDefinition(env *EnvironmentDefinition) (string, error) {
	encoded, err := yaml.Marshal(env)
	if err != nil {
		return "", err
	}
	return string(encoded), nil
}
