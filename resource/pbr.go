package resource

import (
	"bytes"
	"fmt"

	"github.com/mokiat/lacking/async"
	"github.com/mokiat/lacking/graphics"
)

const PBRTypeName = TypeName("pbr")

func NewPBRShaderOperator(gfxWorker *async.Worker) *PBRShaderOperator {
	return &PBRShaderOperator{
		gfxWorker: gfxWorker,
	}
}

type PBRShaderOperator struct {
	gfxWorker *async.Worker
}

func (o *PBRShaderOperator) Allocate(info ShaderInfo) (*Shader, error) {
	shader := &Shader{
		Type:            PBRTypeName,
		Info:            info,
		GeometryProgram: &graphics.Program{},
	}

	gfxTask := o.gfxWorker.Schedule(async.VoidTask(func() error {
		spec := PBRGeometrySpec{
			UsesAlbedoTexture: info.HasAlbedoTexture,
			UsesTexCoord0:     info.HasAlbedoTexture,
		}
		return shader.GeometryProgram.Allocate(graphics.ProgramData{
			VertexShaderSourceCode:   BuildGeometryVertexShader(spec),
			FragmentShaderSourceCode: BuildGeometryFragmentShader(spec),
		})
	}))
	if err := gfxTask.Wait().Err; err != nil {
		return nil, fmt.Errorf("failed to allocate gfx program: %w", err)
	}
	return shader, nil
}

func (o *PBRShaderOperator) Release(shader *Shader) error {
	if shader.GeometryProgram != nil {
		gfxTask := o.gfxWorker.Schedule(async.VoidTask(func() error {
			return shader.GeometryProgram.Release()
		}))
		if err := gfxTask.Wait().Err; err != nil {
			return fmt.Errorf("failed to release gfx program: %w", err)
		}
	}
	if shader.ForwardProgram != nil {
		gfxTask := o.gfxWorker.Schedule(async.VoidTask(func() error {
			return shader.ForwardProgram.Release()
		}))
		if err := gfxTask.Wait().Err; err != nil {
			return fmt.Errorf("failed to release gfx program: %w", err)
		}
	}
	return nil
}

type PBRGeometrySpec struct {
	UsesAlbedoTexture bool
	UsesTexCoord0     bool
}

func BuildGeometryVertexShader(spec PBRGeometrySpec) string {
	return buildGeometryShader(spec, geometryVertexShaderTemplate)
}

func BuildGeometryFragmentShader(spec PBRGeometrySpec) string {
	return buildGeometryShader(spec, geometryFragmentShaderTemplate)
}

func buildGeometryShader(spec PBRGeometrySpec, template string) string {
	buffer := &bytes.Buffer{}
	fmt.Fprintln(buffer, "#version 410")
	fmt.Fprintln(buffer)
	if spec.UsesAlbedoTexture {
		fmt.Fprintln(buffer, "#define USES_ALBEDO_TEXTURE")
	}
	if spec.UsesTexCoord0 {
		fmt.Fprintln(buffer, "#define USES_TEX_COORD0")
	}
	fmt.Fprint(buffer, template)
	return buffer.String()
}

const geometryVertexShaderTemplate = `
layout(location = 0) in vec4 coordIn;
layout(location = 1) in vec3 normalIn;
#if defined(USES_TEX_COORD0)
layout(location = 3) in vec2 texCoordIn;
#endif

uniform mat4 projectionMatrixIn;
uniform mat4 modelMatrixIn;
uniform mat4 viewMatrixIn;

smooth out vec3 normalInOut;
#if defined(USES_TEX_COORD0)
smooth out vec2 texCoordInOut;
#endif

void main()
{
#if defined(USES_TEX_COORD0)
	texCoordInOut = texCoordIn;
#endif
	normalInOut = inverse(transpose(mat3(modelMatrixIn))) * normalIn;
	gl_Position = projectionMatrixIn * (viewMatrixIn * (modelMatrixIn * coordIn));
}
`

const geometryFragmentShaderTemplate = `
layout(location = 0) out vec4 fbColor0Out;
layout(location = 1) out vec4 fbColor1Out;

#if defined(USES_ALBEDO_TEXTURE)
uniform sampler2D albedoTwoDTextureIn;
#endif
uniform vec4 albedoColorIn = vec4(0.5, 0.0, 0.5, 1.0);

uniform float metalnessIn = 0.0;
uniform float roughnessIn = 0.8;
uniform float alphaThresholdIn = 0.5;

smooth in vec3 normalInOut;
#if defined(USES_TEX_COORD0)
smooth in vec2 texCoordInOut;
#endif

void main()
{
#if defined(USES_ALBEDO_TEXTURE) && defined(USES_TEX_COORD0)
	vec4 color = texture(albedoTwoDTextureIn, texCoordInOut);
	if (color.a < alphaThresholdIn) {
		discard;
	}
#else
	vec4 color = albedoColorIn;
#endif

	fbColor0Out = vec4(color.xyz, metalnessIn);
	fbColor1Out = vec4(normalize(normalInOut), roughnessIn);
}
`
