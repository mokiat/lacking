package game

import (
	"fmt"

	"github.com/mokiat/gog"
	"github.com/mokiat/gomath/dprec"
	newasset "github.com/mokiat/lacking/game/newasset"
)

// SceneDefinition2 describes a fragment of a game scene.
//
// Deprecated: Get rid of this and use the newasset.Scene directly.
type SceneDefinition2 struct {
	Nodes             []SceneNodeDefinition
	PointLights       []newasset.PointLight
	SpotLights        []newasset.SpotLight
	DirectionalLights []SceneDirectionalLight
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

// SceneDirectionalLight describes a directional light within a fragment of
// a game scene.
type SceneDirectionalLight struct {
	EmitColor dprec.Vec3
	EmitRange float64
}

func (r *ResourceSet) loadFragment(resource *newasset.Resource) (SceneDefinition2, error) {
	var fragmentAsset newasset.Scene
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

func (r *ResourceSet) transformFragmentAsset(sceneAsset newasset.Scene) (SceneDefinition2, error) {
	nodes := gog.Map(sceneAsset.Nodes, func(node newasset.Node) SceneNodeDefinition {
		return SceneNodeDefinition{
			Name:          node.Name,
			ParentIndex:   int(node.ParentIndex),
			Translation:   node.Translation,
			Rotation:      node.Rotation,
			Scale:         node.Scale,
			IsStationary:  node.Mask&newasset.NodeMaskStationary != 0,
			IsInseparable: node.Mask&newasset.NodeMaskInseparable != 0,
		}
	})

	directionalLights := gog.Map(sceneAsset.DirectionalLights, func(light newasset.DirectionalLight) SceneDirectionalLight {
		return SceneDirectionalLight{
			EmitColor: light.EmitColor,
			EmitRange: light.EmitDistance,
		}
	})

	return SceneDefinition2{
		Nodes:             nodes,
		PointLights:       sceneAsset.PointLights,
		SpotLights:        sceneAsset.SpotLights,
		DirectionalLights: directionalLights,
	}, nil
}
