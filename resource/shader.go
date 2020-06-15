package resource

import "github.com/mokiat/lacking/graphics"

func InjectShader(target **Shader) func(value interface{}) {
	return func(value interface{}) {
		*target = value.(*Shader)
	}
}

type ShaderInfo struct {
	HasMetalnessTexture bool
	HasRoughnessTexture bool
	HasAlbedoTexture    bool
	HasNormalTexture    bool
}

type Shader struct {
	Type            TypeName
	Info            ShaderInfo
	GeometryProgram *graphics.Program
	ForwardProgram  *graphics.Program
}
