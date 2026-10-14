package ioc

func Value[T any](v T) func() T {
	return func() T {
		return v
	}
}
