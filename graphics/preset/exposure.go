package preset

import (
	"github.com/mokiat/lacking/framework/opengl"
	"github.com/mokiat/lacking/graphics"
)

func NewExposureProbeShaderData() graphics.ProgramData {
	vsBuilder := opengl.NewShaderSourceBuilder(exposureProbeVertexShaderTemplate)
	fsBuilder := opengl.NewShaderSourceBuilder(exposureProbeFragmentShaderTemplate)

	return graphics.ProgramData{
		VertexShaderSourceCode:   vsBuilder.Build(),
		FragmentShaderSourceCode: fsBuilder.Build(),
	}
}

const exposureProbeVertexShaderTemplate = `
layout(location = 0) in vec2 coordIn;

void main()
{
	gl_Position = vec4(coordIn, 0.0, 1.0);
}
`

const exposureProbeFragmentShaderTemplate = `
layout(location = 0) out vec4 fragmentColor;

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
	fragmentColor = vec4(mixture / count, 1.0);
}
`
