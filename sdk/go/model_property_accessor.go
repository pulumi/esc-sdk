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

// checks if the PropertyAccessor type satisfies the MappedNullable interface at compile time
var _ MappedNullable = &PropertyAccessor{}

// PropertyAccessor struct for PropertyAccessor
type PropertyAccessor struct {
	Index *int32 `json:"index,omitempty"`
	Key string `json:"key"`
	Range Range `json:"range"`
	Value *Range `json:"value,omitempty"`
}

type _PropertyAccessor PropertyAccessor

// NewPropertyAccessor instantiates a new PropertyAccessor object
// This constructor will assign default values to properties that have it defined,
// and makes sure properties required by API are set, but the set of arguments
// will change when the set of required properties is changed
func NewPropertyAccessor(key string, range_ Range) *PropertyAccessor {
	this := PropertyAccessor{}
	this.Key = key
	this.Range = range_
	return &this
}

// NewPropertyAccessorWithDefaults instantiates a new PropertyAccessor object
// This constructor will only assign default values to properties that have it defined,
// but it doesn't guarantee that properties required by API are set
func NewPropertyAccessorWithDefaults() *PropertyAccessor {
	this := PropertyAccessor{}
	return &this
}

// GetIndex returns the Index field value if set, zero value otherwise.
func (o *PropertyAccessor) GetIndex() int32 {
	if o == nil || IsNil(o.Index) {
		var ret int32
		return ret
	}
	return *o.Index
}

// GetIndexOk returns a tuple with the Index field value if set, nil otherwise
// and a boolean to check if the value has been set.
func (o *PropertyAccessor) GetIndexOk() (*int32, bool) {
	if o == nil || IsNil(o.Index) {
		return nil, false
	}
	return o.Index, true
}

// HasIndex returns a boolean if a field has been set.
func (o *PropertyAccessor) HasIndex() bool {
	if o != nil && !IsNil(o.Index) {
		return true
	}

	return false
}

// SetIndex gets a reference to the given int32 and assigns it to the Index field.
func (o *PropertyAccessor) SetIndex(v int32) {
	o.Index = &v
}

// GetKey returns the Key field value
func (o *PropertyAccessor) GetKey() string {
	if o == nil {
		var ret string
		return ret
	}

	return o.Key
}

// GetKeyOk returns a tuple with the Key field value
// and a boolean to check if the value has been set.
func (o *PropertyAccessor) GetKeyOk() (*string, bool) {
	if o == nil {
		return nil, false
	}
	return &o.Key, true
}

// SetKey sets field value
func (o *PropertyAccessor) SetKey(v string) {
	o.Key = v
}

// GetRange returns the Range field value
func (o *PropertyAccessor) GetRange() Range {
	if o == nil {
		var ret Range
		return ret
	}

	return o.Range
}

// GetRangeOk returns a tuple with the Range field value
// and a boolean to check if the value has been set.
func (o *PropertyAccessor) GetRangeOk() (*Range, bool) {
	if o == nil {
		return nil, false
	}
	return &o.Range, true
}

// SetRange sets field value
func (o *PropertyAccessor) SetRange(v Range) {
	o.Range = v
}

// GetValue returns the Value field value if set, zero value otherwise.
func (o *PropertyAccessor) GetValue() Range {
	if o == nil || IsNil(o.Value) {
		var ret Range
		return ret
	}
	return *o.Value
}

// GetValueOk returns a tuple with the Value field value if set, nil otherwise
// and a boolean to check if the value has been set.
func (o *PropertyAccessor) GetValueOk() (*Range, bool) {
	if o == nil || IsNil(o.Value) {
		return nil, false
	}
	return o.Value, true
}

// HasValue returns a boolean if a field has been set.
func (o *PropertyAccessor) HasValue() bool {
	if o != nil && !IsNil(o.Value) {
		return true
	}

	return false
}

// SetValue gets a reference to the given Range and assigns it to the Value field.
func (o *PropertyAccessor) SetValue(v Range) {
	o.Value = &v
}

func (o PropertyAccessor) MarshalJSON() ([]byte, error) {
	toSerialize,err := o.ToMap()
	if err != nil {
		return []byte{}, err
	}
	return json.Marshal(toSerialize)
}

func (o PropertyAccessor) ToMap() (map[string]interface{}, error) {
	toSerialize := map[string]interface{}{}
	if !IsNil(o.Index) {
		toSerialize["index"] = o.Index
	}
	toSerialize["key"] = o.Key
	toSerialize["range"] = o.Range
	if !IsNil(o.Value) {
		toSerialize["value"] = o.Value
	}
	return toSerialize, nil
}

func (o *PropertyAccessor) UnmarshalJSON(data []byte) (err error) {
	// This validates that all required properties are included in the JSON object
	// by unmarshalling the object into a generic map with string keys and checking
	// that every required field exists as a key in the generic map.
	requiredProperties := []string{
		"key",
		"range",
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

	varPropertyAccessor := _PropertyAccessor{}

	decoder := json.NewDecoder(bytes.NewReader(data))
	decoder.DisallowUnknownFields()
	err = decoder.Decode(&varPropertyAccessor)

	if err != nil {
		return err
	}

	*o = PropertyAccessor(varPropertyAccessor)

	return err
}

type NullablePropertyAccessor struct {
	value *PropertyAccessor
	isSet bool
}

func (v NullablePropertyAccessor) Get() *PropertyAccessor {
	return v.value
}

func (v *NullablePropertyAccessor) Set(val *PropertyAccessor) {
	v.value = val
	v.isSet = true
}

func (v NullablePropertyAccessor) IsSet() bool {
	return v.isSet
}

func (v *NullablePropertyAccessor) Unset() {
	v.value = nil
	v.isSet = false
}

func NewNullablePropertyAccessor(val *PropertyAccessor) *NullablePropertyAccessor {
	return &NullablePropertyAccessor{value: val, isSet: true}
}

func (v NullablePropertyAccessor) MarshalJSON() ([]byte, error) {
	return json.Marshal(v.value)
}

func (v *NullablePropertyAccessor) UnmarshalJSON(src []byte) error {
	v.isSet = true
	return json.Unmarshal(src, &v.value)
}


