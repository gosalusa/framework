package builder

import (
	"context"
	"encoding/json"
	"fmt"
	"reflect"

	"gosalusa.com/database"
	"gosalusa.com/internal/helpers"
	"gosalusa.com/internal/relationship"
)

// ForeignKey describes how two tables are related: the local column on the
// parent model and the column on the related table it references.
type ForeignKey struct {
	LocalKey     string
	RelatedTable string
	RelatedKey   string
}

// Equal reports whether v references the same columns and tables as f.
func (f *ForeignKey) Equal(v *ForeignKey) bool {
	return f.LocalKey == v.LocalKey && f.RelatedKey == v.RelatedKey && f.RelatedTable == v.RelatedTable
}

// Relationship is a model relationship that can be initialized from a parent
// model, give back a query for the related records, and load those records from
// the database.
type Relationship interface {
	relationship.Relationship
	Subquery() *Builder
	Load(ctx context.Context, tx database.DB, relations []Relationship) error
	ForeignKeys() []*ForeignKey
}

type relationValue[T any] struct {
	loaded bool
	value  T
}

var (
	// ErrMissingRelationship is returned when a model does not have the given
	// relationship.
	ErrMissingRelationship = fmt.Errorf("missing relationship")
	// ErrMissingField is returned when a model does not have the given field.
	ErrMissingField = fmt.Errorf("missing related field")
)

// Value will return the related value and if it has been fetched.
func (v *relationValue[T]) Value() (T, bool) {
	return v.value, v.loaded
}

// Loaded returns true if the relationship has been fetched and false if it has
// not.
func (v *relationValue[T]) Loaded() bool {
	return v.loaded
}

func (v *relationValue[T]) MarshalJSON() ([]byte, error) {
	if !v.loaded {
		return json.Marshal(nil)
	}
	return json.Marshal(v.value)
}

func foreignKeyName(field reflect.StructField, tag string, tableType any) (string, error) {
	t, ok := field.Tag.Lookup(tag)
	if ok {
		return t, nil
	}

	pKeys := helpers.PrimaryKey(tableType)
	if len(pKeys) != 1 {
		return "", fmt.Errorf("you must specify keys for relationships with compound primary keys")
	}
	return database.GetTableSingular(tableType) + "_" + pKeys[0], nil
}

func primaryKeyName(field reflect.StructField, tag string, tableType any) (string, error) {
	t, ok := field.Tag.Lookup(tag)
	if ok {
		return t, nil
	}
	pKeys := helpers.PrimaryKey(tableType)
	if len(pKeys) != 1 {
		return "", fmt.Errorf("you must specify keys for relationships with compound primary keys")
	}
	return pKeys[0], nil
}

func getRelation(rv reflect.Value, relation string) (Relationship, bool) {
	if rv.Kind() == reflect.Pointer {
		if rv.IsZero() {
			rv = reflect.New(rv.Type().Elem())
			err := relationship.InitializeRelationships(rv.Interface())
			if err != nil {
				panic(err)
			}
		}
		rv = rv.Elem()
	}
	t := rv.Type()
	for i := 0; i < rv.NumField(); i++ {
		ft := t.Field(i)

		if ft.Anonymous {
			r, ok := getRelation(rv.Field(i), relation)
			if ok {
				return r, true
			}
			continue
		}

		if ft.Name != relation {
			continue
		}

		r, ok := rv.Field(i).Interface().(Relationship)
		if !ok {
			continue
		}
		return r, true
	}
	return nil, false
}
