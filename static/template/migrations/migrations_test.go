package migrations_test

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gosalusa.com/static/template/migrations"
)

func TestUse(t *testing.T) {
	m := migrations.Use()
	require.NotNil(t, m)
	assert.NotNil(t, m)
}
