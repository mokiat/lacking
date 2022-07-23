package datastruct

type LessFunc[T any] func(a, b T) bool

type MoreFunc[T any] func(a, b T) bool
