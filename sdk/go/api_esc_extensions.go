// Copyright 2024, Pulumi Corporation.  All rights reserved.

package esc_sdk

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/url"
	"os"

	esc_workspace "github.com/pulumi/esc/cmd/esc/cli/workspace"
	"gopkg.in/ghodss/yaml.v1"
)

// Enum constants for discriminated union types

// ChangeGateRuleType constants
const (
	ChangeGateRuleTypeApprovalRequired = "approval_required"
)

// ChangeGateTargetActionType constants
const (
	ChangeGateTargetActionTypeUpdate = "update"
)

// EscClient is a client for the ESC API.
// It wraps the raw API client and provides a more convenient interface.
type EscClient struct {
	rawClient *RawAPIClient
	EscAPI    *EscAPIService
}

// NewAuthContext creates a new context with the given access token.
// This context can be used to authenticate requests to the ESC API.
func NewAuthContext(accessToken string) context.Context {
	return context.WithValue(
		context.Background(),
		ContextAPIKeys,
		map[string]APIKey{
			"Authorization": {Key: accessToken, Prefix: "token"},
		},
	)
}

// DefaultAuthContext creates a new context, retrieving Pulumi Access Token
// from either PULUMI_ACCESS_TOKEN environment variable, or from
// currently logged in account in Pulumi CLI or ESC CLI
//
// This context can be used to authenticate requests to the ESC API.
func NewDefaultAuthContext() (context.Context, error) {
	accessToken := os.Getenv("PULUMI_ACCESS_TOKEN")
	if accessToken != "" {
		return NewAuthContext(accessToken), nil
	}

	workspace := esc_workspace.New(esc_workspace.DefaultFS(), esc_workspace.DefaultPulumiWorkspace())
	account, _, err := workspace.GetCurrentAccount(false)
	if err != nil {
		return nil, fmt.Errorf("Error grabbing current account: %w", err)
	}
	if account != nil {
		return NewAuthContext(account.AccessToken), nil
	}

	return nil, errors.New("no default Pulumi Access Token found. Either export PULUMI_ACCESS_TOKEN " +
		"environment variable, or login using Pulumi or ESC CLI")
}

// NewClient creates a new ESC client with the given configuration.
func NewClient(cfg *Configuration) *EscClient {
	client := &EscClient{rawClient: NewRawAPIClient(cfg)}
	client.EscAPI = client.rawClient.EscAPI
	return client
}

// NewCustomBackendConfiguration creates a new Configuration object,
// but replaces default API endpoint with a given custom backend URL
func NewCustomBackendConfiguration(customBackendURL url.URL) (*Configuration, error) {
	cfg := NewConfiguration()
	cfg.DefaultHeader = make(map[string]string)
	cfg.UserAgent = "esc-sdk"
	cfg.Debug = false
	// Override just the server URL, keeping the rest from the generated configuration
	cfg.Servers = ServerConfigurations{
		{
			URL:         fmt.Sprintf("%s://%s/api", customBackendURL.Scheme, customBackendURL.Hostname()),
			Description: "Pulumi Cloud Custom Backend API",
		},
	}
	return cfg, nil
}

// NewDefaultClient creates a new ESC client with default configuration.
// Backend URL is automatically detected from either PULUMI_BACKEND_URL environment variable
// or currently logged in account in Pulumi CLI or ESC CLI
func NewDefaultClient() (*EscClient, error) {
	workspace := esc_workspace.New(esc_workspace.DefaultFS(), esc_workspace.DefaultPulumiWorkspace())
	account, _, err := workspace.GetCurrentAccount(false)
	if err != nil {
		return nil, fmt.Errorf("Error grabbing current account: %w", err)
	}
	customBackendURL := workspace.GetCurrentCloudURL(account)
	parsedUrl, err := url.Parse(customBackendURL)
	if err != nil {
		return nil, fmt.Errorf("Error parsing custom backend url: %w", err)
	}
	config, err := NewCustomBackendConfiguration(*parsedUrl)
	if err != nil {
		return nil, err
	}
	return NewClient(config), nil
}

// This is the easiest way to use ESC SDK. DefaultLogin grabs default client
// and default authorization context, so you can start using SDK right away
func DefaultLogin() (context.Context, *EscClient, error) {
	client, err := NewDefaultClient()
	if err != nil {
		return nil, nil, err
	}
	context, err := NewDefaultAuthContext()
	if err != nil {
		return nil, nil, err
	}
	return context, client, nil
}

// ListEnvironments lists all environments in the given organization.
// If a continuation token is provided, the list will start from that token.
func (c *EscClient) ListEnvironments(ctx context.Context, org string, continuationToken *string) (*OrgEnvironments, error) {
	request := c.EscAPI.ListEnvironments(ctx, org)
	if continuationToken != nil {
		request = request.ContinuationToken(*continuationToken)
	}

	envs, _, err := request.Execute()
	return envs, err
}

// GetEnvironment retrieves the environment with the given name in the given organization.
// The environment is returned along with the raw YAML definition.
func (c *EscClient) GetEnvironment(ctx context.Context, org, projectName, envName string) (*EnvironmentDefinition, string, error) {
	env, resp, err := c.EscAPI.GetEnvironment(ctx, org, projectName, envName).Execute()
	if err != nil {
		return nil, "", err
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, "", err
	}

	return env, string(body), nil
}

// GetEnvironmentAtVersion retrieves the environment with the given name in the given organization at the given version.
// The environment is returned along with the raw YAML definition.
func (c *EscClient) GetEnvironmentAtVersion(ctx context.Context, org, projectName, envName, version string) (*EnvironmentDefinition, string, error) {
	env, resp, err := c.EscAPI.GetEnvironmentAtVersion(ctx, org, projectName, envName, version).Execute()
	if err != nil {
		return nil, "", err
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, "", err
	}

	return env, string(body), nil
}

// OpenEnvironment opens the environment with the given name in the given organization.
// The open environment is returned, which contains the ID of the opened environment session to use with ReadOpenEnvironment.
func (c *EscClient) OpenEnvironment(ctx context.Context, org, projectName, envName string) (*OpenEnvironment, error) {
	openInfo, _, err := c.EscAPI.OpenEnvironment(ctx, org, projectName, envName).Execute()
	return openInfo, err
}

// OpenEnvironmentAtVersion opens the environment with the given name in the given organization at the given version.
// The open environment is returned, which contains the ID of the opened environment session to use with ReadOpenEnvironment.
func (c *EscClient) OpenEnvironmentAtVersion(ctx context.Context, org, projectName, envName, version string) (*OpenEnvironment, error) {
	openInfo, _, err := c.EscAPI.OpenEnvironmentAtVersion(ctx, org, projectName, envName, version).Execute()
	return openInfo, err
}

// ReadOpenEnvironment reads the environment with the given open session ID and returns the config and resolved secret values.
func (c *EscClient) ReadOpenEnvironment(ctx context.Context, org, projectName, envName, openEnvID string) (*Environment, map[string]any, error) {
	env, _, err := c.EscAPI.ReadOpenEnvironment(ctx, org, projectName, envName, openEnvID).Execute()
	if err != nil {
		return nil, nil, err
	}

	if env == nil || env.Properties == nil {
		return nil, nil, nil
	}

	propertyMap := *env.Properties
	for k, v := range propertyMap {
		v.Value = mapValues(v.Value)
		propertyMap[k] = v
	}

	values := make(map[string]any, len(propertyMap))
	for k := range propertyMap {
		v := propertyMap[k]
		values[k] = mapValuesPrimitive(&v)
	}

	return env, values, nil
}

// OpenAndReadEnvironment opens and reads the environment with the given name in the given organization.
// The config and resolved secret values are returned.
func (c *EscClient) OpenAndReadEnvironment(ctx context.Context, org, projectName, envName string) (*Environment, map[string]any, error) {
	openInfo, err := c.OpenEnvironment(ctx, org, projectName, envName)
	if err != nil {
		return nil, nil, err
	}

	return c.ReadOpenEnvironment(ctx, org, projectName, envName, openInfo.Id)
}

// OpenAndReadEnvironmentAtVersion opens and reads the environment with the given name in the given organization at the given version.
// The config and resolved secret values are returned.
func (c *EscClient) OpenAndReadEnvironmentAtVersion(ctx context.Context, org, projectName, envName, version string) (*Environment, map[string]any, error) {
	openInfo, err := c.OpenEnvironmentAtVersion(ctx, org, projectName, envName, version)
	if err != nil {
		return nil, nil, err
	}

	return c.ReadOpenEnvironment(ctx, org, projectName, envName, openInfo.Id)
}

// ReadEnvironmentProperty reads the property at the given path in the environment with the given open session ID.
// The property is returned along with the resolved value.
func (c *EscClient) ReadEnvironmentProperty(ctx context.Context, org, projectName, envName, openEnvID, propPath string) (*Value, any, error) {
	prop, _, err := c.EscAPI.ReadOpenEnvironmentProperty(ctx, org, projectName, envName, openEnvID).Property(propPath).Execute()
	if prop == nil {
		return nil, nil, err
	}

	v := mapValuesPrimitive(prop.Value)
	return prop, v, err
}

// CreateEnvironment creates a new environment with the given name in the given organization.
func (c *EscClient) CreateEnvironment(ctx context.Context, org, projectName, envName string) error {
	createEnvironment := NewCreateEnvironment(projectName, envName)
	_, _, err := c.EscAPI.CreateEnvironment(ctx, org).CreateEnvironment(*createEnvironment).Execute()
	return err
}

type CloneEnvironmentOptions struct {
	PreserveHistory         bool
	PreserveAccess          bool
	PreserveEnvironmentTags bool
	PreserveRevisionTags    bool
}

// CloneEnvironment clones an existing environment into a new environment.
func (c *EscClient) CloneEnvironment(ctx context.Context, org, srcProjectName, srcEnvName, destProjectName, destEnvName string, cloneEnvironmentOptions *CloneEnvironmentOptions) error {
	cloneEnvironment := NewCloneEnvironment(destProjectName, destEnvName)
	cloneEnvironment.PreserveHistory = &cloneEnvironmentOptions.PreserveHistory
	cloneEnvironment.PreserveAccess = &cloneEnvironmentOptions.PreserveAccess
	cloneEnvironment.PreserveEnvironmentTags = &cloneEnvironmentOptions.PreserveEnvironmentTags
	cloneEnvironment.PreserveRevisionTags = &cloneEnvironmentOptions.PreserveRevisionTags

	_, err := c.EscAPI.CloneEnvironment(ctx, org, srcProjectName, srcEnvName).CloneEnvironment(*cloneEnvironment).Execute()
	return err
}

// UpdateEnvironmentYaml updates the environment with the given name in the given organization with the given YAML definition.
func (c *EscClient) UpdateEnvironmentYaml(ctx context.Context, org, projectName, envName, yaml string) (*EnvironmentDiagnostics, error) {
	diags, _, err := c.EscAPI.UpdateEnvironmentYaml(ctx, org, projectName, envName).Body(yaml).Execute()
	return diags, err
}

// UpdateEnvironment updates the environment with the given name in the given organization with the given definition.
func (c *EscClient) UpdateEnvironment(ctx context.Context, org, projectName, envName string, env *EnvironmentDefinition) (*EnvironmentDiagnostics, error) {
	yaml, err := MarshalEnvironmentDefinition(env)
	if err != nil {
		return nil, err
	}

	diags, _, err := c.EscAPI.UpdateEnvironmentYaml(ctx, org, projectName, envName).Body(yaml).Execute()
	return diags, err
}

// DeleteEnvironment deletes the environment with the given name in the given organization.
func (c *EscClient) DeleteEnvironment(ctx context.Context, org, projectName, envName string) error {
	_, _, err := c.EscAPI.DeleteEnvironment(ctx, org, projectName, envName).Execute()
	return err
}

// CheckEnvironment checks the given environment definition for errors.
func (c *EscClient) CheckEnvironment(ctx context.Context, org string, env *EnvironmentDefinition) (*CheckEnvironment, error) {
	yaml, err := MarshalEnvironmentDefinition(env)
	if err != nil {
		return nil, err
	}

	return c.CheckEnvironmentYaml(ctx, org, yaml)
}

// CheckEnvironmentYaml checks the given environment YAML definition for errors.
func (c *EscClient) CheckEnvironmentYaml(ctx context.Context, org, yaml string) (*CheckEnvironment, error) {
	check, _, err := c.EscAPI.CheckEnvironmentYaml(ctx, org).Body(yaml).Execute()
	var genericOpenApiError *GenericOpenAPIError
	if err != nil && errors.As(err, &genericOpenApiError) {
		model := genericOpenApiError.Model().(CheckEnvironment)
		return &model, err
	}

	return check, err
}

// DecryptEnvironment decrypts the environment with the given name in the given organization.
func (c *EscClient) DecryptEnvironment(ctx context.Context, org, projectName, envName string) (*EnvironmentDefinition, string, error) {
	env, resp, err := c.EscAPI.DecryptEnvironment(ctx, org, projectName, envName).Execute()

	body, bodyErr := io.ReadAll(resp.Body)
	if err != nil {
		return nil, "", bodyErr
	}

	return env, string(body), err
}

// ListEnvironmentRevisions lists all revisions of the environment with the given name in the given organization.
func (c *EscClient) ListEnvironmentRevisions(ctx context.Context, org, projectName, envName string) ([]EnvironmentRevision, error) {
	request := c.EscAPI.ListEnvironmentRevisions(ctx, org, projectName, envName)

	revs, _, err := request.Execute()
	return revs, err
}

// ListEnvironmentRevisionsPaginated lists all revisions of the environment with the given name in the given organization, with pagination support.
func (c *EscClient) ListEnvironmentRevisionsPaginated(ctx context.Context, org, projectName, envName string, before, count int32) ([]EnvironmentRevision, error) {
	request := c.EscAPI.ListEnvironmentRevisions(ctx, org, projectName, envName).Before(before).Count(count)

	revs, _, err := request.Execute()
	return revs, err
}

// ListEnvironmentRevisionTags lists all tags of the environment with the given name in the given organization.
func (c *EscClient) ListEnvironmentRevisionTags(ctx context.Context, org, projectName, envName string) (*EnvironmentRevisionTags, error) {
	request := c.EscAPI.client.EscAPI.ListEnvironmentRevisionTags(ctx, org, projectName, envName)

	revs, _, err := request.Execute()
	return revs, err
}

// ListEnvironmentRevisionTagsPaginated lists all tags of the environment with the given name in the given organization, with pagination support.
func (c *EscClient) ListEnvironmentRevisionTagsPaginated(ctx context.Context, org, projectName, envName string, after string, count int32) (*EnvironmentRevisionTags, error) {
	request := c.EscAPI.ListEnvironmentRevisionTags(ctx, org, projectName, envName).After(after).Count(count)

	tags, _, err := request.Execute()
	return tags, err
}

// GetEnvironmentRevisionTag retrieves the tag with the given name of the environment with the given name in the given organization.
func (c *EscClient) GetEnvironmentRevisionTag(ctx context.Context, org, projectName, envName, tagName string) (*EnvironmentRevisionTag, error) {
	request := c.EscAPI.client.EscAPI.GetEnvironmentRevisionTag(ctx, org, projectName, envName, tagName)

	revision, _, err := request.Execute()
	return revision, err
}

// CreateEnvironmentRevisionTag creates a new tag with the given name for the environment with the given name in the given organization.
func (c *EscClient) CreateEnvironmentRevisionTag(ctx context.Context, org, projectName, envName, tagName string, revision int32) error {
	createTag := NewCreateEnvironmentRevisionTag(tagName, revision)
	request := c.EscAPI.client.EscAPI.CreateEnvironmentRevisionTag(ctx, org, projectName, envName).CreateEnvironmentRevisionTag(*createTag)

	_, err := request.Execute()
	return err
}

// UpdateEnvironmentRevisionTag updates the tag's revision with the given name for the environment with the given name in the given organization.
func (c *EscClient) UpdateEnvironmentRevisionTag(ctx context.Context, org, projectName, envName, tagName string, revision int32) error {
	update := NewUpdateEnvironmentRevisionTag(revision)
	request := c.EscAPI.client.EscAPI.UpdateEnvironmentRevisionTag(ctx, org, projectName, envName, tagName).UpdateEnvironmentRevisionTag(*update)

	_, err := request.Execute()
	return err
}

// DeleteEnvironmentRevisionTag deletes the tag with the given name for the environment with the given name in the given organization.
func (c *EscClient) DeleteEnvironmentRevisionTag(ctx context.Context, org, projectName, envName, tagName string) error {
	request := c.EscAPI.client.EscAPI.DeleteEnvironmentRevisionTag(ctx, org, projectName, envName, tagName)

	_, err := request.Execute()
	return err
}

// ListEnvironmentTags lists all tags of the environment with the given name in the given organization.
func (c *EscClient) ListEnvironmentTags(ctx context.Context, org, projectName, envName string) (*ListEnvironmentTags, error) {
	request := c.EscAPI.client.EscAPI.ListEnvironmentTags(ctx, org, projectName, envName)

	tags, _, err := request.Execute()
	return tags, err
}

// ListEnvironmentTagsPaginated lists all tags of the environment with the given name in the given organization, with pagination support.
func (c *EscClient) ListEnvironmentTagsPaginated(ctx context.Context, org, projectName, envName string, after string, count int32) (*ListEnvironmentTags, error) {
	request := c.EscAPI.ListEnvironmentTags(ctx, org, projectName, envName).After(after).Count(count)

	tags, _, err := request.Execute()
	return tags, err
}

// GetEnvironmentTag retrieves the tag with the given name of the environment with the given name in the given organization.
func (c *EscClient) GetEnvironmentTag(ctx context.Context, org, projectName, envName, tagName string) (*EnvironmentTag, error) {
	request := c.EscAPI.client.EscAPI.GetEnvironmentTag(ctx, org, projectName, envName, tagName)

	tag, _, err := request.Execute()
	return tag, err
}

// CreateEnvironmentTag creates a new tag with the given name for the environment with the given name in the given organization.
func (c *EscClient) CreateEnvironmentTag(ctx context.Context, org, projectName, envName, tagName, tagValue string) (*EnvironmentTag, error) {
	createTag := NewCreateEnvironmentTag(tagName, tagValue)
	request := c.EscAPI.client.EscAPI.CreateEnvironmentTag(ctx, org, projectName, envName).CreateEnvironmentTag(*createTag)

	tag, _, err := request.Execute()
	return tag, err
}

// UpdateEnvironmentTag updates the tag's value with the given name for the environment with the given name in the given organization.
func (c *EscClient) UpdateEnvironmentTag(ctx context.Context, org, projectName, envName, tagName, currentTagValue, newTagName, newTagValue string) (*EnvironmentTag, error) {
	update := NewUpdateEnvironmentTag(
		UpdateEnvironmentTagCurrentTag{currentTagValue},
		UpdateEnvironmentTagNewTag{
			Name:  newTagName,
			Value: newTagValue,
		})
	request := c.EscAPI.client.EscAPI.UpdateEnvironmentTag(ctx, org, projectName, envName, tagName).UpdateEnvironmentTag(*update)

	tag, _, err := request.Execute()
	return tag, err
}

// DeleteEnvironmentTag deletes the tag with the given name for the environment with the given name in the given organization.
func (c *EscClient) DeleteEnvironmentTag(ctx context.Context, org, projectName, envName, tagName string) error {
	request := c.EscAPI.client.EscAPI.DeleteEnvironmentTag(ctx, org, projectName, envName, tagName)

	_, err := request.Execute()
	return err
}

func MarshalEnvironmentDefinition(env *EnvironmentDefinition) (string, error) {
	var bs []byte
	bs, err := yaml.Marshal(env)
	if err == nil {
		return string(bs), nil
	}

	return "", err
}

func mapValuesPrimitive(value any) any {
	switch val := value.(type) {
	case *Value:
		return mapValuesPrimitive(val.Value)
	case map[string]Value:
		output := make(map[string]any, len(val))
		for k, v := range val {
			output[k] = mapValuesPrimitive(v.Value)
		}

		return output
	case []any:
		for i, v := range val {
			val[i] = mapValuesPrimitive(v)
		}
		return val
	default:
		return value
	}
}

func mapValues(value any) any {
	if val := getValue(getMapSafe(value)); val != nil {
		val.Value = mapValues(val.Value)
		return val
	}
	if mapData, isMap := value.(map[string]any); isMap {
		output := map[string]Value{}
		for key, v := range mapData {
			value := mapValues(v)
			if value == nil {
				continue
			}

			if v, ok := value.(*Value); ok && v != nil {
				output[key] = *v
			} else {
				output[key] = Value{
					Value: value,
				}
			}
		}
		return output
	} else if sliceData, isSlice := value.([]any); isSlice {
		for i, v := range sliceData {
			sliceData[i] = mapValues(v)
		}
		return sliceData
	}

	return value
}

func getValue(data map[string]any) *Value {
	_, hasValue := data["value"]
	_, hasTrace := data["trace"]
	if hasValue && hasTrace {
		return &Value{
			Value:   mapValues(data["value"]),
			Secret:  getBoolPtr(data, "secret"),
			Unknown: getBoolPtr(data, "unknown"),
			Trace:   getTrace(data["trace"].(map[string]any)),
		}
	}

	return nil
}

func getTrace(data map[string]any) Trace {
	def := getRange(getMapSafe(data["def"]))
	base := getValue(getMapSafe(data["base"]))
	if def != nil || base != nil {
		return Trace{
			Def:  def,
			Base: base,
		}
	}

	return Trace{}
}

func getMapSafe(data any) map[string]any {
	if data == nil {
		return nil
	}

	val, _ := data.(map[string]any)
	return val
}

func getRange(data map[string]any) *Range {
	begin := getPos(getMapSafe(data["begin"]))
	end := getPos(getMapSafe(data["end"]))
	environment := data["environment"].(string)
	if begin != nil && end != nil {
		return &Range{
			Environment: environment,
			Begin:       *begin,
			End:         *end,
		}
	}

	return nil
}

func getPos(data map[string]any) *Pos {
	line, hasLine := data["line"].(float64)
	column, hasColumn := data["column"].(float64)
	byteData, hasByte := data["byte"].(float64)
	if hasLine || hasColumn || hasByte {
		return &Pos{
			Line:   int32(line),
			Column: int32(column),
			Byte:   int32(byteData),
		}
	}

	return nil
}

func getBoolPtr(data map[string]any, key string) *bool {
	val, exists := data[key]
	if exists {
		v, ok := val.(bool)
		if ok {
			return &v
		}
	}

	return nil
}

// Change Management Wrapper Functions

// ListChangeRequests lists change requests for an organization.
func (c *EscClient) ListChangeRequests(ctx context.Context, org string) (*ListChangeRequestsResponse, error) {
	resp, _, err := c.EscAPI.ListChangeRequests(ctx, org).Execute()
	return resp, err
}

// GetChangeRequest retrieves a specific change request with gate evaluation details.
func (c *EscClient) GetChangeRequest(ctx context.Context, org, changeRequestID string) (*GetChangeRequestResponse, error) {
	resp, _, err := c.EscAPI.GetChangeRequest(ctx, org, changeRequestID).Execute()
	return resp, err
}

// UpdateChangeRequest updates a change request's description.
func (c *EscClient) UpdateChangeRequest(ctx context.Context, org, changeRequestID, description string) (*ChangeRequest, error) {
	req := NewUpdateChangeRequestRequest()
	req.Description = &description

	resp, _, err := c.EscAPI.UpdateChangeRequest(ctx, org, changeRequestID).UpdateChangeRequestRequest(*req).Execute()
	return resp, err
}

// SubmitChangeRequest submits a change request for approval.
func (c *EscClient) SubmitChangeRequest(ctx context.Context, org, changeRequestID, description string) error {
	req := NewSubmitChangeRequestRequest()
	req.Description = &description

	_, err := c.EscAPI.SubmitChangeRequest(ctx, org, changeRequestID).SubmitChangeRequestRequest(*req).Execute()
	return err
}

// ApplyChangeRequest applies an approved change request.
func (c *EscClient) ApplyChangeRequest(ctx context.Context, org, changeRequestID string) (*ChangeRequestApplyResult, error) {
	resp, _, err := c.EscAPI.ApplyChangeRequest(ctx, org, changeRequestID).Execute()
	return resp, err
}

// CloseChangeRequest closes a change request without applying it.
func (c *EscClient) CloseChangeRequest(ctx context.Context, org, changeRequestID string, comment *string) error {
	req := NewCloseChangeRequestRequest()
	req.Comment = comment

	_, err := c.EscAPI.CloseChangeRequest(ctx, org, changeRequestID).CloseChangeRequestRequest(*req).Execute()
	return err
}

// ApproveChangeRequest approves a change request.
func (c *EscClient) ApproveChangeRequest(ctx context.Context, org, changeRequestID string, revisionNumber int64, comment *string) error {
	req := NewApproveChangeRequestRequest(revisionNumber)
	req.Comment = comment

	_, err := c.EscAPI.ApproveChangeRequest(ctx, org, changeRequestID).ApproveChangeRequestRequest(*req).Execute()
	return err
}

// UnapproveChangeRequest withdraws approval from a change request.
func (c *EscClient) UnapproveChangeRequest(ctx context.Context, org, changeRequestID string, comment *string) error {
	req := NewUnapproveChangeRequestRequest()
	req.Comment = comment

	_, err := c.EscAPI.UnapproveChangeRequest(ctx, org, changeRequestID).UnapproveChangeRequestRequest(*req).Execute()
	return err
}

// ListChangeRequestEvents lists events for a change request.
func (c *EscClient) ListChangeRequestEvents(ctx context.Context, org, changeRequestID string) (*ListChangeRequestEventsResponse, error) {
	resp, _, err := c.EscAPI.ListChangeRequestEvents(ctx, org, changeRequestID).Execute()
	return resp, err
}

// CreateEnvironmentDraft creates a new environment draft (change request).
func (c *EscClient) CreateEnvironmentDraft(ctx context.Context, org, projectName, envName, yamlContent string) (*ChangeRequestRef, error) {
	resp, _, err := c.EscAPI.CreateEnvironmentDraft(ctx, org, projectName, envName).Body(yamlContent).Execute()
	return resp, err
}

// ReadEnvironmentDraft reads environment draft content.
func (c *EscClient) ReadEnvironmentDraft(ctx context.Context, org, projectName, envName, changeRequestID string, revision *int32) (string, error) {
	req := c.EscAPI.ReadEnvironmentDraft(ctx, org, projectName, envName, changeRequestID)
	if revision != nil {
		req = req.Revision(*revision)
	}

	resp, _, err := req.Execute()
	return resp, err
}

// UpdateEnvironmentDraft updates environment draft content with optimistic concurrency control.
func (c *EscClient) UpdateEnvironmentDraft(ctx context.Context, org, projectName, envName, changeRequestID, etag, yamlContent string) (*ChangeRequestRef, error) {
	resp, _, err := c.EscAPI.UpdateEnvironmentDraft(ctx, org, projectName, envName, changeRequestID).
		IfMatch(etag).
		Body(yamlContent).
		Execute()
	return resp, err
}

// OpenEnvironmentDraft opens an environment draft for evaluation.
func (c *EscClient) OpenEnvironmentDraft(ctx context.Context, org, projectName, envName, changeRequestID string, revision *int32, duration *string) (*OpenEnvironmentDraftResponse, error) {
	req := c.EscAPI.OpenEnvironmentDraft(ctx, org, projectName, envName, changeRequestID)
	if revision != nil {
		req = req.Revision(*revision)
	}
	if duration != nil {
		req = req.Duration(*duration)
	}

	resp, _, err := req.Execute()
	return resp, err
}

// Change Gate Methods

// ListEnvironmentChangeGates lists change gates for an environment.
func (c *EscClient) ListEnvironmentChangeGates(ctx context.Context, org, projectName, envName string) (*EnvironmentChangeGatesResponse, error) {
	return c.listEnvironmentChangeGatesWithRawJSON(ctx, org, projectName, envName)
}

// GetEnvironmentChangeGate retrieves a specific change gate.
func (c *EscClient) GetChangeGate(ctx context.Context, org, gateID string) (*EnvironmentChangeGate, error) {
	return c.getChangeGateWithRawJSON(ctx, org, gateID)
}

// UpdateChangeGate updates a change gate using properly marshaled discriminated unions.
func (c *EscClient) UpdateChangeGate(ctx context.Context, org, gateID string, config *ChangeGateUpdateConfig) (*EnvironmentChangeGate, error) {
	return c.UpdateEnvironmentChangeGate(ctx, org, gateID, config)
}

// UpdateEnvironmentChangeGate updates a change gate using properly marshaled discriminated unions.
// This method uses structured objects and marshals them to preserve discriminated union fields during marshaling.
func (c *EscClient) UpdateEnvironmentChangeGate(ctx context.Context, org, gateID string, config *ChangeGateUpdateConfig) (*EnvironmentChangeGate, error) {
	// Marshal eligible approvers to preserve discriminated union structure
	eligibleApprovers := make([]json.RawMessage, len(config.EligibleApprovers))
	for i, wrapper := range config.EligibleApprovers {
		approverBytes, err := json.Marshal(wrapper.Concrete)
		if err != nil {
			return nil, fmt.Errorf("failed to marshal approver %d: %w", i, err)
		}
		eligibleApprovers[i] = json.RawMessage(approverBytes)
	}

	// Build the request structure
	requestStruct := struct {
		Name    string `json:"name"`
		Enabled bool   `json:"enabled"`
		Rule    struct {
			RuleType                  string            `json:"ruleType"`
			NumApprovalsRequired      int64             `json:"numApprovalsRequired"`
			AllowSelfApproval         bool              `json:"allowSelfApproval"`
			RequireReapprovalOnChange bool              `json:"requireReapprovalOnChange"`
			EligibleApprovers         []json.RawMessage `json:"eligibleApprovers"`
		} `json:"rule"`
		Target struct {
			ActionTypes []string `json:"actionTypes"`
		} `json:"target"`
	}{
		Name:    config.Name,
		Enabled: config.Enabled,
	}

	// Set rule fields
	requestStruct.Rule.RuleType = "approval_required"
	requestStruct.Rule.NumApprovalsRequired = config.NumApprovalsRequired
	requestStruct.Rule.AllowSelfApproval = *config.AllowSelfApproval
	requestStruct.Rule.RequireReapprovalOnChange = *config.RequireReapprovalOnChange
	requestStruct.Rule.EligibleApprovers = eligibleApprovers

	// Set target fields
	requestStruct.Target.ActionTypes = []string{"update"}

	// Marshal the complete request structure
	requestJSON, err := json.Marshal(requestStruct)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal request: %w", err)
	}

	return c.updateChangeGateWithRawJSON(ctx, org, gateID, string(requestJSON))
}

// DeleteEnvironmentChangeGate deletes a change gate.
func (c *EscClient) DeleteEnvironmentChangeGate(ctx context.Context, org, gateID string) error {
	_, err := c.EscAPI.DeleteEnvironmentChangeGate(ctx, org, gateID).Execute()
	return err
}

// DeleteChangeGate deletes a change gate (alias for DeleteEnvironmentChangeGate).
func (c *EscClient) DeleteChangeGate(ctx context.Context, org, gateID string) error {
	return c.DeleteEnvironmentChangeGate(ctx, org, gateID)
}

// CreateEnvironmentChangeGate creates a change gate using properly marshaled discriminated unions.
// This method uses structured objects and marshals them to avoid the OpenAPI generator's issues with discriminated union marshaling.
func (c *EscClient) CreateEnvironmentChangeGate(ctx context.Context, org, projectName, envName string, config *ChangeGateConfig) (*EnvironmentChangeGate, error) {
	// Marshal eligible approvers to preserve discriminated union structure
	eligibleApprovers := make([]json.RawMessage, len(config.EligibleApprovers))
	for i, wrapper := range config.EligibleApprovers {
		approverBytes, err := json.Marshal(wrapper.Concrete)
		if err != nil {
			return nil, fmt.Errorf("failed to marshal approver %d: %w", i, err)
		}
		eligibleApprovers[i] = json.RawMessage(approverBytes)
	}

	// Build the request structure
	requestStruct := struct {
		Name    string `json:"name"`
		Enabled bool   `json:"enabled"`
		Rule    struct {
			RuleType                  string            `json:"ruleType"`
			NumApprovalsRequired      int64             `json:"numApprovalsRequired"`
			AllowSelfApproval         bool              `json:"allowSelfApproval"`
			RequireReapprovalOnChange bool              `json:"requireReapprovalOnChange"`
			EligibleApprovers         []json.RawMessage `json:"eligibleApprovers"`
		} `json:"rule"`
		Target struct {
			EntityType     string   `json:"entityType"`
			QualifiedName  string   `json:"qualifiedName"`
			ActionTypes    []string `json:"actionTypes"`
		} `json:"target"`
	}{
		Name:    config.Name,
		Enabled: config.Enabled,
	}

	// Set rule fields
	requestStruct.Rule.RuleType = "approval_required"
	requestStruct.Rule.NumApprovalsRequired = config.NumApprovalsRequired
	requestStruct.Rule.AllowSelfApproval = *config.AllowSelfApproval
	requestStruct.Rule.RequireReapprovalOnChange = *config.RequireReapprovalOnChange
	requestStruct.Rule.EligibleApprovers = eligibleApprovers

	// Set target fields
	requestStruct.Target.EntityType = "environment"
	requestStruct.Target.QualifiedName = getQualifiedEnvName(projectName, envName)
	requestStruct.Target.ActionTypes = []string{"update"}

	// Marshal the complete request structure
	requestJSON, err := json.Marshal(requestStruct)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal request: %w", err)
	}

	return c.createChangeGateWithRawJSON(ctx, org, string(requestJSON))
}

