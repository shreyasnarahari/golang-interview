package practice

import (
	"reflect"
	"unsafe"
)

// ZeroCopyStringToBytes performs zero-allocation string to byte slice conversion (Go 1.20+).
func ZeroCopyStringToBytes(s string) []byte {
	if s == "" {
		return nil
	}
	return unsafe.Slice(unsafe.StringData(s), len(s))
}

// GetStructFieldTag retrieves the 'json' tag of a struct field by name.
func GetStructFieldTag(structObj any, fieldName string) string {
	t := reflect.TypeOf(structObj)
	if t.Kind() == reflect.Ptr {
		t = t.Elem()
	}
	if field, ok := t.FieldByName(fieldName); ok {
		return field.Tag.Get("json")
	}
	return ""
}
