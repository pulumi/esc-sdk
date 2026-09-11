// Copyright 2026, Pulumi Corporation.  All rights reserved.

package esc_sdk

import (
	"encoding/json"

	"github.com/pulumi/pulumi-cloud-sdk/go/apitype"
)

type (
	Value                     = apitype.EscValue
	Trace                     = apitype.EscTrace
	Range                     = apitype.EscRange
	Pos                       = apitype.EscPos
	Expr                      = apitype.EscExpr
	Environment               = apitype.EscEnvironment
	EvaluatedExecutionContext = apitype.EscEvaluatedExecutionContext
	CheckEnvironment          = apitype.EnvironmentResponse
	EnvironmentDiagnostic     = apitype.EnvironmentDiagnostic
	EnvironmentDiagnostics    = apitype.EnvironmentDiagnosticsResponse
	OpenEnvironment           = apitype.OpenEnvironmentResponse
	OrgEnvironment            = apitype.OrgEnvironment
	OrgEnvironments           = apitype.ListEnvironmentsResponse
	EnvironmentRevision       = apitype.EnvironmentRevision
	EnvironmentRevisionTag    = apitype.EnvironmentRevisionTag
	EnvironmentRevisionTags   = apitype.ListEnvironmentRevisionTagsResponse
	EnvironmentTag            = apitype.EnvironmentTag
	ListEnvironmentTags       = apitype.ListEnvironmentTagsResponse
)

// EnvironmentDefinition is the YAML document that defines an environment.
type EnvironmentDefinition struct {
	Imports []string                     `json:"imports,omitempty"`
	Values  *EnvironmentDefinitionValues `json:"values,omitempty"`
}

// EnvironmentDefinitionValues is the values block of an environment definition.
// Keys other than the three well-known ones live in AdditionalProperties.
type EnvironmentDefinitionValues struct {
	PulumiConfig         map[string]any    `json:"pulumiConfig,omitempty"`
	EnvironmentVariables map[string]string `json:"environmentVariables,omitempty"`
	Files                map[string]string `json:"files,omitempty"`
	AdditionalProperties map[string]any    `json:"-"`
}

type environmentDefinitionValuesFields EnvironmentDefinitionValues

func (v EnvironmentDefinitionValues) MarshalJSON() ([]byte, error) {
	known, err := json.Marshal(environmentDefinitionValuesFields(v))
	if err != nil {
		return nil, err
	}
	merged := make(map[string]json.RawMessage, len(v.AdditionalProperties)+3)
	for key, value := range v.AdditionalProperties {
		encoded, err := json.Marshal(value)
		if err != nil {
			return nil, err
		}
		merged[key] = encoded
	}
	if err := json.Unmarshal(known, &merged); err != nil {
		return nil, err
	}
	return json.Marshal(merged)
}

func (v *EnvironmentDefinitionValues) UnmarshalJSON(data []byte) error {
	var fields environmentDefinitionValuesFields
	if err := json.Unmarshal(data, &fields); err != nil {
		return err
	}
	additional := map[string]any{}
	if err := json.Unmarshal(data, &additional); err != nil {
		return err
	}
	delete(additional, "pulumiConfig")
	delete(additional, "environmentVariables")
	delete(additional, "files")
	*v = EnvironmentDefinitionValues(fields)
	v.AdditionalProperties = additional
	return nil
}
