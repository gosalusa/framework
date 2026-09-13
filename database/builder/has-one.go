package builder

import (
	"context"
	"reflect"

	"gosalusa.com/database"
	"gosalusa.com/database/model"
)

// # Tags:
//   - local: parent model
//   - foreign: related model
type HasOne[T model.Model] struct {
	hasOneOrMany[T]
	relationValue[T]
}

var _ Relationship = &HasOne[model.Model]{}

// Initialize configures the relationship from the parent model and the struct
// field that holds it.
func (r *HasOne[T]) Initialize(parent any, field reflect.StructField) error {
	r.parent = parent
	parentKey, err := primaryKeyName(field, "local", parent)
	if err != nil {
		return err
	}
	relatedKey, err := foreignKeyName(field, "foreign", parent)
	if err != nil {
		return err
	}
	r.parentKey = parentKey
	r.relatedKey = relatedKey
	return nil
}

// Load fills the related value on each of the relationships in relations.
func (r *HasOne[T]) Load(ctx context.Context, tx database.DB, relations []Relationship) error {
	rm, err := r.relatedMap(ctx, tx, relations)
	if err != nil {
		return err
	}

	for relation := range ofType[*HasOne[T]](relations) {
		relation.value = rm.Single(relation.parentKeyValue())
		relation.loaded = true
	}
	return nil
}

// ForeignKeys returns a list of related tables and what columns they are
// related on.
func (r *HasOne[T]) ForeignKeys() []*ForeignKey {
	return []*ForeignKey{}
}
