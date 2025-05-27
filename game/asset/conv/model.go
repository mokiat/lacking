package conv

import (
	"github.com/mokiat/gog"
	"github.com/mokiat/lacking/game/asset"
	"github.com/mokiat/lacking/game/asset/conv/animationconv"
	"github.com/mokiat/lacking/game/asset/conv/backgroundconv"
	"github.com/mokiat/lacking/game/asset/conv/hierarchyconv"
	"github.com/mokiat/lacking/game/asset/conv/lightingconv"
	"github.com/mokiat/lacking/game/asset/conv/meshconv"
	"github.com/mokiat/lacking/game/asset/conv/physicsconv"
	"github.com/mokiat/lacking/game/asset/conv/shadingconv"
	"github.com/mokiat/lacking/game/asset/dto/animationdto"
	"github.com/mokiat/lacking/game/asset/dto/backgrounddto"
	"github.com/mokiat/lacking/game/asset/dto/cameradto"
	"github.com/mokiat/lacking/game/asset/dto/hierarchydto"
	"github.com/mokiat/lacking/game/asset/dto/lightingdto"
	"github.com/mokiat/lacking/game/asset/dto/meshdto"
	"github.com/mokiat/lacking/game/asset/dto/physicsdto"
	"github.com/mokiat/lacking/game/asset/dto/shadingdto"
	"github.com/mokiat/lacking/game/asset/mdl"
)

// TODO: Implement registry and do conversion using it.

func NewConverter(model *mdl.Model) *Converter {
	return &Converter{
		model: model,
	}
}

type Converter struct {
	model *mdl.Model
}

func (c *Converter) Convert() (asset.Model, error) {
	return asset.Model{
		HierarchyChunkHolder: hierarchydto.HierarchyChunkHolder{
			HierarchyChunk: hierarchyconv.CreateHierarchyChunk(c.model),
		},
		AnimationChunkHolder: animationdto.AnimationChunkHolder{
			AnimationChunk: animationconv.CreateAnimationChunk(c.model),
		},
		ShadingChunkHolder: shadingdto.ShadingChunkHolder{
			ShadingChunk: gog.Must(shadingconv.CreateShadingChunk(c.model)),
		},
		LightingChunkHolder: lightingdto.LightingChunkHolder{
			LightingChunk: lightingconv.CreateLightingChunk(c.model),
		},
		MeshChunkHolder: meshdto.MeshChunkHolder{
			MeshChunk: gog.Must(meshconv.CreateMeshChunk(c.model)),
		},
		PhysicsChunkHolder: physicsdto.PhysicsChunkHolder{
			PhysicsChunk: physicsconv.CreatePhysicsChunk(c.model),
		},
		CameraChunkHolder: cameradto.CameraChunkHolder{
			Camera: nil, // TODO: Implement camera conversion.
		},
		BackgroundChunkHolder: backgrounddto.BackgroundChunkHolder{
			BackgroundChunk: gog.Must(backgroundconv.CreateBackgroundChunk(c.model)),
		},
	}, nil
}
