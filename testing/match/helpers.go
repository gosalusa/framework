package match

import (
	"fmt"
	"reflect"
	"time"
)

const maxMessageSize = 1024

// formatUnequalValues takes two values of arbitrary types and returns string
// representations appropriate to be presented to the user.
//
// If the values are not of like type, the returned strings will be prefixed
// with the type name, and the value will be enclosed in parentheses similar
// to a type conversion in the Go grammar.
func formatUnequalValues(expected, actual any) (e string, a string) {
	if reflect.TypeOf(expected) != reflect.TypeOf(actual) {
		return fmt.Sprintf("%T(%s)", expected, truncatingFormat("%#v", expected)),
			fmt.Sprintf("%T(%s)", actual, truncatingFormat("%#v", actual))
	}
	switch expected.(type) {
	case time.Duration, uint, uint8, uint16, uint32, uint64:
		return fmt.Sprint(expected), fmt.Sprint(actual)
	default:
		return truncatingFormat("%#v", expected), truncatingFormat("%#v", actual)
	}
}

// truncatingFormat formats the data and truncates it if it's too long.
//
// This helps keep formatted error messages lines from exceeding maxMessageSize for readability's sake.
func truncatingFormat(format string, data any) string {
	value := fmt.Sprintf(format, data)
	// Give us space for two truncated objects and the surrounding sentence.
	if len(value) > maxMessageSize {
		value = value[0:maxMessageSize] + "<... truncated>"
	}

	return value
}

// getLen tries to get the length of an object.
//
// It returns (0, false) if impossible.
func getLen(x any) (length int, ok bool) {
	v := reflect.ValueOf(x)
	switch v.Kind() {
	case reflect.Array, reflect.Chan, reflect.Map, reflect.Slice, reflect.String:
		return v.Len(), true
	case reflect.Pointer:
		v = v.Elem()
		if v.Kind() != reflect.Array {
			return 0, false
		}
		return v.Len(), true
	default:
		return 0, false
	}
}
