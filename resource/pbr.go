package resource

import (
	"bytes"
	"fmt"
)

type PBRGeometrySpec struct {
	UsesAlbedoTexture bool
	UsesTexCoord0     bool
}

func BuildPBRGeometryVertexShader(spec PBRGeometrySpec) string {
	return buildPBRGeometryShader(spec, pbrGeometryVertexShaderTemplate)
}

func BuildPBRGeometryFragmentShader(spec PBRGeometrySpec) string {
	return buildPBRGeometryShader(spec, pbrGeometryFragmentShaderTemplate)
}

func buildPBRGeometryShader(spec PBRGeometrySpec, template string) string {
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

const pbrGeometryVertexShaderTemplate = `
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

const pbrGeometryFragmentShaderTemplate = `
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
