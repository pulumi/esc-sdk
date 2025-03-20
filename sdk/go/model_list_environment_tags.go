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

// checks if the ListEnvironmentTags type satisfies the MappedNullable interface at compile time
var _ MappedNullable = &ListEnvironmentTags{}

// ListEnvironmentTags struct for ListEnvironmentTags
type ListEnvironmentTags struct {
	Tags map[string]EnvironmentTag `json:"tags"`
	NextToken string `json:"nextToken"`
}

type _ListEnvironmentTags ListEnvironmentTags

// NewListEnvironmentTags instantiates a new ListEnvironmentTags object
// This constructor will assign default values to properties that have it defined,
// and makes sure properties required by API are set, but the set of arguments
// will change when the set of required properties is changed
func NewListEnvironmentTags(tags map[string]EnvironmentTag, nextToken string) *ListEnvironmentTags {
	this := ListEnvironmentTags{}
	this.Tags = tags
	this.NextToken = nextToken
	return &this
}

// NewListEnvironmentTagsWithDefaults instantiates a new ListEnvironmentTags object
// This constructor will only assign default values to properties that have it defined,
// but it doesn't guarantee that properties required by API are set
func NewListEnvironmentTagsWithDefaults() *ListEnvironmentTags {
	this := ListEnvironmentTags{}
	return &this
}

// GetTags returns the Tags field value
func (o *ListEnvironmentTags) GetTags() map[string]EnvironmentTag {
	if o == nil {
		var ret map[string]EnvironmentTag
		return ret
	}

	return o.Tags
}

// GetTagsOk returns a tuple with the Tags field value
// and a boolean to check if the value has been set.
func (o *ListEnvironmentTags) GetTagsOk() (*map[string]EnvironmentTag, bool) {
	if o == nil {
		return nil, false
	}
	return &o.Tags, true
}

// SetTags sets field value
func (o *ListEnvironmentTags) SetTags(v map[string]EnvironmentTag) {
	o.Tags = v
}

// GetNextToken returns the NextToken field value
func (o *ListEnvironmentTags) GetNextToken() string {
	if o == nil {
		var ret string
		return ret
	}

	return o.NextToken
}

// GetNextTokenOk returns a tuple with the NextToken field value
// and a boolean to check if the value has been set.
func (o *ListEnvironmentTags) GetNextTokenOk() (*string, bool) {
	if o == nil {
		return nil, false
	}
	return &o.NextToken, true
}

// SetNextToken sets field value
func (o *ListEnvironmentTags) SetNextToken(v string) {
	o.NextToken = v
}

func (o ListEnvironmentTags) MarshalJSON() ([]byte, error) {
	toSerialize,err := o.ToMap()
	if err != nil {
		return []byte{}, err
	}
	return json.Marshal(toSerialize)
}

func (o ListEnvironmentTags) ToMap() (map[string]interface{}, error) {
	toSerialize := map[string]interface{}{}
	toSerialize["tags"] = o.Tags
	toSerialize["nextToken"] = o.NextToken
	return toSerialize, nil
}

func (o *ListEnvironmentTags) UnmarshalJSON(data []byte) (err error) {
	// This validates that all required properties are included in the JSON object
	// by unmarshalling the object into a generic map with string keys and checking
	// that every required field exists as a key in the generic map.
	requiredProperties := []string{
		"tags",
		"nextToken",
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

	varListEnvironmentTags := _ListEnvironmentTags{}

	decoder := json.NewDecoder(bytes.NewReader(data))
	decoder.DisallowUnknownFields()
	err = decoder.Decode(&varListEnvironmentTags)

	if err != nil {
		return err
	}

	*o = ListEnvironmentTags(varListEnvironmentTags)

	return err
}

type NullableListEnvironmentTags struct {
	value *ListEnvironmentTags
	isSet bool
}

func (v NullableListEnvironmentTags) Get() *ListEnvironmentTags {
	return v.value
}

func (v *NullableListEnvironmentTags) Set(val *ListEnvironmentTags) {
	v.value = val
	v.isSet = true
}

func (v NullableListEnvironmentTags) IsSet() bool {
	return v.isSet
}

func (v *NullableListEnvironmentTags) Unset() {
	v.value = nil
	v.isSet = false
}

func NewNullableListEnvironmentTags(val *ListEnvironmentTags) *NullableListEnvironmentTags {
	return &NullableListEnvironmentTags{value: val, isSet: true}
}

func (v NullableListEnvironmentTags) MarshalJSON() ([]byte, error) {
	return json.Marshal(v.value)
}

func (v *NullableListEnvironmentTags) UnmarshalJSON(src []byte) error {
	v.isSet = true
	return json.Unmarshal(src, &v.value)
}


