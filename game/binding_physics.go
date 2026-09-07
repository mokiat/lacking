package game

import (
	"github.com/mokiat/gomath/dprec"
	"github.com/mokiat/lacking/game/hierarchy"
	"github.com/mokiat/lacking/game/physics"
)

type BodyBindingSolver struct {
	physicsScene *physics.Scene
}

var _ hierarchy.LifecycleBindingSolver[physics.BodyID] = (*BodyBindingSolver)(nil)
var _ hierarchy.SourceBindingSolver[physics.BodyID] = (*BodyBindingSolver)(nil)

func NewBodyBindingSolver(physicsScene *physics.Scene) *BodyBindingSolver {
	return &BodyBindingSolver{
		physicsScene: physicsScene,
	}
}

func (s *BodyBindingSolver) OnDelete(_ *hierarchy.Scene, nodeID hierarchy.NodeID, bodyID physics.BodyID) {
	s.physicsScene.Bodies().Delete(bodyID)
}

func (s *BodyBindingSolver) OnSourceToNode(hierarchyScene *hierarchy.Scene, nodeID hierarchy.NodeID, bodyID physics.BodyID) {
	node := hierarchyScene.Nodes().Handle(nodeID)
	body := s.physicsScene.Bodies().Handle(bodyID)

	currentTranslation := body.Position()
	currentRotation := body.Rotation()

	node.SetAbsoluteMatrix(dprec.TRSMat4(
		currentTranslation,
		currentRotation,
		dprec.NewVec3(1.0, 1.0, 1.0),
	))
}
