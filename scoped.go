package ioc

import (
	"reflect"
	"strings"
)

// AddScoped registers fn as the provider for T. The provider is evaluated once
// per container, and its result is reused within that container.
func (c *Container) AddScoped[T any](fn any) error {
	config, err := createScopedProvider(fn)
	if err != nil {
		return err
	}
	appendToSliceInMap(c.providers, reflect.TypeFor[T](), config)
	return nil
}

// AddNamedScoped registers fn as a named scoped provider for T. It also
// registers the provider for ordinary type-based resolution of T.
func (c *Container) AddNamedScoped[T any](name string, fn any) error {
	if strings.TrimSpace(name) == "" {
		return ErrNameCannotBeBlank
	}

	config, err := createScopedProvider(fn)
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

func createScopedProvider(fn any) (providerConfig, error) {
	spec, err := EnsureFunc(fn)
	if err != nil {
		return providerConfig{}, err
	}
	fnV := reflect.ValueOf(spec.Func)

	config := providerConfig{
		Create: func(c *Container) any {
			// Check cache
			if v, ok := c.cache[fnV]; ok {
				return v
			}

			args := mustGetArgs(c, fnV, spec.Overrides...)
			v := mustCallWithArgs(fnV, args)
			// Set cache
			c.cache[fnV] = v
			c.trackValue(v)
			return v
		},
	}

	return config, nil
}
