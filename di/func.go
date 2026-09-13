package di

import (
	"context"
	"fmt"
	"reflect"
)

var (
	errorType = reflect.TypeFor[error]()
)

// PrepareFuncCtx returns a new function of type T whose extra parameters are
// filled from the dependency provider carried by ctx at call time. It is a
// shorthand for PrepareFunc with a nil provider.
func PrepareFuncCtx[T any](fn any) T {
	return PrepareFunc[T](nil, fn)
}

// PrepareFunc returns a new function of type T built by wrapping fn. fn may
// accept more parameters than T; any parameters after those in T are filled
// from dp (or the provider carried by the context passed to the returned
// function when dp is nil). If a fill error occurs and T does not return an
// error, the call panics.
func PrepareFunc[T any](dp *DependencyProvider, fn any) T {
	t := reflect.TypeFor[T]()
	vFn := reflect.ValueOf(fn)
	tFn := reflect.TypeOf(fn)
	if t.Kind() != reflect.Func {
		panic(fmt.Errorf("di.Func(): non function type parameter"))
	}
	if vFn.Kind() != reflect.Func {
		panic(fmt.Errorf("di.Func(): non function argument"))
	}

	if t.NumOut() != tFn.NumOut() {
		panic(fmt.Errorf("di.Func(): return counts do not match"))
	}

	outError := -1
	for i := range t.NumOut() {
		if t.Out(i) != tFn.Out(i) {
			panic(fmt.Errorf("di.Func(): return types do not match"))
		}

		if t.Out(i) == errorType {
			outError = i
		}
	}

	if t.NumIn() > tFn.NumIn() {
		panic(fmt.Errorf("di.Func(): more parameters on result function"))
	}

	inContext := -1
	for i := range t.NumIn() {
		if t.In(i) != tFn.In(i) {
			panic(fmt.Errorf("di.Func(): parameters do not match"))
		}
		if t.In(i) == contextType {
			inContext = i
		}
	}

	return reflect.MakeFunc(t, func(args []reflect.Value) (results []reflect.Value) {
		ctx := args[inContext].Interface().(context.Context)
		if dp == nil {
			dp = GetDependencyProvider(ctx)
		}
		fullArgs := make([]reflect.Value, tFn.NumIn())
		copy(fullArgs, args)
		for i := t.NumIn(); i < tFn.NumIn(); i++ {
			inV := reflect.New(tFn.In(i))
			err := dp.fill(ctx, inV, "")
			if err != nil {
				if outError == -1 {
					panic(err)
				} else {
					results := make([]reflect.Value, t.NumOut())
					for i := range t.NumOut() {
						if i == outError {
							results[i] = reflect.ValueOf(err)
						} else {
							results[i] = reflect.Zero(t.Out(i))
						}
					}
					return results
				}
			}
			fullArgs[i] = inV.Elem()
		}
		return vFn.Call(fullArgs)
	}).Interface().(T)
}
