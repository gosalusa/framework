package helpers_test

import (
	"reflect"
	"testing"

	"github.com/stretchr/testify/assert"
	"gosalusa.com/internal/helpers"
)

func TestName(t *testing.T) {
	type Foo struct {
		ID  int
		Foo string `db:"foo,autoincrement,primary,type:date"`
		Bar string `db:"bar,readonly"`
		Baz string `db:"baz,size:100"`
	}

	rt := reflect.TypeOf(Foo{})
	assert.Equal(
		t,
		&helpers.Tag{
			Name:          "ID",
			Primary:       false,
			AutoIncrement: false,
			Readonly:      false,
			Index:         false,
		},
		helpers.DBTag(rt.Field(0)),
	)
	assert.Equal(
		t,
		&helpers.Tag{
			Name:          "foo",
			Primary:       true,
			AutoIncrement: true,
			Readonly:      false,
			Index:         false,
			Type:          "date",
		},
		helpers.DBTag(rt.Field(1)),
	)
	assert.Equal(
		t,
		&helpers.Tag{
			Name:          "bar",
			Primary:       false,
			AutoIncrement: false,
			Readonly:      true,
			Index:         false,
		},
		helpers.DBTag(rt.Field(2)),
	)
	assert.Equal(
		t,
		&helpers.Tag{
			Name:          "baz",
			Primary:       false,
			AutoIncrement: false,
			Readonly:      false,
			Index:         false,
			Size:          100,
		},
		helpers.DBTag(rt.Field(3)),
	)
}
