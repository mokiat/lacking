package animation3

import (
	"fmt"
	"maps"
)

// StateMachineNode is a [Node] that plays one of a set of states, identified
// by keys of type T, and transitions between them. Each state is represented
// by a node of its own and all state nodes are expected to be of the same
// kind (see [Node.Transform]), which is also the kind of the result.
//
// The playback of a StateMachineNode is that of its current state: its rate,
// fraction and output are those of the node of the current state.
type StateMachineNode[T comparable] struct {
	synchronizationAspect
	states       map[T]Node
	currentState T
	currentNode  Node
}

var _ Node = (*StateMachineNode[struct{}])(nil)

// NewStateMachineNode creates a new [StateMachineNode] with the specified
// states, keyed by state, that starts out in the specified initial state.
// The StateMachineNode takes control of the playback of all state nodes,
// which should not be advanced or relocated through other means afterwards.
// The node does not retain the specified map and is not synchronized.
//
// The initial state must be one of the specified states, otherwise the
// function panics.
func NewStateMachineNode[T comparable](initial T, states map[T]Node) *StateMachineNode[T] {
	initialNode, ok := states[initial]
	if !ok {
		panic(fmt.Errorf("initial state %v is not among the specified states", initial))
	}
	return &StateMachineNode[T]{
		states:       maps.Clone(states),
		currentState: initial,
		currentNode:  initialNode,
	}
}

// CurrentState returns the state that the node is in.
func (n *StateMachineNode[T]) CurrentState() T {
	return n.currentState
}

// JumpTo switches to the specified state immediately. Jumping to the current
// state has no effect.
//
// If the specified state is synchronized, it is moved to the fraction of the
// node, so that it continues in phase with the state that it replaces.
// Otherwise, it resumes from its own playback position. To start the state
// from a particular position, call [StateMachineNode.SetFraction] afterwards.
//
// The state must be one of the states of the node, otherwise the method
// panics.
func (n *StateMachineNode[T]) JumpTo(state T) {
	if state == n.currentState {
		return
	}
	node, ok := n.states[state]
	if !ok {
		panic(fmt.Errorf("state %v is not among the states of the node", state))
	}
	if node.IsSynchronized() {
		node.SetFraction(n.currentNode.Fraction())
	}
	n.currentState = state
	n.currentNode = node
}

// Rate returns the rate of the current state.
func (n *StateMachineNode[T]) Rate() float64 {
	return n.currentNode.Rate()
}

// Fraction returns the playback position of the current state.
func (n *StateMachineNode[T]) Fraction() float64 {
	return n.currentNode.Fraction()
}

// SetFraction relocates the current state to the specified fraction.
func (n *StateMachineNode[T]) SetFraction(fraction float64) {
	n.currentNode.SetFraction(fraction)
}

// Advance moves the current state by the specified delta seconds. The nodes
// of the other states are not advanced.
func (n *StateMachineNode[T]) Advance(deltaSeconds float64) {
	n.currentNode.Advance(deltaSeconds)
}

// Transform returns the transformation of the specified bone in the current
// state.
//
// The bone must be a valid index in the skeleton of the state nodes.
func (n *StateMachineNode[T]) Transform(bone int) Transform {
	return n.currentNode.Transform(bone)
}

// DeltaTransform returns the motion that the specified bone experiences in
// the current state throughout the next delta seconds.
//
// The bone must be a valid index in the skeleton of the state nodes.
func (n *StateMachineNode[T]) DeltaTransform(bone int, deltaSeconds float64) Transform {
	return n.currentNode.DeltaTransform(bone, deltaSeconds)
}
