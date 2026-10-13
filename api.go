package ioc

import (
	"fmt"
	"reflect"
)

func (c *Container) AddTransient[T any](fn any) error {
	return nil
}

func (c *Container) Get[T any]() (v any, ok bool) {
	t := reflect.TypeFor[T]()

	if err := c.loadServices(t); err != nil {
		ok = false
		return
	}

	// Choose service
	svc := c.services[t][len(c.services[t])-1]

	// Resolve
	switch svc.Scope {
	default:
	}

	return
}

func (c *Container) loadServices(t reflect.Type) error {
	if len(c.services[t]) > 0 {
		return nil
	}
	if c.parent == nil {
		return fmt.Errorf("no service with type %s", t.String())
	}
	c.services[t] = c.parent.services[t]
	return nil
}

func (c *Container) provide(s service) (v any, ok bool) {
	switch s.Scope {
	case Singleton:
		//
	case Scoped:
		//
	case Transient:
		v = s.Provide(c)
	default:
	}
	return
}
