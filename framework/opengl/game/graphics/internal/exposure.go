package internal

import "github.com/mokiat/lacking/framework/opengl"

func NewExposurePresentation() *LightingPresentation {
	vsBuilder := opengl.NewShaderSourceBuilder(exposureVertexShader)
	fsBuilder := opengl.NewShaderSourceBuilder(exposureFragmentShader)
	return NewLightingPresentation(vsBuilder.Build(), fsBuilder.Build())
}

const exposureVertexShader = `
layout(location = 0) in vec2 coordIn;

void main()
{
	gl_Position = vec4(coordIn, 0.0, 1.0);
}
`

const exposureFragmentShader = `
layout(location = 0) out vec4 fbColor0Out;

uniform sampler2D fbColor0TextureIn;

void main()
{
	vec3 mixture = vec3(0.0, 0.0, 0.0);
	float count = 0.0;
	for (float u = 0.0; u <= 1.0; u += 0.05) {
		for (float v = 0.0; v <= 1.0; v += 0.05) {
			mixture += clamp(texture(fbColor0TextureIn, vec2(u, v)).xyz, 0.0, 100.0);
			count += 1.0;
		}
	}
	fbColor0Out = vec4(mixture / count, 1.0);
}
`
