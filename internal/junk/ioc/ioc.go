package ioc

import (
	"fmt"
	"reflect"
)

const (
	Singleton Scope = iota
	Transient
	Scoped
)

type Scope int

type providerFunc func(*Container) any

type service struct {
	Scope   Scope
	Provide providerFunc
}

type Container struct {
	parent   *Container
	services map[reflect.Type][]service
	cache    map[reflect.Value]any
	nameMap  map[string]service
}

func New() *Container {
	return &Container{
		services: make(map[reflect.Type][]service),
		cache:    make(map[reflect.Value]any),
		nameMap:  map[string]service{},
	}
}

func (c *Container) CreateScope() *Container {
	s := New()
	s.parent = c
	return s
}

func (c *Container) add(t reflect.Type, scope Scope, fn any) error {
	// Preload parent's services
	// Important to PRE-load not just load, so newer services take priority
	if c.services[t] == nil && c.parent != nil && c.parent.services[t] != nil {
		for _, svc := range c.parent.services[t] {
			appendToSliceInMap(c.services, t, svc)
		}
	}

	fnType := reflect.TypeOf(fn)
	fnVal := reflect.ValueOf(fn)

	appendToSliceInMap(c.services, t, service{
		Scope: scope,
		Provide: func(c *Container) any {
			deps := make([]reflect.Value, fnType.NumIn())
			for i := range fnType.NumIn() {
				svc, ok := c.get(fnType.In(i))
				if !ok {
					panic(fmt.Errorf("no service with type %s", fnType.In(i).String()))
				}
				deps = append(deps, reflect.ValueOf(svc))
			}
			return fnVal.Call(deps)[0].Interface()
		},
	})
	return nil
}

func (c *Container) get(t reflect.Type) (v any, ok bool) {
	// Preload services
	if c.services[t] == nil && c.parent != nil && c.parent.services[t] != nil {
		c.services[t] = c.parent.services[t]
	}

	// Dynamically provide a slice of each registered provider for a type
	// unless the slice of a type is already directly registered
	if t.Kind() == reflect.Slice {
		if _, exists := c.services[t]; !exists {
			vals := []any{}
			elemType := t.Elem()

			services, ok := c.services[elemType]
			if !ok || len(services) == 0 {
				return nil, false
			}

			for _, service := range services {

				switch service.Scope {
				case Singleton:
					providerVal := reflect.ValueOf(service.Provide)
					// Check local cache
					if val, exists := c.cache[providerVal]; exists {
						vals = append(vals, val)
						continue
					}

					if c.parent != nil {
						// Check parent cache
						if val, exists := c.parent.cache[providerVal]; exists {
							vals = append(vals, val)
							// Set local cache
							c.cache[providerVal] = val
							continue
						}

						val := service.Provide(c.parent)
						vals = append(vals, val)
						// Set parent cache
						c.parent.cache[providerVal] = val
						// Set local cache
						c.cache[providerVal] = val
					} else {
						val := service.Provide(c)
						vals = append(vals, val)
						// Set local cache
						c.cache[providerVal] = val
					}
				case Scoped:
					providerVal := reflect.ValueOf(service.Provide)
					// Check local cache
					if val, exists := c.cache[providerVal]; exists {
						vals = append(vals, val)
						continue
					}

					val := service.Provide(c)
					vals = append(vals, val)

					c.cache[providerVal] = val
				case Transient:
					val := service.Provide(c)
					vals = append(vals, val)
				default:
					//
				}

			}
			// v = vals
			return vals, true
		}
	}
	// END OF SLICE LOGIC

	services, ok := c.services[t]
	if !ok || len(services) == 0 {
		return
	}
	service := services[len(services)-1]

	switch service.Scope {
	case Singleton:
		providerVal := reflect.ValueOf(service.Provide)
		// Check local cache
		if val, exists := c.cache[providerVal]; exists {
			v = val
			return
		}

		if c.parent != nil {
			// Check parent cache
			if val, exists := c.parent.cache[providerVal]; exists {
				v = val
				// Set local cache
				c.cache[providerVal] = v
				return
			}

			v = service.Provide(c.parent)
			// Set parent cache
			c.parent.cache[providerVal] = v
		} else {
			v = service.Provide(c)
		}
		// Set local cache
		c.cache[providerVal] = v
	case Scoped:
		providerVal := reflect.ValueOf(service.Provide)
		// Check local cache
		if val, exists := c.cache[providerVal]; exists {
			v = val
			return
		}

		// if c.parent != nil {
		// 	v = service.Provide(c.parent)
		// } else {
		v = service.Provide(c)
		// }
		// Set local cache
		c.cache[providerVal] = v
	case Transient:
		v = service.Provide(c)
	default:
		//
	}

	return
}
