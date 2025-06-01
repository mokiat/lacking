package game

import (
	"github.com/mokiat/lacking/game/graphics"
	"github.com/mokiat/lacking/game/hierarchy"
)

// SkyInfo contains the information required to create a sky.
type SkyInfo struct {
	Definition *graphics.SkyDefinition
}

// PlaceSky places a sky on the provided node using the provided definition.
func (s *Scene) PlaceSky(node *hierarchy.Node, info SkyInfo) *graphics.Sky {
	sky := s.gfxScene.CreateSky(graphics.SkyInfo{
		Definition: info.Definition,
	})
	node.SetTarget(SkyNodeTarget{
		Sky: sky,
	})
	return sky
}
