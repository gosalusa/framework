package env_test

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"gosalusa.com/env"
)

func TestString(t *testing.T) {
	t.Setenv("TEST_STRING_SET", "hello")
	t.Setenv("TEST_STRING_EMPTY", "")

	assert.Equal(t, "hello", env.String("TEST_STRING_SET", "default"))
	assert.Equal(t, "", env.String("TEST_STRING_EMPTY", "default"))
	assert.Equal(t, "default", env.String("TEST_STRING_MISSING", "default"))
}

func TestBool(t *testing.T) {
	t.Setenv("TEST_BOOL_TRUE", "true")
	t.Setenv("TEST_BOOL_TRUE_CAPS", "TRUE")
	t.Setenv("TEST_BOOL_ONE", "1")
	t.Setenv("TEST_BOOL_FALSE", "false")
	t.Setenv("TEST_BOOL_NO", "nope")

	assert.True(t, env.Bool("TEST_BOOL_TRUE", false))
	assert.True(t, env.Bool("TEST_BOOL_TRUE_CAPS", false))
	assert.True(t, env.Bool("TEST_BOOL_ONE", false))
	assert.False(t, env.Bool("TEST_BOOL_FALSE", true))
	assert.False(t, env.Bool("TEST_BOOL_NO", true))
	assert.True(t, env.Bool("TEST_BOOL_MISSING", true))
}

func TestInt(t *testing.T) {
	t.Setenv("TEST_INT_VALID", "42")
	t.Setenv("TEST_INT_INVALID", "not-an-int")

	assert.Equal(t, 42, env.Int("TEST_INT_VALID", 0))
	assert.Equal(t, 7, env.Int("TEST_INT_INVALID", 7))
	assert.Equal(t, 7, env.Int("TEST_INT_MISSING", 7))
}

func TestFloat64(t *testing.T) {
	t.Setenv("TEST_FLOAT_VALID", "3.14")
	t.Setenv("TEST_FLOAT_INVALID", "not-a-float")

	assert.Equal(t, 3.14, env.Float64("TEST_FLOAT_VALID", 0))
	assert.Equal(t, 2.5, env.Float64("TEST_FLOAT_INVALID", 2.5))
	assert.Equal(t, 2.5, env.Float64("TEST_FLOAT_MISSING", 2.5))
}

func TestFloat32(t *testing.T) {
	t.Setenv("TEST_FLOAT32_VALID", "1.5")
	t.Setenv("TEST_FLOAT32_MISSING", "x")

	assert.Equal(t, float32(1.5), env.Float32("TEST_FLOAT32_VALID", 0))
	assert.Equal(t, float32(1.5), env.Float32("TEST_FLOAT32_MISSING", 1.5))
}
