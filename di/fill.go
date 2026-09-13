package di

import (
	"context"
	"errors"
	"fmt"
	"reflect"
	"strings"

	"gosalusa.com/internal/helpers"
)

// var ErrNotFillable = errors.New("struct is not fillable")

// Fill populates the inject fields of v, a non-nil pointer to a struct, using
// the dependency provider carried by ctx. Fields without an inject tag are left
// untouched. Dependencies that are not registered are filled recursively when
// they are fillable structs; otherwise an error wrapping ErrNotRegistered is
// returned unless the inject tag is marked optional.
func Fill(ctx context.Context, v any) error {
	dp := GetDependencyProvider(ctx)
	return dp.fill(ctx, reflect.ValueOf(v), "")
}

// Fill populates the inject fields of v using this provider, rebinding ctx to
// dp so that recursively filled dependencies resolve against it.
func (dp *DependencyProvider) Fill(ctx context.Context, v any) error {
	if ctx.Value(dpKey) != dp {
		ctx = ContextWithDependencyProvider(ctx, dp)
	}
	return dp.fill(ctx, reflect.ValueOf(v), "")
}
func (dp *DependencyProvider) fill(ctx context.Context, v reflect.Value, tag string) error {
	if (v == reflect.Value{}) {
		return fmt.Errorf("di: Fill(interface): %w", ErrFillParameters)
	}
	if v.Kind() != reflect.Pointer {
		return fmt.Errorf("di: Fill(non-pointer "+v.Type().String()+"): %w", ErrFillParameters)
	}

	if v.IsNil() {
		return fmt.Errorf("di: Fill(nil)")
	}

	if ok, err := dp.resolve(ctx, tag, v); ok {
		return err
	}

	if !isFillable(v.Type()) {
		if isFillable(v.Elem().Type()) {
			v = v.Elem()
			v.Set(helpers.Create(v.Type()))
		} else {
			return errNotRegistered(v.Type())
		}
	}

	err := helpers.EachField(v, func(sf reflect.StructField, fv reflect.Value) error {
		if !sf.IsExported() {
			return nil
		}

		rawTag, ok := sf.Tag.Lookup("inject")
		if !ok {
			return nil
		}
		tag := parseTag(rawTag)

		nv := reflect.New(sf.Type)
		err := dp.fill(ctx, nv, tag.Name)
		if errors.Is(err, ErrNotRegistered) {
			if tag.Optional {
				return nil
			} else {
				return fmt.Errorf("unable to fill field %s.%s: %w", v.Type().String(), sf.Name, err)
			}
		} else if err != nil {
			return fmt.Errorf("failed to fill: %w", err)
		}
		fv.Set(nv.Elem())
		return nil
	})
	if err != nil {
		return err
	}

	return nil
}

// IsFillable reports whether v is a pointer to a struct that has at least one
// field with an inject tag.
func IsFillable(v any) bool {
	return isFillable(reflect.TypeOf(v))
}
func isFillable(t reflect.Type) bool {
	if t.Kind() != reflect.Pointer {
		return false
	}
	if t.Elem().Kind() != reflect.Struct {
		return false
	}
	for _, sf := range helpers.GetFields(t.Elem()) {
		if _, ok := sf.Tag.Lookup("inject"); ok {
			return true
		}
	}
	return false
}

func (dp *DependencyProvider) resolve(ctx context.Context, tag string, v reflect.Value) (bool, error) {
	f, ok := dp.factories.Get(v.Type().Elem())
	if !ok {
		return false, nil
	}

	res, err := f.Build(ctx, dp, tag)
	if err != nil {
		return true, err
	}
	v.Elem().Set(reflect.ValueOf(res))
	return true, nil
}

type fillTag struct {
	Name     string
	Optional bool
}

func parseTag(rawTag string) *fillTag {
	parts := strings.Split(rawTag, ",")
	tag := &fillTag{}
	tag.Name = parts[0]
	for _, p := range parts[1:] {
		if p == "optional" {
			tag.Optional = true
		}
	}
	return tag
}
