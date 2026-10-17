package ioc

import (
	"reflect"
)

// MustCall resolves fn's arguments from c, applies overrides, and invokes fn.
// It panics when fn is not a function or an argument cannot be resolved.
func MustCall(c *Container, fn any, overrides ...override) []reflect.Value {
	fnV := reflect.ValueOf(fn)
	if fnV.Type().Kind() != reflect.Func {
		panic(ErrNotAFunc{t: fnV.Type()})
	}
	args := mustGetArgs(c, fnV, overrides...)
	return fnV.Call(args)
}

// MustCallErr is like MustCall for a function whose first return value is an
// error. A returned error is passed through; Must applies only to invocation
// and dependency-resolution failures.
func MustCallErr(c *Container, fn any, overrides ...override) error {
	out := MustCall(c, fn, overrides...)
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
func MustCallValErr(c *Container, fn any, overrides ...override) (any, error) {
	out := MustCall(c, fn, overrides...)
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
