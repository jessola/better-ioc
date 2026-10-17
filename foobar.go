package ioc

import "reflect"

type foobar struct {
	Index int
	Name  string
	Type  reflect.Type
}

func Foobar[T any](idx int, name string) foobar {
	return foobar{Type: reflect.TypeFor[T](), Index: idx, Name: name}
}
