package ioc

import "reflect"

const (
	overrideBasic overrideKind = iota
	overrideNamed
	overrideValue
	overrideFactory
)

type overrideKind uint

type override struct {
	Kind  overrideKind
	Index int
	Name  string
	Type  reflect.Type
	Value reflect.Value
}

type overrideMap map[int]override

func newOverrideMap(overrides []override) overrideMap {
	return makeMapFromOverrides(overrides...)
}

func (o override) Validate() error {
	return nil
}

func Provide[T any](idx int) override {
	return override{Kind: overrideBasic, Index: idx, Type: reflect.TypeFor[T]()}
}

func ProvideNamed[T any](idx int, name string) override {
	return override{Kind: overrideNamed, Index: idx, Type: reflect.TypeFor[T](), Name: name}
}

func ProvideValue[T any](idx int, value T) override {
	return override{
		Kind:  overrideValue,
		Index: idx,
		Type:  reflect.TypeFor[T](),
		Value: reflect.ValueOf(value),
	}
}

func makeMapFromOverrides(overrides ...override) overrideMap {
	m := overrideMap{}
	for _, x := range overrides {
		m[x.Index] = x
	}
	return m
}
