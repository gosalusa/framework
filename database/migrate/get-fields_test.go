package migrate_test

import (
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/assert"
	"gosalusa.com/database/migrate"
	"gosalusa.com/database/model"
)

func TestGetFieldsErrors(t *testing.T) {
	t.Run("invalid custom type", func(t *testing.T) {
		type TestModel struct {
			model.BaseModel
			ID int `db:"id,primary,type:bogus"`
		}
		_, err := migrate.New().GenerateMigration("2023-01-01T00:00:00Z create test model", "packageName", &TestModel{})
		assert.Error(t, err)
	})

	t.Run("unsupported kind", func(t *testing.T) {
		type TestModel struct {
			model.BaseModel
			ID    int `db:"id,primary"`
			Value complex128
		}
		_, err := migrate.New().GenerateMigration("2023-01-01T00:00:00Z create test model", "packageName", &TestModel{})
		assert.Error(t, err)
	})
}

func TestGetFieldsDataTypes(t *testing.T) {
	type TestModel struct {
		model.BaseModel
		ID      int             `db:"id,primary"`
		Bytes   []byte          `db:"bytes"`
		Raw     json.RawMessage `db:"raw"`
		Int8    int8            `db:"int8"`
		Int16   int16           `db:"int16"`
		Int64   int64           `db:"int64"`
		Uint8   uint8           `db:"uint8"`
		Uint16  uint16          `db:"uint16"`
		Uint32  uint32          `db:"uint32"`
		Uint64  uint64          `db:"uint64"`
		Float32 float32         `db:"float32"`
		Float64 float64         `db:"float64"`
		Map     map[string]any  `db:"map"`
		Slice   []string        `db:"slice"`
		Struct  struct{ A int } `db:"struct"`
	}
	_, err := migrate.New().GenerateMigration("2023-01-01T00:00:00Z create test model", "packageName", &TestModel{})
	assert.NoError(t, err)
}
