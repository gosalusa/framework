package helpers_test

import (
	"reflect"
	"testing"

	"github.com/stretchr/testify/assert"
	"gosalusa.com/internal/helpers"
)

func TestPrimaryKey(t *testing.T) {

	type NoTag struct {
		NoTag int
	}
	type WithTag struct {
		Primary    int `db:"with_tag"`
		NotPrimary int `db:"not_primary"`
	}
	type WithTagAndPrimary struct {
		NotPrimary int `db:"not_primary"`
		Primary    int `db:"with_tag_and_primary,primary"`
	}
	type Composite struct {
		NotPrimary int `db:"not_primary"`
		Primary1   int `db:"primary1,primary"`
		Primary2   int `db:"primary2,primary"`
	}

	testCases := []struct {
		name               string
		model              any
		expectedPrimaryKey []string
	}{
		{"No Tag", &NoTag{}, []string{"NoTag"}},
		{"With Tag", &WithTag{}, []string{"with_tag"}},
		{"With Tag And Primary", &WithTagAndPrimary{}, []string{"with_tag_and_primary"}},
		{"Composite", &Composite{}, []string{"primary1", "primary2"}},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			assert.Equal(t, tc.expectedPrimaryKey, helpers.PrimaryKey(tc.model))
		})
	}
}

type primaryKeyer struct{}

func (primaryKeyer) PrimaryKey() []string {
	return []string{"email"}
}

type withPrimaryKeyer struct {
	primaryKeyer
	ID    int    `db:"id"`
	Name  string `db:"name"`
	Email string `db:"email"`
}

type withPrimaryTag struct {
	ID    int `db:"id,primary"`
	Name  string
	Email string `db:"email,unique"`
}

type badKeyer struct {
	ID int `db:"id"`
}

func (badKeyer) PrimaryKey() []string {
	return []string{"nope"}
}

func TestPrimaryKeyCustom(t *testing.T) {
	assert.Equal(t, []string{"email"}, helpers.PrimaryKey(withPrimaryKeyer{}))
}

func TestPrimaryKeyFields(t *testing.T) {
	t.Run("primary keyer", func(t *testing.T) {
		fields := helpers.PrimaryKeyFields(withPrimaryKeyer{})
		assert.Len(t, fields, 1)
		assert.Equal(t, "email", helpers.DBTag(fields[0]).Name)
	})

	t.Run("tagged", func(t *testing.T) {
		fields := helpers.RPrimaryKeyFields(reflect.TypeOf(withPrimaryTag{}))
		assert.Len(t, fields, 1)
		assert.Equal(t, "id", helpers.DBTag(fields[0]).Name)
	})

}

func TestPrimaryKeyValue(t *testing.T) {
	t.Run("found", func(t *testing.T) {
		values, err := helpers.PrimaryKeyValue(withPrimaryTag{ID: 42})
		assert.NoError(t, err)
		assert.Equal(t, []any{42}, values)
	})

	t.Run("not found", func(t *testing.T) {
		_, err := helpers.PrimaryKeyValue(badKeyer{ID: 42})
		assert.Error(t, err)
	})

	t.Run("nil pointer", func(t *testing.T) {
		type nilKeyer struct {
			ID int `db:"id"`
		}
		_, err := helpers.PrimaryKeyValue((*nilKeyer)(nil))
		assert.Error(t, err)
	})
}

func TestIncludes(t *testing.T) {
	assert.True(t, helpers.Includes([]int{1, 2, 3}, 2))
	assert.False(t, helpers.Includes([]int{1, 2, 3}, 4))
	assert.True(t, helpers.Includes([]string{"a"}, "a"))
	assert.False(t, helpers.Includes([]string{"a"}, "b"))
}
