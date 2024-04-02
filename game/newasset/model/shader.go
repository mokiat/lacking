package model

type Programmable interface {
	SourceCode() string
	SetSourceCode(sourceCode string)
}

type BaseProgrammable struct {
	sourceCode string
}

func (b *BaseProgrammable) SourceCode() string {
	return b.sourceCode
}

func (b *BaseProgrammable) SetSourceCode(sourceCode string) {
	b.sourceCode = sourceCode
}

type Shader struct {
	BaseProgrammable
}
