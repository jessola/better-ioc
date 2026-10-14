package ioc

import (
	"fmt"
	"reflect"
)

type service struct {
	Create func(*Container) any
}

type ErrNotAFunc struct {
	t reflect.Type
}

type Container struct {
	parent    *Container
	providers map[reflect.Type][]service
	cache     map[any]any
}

func New() *Container {
	return &Container{
		providers: map[reflect.Type][]service{},
		cache:     make(map[any]any),
	}
}

func (err ErrNotAFunc) Error() string {
	return fmt.Sprintf("%s is not a func", err.t.String())
}

func (c *Container) CreateScope() *Container {
	child := New()
	child.parent = c
	return child
}

func (c *Container) Close() error {
	return nil
}

func (c *Container) Get[T any]() (v T, ok bool) {
	t := reflect.TypeFor[T]()

	provider, ok := c.chooseDefaultProvider(t)
	if !ok && t.Kind() == reflect.Slice {
		var providers []service
		elemType := t.Elem()

		providers = c.getAllProviders(elemType)
		if len(providers) == 0 {
			return
		}

		arr := reflect.MakeSlice(t, 0, len(providers))

		for _, provider := range providers {
			val := provider.Create(c)
			arr = reflect.Append(arr, reflect.ValueOf(val))
		}

		v, ok = arr.Interface().(T)
		return
	}

	// if !ok {
	// 	return
	// }

	v, ok = provider.Create(c).(T)
	return
}

func (c *Container) AddSingleton[T any](fn any) error {
	fnV := reflect.ValueOf(fn)
	if fnV.Type().Kind() != reflect.Func {
		return ErrNotAFunc{t: fnV.Type()}
	}

	appendToSliceInMap(c.providers, reflect.TypeFor[T](), service{
		Create: func(c *Container) any {
			// Check local cache
			if v, ok := c.cache[fnV]; ok {
				return v
			}

			// Ask parent to create
			if c.parent != nil {
				provider, ok := c.parent.chooseDefaultProvider(reflect.TypeFor[T]())
				if ok {
					v := provider.Create(c.parent)
					c.cache[fnV] = v
					return v
				}
			}

			v := mustCall(c, fnV)
			// Set local cache
			c.cache[fnV] = v
			return v
		},
	})

	return nil
}

func (c *Container) AddScoped[T any](fn any) error {
	fnV := reflect.ValueOf(fn)
	if fnV.Type().Kind() != reflect.Func {
		return ErrNotAFunc{t: fnV.Type()}
	}

	appendToSliceInMap(c.providers, reflect.TypeFor[T](), service{
		Create: func(c *Container) any {
			// Check cache
			if v, ok := c.cache[fnV]; ok {
				return v
			}

			v := mustCall(c, fnV)
			// Set cache
			c.cache[fnV] = v
			return v
		},
	})

	return nil
}

func (c *Container) AddTransient[T any](fn any) error {
	fnV := reflect.ValueOf(fn)
	if fnV.Type().Kind() != reflect.Func {
		return ErrNotAFunc{t: fnV.Type()}
	}

	appendToSliceInMap(c.providers, reflect.TypeFor[T](), service{
		Create: func(c *Container) any {
			return mustCall(c, fnV)
		},
	})

	return nil
}

func (c *Container) chooseDefaultProvider(t reflect.Type) (service, bool) {
	if len(c.providers[t]) == 0 {
		if c.parent == nil {
			return service{}, false
		}
		return c.parent.chooseDefaultProvider(t)
	} else {
		services := c.providers[t]
		return services[len(services)-1], true
	}
}

func (c *Container) getAllProviders(t reflect.Type) []service {
	var appendProviders func(*[]service, *Container)
	providers := []service{}

	appendProviders = func(providers *[]service, c *Container) {
		if c.parent != nil {
			appendProviders(providers, c.parent)
		}
		for _, p := range c.providers[t] {
			*providers = append(*providers, p)
		}
	}

	appendProviders(&providers, c)

	return providers
}

func mustCall(c *Container, fn reflect.Value) any {
	args := mustGetArgs(c, fn)
	out := fn.Call(args)
	return out[0].Interface()
}

func mustGetArgs(c *Container, fn reflect.Value) []reflect.Value {
	t := fn.Type()
	deps := make([]reflect.Value, t.NumIn())

	for i := range t.NumIn() {
		provider, ok := c.chooseDefaultProvider(t.In(i))
		if !ok {
			panic(fmt.Errorf("no provider for type %s", t.In(i).String()))
		}
		deps[i] = reflect.ValueOf(provider.Create(c))
	}

	return deps
}

func appendToSliceInMap[K comparable, V any](m map[K][]V, k K, v ...V) {
	if _, exists := m[k]; !exists {
		m[k] = []V{}
	}
	m[k] = append(m[k], v...)
}
