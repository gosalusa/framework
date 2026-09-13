package hooks

import (
	"context"
	"reflect"

	"gosalusa.com/database"
)

type BeforeSaver interface {
	BeforeSave(ctx context.Context, tx database.DB) error
}
type AfterSaver interface {
	AfterSave(ctx context.Context, tx database.DB) error
}

type AfterLoader interface {
	AfterLoad(ctx context.Context, tx database.DB) error
}

func BeforeSave(ctx context.Context, tx database.DB, model any) error {
	if model, ok := model.(BeforeSaver); ok {
		err := model.BeforeSave(ctx, tx)
		if err != nil {
			return err
		}
	}
	return eachField(reflect.ValueOf(model), func(i any) error {
		return BeforeSave(ctx, tx, i)
	})
}

func AfterSave(ctx context.Context, tx database.DB, model any) error {
	if model, ok := model.(AfterSaver); ok {
		err := model.AfterSave(ctx, tx)
		if err != nil {
			return err
		}
	}
	return eachField(reflect.ValueOf(model), func(i any) error {
		return AfterSave(ctx, tx, i)
	})
}

func AfterLoad(ctx context.Context, tx database.DB, model any) error {
	if model, ok := model.(AfterLoader); ok {
		err := model.AfterLoad(ctx, tx)
		if err != nil {
			return err
		}
	}

	return eachField(reflect.ValueOf(model), func(i any) error {
		return AfterLoad(ctx, tx, i)
	})
}

func eachField(v reflect.Value, callback func(model any) error) error {
	if v.Kind() == reflect.Pointer {
		return eachField(v.Elem(), callback)
	}
	if v.Kind() == reflect.Struct {
		t := v.Type()
		for i := 0; i < v.NumField(); i++ {
			if t.Field(i).Anonymous {
				f := v.Field(i)
				if f.Kind() != reflect.Pointer {
					f = f.Addr()
				}
				err := callback(f.Interface())
				if err != nil {
					return err
				}
			}
		}
		return nil
	}
	if v.Kind() == reflect.Slice {
		for i := 0; i < v.Len(); i++ {
			err := callback(v.Index(i).Interface())
			if err != nil {
				return err
			}
		}
		return nil
	}
	return nil
}
