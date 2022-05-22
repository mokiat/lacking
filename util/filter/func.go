package filter

// Func represents a filtering function.
// If the function returns true for the specified argument then that argument is
// accepted by the specific filter.
type Func[T any] func(T) bool

// Always allows all input data.
func Always[T any]() Func[T] {
	return func(T) bool {
		return true
	}
}

// Never blocks all input data.
func Never[T any]() Func[T] {
	return func(T) bool {
		return false
	}
}

// All passes the input data only if it is accepted by all filters.
func All[T any](filters ...Func[T]) Func[T] {
	if len(filters) == 0 {
		return Always[T]()
	}
	return func(item T) bool {
		for _, filter := range filters {
			if !filter(item) {
				return false
			}
		}
		return true
	}
}

// Any passes the input data if at least one of the filters accepts the input
// data.
func Any[T any](filters ...Func[T]) Func[T] {
	if len(filters) == 0 {
		return Always[T]()
	}
	return func(item T) bool {
		for _, filter := range filters {
			if filter(item) {
				return true
			}
		}
		return false
	}
}

// Not returns the opposite of a specified filter.
func Not[T any](filter Func[T]) Func[T] {
	return func(item T) bool {
		return !filter(item)
	}
}

// Slice filters the specified slice of entries and returns a new slice that
// contains only those that have passed the filter.
func Slice[T any](entries []T, filter Func[T]) []T {
	var result []T
	for _, entry := range entries {
		if filter(entry) {
			result = append(result, entry)
		}
	}
	return result
}

// SliceIterator iterates over the specified slice of entries and calls
// the specified function only for entries that pass the filter.
func SliceIterator[T any](entries []T, filter Func[T], fn func(T)) {
	for _, entry := range entries {
		if filter(entry) {
			fn(entry)
		}
	}
}
