package graphics

import (
	"github.com/mokiat/gomath/sprec"
	"github.com/mokiat/lacking/game/graphics/internal"
	"github.com/mokiat/lacking/render"
)

func newSky() *Sky {
	return &Sky{}
}

// Sky represents the Scene's background.
type Sky struct {
	backgroundColor sprec.Vec3
	skyboxTexture   *CubeTexture
}

// BackgroundColor returns the color of the background.
func (s *Sky) BackgroundColor() sprec.Vec3 {
	return s.backgroundColor
}

// SetBackgroundColor changes the color of the background.
func (s *Sky) SetBackgroundColor(color sprec.Vec3) {
	s.backgroundColor = color
}

// // Skybox returns the cube texture to be used as the background.
// // If one has not been set, this method returns nil.
func (s *Sky) Skybox() *CubeTexture {
	return s.skyboxTexture
}

// SetSkybox sets a cube texture to be used as the background.
// If nil is specified, then a texture will not be used and instead
// the background color will be drawn instead.
func (s *Sky) SetSkybox(skybox *CubeTexture) {
	s.skyboxTexture = skybox
}

type Sky2Info struct {
	Layers []Sky2LayerInfo
}

type Sky2LayerInfo struct {
	Blending     bool
	Textures     []render.Texture
	MaterialData []byte
	Shader       SkyShader
}

func newSky2(scene *Scene, info Sky2Info) *Sky2 {
	var layers []Sky2Layer

	cubeShape := scene.renderer.stageData.CubeShape()

	for _, infoLayer := range info.Layers {
		shader := infoLayer.Shader
		programCode := shader.CreateProgramCode(internal.ShaderProgramCodeInfo{})

		program := scene.renderer.api.CreateProgram(render.ProgramInfo{
			SourceCode:      programCode,
			TextureBindings: []render.TextureBinding{}, // TODO
			UniformBindings: []render.UniformBinding{
				render.NewUniformBinding("Camera", internal.UniformBufferBindingCamera),
				render.NewUniformBinding("Material", internal.UniformBufferBindingMaterial),
			},
		})

		pipeline := scene.renderer.api.CreatePipeline(render.PipelineInfo{
			Program:     program,
			VertexArray: cubeShape.VertexArray(),
			Topology:    cubeShape.Topology(),
			Culling:     render.CullModeBack,
			// We are looking from within the cube shape so we need to flip the winding.
			FrontFace:       render.FaceOrientationCW,
			DepthTest:       true,
			DepthWrite:      false,
			DepthComparison: render.ComparisonLessOrEqual,
			StencilTest:     false,
			ColorWrite:      render.ColorMaskTrue,
			BlendEnabled:    false,
		})
		layer := Sky2Layer{
			pipeline:     pipeline,
			materialData: infoLayer.MaterialData,
		}
		layers = append(layers, layer)
	}

	result := &Sky2{
		scene:  scene,
		layers: layers,
	}
	scene.skies.Add(result)
	return result
}

type Sky2 struct {
	scene  *Scene
	layers []Sky2Layer
}

func (s *Sky2) Delete() {
	panic("TODO")
}

type Sky2Layer struct {
	pipeline     render.Pipeline
	textures     []render.Texture
	materialData []byte
}
