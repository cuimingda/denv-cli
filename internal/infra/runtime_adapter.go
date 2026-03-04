package infra

// RuntimeAdapter keeps a normalized runtime reference and a safe fallback for nil receivers.
type RuntimeAdapter[T any] struct {
	runtime  T
	fallback func() T
}

// NewRuntimeAdapter creates a runtime adapter with optional fallback value.
func NewRuntimeAdapter[T any](runtime T, fallback func() T) *RuntimeAdapter[T] {
	return &RuntimeAdapter[T]{
		runtime:  runtime,
		fallback: fallback,
	}
}

// Runtime returns a non-nil-safe runtime value for orchestration calls.
func (a *RuntimeAdapter[T]) Runtime() T {
	if a == nil {
		if a.fallback != nil {
			return a.fallback()
		}
		var zero T
		return zero
	}
	return a.runtime
}
