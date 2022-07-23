package datastruct

// LessFunc is a function that can determine whether one value is smaller
// than another.
type LessFunc[T any] func(a, b T) bool

// MoreFunc is a function that can determine whether one value is larger
// than another.
type MoreFunc[T any] func(a, b T) bool
