package request

import (
	"context"
	"fmt"
	"net/http"
	"reflect"
	"strings"

	"gosalusa.com/internal/helpers"
	"gosalusa.com/request/rules"
)

type Validator interface {
	Valid() error
}

func Validate(request *http.Request, v any) error {
	ctx := context.Background()
	if request != nil {
		ctx = request.Context()
	}
	s, err := getStruct(reflect.ValueOf(v))
	if err != nil {
		return fmt.Errorf("Validate mast take a struct or pointer to a struct: %w", err)
	}

	vErr := ValidationError{}
	err = helpers.EachField(s, func(sf reflect.StructField, fv reflect.Value) error {
		name := getName(sf)
		newVErr, err := validateField(ctx, name, request, sf, fv)
		if err != nil {
			return err
		}
		if newVErr != nil {
			vErr.Merge(newVErr)
		}
		return nil
	})
	if err != nil {
		return err
	}

	if vErr.HasErrors() {
		return vErr
	}

	return nil
}

func validateField(ctx context.Context, attribute string, request *http.Request, ft reflect.StructField, fv reflect.Value) (ValidationError, error) {
	validate, ok := ft.Tag.Lookup("validate")
	if !ok {
		return nil, nil
	}

	vErr := ValidationError{}

	if validator, ok := fv.Interface().(Validator); ok {
		err := validator.Valid()
		if err != nil {
			vErr.AddError(attribute, err.Error())
		}
	}

	rulesStr := strings.Split(validate, "|")
	for _, ruleStr := range rulesStr {
		err := validateRule(ctx, attribute, request, ft, fv, ruleStr, vErr)
		if err != nil {
			return nil, err
		}
	}
	return vErr, nil
}

func validateRule(ctx context.Context, attribute string, request *http.Request, ft reflect.StructField, fv reflect.Value, ruleStr string, vErr ValidationError) error {
	ruleName, argsStr := split(ruleStr, ":")
	args := filterZeros(strings.Split(argsStr, ","))

	if ruleName == "required" {
		if !fv.IsZero() {
			return nil
		}
		msg, err := getMessage(ctx, ruleName, &MessageOptions{
			Attribute: attribute,
			Value:     fv.Interface(),
			Arguments: args,
			Field:     ft,
		})
		if err != nil {
			return err
		}
		vErr.AddError(attribute, msg)
		return nil
	}

	rule, ok := rules.GetRule(ruleName)
	if !ok {
		return nil
	}

	valid := rule(&rules.ValidationOptions{
		Value:     fv.Interface(),
		Arguments: args,
		Request:   request,
		Name:      attribute,
	})
	if !valid {
		msg, err := getMessage(ctx, ruleName, &MessageOptions{
			Attribute: attribute,
			Value:     fv.Interface(),
			Arguments: args,
			Field:     ft,
		})
		if err != nil {
			return err
		}
		vErr.AddError(attribute, msg)
	}

	return nil
}

func getName(f reflect.StructField) string {
	name, ok := f.Tag.Lookup("json")
	if ok {
		return name
	}

	name, ok = f.Tag.Lookup("query")
	if ok {
		return name
	}

	name, ok = f.Tag.Lookup("di")
	if ok {
		return name
	}

	return f.Name
}

func getStruct(v reflect.Value) (reflect.Value, error) {
	switch v.Kind() {
	case reflect.Struct:
		return v, nil
	case reflect.Interface, reflect.Pointer:
		return getStruct(v.Elem())
	default:
		return reflect.Value{}, fmt.Errorf("expected struct found %s", v.Kind())
	}
}

func split(s, sep string) (string, string) {
	parts := strings.SplitN(s, sep, 2)
	switch len(parts) {
	case 0:
		return "", ""
	case 1:
		return parts[0], ""
	default:
		return parts[0], parts[1]
	}
}

func filterZeros[T comparable](array []T) []T {
	newArray := []T{}
	var zero T
	for _, v := range array {
		if v != zero {
			newArray = append(newArray, v)
		}
	}
	return newArray
}
