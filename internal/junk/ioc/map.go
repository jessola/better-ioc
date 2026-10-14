package ioc

func appendToSliceInMap[K comparable, V any](m map[K][]V, k K, v ...V) {
	_, ok := m[k]
	if !ok {
		m[k] = []V{}

	}
	m[k] = append(m[k], v...)
}
