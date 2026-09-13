package util_test

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gosalusa.com/spice/util"
)

func TestMigrationName(t *testing.T) {
	name := util.MigrationName([]string{"create", "users", "table"})
	assert.Contains(t, name, "create_users_table")
	assert.Len(t, name, len("20060102_150405")+len("create_users_table")+1)
}

func TestMigrationName_spaces(t *testing.T) {
	name := util.MigrationName([]string{"my", "migration", "two words"})
	assert.Contains(t, name, "my_migration_two_words")
}

func TestFileDir(t *testing.T) {
	dir := t.TempDir()
	err := os.WriteFile(filepath.Join(dir, "spice.yml"), []byte("module: foo\n"), 0644)
	require.NoError(t, err)

	sub := filepath.Join(dir, "app", "models")
	err = os.MkdirAll(sub, 0755)
	require.NoError(t, err)

	got, err := util.FileDir(sub, "spice.yml")
	require.NoError(t, err)
	assert.Equal(t, dir, filepath.Clean(got))
}

func TestFileDir_notFound(t *testing.T) {
	_, err := util.FileDir(t.TempDir(), "does-not-exist.txt")
	require.Error(t, err)
	assert.ErrorIs(t, err, util.ErrNotFound)
}

func TestFileDir_invalidDir(t *testing.T) {
	_, err := util.FileDir("/definitely/not/a/real/dir", "spice.yml")
	require.Error(t, err)
}

func TestPkgInfo(t *testing.T) {
	dir := t.TempDir()
	err := os.WriteFile(filepath.Join(dir, "go.mod"), []byte("module github.com/foo/bar\n\ngo 1.23\n"), 0644)
	require.NoError(t, err)

	info, err := util.PkgInfo(dir, dir)
	require.NoError(t, err)
	assert.Equal(t, "github.com/foo/bar", info.RootPackage)
	assert.Equal(t, "github.com/foo/bar", info.SpicePackage)
}

func TestPkgInfo_subdir(t *testing.T) {
	dir := t.TempDir()
	err := os.WriteFile(filepath.Join(dir, "go.mod"), []byte("module github.com/foo/bar\n\ngo 1.23\n"), 0644)
	require.NoError(t, err)
	sub := filepath.Join(dir, "app", "models")
	err = os.MkdirAll(sub, 0755)
	require.NoError(t, err)

	info, err := util.PkgInfo(dir, sub)
	require.NoError(t, err)
	assert.Equal(t, "github.com/foo/bar", info.RootPackage)
	assert.Equal(t, filepath.Join("github.com/foo/bar", "app", "models"), info.SpicePackage)
}

func TestPkgInfo_missingGoMod(t *testing.T) {
	_, err := util.PkgInfo(t.TempDir(), t.TempDir())
	require.Error(t, err)
}

func TestLoadConfig(t *testing.T) {
	dir := t.TempDir()
	err := os.WriteFile(filepath.Join(dir, "go.mod"), []byte("module github.com/foo/bar\n\ngo 1.23\n"), 0644)
	require.NoError(t, err)
	err = os.WriteFile(filepath.Join(dir, "spice.yml"), []byte("module: foo\nmodel:\n  dir: app/models\n"), 0644)
	require.NoError(t, err)

	c, err := util.LoadConfig(dir)
	require.NoError(t, err)
	require.NotNil(t, c)

	assert.Equal(t, "foo", c.Module)
	assert.Equal(t, dir, c.Root)
	require.NotNil(t, c.Model)
	assert.Equal(t, filepath.Join(dir, "app", "models"), c.Model.Dir)
	assert.Equal(t, "models", c.Model.Pkg)
	assert.Equal(t, filepath.Join("github.com/foo/bar", "app", "models"), c.Model.Import)
}

func TestLoadConfig_defaults(t *testing.T) {
	dir := t.TempDir()
	err := os.WriteFile(filepath.Join(dir, "go.mod"), []byte("module github.com/foo/bar\n\ngo 1.23\n"), 0644)
	require.NoError(t, err)
	err = os.WriteFile(filepath.Join(dir, "spice.yml"), []byte("module: foo\n"), 0644)
	require.NoError(t, err)

	c, err := util.LoadConfig(dir)
	require.NoError(t, err)
	require.NotNil(t, c.Model)
	assert.Equal(t, filepath.Join(dir, "app", "models"), c.Model.Dir)
	assert.Equal(t, filepath.Join(dir, "migrations"), c.Migration.Dir)
	assert.Equal(t, "migrations", c.Migration.Pkg)
}

func TestLoadConfig_missingConfig(t *testing.T) {
	_, err := util.LoadConfig(t.TempDir())
	require.Error(t, err)
}

func TestLoadConfig_missingGoMod(t *testing.T) {
	dir := t.TempDir()
	err := os.WriteFile(filepath.Join(dir, "spice.yml"), []byte("module: foo\n"), 0644)
	require.NoError(t, err)

	_, err = util.LoadConfig(dir)
	require.Error(t, err)
}

func TestLoadConfig_badYAML(t *testing.T) {
	dir := t.TempDir()
	err := os.WriteFile(filepath.Join(dir, "go.mod"), []byte("module github.com/foo/bar\n\ngo 1.23\n"), 0644)
	require.NoError(t, err)
	err = os.WriteFile(filepath.Join(dir, "spice.yml"), []byte("module: [unclosed"), 0644)
	require.NoError(t, err)

	_, err = util.LoadConfig(dir)
	require.Error(t, err)
}

func TestErrNotFound(t *testing.T) {
	assert.Error(t, util.ErrNotFound)
}
