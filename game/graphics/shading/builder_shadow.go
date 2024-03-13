package shading

type ShadowVertexBuilderFunc func(b ShadowVertexBuilder)

func (f ShadowVertexBuilderFunc) GenericBuilder() GenericBuilderFunc {
	return func(builder *Builder) {
		f(builder)
	}
}

type ShadowFragmentBuilderFunc func(b ShadowFragmentBuilder)

func (f ShadowFragmentBuilderFunc) GenericBuilder() GenericBuilderFunc {
	return func(builder *Builder) {
		f(builder)
	}
}

type ShadowVertexBuilder interface {
	CommonBuilder
}

type ShadowFragmentBuilder interface {
	CommonBuilder
}
