package shading

type Parameter interface {
	IsUsed() bool
	MarkUsed()
}

type BaseParameter struct {
	used bool
}

func (p *BaseParameter) IsUsed() bool {
	return p.used
}

func (p *BaseParameter) MarkUsed() {
	p.used = true
}

type Vec4Param struct {
	BaseParameter
}
