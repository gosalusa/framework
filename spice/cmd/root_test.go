package cmd

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gosalusa.com/static"
)

func chdir(t *testing.T, dir string) {
	t.Helper()
	old, err := os.Getwd()
	require.NoError(t, err)
	require.NoError(t, os.Chdir(dir))
	t.Cleanup(func() {
		require.NoError(t, os.Chdir(old))
	})
}

func writeGoMod(t *testing.T, dir string) {
	t.Helper()
	err := os.WriteFile(filepath.Join(dir, "go.mod"), []byte("module github.com/foo/bar\n\ngo 1.23\n"), 0644)
	require.NoError(t, err)
}

func writeSpiceYML(t *testing.T, dir string) {
	t.Helper()
	err := os.WriteFile(filepath.Join(dir, "spice.yml"), []byte(`module: foo
model:
  dir: app/models
  package: models
  import: foo/app/models
migration:
  dir: migrations
  package: migrations
  import: foo/migrations
`), 0644)
	require.NoError(t, err)
}

func TestCommandsRegistered(t *testing.T) {
	want := []string{"dev", "init", "make:model", "make:migration", "generate:migration"}
	names := map[string]bool{}
	for _, c := range rootCmd.Commands() {
		names[c.Name()] = true
		names[c.Use] = true
	}
	for _, name := range want {
		assert.True(t, names[name], "command %q not registered", name)
	}
}

func TestExecute(t *testing.T) {
	oldArgs := os.Args
	os.Args = []string{"spice"}
	t.Cleanup(func() { os.Args = oldArgs })

	Execute()
}

func TestCopyDir(t *testing.T) {
	t.Run("copies and replaces module path", func(t *testing.T) {
		dist := t.TempDir()
		err := copyDir(static.Content, "template", dist, "github.com/example/app")
		require.NoError(t, err)

		b, err := os.ReadFile(filepath.Join(dist, "routes", "routes.go"))
		require.NoError(t, err)
		assert.Contains(t, string(b), "github.com/example/app/app/handlers")
		assert.NotContains(t, string(b), "gosalusa.com/static/template")
	})
	t.Run("missing source", func(t *testing.T) {
		err := copyDir(static.Content, "does-not-exist", t.TempDir(), "x")
		require.Error(t, err)
	})
	t.Run("mkdir fails", func(t *testing.T) {
		dist := t.TempDir()
		err := os.WriteFile(filepath.Join(dist, "app"), []byte("x"), 0644)
		require.NoError(t, err)

		err = copyDir(static.Content, "template", dist, "x")
		require.Error(t, err)
	})
	t.Run("write file fails", func(t *testing.T) {
		dist := t.TempDir()
		err := os.MkdirAll(filepath.Join(dist, "spice.yml"), 0755)
		require.NoError(t, err)

		err = copyDir(static.Content, "template", dist, "x")
		require.Error(t, err)
	})
}

func TestMakeModelCmd(t *testing.T) {
	dir := t.TempDir()
	writeGoMod(t, dir)
	writeSpiceYML(t, dir)
	chdir(t, dir)

	err := makeModelCmd.RunE(makeModelCmd, []string{"MyModel"})
	require.NoError(t, err)

	b, err := os.ReadFile(filepath.Join(dir, "app", "models", "my-model.go"))
	require.NoError(t, err)
	content := string(b)
	assert.Contains(t, content, "package models")
	assert.Contains(t, content, "type MyModel struct")
	assert.Contains(t, content, "func MyModelQuery")
}

func TestMakeModelCmd_noConfig(t *testing.T) {
	chdir(t, t.TempDir())
	err := makeModelCmd.RunE(makeModelCmd, []string{"MyModel"})
	require.Error(t, err)
}

func TestMakeMigrationCmd(t *testing.T) {
	dir := t.TempDir()
	writeGoMod(t, dir)
	writeSpiceYML(t, dir)
	chdir(t, dir)

	err := makeMigrationCmd.RunE(makeMigrationCmd, []string{"create", "users"})
	require.NoError(t, err)

	entries, err := os.ReadDir(filepath.Join(dir, "migrations"))
	require.NoError(t, err)
	var found bool
	for _, e := range entries {
		if strings.Contains(e.Name(), "create_users.go") {
			found = true
		}
	}
	assert.True(t, found, "migration file create_users.go not created")
}

func TestMakeMigrationCmd_noConfig(t *testing.T) {
	chdir(t, t.TempDir())
	err := makeMigrationCmd.RunE(makeMigrationCmd, []string{"create", "users"})
	require.Error(t, err)
}

func TestGenerateMigrationCmd(t *testing.T) {
	t.Run("invalid line number", func(t *testing.T) {
		dir := t.TempDir()
		writeGoMod(t, dir)
		writeSpiceYML(t, dir)
		chdir(t, dir)

		t.Setenv("GOFILE", "model.go")
		t.Setenv("GOLINE", "abc")

		err := generateCmd.RunE(generateCmd, nil)
		require.Error(t, err)
	})
	t.Run("no struct found", func(t *testing.T) {
		dir := t.TempDir()
		writeGoMod(t, dir)
		writeSpiceYML(t, dir)
		chdir(t, dir)

		modelFile := filepath.Join(dir, "model.go")
		err := os.WriteFile(modelFile, []byte("package main\n\nvar x = 1\n"), 0644)
		require.NoError(t, err)

		t.Setenv("GOFILE", modelFile)
		t.Setenv("GOLINE", "0")

		err = generateCmd.RunE(generateCmd, nil)
		require.Error(t, err)
		assert.Contains(t, err.Error(), "could not find model struct")
	})
	t.Run("missing file", func(t *testing.T) {
		dir := t.TempDir()
		writeGoMod(t, dir)
		writeSpiceYML(t, dir)
		chdir(t, dir)

		t.Setenv("GOFILE", "does-not-exist.go")
		t.Setenv("GOLINE", "0")

		err := generateCmd.RunE(generateCmd, nil)
		require.Error(t, err)
	})
	t.Run("writes migration and runs generator", func(t *testing.T) {
		dir := t.TempDir()
		writeGoMod(t, dir)
		writeSpiceYML(t, dir)
		chdir(t, dir)

		modelFile := filepath.Join(dir, "model.go")
		err := os.WriteFile(modelFile, []byte("package main\n\ntype User struct {\n\tID int\n}\n"), 0644)
		require.NoError(t, err)

		t.Setenv("GOFILE", modelFile)
		t.Setenv("GOLINE", "2")

		err = generateCmd.RunE(generateCmd, nil)
		require.Error(t, err)
	})
	t.Run("migration dir is a file", func(t *testing.T) {
		dir := t.TempDir()
		writeGoMod(t, dir)
		writeSpiceYML(t, dir)
		chdir(t, dir)

		err := os.WriteFile(filepath.Join(dir, "migrations"), []byte("x"), 0644)
		require.NoError(t, err)

		err = generateCmd.RunE(generateCmd, nil)
		require.Error(t, err)
	})
	t.Run("migrations.go is a directory", func(t *testing.T) {
		dir := t.TempDir()
		writeGoMod(t, dir)
		writeSpiceYML(t, dir)
		chdir(t, dir)

		err := os.MkdirAll(filepath.Join(dir, "migrations", "migrations.go"), 0755)
		require.NoError(t, err)

		err = generateCmd.RunE(generateCmd, nil)
		require.Error(t, err)
	})
	t.Run("no config", func(t *testing.T) {
		chdir(t, t.TempDir())
		err := generateCmd.RunE(generateCmd, nil)
		require.Error(t, err)
	})
}

func TestRun(t *testing.T) {
	t.Run("command not found", func(t *testing.T) {
		err := run("definitely-not-a-real-command-xyz")
		require.Error(t, err)
	})
}

func TestDevCmd(t *testing.T) {
	oldArgs := os.Args
	os.Args = []string{"spice", "init"}
	t.Cleanup(func() { os.Args = oldArgs })

	chdir(t, t.TempDir())
	err := devCmd.RunE(devCmd, nil)
	require.NoError(t, err)

	b, err := os.ReadFile(".air.toml")
	require.NoError(t, err)
	assert.NotEmpty(t, b)
}

func TestInitCmd_gomodInitFails(t *testing.T) {
	dir := t.TempDir()
	chdir(t, dir)

	err := os.WriteFile(filepath.Join(dir, "go.mod"), []byte("module invalid\n"), 0644)
	require.NoError(t, err)

	err = initCmd.RunE(initCmd, []string{"github.com/foo/bar"})
	require.Error(t, err)
}
