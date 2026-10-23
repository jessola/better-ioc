package ioc

import (
	"reflect"
)

// MustCall resolves fn's arguments from c, applies any overrides attached with
// WithOverrides, and invokes fn. It panics when fn is not a function or an
// argument cannot be resolved.
func MustCall(c *Container, fn any) []reflect.Value {
	spec, err := EnsureFunc(fn)
	if err != nil {
		panic(err)
	}
	fnV := reflect.ValueOf(spec.Func)
	args := mustGetArgs(c, fnV, spec.Overrides...)
	return fnV.Call(args)
}

// MustCallErr is like MustCall for a function whose first return value is an
// error. A returned error is passed through; Must applies only to invocation
// and dependency-resolution failures.
func MustCallErr(c *Container, fn any) error {
	out := MustCall(c, fn)
	if len(out) == 0 {
		// TODO: Maybe leave this to reflect
		panic("not enough return values")
	}
	// Important, otherwise returning nil will panic at runtime.
	if err, ok := out[0].Interface().(error); ok {
		return err
	}
	return nil
}

// MustCallValErr is like MustCall for a function returning a value and an
// error. A returned error is passed through; Must applies only to invocation
// and dependency-resolution failures.
func MustCallValErr(c *Container, fn any) (any, error) {
	out := MustCall(c, fn)
	if len(out) < 2 {
		// TODO: Maybe leave this to reflect
		panic("not enough return values")
	}
	// Important, otherwise returning a nil error will panic at runtime.
	if err, ok := out[1].Interface().(error); ok {
		return nil, err
	}
	return out[0].Interface(), nil
}

// MustMakeScopedFunc is like MakeScopedFunc but panics when the wrapper
// cannot be created.
func MustMakeScopedFunc[T any](c *Container, fn any) T {
	v, err := MakeScopedFunc[T](c, fn)
	if err != nil {
		panic(err)
	}
	return v
}

// MustMakeFuncWithScope is like MakeFuncWithScope but panics when the wrapper
// cannot be created.
func MustMakeFuncWithScope[T any](s *Container, fn any) T {
	v, err := MakeFuncWithScope[T](s, fn)
	if err != nil {
		panic(err)
	}
	return v
}

// MakeFuncWithScope returns a function of type T that resolves dependencies
// from s for each invocation. Its invocation arguments are registered as
// scoped values in s. The caller owns s and is responsible for closing it.
// T must be a function type compatible with the generated wrapper.
func MakeFuncWithScope[T any](s *Container, fn any) (v T, err error) {
	spec, err := EnsureFunc(fn)
	if err != nil {
		return
	}
	fnV := reflect.ValueOf(spec.Func)

	wrapped := reflect.MakeFunc(reflect.TypeFor[T](), func(args []reflect.Value) (results []reflect.Value) {
		for _, arg := range args {
			config, err := createScopedProvider(Value(arg.Interface()))
			if err != nil {
				panic(err)
			}
			appendToSliceInMap(s.providers, arg.Type(), config)
		}

		deps := mustGetArgs(s, fnV, spec.Overrides...)
		return fnV.Call(deps)
	})

	var ok bool
	if v, ok = wrapped.Interface().(T); !ok {
		err = ErrCouldNotCast{}
	}

	return
}

// MakeScopedFunc returns a function of type T that creates a child scope for
// each invocation. Its invocation arguments are registered as scoped values,
// then fn is called with dependencies resolved from that scope. T must be a
// function type compatible with the generated wrapper.
func MakeScopedFunc[T any](c *Container, fn any) (v T, err error) {
	spec, err := EnsureFunc(fn)
	if err != nil {
		return
	}

	wrapped := reflect.MakeFunc(reflect.TypeFor[T](), func(args []reflect.Value) (results []reflect.Value) {
		s := c.CreateScope()
		defer s.Close()

		fn := MustMakeFuncWithScope[T](s, spec)
		return reflect.ValueOf(fn).Call(args)
	})

	var ok bool
	if v, ok = wrapped.Interface().(T); !ok {
		err = ErrCouldNotCast{}
		return
	}

	return
}
