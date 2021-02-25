package influxdb

import (
	"errors"
	"fmt"
)

type FieldValue interface {
	IsBool() bool
	IsString() bool
	IsFloat64() bool
	IsInt64() bool
	IsUint64() bool

	Bool() bool
	String() string
	Float64() float64
	Int64() int64
	Uint64() uint64

	SetBool(bool)
	SetString(string)
	SetFloat64(float64)
	SetInt64(int64)
	SetUint64(uint64)

	Marshal() ([]byte, error)
}

const (
	boolFlag   = 0b00000001
	stringFlag = 0b00000010
	floatFlag  = 0b00000100
	intFlag    = 0b00001000
	uintFlag   = 0b00010000
)

// NewFieldValue takes in a bool, float, int, uint or string and returns it as a usable FieldValue
//
// float32 is casted to float64
// any int type other than int64 is casted to int64
// any uint type other than uint64 is casted to uint64
// []byte is casted to string
func NewFieldValue(v interface{}) FieldValue {
	fv := fieldValue{}
	switch v.(type) {
	case bool:
		fv.SetBool(v.(bool))
	case []byte:
		fv.SetString(string(v.([]byte)))
	case string:
		fv.SetString(v.(string))
	case float32:
		fv.SetFloat64(float64(v.(float32)))
	case float64:
		fv.SetFloat64(v.(float64))
	case int:
		fv.SetInt64(int64(v.(int)))
	case int8:
		fv.SetInt64(int64(v.(int8)))
	case int16:
		fv.SetInt64(int64(v.(int16)))
	case int32:
		fv.SetInt64(int64(v.(int32)))
	case int64:
		fv.SetInt64(v.(int64))
	case uint:
		fv.SetUint64(uint64(v.(uint)))
	case uint8:
		fv.SetUint64(uint64(v.(uint8)))
	case uint16:
		fv.SetUint64(uint64(v.(uint16)))
	case uint32:
		fv.SetUint64(uint64(v.(uint32)))
	case uint64:
		fv.SetUint64(v.(uint64))
	default:
		panic("unsupported FieldValue type")
	}

	return &fv
}

type fieldValue struct {
	value     interface{}
	valueFlag uint8
}

func (f fieldValue) IsBool() bool {
	return f.valueFlag&boolFlag > 0
}

func (f fieldValue) IsString() bool {
	return f.valueFlag&stringFlag > 0
}

func (f fieldValue) IsFloat64() bool {
	return f.valueFlag&floatFlag > 0
}

func (f fieldValue) IsInt64() bool {
	return f.valueFlag&intFlag > 0
}

func (f fieldValue) IsUint64() bool {
	return f.valueFlag&uintFlag > 0
}

func (f *fieldValue) SetBool(b bool) {
	f.value = b
	f.valueFlag = boolFlag
}

func (f *fieldValue) SetString(s string) {
	f.value = s
	f.valueFlag = stringFlag
}

func (f *fieldValue) SetFloat64(v float64) {
	f.value = v
	f.valueFlag = floatFlag
}

func (f *fieldValue) SetInt64(i int64) {
	f.value = i
	f.valueFlag = intFlag
}

func (f *fieldValue) SetUint64(u uint64) {
	f.value = u
	f.valueFlag = uintFlag
}

func (f fieldValue) Bool() bool {
	if f.valueFlag&boolFlag == 0 {
		panic("value is not bool")
	}
	return f.value.(bool)
}

func (f fieldValue) String() string {
	if f.valueFlag&stringFlag == 0 {
		panic("value is not string")
	}
	return f.value.(string)
}

func (f fieldValue) Float64() float64 {
	if f.valueFlag&floatFlag == 0 {
		panic("value is not float64")
	}
	return f.value.(float64)
}

func (f fieldValue) Int64() int64 {
	if f.valueFlag&intFlag == 0 {
		panic("value is not int64")
	}
	return f.value.(int64)
}

func (f fieldValue) Uint64() uint64 {
	if f.valueFlag&uintFlag == 0 {
		panic("value is not uint64")
	}
	return f.value.(uint64)
}

func (f fieldValue) Marshal() ([]byte, error) {
	switch f.value.(type) {
	case bool:
		return []byte(fmt.Sprintf("%t", f.value.(bool))), nil
	case string:
		return []byte(fmt.Sprintf(`"%s"`, f.value.(string))), nil
	case float64:
		return []byte(fmt.Sprintf("%f", f.value.(float64))), nil
	case int:
		return []byte(fmt.Sprintf("%di", f.value.(int64))), nil
	case uint:
		return []byte(fmt.Sprintf("%du", f.value.(uint64))), nil
	default:
		return nil, errors.New("Unsupported fieldValue type")
	}
}
