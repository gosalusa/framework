package request

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
	"gosalusa.com/request/rules"
)

type testValidator struct {
	ok bool
}

func (v *testValidator) Valid() error {
	if v.ok {
		return nil
	}
	return fmt.Errorf("not ok")
}

func Test_Validate_uses_validator_interface(t *testing.T) {
	type Request struct {
		Foo *testValidator `json:"foo" validate:""`
	}

	err := Validate(nil, &Request{Foo: &testValidator{ok: false}})
	assert.Equal(t, ValidationError{
		"foo": []string{"not ok"},
	}, err)

	err = Validate(nil, &Request{Foo: &testValidator{ok: true}})
	assert.NoError(t, err)
}

func Test_Validate_ignores_unknown_rules(t *testing.T) {
	type Request struct {
		Foo int `validate:"does_not_exist"`
	}

	err := Validate(nil, &Request{Foo: 1})
	assert.NoError(t, err)
}

func Test_Validate_uses_di_tag_name(t *testing.T) {
	rules.AddRule("should_fail_di", func(*rules.ValidationOptions) bool {
		return false
	})

	type Request struct {
		Foo int `di:"bar" validate:"should_fail_di"`
	}

	err := Validate(nil, &Request{})
	assert.Equal(t, ValidationError{
		"bar": []string{"should_fail_di"},
	}, err)
}

func Test_Validate_uses_required_pass(t *testing.T) {
	type Request struct {
		Foo int `validate:"required"`
	}

	err := Validate(nil, &Request{Foo: 1})
	assert.NoError(t, err)
}

func Test_Validate_returns_getMessage_error(t *testing.T) {
	rules.AddRule("prohibits", func(*rules.ValidationOptions) bool {
		return false
	})

	type Request struct {
		Foo string `validate:"prohibits"`
	}

	err := Validate(nil, &Request{Foo: "x"})
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "failed to execute template")
}

func Test_Validate_nested_non_struct_anonymous(t *testing.T) {
	type Request struct {
		Foo int `json:"foo"`
	}
	_ = Request{}
	type weird struct {
		Request
		int
	}

	err := Validate(nil, &weird{})
	assert.Error(t, err)
}
