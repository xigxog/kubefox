// Copyright 2023 XigXog
//
// This Source Code Form is subject to the terms of the Mozilla Public
// License, v. 2.0. If a copy of the MPL was not distributed with this
// file, You can obtain one at https://mozilla.org/MPL/2.0/.
//
// SPDX-License-Identifier: MPL-2.0

package api

import (
	"encoding/json"
	"errors"
	"fmt"
	"strconv"
)

// ValType represents the stored type of Var.
type ValType int

const (
	Unknown        ValType = iota // holds an unknown
	Nil                           // holds a null
	Bool                          // holds a boolean
	Number                        // holds an int or float
	String                        // holds a string
	ArrayNumber                   // holds an array of ints or floats
	ArrayString                   // holds an array of strings
	MapArrayString                // holds a map of array of strings
)

type Val struct {
	boolVal     bool                `json:"-"`
	numVal      float64             `json:"-"`
	strVal      string              `json:"-"`
	arrayNumVal []float64           `json:"-"`
	arrayStrVal []string            `json:"-"`
	mapStrVal   map[string][]string `json:"-"`
	Type        ValType             `json:"-"`
}

func ValNil() *Val {
	return &Val{Type: Nil}
}

func ValBool(val bool) *Val {
	return &Val{Type: Bool, boolVal: val}
}

func ValInt(val int) *Val {
	return &Val{Type: Number, numVal: float64(val)}
}

func ValFloat(val float64) *Val {
	return &Val{Type: Number, numVal: val}
}

func ValString(val string) *Val {
	return &Val{Type: String, strVal: val}
}

func ValArrayInt(val []int) *Val {
	arr := make([]float64, len(val))
	for i, v := range val {
		arr[i] = float64(v)
	}
	return &Val{Type: ArrayNumber, arrayNumVal: arr}
}
func ValArrayFloat(val []float64) *Val {
	return &Val{Type: ArrayNumber, arrayNumVal: val}
}

func ValArrayString(val []string) *Val {
	return &Val{Type: ArrayString, arrayStrVal: val}
}

func ValMapArrayString(val map[string][]string) *Val {
	return &Val{Type: ArrayString, mapStrVal: val}
}

func (val *Val) Any() any {
	switch val.Type {
	case Bool:
		return val.boolVal
	case Number:
		return val.numVal
	case String:
		return val.strVal
	case ArrayNumber:
		return val.arrayNumVal
	case ArrayString:
		return val.arrayStrVal
	case MapArrayString:
		return val.mapStrVal
	default:
		return ""
	}
}

// Bool returns the boolean value if type is Bool. If type is Number, false will
// be returned if value is 0, otherwise true is returned. If type is String, an
// attempt to parse the boolean value will be made. If parsing fails or type is
// Array false will be returned.
func (val *Val) Bool() bool {
	switch val.Type {
	case Bool:
		return val.boolVal
	case Number:
		if val.numVal == 0 {
			return false
		} else {
			return true
		}
	case String:
		b, _ := strconv.ParseBool(val.strVal)
		return b
	default:
		return false
	}
}

func (val *Val) BoolDef(def bool) bool {
	if val.Type != Bool {
		return def
	}
	return val.boolVal
}

// Int returns the int value if type is Number. If type is Bool 1 will be
// returned if true, otherwise 0 is returned. If type is String an attempt to
// parse the number will be made. If parsing fails or type is Array 0 will be
// returned.
func (val *Val) Int() int {
	return int(val.Float())
}

func (val *Val) IntDef(def int) int {
	if val.Type != Number {
		return def
	}
	return int(val.numVal)
}

// Float returns the float64 value if type is Number. If type is Bool 1 will be
// returned if true, otherwise 0 is returned. If  type is String an attempt to
// parse the number will be made. If parsing fails or type is Array 0 will be
// returned.
func (val *Val) Float() float64 {
	switch val.Type {
	case Bool:
		if val.boolVal {
			return 1
		} else {
			return 0
		}
	case Number:
		return val.numVal
	case String:
		i, _ := strconv.ParseFloat(val.strVal, 64)
		return i
	default:
		return 0
	}
}

func (val *Val) FloatDef(def float64) float64 {
	if val.Type != Number {
		return def
	}
	return val.numVal
}

// String returns the string value if type is String. If type is Array the JSON
// representation of the array is returned. Otherwise default string format of
// the value is returned.
func (val *Val) String() string {
	switch val.Type {
	case Bool:
		return fmt.Sprint(val.boolVal)
	case Number:
		return fmt.Sprint(val.numVal)
	case String:
		return val.strVal
	case ArrayNumber:
		b, _ := json.Marshal(val.arrayNumVal)
		return string(b)
	case ArrayString:
		b, _ := json.Marshal(val.arrayStrVal)
		return string(b)
	default:
		return ""
	}
}

func (val *Val) StringDef(def string) string {
	if val.Type != String {
		return def
	}
	return val.strVal
}

// ArrayInt returns the array value if type is ArrayNumber. Otherwise nil is
// returned.
func (val *Val) ArrayInt() []int {
	if val.Type != ArrayNumber {
		return nil
	}

	arr := make([]int, len(val.arrayNumVal))
	for i, v := range val.arrayNumVal {
		arr[i] = int(v)
	}
	return arr
}

// ArrayFloat returns the array value if type is ArrayNumber. Otherwise nil is
// returned.
func (val *Val) ArrayFloat() []float64 {
	if val.Type != ArrayNumber {
		return nil
	}
	return val.arrayNumVal
}

// ArrayString returns the array value if type is ArrayString or ArrayNumber.
// Otherwise nil is returned.
func (val *Val) ArrayString() []string {
	if val.Type == ArrayString {
		return val.arrayStrVal
	}

	if val.Type == ArrayNumber {
		a := make([]string, len(val.arrayNumVal))
		for i, v := range val.arrayNumVal {
			a[i] = fmt.Sprint(v)
		}
		return a
	}

	return nil
}

func (val *Val) MapArrayString() map[string][]string {
	if val.Type != MapArrayString {
		return map[string][]string{}
	}

	return val.mapStrVal
}

func (val *Val) Equals(rhs *Val) bool {
	if val == nil && rhs == nil {
		return true
	}
	if val == nil || rhs == nil {
		return false
	}

	return val.Type == rhs.Type && val.Any() == rhs.Any()
}

func (val *Val) IsUnknown() bool {
	return val.Type == Unknown
}

func (val *Val) IsNil() bool {
	return val == nil || val.Type == Nil
}

func (val *Val) IsBool() bool {
	return val.Type == Bool
}

func (val *Val) IsString() bool {
	return val.Type == String
}

func (val *Val) IsNumber() bool {
	return val.Type == Number
}

func (val *Val) IsArrayNumber() bool {
	return val.Type == ArrayNumber
}

func (val *Val) IsArrayString() bool {
	return val.Type == ArrayString
}

func (val *Val) IsMapArrayString() bool {
	return val.Type == MapArrayString
}

func (val *Val) IsEmpty() bool {
	switch val.Type {
	case Unknown, Nil:
		return true
	case String:
		return val.String() == ""
	case ArrayNumber:
		return len(val.arrayNumVal) == 0
	case ArrayString:
		return len(val.arrayStrVal) == 0
	case MapArrayString:
		return len(val.mapStrVal) == 0
	default:
		return false
	}
}

func (val *Val) EnvVarType() EnvVarType {
	switch val.Type {
	case ArrayNumber, ArrayString:
		return EnvVarTypeArray
	case Bool:
		return EnvVarTypeBoolean
	case Number:
		return EnvVarTypeNumber
	default:
		return EnvVarTypeString
	}
}

// UnmarshalJSON implements the json.Unmarshaller interface.
func (val *Val) UnmarshalJSON(value []byte) error {
	defErr := errors.New("value must be type boolean, number, string, []number, or []string; nested objects are not supported")

	if value[0] == '{' {
		return defErr
	}

	switch value[0] {
	case '{':
		// try to unmarshal map of strings
		if err := json.Unmarshal(value, &val.mapStrVal); err != nil {
			return err
		}
		val.Type = MapArrayString
		return nil

	case '[':
		if value[1] == '"' {
			// then try array of string
			if err := json.Unmarshal(value, &val.arrayStrVal); err != nil {
				return err
			}
			val.Type = ArrayString
			return nil
		} else {
			// try to unmarshal array of numbers
			if err := json.Unmarshal(value, &val.arrayNumVal); err != nil {
				return err
			}
			val.Type = ArrayNumber
			return nil
		}

	case '"':
		if err := json.Unmarshal(value, &val.strVal); err != nil {
			return err
		}
		val.Type = String
		return nil

	case 't', 'f':
		if err := json.Unmarshal(value, &val.boolVal); err != nil {
			return err
		}
		val.Type = Bool
		return nil

	case 'n':
		val.Type = Nil
		return nil

	default:
		if err := json.Unmarshal(value, &val.numVal); err != nil {
			return fmt.Errorf("%s: %w", defErr, err)
		}
		val.Type = Number
		return nil
	}
}

// MarshalJSON implements the json.Marshaller interface.
func (val *Val) MarshalJSON() ([]byte, error) {
	// TODO just use any?
	switch val.Type {
	case Unknown:
		return json.Marshal(nil)
	case Nil:
		return json.Marshal(nil)
	case Bool:
		return json.Marshal(val.boolVal)
	case Number:
		return json.Marshal(val.numVal)
	case String:
		return json.Marshal(val.strVal)
	case ArrayNumber:
		return json.Marshal(val.arrayNumVal)
	case ArrayString:
		return json.Marshal(val.arrayStrVal)
	case MapArrayString:
		return json.Marshal(val.mapStrVal)
	default:
		return []byte{}, fmt.Errorf("impossible var type")
	}
}
