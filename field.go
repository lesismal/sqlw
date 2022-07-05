// Copyright 2022 lesismal. All rights reserved.
// Use of this source code is governed by an MIT-style
// license that can be found in the LICENSE file.

package sqlw

import (
	"reflect"
	"strconv"
	"strings"
	"sync"
	"time"
	"unsafe"
)

var (
	emptyField = Field{}

	fieldPool = sync.Pool{
		New: func() interface{} {
			return &Field{}
		},
	}

	fieldsPool = sync.Pool{
		New: func() interface{} {
			fields := make([]interface{}, 32)[0:0]
			return &fields
		},
	}
)

// sql.Scanner
type Field struct {
	null bool
	typ  reflect.Kind
	ptr  unsafe.Pointer
}

func (field *Field) Scan(v interface{}) error {
	switch typVal := v.(type) {
	case int64:
		field.null = false
		field.typ = reflect.Int64
		field.ptr = unsafe.Pointer(&typVal)
	case float64:
		field.null = false
		field.typ = reflect.Float64
		field.ptr = unsafe.Pointer(&typVal)
	case bool:
		field.null = false
		field.typ = reflect.Bool
		field.ptr = unsafe.Pointer(&typVal)
	case []byte:
		field.null = false
		field.typ = reflect.Slice
		field.ptr = unsafe.Pointer(&typVal)
	case string:
		field.null = false
		field.typ = reflect.String
		field.ptr = unsafe.Pointer(&typVal)
	case time.Time:
		field.null = false
		field.typ = reflect.Struct
		field.ptr = unsafe.Pointer(&typVal)
	default:
		field.null = true
	}
	return nil
}

var timeTypString = reflect.TypeOf(time.Time{}).String()

func (field *Field) ToValue(dstVal reflect.Value) {
	if !field.null {
		switch typ := dstVal.Type().Kind(); typ {
		case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
			dstVal.SetInt(field.Int64())
		case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
			dstVal.SetUint(field.Uint64())
		case reflect.Float32, reflect.Float64:
			dstVal.SetFloat(field.Float64())
		case reflect.Bool:
			dstVal.SetBool(field.Bool())
		case reflect.Array, reflect.Slice:
			dstVal.SetBytes(field.Bytes())
		case reflect.String:
			dstVal.SetString(field.String())
		case reflect.Struct:
			if dstVal.Type().String() == timeTypString {
				t := field.Time()
				dstVal.Set(reflect.ValueOf(t))
			}
		default:
		}
	}
}

func (field *Field) Int64() int64 {
	if !field.null {
		switch field.typ {
		case reflect.Int64:
			return *(*int64)(field.ptr)
		case reflect.Float64:
			return int64(*(*float64)(field.ptr))
		case reflect.Bool:
			if *(*bool)(field.ptr) {
				return 1
			}
			return 0
		case reflect.Slice: // *[]byte
			s := string(*(*[]byte)(field.ptr))
			if strings.Contains(s, ".") {
				v, err := strconv.ParseFloat(s, 64)
				if err != nil {
					return 0
				}
				return int64(v)
			}
			v, err := strconv.ParseInt(s, 10, 64)
			if err != nil {
				return 0
			}
			return v
		case reflect.String:
			s := *(*string)(field.ptr)
			if strings.Contains(s, ".") {
				v, err := strconv.ParseFloat(s, 64)
				if err != nil {
					return 0
				}
				return int64(v)
			}
			v, err := strconv.ParseInt(s, 10, 64)
			if err != nil {
				return 0
			}
			return v
		case reflect.Struct: // time.Time
			// t := *(*time.Time)(field.ptr)
			// return t.UnixNano()
		default:
		}
	}

	return 0
}

func (field *Field) Uint64() uint64 {
	return uint64(field.Int64())
}

func (field *Field) Float64() float64 {
	if !field.null {
		switch field.typ {
		case reflect.Int64:
			return float64(*(*int64)(field.ptr))
		case reflect.Float64:
			return *(*float64)(field.ptr)
		case reflect.Bool:
			if *(*bool)(field.ptr) {
				return 1
			}
			return 0
		case reflect.Slice: // *[]byte
			s := string(*(*[]byte)(field.ptr))
			v, err := strconv.ParseFloat(s, 64)
			if err != nil {
				return 0
			}
			return v
		case reflect.String:
			s := *(*string)(field.ptr)
			v, err := strconv.ParseFloat(s, 64)
			if err != nil {
				return 0
			}
			return v
		case reflect.Struct: // time.Time
			// t := *(*time.Time)(field.ptr)
			// return float64(t.UnixNano())
		default:
		}
	}

	return 0.0
}

func (field *Field) Bool() bool {
	if !field.null {
		switch field.typ {
		case reflect.Int64:
			return *(*int64)(field.ptr) != 0
		case reflect.Float64:
			return int64(*(*float64)(field.ptr)) != 0
		case reflect.Bool:
			return *(*bool)(field.ptr)
		case reflect.Slice: // *[]byte
			s := string(*(*[]byte)(field.ptr))
			switch s {
			case "", "0", "false":
				return false
			default:
			}
			return true
		case reflect.String:
			s := *(*string)(field.ptr)
			switch s {
			case "", "0", "false":
				return false
			default:
			}
			return true
		case reflect.Struct: // time.Time
			// t := *(*time.Time)(field.ptr)
			// return t.IsZero()
		default:
		}
	}

	return false
}

func (field *Field) Bytes() []byte {
	if !field.null {
		switch field.typ {
		// case reflect.Int64:
		// 	return []byte(fmt.Sprintf("%v", *(*int64)(field.ptr)))
		// case reflect.Float64:
		// 	return []byte(fmt.Sprintf("%v", *(*float64)(field.ptr)))
		// case reflect.Bool:
		// 	return []byte(fmt.Sprintf("%v", *(*bool)(field.ptr)))
		case reflect.Slice: // *[]byte
			return *(*[]byte)(field.ptr)
		case reflect.String:
			s := *(*string)(field.ptr)
			return []byte(s)
		case reflect.Struct: // time.Time
			t := *(*time.Time)(field.ptr)
			return []byte(t.Format(time.RFC3339Nano))
		default:
		}
	}

	return nil
}

func (field *Field) String() string {
	if !field.null {
		switch field.typ {
		// case reflect.Int64:
		// 	return fmt.Sprintf("%v", *(*int64)(field.ptr))
		// case reflect.Float64:
		// 	return fmt.Sprintf("%v", *(*float64)(field.ptr))
		// case reflect.Bool:
		// 	return fmt.Sprintf("%v", *(*bool)(field.ptr))
		case reflect.Slice: // *[]byte
			return string(*(*[]byte)(field.ptr))
		case reflect.String:
			return *(*string)(field.ptr)
		case reflect.Struct: // time.Time
			t := *(*time.Time)(field.ptr)
			return t.Format(time.RFC3339Nano)
		default:
		}
	}

	return ""
}

func (field *Field) Time() time.Time {
	if !field.null {
		switch field.typ {
		case reflect.Int64:
			timestamp := *(*int64)(field.ptr)
			t := time.Unix(timestamp, 0)
			return t
		case reflect.Float64:
		case reflect.Bool:
		case reflect.Slice: // *[]byte
			// YYYY-mm-dd HH:ii:ss
			s := string(*(*[]byte)(field.ptr))
			t, err := time.Parse("2006-01-02 15:04:05", s)
			if err != nil {
				return time.Time{}
			}
			return t
		case reflect.String:
			// YYYY-mm-dd HH:ii:ss
			s := *(*string)(field.ptr)
			t, err := time.Parse("2006-01-02 15:04:05", s)
			if err != nil {
				return time.Time{}
			}
			return t
		case reflect.Struct: // time.Time
			t := *(*time.Time)(field.ptr)
			return t
		default:
		}
	}

	return time.Time{}
}

func releaseFields(fields []interface{}) {
	for _, v := range fields {
		pfield := v.(*Field)
		*pfield = emptyField
		fieldPool.Put(pfield)
	}
	fields = fields[0:0]
	fieldsPool.Put(&fields)
}

func newFields(n int) []interface{} {
	fields := *(fieldsPool.Get().(*[]interface{}))
	for i := 0; i < n; i++ {
		fields = append(fields, fieldPool.Get())
	}
	return fields
}
