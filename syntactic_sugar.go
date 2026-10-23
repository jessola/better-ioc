package ioc

// Must returns value or panics if err is non-nil.
func Must[T any](value T, err error) T {
	if err != nil {
		panic(err)
	}
	return value
}

// Value returns a provider function that always returns v.
func Value[T any](v T) func() T {
	return func() T {
		return v
	}
}

// Resolve returns a provider that returns a resolved value of T. It is useful
// for registering a concrete type under another type, such as an interface.
func Resolve[T any]() func(v T) T {
	return func(v T) T {
		return v
	}
}

// ResolveNamed returns a provider that resolves the provider registered for T
// with name. It panics with ErrNotFound when no matching provider is registered.
func ResolveNamed[T any](name string) func(*Container) T {
	return func(c *Container) T {
		return c.MustGetNamed[T](name)
	}
}

// BindSingleton registers V as a singleton provider for T. V must be
// registered when T is resolved.
func (c *Container) BindSingleton[T any, V any]() error {
	return c.AddSingleton[T](Resolve[V]())
}

// BindNamedSingleton registers V as a named singleton provider for T. It also
// registers the provider for ordinary type-based resolution of T. V must be
// registered when T is resolved.
func (c *Container) BindNamedSingleton[T any, V any](name string) error {
	return c.AddNamedSingleton[T](name, Resolve[V]())
}

// BindScoped registers V as a scoped provider for T. V must be registered
// when T is resolved.
func (c *Container) BindScoped[T any, V any]() error {
	return c.AddScoped[T](Resolve[V]())
}

// BindNamedScoped registers V as a named scoped provider for T. It also
// registers the provider for ordinary type-based resolution of T. V must be
// registered when T is resolved.
func (c *Container) BindNamedScoped[T any, V any](name string) error {
	return c.AddNamedScoped[T](name, Resolve[V]())
}

// BindTransient registers V as a transient provider for T. V must be
// registered when T is resolved.
func (c *Container) BindTransient[T any, V any]() error {
	return c.AddTransient[T](Resolve[V]())
}

// BindNamedTransient registers V as a named transient provider for T. It also
// registers the provider for ordinary type-based resolution of T. V must be
// registered when T is resolved.
func (c *Container) BindNamedTransient[T any, V any](name string) error {
	return c.AddNamedTransient[T](name, Resolve[V]())
}
