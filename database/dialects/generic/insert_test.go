package generic_test

import (
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
	"gosalusa.com/database/dialects"
	"gosalusa.com/database/dialects/generic"
	"gosalusa.com/internal/test"
)

func TestGeneric_EncodeInsertQuery(t *testing.T) {
	g := generic.New(&testCore{})
	test.EncoderTest(t, g.EncodeInsertQuery, []test.EncoderTestCase[*dialects.InsertQuery]{
		{
			Name: "single",
			Builder: &dialects.InsertQuery{
				Table:  "foo",
				Values: []map[string]any{{"a": "b"}},
			},
			ExpectedSQL:      "INSERT INTO `foo` (`a`) VALUES (?)",
			ExpectedBindings: []any{"b"},
		},
		{
			Name: "multi",
			Builder: &dialects.InsertQuery{
				Table:  "foo",
				Values: []map[string]any{{"a": "b", "c": 1}, {"a": "d", "c": 2}},
			},
			ExpectedSQL:      "INSERT INTO `foo` (`a`, `c`) VALUES (?, ?), (?, ?)",
			ExpectedBindings: []any{"b", 1, "d", 2},
		},
	})
}

func TestGeneric_EncodeInsertQueryErrors(t *testing.T) {
	g := generic.New(&testCore{})
	t.Run("no rows", func(t *testing.T) {
		_, err := g.EncodeInsertQuery(&dialects.InsertQuery{
			Table: "foo",
		})
		assert.True(t, errors.Is(err, generic.ErrInsertNoRows))
	})
	t.Run("mismatched keys", func(t *testing.T) {
		_, err := g.EncodeInsertQuery(&dialects.InsertQuery{
			Table:  "foo",
			Values: []map[string]any{{"a": "b", "c": 1}, {"a": "d"}},
		})
		assert.True(t, errors.Is(err, generic.ErrInsertMismatchedValueKeys))
	})
}
