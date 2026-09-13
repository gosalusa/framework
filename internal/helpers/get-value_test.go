package helpers_test

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"gosalusa.com/internal/helpers"
)

func TestGetValue(t *testing.T) {
	type A struct {
		Foo string
	}
	type B struct {
		A
	}

	type C struct {
		Foo *A `db:"foo"`
	}

	testCases := []struct {
		Name          string
		Value         any
		Field         string
		ExpectedValue any
		ExpectedOk    bool
	}{
		{Name: "stack find", Value: A{Foo: "bar"}, Field: "Foo", ExpectedValue: "bar", ExpectedOk: true},
		{Name: "stack miss", Value: A{}, Field: "Bar", ExpectedValue: nil, ExpectedOk: false},
		{Name: "pointer find", Value: &A{Foo: "bar"}, Field: "Foo", ExpectedValue: "bar", ExpectedOk: true},
		{Name: "pointer miss", Value: &A{}, Field: "Bar", ExpectedValue: nil, ExpectedOk: false},
		{Name: "nil pointer find", Value: (*A)(nil), Field: "Foo", ExpectedValue: nil, ExpectedOk: false},
		{Name: "nil pointer miss", Value: (*A)(nil), Field: "Bar", ExpectedValue: nil, ExpectedOk: false},
		{Name: "anonymise find", Value: &B{A: A{Foo: "bar"}}, Field: "Foo", ExpectedValue: "bar", ExpectedOk: true},
		{Name: "anonymise miss", Value: &B{}, Field: "Bar", ExpectedValue: nil, ExpectedOk: false},
		{Name: "nil tag find", Value: (*C)(nil), Field: "foo", ExpectedValue: nil, ExpectedOk: false},
		{Name: "nil tag miss Foo", Value: (*C)(nil), Field: "Foo", ExpectedValue: nil, ExpectedOk: false},
		{Name: "nil tag miss", Value: (*C)(nil), Field: "bar", ExpectedValue: nil, ExpectedOk: false},
		{Name: "nil value", Value: C{}, Field: "foo", ExpectedValue: (*A)(nil), ExpectedOk: true},
	}

	for _, tc := range testCases {
		t.Run(tc.Name, func(t *testing.T) {
			v, ok := helpers.GetValue(tc.Value, tc.Field)
			assert.Equal(t, tc.ExpectedOk, ok)
			assert.Equal(t, tc.ExpectedValue, v)
		})
	}
}
