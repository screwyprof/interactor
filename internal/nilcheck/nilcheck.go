// Package nilcheck holds the one definition of "nil" for use cases, shared by the dispatcher
// and testdsl so they reject the same values.
package nilcheck

import "reflect"

// IsNil reports whether v is a nil interface, nil pointer or nil func.
func IsNil(v any) bool {
	if v == nil {
		return true
	}

	rv := reflect.ValueOf(v)

	return (rv.Kind() == reflect.Pointer || rv.Kind() == reflect.Func) && rv.IsNil()
}
