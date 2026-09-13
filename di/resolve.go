package di

import (
	"context"
)

// Resolve returns a value of type T built by the dependency provider carried
// by ctx. If T is a fillable struct, its inject fields are filled recursively;
// otherwise a registered factory is required.
func Resolve[T any](ctx context.Context) (T, error) {
	dp := GetDependencyProvider(ctx)
	var result T
	err := dp.Fill(ctx, &result)
	if err != nil {
		var zero T
		return zero, err
	}
	return result, err
}
