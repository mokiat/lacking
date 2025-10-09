package hierarchy

import "github.com/mokiat/gomath/dprec"

// initialTimestamp is a value indicating that the node is uninitialized.
//
// The library uses an approach that borrows ideas from Lamport timestamps used
// in distributed systems. The state's absolute matrix is considered up to date
// if the revision is larger than the parent's revision.
const initialTimestamp int32 = -1

// NodeID represents a unique identifier for a node in the scene.
type NodeID struct {
	index    int32
	revision uint32
}

// IsNil returns whether the NodeID is nil.
func (id NodeID) IsNil() bool {
	return id == NilNodeID
}

// NilNodeID is a sentinel value representing a nil NodeID.
var NilNodeID = NodeID{0, 0}

type node struct {
	index    int32
	revision uint32
	name     string

	isIndependent bool
	isVisible     bool
	transformFunc TransformFunc
	timestamp     int32

	parentIndex       int32
	firstChildIndex   int32
	lastChildIndex    int32
	leftSiblingIndex  int32
	rightSiblingIndex int32

	previousState nodeState
	currentState  nodeState
}

func (n *node) initialize(index int32, revision uint32) {
	n.index = index
	n.revision = revision
	n.name = ""

	n.isIndependent = false
	n.isVisible = true
	n.transformFunc = DefaultTransformFunc
	n.timestamp = initialTimestamp

	n.parentIndex = -1
	n.firstChildIndex = -1
	n.lastChildIndex = -1
	n.leftSiblingIndex = -1
	n.rightSiblingIndex = -1

	n.previousState.initialize()
	n.currentState.initialize()
}

func (n *node) getID() NodeID {
	return NodeID{
		index:    n.index,
		revision: n.revision,
	}
}

func (n *node) prependSibling(scene *Scene, sibling *node) {
	sibling.leftSiblingIndex = n.leftSiblingIndex
	if n.leftSiblingIndex != -1 {
		leftSibling := &scene.nodes[n.leftSiblingIndex]
		leftSibling.rightSiblingIndex = sibling.index
	}

	sibling.rightSiblingIndex = n.index
	n.leftSiblingIndex = sibling.index

	if n.parentIndex != -1 && sibling.leftSiblingIndex == -1 {
		parent := &scene.nodes[n.parentIndex]
		parent.firstChildIndex = sibling.index
	}

	sibling.timestamp = initialTimestamp
}

func (n *node) appendSibling(scene *Scene, sibling *node) {
	sibling.rightSiblingIndex = n.rightSiblingIndex
	if n.rightSiblingIndex != -1 {
		rightSibling := &scene.nodes[n.rightSiblingIndex]
		rightSibling.leftSiblingIndex = sibling.index
	}

	sibling.leftSiblingIndex = n.index
	n.rightSiblingIndex = sibling.index

	if n.parentIndex != -1 && sibling.rightSiblingIndex == -1 {
		parent := &scene.nodes[n.parentIndex]
		parent.lastChildIndex = sibling.index
	}

	sibling.timestamp = initialTimestamp
}

func (n *node) prependChild(scene *Scene, child *node) {
	child.parentIndex = n.index

	child.leftSiblingIndex = -1
	child.rightSiblingIndex = n.firstChildIndex
	if n.firstChildIndex != -1 {
		firstChild := &scene.nodes[n.firstChildIndex]
		firstChild.leftSiblingIndex = child.index
	}
	n.firstChildIndex = child.index

	if n.lastChildIndex == -1 {
		n.lastChildIndex = child.index
	}

	child.timestamp = initialTimestamp
}

func (n *node) appendChild(scene *Scene, child *node) {
	child.parentIndex = n.index

	child.leftSiblingIndex = n.lastChildIndex
	child.rightSiblingIndex = -1
	if n.lastChildIndex != -1 {
		lastChild := &scene.nodes[n.lastChildIndex]
		lastChild.rightSiblingIndex = child.index
	}
	n.lastChildIndex = child.index

	if n.firstChildIndex == -1 {
		n.firstChildIndex = child.index
	}

	child.timestamp = initialTimestamp
}

func (n *node) detach(scene *Scene) {
	if n.parentIndex == -1 {
		return // no-op
	}

	if n.leftSiblingIndex != -1 {
		leftSibling := &scene.nodes[n.leftSiblingIndex]
		leftSibling.rightSiblingIndex = n.rightSiblingIndex
	}
	if n.rightSiblingIndex != -1 {
		rightSibling := &scene.nodes[n.rightSiblingIndex]
		rightSibling.leftSiblingIndex = n.leftSiblingIndex
	}
	parent := &scene.nodes[n.parentIndex]
	if parent.firstChildIndex == n.index {
		parent.firstChildIndex = n.rightSiblingIndex
	}
	if parent.lastChildIndex == n.index {
		parent.lastChildIndex = n.leftSiblingIndex
	}
	n.parentIndex = -1
	n.leftSiblingIndex = -1
	n.rightSiblingIndex = -1
	n.timestamp = initialTimestamp
}

func (n *node) reconstructTransforms(scene *Scene) {
	previousAbsMatrix := n.previousState.absMatrix // should not have been modified
	previousTranslation, previousRotation, previousScale := previousAbsMatrix.TRS()
	n.previousState.translation = previousTranslation
	n.previousState.rotation = previousRotation
	n.previousState.scale = previousScale

	currentAbsMatrix := n.currentState.absMatrix // should not have been modified
	currentTranslation, currentRotation, currentScale := currentAbsMatrix.TRS()
	n.currentState.translation = currentTranslation
	n.currentState.rotation = currentRotation
	n.currentState.scale = currentScale
	n.timestamp = scene.nextTimestamp()
}

func (n *node) resetDelta(scene *Scene, recursive bool) {
	n.ensureNotStale(scene)
	n.previousState = n.currentState

	if recursive {
		for childIndex := n.firstChildIndex; childIndex != -1; {
			child := &scene.nodes[childIndex]
			child.resetDelta(scene, recursive)
			childIndex = child.rightSiblingIndex
		}
	}
}

func (n *node) getName(_ *Scene) string {
	return n.name
}

func (n *node) setName(_ *Scene, name string) {
	n.name = name
}

func (n *node) getIsIndependent(_ *Scene) bool {
	return n.isIndependent
}

func (n *node) setIsIndependent(_ *Scene, independent bool) {
	n.isIndependent = independent
	n.timestamp = initialTimestamp
}

func (n *node) getIsVisible(_ *Scene) bool {
	return n.isVisible
}

func (n *node) setIsVisible(_ *Scene, visible bool) {
	n.isVisible = visible
}

func (n *node) getTransform(_ *Scene) TransformFunc {
	return n.transformFunc
}

func (n *node) setTransform(_ *Scene, transformFunc TransformFunc) {
	if transformFunc == nil {
		n.transformFunc = DefaultTransformFunc
	} else {
		n.transformFunc = transformFunc
	}
	n.timestamp = initialTimestamp
}

func (n *node) getPosition(_ *Scene) dprec.Vec3 {
	return n.currentState.translation
}

func (n *node) setPosition(_ *Scene, position dprec.Vec3) {
	n.currentState.translation = position
	n.timestamp = initialTimestamp
}

func (n *node) getRotation(_ *Scene) dprec.Quat {
	return n.currentState.rotation
}

func (n *node) setRotation(_ *Scene, rotation dprec.Quat) {
	n.currentState.rotation = rotation
	n.timestamp = initialTimestamp
}

func (n *node) getScale(_ *Scene) dprec.Vec3 {
	return n.currentState.scale
}

func (n *node) setScale(_ *Scene, scale dprec.Vec3) {
	n.currentState.scale = scale
	n.timestamp = initialTimestamp
}

func (n *node) getMatrix(_ *Scene) dprec.Mat4 {
	return dprec.TRSMat4(
		n.currentState.translation,
		n.currentState.rotation,
		n.currentState.scale,
	)
}

func (n *node) setMatrix(_ *Scene, matrix dprec.Mat4) {
	translation, rotation, scale := matrix.TRS()
	n.currentState.translation = translation
	n.currentState.rotation = rotation
	n.currentState.scale = scale
	n.timestamp = initialTimestamp
}

func (n *node) getAbsoluteMatrix(scene *Scene) dprec.Mat4 {
	n.ensureNotStale(scene)
	return n.currentState.absMatrix
}

func (n *node) getPreviousAbsoluteTRS(_ *Scene) (dprec.Vec3, dprec.Quat, dprec.Vec3) {
	return n.previousState.absMatrix.TRS()
}

func (n *node) getCurrentAbsoluteTRS(scene *Scene) (dprec.Vec3, dprec.Quat, dprec.Vec3) {
	n.ensureNotStale(scene)
	return n.currentState.absMatrix.TRS()
}

func (n *node) ensureNotStale(scene *Scene) {
	if n.isStale(scene) {
		if n.isIndependent {
			n.currentState.absMatrix = n.getMatrix(scene)
		} else {
			n.currentState.absMatrix = n.transformFunc(scene, NodeID{
				index:    n.index,
				revision: n.revision,
			})
		}
		n.timestamp = scene.nextTimestamp() // make sure to do last
	}
}

func (n *node) isStale(scene *Scene) bool {
	if n.timestamp == initialTimestamp {
		return true // default revision is always stale
	}
	if n.isIndependent || n.parentIndex == -1 {
		return false // no parent and not default revision
	}
	parent := &scene.nodes[n.parentIndex]
	if n.timestamp <= parent.timestamp {
		return true // parent changed since last update
	}
	return parent.isStale(scene)
}
