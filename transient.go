package ioc

import (
	"reflect"
	"strings"
)

func (c *Container) AddTransient[T any](fn any, foobars ...foobar) error {
	config, err := createTransientProvider(fn, foobars...)
	if err != nil {
		return err
	}
	appendToSliceInMap(c.providers, reflect.TypeFor[T](), config)
	return nil
}

func (c *Container) AddNamedTransient[T any](name string, fn any, foobars ...foobar) error {
	if strings.TrimSpace(name) == "" {
		return ErrNameCannotBeBlank
	}

	config, err := createTransientProvider(fn, foobars...)
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

func createTransientProvider(fn any, foobars ...foobar) (providerConfig, error) {
	fnV := reflect.ValueOf(fn)
	if fnV.Type().Kind() != reflect.Func {
		return providerConfig{}, ErrNotAFunc{t: fnV.Type()}
	}

	config := providerConfig{
		Create: func(c *Container) any {
			args := mustGetArgs(c, fnV, foobars...)
			return mustCallWithArgs(fnV, args)
		},
	}

	return config, nil
}
