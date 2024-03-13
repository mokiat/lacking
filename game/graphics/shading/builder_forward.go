package shading

type ForwardVertexBuilderFunc func(b ForwardVertexBuilder)

func (f ForwardVertexBuilderFunc) GenericBuilder() GenericBuilderFunc {
	return func(builder *Builder) {
		f(builder)
	}
}

type ForwardFragmentBuilderFunc func(b ForwardFragmentBuilder)

func (f ForwardFragmentBuilderFunc) GenericBuilder() GenericBuilderFunc {
	return func(builder *Builder) {
		f(builder)
	}
}

type ForwardVertexBuilder interface {
	CommonBuilder
}

type ForwardFragmentBuilder interface {
	CommonBuilder

	ForwardOutputColor(color Vec4Variable)
	ForwardAlphaDiscard(alpha Vec1Variable, threshold Vec1Variable)
}
