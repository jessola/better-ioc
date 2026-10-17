package ioc

import (
	"reflect"
	"strings"
)

// AddScoped registers fn as the provider for T. The provider is evaluated once
// per container, and its result is reused within that container.
func (c *Container) AddScoped[T any](fn any, overrides ...Override) error {
	config, err := createScopedProvider(fn, overrides...)
	if err != nil {
		return err
	}
	appendToSliceInMap(c.providers, reflect.TypeFor[T](), config)
	return nil
}

// AddNamedScoped registers fn as a named scoped provider for T. It also
// registers the provider for ordinary type-based resolution of T.
func (c *Container) AddNamedScoped[T any](name string, fn any, overrides ...Override) error {
	if strings.TrimSpace(name) == "" {
		return ErrNameCannotBeBlank
	}

	config, err := createScopedProvider(fn, overrides...)
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

func createScopedProvider(fn any, overrides ...Override) (providerConfig, error) {
	fnV := reflect.ValueOf(fn)
	if fnV.Type().Kind() != reflect.Func {
		return providerConfig{}, ErrNotAFunc{t: fnV.Type()}
	}

	config := providerConfig{
		Create: func(c *Container) any {
			// Check cache
			if v, ok := c.cache[fnV]; ok {
				return v
			}

			args := mustGetArgs(c, fnV, overrides...)
			v := mustCallWithArgs(fnV, args)
			// Set cache
			c.cache[fnV] = v
			c.trackValue(v)
			return v
		},
	}

	return config, nil
}
