package ioc

// Value returns a provider function that always returns v.
func Value[T any](v T) func() T {
	return func() T {
		return v
	}
}
