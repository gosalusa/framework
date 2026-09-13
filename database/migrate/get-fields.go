package migrate

import (
	"encoding/json"
	"fmt"
	"reflect"
	"time"

	"gosalusa.com/database/builder"
	"gosalusa.com/database/dialects"
	"gosalusa.com/database/model"
	"gosalusa.com/internal/helpers"
)

type field struct {
	dataType dialects.DataType
	tag      *helpers.Tag
	nullable bool
	relation builder.Relationship
}

func getFields(m model.Model) ([]*field, error) {
	fields := []*field{}
	relationshipInterface := reflect.TypeOf((*builder.Relationship)(nil)).Elem()
	err := helpers.EachField(reflect.ValueOf(m), func(sf reflect.StructField, fv reflect.Value) error {
		if !sf.IsExported() {
			return nil
		}
		if sf.Type.Implements(relationshipInterface) {
			fields = append(fields, &field{
				relation: fv.Interface().(builder.Relationship),
			})

			return nil
		}
		tag := helpers.DBTag(sf)
		if tag.Name == "-" {
			return nil
		}

		f := &field{
			tag:      tag,
			nullable: tag.Nullable,
		}
		t := sf.Type
		if t.Kind() == reflect.Pointer {
			t = t.Elem()
			fv = reflect.New(t).Elem()
			f.nullable = true
		}

		if tag.Type != "" {
			f.dataType = dialects.DataType{
				Name: tag.Type,
			}

			// if !f.dataType.IsValid() {
			// 	return fmt.Errorf("data type %s is not valid", tag.Type)
			// }
		} else {
			switch field := fv.Interface().(type) {
			case dialects.DataTyper:
				f.dataType = field.DataType()
			case time.Time:
				f.dataType = dialects.DataTypeDateTime
			case []byte:
				f.dataType = dialects.DataTypeBlob
			case json.RawMessage:
				f.dataType = dialects.DataTypeJSON
			default:
				switch t.Kind() {
				case reflect.Bool:
					f.dataType = dialects.DataTypeBoolean
				case reflect.Int8:
					f.dataType = dialects.DataTypeInt8
				case reflect.Int16:
					f.dataType = dialects.DataTypeInt16
				case reflect.Int, reflect.Int32:
					f.dataType = dialects.DataTypeInt32
				case reflect.Int64:
					f.dataType = dialects.DataTypeInt64
				case reflect.Uint8:
					f.dataType = dialects.DataTypeUInt8
				case reflect.Uint16:
					f.dataType = dialects.DataTypeUInt16
				case reflect.Uint, reflect.Uint32:
					f.dataType = dialects.DataTypeUInt32
				case reflect.Uint64:
					f.dataType = dialects.DataTypeUInt64
				case reflect.Float32:
					f.dataType = dialects.DataTypeFloat32
				case reflect.Float64:
					f.dataType = dialects.DataTypeFloat64
				case reflect.String:
					f.dataType = dialects.DataTypeString
				case reflect.Map, reflect.Slice, reflect.Struct:
					f.dataType = dialects.DataTypeJSON
				case reflect.Array:
					f.dataType = dialects.DataTypeBlob
				default:
					return fmt.Errorf("no datatype for %v", t.Kind())
				}
			}
		}

		if tag.Size != 0 {
			f.dataType.Size = tag.Size
		}

		fields = append(fields, f)
		return nil
	})
	if err != nil {
		return nil, err
	}
	return fields, nil
}
