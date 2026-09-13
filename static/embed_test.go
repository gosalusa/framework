package static_test

import (
	"io/fs"
	"testing"
	"testing/fstest"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gosalusa.com/static"
)

func TestContent(t *testing.T) {
	err := fstest.TestFS(static.Content, "template/main.go", "template/spice.yml", "template/routes/routes.go")
	require.NoError(t, err)

	b, err := fs.ReadFile(static.Content, "template/main.go")
	require.NoError(t, err)
	assert.Contains(t, string(b), "package main")
}
