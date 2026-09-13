package builder

import (
	"context"
	"encoding"
	"fmt"
	"log/slog"
	"reflect"

	"gosalusa.com/database"
	"gosalusa.com/database/model"
	"gosalusa.com/internal/helpers"
)

type iHasOneOrMany interface {
	getParentKey() string
	getRelatedKey() string
	parentKeyValue() (any, bool)
	relatedKeyValue() (any, bool)
}

type hasOneOrMany[T model.Model] struct {
	parent     any
	relatedKey string
	parentKey  string
}

var _ iHasOneOrMany = hasOneOrMany[model.Model]{}

// Subquery returns a Builder scoped to the relationship.
func (r hasOneOrMany[T]) Subquery() *Builder {
	return From[T]().
		WhereColumn(r.relatedKey, "=", database.GetTable(r.parent)+"."+r.parentKey).
		builder
}

// Query returns a ModelBuilder scoped to the relationship.
func (r hasOneOrMany[T]) Query() *ModelBuilder[T] {
	v, ok := helpers.GetValue(r.parent, r.parentKey)
	if !ok {
		panic(fmt.Errorf("no column %s in %v", r.parentKey, reflect.TypeOf(r.parent)))
	}
	return From[T]().Where(r.relatedKey, "=", v)
}

func (r hasOneOrMany[T]) parentKeyValue() (any, bool) {
	return helpers.GetValue(r.parent, r.parentKey)
}
func (r hasOneOrMany[T]) relatedKeyValue() (any, bool) {
	var related T
	return helpers.GetValue(related, r.relatedKey)
}

func (r hasOneOrMany[T]) getParentKey() string {
	return r.parentKey
}
func (r hasOneOrMany[T]) getRelatedKey() string {
	return r.relatedKey
}

func (r hasOneOrMany[T]) getRelated(ctx context.Context, tx database.DB, relations []Relationship) ([]T, error) {
	localKeys := make([]any, 0, len(relations))
	for _, r := range relations {
		local, ok := r.(iHasOneOrMany).parentKeyValue()
		if !ok {
			var related T
			return nil, fmt.Errorf("%s has no field %s: %w", reflect.TypeOf(related).Name(), r.(iHasOneOrMany).getParentKey(), ErrMissingField)
		}
		if local != nil {
			localKeys = append(localKeys, local)
		}
	}

	return From[T]().
		WhereIn(r.getRelatedKey(), localKeys).
		WithContext(ctx).
		Get(tx)
}

type relatedMap[T model.Model] map[any][]T

func newRelatedMap[T model.Model]() relatedMap[T] {
	return relatedMap[T]{}
}

func (rm relatedMap[T]) Get(k any) []T {
	k = stringify(k)
	if k == nil {
		return []T{}
	}
	v, ok := rm[k]
	if !ok {
		return []T{}
	}
	return v
}

func (rm relatedMap[T]) Single(k any, ok bool) T {
	if !ok {
		var zero T
		return zero
	}

	v := rm.Get(k)
	if len(v) == 0 {
		var zero T
		return zero
	}
	return v[0]
}
func (rm relatedMap[T]) Multi(k any, ok bool) []T {
	if !ok {
		return []T{}
	}

	return rm.Get(k)
}

func (rm relatedMap[T]) Add(k any, v T) {
	k = stringify(k)
	if k == nil {
		return
	}

	m, ok := rm[k]
	if !ok {
		m = []T{v}
	} else {
		m = append(m, v)
	}
	rm[k] = m
}

func (r hasOneOrMany[T]) relatedMap(ctx context.Context, tx database.DB, relations []Relationship) (relatedMap[T], error) {
	var related T
	if !helpers.HasField(related, r.getRelatedKey()) {
		return nil, fmt.Errorf("%s has no field %s: %w", reflect.TypeOf(related).Name(), r.getRelatedKey(), ErrMissingField)
	}

	relatedLists, err := r.getRelated(ctx, tx, relations)
	if err != nil {
		return nil, err
	}

	rm := newRelatedMap[T]()
	for _, related := range relatedLists {
		foreign, ok := helpers.GetValue(related, r.getRelatedKey())
		if !ok {
			return nil, fmt.Errorf("%s has no field %s: %w", reflect.TypeOf(related).Name(), r.getRelatedKey(), ErrMissingField)
		}
		rm.Add(foreign, related)
	}

	return rm, nil
}

func stringify(v any) any {
	rv := reflect.ValueOf(v)
	if rv.Kind() == reflect.Pointer && rv.IsNil() {
		return nil
	}
	if str, ok := v.(fmt.Stringer); ok {
		return str.String()
	}
	if txt, ok := v.(encoding.TextMarshaler); ok {
		b, err := txt.MarshalText()
		if err != nil {
			slog.Error("Failed to marshal key", "err", err)
		} else {
			return string(b)
		}
	}
	return v
}
