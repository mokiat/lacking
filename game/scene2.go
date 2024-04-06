package game

import (
	"fmt"

	"github.com/mokiat/gog"
	"github.com/mokiat/gomath/dprec"
	"github.com/mokiat/lacking/game/graphics"
	newasset "github.com/mokiat/lacking/game/newasset"
	"github.com/mokiat/lacking/render"
)

// SceneDefinition2 describes a fragment of a game scene.
type SceneDefinition2 struct {
	SkyShaders     []*graphics.SkyShader
	SkyDefinitions []*graphics.SkyDefinition

	Nodes             []newasset.Node
	PointLights       []newasset.PointLight
	SpotLights        []newasset.SpotLight
	DirectionalLights []newasset.DirectionalLight
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

func (r *ResourceSet) loadModel2(resource *newasset.Resource) (SceneDefinition2, error) {
	var fragmentAsset newasset.Model
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

func (r *ResourceSet) transformModel2Asset(sceneAsset newasset.Model) (SceneDefinition2, error) {
	gfxEngine := r.engine.Graphics()

	// TODO: Load textures, shaders, etc. This is what differentiates
	// SceneDefinition with just the asset.Scene.

	textures := make([]render.Texture, 0) // TODO

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
			Layers: gog.Map(skyAsset.Layers, func(layerAsset newasset.SkyLayer) graphics.SkyLayerDefinitionInfo {
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
		SkyShaders:     skyShaders,
		SkyDefinitions: skyDefinitions,

		Nodes:             sceneAsset.Nodes,
		PointLights:       sceneAsset.PointLights,
		SpotLights:        sceneAsset.SpotLights,
		DirectionalLights: sceneAsset.DirectionalLights,
		Skies:             skies,
	}, nil
}
