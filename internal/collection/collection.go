package collection

import "reflect"

type Collection struct {
	inner map[reflect.Type][]any
}

func New() *Collection {
	return &Collection{
		inner: make(map[reflect.Type][]any),
	}
}

func (c *Collection) Last[T any]() (v any, ok bool) {
	return
}
