// Copyright 2024, Pulumi Corporation.  All rights reserved.
/*
ESC (Environments, Secrets, Config) API

Pulumi ESC allows you to compose and manage hierarchical collections of configuration and secrets and consume them in various ways.

API version: 0.1.0
*/

// Code generated by OpenAPI Generator (https://openapi-generator.tech); DO NOT EDIT.

package esc_sdk

import (
	"encoding/json"
)

// checks if the CheckEnvironment type satisfies the MappedNullable interface at compile time
var _ MappedNullable = &CheckEnvironment{}

// CheckEnvironment struct for CheckEnvironment
type CheckEnvironment struct {
	Exprs *map[string]Expr `json:"exprs,omitempty"`
	Properties *map[string]Value `json:"properties,omitempty"`
	Schema interface{} `json:"schema,omitempty"`
	ExecutionContext *EvaluatedExecutionContext `json:"executionContext,omitempty"`
	Diagnostics []EnvironmentDiagnostic `json:"diagnostics,omitempty"`
}

// NewCheckEnvironment instantiates a new CheckEnvironment object
// This constructor will assign default values to properties that have it defined,
// and makes sure properties required by API are set, but the set of arguments
// will change when the set of required properties is changed
func NewCheckEnvironment() *CheckEnvironment {
	this := CheckEnvironment{}
	return &this
}

// NewCheckEnvironmentWithDefaults instantiates a new CheckEnvironment object
// This constructor will only assign default values to properties that have it defined,
// but it doesn't guarantee that properties required by API are set
func NewCheckEnvironmentWithDefaults() *CheckEnvironment {
	this := CheckEnvironment{}
	return &this
}

// GetExprs returns the Exprs field value if set, zero value otherwise.
func (o *CheckEnvironment) GetExprs() map[string]Expr {
	if o == nil || IsNil(o.Exprs) {
		var ret map[string]Expr
		return ret
	}
	return *o.Exprs
}

// GetExprsOk returns a tuple with the Exprs field value if set, nil otherwise
// and a boolean to check if the value has been set.
func (o *CheckEnvironment) GetExprsOk() (*map[string]Expr, bool) {
	if o == nil || IsNil(o.Exprs) {
		return nil, false
	}
	return o.Exprs, true
}

// HasExprs returns a boolean if a field has been set.
func (o *CheckEnvironment) HasExprs() bool {
	if o != nil && !IsNil(o.Exprs) {
		return true
	}

	return false
}

// SetExprs gets a reference to the given map[string]Expr and assigns it to the Exprs field.
func (o *CheckEnvironment) SetExprs(v map[string]Expr) {
	o.Exprs = &v
}

// GetProperties returns the Properties field value if set, zero value otherwise.
func (o *CheckEnvironment) GetProperties() map[string]Value {
	if o == nil || IsNil(o.Properties) {
		var ret map[string]Value
		return ret
	}
	return *o.Properties
}

// GetPropertiesOk returns a tuple with the Properties field value if set, nil otherwise
// and a boolean to check if the value has been set.
func (o *CheckEnvironment) GetPropertiesOk() (*map[string]Value, bool) {
	if o == nil || IsNil(o.Properties) {
		return nil, false
	}
	return o.Properties, true
}

// HasProperties returns a boolean if a field has been set.
func (o *CheckEnvironment) HasProperties() bool {
	if o != nil && !IsNil(o.Properties) {
		return true
	}

	return false
}

// SetProperties gets a reference to the given map[string]Value and assigns it to the Properties field.
func (o *CheckEnvironment) SetProperties(v map[string]Value) {
	o.Properties = &v
}

// GetSchema returns the Schema field value if set, zero value otherwise (both if not set or set to explicit null).
func (o *CheckEnvironment) GetSchema() interface{} {
	if o == nil {
		var ret interface{}
		return ret
	}
	return o.Schema
}

// GetSchemaOk returns a tuple with the Schema field value if set, nil otherwise
// and a boolean to check if the value has been set.
// NOTE: If the value is an explicit nil, `nil, true` will be returned
func (o *CheckEnvironment) GetSchemaOk() (*interface{}, bool) {
	if o == nil || IsNil(o.Schema) {
		return nil, false
	}
	return &o.Schema, true
}

// HasSchema returns a boolean if a field has been set.
func (o *CheckEnvironment) HasSchema() bool {
	if o != nil && IsNil(o.Schema) {
		return true
	}

	return false
}

// SetSchema gets a reference to the given interface{} and assigns it to the Schema field.
func (o *CheckEnvironment) SetSchema(v interface{}) {
	o.Schema = v
}

// GetExecutionContext returns the ExecutionContext field value if set, zero value otherwise.
func (o *CheckEnvironment) GetExecutionContext() EvaluatedExecutionContext {
	if o == nil || IsNil(o.ExecutionContext) {
		var ret EvaluatedExecutionContext
		return ret
	}
	return *o.ExecutionContext
}

// GetExecutionContextOk returns a tuple with the ExecutionContext field value if set, nil otherwise
// and a boolean to check if the value has been set.
func (o *CheckEnvironment) GetExecutionContextOk() (*EvaluatedExecutionContext, bool) {
	if o == nil || IsNil(o.ExecutionContext) {
		return nil, false
	}
	return o.ExecutionContext, true
}

// HasExecutionContext returns a boolean if a field has been set.
func (o *CheckEnvironment) HasExecutionContext() bool {
	if o != nil && !IsNil(o.ExecutionContext) {
		return true
	}

	return false
}

// SetExecutionContext gets a reference to the given EvaluatedExecutionContext and assigns it to the ExecutionContext field.
func (o *CheckEnvironment) SetExecutionContext(v EvaluatedExecutionContext) {
	o.ExecutionContext = &v
}

// GetDiagnostics returns the Diagnostics field value if set, zero value otherwise.
func (o *CheckEnvironment) GetDiagnostics() []EnvironmentDiagnostic {
	if o == nil || IsNil(o.Diagnostics) {
		var ret []EnvironmentDiagnostic
		return ret
	}
	return o.Diagnostics
}

// GetDiagnosticsOk returns a tuple with the Diagnostics field value if set, nil otherwise
// and a boolean to check if the value has been set.
func (o *CheckEnvironment) GetDiagnosticsOk() ([]EnvironmentDiagnostic, bool) {
	if o == nil || IsNil(o.Diagnostics) {
		return nil, false
	}
	return o.Diagnostics, true
}

// HasDiagnostics returns a boolean if a field has been set.
func (o *CheckEnvironment) HasDiagnostics() bool {
	if o != nil && !IsNil(o.Diagnostics) {
		return true
	}

	return false
}

// SetDiagnostics gets a reference to the given []EnvironmentDiagnostic and assigns it to the Diagnostics field.
func (o *CheckEnvironment) SetDiagnostics(v []EnvironmentDiagnostic) {
	o.Diagnostics = v
}

func (o CheckEnvironment) MarshalJSON() ([]byte, error) {
	toSerialize,err := o.ToMap()
	if err != nil {
		return []byte{}, err
	}
	return json.Marshal(toSerialize)
}

func (o CheckEnvironment) ToMap() (map[string]interface{}, error) {
	toSerialize := map[string]interface{}{}
	if !IsNil(o.Exprs) {
		toSerialize["exprs"] = o.Exprs
	}
	if !IsNil(o.Properties) {
		toSerialize["properties"] = o.Properties
	}
	if o.Schema != nil {
		toSerialize["schema"] = o.Schema
	}
	if !IsNil(o.ExecutionContext) {
		toSerialize["executionContext"] = o.ExecutionContext
	}
	if !IsNil(o.Diagnostics) {
		toSerialize["diagnostics"] = o.Diagnostics
	}
	return toSerialize, nil
}

type NullableCheckEnvironment struct {
	value *CheckEnvironment
	isSet bool
}

func (v NullableCheckEnvironment) Get() *CheckEnvironment {
	return v.value
}

func (v *NullableCheckEnvironment) Set(val *CheckEnvironment) {
	v.value = val
	v.isSet = true
}

func (v NullableCheckEnvironment) IsSet() bool {
	return v.isSet
}

func (v *NullableCheckEnvironment) Unset() {
	v.value = nil
	v.isSet = false
}

func NewNullableCheckEnvironment(val *CheckEnvironment) *NullableCheckEnvironment {
	return &NullableCheckEnvironment{value: val, isSet: true}
}

func (v NullableCheckEnvironment) MarshalJSON() ([]byte, error) {
	return json.Marshal(v.value)
}

func (v *NullableCheckEnvironment) UnmarshalJSON(src []byte) error {
	v.isSet = true
	return json.Unmarshal(src, &v.value)
}


