package ioc

import "reflect"

const (
	overrideBasic overrideKind = iota
	overrideNamed
	overrideValue
	overrideFactory
)

type overrideKind uint

type Override struct {
	Kind    overrideKind
	Index   int
	Name    string
	Type    reflect.Type
	Value   reflect.Value
	Factory func(*Container) any
}

type overrideMap map[int]Override

type FuncSpec struct {
	Func      any
	Overrides []Override
}

func newOverrideMap(overrides []Override) overrideMap {
	return makeMapFromOverrides(overrides...)
}

func (o Override) Validate() error {
	return nil
}

// func WithOverrides(fn any, overrides ...Override) func(*Container) any {
// 	fnV := reflect.ValueOf(fn)
// 	if fnV.Type().Kind() != reflect.Func {
// 		panic(ErrNotAFunc{t: fnV.Type()})
// 	}

// 	return func(c *Container) any {
// 		args := mustGetArgs(c, fnV, overrides...)
// 		return mustCallWithArgs(fnV, args)
// 	}
// }

func EnsureFunc(fn any) (FuncSpec, error) {
	if fn == nil {
		return FuncSpec{}, ErrCannotBeNil
	}

	if fn, ok := fn.(FuncSpec); ok {
		if fn.Func == nil {
			return FuncSpec{}, ErrCannotBeNil
		}

		fnV := reflect.ValueOf(fn.Func)
		if fnV.Type().Kind() == reflect.Func {
			return fn, nil
		} else {
			return FuncSpec{}, NewErrNotAFunc(fnV.Type())
		}
	}

	fnT := reflect.TypeOf(fn)
	if fnT.Kind() != reflect.Func {
		return FuncSpec{}, NewErrNotAFunc(fnT)
	}

	return FuncSpec{Func: fn}, nil
}

func WithOverrides(fn any, overrides ...Override) FuncSpec {
	fnV := reflect.ValueOf(fn)
	if fnV.Type().Kind() != reflect.Func {
		panic(ErrNotAFunc{t: fnV.Type()})
	}

	return FuncSpec{
		Func:      fn,
		Overrides: overrides,
	}
}

// Provide overrides the dependency at idx with the default registration for T.
func Provide[T any](idx int) Override {
	return Override{Kind: overrideBasic, Index: idx, Type: reflect.TypeFor[T]()}
}

// ProvideNamed overrides the dependency at idx with the registration for T
// named name.
func ProvideNamed[T any](idx int, name string) Override {
	return Override{Kind: overrideNamed, Index: idx, Type: reflect.TypeFor[T](), Name: name}
}

// ProvideValue overrides the dependency at idx with value.
func ProvideValue[T any](idx int, value T) Override {
	return Override{
		Kind:  overrideValue,
		Index: idx,
		Type:  reflect.TypeFor[T](),
		Value: reflect.ValueOf(value),
	}
}

// ProvideFactory overrides the dependency at idx with the value returned by
// fn. The factory receives the container resolving the dependency, allowing it
// to resolve values from the current scope.
func ProvideFactory(idx int, fn func(*Container) any) Override {
	return Override{
		Kind:    overrideFactory,
		Index:   idx,
		Factory: fn,
	}
}

func makeMapFromOverrides(overrides ...Override) overrideMap {
	m := overrideMap{}
	for _, x := range overrides {
		m[x.Index] = x
	}
	return m
}
