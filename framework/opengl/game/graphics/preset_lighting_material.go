package graphics

import (
	"github.com/mokiat/lacking/framework/opengl"
)

func newLightingMaterial() *LightingMaterial {
	return &LightingMaterial{
		Program: opengl.NewProgram(),
	}
}

type LightingMaterial struct {
	Program *opengl.Program
}

func (m *LightingMaterial) Allocate() {
	vsBuilder := opengl.NewShaderSourceBuilder(lightingVertexSource)
	fsBuilder := opengl.NewShaderSourceBuilder(lightingFragmentSource)

	m.Program = buildProgram(vsBuilder.Build(), fsBuilder.Build())
}

func (m *LightingMaterial) Release() {
	m.Program.Release()
}

const lightingVertexSource = `
layout(location = 0) in vec3 coordIn;

noperspective out vec2 texCoordInOut;

void main()
{
	texCoordInOut = (coordIn.xy + 1.0) / 2.0;
	gl_Position = vec4(coordIn.xy, 0.0, 1.0);
}
`

const lightingFragmentSource = `
layout(location = 0) out vec4 fragmentColor;

uniform sampler2D fbColor0TextureIn;
uniform sampler2D fbColor1TextureIn;
uniform sampler2D fbDepthTextureIn;
uniform mat4 projectionMatrixIn;
uniform mat4 viewMatrixIn;
uniform mat4 cameraMatrixIn;
uniform vec3 lightDirectionWSIn;

noperspective in vec2 texCoordInOut;

const float pi = 3.141592;
const vec3 lightIntensity = vec3(1.2, 1.2, 1.2);

struct fresnelInput {
	vec3 reflectanceF0;
	vec3 halfDirection;
	vec3 lightDirection;
};

vec3 calculateFresnel(fresnelInput i) {
	float halfLightDot = clamp(dot(i.halfDirection, i.lightDirection), 0.0, 1.0);
	return i.reflectanceF0 + (1.0 - i.reflectanceF0) * pow(1.0 - halfLightDot, 5);
}

struct ambientFresnelInput {
	vec3 reflectanceF0;
	vec3 normal;
	vec3 viewDirection;
	float roughness;
};

vec3 calculateAmbientFresnel(ambientFresnelInput i) {
	float normViewDot = clamp(dot(i.normal, i.viewDirection), 0.0, 1.0);
	return i.reflectanceF0 + (max(vec3(1.0 - i.roughness), i.reflectanceF0) - i.reflectanceF0) * pow(1.0 - normViewDot, 5);
}

struct distributionInput {
	float roughness;
	vec3 normal;
	vec3 halfDirection;
};

float calculateDistribution(distributionInput i) {
	float sqrRough = i.roughness * i.roughness;
	float halfNormDot = dot(i.normal, i.halfDirection);
	float denom = halfNormDot * halfNormDot * (sqrRough - 1.0) + 1.0;
	return sqrRough / (pi * denom * denom);
}

struct geometryInput {
	float roughness;
};

// TODO: Use better model
float calculateGeometry(geometryInput i) {
	return 1.0 / 4.0;
}

struct ambientSetup {
	samplerCube reflectionTexture;
	samplerCube refractionTexture;
	float roughness;
	vec3 reflectedColor;
	vec3 refractedColor;
	vec3 viewDirection;
	vec3 normal;
};

vec3 calculateAmbientHDR(ambientSetup s) {
	vec3 fresnel = calculateAmbientFresnel(ambientFresnelInput(
		s.reflectedColor,
		s.normal,
		s.viewDirection,
		s.roughness
	));

	vec3 lightDirection = reflect(s.viewDirection, s.normal);
	vec3 reflectedLightIntensity = pow(mix(
			pow(texture(s.refractionTexture, lightDirection) / pi, vec4(0.25)),
			pow(texture(s.reflectionTexture, lightDirection), vec4(0.25)),
			pow(1.0 - s.roughness, 4)
		), vec4(4)).xyz;
	float geometry = calculateGeometry(geometryInput(
		s.roughness
	));
	vec3 reflectedHDR = fresnel * s.reflectedColor * reflectedLightIntensity * geometry;

	vec3 refractedLightIntensity = texture(s.refractionTexture, -s.normal).xyz;
	vec3 refractedHDR = (vec3(1.0) - fresnel) * s.refractedColor * refractedLightIntensity / pi;

	return (reflectedHDR + refractedHDR);
}

struct directionalSetup {
	float roughness;
	vec3 reflectedColor;
	vec3 refractedColor;
	vec3 viewDirection;
	vec3 lightDirection;
	vec3 normal;
	vec3 lightIntensity;
};

vec3 calculateDirectionalHDR(directionalSetup s) {
	vec3 halfDirection = normalize(s.lightDirection + s.viewDirection);
	vec3 fresnel = calculateFresnel(fresnelInput(
		s.reflectedColor,
		halfDirection,
		s.lightDirection
	));
	float distributionFactor = calculateDistribution(distributionInput(
		s.roughness,
		s.normal,
		halfDirection
	));
	float geometryFactor = calculateGeometry(geometryInput(
		s.roughness
	));
	vec3 reflectedHDR = fresnel * distributionFactor * geometryFactor;
	vec3 refractedHDR = (vec3(1.0) - fresnel) * s.refractedColor / pi;
	return (reflectedHDR + refractedHDR) * s.lightIntensity * clamp(dot(s.normal, s.lightDirection), 0.0, 1.0);
}

void main()
{
	vec3 ndcPosition = vec3(
		(texCoordInOut.x - 0.5) * 2.0,
		(texCoordInOut.y - 0.5) * 2.0,
		texture(fbDepthTextureIn, texCoordInOut).x * 2.0 - 1.0
	);
	vec3 clipPosition = vec3(
		ndcPosition.x / projectionMatrixIn[0][0],
		ndcPosition.y / projectionMatrixIn[1][1],
		-1.0
	);
	vec3 viewPosition = clipPosition * projectionMatrixIn[3][2] / (projectionMatrixIn[2][2] + ndcPosition.z);
	vec3 worldPosition = (cameraMatrixIn * vec4(viewPosition, 1.0)).xyz;
	vec3 cameraPosition = cameraMatrixIn[3].xyz;

	vec4 albedoMetalness = texture(fbColor0TextureIn, texCoordInOut);
	vec4 normalRoughness = texture(fbColor1TextureIn, texCoordInOut);
	vec3 baseColor = albedoMetalness.xyz;
	vec3 normal = normalize(normalRoughness.xyz);
	float metalness = albedoMetalness.w;
	float roughness = normalRoughness.w;

	vec3 refractedColor = baseColor * (1.0 - metalness);
	vec3 reflectedColor = mix(vec3(0.02), baseColor, metalness);

	vec3 totalHDR = vec3(0.0, 0.0, 0.0);
	// totalHDR += calculateAmbientHDR(ambientSetup(
	// 	ambientReflectionTextureIn,
	// 	ambientRefractionTextureIn,
	// 	roughness,
	// 	reflectedColor,
	// 	refractedColor,
	// 	normalize(cameraPosition - worldPosition),
	// 	normal
	// ));
	totalHDR += calculateDirectionalHDR(directionalSetup(
		roughness,
		reflectedColor,
		refractedColor,
		normalize(cameraPosition - worldPosition),
		normalize(lightDirectionWSIn),
		normal,
		lightIntensity
	));
	fragmentColor = vec4(totalHDR, 1.0);
}
`
