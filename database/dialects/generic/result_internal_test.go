package generic

import (
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
	"gosalusa.com/database/dialects"
)

func TestRawQueryBuilder(t *testing.T) {
	t.Run("add error", func(t *testing.T) {
		b := newRawQueryBuilder()
		b.Add(dialects.RawQuery{SQL: "x"}, errors.New("boom"))
		assert.Error(t, b.err)
		_, err := b.Build()
		assert.EqualError(t, err, "boom")
	})
	t.Run("add skips empty", func(t *testing.T) {
		b := newRawQueryBuilder()
		b.Add(dialects.RawQuery{}, errors.New("boom"))
		assert.NoError(t, b.err)
	})
	t.Run("add skips after error", func(t *testing.T) {
		b := newRawQueryBuilder()
		b.Add(dialects.RawQuery{SQL: "x"}, errors.New("boom"))
		b.Add(dialects.RawQuery{SQL: "y"}, nil)
		assert.EqualError(t, b.err, "boom")
	})
	t.Run("add stringf", func(t *testing.T) {
		b := newRawQueryBuilder()
		b.AddStringf("%s %d", "foo", 5)
		r, err := b.Build()
		assert.NoError(t, err)
		assert.Equal(t, "foo 5", r.SQL)
	})
	t.Run("add string no space", func(t *testing.T) {
		b := newRawQueryBuilder()
		b.AddString("a")
		b.AddStringNoSpace("b")
		r, err := b.Build()
		assert.NoError(t, err)
		assert.Equal(t, "ab", r.SQL)
	})
	t.Run("add string", func(t *testing.T) {
		b := newRawQueryBuilder()
		b.AddString("a")
		b.AddString("b")
		r, err := b.Build()
		assert.NoError(t, err)
		assert.Equal(t, "a b", r.SQL)
	})
	t.Run("build with bindings", func(t *testing.T) {
		b := newRawQueryBuilder()
		b.Add(dialects.RawQuery{SQL: "a", Bindings: []any{1}}, nil)
		r, err := b.Build()
		assert.NoError(t, err)
		assert.Equal(t, "a", r.SQL)
		assert.Equal(t, []any{1}, r.Bindings)
	})
}

func TestJoinRawQueries(t *testing.T) {
	t.Run("empty", func(t *testing.T) {
		r := joinRawQueries(nil, ", ")
		assert.Equal(t, "", r.SQL)
		assert.Empty(t, r.Bindings)
	})
	t.Run("single", func(t *testing.T) {
		r := joinRawQueries([]dialects.RawQuery{{SQL: "a", Bindings: []any{1}}}, ", ")
		assert.Equal(t, "a", r.SQL)
		assert.Equal(t, []any{1}, r.Bindings)
	})
}

func TestMapJoinRawQueriesError(t *testing.T) {
	_, err := mapJoinRawQueries([]int{1}, ", ", func(v int) (dialects.RawQuery, error) {
		return dialects.RawQuery{}, errors.New("boom")
	})
	assert.EqualError(t, err, "boom")
}
