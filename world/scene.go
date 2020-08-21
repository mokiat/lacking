package world

import (
	"github.com/mokiat/gomath/sprec"
	"github.com/mokiat/lacking/async"
	"github.com/mokiat/lacking/data"
	"github.com/mokiat/lacking/game"
	"github.com/mokiat/lacking/graphics"
	"github.com/mokiat/lacking/resource"
)

const (
	framebufferWidth  = int32(1024)
	framebufferHeight = int32(576)
)

func NewScene(registry *resource.Registry, gfxWorker *async.Worker) *Scene {
	scene := &Scene{
		geometryFramebuffer: &graphics.Framebuffer{},
		screenFramebuffer:   &graphics.Framebuffer{},
		pbrLightingProgram:  &graphics.Program{},
		fwdSkyboxProgram:    &graphics.Program{},
		layout:              NewLayout(32000.0, 3),
	}

	gfxTask := func() error {
		geometryFramebufferData := graphics.FramebufferData{
			Width:               framebufferWidth,
			Height:              framebufferHeight,
			HasAlbedoAttachment: true,
			HasNormalAttachment: true,
			HasDepthAttachment:  true,
		}
		if err := scene.geometryFramebuffer.Allocate(geometryFramebufferData); err != nil {
			return err
		}

		lightingShaderData := graphics.ProgramData{
			VertexShaderSourceCode:   lightingVertexSource,
			FragmentShaderSourceCode: lightingFragmentSource,
		}
		if err := scene.pbrLightingProgram.Allocate(lightingShaderData); err != nil {
			return err
		}

		scene.quadMesh = &resource.Mesh{
			GFXVertexArray: &graphics.VertexArray{},
			SubMeshes: []resource.SubMesh{
				{
					Primitive:   graphics.RenderPrimitiveTriangles,
					IndexOffset: 0,
					IndexCount:  6,
				},
			},
		}

		vertexData := data.Buffer(make([]byte, 3*4*4))
		vertexData.SetFloat32(4*0, -1.0)
		vertexData.SetFloat32(4*1, 1.0)
		vertexData.SetFloat32(4*2, 0.0)

		vertexData.SetFloat32(4*3, -1.0)
		vertexData.SetFloat32(4*4, -1.0)
		vertexData.SetFloat32(4*5, 0.0)

		vertexData.SetFloat32(4*6, 1.0)
		vertexData.SetFloat32(4*7, -1.0)
		vertexData.SetFloat32(4*8, 0.0)

		vertexData.SetFloat32(4*9, 1.0)
		vertexData.SetFloat32(4*10, 1.0)
		vertexData.SetFloat32(4*11, 0.0)

		indexData := data.Buffer(make([]byte, 6*2))
		indexData.SetUInt16(0, 0)
		indexData.SetUInt16(2, 1)
		indexData.SetUInt16(4, 2)
		indexData.SetUInt16(6, 0)
		indexData.SetUInt16(8, 2)
		indexData.SetUInt16(10, 3)

		vertexArrayData := graphics.VertexArrayData{
			VertexData: vertexData,
			Layout: graphics.VertexArrayLayout{
				HasCoord:    true,
				CoordOffset: 0,
				CoordStride: 3 * 4,
			},
			IndexData: indexData,
		}
		if err := scene.quadMesh.GFXVertexArray.Allocate(vertexArrayData); err != nil {
			return err
		}

		skyboxShaderData := graphics.ProgramData{
			VertexShaderSourceCode:   skyboxVertexSource,
			FragmentShaderSourceCode: skyboxFragmentSource,
		}
		if err := scene.fwdSkyboxProgram.Allocate(skyboxShaderData); err != nil {
			return err
		}

		scene.cubeMesh = &resource.Mesh{
			GFXVertexArray: &graphics.VertexArray{},
			SubMeshes: []resource.SubMesh{
				{
					Primitive:   graphics.RenderPrimitiveTriangles,
					IndexOffset: 0,
					IndexCount:  36,
				},
			},
		}

		vertexData = data.Buffer(make([]byte, 3*8*4))
		vertexData.SetFloat32(4*0, -1.0)
		vertexData.SetFloat32(4*1, 1.0)
		vertexData.SetFloat32(4*2, 1.0)

		vertexData.SetFloat32(4*3, -1.0)
		vertexData.SetFloat32(4*4, -1.0)
		vertexData.SetFloat32(4*5, 1.0)

		vertexData.SetFloat32(4*6, 1.0)
		vertexData.SetFloat32(4*7, -1.0)
		vertexData.SetFloat32(4*8, 1.0)

		vertexData.SetFloat32(4*9, 1.0)
		vertexData.SetFloat32(4*10, 1.0)
		vertexData.SetFloat32(4*11, 1.0)

		vertexData.SetFloat32(4*12, -1.0)
		vertexData.SetFloat32(4*13, 1.0)
		vertexData.SetFloat32(4*14, -1.0)

		vertexData.SetFloat32(4*15, -1.0)
		vertexData.SetFloat32(4*16, -1.0)
		vertexData.SetFloat32(4*17, -1.0)

		vertexData.SetFloat32(4*18, 1.0)
		vertexData.SetFloat32(4*19, -1.0)
		vertexData.SetFloat32(4*20, -1.0)

		vertexData.SetFloat32(4*21, 1.0)
		vertexData.SetFloat32(4*22, 1.0)
		vertexData.SetFloat32(4*23, -1.0)

		indexData = data.Buffer(make([]byte, 36*2))
		indexData.SetUInt16(0, 3)
		indexData.SetUInt16(2, 2)
		indexData.SetUInt16(4, 1)

		indexData.SetUInt16(6, 3)
		indexData.SetUInt16(8, 1)
		indexData.SetUInt16(10, 0)

		indexData.SetUInt16(12, 0)
		indexData.SetUInt16(14, 1)
		indexData.SetUInt16(16, 5)

		indexData.SetUInt16(18, 0)
		indexData.SetUInt16(20, 5)
		indexData.SetUInt16(22, 4)

		indexData.SetUInt16(24, 7)
		indexData.SetUInt16(26, 6)
		indexData.SetUInt16(28, 2)

		indexData.SetUInt16(30, 7)
		indexData.SetUInt16(32, 2)
		indexData.SetUInt16(34, 3)

		indexData.SetUInt16(36, 4)
		indexData.SetUInt16(38, 5)
		indexData.SetUInt16(40, 6)

		indexData.SetUInt16(42, 4)
		indexData.SetUInt16(44, 6)
		indexData.SetUInt16(46, 7)

		indexData.SetUInt16(48, 5)
		indexData.SetUInt16(50, 1)
		indexData.SetUInt16(52, 2)

		indexData.SetUInt16(54, 5)
		indexData.SetUInt16(56, 2)
		indexData.SetUInt16(58, 6)

		indexData.SetUInt16(60, 0)
		indexData.SetUInt16(62, 4)
		indexData.SetUInt16(64, 7)

		indexData.SetUInt16(66, 0)
		indexData.SetUInt16(68, 7)
		indexData.SetUInt16(70, 3)

		vertexArrayData = graphics.VertexArrayData{
			VertexData: vertexData,
			Layout: graphics.VertexArrayLayout{
				HasCoord:    true,
				CoordOffset: 0,
				CoordStride: 3 * 4,
			},
			IndexData: indexData,
		}
		if err := scene.cubeMesh.GFXVertexArray.Allocate(vertexArrayData); err != nil {
			return err
		}

		return nil
	}
	if err := gfxWorker.Wait(async.VoidTask(gfxTask)).Err; err != nil {
		panic(err) // FIXME
	}

	return scene
}

type Scene struct {
	geometryFramebuffer *graphics.Framebuffer
	screenFramebuffer   *graphics.Framebuffer

	pbrLightingProgram *graphics.Program
	quadMesh           *resource.Mesh
	cubeMesh           *resource.Mesh

	fwdSkyboxProgram *graphics.Program

	layout       *Layout
	activeCamera *Camera
}

func (s *Scene) Layout() *Layout {
	return s.layout
}

func (s *Scene) Load(level string) async.Outcome {
	outcome := async.NewOutcome()

	return outcome
}

func (s *Scene) Unload() async.Outcome {
	return async.NewOutcome()
}

func (s *Scene) SetActiveCamera(camera *Camera) {
	s.activeCamera = camera
}

func (s *Scene) Update(ctx game.UpdateContext) {
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
	lightingSequence.TargetFramebuffer = s.screenFramebuffer
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
	lightingSequence.EndItem(quadItem)
	pipeline.EndSequence(lightingSequence)
}

func (s *Scene) renderForwardPass(pipeline *graphics.Pipeline) {
	forwardSequence := pipeline.BeginSequence()
	forwardSequence.SourceFramebuffer = s.geometryFramebuffer
	forwardSequence.TargetFramebuffer = s.screenFramebuffer
	forwardSequence.TestDepth = true
	forwardSequence.WriteDepth = false
	forwardSequence.DepthFunc = graphics.DepthFuncLessOrEqual
	forwardSequence.ProjectionMatrix = s.activeCamera.ProjectionMatrix()
	forwardSequence.ViewMatrix = s.activeCamera.ViewMatrix()
	forwardSequence.CameraMatrix = s.activeCamera.Matrix()
	for _, subMesh := range s.cubeMesh.SubMeshes {
		item := forwardSequence.BeginItem()
		item.Program = s.fwdSkyboxProgram
		item.AlbedoCubeTexture = s.layout.environment.SkyboxTexture
		item.VertexArray = s.cubeMesh.GFXVertexArray
		item.IndexOffset = subMesh.IndexOffset
		item.IndexCount = subMesh.IndexCount
		forwardSequence.EndItem(item)
	}
	pipeline.EndSequence(forwardSequence)
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

noperspective in vec2 texCoordInOut;

const float pi = 3.141592;
const vec3 lightIntensity = vec3(1.2, 1.2, 1.2);
const vec3 lightDirection = vec3(1.0, 1.0, 1.0);

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
	vec3 lightDir = normalize(lightDirection);
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

	vec3 ldr = totalHDR;
	fragmentColor = vec4(ldr, 1.0);
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
