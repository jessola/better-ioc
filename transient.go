package ioc

import (
	"reflect"
	"strings"
)

// AddTransient registers fn as the provider for T. The provider is evaluated
// each time T is resolved.
func (c *Container) AddTransient[T any](fn any, overrides ...Override) error {
	config, err := createTransientProvider(fn, overrides...)
	if err != nil {
		return err
	}
	appendToSliceInMap(c.providers, reflect.TypeFor[T](), config)
	return nil
}

// AddNamedTransient registers fn as a named transient provider for T. It also
// registers the provider for ordinary type-based resolution of T.
func (c *Container) AddNamedTransient[T any](name string, fn any, overrides ...Override) error {
	if strings.TrimSpace(name) == "" {
		return ErrNameCannotBeBlank
	}

	config, err := createTransientProvider(fn, overrides...)
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

func createTransientProvider(fn any, overrides ...Override) (providerConfig, error) {
	fnV := reflect.ValueOf(fn)
	if fnV.Type().Kind() != reflect.Func {
		return providerConfig{}, ErrNotAFunc{t: fnV.Type()}
	}

	config := providerConfig{
		Create: func(c *Container) any {
			args := mustGetArgs(c, fnV, overrides...)
			return mustCallWithArgs(fnV, args)
		},
	}

	return config, nil
}
