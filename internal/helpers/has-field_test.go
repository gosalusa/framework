package helpers_test

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"gosalusa.com/internal/helpers"
)

func TestHasField(t *testing.T) {
	type A struct {
		Foo string
	}
	type B struct {
		A
	}

	type C struct {
		Foo string `db:"foo"`
	}
	var nilA *A
	var nilC *C

	testCases := []struct {
		Name     string
		Value    any
		Field    string
		Expected bool
	}{
		{"heap find", A{}, "Foo", true},
		{"heap miss", A{}, "Bar", false},
		{"pointer find", &A{}, "Foo", true},
		{"pointer miss", &A{}, "Bar", false},
		{"nil pointer find", nilA, "Foo", true},
		{"nil pointer miss", nilA, "Bar", false},
		{"anonymise find", &B{}, "Foo", true},
		{"anonymise miss", &B{}, "Bar", false},
		{"nil tag find", nilC, "foo", true},
		{"nil tag miss Foo", nilC, "Foo", false},
		{"nil tag miss", nilC, "bar", false},
	}

	for _, tc := range testCases {
		t.Run(tc.Name, func(t *testing.T) {
			ok := helpers.HasField(tc.Value, tc.Field)
			assert.Equal(t, tc.Expected, ok)
		})
	}
}
