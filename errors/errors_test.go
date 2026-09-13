package errors_test

import (
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
	salusaerrors "gosalusa.com/errors"
)

func TestSentinelError(t *testing.T) {
	e := salusaerrors.SentinelError("boom")
	assert.Equal(t, "boom", e.Error())
	assert.ErrorIs(t, e, e)
}

func TestNew(t *testing.T) {
	err := salusaerrors.New("some message")
	assert.EqualError(t, err, "some message")

	var stacker salusaerrors.Stacker
	assert.ErrorAs(t, err, &stacker)
	assert.NotNil(t, stacker.Stack())
	assert.NotEmpty(t, stacker.Stack())
}

func TestWithStack(t *testing.T) {
	base := errors.New("base error")
	err := salusaerrors.WithStack(base)
	assert.EqualError(t, err, "base error")
	assert.ErrorIs(t, err, base)
	assert.ErrorIs(t, errors.Unwrap(err), base)
	assert.Equal(t, base, errors.Unwrap(err))

	var stacker salusaerrors.Stacker
	assert.ErrorAs(t, err, &stacker)
	assert.NotEmpty(t, stacker.Stack())
}

func TestAsIs(t *testing.T) {
	var target *MyErr
	err := &MyErr{msg: "nested"}
	wrapped := salusaerrors.WithStack(err)

	assert.True(t, salusaerrors.As(wrapped, &target))
	assert.Equal(t, err, target)

	assert.True(t, salusaerrors.Is(wrapped, err))
	assert.False(t, salusaerrors.Is(wrapped, errors.New("other")))
}

func TestJoin(t *testing.T) {
	a := errors.New("a")
	b := errors.New("b")
	joined := salusaerrors.Join(a, b)
	assert.ErrorIs(t, joined, a)
	assert.ErrorIs(t, joined, b)
	assert.ErrorIs(t, errors.Join(joined), joined)
}

func TestUnwrap(t *testing.T) {
	base := errors.New("base")
	wrapped := salusaerrors.WithStack(base)
	assert.Equal(t, base, salusaerrors.Unwrap(wrapped))
	assert.Nil(t, salusaerrors.Unwrap(salusaerrors.New("plain")))
}

type MyErr struct {
	msg string
}

func (e *MyErr) Error() string { return e.msg }
