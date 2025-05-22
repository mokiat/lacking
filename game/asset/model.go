package asset

import (
	"github.com/mokiat/lacking/game/asset/animationdto"
	"github.com/mokiat/lacking/game/asset/backgrounddto"
	"github.com/mokiat/lacking/game/asset/cameradto"
	"github.com/mokiat/lacking/game/asset/hierarchydto"
	"github.com/mokiat/lacking/game/asset/lightingdto"
	"github.com/mokiat/lacking/game/asset/meshdto"
	"github.com/mokiat/lacking/game/asset/physicsdto"
	"github.com/mokiat/lacking/game/asset/shadingdto"
)

// Model represents a virtual world that is composed of various visual
// and logical elements.
type Model struct {
	hierarchydto.HierarchyChunk
	animationdto.AnimationChunk
	shadingdto.ShadingChunk
	lightingdto.LightingChunk
	meshdto.MeshChunk
	physicsdto.PhysicsChunk
	cameradto.CameraChunk
	backgrounddto.BackgroundChunk
}
