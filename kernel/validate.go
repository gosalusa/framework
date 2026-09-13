package kernel

import (
	"context"
	"reflect"

	"gosalusa.com/validate"
)

var _ validate.Validator = (*Kernel)(nil)

func (k *Kernel) Validate(ctx context.Context) error {
	var err error

	err = validate.Append(ctx, err, k.dependencyProvider)

	for _, s := range k.services {
		if v, ok := s.(validate.Validator); ok {
			err = validate.Append(ctx, err, v)
		}

		err = validate.Append(ctx, err, k.dependencyProvider.Validator(reflect.TypeOf(s)))
	}

	h := k.rootHandler
	if v, ok := h.(validate.Validator); ok {
		err = validate.Append(ctx, err, v)
	}

	return err
}
