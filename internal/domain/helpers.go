package domain

func zeroValue[T any]() T {
	var zero T
	return zero
}
