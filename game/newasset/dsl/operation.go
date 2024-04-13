package dsl

// ApplyFunc is a function that applies an operation to a target.
type ApplyFunc func(target any) error

// Operation is an operation that can be applied to a target.
type Operation interface {
	Digestable
	Apply(target any) error
}

// FuncOperation creates an operation from an apply and a digest functions.
func FuncOperation(apply ApplyFunc, digest DigestFunc) Operation {
	return &funcOperation{
		applyFunc:  apply,
		digestFunc: digest,
	}
}

type funcOperation struct {
	applyFunc  ApplyFunc
	digestFunc DigestFunc
}

func (o *funcOperation) Apply(target any) error {
	return o.applyFunc(target)
}

func (o *funcOperation) Digest() ([]byte, error) {
	return o.digestFunc()
}
