package util

import (
	"reflect"
)

func DerefString(str *string) string {
	if str != nil {
		return *str
	}
	return ""
}

func DedupeSlice(slice []string) []string {
	hash := make(map[string]int)

	for _, em := range slice {
		if hash[em] == 1 {
			continue
		}
		hash[em] = 1
	}

	deduped := make([]string, 0, len(hash))
	for val, _ := range hash {
		deduped = append(deduped, val)
	}
	return deduped
}

func TypeOf(i interface{}) string {
	if t := reflect.TypeOf(i); t.Kind() == reflect.Ptr {
		return t.Elem().Name()
	} else {
		return t.Name()
	}
}

func Empty(i interface{}) bool {
	switch v := i.(type) {
	case string:
		return v == ""
	case *string:
		return v == nil
	case *[]string:
		return v == nil
	default:
		return reflect.ValueOf(v).IsNil()
	}
}

func StrPtr(str string) *string {
	return &str
}

func IntPtr(i int) *int {
	return &i
}
