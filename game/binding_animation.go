package game

import (
	"github.com/mokiat/lacking/game/animation"
	"github.com/mokiat/lacking/game/hierarchy"
)

// AnimationSource specifies the bone of an animation node tree that drives
// a hierarchy node through the [Scene.ArmatureBindingSet].
type AnimationSource struct {

	// Node is the animation node tree whose output drives the hierarchy node.
	Node animation.Node

	// Bone is the index of the bone, in the skeleton of the node tree, whose
	// transformation is applied to the hierarchy node.
	Bone int
}

// AnimationBindingSolver applies the transformation of the bone of an
// [AnimationSource] to the hierarchy node that the source is bound to.
type AnimationBindingSolver struct{}

var _ hierarchy.SourceBindingSolver[AnimationSource] = (*AnimationBindingSolver)(nil)

// NewAnimationBindingSolver creates a new [AnimationBindingSolver].
func NewAnimationBindingSolver() *AnimationBindingSolver {
	return &AnimationBindingSolver{}
}

// OnSourceToNode sets the local transformation of the hierarchy node to the
// current transformation of the bone of the source.
func (s *AnimationBindingSolver) OnSourceToNode(hierarchyScene *hierarchy.Scene, nodeID hierarchy.NodeID, source AnimationSource) {
	transform := source.Node.Transform(source.Bone)
	node := hierarchyScene.Nodes().Handle(nodeID)
	node.SetTRS(transform.Translation, transform.Rotation, transform.Scale)
}
