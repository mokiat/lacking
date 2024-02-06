package game

import (
	"fmt"

	"github.com/mokiat/gog"
	"github.com/mokiat/gomath/dprec"
	newasset "github.com/mokiat/lacking/game/newasset"
)

// FragmentDefinition describes a fragment of a game scene.
type FragmentDefinition struct {
	Nodes             []FragmentNodeDefinition
	PointLights       []FragmentPointLight
	DirectionalLights []FragmentDirectionalLight
}

// FragmentNodeDefinition describes a node within a fragment of a game scene.
type FragmentNodeDefinition struct {
	Name          string
	ParentIndex   int
	Translation   dprec.Vec3
	Rotation      dprec.Quat
	Scale         dprec.Vec3
	IsStationary  bool
	IsInseparable bool
}

// FragmentPointLight describes a point light within a fragment of a game scene.
type FragmentPointLight struct {
	EmitColor dprec.Vec3
	EmitRange float64
}

// FragmentDirectionalLight describes a directional light within a fragment of
// a game scene.
type FragmentDirectionalLight struct {
	EmitColor dprec.Vec3
	EmitRange float64
}

func (r *ResourceSet) loadFragment(resource *newasset.Resource) (FragmentDefinition, error) {
	var fragmentAsset newasset.Fragment
	ioTask := func() error {
		var err error
		fragmentAsset, err = resource.OpenContent()
		return err
	}
	if err := r.ioWorker.Schedule(ioTask).Wait(); err != nil {
		return FragmentDefinition{}, fmt.Errorf("failed to read asset: %w", err)
	}
	return r.transformFragmentAsset(fragmentAsset)
}

func (r *ResourceSet) transformFragmentAsset(fragmentAsset newasset.Fragment) (FragmentDefinition, error) {
	nodes := gog.Map(fragmentAsset.Nodes, func(node newasset.Node) FragmentNodeDefinition {
		return FragmentNodeDefinition{
			Name:          node.Name,
			ParentIndex:   int(node.ParentIndex),
			Translation:   node.Translation,
			Rotation:      node.Rotation,
			Scale:         node.Scale,
			IsStationary:  node.Mask&newasset.NodeMaskStationary != 0,
			IsInseparable: node.Mask&newasset.NodeMaskInseparable != 0,
		}
	})

	pointLights := gog.Map(fragmentAsset.PointLights, func(light newasset.PointLight) FragmentPointLight {
		return FragmentPointLight{
			EmitColor: dprec.Vec3Prod(light.EmitColor, light.EmitIntensity),
			EmitRange: light.EmitRange,
		}
	})

	directionalLights := gog.Map(fragmentAsset.DirectionalLights, func(light newasset.DirectionalLight) FragmentDirectionalLight {
		return FragmentDirectionalLight{
			EmitColor: dprec.Vec3Prod(light.EmitColor, light.EmitIntensity),
			EmitRange: light.EmitRange,
		}
	})

	return FragmentDefinition{
		Nodes:             nodes,
		PointLights:       pointLights,
		DirectionalLights: directionalLights,
	}, nil
}
