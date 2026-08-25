package game

import (
	"github.com/mokiat/lacking/game/animation"
	"github.com/mokiat/lacking/game/hierarchy"
)

type AnimationBindingSolver struct{}

var _ hierarchy.SourceBindingSolver[*animation.Player] = (*AnimationBindingSolver)(nil)

func NewAnimationBindingSolver() *AnimationBindingSolver {
	return &AnimationBindingSolver{}
}

func (s *AnimationBindingSolver) OnSourceToNode(hierarchyScene *hierarchy.Scene, nodeID hierarchy.NodeID, player *animation.Player) {
	node := hierarchyScene.Nodes().Handle(nodeID)
	name := node.Name()

	transform := player.BoneTransform(name)
	if transform.Translation.Specified {
		node.SetPosition(transform.Translation.Value)
	}
	if transform.Rotation.Specified {
		node.SetRotation(transform.Rotation.Value)
	}
	if transform.Scale.Specified {
		node.SetScale(transform.Scale.Value)
	}
}
