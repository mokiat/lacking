package game

import (
	"time"

	"github.com/mokiat/gomath/dprec"
	"github.com/mokiat/lacking/game/hierarchy"
	"github.com/mokiat/lacking/game/sound"
)

// EmitterBindingSolver keeps sound emitters in sync with the hierarchy
// nodes that they are bound to. The position and rotation of an emitter
// follow the interpolated transformation of its node, and its velocity is
// derived from the distance that the node covers during a fixed step.
type EmitterBindingSolver struct {
	soundScene  *sound.Scene
	stepSeconds float64
}

var _ hierarchy.LifecycleBindingSolver[sound.EmitterID] = (*EmitterBindingSolver)(nil)
var _ hierarchy.TargetBindingSolver[sound.EmitterID] = (*EmitterBindingSolver)(nil)
var _ hierarchy.InterpolationBindingSolver[sound.EmitterID] = (*EmitterBindingSolver)(nil)

// NewEmitterBindingSolver creates a new [EmitterBindingSolver] for the
// emitters of the specified sound scene, where fixedTimestep is the
// duration of a single fixed step of the hierarchy.
func NewEmitterBindingSolver(soundScene *sound.Scene, fixedTimestep time.Duration) *EmitterBindingSolver {
	return &EmitterBindingSolver{
		soundScene:  soundScene,
		stepSeconds: fixedTimestep.Seconds(),
	}
}

func (s *EmitterBindingSolver) OnDelete(_ *hierarchy.Scene, nodeID hierarchy.NodeID, emitterID sound.EmitterID) {
	s.soundScene.Emitters().Delete(emitterID)
}

func (s *EmitterBindingSolver) OnTargetFromNode(hierarchyScene *hierarchy.Scene, nodeID hierarchy.NodeID, emitterID sound.EmitterID) {
	node := hierarchyScene.Nodes().Handle(nodeID)
	velocity := nodeStepVelocity(node, s.stepSeconds)

	s.soundScene.Emitters().SetVelocity(emitterID, velocity)
}

func (s *EmitterBindingSolver) OnInterpolationFromNode(hierarchyScene *hierarchy.Scene, nodeID hierarchy.NodeID, emitterID sound.EmitterID, fraction float64) {
	node := hierarchyScene.Nodes().Handle(nodeID)
	matrix := node.InterpolatedAbsoluteMatrix(fraction)
	translation, rotation, _ := matrix.TRS()

	emitter := s.soundScene.Emitters().Handle(emitterID)
	emitter.SetPosition(translation)
	emitter.SetRotation(rotation)
}

// ListenerBindingSolver keeps sound listeners in sync with the hierarchy
// nodes that they are bound to. The position and rotation of a listener
// follow the interpolated transformation of its node, and its velocity is
// derived from the distance that the node covers during a fixed step.
type ListenerBindingSolver struct {
	soundScene  *sound.Scene
	stepSeconds float64
}

var _ hierarchy.LifecycleBindingSolver[sound.ListenerID] = (*ListenerBindingSolver)(nil)
var _ hierarchy.TargetBindingSolver[sound.ListenerID] = (*ListenerBindingSolver)(nil)
var _ hierarchy.InterpolationBindingSolver[sound.ListenerID] = (*ListenerBindingSolver)(nil)

// NewListenerBindingSolver creates a new [ListenerBindingSolver] for the
// listeners of the specified sound scene, where fixedTimestep is the
// duration of a single fixed step of the hierarchy.
func NewListenerBindingSolver(soundScene *sound.Scene, fixedTimestep time.Duration) *ListenerBindingSolver {
	return &ListenerBindingSolver{
		soundScene:  soundScene,
		stepSeconds: fixedTimestep.Seconds(),
	}
}

func (s *ListenerBindingSolver) OnDelete(_ *hierarchy.Scene, nodeID hierarchy.NodeID, listenerID sound.ListenerID) {
	s.soundScene.Listeners().Delete(listenerID)
}

func (s *ListenerBindingSolver) OnTargetFromNode(hierarchyScene *hierarchy.Scene, nodeID hierarchy.NodeID, listenerID sound.ListenerID) {
	node := hierarchyScene.Nodes().Handle(nodeID)
	velocity := nodeStepVelocity(node, s.stepSeconds)

	s.soundScene.Listeners().SetVelocity(listenerID, velocity)
}

func (s *ListenerBindingSolver) OnInterpolationFromNode(hierarchyScene *hierarchy.Scene, nodeID hierarchy.NodeID, listenerID sound.ListenerID, fraction float64) {
	node := hierarchyScene.Nodes().Handle(nodeID)
	matrix := node.InterpolatedAbsoluteMatrix(fraction)
	translation, rotation, _ := matrix.TRS()

	listener := s.soundScene.Listeners().Handle(listenerID)
	listener.SetPosition(translation)
	listener.SetRotation(rotation)
}

// nodeStepVelocity returns the velocity of the node, based on the distance
// between its absolute positions before and after the current fixed step.
// A node that has been snapped is reported as stationary.
func nodeStepVelocity(node hierarchy.NodeHandle, stepSeconds float64) dprec.Vec3 {
	if stepSeconds <= 0.0 {
		return dprec.ZeroVec3()
	}
	oldPosition := node.InterpolatedAbsoluteMatrix(0.0).Translation()
	newPosition := node.InterpolatedAbsoluteMatrix(1.0).Translation()
	return dprec.Vec3Quot(dprec.Vec3Diff(newPosition, oldPosition), stepSeconds)
}
