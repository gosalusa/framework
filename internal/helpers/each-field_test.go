package helpers_test

import (
	"fmt"
	"reflect"
	"testing"

	"github.com/stretchr/testify/assert"
	"gosalusa.com/internal/helpers"
)

func TestEachField(t *testing.T) {
	type Inner struct {
		Baz string
	}
	type Outer struct {
		Inner
		Foo string
	}

	t.Run("struct", func(t *testing.T) {
		var fields []string
		err := helpers.EachField(reflect.ValueOf(Outer{Foo: "x"}), func(sf reflect.StructField, fv reflect.Value) error {
			fields = append(fields, sf.Name)
			return nil
		})
		assert.NoError(t, err)
		assert.Equal(t, []string{"Baz", "Foo"}, fields)
	})

	t.Run("pointer to struct", func(t *testing.T) {
		var fields []string
		err := helpers.EachField(reflect.ValueOf(&Outer{Foo: "x"}), func(sf reflect.StructField, fv reflect.Value) error {
			fields = append(fields, sf.Name)
			return nil
		})
		assert.NoError(t, err)
		assert.Equal(t, []string{"Baz", "Foo"}, fields)
	})

	t.Run("not a struct", func(t *testing.T) {
		err := helpers.EachField(reflect.ValueOf("string"), func(sf reflect.StructField, fv reflect.Value) error {
			return nil
		})
		assert.ErrorIs(t, err, helpers.ErrExpectedStruct)
	})

	t.Run("callback error", func(t *testing.T) {
		expectedErr := fmt.Errorf("boom")
		err := helpers.EachField(reflect.ValueOf(Outer{}), func(sf reflect.StructField, fv reflect.Value) error {
			return expectedErr
		})
		assert.ErrorIs(t, err, expectedErr)
	})
}

func TestGetFields(t *testing.T) {
	type Inner struct {
		Baz string
		bar string
	}
	type Outer struct {
		Inner
		Foo string
	}

	t.Run("struct", func(t *testing.T) {
		fields := helpers.GetFields(reflect.TypeOf(Outer{}))
		var names []string
		for _, f := range fields {
			names = append(names, f.Name)
		}
		assert.Equal(t, []string{"Baz", "Foo"}, names)
	})

}

func TestEach(t *testing.T) {
	type Foo struct {
		Bar string
	}

	t.Run("struct", func(t *testing.T) {
		called := 0
		err := helpers.Each(Foo{}, func(v reflect.Value, pointer bool) error {
			called++
			assert.False(t, pointer)
			return nil
		})
		assert.NoError(t, err)
		assert.Equal(t, 1, called)
	})

	t.Run("pointer to struct", func(t *testing.T) {
		called := 0
		err := helpers.Each(&Foo{}, func(v reflect.Value, pointer bool) error {
			called++
			assert.True(t, pointer)
			return nil
		})
		assert.NoError(t, err)
		assert.Equal(t, 1, called)
	})

	t.Run("slice", func(t *testing.T) {
		var ptrs []bool
		err := helpers.Each([]Foo{{}, {}}, func(v reflect.Value, pointer bool) error {
			ptrs = append(ptrs, pointer)
			return nil
		})
		assert.NoError(t, err)
		assert.Equal(t, []bool{false, false}, ptrs)
	})

	t.Run("not a struct", func(t *testing.T) {
		err := helpers.Each("string", func(v reflect.Value, pointer bool) error {
			return nil
		})
		assert.NoError(t, err)
	})

	t.Run("callback error", func(t *testing.T) {
		expectedErr := fmt.Errorf("boom")
		err := helpers.Each(Foo{}, func(v reflect.Value, pointer bool) error {
			return expectedErr
		})
		assert.ErrorIs(t, err, expectedErr)
	})

	t.Run("slice callback error", func(t *testing.T) {
		expectedErr := fmt.Errorf("boom")
		err := helpers.Each([]Foo{{}}, func(v reflect.Value, pointer bool) error {
			return expectedErr
		})
		assert.ErrorIs(t, err, expectedErr)
	})
}
