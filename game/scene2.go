package game

import (
	"fmt"

	"github.com/mokiat/gog"
	"github.com/mokiat/gomath/dprec"
	"github.com/mokiat/lacking/game/graphics"
	asset "github.com/mokiat/lacking/game/newasset"
	"github.com/mokiat/lacking/render"
)

// SceneDefinition2 describes a fragment of a game scene.
type SceneDefinition2 struct {
	Textures       []render.Texture
	SkyShaders     []*graphics.SkyShader
	SkyDefinitions []*graphics.SkyDefinition

	Nodes             []asset.Node
	AmbientLights     []asset.AmbientLight
	PointLights       []asset.PointLight
	SpotLights        []asset.SpotLight
	DirectionalLights []asset.DirectionalLight
	Skies             []modelSkyInstance
}

func (s *SceneDefinition2) Delete() {
	// TODO: Implement.
}

type modelSkyInstance struct {
	NodeIndex          uint32
	SkyDefinitionIndex uint32
}

// SceneNodeDefinition describes a node within a fragment of a game scene.
type SceneNodeDefinition struct {
	Name          string
	ParentIndex   int
	Translation   dprec.Vec3
	Rotation      dprec.Quat
	Scale         dprec.Vec3
	IsStationary  bool
	IsInseparable bool
}

func (r *ResourceSet) loadModel2(resource *asset.Resource) (SceneDefinition2, error) {
	var fragmentAsset asset.Model
	ioTask := func() error {
		var err error
		fragmentAsset, err = resource.OpenContent()
		return err
	}
	if err := r.ioWorker.Schedule(ioTask).Wait(); err != nil {
		return SceneDefinition2{}, fmt.Errorf("failed to read asset: %w", err)
	}
	return r.transformModel2Asset(fragmentAsset)
}

func (r *ResourceSet) transformModel2Asset(sceneAsset asset.Model) (SceneDefinition2, error) {
	gfxEngine := r.engine.Graphics()
	renderAPI := gfxEngine.API()

	// TODO: Load textures, shaders, etc. This is what differentiates
	// SceneDefinition with just the asset.Scene.

	textures := make([]render.Texture, len(sceneAsset.Textures))
	for i, textureAsset := range sceneAsset.Textures {
		textures[i] = r.allocateTexture(textureAsset)
	}

	skyShaders := make([]*graphics.SkyShader, len(sceneAsset.SkyShaders))
	for i, skyShaderAsset := range sceneAsset.SkyShaders {
		shaderInfo := graphics.ShaderInfo{
			SourceCode: skyShaderAsset.SourceCode,
		}
		r.gfxWorker.ScheduleVoid(func() {
			skyShaders[i] = gfxEngine.CreateSkyShader(shaderInfo)
		}).Wait()
	}

	skyDefinitions := make([]*graphics.SkyDefinition, len(sceneAsset.Skies))
	for i, skyAsset := range sceneAsset.Skies {
		skyDefinitionInfo := graphics.SkyDefinitionInfo{
			Layers: gog.Map(skyAsset.Layers, func(layerAsset asset.SkyLayer) graphics.SkyLayerDefinitionInfo {
				return graphics.SkyLayerDefinitionInfo{
					Shader:   skyShaders[layerAsset.ShaderIndex],
					Blending: layerAsset.Blending,
				}
			}),
		}
		r.gfxWorker.ScheduleVoid(func() {
			skyDefinition := gfxEngine.CreateSkyDefinition(skyDefinitionInfo)
			for _, binding := range skyAsset.Textures {
				texture := textures[binding.TextureIndex]
				skyDefinition.SetTexture(binding.BindingName, texture)
			}
			for _, binding := range skyAsset.Textures {
				sampler := renderAPI.CreateSampler(render.SamplerInfo{
					Wrapping:   resolveWrapMode(binding.Wrapping),
					Filtering:  resolveFiltering(binding.Filtering),
					Mipmapping: binding.Mipmapping,
				})
				skyDefinition.SetSampler(binding.BindingName, sampler)
			}
			for _, binding := range skyAsset.Properties {
				skyDefinition.SetProperty(binding.BindingName, binding.Data)
			}
			skyDefinitions[i] = skyDefinition
		})
	}

	skies := make([]modelSkyInstance, len(sceneAsset.Skies))
	for i, skyAsset := range sceneAsset.Skies {
		skies[i] = modelSkyInstance{
			NodeIndex:          skyAsset.NodeIndex,
			SkyDefinitionIndex: uint32(i),
		}
	}

	// TODO: Register SceneDefinition2 resource set for cleanup.
	return SceneDefinition2{
		Textures:       textures,
		SkyShaders:     skyShaders,
		SkyDefinitions: skyDefinitions,

		Nodes:             sceneAsset.Nodes,
		AmbientLights:     sceneAsset.AmbientLights,
		PointLights:       sceneAsset.PointLights,
		SpotLights:        sceneAsset.SpotLights,
		DirectionalLights: sceneAsset.DirectionalLights,
		Skies:             skies,
	}, nil
}

func resolveWrapMode(wrap asset.WrapMode) render.WrapMode {
	switch wrap {
	case asset.WrapModeClamp:
		return render.WrapModeClamp
	case asset.WrapModeRepeat:
		return render.WrapModeRepeat
	case asset.WrapModeMirroredRepeat:
		return render.WrapModeMirroredRepeat
	default:
		panic(fmt.Errorf("unknown wrap mode: %v", wrap))
	}
}

func resolveFiltering(filter asset.FilterMode) render.FilterMode {
	switch filter {
	case asset.FilterModeNearest:
		return render.FilterModeNearest
	case asset.FilterModeLinear:
		return render.FilterModeLinear
	case asset.FilterModeAnisotropic:
		return render.FilterModeAnisotropic
	default:
		panic(fmt.Errorf("unknown filter mode: %v", filter))
	}
}

func resolveDataFormat(format asset.TexelFormat) render.DataFormat {
	switch format {
	case asset.TexelFormatRGBA8:
		return render.DataFormatRGBA8
	case asset.TexelFormatRGBA16F:
		return render.DataFormatRGBA16F
	case asset.TexelFormatRGBA32F:
		return render.DataFormatRGBA32F
	default:
		panic(fmt.Errorf("unknown format: %v", format))
	}
}

func resolveCullMode(mode asset.CullMode) render.CullMode {
	switch mode {
	case asset.CullModeNone:
		return render.CullModeNone
	case asset.CullModeFront:
		return render.CullModeFront
	case asset.CullModeBack:
		return render.CullModeBack
	case asset.CullModeFrontAndBack:
		return render.CullModeFrontAndBack
	default:
		panic(fmt.Errorf("unknown cull mode: %v", mode))
	}
}

func resolveFaceOrientation(orientation asset.FaceOrientation) render.FaceOrientation {
	switch orientation {
	case asset.FaceOrientationCCW:
		return render.FaceOrientationCCW
	case asset.FaceOrientationCW:
		return render.FaceOrientationCW
	default:
		panic(fmt.Errorf("unknown face orientation: %v", orientation))
	}
}

func resolveComparison(comparison asset.Comparison) render.Comparison {
	switch comparison {
	case asset.ComparisonNever:
		return render.ComparisonNever
	case asset.ComparisonLess:
		return render.ComparisonLess
	case asset.ComparisonEqual:
		return render.ComparisonEqual
	case asset.ComparisonLessOrEqual:
		return render.ComparisonLessOrEqual
	case asset.ComparisonGreater:
		return render.ComparisonGreater
	case asset.ComparisonNotEqual:
		return render.ComparisonNotEqual
	case asset.ComparisonGreaterOrEqual:
		return render.ComparisonGreaterOrEqual
	case asset.ComparisonAlways:
		return render.ComparisonAlways
	default:
		panic(fmt.Errorf("unknown comparison: %v", comparison))
	}
}
