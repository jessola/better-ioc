package ioc

// DefaultContainer is the package-level container for application-wide
// registrations.
var DefaultContainer = NewContainer()

// CreateScope creates a child scope of DefaultContainer.
func CreateScope() *Container {
	return DefaultContainer.CreateScope()
}

// AddSingleton registers fn as a singleton provider for T in DefaultContainer.
func AddSingleton[T any](fn any) error {
	return DefaultContainer.AddSingleton[T](fn)
}

// AddNamedSingleton registers fn as a named singleton provider for T in
// DefaultContainer.
func AddNamedSingleton[T any](name string, fn any) error {
	return DefaultContainer.AddNamedSingleton[T](name, fn)
}

// AddScoped registers fn as a scoped provider for T in DefaultContainer.
func AddScoped[T any](fn any) error {
	return DefaultContainer.AddScoped[T](fn)
}

// AddNamedScoped registers fn as a named scoped provider for T in
// DefaultContainer.
func AddNamedScoped[T any](name string, fn any) error {
	return DefaultContainer.AddNamedScoped[T](name, fn)
}

// AddTransient registers fn as a transient provider for T in DefaultContainer.
func AddTransient[T any](fn any) error {
	return DefaultContainer.AddTransient[T](fn)
}

// AddNamedTransient registers fn as a named transient provider for T in
// DefaultContainer.
func AddNamedTransient[T any](name string, fn any) error {
	return DefaultContainer.AddNamedTransient[T](name, fn)
}

// BindSingleton registers V as a singleton provider for T in DefaultContainer.
func BindSingleton[T any, V any]() error {
	return DefaultContainer.BindSingleton[T, V]()
}

// BindNamedSingleton registers V as a named singleton provider for T in
// DefaultContainer.
func BindNamedSingleton[T any, V any](name string) error {
	return DefaultContainer.BindNamedSingleton[T, V](name)
}

// BindScoped registers V as a scoped provider for T in DefaultContainer.
func BindScoped[T any, V any]() error {
	return DefaultContainer.BindScoped[T, V]()
}

// BindNamedScoped registers V as a named scoped provider for T in
// DefaultContainer.
func BindNamedScoped[T any, V any](name string) error {
	return DefaultContainer.BindNamedScoped[T, V](name)
}

// BindTransient registers V as a transient provider for T in DefaultContainer.
func BindTransient[T any, V any]() error {
	return DefaultContainer.BindTransient[T, V]()
}

// BindNamedTransient registers V as a named transient provider for T in
// DefaultContainer.
func BindNamedTransient[T any, V any](name string) error {
	return DefaultContainer.BindNamedTransient[T, V](name)
}
