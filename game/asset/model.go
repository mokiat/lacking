package asset

import (
	"github.com/mokiat/lacking/game/asset/dto/animationdto"
	"github.com/mokiat/lacking/game/asset/dto/backgrounddto"
	"github.com/mokiat/lacking/game/asset/dto/cameradto"
	"github.com/mokiat/lacking/game/asset/dto/hierarchydto"
	"github.com/mokiat/lacking/game/asset/dto/lightingdto"
	"github.com/mokiat/lacking/game/asset/dto/meshdto"
	"github.com/mokiat/lacking/game/asset/dto/physicsdto"
	"github.com/mokiat/lacking/game/asset/dto/shadingdto"
)

// Model represents a virtual world that is composed of various visual
// and logical elements.
type Model struct {
	hierarchydto.HierarchyChunkHolder
	animationdto.AnimationChunkHolder
	shadingdto.ShadingChunkHolder
	lightingdto.LightingChunkHolder
	meshdto.MeshChunkHolder
	physicsdto.PhysicsChunkHolder
	cameradto.CameraChunkHolder
	backgrounddto.BackgroundChunkHolder
}
