package ioc

import (
	"reflect"
	"strings"
)

func (c *Container) AddSingleton[T any](fn any, foobars ...foobar) error {
	config, err := createSingletonProvider(fn, foobars...)
	if err != nil {
		return err
	}
	appendToSliceInMap(c.providers, reflect.TypeFor[T](), config)
	return nil
}

func (c *Container) AddNamedSingleton[T any](name string, fn any, foobars ...foobar) error {
	if strings.TrimSpace(name) == "" {
		return ErrNameCannotBeBlank
	}

	config, err := createSingletonProvider(fn, foobars...)
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

func createSingletonProvider(fn any, foobars ...foobar) (providerConfig, error) {
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

		args := mustGetArgs(c, fnV, foobars...)
		v := mustCallWithArgs(fnV, args)
		c.cache[fnV] = v
		return v
	}

	config := providerConfig{Create: create}

	return config, nil
}
