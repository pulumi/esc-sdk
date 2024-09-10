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
	"bytes"
	"fmt"
)

// checks if the UpdateEnvironmentTagCurrentTag type satisfies the MappedNullable interface at compile time
var _ MappedNullable = &UpdateEnvironmentTagCurrentTag{}

// UpdateEnvironmentTagCurrentTag struct for UpdateEnvironmentTagCurrentTag
type UpdateEnvironmentTagCurrentTag struct {
	Value string `json:"value"`
}

type _UpdateEnvironmentTagCurrentTag UpdateEnvironmentTagCurrentTag

// NewUpdateEnvironmentTagCurrentTag instantiates a new UpdateEnvironmentTagCurrentTag object
// This constructor will assign default values to properties that have it defined,
// and makes sure properties required by API are set, but the set of arguments
// will change when the set of required properties is changed
func NewUpdateEnvironmentTagCurrentTag(value string) *UpdateEnvironmentTagCurrentTag {
	this := UpdateEnvironmentTagCurrentTag{}
	this.Value = value
	return &this
}

// NewUpdateEnvironmentTagCurrentTagWithDefaults instantiates a new UpdateEnvironmentTagCurrentTag object
// This constructor will only assign default values to properties that have it defined,
// but it doesn't guarantee that properties required by API are set
func NewUpdateEnvironmentTagCurrentTagWithDefaults() *UpdateEnvironmentTagCurrentTag {
	this := UpdateEnvironmentTagCurrentTag{}
	return &this
}

// GetValue returns the Value field value
func (o *UpdateEnvironmentTagCurrentTag) GetValue() string {
	if o == nil {
		var ret string
		return ret
	}

	return o.Value
}

// GetValueOk returns a tuple with the Value field value
// and a boolean to check if the value has been set.
func (o *UpdateEnvironmentTagCurrentTag) GetValueOk() (*string, bool) {
	if o == nil {
		return nil, false
	}
	return &o.Value, true
}

// SetValue sets field value
func (o *UpdateEnvironmentTagCurrentTag) SetValue(v string) {
	o.Value = v
}

func (o UpdateEnvironmentTagCurrentTag) MarshalJSON() ([]byte, error) {
	toSerialize,err := o.ToMap()
	if err != nil {
		return []byte{}, err
	}
	return json.Marshal(toSerialize)
}

func (o UpdateEnvironmentTagCurrentTag) ToMap() (map[string]interface{}, error) {
	toSerialize := map[string]interface{}{}
	toSerialize["value"] = o.Value
	return toSerialize, nil
}

func (o *UpdateEnvironmentTagCurrentTag) UnmarshalJSON(data []byte) (err error) {
	// This validates that all required properties are included in the JSON object
	// by unmarshalling the object into a generic map with string keys and checking
	// that every required field exists as a key in the generic map.
	requiredProperties := []string{
		"value",
	}

	allProperties := make(map[string]interface{})

	err = json.Unmarshal(data, &allProperties)

	if err != nil {
		return err;
	}

	for _, requiredProperty := range(requiredProperties) {
		if _, exists := allProperties[requiredProperty]; !exists {
			return fmt.Errorf("no value given for required property %v", requiredProperty)
		}
	}

	varUpdateEnvironmentTagCurrentTag := _UpdateEnvironmentTagCurrentTag{}

	decoder := json.NewDecoder(bytes.NewReader(data))
	decoder.DisallowUnknownFields()
	err = decoder.Decode(&varUpdateEnvironmentTagCurrentTag)

	if err != nil {
		return err
	}

	*o = UpdateEnvironmentTagCurrentTag(varUpdateEnvironmentTagCurrentTag)

	return err
}

type NullableUpdateEnvironmentTagCurrentTag struct {
	value *UpdateEnvironmentTagCurrentTag
	isSet bool
}

func (v NullableUpdateEnvironmentTagCurrentTag) Get() *UpdateEnvironmentTagCurrentTag {
	return v.value
}

func (v *NullableUpdateEnvironmentTagCurrentTag) Set(val *UpdateEnvironmentTagCurrentTag) {
	v.value = val
	v.isSet = true
}

func (v NullableUpdateEnvironmentTagCurrentTag) IsSet() bool {
	return v.isSet
}

func (v *NullableUpdateEnvironmentTagCurrentTag) Unset() {
	v.value = nil
	v.isSet = false
}

func NewNullableUpdateEnvironmentTagCurrentTag(val *UpdateEnvironmentTagCurrentTag) *NullableUpdateEnvironmentTagCurrentTag {
	return &NullableUpdateEnvironmentTagCurrentTag{value: val, isSet: true}
}

func (v NullableUpdateEnvironmentTagCurrentTag) MarshalJSON() ([]byte, error) {
	return json.Marshal(v.value)
}

func (v *NullableUpdateEnvironmentTagCurrentTag) UnmarshalJSON(src []byte) error {
	v.isSet = true
	return json.Unmarshal(src, &v.value)
}


