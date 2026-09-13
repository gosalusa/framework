package validate_test

import (
	"context"
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
	"gosalusa.com/validate"
)

type fakeValidator struct {
	err error
}

func (f fakeValidator) Validate(ctx context.Context) error {
	return f.err
}

type multiError struct {
	errs []error
}

func (m multiError) Error() string {
	return "multi"
}

func (m multiError) Unwrap() []error {
	return m.errs
}

func TestAppend(t *testing.T) {
	t.Run("nil err nil validator", func(t *testing.T) {
		err := validate.Append(context.Background(), nil, fakeValidator{})
		assert.NoError(t, err)
	})

	t.Run("nil err failing validator", func(t *testing.T) {
		fail := errors.New("fail")
		err := validate.Append(context.Background(), nil, fakeValidator{err: fail})
		assert.ErrorIs(t, err, fail)
	})

	t.Run("existing err and nil validator", func(t *testing.T) {
		base := errors.New("base")
		err := validate.Append(context.Background(), base, fakeValidator{})
		assert.ErrorIs(t, err, base)
	})

	t.Run("multi error unwrap", func(t *testing.T) {
		a := errors.New("a")
		b := errors.New("b")
		base := multiError{errs: []error{a, b}}
		c := errors.New("c")
		err := validate.Append(context.Background(), base, fakeValidator{err: c})
		assert.ErrorIs(t, err, a)
		assert.ErrorIs(t, err, b)
		assert.ErrorIs(t, err, c)
	})
}
