package render

import (
	"github.com/go-gl/gl/v4.1-core/gl"
	"github.com/mokiat/gomath/sprec"
	"github.com/mokiat/lacking/async"
	"github.com/mokiat/lacking/game"
	"github.com/mokiat/lacking/graphics"
	"github.com/mokiat/lacking/graphics/preset"
	"github.com/mokiat/lacking/resource"
)

const (
	framebufferWidth  = int32(1920)
	framebufferHeight = int32(1080)
)

func NewScene() *Scene {
	return &Scene{
		geometryFramebuffer: &graphics.Framebuffer{},
		lightingFramebuffer: &graphics.Framebuffer{},
		exposureFramebuffer: &graphics.Framebuffer{},
		screenFramebuffer:   &graphics.Framebuffer{},

		pbrLightingProgram:    &graphics.Program{},
		postprocessingProgram: &graphics.Program{},
		exposureProbeProgram:  &graphics.Program{},
		fwdSkyboxProgram:      &graphics.Program{},
		quadMesh:              &resource.Mesh{},
		cubeMesh:              &resource.Mesh{},
		lastExposure:          1.0,
		layout:                NewLayout(32000.0, 3),
	}
}

type Scene struct {
	geometryFramebuffer *graphics.Framebuffer
	lightingFramebuffer *graphics.Framebuffer
	exposureFramebuffer *graphics.Framebuffer
	screenFramebuffer   *graphics.Framebuffer

	pbrLightingProgram    *graphics.Program
	postprocessingProgram *graphics.Program
	exposureProbeProgram  *graphics.Program
	fwdSkyboxProgram      *graphics.Program
	quadMesh              *resource.Mesh
	cubeMesh              *resource.Mesh

	lastExposure float32
	layout       *Layout
	activeCamera *Camera
}

func (s *Scene) SetActiveCamera(camera *Camera) {
	s.activeCamera = camera
}

func (s *Scene) Layout() *Layout {
	return s.layout
}

func (s *Scene) Init(gfxWorker *async.Worker) async.Outcome {
	gfxTask := func() error {
		geometryFramebufferData := graphics.FramebufferData{
			Width:               framebufferWidth,
			Height:              framebufferHeight,
			HasAlbedoAttachment: true,
			HasNormalAttachment: true,
			HasDepthAttachment:  true,
		}
		if err := s.geometryFramebuffer.Allocate(geometryFramebufferData); err != nil {
			return err
		}
		lightingFramebufferData := graphics.FramebufferData{
			Width:               framebufferWidth,
			Height:              framebufferHeight,
			HasAlbedoAttachment: true,
			UsesHDRAlbedo:       true,
			HasDepthAttachment:  true,
		}
		if err := s.lightingFramebuffer.Allocate(lightingFramebufferData); err != nil {
			return err
		}
		exposureFramebufferData := graphics.FramebufferData{
			Width:               1,
			Height:              1,
			HasAlbedoAttachment: true,
			UsesHDRAlbedo:       true,
		}
		if err := s.exposureFramebuffer.Allocate(exposureFramebufferData); err != nil {
			return err
		}

		lightingShaderData := graphics.ProgramData{
			VertexShaderSourceCode:   lightingVertexSource,
			FragmentShaderSourceCode: lightingFragmentSource,
		}
		if err := s.pbrLightingProgram.Allocate(lightingShaderData); err != nil {
			return err
		}

		postprocessingShaderData := preset.NewPostprocessingShaderData(preset.ReinhardToneMapping)
		if err := s.postprocessingProgram.Allocate(postprocessingShaderData); err != nil {
			return err
		}

		exposureProbeShaderData := preset.NewExposureProbeShaderData()
		if err := s.exposureProbeProgram.Allocate(exposureProbeShaderData); err != nil {
			return err
		}

		skyboxShaderData := graphics.ProgramData{
			VertexShaderSourceCode:   skyboxVertexSource,
			FragmentShaderSourceCode: skyboxFragmentSource,
		}
		if err := s.fwdSkyboxProgram.Allocate(skyboxShaderData); err != nil {
			return err
		}

		quadData := preset.NewQuadVertexArrayData(-1.0, 1.0, 1.0, -1.0)
		s.quadMesh = &resource.Mesh{
			GFXVertexArray: &graphics.VertexArray{},
			SubMeshes: []resource.SubMesh{ // TODO: Get this from preset somehow
				{
					Primitive:   graphics.RenderPrimitiveTriangles,
					IndexOffset: 0,
					IndexCount:  6,
				},
			},
		}
		if err := s.quadMesh.GFXVertexArray.Allocate(quadData); err != nil {
			return err
		}

		cubeData := preset.NewCubeVertexArrayData(-1.0, 1.0, 1.0, -1.0, 1.0, -1.0)
		s.cubeMesh = &resource.Mesh{
			GFXVertexArray: &graphics.VertexArray{},
			SubMeshes: []resource.SubMesh{ // TODO: Get this from preset somehow
				{
					Primitive:   graphics.RenderPrimitiveTriangles,
					IndexOffset: 0,
					IndexCount:  36,
				},
			},
		}
		if err := s.cubeMesh.GFXVertexArray.Allocate(cubeData); err != nil {
			return err
		}
		return nil
	}
	return gfxWorker.Schedule(async.VoidTask(gfxTask))
}

func (s *Scene) Release(registry *resource.Registry, gfxWorker *async.Worker) async.Outcome {
	return async.NewValueOutcome(nil)
}

func (s *Scene) Render(ctx game.RenderContext) {
	// Initialize new framebuffer instance, to avoid race conditions
	s.screenFramebuffer = &graphics.Framebuffer{
		Width:  int32(ctx.WindowSize.Width),
		Height: int32(ctx.WindowSize.Height),
	}

	s.renderGeometryPass(ctx.GFXPipeline)
	s.renderLightingPass(ctx.GFXPipeline)
	s.renderForwardPass(ctx.GFXPipeline)
	s.renderExposureProbePass(ctx.GFXPipeline)
	s.renderPostprocessingPass(ctx.GFXPipeline)
}

func (s *Scene) renderGeometryPass(pipeline *graphics.Pipeline) {
	geometrySequence := pipeline.BeginSequence()
	geometrySequence.TargetFramebuffer = s.geometryFramebuffer
	geometrySequence.BackgroundColor = sprec.NewVec4(0.0, 0.6, 1.0, 1.0)
	geometrySequence.ClearColor = true
	geometrySequence.ClearDepth = true
	geometrySequence.WriteDepth = true
	geometrySequence.DepthFunc = graphics.DepthFuncLessOrEqual
	geometrySequence.ProjectionMatrix = s.activeCamera.ProjectionMatrix()
	geometrySequence.ViewMatrix = s.activeCamera.ViewMatrix()
	// TODO: Pick relevant renderables
	renderable := s.layout.root.renderableList.next
	for renderable != nil {
		for _, node := range renderable.Model.Nodes {
			s.renderModelNode(geometrySequence, renderable.Matrix, node)
		}
		renderable = renderable.next
	}
	pipeline.EndSequence(geometrySequence)
}

func (s *Scene) renderLightingPass(pipeline *graphics.Pipeline) {
	lightingSequence := pipeline.BeginSequence()
	lightingSequence.SourceFramebuffer = s.geometryFramebuffer
	lightingSequence.TargetFramebuffer = s.lightingFramebuffer
	lightingSequence.BlitFramebufferDepth = true
	lightingSequence.BackgroundColor = sprec.NewVec4(1.0, 0.6, 0.0, 1.0)
	lightingSequence.ClearColor = true
	lightingSequence.TestDepth = false
	lightingSequence.WriteDepth = false
	lightingSequence.ProjectionMatrix = s.activeCamera.ProjectionMatrix()
	lightingSequence.ViewMatrix = s.activeCamera.ViewMatrix()
	lightingSequence.CameraMatrix = s.activeCamera.Matrix()
	quadItem := lightingSequence.BeginItem()
	quadItem.Program = s.pbrLightingProgram
	quadItem.VertexArray = s.quadMesh.GFXVertexArray
	quadItem.IndexOffset = s.quadMesh.SubMeshes[0].IndexOffset
	quadItem.IndexCount = s.quadMesh.SubMeshes[0].IndexCount
	quadItem.LightDirectionWS = sprec.UnitVec3(sprec.NewVec3(1.0, 1.0, 1.0))
	lightingSequence.EndItem(quadItem)
	pipeline.EndSequence(lightingSequence)
}

func (s *Scene) renderForwardPass(pipeline *graphics.Pipeline) {
	forwardSequence := pipeline.BeginSequence()
	forwardSequence.SourceFramebuffer = s.geometryFramebuffer // TODO: Do we really need a source?
	forwardSequence.TargetFramebuffer = s.lightingFramebuffer
	forwardSequence.TestDepth = true
	forwardSequence.WriteDepth = false
	forwardSequence.DepthFunc = graphics.DepthFuncLessOrEqual
	forwardSequence.ProjectionMatrix = s.activeCamera.ProjectionMatrix()
	forwardSequence.ViewMatrix = s.activeCamera.ViewMatrix()
	forwardSequence.CameraMatrix = s.activeCamera.Matrix()
	if s.layout.skybox != nil {
		s.renderSkybox(forwardSequence, s.layout.skybox)
	}
	pipeline.EndSequence(forwardSequence)
}

func (s *Scene) renderExposureProbePass(pipeline *graphics.Pipeline) {
	probeSequence := pipeline.BeginSequence()
	probeSequence.SourceFramebuffer = s.lightingFramebuffer
	probeSequence.TargetFramebuffer = s.exposureFramebuffer
	probeSequence.ClearColor = true
	probeSequence.TestDepth = false
	probeSequence.WriteDepth = false
	probeSequence.ProjectionMatrix = s.activeCamera.ProjectionMatrix()
	probeSequence.ViewMatrix = s.activeCamera.ViewMatrix()
	probeSequence.CameraMatrix = s.activeCamera.Matrix()
	quadItem := probeSequence.BeginItem()
	quadItem.Program = s.exposureProbeProgram
	quadItem.VertexArray = s.quadMesh.GFXVertexArray
	quadItem.IndexOffset = s.quadMesh.SubMeshes[0].IndexOffset
	quadItem.IndexCount = s.quadMesh.SubMeshes[0].IndexCount
	probeSequence.EndItem(quadItem)
	pipeline.EndSequence(probeSequence)

	pipeline.SchedulePostRender(func() {
		gl.BindTexture(gl.TEXTURE_2D, s.exposureFramebuffer.AlbedoTextureID)
		data := make([]float32, 4)
		gl.GetTexImage(gl.TEXTURE_2D, 0, gl.RGBA, gl.FLOAT, gl.Ptr(&data[0]))
		brightness := 0.2126*data[0] + 0.7152*data[1] + 0.0722*data[2]
		mix := float32(0.995)
		targetExposure := 1.0 / (9.8 * brightness)
		s.lastExposure = mix*s.lastExposure + (1.0-mix)*targetExposure // FIXME: race condition
	})
}

func (s *Scene) renderPostprocessingPass(pipeline *graphics.Pipeline) {
	exposureSequence := pipeline.BeginSequence()
	exposureSequence.SourceFramebuffer = s.lightingFramebuffer
	exposureSequence.TargetFramebuffer = s.screenFramebuffer
	exposureSequence.ClearColor = true
	exposureSequence.TestDepth = false
	exposureSequence.WriteDepth = false
	exposureSequence.ProjectionMatrix = s.activeCamera.ProjectionMatrix()
	exposureSequence.ViewMatrix = s.activeCamera.ViewMatrix()
	exposureSequence.CameraMatrix = s.activeCamera.Matrix()
	quadItem := exposureSequence.BeginItem()
	quadItem.Program = s.postprocessingProgram
	quadItem.Exposure = s.lastExposure
	quadItem.VertexArray = s.quadMesh.GFXVertexArray
	quadItem.IndexOffset = s.quadMesh.SubMeshes[0].IndexOffset
	quadItem.IndexCount = s.quadMesh.SubMeshes[0].IndexCount
	exposureSequence.EndItem(quadItem)
	pipeline.EndSequence(exposureSequence)
}

func (s *Scene) renderSkybox(sequence *graphics.Sequence, skybox *Skybox) {
	for _, subMesh := range s.cubeMesh.SubMeshes {
		item := sequence.BeginItem()
		item.Program = s.fwdSkyboxProgram
		item.AlbedoCubeTexture = s.layout.skybox.SkyboxTexture
		item.VertexArray = s.cubeMesh.GFXVertexArray
		item.IndexOffset = subMesh.IndexOffset
		item.IndexCount = subMesh.IndexCount
		sequence.EndItem(item)
	}
}

func (s *Scene) renderModelNode(sequence *graphics.Sequence, parentMatrix sprec.Mat4, node *resource.Node) {
	matrix := sprec.Mat4Prod(parentMatrix, node.Matrix)
	s.renderMesh(sequence, matrix, node.Mesh)
	for _, child := range node.Children {
		s.renderModelNode(sequence, matrix, child)
	}
}

func (s *Scene) renderMesh(sequence *graphics.Sequence, modelMatrix sprec.Mat4, mesh *resource.Mesh) {
	for _, subMesh := range mesh.SubMeshes {
		meshItem := sequence.BeginItem()
		meshItem.Program = subMesh.Material.Shader.GeometryProgram
		meshItem.Primitive = subMesh.Primitive
		meshItem.ModelMatrix = modelMatrix
		meshItem.BackfaceCulling = subMesh.Material.BackfaceCulling
		meshItem.Metalness = subMesh.Material.Metalness
		if subMesh.Material.MetalnessTexture != nil {
			meshItem.MetalnessTwoDTexture = subMesh.Material.MetalnessTexture.GFXTexture
		}
		meshItem.Roughness = subMesh.Material.Roughness
		if subMesh.Material.RoughnessTexture != nil {
			meshItem.RoughnessTwoDTexture = subMesh.Material.RoughnessTexture.GFXTexture
		}
		meshItem.AlbedoColor = subMesh.Material.AlbedoColor
		if subMesh.Material.AlbedoTexture != nil {
			meshItem.AlbedoTwoDTexture = subMesh.Material.AlbedoTexture.GFXTexture
		}
		meshItem.NormalScale = subMesh.Material.NormalScale
		if subMesh.Material.NormalTexture != nil {
			meshItem.NormalTwoDTexture = subMesh.Material.NormalTexture.GFXTexture
		}
		meshItem.VertexArray = mesh.GFXVertexArray
		meshItem.IndexOffset = subMesh.IndexOffset
		meshItem.IndexCount = subMesh.IndexCount
		sequence.EndItem(meshItem)
	}
}

const lightingVertexSource = `#version 410

layout(location = 0) in vec3 coordIn;

noperspective out vec2 texCoordInOut;

void main()
{
	texCoordInOut = (coordIn.xy + 1.0) / 2.0;
	gl_Position = vec4(coordIn.xy, 0.0, 1.0);
}
`

const lightingFragmentSource = `#version 410

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

struct brdfInfo {
	float roughness;
	vec3 normal;
	vec3 lightDir;
	vec3 viewDir;
	vec3 halfDir;
};

vec3 getFresnel(vec3 reflectanceF0, brdfInfo brdf) {
	return reflectanceF0 + (1.0 - reflectanceF0) * pow(1.0 - clamp(dot(brdf.halfDir, brdf.lightDir), 0, 1), 5);
}

float getDistribution(brdfInfo brdf) {
	float sqrRough = brdf.roughness * brdf.roughness;
	float hnDot = dot(brdf.normal, brdf.halfDir);
	float denom = hnDot * hnDot * (sqrRough - 1.0) + 1.0;
	return sqrRough / (pi * denom * denom);
}

float getGeometry(brdfInfo brdf) {
	return 1.0 / 4.0;
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

	vec4 albedoMetalness = texture(fbColor0TextureIn, texCoordInOut);
	vec4 normalRoughness = texture(fbColor1TextureIn, texCoordInOut);
	vec3 baseColor = albedoMetalness.xyz;
	vec3 normal = normalize(normalRoughness.xyz);
	float metalness = albedoMetalness.w;
	float roughness = normalRoughness.w;

	vec3 refractedColor = baseColor * (1.0 - metalness);
	vec3 reflectedColor = mix(vec3(0.02), baseColor, metalness);

	vec3 cameraPosition = cameraMatrixIn[3].xyz;
	vec3 lightDir = normalize(lightDirectionWSIn);
	vec3 viewDir = normalize(cameraPosition - worldPosition);
	vec3 halfDir = normalize(lightDir + viewDir);

	brdfInfo brdf = brdfInfo(
		roughness,
		normal,
		lightDir,
		viewDir,
		halfDir
	);

	vec3 fresnel = getFresnel(reflectedColor, brdf);
	vec3 reflectedHDR = fresnel * getDistribution(brdf) * getGeometry(brdf);
	vec3 refractedHDR = (vec3(1.0) - fresnel) * (0.6 + dot(lightDir, normal) * 0.4) * refractedColor / pi;
	vec3 totalHDR = lightIntensity * (refractedHDR + reflectedHDR);

	fragmentColor = vec4(totalHDR, 1.0);
}
`

const skyboxVertexSource = `#version 410

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

const skyboxFragmentSource = `#version 410

layout(location = 0) out vec4 fbColor0Out;

uniform samplerCube albedoCubeTextureIn;

smooth in vec3 texCoordInOut;

void main()
{
	fbColor0Out = texture(albedoCubeTextureIn, texCoordInOut);
}`
