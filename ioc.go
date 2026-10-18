// Package ioc provides an experimental reflection-based dependency injection
// container for Go. It resolves dependencies from constructor function parameters
// at runtime. Its public API may change without notice before a stable release.
package ioc

import (
	"errors"
	"io"
	"reflect"
	"sync"
)

type key struct {
	Name string
	Type reflect.Type
}

type providerConfig struct {
	Create func(*Container) any
}

// Container stores dependency providers and resolves their dependencies.
type Container struct {
	parent *Container

	providersMu sync.RWMutex
	providers   map[reflect.Type][]providerConfig

	cacheMu sync.Mutex
	cache   map[any]any

	nameMap map[key]providerConfig

	uniqueValsMu sync.Mutex
	uniqueValMap map[any]bool
	uniqueVals   []any
}

// NewContainer creates a root dependency injection container.
func NewContainer() *Container {
	c := &Container{
		providers:    map[reflect.Type][]providerConfig{},
		cache:        map[any]any{},
		nameMap:      map[key]providerConfig{},
		uniqueValMap: map[any]bool{},
	}

	c.AddScoped[*Container](Value(c))

	return c
}

// CreateScope creates a child container that inherits providers from c.
func (c *Container) CreateScope() *Container {
	child := NewContainer()
	child.parent = c
	return child
}

// Close closes values created by c that implement io.Closer in reverse creation
// order. It returns any errors reported while closing those values.
func (c *Container) Close() error {
	c.uniqueValsMu.Lock()
	values := c.uniqueVals
	c.uniqueVals = nil
	c.uniqueValMap = map[any]bool{}
	c.uniqueValsMu.Unlock()

	var errs []error
	for i := len(values) - 1; i >= 0; i-- {
		closer := values[i].(io.Closer)
		if err := closer.Close(); err != nil {
			errs = append(errs, err)
		}
	}

	return errors.Join(errs...)
}

// MustGet is like Get but panics with ErrNotFound when no matching provider is
// registered.
func (c *Container) MustGet[T any]() (v T) {
	v, ok := c.Get[T]()
	if !ok {
		panic(ErrNotFound{t: reflect.TypeFor[T]()})
	}
	return
}

// MustGetNamed is like GetNamed but panics with ErrNotFound when no matching
// provider is registered.
func (c *Container) MustGetNamed[T any](name string) (v T) {
	v, ok := c.GetNamed[T](name)
	if !ok {
		panic(ErrNotFound{t: reflect.TypeFor[T]()})
	}
	return
}

// Get resolves the most recently registered provider for T. When T is a slice
// with no directly registered provider, Get resolves all providers registered
// for the slice's element type. A provider registered directly for T takes
// precedence over element-type providers. Get returns false when no matching
// provider is registered.
func (c *Container) Get[T any]() (v T, ok bool) {
	t := reflect.TypeFor[T]()

	provider, ok := c.getDefaultProvider(t)
	if !ok && t.Kind() == reflect.Slice {
		var providers []providerConfig
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
	if !ok {
		return
	}

	v, ok = provider.Create(c).(T)
	return
}

// GetNamed resolves the provider registered for T with name.
// It returns false when no matching provider is registered.
func (c *Container) GetNamed[T any](name string) (v T, ok bool) {
	t := reflect.TypeFor[T]()
	provider, ok := c.getNamedProvider(t, name)
	if !ok {
		return
	}
	v, ok = provider.Create(c).(T)
	return
}

func (c *Container) getDefaultProvider(t reflect.Type) (providerConfig, bool) {
	c.providersMu.RLock()
	defer c.providersMu.RUnlock()

	if len(c.providers[t]) == 0 {
		if c.parent == nil {
			return providerConfig{}, false
		}
		return c.parent.getDefaultProvider(t)
	} else {
		services := c.providers[t]
		return services[len(services)-1], true
	}
}

func (c *Container) getNamedProvider(t reflect.Type, name string) (providerConfig, bool) {
	provider, ok := c.nameMap[key{Type: t, Name: name}]
	if ok {
		return provider, true
	}

	if c.parent == nil {
		return providerConfig{}, false
	}

	return c.parent.getNamedProvider(t, name)
}

func (c *Container) getAllProviders(t reflect.Type) []providerConfig {
	c.providersMu.RLock()
	defer c.providersMu.RUnlock()

	var appendProviders func(*[]providerConfig, *Container)
	providers := []providerConfig{}

	appendProviders = func(providers *[]providerConfig, c *Container) {
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

func (c *Container) trackValue(v any) {
	c.uniqueValsMu.Lock()
	defer c.uniqueValsMu.Unlock()

	if _, ok := v.(io.Closer); !ok {
		return
	}
	if !reflect.TypeOf(v).Comparable() {
		return
	}
	if _, ok := c.uniqueValMap[v]; ok {
		return
	}

	c.uniqueValMap[v] = true
	c.uniqueVals = append(c.uniqueVals, v)
}

func mustCall(c *Container, fn reflect.Value, overrides ...Override) any {
	args := mustGetArgs(c, fn, overrides...)
	out := fn.Call(args)
	return out[0].Interface()
}

func mustCallWithArgs(fn reflect.Value, args []reflect.Value) any {
	out := fn.Call(args)
	return out[0].Interface()
}

func callWithArgs(fn reflect.Value, args []reflect.Value) (any, error) {
	out := fn.Call(args)
	return out[0].Interface(), nil
}

func getArgs(c *Container, fn reflect.Value, overrides ...Override) ([]reflect.Value, error) {
	t := fn.Type()
	deps := make([]reflect.Value, t.NumIn())
	overrideMap := newOverrideMap(overrides)

	for i := range t.NumIn() {
		if ovr, ok := overrideMap[i]; ok {
			// To validate or not to validate, that is the question
			// if err := x.Validate(); err != nil {
			//   panic(err)
			// }

			// TODO: finish logic
			// TODO: Consider checking the types match, though function panics anyway
			switch ovr.Kind {
			case overrideNamed:
				provider, ok := c.getNamedProvider(ovr.Type, ovr.Name)
				if !ok {
					return []reflect.Value{}, ErrNotFound{t: t.In(i)}
				}
				deps[i] = reflect.ValueOf(provider.Create(c))
			case overrideValue:
				deps[i] = ovr.Value
			case overrideFactory:
				deps[i] = reflect.ValueOf(ovr.Factory(c))
			default: // overrideBasic
				provider, ok := c.getDefaultProvider(ovr.Type)
				if !ok {
					return []reflect.Value{}, ErrNotFound{t: t.In(i)}
				}
				deps[i] = reflect.ValueOf(provider.Create(c))
			}
		} else {
			provider, ok := c.getDefaultProvider(t.In(i))
			if !ok {
				return []reflect.Value{}, ErrNotFound{t: t.In(i)}
			}
			deps[i] = reflect.ValueOf(provider.Create(c))
		}
	}

	return deps, nil
}

func mustGetArgs(c *Container, fn reflect.Value, overrides ...Override) []reflect.Value {
	args, err := getArgs(c, fn, overrides...)
	if err != nil {
		panic(err)
	}
	return args
}

func appendToSliceInMap[K comparable, V any](m map[K][]V, k K, v ...V) {
	if _, exists := m[k]; !exists {
		m[k] = []V{}
	}
	m[k] = append(m[k], v...)
}
