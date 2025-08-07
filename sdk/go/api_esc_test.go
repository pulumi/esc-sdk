// Copyright 2024, Pulumi Corporation.  All rights reserved.

/*
ESC (Environments, Secrets, Config) API

Testing EscAPIService

*/

package esc_sdk

import (
	"context"
	"fmt"
	"os"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
)

const PROJECT_NAME = "sdk-go-test"
const CLONE_PROJECT_NAME = "sdk-go-test-clone"
const ENV_PREFIX = "env-"
const BASE_ENV_PREFIX = "base-"

func Test_EscClient(t *testing.T) {
	orgName := os.Getenv("PULUMI_ORG")
	require.NotEmpty(t, orgName, "PULUMI_ORG must be set")
	auth, apiClient, err := DefaultLogin()
	require.NoError(t, err)

	removeAllGoTestEnvs(t, apiClient, auth, orgName)

	baseEnvName := "base-" + time.Now().Format("20060102150405")
	err = apiClient.CreateEnvironment(auth, orgName, PROJECT_NAME, baseEnvName)
	require.Nil(t, err)
	t.Cleanup(func() {
		err := apiClient.DeleteEnvironment(auth, orgName, PROJECT_NAME, baseEnvName)
		require.Nil(t, err)
	})

	baseEnv := &EnvironmentDefinition{
		Values: &EnvironmentDefinitionValues{
			AdditionalProperties: map[string]any{
				"base": baseEnvName,
			},
		},
	}

	_, err = apiClient.UpdateEnvironment(auth, orgName, PROJECT_NAME, baseEnvName, baseEnv)
	require.Nil(t, err)

	t.Run("should create, clone, list, update, get, decrypt, open and delete an environment", func(t *testing.T) {
		envName := ENV_PREFIX + time.Now().Format("20060102150405")
		err := apiClient.CreateEnvironment(auth, orgName, PROJECT_NAME, envName)
		require.Nil(t, err)

		cloneProject := fmt.Sprintf("%s-clone", PROJECT_NAME)
		cloneName := fmt.Sprintf("%s-clone", envName)
		err = apiClient.CloneEnvironment(auth, orgName, PROJECT_NAME, envName, cloneProject, cloneName, &CloneEnvironmentOptions{})
		require.Nil(t, err)

		t.Cleanup(func() {
			err := apiClient.DeleteEnvironment(auth, orgName, PROJECT_NAME, envName)
			require.Nil(t, err)
			err = apiClient.DeleteEnvironment(auth, orgName, cloneProject, cloneName)
			require.Nil(t, err)
		})

		envs, err := apiClient.ListEnvironments(auth, orgName, nil)
		require.Nil(t, err)

		requireFindEnvironment(t, envs, PROJECT_NAME, envName)
		requireFindEnvironment(t, envs, cloneProject, cloneName)

		_, values, err := apiClient.OpenAndReadEnvironment(auth, orgName, PROJECT_NAME, envName)
		require.Nil(t, err)
		var nilValues map[string]any = nil
		require.Equal(t, values, nilValues)

		yaml := "imports:\n  - " + PROJECT_NAME + "/" + baseEnvName + "\n" + `
values:
  foo: bar
  my_secret:
    fn::secret: "shh! don't tell anyone"
  my_array: [1, 2, 3]
  pulumiConfig:
    foo: ${foo}
  environmentVariables:
    FOO: ${foo}
`

		diags, err := apiClient.UpdateEnvironmentYaml(auth, orgName, PROJECT_NAME, envName, yaml)
		require.Nil(t, err)
		require.NotNil(t, diags)
		require.Len(t, diags.Diagnostics, 0)

		env, newYaml, err := apiClient.GetEnvironment(auth, orgName, PROJECT_NAME, envName)
		require.Nil(t, err)
		require.NotNil(t, newYaml)

		assertEnvDef(t, env, baseEnvName)

		require.NotNil(t, env.Values.AdditionalProperties["my_secret"].(map[string]any))

		decryptEnv, _, err := apiClient.DecryptEnvironment(auth, orgName, PROJECT_NAME, envName)
		require.Nil(t, err)

		assertEnvDef(t, decryptEnv, baseEnvName)

		mySecret, ok := decryptEnv.Values.AdditionalProperties["my_secret"].(map[string]any)
		require.True(t, ok)
		require.Equal(t, "shh! don't tell anyone", mySecret["fn::secret"])

		_, values, err = apiClient.OpenAndReadEnvironment(auth, orgName, PROJECT_NAME, envName)
		require.Nil(t, err)

		require.Equal(t, baseEnvName, values["base"])
		require.Equal(t, "bar", values["foo"])
		require.Equal(t, []any{1.0, 2.0, 3.0}, values["my_array"])
		require.Equal(t, "shh! don't tell anyone", values["my_secret"])
		pulumiConfig, ok := values["pulumiConfig"].(map[string]any)
		require.True(t, ok)
		require.Equal(t, "bar", pulumiConfig["foo"])
		environmentVariables, ok := values["environmentVariables"].(map[string]any)
		require.True(t, ok)
		require.Equal(t, "bar", environmentVariables["FOO"])

		openInfo, err := apiClient.OpenEnvironment(auth, orgName, PROJECT_NAME, envName)
		require.Nil(t, err)

		v, value, err := apiClient.ReadEnvironmentProperty(auth, orgName, PROJECT_NAME, envName, openInfo.Id, "pulumiConfig.foo")
		require.Nil(t, err)
		require.Equal(t, "bar", v.Value)
		require.Equal(t, "bar", value)

		env, _, err = apiClient.GetEnvironmentAtVersion(auth, orgName, PROJECT_NAME, envName, "2")
		require.Nil(t, err)

		env.Values.AdditionalProperties["versioned"] = "true"

		_, err = apiClient.UpdateEnvironment(auth, orgName, PROJECT_NAME, envName, env)
		require.Nil(t, err)

		revisions, err := apiClient.ListEnvironmentRevisions(auth, orgName, PROJECT_NAME, envName)
		require.Nil(t, err)
		require.NotNil(t, revisions)
		require.Len(t, revisions, 3)

		err = apiClient.CreateEnvironmentRevisionTag(auth, orgName, PROJECT_NAME, envName, "testTag", 2)
		require.Nil(t, err)

		_, values, err = apiClient.OpenAndReadEnvironmentAtVersion(auth, orgName, PROJECT_NAME, envName, "testTag")
		require.Nil(t, err)
		_, ok = values["versioned"]
		require.Equal(t, ok, false)

		tags, err := apiClient.ListEnvironmentRevisionTags(auth, orgName, PROJECT_NAME, envName)
		require.Nil(t, err)
		require.NotNil(t, tags)
		require.Len(t, tags.Tags, 2)
		require.Equal(t, "latest", tags.Tags[0].Name)
		require.Equal(t, "testTag", tags.Tags[1].Name)

		err = apiClient.UpdateEnvironmentRevisionTag(auth, orgName, PROJECT_NAME, envName, "testTag", 3)
		require.Nil(t, err)

		_, values, err = apiClient.OpenAndReadEnvironmentAtVersion(auth, orgName, PROJECT_NAME, envName, "testTag")
		require.Nil(t, err)
		require.Equal(t, "true", values["versioned"])

		testTag, err := apiClient.GetEnvironmentRevisionTag(auth, orgName, PROJECT_NAME, envName, "testTag")
		require.Nil(t, err)
		require.NotNil(t, testTag)
		require.Equal(t, int32(3), testTag.Revision)

		err = apiClient.DeleteEnvironmentRevisionTag(auth, orgName, PROJECT_NAME, envName, "testTag")
		require.Nil(t, err)

		tags, err = apiClient.ListEnvironmentRevisionTags(auth, orgName, PROJECT_NAME, envName)
		require.Nil(t, err)
		require.NotNil(t, tags)
		require.Len(t, tags.Tags, 1)

		_, err = apiClient.CreateEnvironmentTag(auth, orgName, PROJECT_NAME, envName, "owner", "esc-sdk-test")
		require.Nil(t, err)

		envTags, err := apiClient.ListEnvironmentTags(auth, orgName, PROJECT_NAME, envName)
		require.Nil(t, err)
		require.NotNil(t, envTags)
		require.Len(t, envTags.Tags, 1)
		require.Equal(t, "owner", envTags.Tags["owner"].Name)
		require.Equal(t, "esc-sdk-test", *envTags.Tags["owner"].Value)

		_, err = apiClient.UpdateEnvironmentTag(auth, orgName, PROJECT_NAME, envName, "owner", "esc-sdk-test", "new-owner", "esc-sdk-test-updated")
		require.Nil(t, err)

		envTag, err := apiClient.GetEnvironmentTag(auth, orgName, PROJECT_NAME, envName, "new-owner")
		require.Nil(t, err)
		require.NotNil(t, envTag)
		require.Equal(t, "new-owner", envTag.Name)
		require.Equal(t, "esc-sdk-test-updated", *envTag.Value)

		err = apiClient.DeleteEnvironmentTag(auth, orgName, PROJECT_NAME, envName, "new-owner")
		require.Nil(t, err)

		envTags, err = apiClient.ListEnvironmentTags(auth, orgName, PROJECT_NAME, envName)
		require.Nil(t, err)
		require.NotNil(t, envTags)
		require.Len(t, envTags.Tags, 0)
	})

	t.Run("check environment definition valid", func(t *testing.T) {
		env := &EnvironmentDefinition{
			Values: &EnvironmentDefinitionValues{
				AdditionalProperties: map[string]any{
					"foo": "bar",
				},
			},
		}

		diags, err := apiClient.CheckEnvironment(auth, orgName, env)
		require.Nil(t, err)
		require.NotNil(t, diags)
		require.Len(t, diags.Diagnostics, 0)
	})

	t.Run("check environment yaml invalid", func(t *testing.T) {
		yaml := `
values:
  foo: bar
  pulumiConfig:
    foo: ${bad_ref}
`
		diags, err := apiClient.CheckEnvironmentYaml(auth, orgName, yaml)
		require.Error(t, err, "400 Bad Request")
		require.NotNil(t, diags)
		require.Len(t, diags.Diagnostics, 1)
		require.Equal(t, "unknown property \"bad_ref\"", diags.Diagnostics[0].Summary)

	})

	t.Run("change management workflow - gates, drafts, approvals, and apply", func(t *testing.T) {
		envName := ENV_PREFIX + "changmgmt-" + time.Now().Format("20060102150405")
		err := apiClient.CreateEnvironment(auth, orgName, PROJECT_NAME, envName)
		require.Nil(t, err)

		t.Cleanup(func() {
			err := apiClient.DeleteEnvironment(auth, orgName, PROJECT_NAME, envName)
			require.Nil(t, err)
		})

		// Step 1: List change gates (should be empty initially)
		gates, err := apiClient.ListEnvironmentChangeGates(auth, orgName, PROJECT_NAME, envName)
		require.Nil(t, err)
		require.NotNil(t, gates)
		require.Len(t, gates.Gates, 0)
		t.Logf("Initial gates count: %d", len(gates.Gates))

		// Step 2: Create a change gate using our wrapper-based discriminated union fix
		eligibleApprovers := NewApprovalRuleEligibility().
			AddEnvironmentWriters(). // Users with environment:write permission can approve
			BuildWrappers()          // Use wrappers that preserve concrete types for proper JSON marshaling

		// Create change gate config with the new structured approach
		gateConfig := NewChangeGateConfig("Test Security Gate", true, 1, eligibleApprovers).
			WithSelfApproval(true).       // Explicitly disable self approval
			WithReapprovalOnChange(false) // Don't require reapproval on changes

		// Use our working create method that handles discriminated unions correctly
		createdGate, err := apiClient.CreateEnvironmentChangeGate(auth, orgName, PROJECT_NAME, envName, gateConfig)
		require.Nil(t, err)
		require.NotNil(t, createdGate)
		require.Equal(t, "Test Security Gate", createdGate.Name)
		require.True(t, createdGate.Enabled)
		require.True(t, *createdGate.AllowSelfApproval)
		require.False(t, *createdGate.RequireReapprovalOnChange)
		gateID := createdGate.Id
		t.Logf("Created gate with ID: %s", gateID)

		// Step 3: List change gates again (should have 1 gate now)
		gates, err = apiClient.ListEnvironmentChangeGates(auth, orgName, PROJECT_NAME, envName)
		require.Nil(t, err)
		require.NotNil(t, gates)
		require.Len(t, gates.Gates, 1)
		require.Equal(t, gateID, gates.Gates[0].Id)
		require.Equal(t, "Test Security Gate", gates.Gates[0].Name)

		// Test gates - verify discriminated union fields are properly deserialized
		listedGate := gates.Gates[0]
		require.Equal(t, gateID, listedGate.Id)
		require.Equal(t, "Test Security Gate", listedGate.Name)
		require.Equal(t, int64(1), listedGate.NumApprovalsRequired, "Listed gate should require 1 approval")
		require.True(t, *listedGate.AllowSelfApproval, "Listed gate should allow self approval")
		require.False(t, *listedGate.RequireReapprovalOnChange, "Listed gate should not require reapproval on change")

		// Test eligible approvers deserialization
		require.Len(t, listedGate.EligibleApprovers, 1, "Listed gate should have 1 eligible approver")
		approver := listedGate.EligibleApprovers[0]
		require.Equal(t, "has_permission_on_target", approver.GetEligibilityType(), "Approver should be permission-based")

		// Cast to concrete permission approver type
		permissionApprover, ok := approver.(*EnvironmentChangeGatePermissionApprover)
		require.True(t, ok, "Approver should be a EnvironmentChangeGatePermissionApprover")
		require.Equal(t, "environment:write", permissionApprover.Permission, "Approver should require environment:write permission")

		t.Logf("Gates count after creation: %d", len(gates.Gates))
		t.Logf("Listed gate - NumApprovals: %d, AllowSelfApproval: %v, RequireReapprovalOnChange: %v",
			listedGate.NumApprovalsRequired, *listedGate.AllowSelfApproval, *listedGate.RequireReapprovalOnChange)
		t.Logf("Listed gate approver - Type: %s, Permission: %s",
			approver.GetEligibilityType(), permissionApprover.Permission)

		// Step 4: Get the individual gate
		gate, err := apiClient.GetChangeGate(auth, orgName, gateID)
		require.Nil(t, err)
		require.NotNil(t, gate)
		require.Equal(t, gateID, gate.Id)
		require.Equal(t, "Test Security Gate", gate.Name)
		t.Logf("Retrieved gate: %s", gate.Name)

		// Step 4.1: Verify discriminated union response unmarshaling works
		// This verifies our OpenAPI template fix (removing DisallowUnknownFields) worked correctly

		// Verify basic discriminator field is present
		require.NotNil(t, gate.Rule.RuleType)
		require.Equal(t, ChangeGateRuleTypeApprovalRequired, *gate.Rule.RuleType)
		t.Logf("✅ Rule discriminator verified: ruleType=%s", *gate.Rule.RuleType)

		// Verify target discriminator
		require.Equal(t, ChangeGateTargetActionTypeUpdate, gate.Target.ActionTypes[0])
		t.Logf("✅ Target discriminator verified: actionType=%s", gate.Target.ActionTypes[0])

		// Log the actual types we received to understand response structure
		t.Logf("✅ Response types - Rule: %T, Target: %T", gate.Rule, gate.Target)

		// Most importantly: the response was successfully unmarshaled without unknown field errors
		// This proves our discriminated union marshaling fix works end-to-end
		t.Logf("✅ Discriminated union round-trip successful: request marshaling ✓, response unmarshaling ✓")

		// Step 5: Create an environment draft (change request)
		draftYaml := `values:
  foo: bar
  environment: test
  changeManagement: enabled
`
		draftRef, err := apiClient.CreateEnvironmentDraft(auth, orgName, PROJECT_NAME, envName, draftYaml)
		require.Nil(t, err)
		require.NotNil(t, draftRef)
		require.NotNil(t, draftRef.ChangeRequestId)
		require.NotEmpty(t, *draftRef.ChangeRequestId)
		changeRequestID := *draftRef.ChangeRequestId
		t.Logf("Created draft with change request ID: %s", changeRequestID)

		// Step 6: Get the change request to test that gate is not met
		changeRequest, err := apiClient.GetChangeRequest(auth, orgName, changeRequestID)
		require.Nil(t, err)
		require.NotNil(t, changeRequest)
		require.NotNil(t, changeRequest.GateEvaluation)
		require.False(t, changeRequest.GateEvaluation.Satisfied, "Gate should not be satisfied initially")
		require.Len(t, changeRequest.GateEvaluation.ApplicableGates, 1)
		require.False(t, changeRequest.GateEvaluation.ApplicableGates[0].Satisfied)
		t.Logf("Change request gate evaluation - satisfied: %v", changeRequest.GateEvaluation.Satisfied)

		// Step 7: Submit the change request for approval
		err = apiClient.SubmitChangeRequest(auth, orgName, changeRequestID, "Ready for review - adding test values")
		require.Nil(t, err)
		t.Logf("Submitted change request for approval")

		// Step 8: Approve the change request
		err = apiClient.ApproveChangeRequest(auth, orgName, changeRequestID, int64(0), nil)
		require.Nil(t, err)
		t.Logf("Approved change request")

		// Step 9: Test that the gate is now met
		changeRequest, err = apiClient.GetChangeRequest(auth, orgName, changeRequestID)
		require.Nil(t, err)
		require.NotNil(t, changeRequest.GateEvaluation)
		require.True(t, changeRequest.GateEvaluation.Satisfied, "Gate should be satisfied after approval")
		require.Len(t, changeRequest.GateEvaluation.ApplicableGates, 1)
		require.True(t, changeRequest.GateEvaluation.ApplicableGates[0].Satisfied)
		t.Logf("After approval - gate evaluation satisfied: %v", changeRequest.GateEvaluation.Satisfied)

		// Step 10: Change the gate to make it harder (require 2 approvals)
		updateConfig := NewChangeGateUpdateFromGate(gate).
			WithName("Test Security Gate - Stricter").
			WithApprovalRule(2, eligibleApprovers) // Now require 2 approvals

		updatedGate, err := apiClient.UpdateChangeGate(auth, orgName, gateID, updateConfig)
		require.Nil(t, err)
		require.NotNil(t, updatedGate)
		require.Equal(t, "Test Security Gate - Stricter", updatedGate.Name)
		require.Equal(t, int64(2), updatedGate.NumApprovalsRequired, "Gate should require 2 approvals")
		require.True(t, *createdGate.AllowSelfApproval)
		require.False(t, *createdGate.RequireReapprovalOnChange)
		t.Logf("Updated gate to require 2 approvals - verified: %d", updatedGate.NumApprovalsRequired)

		// Step 11: Test that the gate is not met again (1 approval < 2 required)
		changeRequest, err = apiClient.GetChangeRequest(auth, orgName, changeRequestID)
		require.Nil(t, err)
		require.NotNil(t, changeRequest.GateEvaluation)
		require.False(t, changeRequest.GateEvaluation.Satisfied, "Gate should not be satisfied with stricter requirements")
		t.Logf("After stricter gate - satisfied: %v", changeRequest.GateEvaluation.Satisfied)

		// Step 12: Update gate back to easier version (require 1 approval)
		restoreConfig := NewChangeGateUpdateFromGate(updatedGate).
			WithName("Test Security Gate - Restored").
			WithApprovalRule(1, eligibleApprovers) // Back to 1 approval

		updatedGate, err = apiClient.UpdateChangeGate(auth, orgName, gateID, restoreConfig)
		require.Nil(t, err)
		require.Equal(t, "Test Security Gate - Restored", updatedGate.Name)
		require.Equal(t, int64(1), updatedGate.NumApprovalsRequired, "Gate should require 1 approval")
		require.True(t, *createdGate.AllowSelfApproval)
		require.False(t, *createdGate.RequireReapprovalOnChange)
		t.Logf("Restored gate to require 1 approval - verified: %d", updatedGate.NumApprovalsRequired)

		// Step 13: Remove approval to test gate becomes unsatisfied
		err = apiClient.UnapproveChangeRequest(auth, orgName, changeRequestID, nil)
		require.Nil(t, err)
		t.Logf("Removed approval from change request")

		// Step 14: Test that the gate is not met after removing approval
		changeRequest, err = apiClient.GetChangeRequest(auth, orgName, changeRequestID)
		require.Nil(t, err)
		require.NotNil(t, changeRequest.GateEvaluation)
		require.False(t, changeRequest.GateEvaluation.Satisfied, "Gate should not be satisfied after removing approval")
		t.Logf("After removing approval - satisfied: %v", changeRequest.GateEvaluation.Satisfied)

		// Step 15: Re-approve the change request
		err = apiClient.ApproveChangeRequest(auth, orgName, changeRequestID, int64(0), nil)
		require.Nil(t, err)
		t.Logf("Re-approved change request")

		// Step 16: Verify gate is satisfied again
		changeRequest, err = apiClient.GetChangeRequest(auth, orgName, changeRequestID)
		require.Nil(t, err)
		require.NotNil(t, changeRequest.GateEvaluation)
		require.True(t, changeRequest.GateEvaluation.Satisfied, "Gate should be satisfied after re-approval")
		t.Logf("After re-approval - satisfied: %v", changeRequest.GateEvaluation.Satisfied)

		// Step 17: Apply the change request
		applyResult, err := apiClient.ApplyChangeRequest(auth, orgName, changeRequestID)
		require.Nil(t, err)
		require.NotNil(t, applyResult)
		require.NotEmpty(t, applyResult.EntityUrl)
		t.Logf("Applied change request successfully - entity URL: %s", applyResult.EntityUrl)

		// Step 18: Verify the environment was updated by reading it
		env, _, err := apiClient.GetEnvironment(auth, orgName, PROJECT_NAME, envName)
		require.Nil(t, err)
		require.NotNil(t, env)
		require.NotNil(t, env.Values)
		require.NotNil(t, env.Values.AdditionalProperties)
		require.Equal(t, "bar", env.Values.AdditionalProperties["foo"])
		require.Equal(t, "test", env.Values.AdditionalProperties["environment"])
		require.Equal(t, "enabled", env.Values.AdditionalProperties["changeManagement"])
		t.Logf("Verified environment was updated with draft changes")

		// Step 19: Test additional change request events
		events, err := apiClient.ListChangeRequestEvents(auth, orgName, changeRequestID)
		require.Nil(t, err)
		require.NotNil(t, events)
		require.Greater(t, len(events.Events), 0, "Should have change request events")
		t.Logf("Change request has %d events", len(events.Events))

		// Step 20: Clean up - delete the change gate
		err = apiClient.DeleteEnvironmentChangeGate(auth, orgName, gateID)
		require.Nil(t, err)
		t.Logf("Deleted change gate")

		// Step 21: Verify gate was deleted
		gates, err = apiClient.ListEnvironmentChangeGates(auth, orgName, PROJECT_NAME, envName)
		require.Nil(t, err)
		require.NotNil(t, gates)
		require.Len(t, gates.Gates, 0, "All gates should be deleted")
		t.Logf("Verified all gates were deleted")
	})
}

func assertEnvDef(t *testing.T, env *EnvironmentDefinition, baseEnvName string) {
	require.Len(t, env.Imports, 1)
	require.Equal(t, PROJECT_NAME+"/"+baseEnvName, env.Imports[0])
	require.Equal(t, "bar", env.Values.AdditionalProperties["foo"])
	require.Equal(t, []any{1.0, 2.0, 3.0}, env.Values.AdditionalProperties["my_array"])

	require.Equal(t, "${foo}", env.Values.PulumiConfig["foo"])
	require.NotNil(t, env.Values.EnvironmentVariables)
	envVariables := *env.Values.EnvironmentVariables
	require.Equal(t, "${foo}", envVariables["FOO"])
}

func requireFindEnvironment(t *testing.T, envs *OrgEnvironments, findProject, findName string) {
	found := false
	for _, env := range envs.Environments {
		if env.Project == findProject && env.Name == findName {
			found = true
		}
	}

	require.True(t, found)
}

func removeAllGoTestEnvs(t *testing.T, apiClient *EscClient, auth context.Context, orgName string) {
	var continuationToken *string
	for {
		envs, err := apiClient.ListEnvironments(auth, orgName, continuationToken)
		require.Nil(t, err)

		if len(envs.Environments) == 0 {
			break
		}
		for _, env := range envs.Environments {
			// Clean up environments that match the test naming convention
			if env.Project == PROJECT_NAME && strings.HasPrefix(env.Name, ENV_PREFIX) {
				err := apiClient.DeleteEnvironment(auth, orgName, PROJECT_NAME, env.Name)
				require.Nil(t, err)
			}

			// Clean up clone environments
			if env.Project == CLONE_PROJECT_NAME && strings.HasPrefix(env.Name, ENV_PREFIX) {
				err := apiClient.DeleteEnvironment(auth, orgName, CLONE_PROJECT_NAME, env.Name)
				require.Nil(t, err)
			}

			// Clean up base environments
			if env.Project == PROJECT_NAME && strings.HasPrefix(env.Name, BASE_ENV_PREFIX) {
				err := apiClient.DeleteEnvironment(auth, orgName, PROJECT_NAME, env.Name)
				require.Nil(t, err)
			}
		}

		continuationToken = envs.NextToken
		if continuationToken == nil || *continuationToken == "" {
			break
		}
	}
}
