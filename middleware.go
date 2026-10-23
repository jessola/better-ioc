package ioc

import "reflect"

// MiddlewareFunc wraps next with compatible function behavior. It must return
// a function compatible with next.
type MiddlewareFunc func(next any) any

// WithMiddleware applies middlewares around fn, with the first middleware as
// the outermost wrapper. Each middleware is responsible for returning a
// function compatible with the next function it receives.
func WithMiddleware(fn any, middlewares ...MiddlewareFunc) (any, error) {
	// For now, middleware must be regular funcs, not FuncSpec
	// It's assumed that ioc.MustCallX will be used to call next, which will
	// allow fn itself to be FuncSpec. We don't control the shape of things
	// so we cannot use ioc.MustCallX within this function.

	if _, err := EnsureFunc(fn); err != nil {
		return nil, err
	}

	for _, m := range middlewares {
		if m == nil {
			return nil, ErrCannotBeNil
		}
		t := reflect.TypeOf(m)
		if t.Kind() != reflect.Func {
			return nil, NewErrNotAFunc(t)
		}
	}

	effectiveFn := fn

	for i := len(middlewares) - 1; i >= 0; i-- {
		effectiveFn = middlewares[i](effectiveFn)
	}

	return effectiveFn, nil
}
