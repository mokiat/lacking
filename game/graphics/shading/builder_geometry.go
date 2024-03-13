package shading

type GeometryVertexBuilderFunc func(b GeometryVertexBuilder)

func (f GeometryVertexBuilderFunc) GenericBuilder() GenericBuilderFunc {
	return func(builder *Builder) {
		f(builder)
	}
}

type GeometryFragmentBuilderFunc func(b GeometryFragmentBuilder)

func (f GeometryFragmentBuilderFunc) GenericBuilder() GenericBuilderFunc {
	return func(builder *Builder) {
		f(builder)
	}
}

type GeometryVertexBuilder interface {
	CommonBuilder
}

type GeometryFragmentBuilder interface {
	CommonBuilder
}
