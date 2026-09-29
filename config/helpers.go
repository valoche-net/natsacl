package config

import (
	"fmt"
	"reflect"
	"slices"
)

// mustString returns the string value of an interface, and an error if the interface is not a string
func mustString(name any) (string, error) {
	typ := reflect.TypeOf(name)
	if typ == nil {
		return "", fmt.Errorf("missing kv, obj, service or stream name")
	}
	if kind := typ.Kind(); kind != reflect.String {
		return "", fmt.Errorf("resource name must be a string, got %q", kind)
	}
	return name.(string), nil
}

// GetList returns a list of strings
func getList(key any) (*[]string, error) {
	if key == nil {
		return nil, fmt.Errorf("missing key")
	}
	kind := reflect.TypeOf(key).Kind()
	switch kind {
	case reflect.Slice:
		// Construct a slice of strings
		slice := []string{}
		for _, e := range key.([]any) {
			slice = append(slice, fmt.Sprintf("%v", e))
		}
		if len(slice) > 1 && slices.Contains(slice, "*") {
			return nil, fmt.Errorf(`"*" cannot be combined with other values in %q`, key)
		}
		uniq := make(map[string]any)
		for _, s := range slice {
			uniq[s] = nil
		}
		if len(slice) != len(uniq) {
			return nil, fmt.Errorf("duplicate value in list %q", key)
		}
		return &slice, nil
	}
	return nil, fmt.Errorf("is not a list")
}

// getReadWriteKeys returns a list of keys for read or write
//
// nil means the read/write is absent or false
//
// an empty array means read/write is true
func getReadWriteKeys(key any) (*[]string, error) {
	if key == nil {
		return nil, nil
	}
	kind := reflect.TypeOf(key).Kind()
	switch kind {
	case reflect.Bool:
		if !key.(bool) {
			return nil, nil
		}
		return &[]string{}, nil
	case reflect.Slice:
		return getList(key)
	}
	return nil, nil
}
