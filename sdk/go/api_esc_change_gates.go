// Copyright 2024, Pulumi Corporation.  All rights reserved.

package esc_sdk

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
)

// getQualifiedEnvName creates a qualified environment name for API calls.
func getQualifiedEnvName(projectName string, envName string) string {
	qualifiedName := fmt.Sprintf("%s/%s", projectName, envName)
	return qualifiedName
}

// ApprovalRuleEligibilityType constants
const (
	ApprovalRuleEligibilityTypeHasPermissionOnTarget = "has_permission_on_target"
	ApprovalRuleEligibilityTypeHasUserLogin          = "has_user_login"
	ApprovalRuleEligibilityTypeHasTeamMembership     = "has_team_membership"
)

// EnvironmentChangeGatesResponse has Gates with properly deserialized discriminated union fields.
type EnvironmentChangeGatesResponse struct {
	Gates []*EnvironmentChangeGate `json:"gates"`
}

// EligibilityInputWrapper wraps ApprovalRuleEligibilityInput concrete types
// to preserve their type information for proper JSON marshaling
type EligibilityInputWrapper struct {
	Concrete any // Will hold *ApprovalRuleEligibilityInputPermission, *ApprovalRuleEligibilityInputUser, or *ApprovalRuleEligibilityInputTeam
}

// ToApprovalRuleEligibilityInput converts to base type for generated API calls
func (w EligibilityInputWrapper) ToApprovalRuleEligibilityInput() ApprovalRuleEligibilityInput {
	switch concrete := w.Concrete.(type) {
	case *ApprovalRuleEligibilityInputPermission:
		return concrete.ApprovalRuleEligibilityInput
	case *ApprovalRuleEligibilityInputUser:
		return concrete.ApprovalRuleEligibilityInput
	case *ApprovalRuleEligibilityInputTeam:
		return concrete.ApprovalRuleEligibilityInput
	default:
		return ApprovalRuleEligibilityInput{}
	}
}

// ApprovalRuleEligibilityBuilder helps build eligibility rules for change gates.
type ApprovalRuleEligibilityBuilder struct {
	rules []EligibilityInputWrapper
}

// NewApprovalRuleEligibility creates eligibility rules for approvers.
func NewApprovalRuleEligibility() *ApprovalRuleEligibilityBuilder {
	return &ApprovalRuleEligibilityBuilder{
		rules: make([]EligibilityInputWrapper, 0),
	}
}

// AddEnvironmentAdmins adds users with environment:admin permission as eligible approvers.
func (b *ApprovalRuleEligibilityBuilder) AddEnvironmentAdmins() *ApprovalRuleEligibilityBuilder {
	return b.addPermission("environment:admin")
}

// AddEnvironmentWriters adds users with environment:write permission as eligible approvers.
func (b *ApprovalRuleEligibilityBuilder) AddEnvironmentWriters() *ApprovalRuleEligibilityBuilder {
	return b.addPermission("environment:write")
}

// AddCustomPermission adds a custom permission-based eligibility rule.
func (b *ApprovalRuleEligibilityBuilder) AddCustomPermission(permission string) *ApprovalRuleEligibilityBuilder {
	return b.addPermission(permission)
}

// addPermission is the internal method to add permission-based eligibility rules.
func (b *ApprovalRuleEligibilityBuilder) addPermission(permission string) *ApprovalRuleEligibilityBuilder {
	// Create the concrete permission type with the discriminator
	permissionRule := NewApprovalRuleEligibilityInputPermission(permission)
	eligibilityType := ApprovalRuleEligibilityTypeHasPermissionOnTarget
	permissionRule.EligibilityType = &eligibilityType

	// Store in wrapper to preserve concrete type for JSON marshaling
	wrapper := EligibilityInputWrapper{Concrete: permissionRule}
	b.rules = append(b.rules, wrapper)
	return b
}

// AddUser adds a specific user as an eligible approver.
func (b *ApprovalRuleEligibilityBuilder) AddUser(userLogin string) *ApprovalRuleEligibilityBuilder {
	// Create the concrete user type with the discriminator
	userRule := NewApprovalRuleEligibilityInputUser(userLogin)
	eligibilityType := ApprovalRuleEligibilityTypeHasUserLogin
	userRule.EligibilityType = &eligibilityType

	// Store in wrapper to preserve concrete type for JSON marshaling
	wrapper := EligibilityInputWrapper{Concrete: userRule}
	b.rules = append(b.rules, wrapper)
	return b
}

// AddTeam adds a specific team as eligible approvers.
func (b *ApprovalRuleEligibilityBuilder) AddTeam(teamName string) *ApprovalRuleEligibilityBuilder {
	// Create the concrete team type with the discriminator
	teamRule := NewApprovalRuleEligibilityInputTeam(teamName)
	eligibilityType := ApprovalRuleEligibilityTypeHasTeamMembership
	teamRule.EligibilityType = &eligibilityType

	// Store in wrapper to preserve concrete type for JSON marshaling
	wrapper := EligibilityInputWrapper{Concrete: teamRule}
	b.rules = append(b.rules, wrapper)
	return b
}

// Build returns the eligibility rules as ApprovalRuleEligibilityInput for generated API usage.
func (b *ApprovalRuleEligibilityBuilder) Build() []ApprovalRuleEligibilityInput {
	result := make([]ApprovalRuleEligibilityInput, len(b.rules))
	for i, wrapper := range b.rules {
		result[i] = wrapper.ToApprovalRuleEligibilityInput()
	}
	return result
}

// BuildWrappers returns the eligibility rules as wrappers for custom JSON marshaling.
func (b *ApprovalRuleEligibilityBuilder) BuildWrappers() []EligibilityInputWrapper {
	return b.rules
}

// ChangeGateConfig represents the configuration for creating a change gate.
type ChangeGateConfig struct {
	Name                      string
	Enabled                   bool
	NumApprovalsRequired      int64
	EligibleApprovers         []EligibilityInputWrapper
	AllowSelfApproval         *bool
	RequireReapprovalOnChange *bool
}

// NewChangeGateConfig creates a new ChangeGateConfig with required fields and sensible defaults.
//
// Example usage:
//
//	eligibleApprovers := NewApprovalRuleEligibility().
//	    AddEnvironmentWriters().
//	    BuildWrappers()
//
//	config := NewChangeGateConfig("Security Review Gate", true, 2, eligibleApprovers).
//	    WithSelfApproval(false).
//	    WithReapprovalOnChange(true)
//
//	gate, err := client.CreateEnvironmentChangeGate(ctx, org, project, env, config)
func NewChangeGateConfig(name string, enabled bool, numApprovals int64, eligibleApprovers []EligibilityInputWrapper) *ChangeGateConfig {
	return &ChangeGateConfig{
		Name:                      name,
		Enabled:                   enabled,
		NumApprovalsRequired:      numApprovals,
		EligibleApprovers:         eligibleApprovers,
		AllowSelfApproval:         &[]bool{false}[0], // Default to false
		RequireReapprovalOnChange: &[]bool{true}[0],  // Default to true
	}
}

// WithSelfApproval sets whether approvers can approve their own changes.
func (c *ChangeGateConfig) WithSelfApproval(allow bool) *ChangeGateConfig {
	c.AllowSelfApproval = &allow
	return c
}

// WithReapprovalOnChange sets whether approvals are required again when changes are made.
func (c *ChangeGateConfig) WithReapprovalOnChange(require bool) *ChangeGateConfig {
	c.RequireReapprovalOnChange = &require
	return c
}

// EnvironmentChangeGateApprover is the base interface for all approver types.
type EnvironmentChangeGateApprover interface {
	GetEligibilityType() string
}

// EnvironmentChangeGatePermissionApprover represents a permission-based approver.
type EnvironmentChangeGatePermissionApprover struct {
	EligibilityType string `json:"eligibilityType"`
	Permission      string `json:"permission"`
}

func (a *EnvironmentChangeGatePermissionApprover) GetEligibilityType() string {
	return a.EligibilityType
}

// EnvironmentChangeGateUserApprover represents a user-based approver.
type EnvironmentChangeGateUserApprover struct {
	EligibilityType string `json:"eligibilityType"`
	UserLogin       string `json:"userLogin"`
}

func (a *EnvironmentChangeGateUserApprover) GetEligibilityType() string {
	return a.EligibilityType
}

// EnvironmentChangeGateTeamApprover represents a team-based approver.
type EnvironmentChangeGateTeamApprover struct {
	EligibilityType string `json:"eligibilityType"`
	TeamName        string `json:"teamName"`
}

func (a *EnvironmentChangeGateTeamApprover) GetEligibilityType() string {
	return a.EligibilityType
}

// EnvironmentChangeGate extends ChangeGate with easily accessible approval rule fields and eligible approvers.
type EnvironmentChangeGate struct {
	*ChangeGate
	// Approval rule fields for direct access
	NumApprovalsRequired      int64                              `json:"-"`
	AllowSelfApproval         *bool                              `json:"-"`
	RequireReapprovalOnChange *bool                              `json:"-"`
	EligibleApprovers         []EnvironmentChangeGateApprover    `json:"-"`
}

// ChangeGateUpdateConfig holds the configuration for updating a change gate.
type ChangeGateUpdateConfig struct {
	Name                      string
	Enabled                   bool
	NumApprovalsRequired      int64
	EligibleApprovers         []EligibilityInputWrapper
	AllowSelfApproval         *bool
	RequireReapprovalOnChange *bool
}

// NewChangeGateUpdateFromGate creates a new update config starting from an existing ChangeGate.
func NewChangeGateUpdateFromGate(gate *EnvironmentChangeGate) *ChangeGateUpdateConfig {
	return &ChangeGateUpdateConfig{
		Name:                      gate.Name,
		Enabled:                   gate.Enabled,
		NumApprovalsRequired:      gate.NumApprovalsRequired,
		EligibleApprovers:         []EligibilityInputWrapper{}, // Will be set by WithApprovalRule
		AllowSelfApproval:         gate.AllowSelfApproval,
		RequireReapprovalOnChange: gate.RequireReapprovalOnChange,
	}
}

// WithName sets the gate name.
func (c *ChangeGateUpdateConfig) WithName(name string) *ChangeGateUpdateConfig {
	c.Name = name
	return c
}

// WithEnabled sets whether the gate is enabled.
func (c *ChangeGateUpdateConfig) WithEnabled(enabled bool) *ChangeGateUpdateConfig {
	c.Enabled = enabled
	return c
}

// WithApprovalRule sets the approval requirements.
func (c *ChangeGateUpdateConfig) WithApprovalRule(numApprovals int64, eligibleApprovers []EligibilityInputWrapper) *ChangeGateUpdateConfig {
	c.NumApprovalsRequired = numApprovals
	c.EligibleApprovers = eligibleApprovers
	return c
}

// WithSelfApproval sets whether approvers can approve their own changes.
func (c *ChangeGateUpdateConfig) WithSelfApproval(allow bool) *ChangeGateUpdateConfig {
	c.AllowSelfApproval = &allow
	return c
}

// WithReapprovalOnChange sets whether approvals are required again when changes are made.
func (c *ChangeGateUpdateConfig) WithReapprovalOnChange(require bool) *ChangeGateUpdateConfig {
	c.RequireReapprovalOnChange = &require
	return c
}

// Helper functions for creating eligibility rules

// CreatePermissionEligibilityRule creates a permission-based eligibility rule.
func CreatePermissionEligibilityRule(permission string) *ApprovalRuleEligibilityInputPermission {
	rule := NewApprovalRuleEligibilityInputPermission(permission)
	eligibilityType := ApprovalRuleEligibilityTypeHasPermissionOnTarget
	rule.EligibilityType = &eligibilityType
	return rule
}

// CreateUserEligibilityRule creates a user-based eligibility rule.
func CreateUserEligibilityRule(userLogin string) *ApprovalRuleEligibilityInputUser {
	rule := NewApprovalRuleEligibilityInputUser(userLogin)
	eligibilityType := ApprovalRuleEligibilityTypeHasUserLogin
	rule.EligibilityType = &eligibilityType
	return rule
}

// CreateTeamEligibilityRule creates a team-based eligibility rule.
func CreateTeamEligibilityRule(teamName string) *ApprovalRuleEligibilityInputTeam {
	rule := NewApprovalRuleEligibilityInputTeam(teamName)
	eligibilityType := ApprovalRuleEligibilityTypeHasTeamMembership
	rule.EligibilityType = &eligibilityType
	return rule
}

// createChangeGateWithRawJSON makes the HTTP request with raw JSON for creation
func (c *EscClient) createChangeGateWithRawJSON(ctx context.Context, org, jsonBody string) (*EnvironmentChangeGate, error) {
	// Build the URL using the same method as the generated code
	basePath, err := c.rawClient.cfg.ServerURLWithContext(ctx, "EscAPIService.CreateEnvironmentChangeGate")
	if err != nil {
		return nil, fmt.Errorf("failed to get server URL: %w", err)
	}
	requestURL := fmt.Sprintf("%s/change-gates/%s", basePath, url.PathEscape(org))

	// Create HTTP request with raw JSON body
	httpReq, err := http.NewRequestWithContext(ctx, "POST", requestURL, strings.NewReader(jsonBody))
	if err != nil {
		return nil, fmt.Errorf("failed to create HTTP request: %w", err)
	}

	// Set headers
	httpReq.Header.Set("Content-Type", "application/json")
	httpReq.Header.Set("Accept", "application/json")
	httpReq.Header.Set("User-Agent", c.rawClient.cfg.UserAgent)

	// Add authentication
	if auth, ok := ctx.Value(ContextAPIKeys).(map[string]APIKey); ok {
		if apiKey, exists := auth["Authorization"]; exists {
			httpReq.Header.Set("Authorization", fmt.Sprintf("%s %s", apiKey.Prefix, apiKey.Key))
		}
	}

	// Make the request
	resp, err := c.rawClient.cfg.HTTPClient.Do(httpReq)
	if err != nil {
		return nil, fmt.Errorf("failed to make HTTP request: %w", err)
	}
	defer resp.Body.Close()

	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response body: %w", err)
	}

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return nil, fmt.Errorf("HTTP request failed with status %d: %s", resp.StatusCode, string(respBody))
	}

	// Parse into generated ChangeGate struct first
	var changeGate ChangeGate
	err = json.Unmarshal(respBody, &changeGate)
	if err != nil {
		return nil, fmt.Errorf("failed to unmarshal base response: %w", err)
	}

	// Then enhance with discriminated union parsing
	enhanced, err := c.parseEnhancedGateFromJSON(respBody, &changeGate)
	if err != nil {
		return nil, fmt.Errorf("failed to parse enhanced gate: %w", err)
	}

	return enhanced, nil
}

// updateChangeGateWithRawJSON makes the HTTP request with raw JSON for updating
func (c *EscClient) updateChangeGateWithRawJSON(ctx context.Context, org, gateID string, jsonBody string) (*EnvironmentChangeGate, error) {

	// Build the URL using the same method as the generated code
	basePath, err := c.rawClient.cfg.ServerURLWithContext(ctx, "EscAPIService.UpdateEnvironmentChangeGate")
	if err != nil {
		return nil, fmt.Errorf("failed to get server URL: %w", err)
	}
	requestURL := fmt.Sprintf("%s/change-gates/%s/%s", basePath,
		url.PathEscape(org), url.PathEscape(gateID))

	// Create HTTP request with raw JSON body
	httpReq, err := http.NewRequestWithContext(ctx, "PUT", requestURL, strings.NewReader(jsonBody))
	if err != nil {
		return nil, fmt.Errorf("failed to create HTTP request: %w", err)
	}

	// Set headers
	httpReq.Header.Set("Content-Type", "application/json")
	httpReq.Header.Set("Accept", "application/json")
	httpReq.Header.Set("User-Agent", c.rawClient.cfg.UserAgent)

	// Add authentication
	if auth, ok := ctx.Value(ContextAPIKeys).(map[string]APIKey); ok {
		if apiKey, exists := auth["Authorization"]; exists {
			httpReq.Header.Set("Authorization", fmt.Sprintf("%s %s", apiKey.Prefix, apiKey.Key))
		}
	}

	// Make the request
	resp, err := c.rawClient.cfg.HTTPClient.Do(httpReq)
	if err != nil {
		return nil, fmt.Errorf("failed to make HTTP request: %w", err)
	}
	defer resp.Body.Close()

	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response body: %w", err)
	}

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return nil, fmt.Errorf("HTTP request failed with status %d: %s", resp.StatusCode, string(respBody))
	}

	// Parse into generated ChangeGate struct first
	var changeGate ChangeGate
	err = json.Unmarshal(respBody, &changeGate)
	if err != nil {
		return nil, fmt.Errorf("failed to unmarshal base response: %w", err)
	}

	// Then enhance with discriminated union parsing
	enhanced, err := c.parseEnhancedGateFromJSON(respBody, &changeGate)
	if err != nil {
		return nil, fmt.Errorf("failed to parse enhanced gate: %w", err)
	}

	return enhanced, nil
}

// unmarshalChangeGateRule extracts the approval rule details from ChangeGateRuleOutput.
// This converts the base ChangeGateRuleOutput to an EnvironmentChangeGate so
// clients can access rule fields directly without manual type assertions.
// Note: This method provides limited parsing and is mostly used as fallback.
// For full discriminated union parsing, use parseEnhancedGateFromJSON instead.
func (c *EscClient) unmarshalChangeGateRule(gate *ChangeGate) (*EnvironmentChangeGate, error) {
	// This method provides basic enhancement without discriminated union parsing
	// since ChangeGateRuleOutput doesn't have direct field access to the discriminated union fields
	enhanced := &EnvironmentChangeGate{
		ChangeGate:                gate,
		NumApprovalsRequired:      1,                                 // Default value
		AllowSelfApproval:         nil,                               // Unknown without raw JSON parsing
		RequireReapprovalOnChange: nil,                               // Unknown without raw JSON parsing
		EligibleApprovers:         []EnvironmentChangeGateApprover{}, // Empty - requires raw JSON parsing
	}

	return enhanced, nil
}

// parseEnhancedGateFromJSON extracts approval rule details directly from JSON response.
// This works around the issue where generated structs lose discriminated union fields.
func (c *EscClient) parseEnhancedGateFromJSON(jsonData []byte, gate *ChangeGate) (*EnvironmentChangeGate, error) {
	// Parse the raw JSON to extract rule fields
	var responseData struct {
		Rule struct {
			NumApprovalsRequired      int64             `json:"numApprovalsRequired"`
			AllowSelfApproval         *bool             `json:"allowSelfApproval"`
			RequireReapprovalOnChange *bool             `json:"requireReapprovalOnChange"`
			EligibleApprovers         []json.RawMessage `json:"eligibleApprovers"`
		} `json:"rule"`
	}

	err := json.Unmarshal(jsonData, &responseData)
	if err != nil {
		return nil, fmt.Errorf("failed to parse response data: %w", err)
	}

	// Parse eligible approvers as concrete types based on eligibilityType
	eligibleApprovers := make([]EnvironmentChangeGateApprover, len(responseData.Rule.EligibleApprovers))
	for i, approverData := range responseData.Rule.EligibleApprovers {
		// First parse to determine the eligibilityType
		var baseApprover struct {
			EligibilityType string `json:"eligibilityType"`
		}
		err := json.Unmarshal(approverData, &baseApprover)
		if err != nil {
			return nil, fmt.Errorf("failed to parse approver %d eligibilityType: %w", i, err)
		}

		// Create the appropriate concrete type based on eligibilityType
		switch baseApprover.EligibilityType {
		case ApprovalRuleEligibilityTypeHasPermissionOnTarget:
			var permissionApprover EnvironmentChangeGatePermissionApprover
			err := json.Unmarshal(approverData, &permissionApprover)
			if err != nil {
				return nil, fmt.Errorf("failed to parse permission approver %d: %w", i, err)
			}
			eligibleApprovers[i] = &permissionApprover
		case ApprovalRuleEligibilityTypeHasUserLogin:
			var userApprover EnvironmentChangeGateUserApprover
			err := json.Unmarshal(approverData, &userApprover)
			if err != nil {
				return nil, fmt.Errorf("failed to parse user approver %d: %w", i, err)
			}
			eligibleApprovers[i] = &userApprover
		case ApprovalRuleEligibilityTypeHasTeamMembership:
			var teamApprover EnvironmentChangeGateTeamApprover
			err := json.Unmarshal(approverData, &teamApprover)
			if err != nil {
				return nil, fmt.Errorf("failed to parse team approver %d: %w", i, err)
			}
			eligibleApprovers[i] = &teamApprover
		default:
			return nil, fmt.Errorf("unknown eligibilityType '%s' for approver %d", baseApprover.EligibilityType, i)
		}
	}

	enhanced := &EnvironmentChangeGate{
		ChangeGate:                gate,
		NumApprovalsRequired:      responseData.Rule.NumApprovalsRequired,
		AllowSelfApproval:         responseData.Rule.AllowSelfApproval,
		RequireReapprovalOnChange: responseData.Rule.RequireReapprovalOnChange,
		EligibleApprovers:         eligibleApprovers,
	}

	return enhanced, nil
}

// getChangeGateWithRawJSON makes the HTTP request with raw JSON response parsing
func (c *EscClient) getChangeGateWithRawJSON(ctx context.Context, org, gateID string) (*EnvironmentChangeGate, error) {
	// Build the URL using the same method as the generated code
	basePath, err := c.rawClient.cfg.ServerURLWithContext(ctx, "EscAPIService.GetEnvironmentChangeGate")
	if err != nil {
		return nil, fmt.Errorf("failed to get server URL: %w", err)
	}
	requestURL := fmt.Sprintf("%s/change-gates/%s/%s", basePath,
		url.PathEscape(org), url.PathEscape(gateID))

	// Create HTTP request
	httpReq, err := http.NewRequestWithContext(ctx, "GET", requestURL, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create HTTP request: %w", err)
	}

	// Set headers
	httpReq.Header.Set("Accept", "application/json")
	httpReq.Header.Set("User-Agent", c.rawClient.cfg.UserAgent)

	// Add authentication
	if auth, ok := ctx.Value(ContextAPIKeys).(map[string]APIKey); ok {
		if apiKey, exists := auth["Authorization"]; exists {
			httpReq.Header.Set("Authorization", fmt.Sprintf("%s %s", apiKey.Prefix, apiKey.Key))
		}
	}

	// Make the request
	resp, err := c.rawClient.cfg.HTTPClient.Do(httpReq)
	if err != nil {
		return nil, fmt.Errorf("failed to make HTTP request: %w", err)
	}
	defer resp.Body.Close()

	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response body: %w", err)
	}

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return nil, fmt.Errorf("HTTP request failed with status %d: %s", resp.StatusCode, string(respBody))
	}

	// Parse into generated ChangeGate struct first
	var changeGate ChangeGate
	err = json.Unmarshal(respBody, &changeGate)
	if err != nil {
		return nil, fmt.Errorf("failed to unmarshal base response: %w", err)
	}

	// Then enhance with discriminated union parsing
	enhanced, err := c.parseEnhancedGateFromJSON(respBody, &changeGate)
	if err != nil {
		return nil, fmt.Errorf("failed to parse enhanced gate: %w", err)
	}

	return enhanced, nil
}

// listEnvironmentChangeGatesWithRawJSON makes the HTTP request for listing change gates with raw JSON response.
func (c *EscClient) listEnvironmentChangeGatesWithRawJSON(ctx context.Context, org, projectName, envName string) (*EnvironmentChangeGatesResponse, error) {
	// Build the URL using the same method as the generated code
	basePath, err := c.rawClient.cfg.ServerURLWithContext(ctx, "EscAPIService.ListEnvironmentChangeGates")
	if err != nil {
		return nil, fmt.Errorf("failed to get server URL: %w", err)
	}

	// Build query parameters
	qualifiedName := getQualifiedEnvName(projectName, envName)
	requestURL := fmt.Sprintf("%s/change-gates/%s?entityType=environment&qualifiedName=%s",
		basePath, url.PathEscape(org), url.QueryEscape(qualifiedName))

	// Create HTTP request
	httpReq, err := http.NewRequestWithContext(ctx, "GET", requestURL, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create HTTP request: %w", err)
	}

	// Set headers
	httpReq.Header.Set("Accept", "application/json")
	httpReq.Header.Set("User-Agent", c.rawClient.cfg.UserAgent)

	// Add authentication
	if auth, ok := ctx.Value(ContextAPIKeys).(map[string]APIKey); ok {
		if apiKey, exists := auth["Authorization"]; exists {
			httpReq.Header.Set("Authorization", fmt.Sprintf("%s %s", apiKey.Prefix, apiKey.Key))
		}
	}

	// Make the request
	resp, err := c.rawClient.cfg.HTTPClient.Do(httpReq)
	if err != nil {
		return nil, fmt.Errorf("failed to make HTTP request: %w", err)
	}
	defer resp.Body.Close()

	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response body: %w", err)
	}

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return nil, fmt.Errorf("HTTP request failed with status %d: %s", resp.StatusCode, string(respBody))
	}

	// Parse the response as raw JSON to preserve discriminated union fields
	var rawResponse struct {
		Gates []json.RawMessage `json:"gates"`
	}
	err = json.Unmarshal(respBody, &rawResponse)
	if err != nil {
		return nil, fmt.Errorf("failed to unmarshal raw response: %w", err)
	}

	// Also parse into the generated struct for fallback
	var listResponse ListChangeGatesResponse
	err = json.Unmarshal(respBody, &listResponse)
	if err != nil {
		return nil, fmt.Errorf("failed to unmarshal list response: %w", err)
	}

	// Create enhanced gates by parsing each gate's raw JSON
	enhancedGates := make([]*EnvironmentChangeGate, len(listResponse.Gates))
	for i, gate := range listResponse.Gates {
		if i < len(rawResponse.Gates) {
			// Use raw JSON for enhanced parsing
			enhanced, err := c.parseEnhancedGateFromJSON(rawResponse.Gates[i], &gate)
			if err != nil {
				return nil, fmt.Errorf("failed to parse enhanced gate %d: %w", i, err)
			}
			enhancedGates[i] = enhanced
		} else {
			// Fallback to basic enhancement if raw JSON is not available
			enhanced, err := c.unmarshalChangeGateRule(&gate)
			if err != nil {
				return nil, fmt.Errorf("failed to enhance gate %d: %w", i, err)
			}
			enhancedGates[i] = enhanced
		}
	}

	return &EnvironmentChangeGatesResponse{
		Gates: enhancedGates,
	}, nil
}