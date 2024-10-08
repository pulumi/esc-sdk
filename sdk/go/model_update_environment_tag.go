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

// checks if the UpdateEnvironmentTag type satisfies the MappedNullable interface at compile time
var _ MappedNullable = &UpdateEnvironmentTag{}

// UpdateEnvironmentTag struct for UpdateEnvironmentTag
type UpdateEnvironmentTag struct {
	CurrentTag UpdateEnvironmentTagCurrentTag `json:"currentTag"`
	NewTag UpdateEnvironmentTagNewTag `json:"newTag"`
}

type _UpdateEnvironmentTag UpdateEnvironmentTag

// NewUpdateEnvironmentTag instantiates a new UpdateEnvironmentTag object
// This constructor will assign default values to properties that have it defined,
// and makes sure properties required by API are set, but the set of arguments
// will change when the set of required properties is changed
func NewUpdateEnvironmentTag(currentTag UpdateEnvironmentTagCurrentTag, newTag UpdateEnvironmentTagNewTag) *UpdateEnvironmentTag {
	this := UpdateEnvironmentTag{}
	this.CurrentTag = currentTag
	this.NewTag = newTag
	return &this
}

// NewUpdateEnvironmentTagWithDefaults instantiates a new UpdateEnvironmentTag object
// This constructor will only assign default values to properties that have it defined,
// but it doesn't guarantee that properties required by API are set
func NewUpdateEnvironmentTagWithDefaults() *UpdateEnvironmentTag {
	this := UpdateEnvironmentTag{}
	return &this
}

// GetCurrentTag returns the CurrentTag field value
func (o *UpdateEnvironmentTag) GetCurrentTag() UpdateEnvironmentTagCurrentTag {
	if o == nil {
		var ret UpdateEnvironmentTagCurrentTag
		return ret
	}

	return o.CurrentTag
}

// GetCurrentTagOk returns a tuple with the CurrentTag field value
// and a boolean to check if the value has been set.
func (o *UpdateEnvironmentTag) GetCurrentTagOk() (*UpdateEnvironmentTagCurrentTag, bool) {
	if o == nil {
		return nil, false
	}
	return &o.CurrentTag, true
}

// SetCurrentTag sets field value
func (o *UpdateEnvironmentTag) SetCurrentTag(v UpdateEnvironmentTagCurrentTag) {
	o.CurrentTag = v
}

// GetNewTag returns the NewTag field value
func (o *UpdateEnvironmentTag) GetNewTag() UpdateEnvironmentTagNewTag {
	if o == nil {
		var ret UpdateEnvironmentTagNewTag
		return ret
	}

	return o.NewTag
}

// GetNewTagOk returns a tuple with the NewTag field value
// and a boolean to check if the value has been set.
func (o *UpdateEnvironmentTag) GetNewTagOk() (*UpdateEnvironmentTagNewTag, bool) {
	if o == nil {
		return nil, false
	}
	return &o.NewTag, true
}

// SetNewTag sets field value
func (o *UpdateEnvironmentTag) SetNewTag(v UpdateEnvironmentTagNewTag) {
	o.NewTag = v
}

func (o UpdateEnvironmentTag) MarshalJSON() ([]byte, error) {
	toSerialize,err := o.ToMap()
	if err != nil {
		return []byte{}, err
	}
	return json.Marshal(toSerialize)
}

func (o UpdateEnvironmentTag) ToMap() (map[string]interface{}, error) {
	toSerialize := map[string]interface{}{}
	toSerialize["currentTag"] = o.CurrentTag
	toSerialize["newTag"] = o.NewTag
	return toSerialize, nil
}

func (o *UpdateEnvironmentTag) UnmarshalJSON(data []byte) (err error) {
	// This validates that all required properties are included in the JSON object
	// by unmarshalling the object into a generic map with string keys and checking
	// that every required field exists as a key in the generic map.
	requiredProperties := []string{
		"currentTag",
		"newTag",
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

	varUpdateEnvironmentTag := _UpdateEnvironmentTag{}

	decoder := json.NewDecoder(bytes.NewReader(data))
	decoder.DisallowUnknownFields()
	err = decoder.Decode(&varUpdateEnvironmentTag)

	if err != nil {
		return err
	}

	*o = UpdateEnvironmentTag(varUpdateEnvironmentTag)

	return err
}

type NullableUpdateEnvironmentTag struct {
	value *UpdateEnvironmentTag
	isSet bool
}

func (v NullableUpdateEnvironmentTag) Get() *UpdateEnvironmentTag {
	return v.value
}

func (v *NullableUpdateEnvironmentTag) Set(val *UpdateEnvironmentTag) {
	v.value = val
	v.isSet = true
}

func (v NullableUpdateEnvironmentTag) IsSet() bool {
	return v.isSet
}

func (v *NullableUpdateEnvironmentTag) Unset() {
	v.value = nil
	v.isSet = false
}

func NewNullableUpdateEnvironmentTag(val *UpdateEnvironmentTag) *NullableUpdateEnvironmentTag {
	return &NullableUpdateEnvironmentTag{value: val, isSet: true}
}

func (v NullableUpdateEnvironmentTag) MarshalJSON() ([]byte, error) {
	return json.Marshal(v.value)
}

func (v *NullableUpdateEnvironmentTag) UnmarshalJSON(src []byte) error {
	v.isSet = true
	return json.Unmarshal(src, &v.value)
}


