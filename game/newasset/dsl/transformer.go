package dsl

type Operation interface {
	Apply(target any) error
	Digest() ([]byte, error)
}
