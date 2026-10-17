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
	Kind  overrideKind
	Index int
	Name  string
	Type  reflect.Type
	Value reflect.Value
}

type overrideMap map[int]Override

func newOverrideMap(overrides []Override) overrideMap {
	return makeMapFromOverrides(overrides...)
}

func (o Override) Validate() error {
	return nil
}

func Provide[T any](idx int) Override {
	return Override{Kind: overrideBasic, Index: idx, Type: reflect.TypeFor[T]()}
}

func ProvideNamed[T any](idx int, name string) Override {
	return Override{Kind: overrideNamed, Index: idx, Type: reflect.TypeFor[T](), Name: name}
}

func ProvideValue[T any](idx int, value T) Override {
	return Override{
		Kind:  overrideValue,
		Index: idx,
		Type:  reflect.TypeFor[T](),
		Value: reflect.ValueOf(value),
	}
}

func makeMapFromOverrides(overrides ...Override) overrideMap {
	m := overrideMap{}
	for _, x := range overrides {
		m[x.Index] = x
	}
	return m
}
