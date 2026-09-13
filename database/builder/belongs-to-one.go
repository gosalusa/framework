package builder

import (
	"context"
	"reflect"

	"gosalusa.com/database"
	"gosalusa.com/database/model"
)

// BelongsTo represents a belongs to relationship on a model. The parent model
// with a BelongsTo property will have a column referencing another tables
// primary key. For example if model Foo had a BelongsTo[*Bar] property the foos
// table would have a foos.bar_id column related to the bars.id column. Struct
// tags can be used to change the column names if they don't follow the default
// naming convention. The column on the parent model can be set with a foreign
// tag and the column on the related model can be set with an owner tag.
//
// # Tags:
//   - owner: parent model
//   - foreign: related model
type BelongsTo[T model.Model] struct {
	hasOneOrMany[T]
	relationValue[T]
}

var _ Relationship = &BelongsTo[model.Model]{}

// Initialize configures the relationship from the parent model and the struct
// field that holds it.
func (r *BelongsTo[T]) Initialize(parent any, field reflect.StructField) error {
	var related T
	r.parent = parent
	parentKey, err := foreignKeyName(field, "foreign", related)
	if err != nil {
		return err
	}
	relatedKey, err := primaryKeyName(field, "owner", related)
	if err != nil {
		return err
	}

	r.parentKey = parentKey
	r.relatedKey = relatedKey

	return nil
}

// Load fills the related value on each of the relationships in relations.
func (r *BelongsTo[T]) Load(ctx context.Context, tx database.DB, relations []Relationship) error {
	rm, err := r.relatedMap(ctx, tx, relations)
	if err != nil {
		return err
	}

	for relation := range ofType[*BelongsTo[T]](relations) {
		relation.value = rm.Single(relation.parentKeyValue())
		relation.loaded = true
	}
	return nil
}

// ForeignKeys returns a list of related tables and what columns they are
// related on.
func (r *BelongsTo[T]) ForeignKeys() []*ForeignKey {
	var related T
	return []*ForeignKey{{
		LocalKey:     r.getParentKey(),
		RelatedTable: database.GetTable(related),
		RelatedKey:   r.getRelatedKey(),
	}}
}
