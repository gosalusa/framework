package builder

import (
	"context"
	"fmt"
	"iter"
	"reflect"
	"strings"

	"gosalusa.com/database"
	"gosalusa.com/internal/helpers"
	"gosalusa.com/internal/relationship"
)

// Load eagerly loads the given relationship on each of the models.
//
//	Load(db, users, "posts")
//
// Nested relationships can be loaded by separating them with dots.
//
//	Load(db, users, "posts.comments")
func Load(tx database.DB, models any, relation string) error {
	return LoadContext(context.Background(), tx, models, relation)
}

// LoadContext is Load with a context used when executing the relationship
// queries.
func LoadContext(ctx context.Context, tx database.DB, models any, relation string) error {
	return loadContext(ctx, tx, models, relation, false)
}

// LoadMissing loads the given relationship on each of the models, skipping any
// that have already been loaded.
func LoadMissing(tx database.DB, models any, relation string) error {
	return LoadMissingContext(context.Background(), tx, models, relation)
}

// LoadMissingContext is LoadMissing with a context used when executing the
// relationship queries.
func LoadMissingContext(ctx context.Context, tx database.DB, models any, relation string) error {
	return loadContext(ctx, tx, models, relation, true)
}
func loadContext(ctx context.Context, tx database.DB, models any, relationPath string, onlyMissing bool) error {
	err := relationship.InitializeRelationships(models)
	if err != nil {
		return err
	}
	relationNames := strings.Split(relationPath, ".")
	currentModels := models
	for i, relationName := range relationNames {
		relations := []Relationship{}
		err := helpers.Each(currentModels, func(v reflect.Value, pointer bool) error {
			r, ok := getRelation(v, relationName)
			if !ok {
				return fmt.Errorf("%s has no relation %s: %w", v.Type().Name(), relationName, ErrMissingRelationship)
			}

			if onlyMissing && r.Loaded() {
				return nil
			}
			relations = append(relations, r)
			return nil
		})
		if err != nil {
			return err
		}

		if len(relations) == 0 {
			return nil
		}

		err = relations[0].Load(ctx, tx, relations)
		if err != nil {
			return err
		}

		if i <= len(relationNames)-1 {
			values := []any{}
			err = helpers.Each(currentModels, func(v reflect.Value, pointer bool) error {
				related, ok := getValue(v, relationName)
				if !ok {
					return nil
				}
				out := related.MethodByName("Value").Call([]reflect.Value{})
				if len(out) != 2 {
					return fmt.Errorf("invalid value method")
				}
				if !out[1].IsZero() {
					values = append(values, out[0].Interface())
				}
				return nil
			})
			if err != nil {
				return err
			}
			currentModels = values
		}
	}

	return nil
}

func getValue(rv reflect.Value, key string) (reflect.Value, bool) {
	if rv.Kind() == reflect.Pointer {
		rv = rv.Elem()
	}
	if rv.Kind() != reflect.Struct {
		return reflect.Value{}, false
	}
	rt := rv.Type()
	for i := 0; i < rt.NumField(); i++ {
		ft := rt.Field(i)
		if ft.Anonymous {
			result, ok := getValue(rv.Field(i), key)
			if ok {
				return result, true
			}
			continue
		}
		if ft.Name == key {
			return rv.Field(i), true
		}
	}
	return reflect.Value{}, false
}
func ofType[T Relationship](relations []Relationship) iter.Seq[T] {
	return func(yield func(T) bool) {
		for _, r := range relations {
			rOfType, ok := r.(T)
			if !ok {
				continue
			}
			if !yield(rOfType) {
				return
			}
		}
	}
}
