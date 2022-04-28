package optional

// Unset returns an unspecified value V.
func Unset[T any]() V[T] {
	return V[T]{}
}

// Value returns a specified value V.
func Value[T any](value T) V[T] {
	return V[T]{
		Specified: true,
		Value:     value,
	}
}

// V represents an optional value.
type V[T any] struct {

	// Specified returns whether Value can be used.
	Specified bool

	// Value holds the actual value.
	Value T
}
