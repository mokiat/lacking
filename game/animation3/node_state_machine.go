package animation3

import (
	"fmt"
	"maps"
)

// StateMachineTransition specifies a cross-fade between two states of a
// [StateMachineNode].
type StateMachineTransition[T comparable] struct {

	// From is the state that the transition starts from.
	From T

	// To is the state that the transition leads to.
	To T

	// Duration is the time in seconds that it takes for the transition to
	// fade from the From state to the To state. A duration of 0.0 switches
	// between the states immediately.
	Duration float64
}

// StateMachineNode is a [Node] that plays one of a set of states, identified
// by keys of type T, and transitions between them. Each state is represented
// by a node of its own and all state nodes are expected to be of the same
// kind (see [Node.Transform]), which is also the kind of the result.
//
// A transition cross-fades from the state that the node is in to another
// state over the duration of the transition. While no transition is in
// progress, the playback of a StateMachineNode is that of its current state:
// its rate, fraction and output are those of the node of the current state.
// During a transition, the node keeps a playback position of its own and
// advances at a rate that is blended from the rates of the two states in the
// same way as by [BlendNode.Rate], so that synchronized states stay in phase.
//
// Only the states that take part in the playback, which are the current
// state and the state that is being faded out, are advanced and relocated.
// The remaining states keep their playback position until they are entered
// again, at which point a synchronized state is moved to the fraction of the
// node.
type StateMachineNode[T comparable] struct {
	synchronizationAspect
	states      map[T]Node
	transitions map[stateMachineEdge[T]]float64

	currentState T
	currentNode  Node

	// The following are only meaningful during a transition, which is in
	// progress when sourceNode is not nil.
	sourceState T
	sourceNode  Node
	duration    float64
	progress    float64
	fraction    float64
}

var _ Node = (*StateMachineNode[struct{}])(nil)

// NewStateMachineNode creates a new [StateMachineNode] with the specified
// states, keyed by state, and the specified transitions between them. The
// node starts out in the specified initial state. The StateMachineNode takes
// control of the playback of all state nodes, which should not be advanced
// or relocated through other means afterwards. The node does not retain the
// specified map or slice and is not synchronized.
//
// The initial state, as well as the From and To states of every transition,
// must be among the specified states. A transition must lead to a state
// other than the one it starts from, must have a non-negative duration and
// must be the only one between its two states in that direction. Otherwise,
// the function panics.
func NewStateMachineNode[T comparable](initial T, states map[T]Node, transitions []StateMachineTransition[T]) *StateMachineNode[T] {
	initialNode, ok := states[initial]
	if !ok {
		panic(fmt.Errorf("initial state %v is not among the specified states", initial))
	}
	edges := make(map[stateMachineEdge[T]]float64, len(transitions))
	for _, transition := range transitions {
		if _, ok := states[transition.From]; !ok {
			panic(fmt.Errorf("transition from unknown state %v", transition.From))
		}
		if _, ok := states[transition.To]; !ok {
			panic(fmt.Errorf("transition to unknown state %v", transition.To))
		}
		if transition.From == transition.To {
			panic(fmt.Errorf("transition from state %v to itself", transition.From))
		}
		if transition.Duration < 0.0 {
			panic(fmt.Errorf("transition from %v to %v has a negative duration", transition.From, transition.To))
		}
		edge := stateMachineEdge[T]{from: transition.From, to: transition.To}
		if _, ok := edges[edge]; ok {
			panic(fmt.Errorf("duplicate transition from %v to %v", transition.From, transition.To))
		}
		edges[edge] = transition.Duration
	}
	return &StateMachineNode[T]{
		states:       maps.Clone(states),
		transitions:  edges,
		currentState: initial,
		currentNode:  initialNode,
	}
}

// CurrentState returns the state that the node is in. During a transition,
// this is the state that the transition leads to.
func (n *StateMachineNode[T]) CurrentState() T {
	return n.currentState
}

// SourceState returns the state that is being faded out by the transition
// in progress. While no transition is in progress, this is the current
// state.
func (n *StateMachineNode[T]) SourceState() T {
	if n.sourceNode == nil {
		return n.currentState
	}
	return n.sourceState
}

// IsTransitioning returns whether a transition is in progress.
func (n *StateMachineNode[T]) IsTransitioning() bool {
	return n.sourceNode != nil
}

// TransitionProgress returns the progress of the transition in progress, in
// the range [0.0..1.0], where 0.0 shows only the
// [StateMachineNode.SourceState] and 1.0 shows only the
// [StateMachineNode.CurrentState]. While no transition is in progress, the
// progress is 1.0.
func (n *StateMachineNode[T]) TransitionProgress() float64 {
	if n.sourceNode == nil {
		return 1.0
	}
	return n.progress
}

// JumpTo switches to the specified state immediately, canceling any
// transition in progress. Jumping to the current state completes a
// transition in progress immediately and otherwise has no effect.
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
		n.sourceNode = nil
		return
	}
	node, ok := n.states[state]
	if !ok {
		panic(fmt.Errorf("state %v is not among the states of the node", state))
	}
	n.jump(state, node)
}

// TransitionTo starts a transition from the current state to the specified
// state, using the transition between the two that was specified at
// construction. If there is no such transition, or if its duration is 0.0,
// this is the same as [StateMachineNode.JumpTo]. Transitioning to the
// current state has no effect.
//
// If the specified state is synchronized, it is moved to the fraction of the
// node when the transition starts, so that it is in phase with the state
// that it replaces while it fades in. Otherwise, it resumes from its own playback
// position.
//
// If a transition is already in progress, the state that it leads to becomes
// the state that is faded out, without a fade, and the new transition starts
// from it. As an exception, transitioning back to the state that is being
// faded out reverses the transition in progress from where it is, using the
// duration of the transition in the opposite direction.
//
// The state must be one of the states of the node, otherwise the method
// panics.
func (n *StateMachineNode[T]) TransitionTo(state T) {
	if state == n.currentState {
		return
	}
	node, ok := n.states[state]
	if !ok {
		panic(fmt.Errorf("state %v is not among the states of the node", state))
	}
	duration, ok := n.transitions[stateMachineEdge[T]{from: n.currentState, to: state}]
	if !ok || duration == 0.0 {
		n.jump(state, node)
		return
	}
	if n.sourceNode != nil && state == n.sourceState {
		n.sourceState, n.currentState = n.currentState, n.sourceState
		n.sourceNode, n.currentNode = n.currentNode, n.sourceNode
		n.duration = duration
		n.progress = 1.0 - n.progress
		return
	}
	n.fraction = n.Fraction()
	if node.IsSynchronized() {
		node.SetFraction(n.fraction)
	}
	n.sourceState, n.sourceNode = n.currentState, n.currentNode
	n.currentState, n.currentNode = state, node
	n.duration = duration
	n.progress = 0.0
}

// Rate returns the rate at which the node advances. While no transition is
// in progress, this is the rate of the current state. During a transition,
// the rate is blended from the rates of the two states in the same way as by
// [BlendNode.Rate].
func (n *StateMachineNode[T]) Rate() float64 {
	if n.sourceNode == nil {
		return n.currentNode.Rate()
	}
	return blendRates(n.sourceNode, n.currentNode, n.progress)
}

// Fraction returns the playback position of the node. While no transition is
// in progress, this is the playback position of the current state.
func (n *StateMachineNode[T]) Fraction() float64 {
	if n.sourceNode == nil {
		return n.currentNode.Fraction()
	}
	return n.fraction
}

// SetFraction relocates the node to the specified fraction. While no
// transition is in progress, this relocates the current state. During a
// transition, the current state is relocated as well, along with the state
// that is being faded out if it is synchronized.
func (n *StateMachineNode[T]) SetFraction(fraction float64) {
	if n.sourceNode == nil {
		n.currentNode.SetFraction(fraction)
		return
	}
	n.fraction = wrapFraction(fraction)
	if n.sourceNode.IsSynchronized() {
		n.sourceNode.SetFraction(n.fraction)
	}
	n.currentNode.SetFraction(n.fraction)
}

// Advance moves the node by the specified delta seconds. While no transition
// is in progress, only the current state is advanced. During a transition,
// both states are advanced by the same delta seconds and synchronized ones
// are then snapped to the fraction of the node, which advances at its
// [StateMachineNode.Rate]. The transition completes once its duration has
// elapsed, after which the state that was faded out is no longer advanced.
func (n *StateMachineNode[T]) Advance(deltaSeconds float64) {
	if n.sourceNode == nil {
		n.currentNode.Advance(deltaSeconds)
		return
	}
	n.fraction = wrapFraction(n.fraction + n.Rate()*deltaSeconds)
	n.sourceNode.Advance(deltaSeconds)
	if n.sourceNode.IsSynchronized() {
		n.sourceNode.SetFraction(n.fraction)
	}
	n.currentNode.Advance(deltaSeconds)
	if n.currentNode.IsSynchronized() {
		n.currentNode.SetFraction(n.fraction)
	}
	n.progress = max(n.progress+deltaSeconds/n.duration, 0.0)
	if n.progress >= 1.0 {
		n.sourceNode = nil
	}
}

// Transform returns the transformation of the specified bone in the current
// state. During a transition, the transformation is blended through
// [BlendTransforms] from that of the state that is being faded out towards
// that of the current state by the progress of the transition.
//
// The bone must be a valid index in the skeleton of the state nodes.
func (n *StateMachineNode[T]) Transform(bone int) Transform {
	if n.sourceNode == nil {
		return n.currentNode.Transform(bone)
	}
	return BlendTransforms(n.sourceNode.Transform(bone), n.currentNode.Transform(bone), n.progress)
}

// DeltaTransform returns the motion that the specified bone experiences in
// the current state throughout the next delta seconds. During a transition,
// the motion is blended through [BlendTransforms] from that of the state
// that is being faded out towards that of the current state by the progress
// of the transition.
//
// During a transition, the result approximates the motion of the blended
// transformation. It does not account for the change in progress throughout
// the delta seconds, and the motion of a synchronized state is measured in
// real time, so it is approximate when the rate of the state differs from
// that of the node.
//
// The bone must be a valid index in the skeleton of the state nodes.
func (n *StateMachineNode[T]) DeltaTransform(bone int, deltaSeconds float64) Transform {
	if n.sourceNode == nil {
		return n.currentNode.DeltaTransform(bone, deltaSeconds)
	}
	return BlendTransforms(n.sourceNode.DeltaTransform(bone, deltaSeconds), n.currentNode.DeltaTransform(bone, deltaSeconds), n.progress)
}

// jump switches to the specified state immediately, canceling any transition
// in progress.
func (n *StateMachineNode[T]) jump(state T, node Node) {
	if node.IsSynchronized() {
		node.SetFraction(n.Fraction())
	}
	n.currentState, n.currentNode = state, node
	n.sourceNode = nil
}

// stateMachineEdge identifies a transition between two states.
type stateMachineEdge[T comparable] struct {
	from T
	to   T
}
