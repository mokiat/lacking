package graphics

import (
	"fmt"

	"github.com/go-gl/gl/v4.6-core/gl"
	"github.com/mokiat/lacking/framework/opengl"
)

type ToneMapping string

const (
	ReinhardToneMapping    ToneMapping = "reinhard"
	ExponentialToneMapping ToneMapping = "exponential"
)

func newPostprocessingMaterial() *PostprocessingMaterial {
	return &PostprocessingMaterial{
		Program: opengl.NewProgram(),
	}
}

type PostprocessingMaterial struct {
	Program *opengl.Program
}

func (m *PostprocessingMaterial) Allocate(toneMapping ToneMapping) {
	vsBuilder := opengl.NewShaderSourceBuilder(postprocessingVertexShaderTemplate)
	fsBuilder := opengl.NewShaderSourceBuilder(postprocessingFragmentShaderTemplate)
	switch toneMapping {
	case ReinhardToneMapping:
		fsBuilder.AddFeature("MODE_REINHARD")
	case ExponentialToneMapping:
		fsBuilder.AddFeature("MODE_EXPONENTIAL")
	default:
		panic(fmt.Errorf("unknown tone mapping mode: %s", toneMapping))
	}

	vertexShader := opengl.NewShader()
	vertexShaderInfo := opengl.ShaderAllocateInfo{
		ShaderType: gl.VERTEX_SHADER,
		SourceCode: vsBuilder.Build(),
	}
	vertexShader.Allocate(vertexShaderInfo)
	defer func() {
		vertexShader.Release()
	}()

	fragmentShader := opengl.NewShader()
	fragmentShaderInfo := opengl.ShaderAllocateInfo{
		ShaderType: gl.FRAGMENT_SHADER,
		SourceCode: fsBuilder.Build(),
	}
	fragmentShader.Allocate(fragmentShaderInfo)
	defer func() {
		fragmentShader.Release()
	}()

	programInfo := opengl.ProgramAllocateInfo{
		VertexShader:   vertexShader,
		FragmentShader: fragmentShader,
	}
	m.Program.Allocate(programInfo)
}

func (m *PostprocessingMaterial) Release() {
	m.Program.Release()
}

const postprocessingVertexShaderTemplate = `
layout(location = 0) in vec2 coordIn;

noperspective out vec2 texCoordInOut;

void main()
{
	texCoordInOut = (coordIn + 1.0) / 2.0;
	gl_Position = vec4(coordIn, 0.0, 1.0);
}
`

const postprocessingFragmentShaderTemplate = `
layout(location = 0) out vec4 fbColor0Out;

uniform sampler2D fbColor0TextureIn;
uniform float exposureIn = 1.0;

noperspective in vec2 texCoordInOut;

void main()
{
	vec3 hdr = texture(fbColor0TextureIn, texCoordInOut).xyz;
	vec3 exposedHDR = hdr * exposureIn;
	#if defined(MODE_REINHARD)
	vec3 ldr = exposedHDR / (exposedHDR + vec3(1.0));
	#endif
	#if defined(MODE_EXPONENTIAL)
	vec3 ldr = vec3(1.0) - exp2(-exposedHDR);
	#endif
	fbColor0Out = vec4(ldr, 1.0);
}
`
