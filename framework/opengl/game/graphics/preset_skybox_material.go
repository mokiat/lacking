package graphics

import (
	"github.com/go-gl/gl/v4.6-core/gl"
	"github.com/mokiat/lacking/framework/opengl"
)

func newSkyboxMaterial() *SkyboxMaterial {
	return &SkyboxMaterial{
		Program: opengl.NewProgram(),
	}
}

type SkyboxMaterial struct {
	Program *opengl.Program
}

func (m *SkyboxMaterial) Allocate() {
	vsBuilder := opengl.NewShaderSourceBuilder(skyboxVertexShaderTemplate)
	fsBuilder := opengl.NewShaderSourceBuilder(skyboxFragmentShaderTemplate)

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

func (m *SkyboxMaterial) Release() {
	m.Program.Release()
}

const skyboxVertexShaderTemplate = `
layout(location = 0) in vec3 coordIn;

uniform mat4 projectionMatrixIn;
uniform mat4 viewMatrixIn;

smooth out vec3 texCoordInOut;

void main()
{
	// we optimize by using vertex coords as cube texture coords
	// additionally, we need to flip the coords. opengl uses renderman coordinate
	// system for cube maps, contrary to the rest of the opengl api
	texCoordInOut = -coordIn;

	// ensure that translations are ignored by setting w to 0.0
	vec4 viewPosition = viewMatrixIn * vec4(coordIn, 0.0);

	// restore w to 1.0 so that projection works
	vec4 position = projectionMatrixIn * vec4(viewPosition.xyz, 1.0);

	// set z to w so that it has maximum depth (1.0) after projection division
	gl_Position = vec4(position.xy, position.w, position.w);
}`

const skyboxFragmentShaderTemplate = `
layout(location = 0) out vec4 fbColor0Out;

uniform samplerCube albedoCubeTextureIn;

smooth in vec3 texCoordInOut;

void main()
{
	fbColor0Out = texture(albedoCubeTextureIn, texCoordInOut);
}`
