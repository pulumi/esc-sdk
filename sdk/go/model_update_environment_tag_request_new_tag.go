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

// checks if the UpdateEnvironmentTagRequestNewTag type satisfies the MappedNullable interface at compile time
var _ MappedNullable = &UpdateEnvironmentTagRequestNewTag{}

// UpdateEnvironmentTagRequestNewTag struct for UpdateEnvironmentTagRequestNewTag
type UpdateEnvironmentTagRequestNewTag struct {
	Name string `json:"name"`
	Value string `json:"value"`
}

type _UpdateEnvironmentTagRequestNewTag UpdateEnvironmentTagRequestNewTag

// NewUpdateEnvironmentTagRequestNewTag instantiates a new UpdateEnvironmentTagRequestNewTag object
// This constructor will assign default values to properties that have it defined,
// and makes sure properties required by API are set, but the set of arguments
// will change when the set of required properties is changed
func NewUpdateEnvironmentTagRequestNewTag(name string, value string) *UpdateEnvironmentTagRequestNewTag {
	this := UpdateEnvironmentTagRequestNewTag{}
	this.Name = name
	this.Value = value
	return &this
}

// NewUpdateEnvironmentTagRequestNewTagWithDefaults instantiates a new UpdateEnvironmentTagRequestNewTag object
// This constructor will only assign default values to properties that have it defined,
// but it doesn't guarantee that properties required by API are set
func NewUpdateEnvironmentTagRequestNewTagWithDefaults() *UpdateEnvironmentTagRequestNewTag {
	this := UpdateEnvironmentTagRequestNewTag{}
	return &this
}

// GetName returns the Name field value
func (o *UpdateEnvironmentTagRequestNewTag) GetName() string {
	if o == nil {
		var ret string
		return ret
	}

	return o.Name
}

// GetNameOk returns a tuple with the Name field value
// and a boolean to check if the value has been set.
func (o *UpdateEnvironmentTagRequestNewTag) GetNameOk() (*string, bool) {
	if o == nil {
		return nil, false
	}
	return &o.Name, true
}

// SetName sets field value
func (o *UpdateEnvironmentTagRequestNewTag) SetName(v string) {
	o.Name = v
}

// GetValue returns the Value field value
func (o *UpdateEnvironmentTagRequestNewTag) GetValue() string {
	if o == nil {
		var ret string
		return ret
	}

	return o.Value
}

// GetValueOk returns a tuple with the Value field value
// and a boolean to check if the value has been set.
func (o *UpdateEnvironmentTagRequestNewTag) GetValueOk() (*string, bool) {
	if o == nil {
		return nil, false
	}
	return &o.Value, true
}

// SetValue sets field value
func (o *UpdateEnvironmentTagRequestNewTag) SetValue(v string) {
	o.Value = v
}

func (o UpdateEnvironmentTagRequestNewTag) MarshalJSON() ([]byte, error) {
	toSerialize,err := o.ToMap()
	if err != nil {
		return []byte{}, err
	}
	return json.Marshal(toSerialize)
}

func (o UpdateEnvironmentTagRequestNewTag) ToMap() (map[string]interface{}, error) {
	toSerialize := map[string]interface{}{}
	toSerialize["name"] = o.Name
	toSerialize["value"] = o.Value
	return toSerialize, nil
}

func (o *UpdateEnvironmentTagRequestNewTag) UnmarshalJSON(data []byte) (err error) {
	// This validates that all required properties are included in the JSON object
	// by unmarshalling the object into a generic map with string keys and checking
	// that every required field exists as a key in the generic map.
	requiredProperties := []string{
		"name",
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

	varUpdateEnvironmentTagRequestNewTag := _UpdateEnvironmentTagRequestNewTag{}

	decoder := json.NewDecoder(bytes.NewReader(data))
	decoder.DisallowUnknownFields()
	err = decoder.Decode(&varUpdateEnvironmentTagRequestNewTag)

	if err != nil {
		return err
	}

	*o = UpdateEnvironmentTagRequestNewTag(varUpdateEnvironmentTagRequestNewTag)

	return err
}

type NullableUpdateEnvironmentTagRequestNewTag struct {
	value *UpdateEnvironmentTagRequestNewTag
	isSet bool
}

func (v NullableUpdateEnvironmentTagRequestNewTag) Get() *UpdateEnvironmentTagRequestNewTag {
	return v.value
}

func (v *NullableUpdateEnvironmentTagRequestNewTag) Set(val *UpdateEnvironmentTagRequestNewTag) {
	v.value = val
	v.isSet = true
}

func (v NullableUpdateEnvironmentTagRequestNewTag) IsSet() bool {
	return v.isSet
}

func (v *NullableUpdateEnvironmentTagRequestNewTag) Unset() {
	v.value = nil
	v.isSet = false
}

func NewNullableUpdateEnvironmentTagRequestNewTag(val *UpdateEnvironmentTagRequestNewTag) *NullableUpdateEnvironmentTagRequestNewTag {
	return &NullableUpdateEnvironmentTagRequestNewTag{value: val, isSet: true}
}

func (v NullableUpdateEnvironmentTagRequestNewTag) MarshalJSON() ([]byte, error) {
	return json.Marshal(v.value)
}

func (v *NullableUpdateEnvironmentTagRequestNewTag) UnmarshalJSON(src []byte) error {
	v.isSet = true
	return json.Unmarshal(src, &v.value)
}


