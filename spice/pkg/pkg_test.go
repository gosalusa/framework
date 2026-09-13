package pkg_test

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gosalusa.com/spice/pkg"
)

func TestPackage_Add(t *testing.T) {
	p := pkg.New()
	require.NotNil(t, p)

	fc := p.Add("gosalusa.com/database/schema.Create", "users")
	require.NotNil(t, fc)
	assert.Contains(t, fc.GoString(), "gosalusa_com_database_schema.Create(\"users\")")
}

func TestPackage_GoString(t *testing.T) {
	t.Run("no calls", func(t *testing.T) {
		p := pkg.New()
		assert.Equal(t, "package main\n\nimport (\n)\n\nfunc main() {\n}\n", p.GoString())
	})
	t.Run("with function call", func(t *testing.T) {
		p := pkg.New()
		p.Add("fmt.Println", "hello", 42)
		src := p.GoString()
		assert.Contains(t, src, `import (
	"fmt"
)`)
		assert.Contains(t, src, `	fmt.Println("hello", 42)`)
	})
	t.Run("add and run", func(t *testing.T) {
		p := pkg.New()
		fc := p.Add("fmt.Println")
		assert.NotNil(t, fc)
		assert.NoError(t, p.Run())
	})
}

func TestFunctionCall_GoString(t *testing.T) {
	t.Run("with error", func(t *testing.T) {
		p := pkg.New()
		fc := p.Add("os.WriteFile", "file.txt", pkg.Raw(`[]byte("hello")`), 0644)
		fc.WithError()
		fc.ReturnCount(1)

		src := fc.GoString()
		assert.Contains(t, src, "err = ")
		assert.Contains(t, src, `os.WriteFile("file.txt", []byte("hello"), 420)`)
		assert.Contains(t, src, "\tif err != nil {\n")
		assert.Contains(t, src, "\t\tpanic(err)\n")
	})
	t.Run("with multiple return values", func(t *testing.T) {
		p := pkg.New()
		fc := p.Add("os.Create", "file.txt")
		fc.WithError()
		fc.ReturnCount(2)
		fc.ReturnVariable(0, "f")
		fc.ReturnVariable(1, "err")

		src := fc.GoString()
		assert.Contains(t, src, "_, err = ")
	})
	t.Run("no error", func(t *testing.T) {
		p := pkg.New()
		fc := p.Add("fmt.Println", "hi")

		src := fc.GoString()
		assert.Contains(t, src, `fmt.Println("hi")`)
		assert.NotContains(t, src, "if err != nil")
	})
	t.Run("args with go stringer", func(t *testing.T) {
		p := pkg.New()
		fc := p.Add("fmt.Println", pkg.Raw("1 + 1"))

		src := fc.GoString()
		assert.Contains(t, src, `fmt.Println(1 + 1)`)
	})
	t.Run("import with special chars", func(t *testing.T) {
		p := pkg.New()
		p.Add("github.com/foo/bar.baz", "x")

		src := p.GoString()
		assert.Contains(t, src, "github_com_foo_bar.baz(")
		assert.Contains(t, src, `github.com/foo/bar`)
	})
}

func TestRaw(t *testing.T) {
	r := pkg.Raw("hello world")
	assert.Equal(t, "hello world", r.GoString())
}

func TestGoStringFunc(t *testing.T) {
	f := pkg.GoStringFunc(func() string {
		return "custom"
	})
	assert.Equal(t, "custom", f.GoString())
}
