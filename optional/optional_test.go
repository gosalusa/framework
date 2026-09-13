package optional_test

import (
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/assert"
	"gosalusa.com/optional"
)

func TestSome(t *testing.T) {
	o := optional.Some(42)
	assert.Equal(t, 42, o.Value)
	assert.True(t, o.Valid)

	s := optional.Some("hello")
	assert.Equal(t, "hello", s.Value)
	assert.True(t, s.Valid)
}

func TestNone(t *testing.T) {
	var o optional.Option[int]
	assert.False(t, o.Valid)
	assert.Equal(t, 0, o.Value)

	n := optional.None[string]()
	assert.False(t, n.Valid)
	assert.Equal(t, "", n.Value)
}

func TestOptionJSONPanics(t *testing.T) {
	o := optional.Some(1)

	var _ json.Marshaler = o
	var _ json.Unmarshaler = &o

	assert.Panics(t, func() {
		o.MarshalJSON()
	})
	assert.Panics(t, func() {
		o.UnmarshalJSON(nil)
	})
	assert.Panics(t, func() {
		o.MarshalBinary()
	})
	assert.Panics(t, func() {
		o.UnmarshalBinary(nil)
	})
	assert.Panics(t, func() {
		o.MarshalText()
	})
	assert.Panics(t, func() {
		o.UnmarshalText(nil)
	})
}
