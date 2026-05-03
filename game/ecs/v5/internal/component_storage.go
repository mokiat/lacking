package internal

type BaseComponentStorage interface {
	// TODO
}

type ComponentStorage[T any] struct {
}

var _ BaseComponentStorage = (*ComponentStorage[struct{}])(nil)
