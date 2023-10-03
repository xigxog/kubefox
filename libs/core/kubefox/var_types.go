// +kubebuilder:object:generate=false
package kubefox

import (
	"encoding/json"
	"errors"
	"fmt"
	"strconv"

	"google.golang.org/protobuf/types/known/structpb"
)

// VarType represents the stored type of Var.
type VarType int

const (
	Unknown     VarType = iota // holds an unknown
	Nil                        // holds a null
	Bool                       // holds a boolean
	Number                     // holds an int or float
	String                     // holds a string
	ArrayNumber                // holds an array of ints or floats
	ArrayString                // holds an array of strings
)

// +kubebuilder:object:generate=true
type Var struct {
	booVal      bool      `json:"-"`
	numVal      float64   `json:"-"`
	strVal      string    `json:"-"`
	arrayNumVal []float64 `json:"-"`
	arrayStrVal []string  `json:"-"`
	Type        VarType   `json:"-"`
}

func VarFromValue(val *structpb.Value) (*Var, error) {
	if val == nil {
		return &Var{Type: Nil}, nil
	}
	if v, ok := val.GetKind().(*structpb.Value_BoolValue); ok {
		return NewVarBool(v.BoolValue), nil
	}
	if v, ok := val.GetKind().(*structpb.Value_NumberValue); ok {
		return NewVarFloat(v.NumberValue), nil
	}
	if v, ok := val.GetKind().(*structpb.Value_StringValue); ok {
		return NewVarString(v.StringValue), nil
	}
	if l, ok := val.GetKind().(*structpb.Value_ListValue); ok && l.ListValue != nil && len(l.ListValue.Values) > 0 {
		var numArr []float64
		var strArr []string

		k := l.ListValue.Values[0].GetKind()
		if _, ok := k.(*structpb.Value_NumberValue); ok {
			numArr = make([]float64, len(l.ListValue.Values))
		} else if _, ok := k.(*structpb.Value_StringValue); ok {
			strArr = make([]string, len(l.ListValue.Values))
		} else {
			return &Var{}, fmt.Errorf("list contains values of unsupported types")
		}

		for i, v := range l.ListValue.Values {
			if v == nil {
				return &Var{}, fmt.Errorf("list contains a nil value")
			}
			if v.GetKind() != k {
				return &Var{}, fmt.Errorf("list contains values of mixed types")
			}
			if numArr != nil {
				numArr[i] = v.GetNumberValue()
			} else {
				strArr[i] = v.GetStringValue()
			}
		}

		if numArr != nil {
			return NewVarArrayFloat(numArr), nil
		} else {
			return NewVarArrayString(strArr), nil
		}
	}

	return &Var{}, fmt.Errorf("value is of unsupported type %v", val.GetKind())
}

func NewVarBool(val bool) *Var {
	return &Var{Type: Bool, booVal: val}
}

func NewVarInt(val int) *Var {
	return &Var{Type: Number, numVal: float64(val)}
}

func NewVarFloat(val float64) *Var {
	return &Var{Type: Number, numVal: val}
}

func NewVarString(val string) *Var {
	return &Var{Type: String, strVal: val}
}

func NewVarArrayInt(val []int) *Var {
	arr := make([]float64, len(val))
	for i, v := range val {
		arr[i] = float64(v)
	}
	return &Var{Type: ArrayNumber, arrayNumVal: arr}
}
func NewVarArrayFloat(val []float64) *Var {
	return &Var{Type: ArrayNumber, arrayNumVal: val}
}

func NewVarArrayString(val []string) *Var {
	return &Var{Type: ArrayString, arrayStrVal: val}
}

func (val *Var) Any() any {
	switch val.Type {
	case Bool:
		return val.booVal
	case Number:
		return val.numVal
	case String:
		return val.strVal
	case ArrayNumber:
		return val.arrayNumVal
	case ArrayString:
		return val.arrayStrVal
	default:
		return ""
	}
}

func (val *Var) Value() *structpb.Value {
	switch val.Type {
	case Bool:
		return structpb.NewBoolValue(val.booVal)
	case Number:
		return structpb.NewNumberValue(val.numVal)
	case String:
		return structpb.NewStringValue(val.strVal)
	case ArrayNumber:
		if v, err := structpb.NewValue(val.arrayNumVal); err == nil {
			return v
		}
	case ArrayString:
		if v, err := structpb.NewValue(val.arrayStrVal); err == nil {
			return v
		}
	}

	return structpb.NewNullValue()
}

// Bool returns the boolean value if type is Bool. If type is Number, false will
// be returned if value is 0, otherwise true is returned. If type is String, an
// attempt to parse the boolean value will be made. If parsing fails or type is
// Array false will be returned.
func (val *Var) Bool() bool {
	switch val.Type {
	case Bool:
		return val.booVal
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

func (val *Var) BoolOrDefault(def bool) bool {
	if val.Type != Bool {
		return def
	}
	return val.booVal
}

// Int returns the int value if type is Number. If type is Bool 1 will be
// returned if true, otherwise 0 is returned. If type is String an attempt to
// parse the number will be made. If parsing fails or type is Array 0 will be
// returned.
func (val *Var) Int() int {
	return int(val.Float())
}

func (val *Var) IntOrDefault(def int) int {
	if val.Type != Number {
		return def
	}
	return int(val.numVal)
}

// Float returns the float64 value if type is Number. If type is Bool 1 will be
// returned if true, otherwise 0 is returned. If  type is String an attempt to
// parse the number will be made. If parsing fails or type is Array 0 will be
// returned.
func (val *Var) Float() float64 {
	switch val.Type {
	case Bool:
		if val.booVal {
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

func (val *Var) FloatOrDefault(def float64) float64 {
	if val.Type != Number {
		return def
	}
	return val.numVal
}

// String returns the string value if type is String. If type is Bool the
// `fmt.Sprintf("%t", bool)` of the bool value is returned. If type is Number
// the `fmt.Sprintf("%f", float)` of the number value is returned.
// If type is Array the JSON representation of the array is returned.
func (val *Var) String() string {
	switch val.Type {
	case Bool:
		return fmt.Sprintf("%t", val.booVal)
	case Number:
		return ftos(val.numVal)
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

func (val *Var) StringOrDefault(def string) string {
	if val.Type != String {
		return def
	}
	return val.strVal
}

// ArrayInt returns the array value if type is ArrayNumber. Otherwise nil is
// returned.
func (val *Var) ArrayInt() []int {
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
func (val *Var) ArrayFloat() []float64 {
	if val.Type != ArrayNumber {
		return nil
	}
	return val.arrayNumVal
}

// ArrayString returns the array value if type is ArrayString or ArrayNumber.
// Otherwise nil is returned.
func (val *Var) ArrayString() []string {
	if val.Type == ArrayString {
		return val.arrayStrVal
	}

	if val.Type == ArrayNumber {
		a := make([]string, len(val.arrayNumVal))
		for i, v := range val.arrayNumVal {
			a[i] = ftos(v)
		}
		return a
	}

	return nil
}

func (val *Var) IsUnknown() bool {
	return val.Type == Unknown
}

func (val *Var) IsNil() bool {
	return val.Type == Nil
}

func (val *Var) IsBool() bool {
	return val.Type == Bool
}

func (val *Var) IsString() bool {
	return val.Type == String
}

func (val *Var) IsNumber() bool {
	return val.Type == Number
}

func (val *Var) IsArrayNumber() bool {
	return val.Type == ArrayNumber
}

func (val *Var) IsArrayString() bool {
	return val.Type == ArrayString
}

// UnmarshalJSON implements the json.Unmarshaller interface.
func (val *Var) UnmarshalJSON(value []byte) error {
	defErr := errors.New("value must be type boolean, number, string, []number, or []string; nested objects are not supported")

	if value[0] == '{' {
		return defErr
	}

	switch value[0] {
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

	case 't':
		fallthrough
	case 'f':
		if err := json.Unmarshal(value, &val.booVal); err != nil {
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
func (val *Var) MarshalJSON() ([]byte, error) {
	switch val.Type {
	case Unknown:
		return json.Marshal(nil)
	case Nil:
		return json.Marshal(nil)
	case Bool:
		return json.Marshal(val.booVal)
	case Number:
		return json.Marshal(val.numVal)
	case String:
		return json.Marshal(val.strVal)
	case ArrayNumber:
		return json.Marshal(val.arrayNumVal)
	case ArrayString:
		return json.Marshal(val.arrayStrVal)
	default:
		return []byte{}, fmt.Errorf("impossible var type")
	}
}

func ftos(val float64) string {
	if val == float64(int(val)) {
		// float is an int
		return fmt.Sprintf("%d", int(val))
	} else {
		// float is a float
		return fmt.Sprintf("%f", val)
	}
}
