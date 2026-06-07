package animation

import (
	"fmt"

	"github.com/mokiat/gog/opt"
	"github.com/mokiat/gomath/dprec"
)

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
	progress           float64
	synchronized       bool
}

var _ Node = (*GraphNode[struct{}])(nil)

// SourceState returns the current source state of the graph.
func (n *GraphNode[T]) SourceState() T {
	return n.fromState
}

// TargetState returns the current target state of the graph.
func (n *GraphNode[T]) TargetState() T {
	return n.toState
}

// AddState registers a new animation state.
func (n *GraphNode[T]) AddState(state T, animation Node) {
	// TODO: Allow synchronization configuration.
	n.animations[state] = animation
}

// AddTransition registers a new state transition.
func (n *GraphNode[T]) AddTransition(from, to T, transition GraphNodeTransition) {
	pair := graphStatePair[T]{
		from: from,
		to:   to,
	}
	n.transitions[pair] = transition
}

// JumpTo jumps to a specific state and cancels any transitions.
func (n *GraphNode[T]) JumpTo(state T, rewind bool) {
	n.fromState = state
	n.toState = state
	n.transitionFraction = 1.0
	n.SetFraction(0.0)
	if rewind {
		n.animations[state].SetFraction(0.0)
	}
}

// TransitionTo triggers a new transition. Calling this while there is an
// ongoing transition has undefined behavior.
func (n *GraphNode[T]) TransitionTo(to T) {
	if n.fromState == to {
		return // already in the desired state
	}
	if n.toState == to {
		return // already transitioning to state
	}
	transition, ok := n.transitions[graphStatePair[T]{
		from: n.fromState,
		to:   to,
	}]
	if !ok {
		panic(fmt.Errorf("unknown transition %v -> %v", n.fromState, to)) // TODO: Just jump to target state.
	}
	if transitionAnimation, ok := transition.Animation.Unwrap(); ok {
		transitionAnimation.SetFraction(0.0)
	}
	if transition.Rewind {
		n.animations[to].SetFraction(0.0)
	}
	n.toState = to
	n.transitionFraction = 0.0
}

// TransitionFraction returns the amount (fraction) of transition that has
// occurred so far. Returns 1.0 if not in transition.
func (n *GraphNode[T]) TransitionFraction() float64 {
	if n.fromState == n.toState {
		return 1.0
	}
	return n.transitionFraction
}

// IsTransitioned returns whether the graph has transitioned to a stable state.
func (n *GraphNode[T]) IsTransitioned() bool {
	return n.fromState == n.toState
}

// Rate returns the fraction of the animation length that advances each
// second (fraction per second).
func (n *GraphNode[T]) Rate() float64 {
	fromNode := n.animations[n.fromState]
	toNode := n.animations[n.toState]
	return blendRates(fromNode, toNode, n.transitionFraction)
}

// Fraction returns the amount of animation that has elapsed. In case of
// looping, the value will wrap around.
//
// The returned value is in the range [0.0..1.0).
func (n *GraphNode[T]) Fraction() float64 {
	return wrapFraction(n.progress)
}

// SetFraction relocates the animation to the specified fractional position.
//
// NOTE: This resets the animation and accumulated delta is lost.
func (n *GraphNode[T]) SetFraction(fraction float64) {
	n.progress = wrapFraction(fraction)

	for _, animation := range n.animations {
		if animation.IsSynchronized() {
			animation.SetFraction(n.progress)
		}
	}
}

// Advance moves the animation forward by the specified delta seconds.
//
// The synchronizationRate determines the amount of scaling on the seconds
// that should be applied in order to be correctly synchronized with sibling
// and parent nodes in case of synchronization.
func (n *GraphNode[T]) Advance(seconds, synchronizationRate float64) {
	rate := n.Rate()
	n.progress += rate * seconds * synchronizationRate
	n.progress = wrapFraction(n.progress)

	for _, animation := range n.animations {
		if animation.IsSynchronized() {
			adjustedRate := rate / animation.Rate()
			animation.Advance(seconds, synchronizationRate*adjustedRate)
		} else {
			animation.Advance(seconds, 1.0)
		}
	}

	var transitionRate float64
	if n.fromState != n.toState {
		transition := n.transitions[graphStatePair[T]{
			from: n.fromState,
			to:   n.toState,
		}]

		transitionRate = 1.0 / transition.Duration.ValueOrDefault(1.0)
		if animation, ok := transition.Animation.Unwrap(); ok {
			if animation.IsSynchronized() {
				adjustedRate := rate / animation.Rate()
				animation.Advance(seconds, synchronizationRate*adjustedRate)
			} else {
				transitionRate = animation.Rate()
				animation.Advance(seconds, 1.0)
			}
		}

		n.transitionFraction += transitionRate * seconds
		if n.transitionFraction >= 1.0 {
			n.fromState = n.toState // complete transition
		}
		n.transitionFraction = dprec.Clamp(n.transitionFraction, 0.0, 1.0)
	}
}

// IsSynchronized returns whether the node should be synchronized.
func (n *GraphNode[T]) IsSynchronized() bool {
	return n.synchronized
}

// SetSynchronized configures whether the node should be synchronized.
func (n *GraphNode[T]) SetSynchronized(synchronized bool) {
	n.synchronized = synchronized
}

// Synchronize is called each frame to allow a node to synchronized its
// children (depending on their setting).
//
// This will be called (and should be called on children) regardless if
// the current or any child node is synchronized or not.
func (n *GraphNode[T]) Synchronize() {
	for _, animation := range n.animations {
		if animation.IsSynchronized() {
			animation.SetFraction(n.progress)
		}
		animation.Synchronize()
	}
	for _, transition := range n.transitions {
		if animation, ok := transition.Animation.Unwrap(); ok {
			if animation.IsSynchronized() {
				animation.SetFraction(n.progress)
			}
			animation.Synchronize()
		}
	}
}

// BoneTransform returns the transformation of the specified bone. Keep in
// mind that this is after a fixed interval update has been applied. If
// this is called from within a dynamic update handler, the
// BoneTransformInterpolation method should be used instead.
func (n *GraphNode[T]) BoneTransform(bone string) NodeTransform {
	fromNode := n.animations[n.fromState]
	toNode := n.animations[n.toState]
	fromTransform := fromNode.BoneTransform(bone)
	toTransform := toNode.BoneTransform(bone)

	result := BlendNodeTransforms(fromTransform, toTransform, n.transitionFraction)
	if transition, ok := n.animatedTransition(); ok {
		transitionNode := transition.Animation.Value
		transitionTransform := transitionNode.BoneTransform(bone)
		blendFactor := transition.blendFactor(n.transitionFraction)
		result = BlendNodeTransforms(result, transitionTransform, blendFactor)
	}
	return result
}

// BoneDeltaTransform returns the transformation that the bone will experience
// throughout the next delta interval. This is used for root motion.
func (n *GraphNode[T]) BoneDeltaTransform(bone string, delta float64) NodeTransform {
	fromNode := n.animations[n.fromState]
	toNode := n.animations[n.toState]
	fromTransform := fromNode.BoneDeltaTransform(bone, delta)
	toTransform := toNode.BoneDeltaTransform(bone, delta)
	return BlendNodeTransforms(fromTransform, toTransform, n.transitionFraction)
}

func (n *GraphNode[T]) animatedTransition() (GraphNodeTransition, bool) {
	transition, ok := n.transitions[graphStatePair[T]{
		from: n.fromState,
		to:   n.toState,
	}]
	if !ok {
		return GraphNodeTransition{}, false
	}
	if !transition.Animation.Specified {
		return GraphNodeTransition{}, false
	}
	return transition, true
}

type graphStatePair[T comparable] struct {
	from T
	to   T
}

// GraphNodeTransition represents a transition in a GraphNode.
type GraphNodeTransition struct {

	// Rewind indicates whether the target state's animation should be rewound
	// to the beginning when the transition starts.
	Rewind bool

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

func (t GraphNodeTransition) blendFactor(transitionFraction float64) float64 {
	if transitionFraction < t.FadeInFraction {
		return transitionFraction / t.FadeInFraction
	}
	if transitionFraction > 1.0-t.FadeOutFraction {
		return (1.0 - transitionFraction) / t.FadeOutFraction
	}
	return 1.0
}
