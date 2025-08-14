package animation

import "github.com/mokiat/gog/opt"

// GraphNodeTransition represents a transition in a GraphNode.
type GraphNodeTransition struct {
	// FadeInFraction determines the amount of time (in fraction of the total
	// animation) that it takes to fade into the transition animation (if there
	// is one).
	FadeInFraction float64

	// FadeOutFraction determines the amount of time (in fraction of the total
	// animation) that it takes to fade out of the transition animation (if there
	// is one).
	FadeOutFraction float64

	// Duration optionally specifies a custom duration for the transition in
	// case that a transition animation is not specified.
	//
	// Defaults to 1.0 (one second).
	Duration opt.T[float64]

	// Animation optionally specifies an animation to be blended on top of the
	// blending from the start state to the end state. If this is not specified
	// then Duration is used to determine the length of the transition.
	Animation opt.T[Node]
}

// NewGraphNode creates a new transition graph animation node.
func NewGraphNode[T comparable]() *GraphNode[T] {
	return &GraphNode[T]{
		animations:  make(map[T]Node),
		transitions: make(map[graphStatePair[T]]GraphNodeTransition),
	}
}

// GraphNode represents an animation node that is comprised of a graph of
// animation states and transitions between them.
type GraphNode[T comparable] struct {
	animations  map[T]Node
	transitions map[graphStatePair[T]]GraphNodeTransition

	fromState          T
	toState            T
	transitionFraction float64
}

var _ Node = (*GraphNode[struct{}])(nil)

// AddState registers a new animation state.
func (n *GraphNode[T]) AddState(state T, animation Node) {
	n.animations[state] = animation
}

// SetState jumps to a specific state and cancels any transitions.
func (n *GraphNode[T]) SetState(state T) {
	n.fromState = state
	n.toState = state
	n.transitionFraction = 1.0
}

// AddTransition registers a new state transition.
func (n *GraphNode[T]) AddTransition(from, to T, transition GraphNodeTransition) {
	pair := graphStatePair[T]{
		from: from,
		to:   to,
	}
	n.transitions[pair] = transition
}

// TransitionTo triggers a new transition. Calling this while there is an
// ongoing transition has undefined behavior.
func (n *GraphNode[T]) TransitionTo(to T) {
	transition, ok := n.transitions[graphStatePair[T]{
		from: n.fromState,
		to:   to,
	}]
	if !ok {
		panic("unknown transition")
	}
	if transitionAnimation, ok := transition.Animation.Unwrap(); ok {
		transitionAnimation.SetFraction(0.0)
	}
	n.toState = to
	n.transitionFraction = 0.0
}

// IsTransitioned returns whether the graph has transitioned to a stable state.
func (n *GraphNode[T]) IsTransitioned() bool {
	return n.fromState == n.toState
}

// Reset clears any update delta information, so that new interpolations can
// be tracked.
func (n *GraphNode[T]) Reset() {
	// TODO: Implement
}

// Rate returns the fraction of the animation length that advances each
// second (fraction per second).
func (n *GraphNode[T]) Rate() float64 {
	return 1.0 // TODO: Implement
}

// Fraction returns the amount of animation that has elapsed. In case of
// looping, the value will wrap around.
//
// The returned value is in the range [0.0..1.0).
func (n *GraphNode[T]) Fraction() float64 {
	return 0.0 // TODO: Implement
}

// SetFraction relocates the animation to the specified fractional position.
//
// NOTE: This resets the animation and accumulated delta is lost.
func (n *GraphNode[T]) SetFraction(fraction float64) {
	// TODO: Implement
}

// Advance moves the animation forward by the specified delta seconds.
//
// The synchronizationRate determines the amount of scaling on the seconds
// that should be applied in order to be correctly synchronized with sibling
// and parent nodes in case of synchronization.
func (n *GraphNode[T]) Advance(seconds, synchronizationRate float64) {
	// TODO: Implement
}

// BoneTransform returns the transformation of the specified bone. Keep in
// mind that this is after a fixed interval update has been applied. If
// this is called from within a dynamic update handler, the
// BoneTransformInterpolation method should be used instead.
func (n *GraphNode[T]) BoneTransform(bone string) NodeTransform {
	return NodeTransform{} // TODO: Implement
}

// BoneTransformDelta returns the transformation that was applied to the
// specified bone since the last reset.
func (n *GraphNode[T]) BoneTransformDelta(bone string) NodeTransform {
	return NodeTransform{}
}

// BoneTransformInterpolation returns the transformation of the specified bone
// at the specified interpolation fraction.
func (n *GraphNode[T]) BoneTransformInterpolation(bone string, fraction float64) NodeTransform {
	return NodeTransform{}
}

type graphStatePair[T comparable] struct {
	from T
	to   T
}
