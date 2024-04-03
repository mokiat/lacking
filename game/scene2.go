package game

import (
	"fmt"

	"github.com/mokiat/gomath/dprec"
	"github.com/mokiat/lacking/game/graphics"
	newasset "github.com/mokiat/lacking/game/newasset"
)

// SceneDefinition2 describes a fragment of a game scene.
type SceneDefinition2 struct {
	SkyShaders []graphics.SkyShader

	Nodes             []newasset.Node
	PointLights       []newasset.PointLight
	SpotLights        []newasset.SpotLight
	DirectionalLights []newasset.DirectionalLight
	Skies             []newasset.Sky
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

func (r *ResourceSet) loadFragment(resource *newasset.Resource) (SceneDefinition2, error) {
	var fragmentAsset newasset.Model
	ioTask := func() error {
		var err error
		fragmentAsset, err = resource.OpenContent()
		return err
	}
	if err := r.ioWorker.Schedule(ioTask).Wait(); err != nil {
		return SceneDefinition2{}, fmt.Errorf("failed to read asset: %w", err)
	}
	return r.transformFragmentAsset(fragmentAsset)
}

func (r *ResourceSet) transformFragmentAsset(sceneAsset newasset.Model) (SceneDefinition2, error) {
	// TODO: Load textures, shaders, etc. This is what differentiates
	// SceneDefinition with just the asset.Scene.

	skyShaders := make([]graphics.SkyShader, len(sceneAsset.SkyShaders))
	for i, assetShader := range sceneAsset.SkyShaders {
		r.gfxWorker.ScheduleVoid(func() {
			gfxEngine := r.engine.Graphics()
			skyShader := gfxEngine.CreateSkyShader(graphics.ShaderInfo{
				SourceCode: assetShader.SourceCode,
			})
			skyShaders[i] = skyShader
		}).Wait()
	}
	// TODO: Register sky shaders to resource set for cleanup.

	return SceneDefinition2{
		SkyShaders: skyShaders,

		Nodes:             sceneAsset.Nodes,
		PointLights:       sceneAsset.PointLights,
		SpotLights:        sceneAsset.SpotLights,
		DirectionalLights: sceneAsset.DirectionalLights,
		Skies:             sceneAsset.Skies,
	}, nil
}
