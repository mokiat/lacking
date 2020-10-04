package preset

import (
	"fmt"

	"github.com/mokiat/lacking/graphics"
)

type ToneMapping int

const (
	ReinhardToneMapping ToneMapping = iota
	ExponentialToneMapping
)

func NewPostprocessingShaderData(toneMapping ToneMapping) graphics.ProgramData {
	vsBuilder := NewShaderBuilder(postprocessingVertexShaderTemplate)
	fsBuilder := NewShaderBuilder(postprocessingFragmentShaderTemplate)
	switch toneMapping {
	case ReinhardToneMapping:
		fsBuilder.AddFeature("MODE_REINHARD")
	case ExponentialToneMapping:
		fsBuilder.AddFeature("MODE_EXPONENTIAL")
	default:
		panic(fmt.Errorf("unknown tone mapping type: %d", toneMapping))
	}

	return graphics.ProgramData{
		VertexShaderSourceCode:   vsBuilder.Build(),
		FragmentShaderSourceCode: fsBuilder.Build(),
	}
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
layout(location = 0) out vec4 fragmentColor;

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
	fragmentColor = vec4(ldr, 1.0);
}
`
