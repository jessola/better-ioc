package ioc

import (
	"reflect"
	"strings"
)

// AddSingleton registers fn as the provider for T. The provider is evaluated at
// most once for the container hierarchy and its result is reused thereafter.
func (c *Container) AddSingleton[T any](fn any, overrides ...override) error {
	config, err := createSingletonProvider(fn, overrides...)
	if err != nil {
		return err
	}
	appendToSliceInMap(c.providers, reflect.TypeFor[T](), config)
	return nil
}

// AddNamedSingleton registers fn as a named singleton provider for T. It also
// registers the provider for ordinary type-based resolution of T.
func (c *Container) AddNamedSingleton[T any](name string, fn any, overrides ...override) error {
	if strings.TrimSpace(name) == "" {
		return ErrNameCannotBeBlank
	}

	config, err := createSingletonProvider(fn, overrides...)
	if err != nil {
		return err
	}

	t := reflect.TypeFor[T]()

	c.providersMu.Lock()
	c.nameMap[key{Name: name, Type: t}] = config
	appendToSliceInMap(c.providers, t, config)
	c.providersMu.Unlock()

	return nil
}

func createSingletonProvider(fn any, overrides ...override) (providerConfig, error) {
	fnV := reflect.ValueOf(fn)
	if fnV.Type().Kind() != reflect.Func {
		return providerConfig{}, ErrNotAFunc{t: fnV.Type()}
	}

	var create func(*Container) any

	create = func(c *Container) any {
		// Check local cache
		if v, ok := c.cache[fnV]; ok {
			return v
		}
		// Check ancestors
		if c.parent != nil {
			v := create(c.parent)
			c.cache[fnV] = v
			return v
		}

		args := mustGetArgs(c, fnV, overrides...)
		v := mustCallWithArgs(fnV, args)
		c.cache[fnV] = v
		return v
	}

	config := providerConfig{Create: create}

	return config, nil
}
