package migrate_test

import (
	"encoding/json"
	"fmt"
	"testing"

	"github.com/bradleyjkemp/cupaloy"
	"github.com/stretchr/testify/assert"
	"gosalusa.com/database/migrate"
	"gosalusa.com/database/model"
)

func TestCreateFromModel(t *testing.T) {
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
		String  string          `db:"string,size:100"`
	}
	got, err := migrate.CreateFromModel(&TestModel{})
	if !assert.NoError(t, err) {
		return
	}
	cupaloy.SnapshotT(t, fmt.Sprintf("%#v", got))

}
