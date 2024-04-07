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
		var texture render.Texture
		switch {
		case textureAsset.Flags.Has(asset.TextureFlag2D):
			// TODO: Use render API directly?
			textureInfo := graphics.TwoDTextureDefinition{
				Width:           int(textureAsset.Width),
				Height:          int(textureAsset.Height),
				Wrapping:        graphics.WrapClampToEdge,
				Filtering:       graphics.FilterNearest,
				GenerateMipmaps: textureAsset.Flags.Has(asset.TextureFlagMipmapping),
				GammaCorrection: !textureAsset.Flags.Has(asset.TextureFlagLinearSpace),
				DataFormat:      resolveDataFormat2(textureAsset.Format),
				InternalFormat:  resolveInternalFormat2(textureAsset.Format),
				Data:            textureAsset.Layers[0].Data,
			}
			r.gfxWorker.ScheduleVoid(func() {
				texture = gfxEngine.CreateTwoDTexture(textureInfo).Texture()
			}).Wait()
		case textureAsset.Flags.Has(asset.TextureFlagCubeMap):
			// TODO: Use render API directly?
			textureInfo := graphics.CubeTextureDefinition{
				Dimension:      int(textureAsset.Width),
				Filtering:      graphics.FilterNearest,
				DataFormat:     resolveDataFormat2(textureAsset.Format),
				InternalFormat: resolveInternalFormat2(textureAsset.Format),
				FrontSideData:  textureAsset.Layers[0].Data,
				BackSideData:   textureAsset.Layers[1].Data,
				LeftSideData:   textureAsset.Layers[2].Data,
				RightSideData:  textureAsset.Layers[3].Data,
				TopSideData:    textureAsset.Layers[4].Data,
				BottomSideData: textureAsset.Layers[5].Data,
			}
			r.gfxWorker.ScheduleVoid(func() {
				texture = gfxEngine.CreateCubeTexture(textureInfo).Texture()
			}).Wait()
		default:
			return SceneDefinition2{}, fmt.Errorf("unsupported texture kind (flags: %v)", textureAsset.Flags)
		}
		textures[i] = texture
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
					Wrapping:   resolveWrapMode2(binding.Wrapping),
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
		PointLights:       sceneAsset.PointLights,
		SpotLights:        sceneAsset.SpotLights,
		DirectionalLights: sceneAsset.DirectionalLights,
		Skies:             skies,
	}, nil
}

func resolveWrapMode2(wrap asset.WrapMode) render.WrapMode {
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

func resolveDataFormat2(format asset.TexelFormat) graphics.DataFormat {
	// FIXME: Support other formats as well
	switch format {
	case asset.TexelFormatRGBA8:
		return graphics.DataFormatRGBA8
	case asset.TexelFormatRGBA16F:
		return graphics.DataFormatRGBA16F
	case asset.TexelFormatRGBA32F:
		return graphics.DataFormatRGBA32F
	default:
		panic(fmt.Errorf("unknown format: %v", format))
	}
}

func resolveInternalFormat2(format asset.TexelFormat) graphics.InternalFormat {
	// FIXME: Support other formats as well
	switch format {
	case asset.TexelFormatRGBA8:
		return graphics.InternalFormatRGBA8
	case asset.TexelFormatRGBA16F:
		return graphics.InternalFormatRGBA16F
	case asset.TexelFormatRGBA32F:
		return graphics.InternalFormatRGBA32F
	default:
		panic(fmt.Errorf("unknown format: %v", format))
	}
}
