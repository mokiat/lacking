package dsl

type ApplyFunc func(target any) error

type Operation interface {
	Apply(target any) error
	Digest() ([]byte, error)
}

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
