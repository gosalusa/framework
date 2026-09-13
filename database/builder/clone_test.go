package builder_test

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"gosalusa.com/database/builder"
)

func TestBuilder_Clone(t *testing.T) {
	b := builder.NewBuilder().
		Select("c1", "c2").Distinct().
		Join("t1", "t1_id", "=", "id").
		Where("a", "!=", 5).
		GroupBy("c1").
		Having("c1", "=", 7).
		OrderBy("c2").
		Limit(1).Offset(10)

	clone := b.Clone()
	assert.NotSame(t, b, clone)
	assert.Equal(t, b, clone)
}
