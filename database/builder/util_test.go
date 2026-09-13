package builder_test

import (
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/assert"
	"gosalusa.com/database"
	"gosalusa.com/database/builder"
	"gosalusa.com/database/model"
	"gosalusa.com/internal/test"
)

func NewTestBuilder() *builder.ModelBuilder[*test.Foo] {
	return builder.From[*test.Foo]()
}

func MustSave[T model.Model](tx database.DB, v T) T {
	err := model.Save(tx, v)
	if err != nil {
		panic(err)
	}
	return v
}

func assertJsonEqual(t *testing.T, rawJson string, v any) bool {
	b, err := json.Marshal(v)
	if !assert.NoError(t, err) {
		return false
	}
	var data any
	err = json.Unmarshal([]byte(rawJson), &data)
	if !assert.NoError(t, err) {
		return false
	}
	formattedJson, err := json.Marshal(data)
	if !assert.NoError(t, err) {
		return false
	}

	return assert.JSONEq(t, string(formattedJson), string(b))
}
