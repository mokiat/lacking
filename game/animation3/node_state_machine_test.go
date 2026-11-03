package animation3_test

import (
	"github.com/mokiat/gomath/dprec"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"github.com/mokiat/lacking/game/animation3"
)

var _ = Describe("StateMachineNode", func() {

	const bone = 0

	type state int

	const (
		idleState state = iota
		walkState
		runState
		jumpState
		unknownState
	)

	var (
		skeleton *animation3.Skeleton
		idle     *animation3.ClipNode
		walk     *animation3.ClipNode
		node     *animation3.StateMachineNode[state]
	)

	// cycleNode creates a looping clip node of the specified length that
	// moves the bone along the specified axis by one unit per second.
	cycleNode := func(length float64, axis dprec.Vec3) *animation3.ClipNode {
		clip := animation3.NewClip("cycle", 0.0, length, map[string]animation3.Channel{
			"bone": animation3.NewComponentKeyframeChannel(
				animation3.KeyframeList[dprec.Vec3]{
					{Timestamp: 0.0, Value: dprec.ZeroVec3()},
					{Timestamp: length, Value: dprec.Vec3Prod(axis, length)},
				},
				animation3.KeyframeList[dprec.Quat]{
					{Timestamp: 0.0, Value: dprec.IdentityQuat()},
				},
				animation3.KeyframeList[dprec.Vec3]{
					{Timestamp: 0.0, Value: dprec.NewVec3(1.0, 1.0, 1.0)},
				},
			),
		})
		return animation3.NewClipNode(skeleton, clip, true)
	}

	// positionNode creates a static node that places the bone at the
	// specified X coordinate.
	positionNode := func(x float64) *animation3.FixedPoseNode {
		return animation3.NewFixedPoseNode(skeleton, map[string]animation3.Transform{
			"bone": {
				Translation: dprec.NewVec3(x, 0.0, 0.0),
				Rotation:    dprec.IdentityQuat(),
				Scale:       dprec.NewVec3(1.0, 1.0, 1.0),
			},
		})
	}

	expectX := func(node animation3.Node, expected float64) {
		GinkgoHelper()
		expectVec3(node.Transform(bone).Translation, dprec.NewVec3(expected, 0.0, 0.0))
	}

	BeforeEach(func() {
		skeleton = animation3.NewSkeleton([]animation3.SkeletonBone{
			{Name: "bone", RestTransform: animation3.IdentityTransform()},
		})
		idle = cycleNode(4.0, dprec.BasisYVec3())
		walk = cycleNode(2.0, dprec.BasisXVec3())
		node = animation3.NewStateMachineNode(walkState, map[state]animation3.Node{
			idleState: idle,
			walkState: walk,
		}, animation3.StateMachineTransitionSettings{}, nil)
	})

	Describe("NewStateMachineNode", func() {

		It("starts out in the initial state", func() {
			Expect(node.CurrentState()).To(Equal(walkState))
		})

		It("is not synchronized", func() {
			Expect(node.IsSynchronized()).To(BeFalse())
		})

		It("panics when the initial state is not among the states", func() {
			Expect(func() {
				animation3.NewStateMachineNode(unknownState, map[state]animation3.Node{
					idleState: idle,
				}, animation3.StateMachineTransitionSettings{}, nil)
			}).To(Panic())
		})

		It("panics without states", func() {
			Expect(func() {
				animation3.NewStateMachineNode(idleState, nil, animation3.StateMachineTransitionSettings{}, nil)
			}).To(Panic())
		})

		It("does not retain the specified map", func() {
			states := map[state]animation3.Node{
				idleState: idle,
			}
			node := animation3.NewStateMachineNode(idleState, states, animation3.StateMachineTransitionSettings{}, nil)
			states[idleState] = walk
			node.Advance(1.0)
			Expect(idle.Fraction()).To(BeNumerically("~", 0.25, 1e-9))
			Expect(walk.Fraction()).To(Equal(0.0))
		})

		It("does not retain the specified transitions", func() {
			transitions := []animation3.StateMachineTransition[state]{
				{From: walkState, To: idleState, Duration: 1.0},
			}
			node := animation3.NewStateMachineNode(walkState, map[state]animation3.Node{
				idleState: idle,
				walkState: walk,
			}, animation3.StateMachineTransitionSettings{}, transitions)
			transitions[0].Duration = 0.0
			node.TransitionTo(idleState)
			node.Advance(0.5)
			Expect(walk.Fraction()).To(BeNumerically("~", 0.25, 1e-9))
		})

		DescribeTable("panics for an invalid transition",
			func(transition animation3.StateMachineTransition[state]) {
				Expect(func() {
					animation3.NewStateMachineNode(walkState, map[state]animation3.Node{
						idleState: idle,
						walkState: walk,
					}, animation3.StateMachineTransitionSettings{}, []animation3.StateMachineTransition[state]{transition})
				}).To(Panic())
			},
			Entry("from an unknown state", animation3.StateMachineTransition[state]{From: unknownState, To: idleState, Duration: 1.0}),
			Entry("to an unknown state", animation3.StateMachineTransition[state]{From: walkState, To: unknownState, Duration: 1.0}),
			Entry("to the same state", animation3.StateMachineTransition[state]{From: walkState, To: walkState, Duration: 1.0}),
			Entry("with a negative duration", animation3.StateMachineTransition[state]{From: walkState, To: idleState, Duration: -1.0}),
		)

		It("panics for duplicate transitions", func() {
			Expect(func() {
				animation3.NewStateMachineNode(walkState, map[state]animation3.Node{
					idleState: idle,
					walkState: walk,
				}, animation3.StateMachineTransitionSettings{}, []animation3.StateMachineTransition[state]{
					{From: walkState, To: idleState, Duration: 1.0},
					{From: walkState, To: idleState, Duration: 2.0},
				})
			}).To(Panic())
		})
	})

	Describe("JumpTo", func() {

		It("switches to the specified state", func() {
			node.JumpTo(idleState)
			Expect(node.CurrentState()).To(Equal(idleState))
			Expect(node.Rate()).To(BeNumerically("~", 0.25, 1e-9))
			node.Advance(1.0)
			expectVec3(node.Transform(bone).Translation, dprec.NewVec3(0.0, 1.0, 0.0))
		})

		It("panics for an unknown state", func() {
			Expect(func() {
				node.JumpTo(unknownState)
			}).To(Panic())
		})

		It("has no effect when jumping to the current state", func() {
			node.Advance(0.5)
			node.JumpTo(walkState)
			Expect(node.CurrentState()).To(Equal(walkState))
			Expect(node.Fraction()).To(BeNumerically("~", 0.25, 1e-9))
		})

		It("stops advancing the previous state", func() {
			node.JumpTo(idleState)
			node.Advance(1.0)
			Expect(idle.Fraction()).To(BeNumerically("~", 0.25, 1e-9))
			Expect(walk.Fraction()).To(Equal(0.0))
		})

		DescribeTable("moves a synchronized state to the fraction of the node",
			func(walkSynchronized bool) {
				walk.SetSynchronized(walkSynchronized)
				idle.SetSynchronized(true)
				node.Advance(1.0) // walk at 0.5
				node.JumpTo(idleState)
				Expect(node.Fraction()).To(BeNumerically("~", 0.5, 1e-9))
				Expect(idle.Fraction()).To(BeNumerically("~", 0.5, 1e-9))
			},
			Entry("from a synchronized state", true),
			Entry("from a state that is not synchronized", false),
		)

		DescribeTable("resumes a state that is not synchronized from its own position",
			func(walkSynchronized bool) {
				walk.SetSynchronized(walkSynchronized)
				node.JumpTo(idleState)
				node.Advance(1.0) // idle at 0.25
				node.JumpTo(walkState)
				node.Advance(1.0) // walk at 0.5
				node.JumpTo(idleState)
				Expect(node.Fraction()).To(BeNumerically("~", 0.25, 1e-9))
				node.Advance(0.5)
				Expect(node.Fraction()).To(BeNumerically("~", 0.375, 1e-9))
			},
			Entry("from a synchronized state", true),
			Entry("from a state that is not synchronized", false),
		)

		It("joins the phase of a parent that keeps the node synchronized", func() {
			idle.SetSynchronized(true)
			parent := animation3.NewBlendNode(
				animation3.Synchronized(cycleNode(1.0, dprec.BasisZVec3())),
				animation3.Synchronized(node),
			)
			parent.Advance(0.25) // parent and walk at 0.25
			Expect(walk.Fraction()).To(BeNumerically("~", 0.25, 1e-9))
			node.JumpTo(idleState)
			Expect(idle.Fraction()).To(BeNumerically("~", 0.25, 1e-9))
		})

		It("can be combined with SetFraction to restart the state", func() {
			node.JumpTo(idleState)
			node.Advance(1.0)
			node.JumpTo(walkState)
			node.JumpTo(idleState)
			node.SetFraction(0.0)
			Expect(idle.Fraction()).To(Equal(0.0))
		})
	})

	Describe("TransitionTo", func() {

		var machine *animation3.StateMachineNode[state]

		BeforeEach(func() {
			machine = animation3.NewStateMachineNode(idleState, map[state]animation3.Node{
				idleState: positionNode(1.0),
				walkState: positionNode(2.0),
				runState:  positionNode(4.0),
				jumpState: positionNode(8.0),
			}, animation3.StateMachineTransitionSettings{}, []animation3.StateMachineTransition[state]{
				{From: idleState, To: walkState, Duration: 1.0},
				{From: walkState, To: idleState, Duration: 2.0},
				{From: walkState, To: runState, Duration: 1.0},
				{From: idleState, To: runState, Duration: 0.0},
			})
		})

		It("switches the current state immediately", func() {
			machine.TransitionTo(walkState)
			Expect(machine.CurrentState()).To(Equal(walkState))
			expectX(machine, 1.0)
		})

		It("reports a stable state while no transition is in progress", func() {
			Expect(machine.IsTransitioning()).To(BeFalse())
			Expect(machine.SourceState()).To(Equal(idleState))
			Expect(machine.TransitionProgress()).To(Equal(1.0))
		})

		It("reports the progress of the transition", func() {
			machine.TransitionTo(walkState)
			Expect(machine.IsTransitioning()).To(BeTrue())
			Expect(machine.SourceState()).To(Equal(idleState))
			Expect(machine.TransitionProgress()).To(Equal(0.0))
			machine.Advance(0.25)
			Expect(machine.TransitionProgress()).To(BeNumerically("~", 0.25, 1e-9))
			machine.Advance(0.75)
			Expect(machine.IsTransitioning()).To(BeFalse())
			Expect(machine.SourceState()).To(Equal(walkState))
			Expect(machine.TransitionProgress()).To(Equal(1.0))
		})

		It("reports a stable state after a jump", func() {
			machine.TransitionTo(runState)
			Expect(machine.IsTransitioning()).To(BeFalse())
			Expect(machine.SourceState()).To(Equal(runState))
			Expect(machine.TransitionProgress()).To(Equal(1.0))
		})

		It("cross-fades to the state over the duration of the transition", func() {
			machine.TransitionTo(walkState)
			machine.Advance(0.25)
			expectX(machine, 1.25)
			machine.Advance(0.5)
			expectX(machine, 1.75)
			machine.Advance(0.25)
			expectX(machine, 2.0)
			machine.Advance(1.0)
			expectX(machine, 2.0)
		})

		It("panics for an unknown state", func() {
			Expect(func() {
				machine.TransitionTo(unknownState)
			}).To(Panic())
		})

		It("has no effect when transitioning to the current state", func() {
			machine.TransitionTo(idleState)
			expectX(machine, 1.0)
			machine.TransitionTo(walkState)
			machine.Advance(0.5)
			machine.TransitionTo(walkState)
			expectX(machine, 1.5)
			machine.Advance(0.25)
			expectX(machine, 1.75)
		})

		It("jumps for a transition with zero duration", func() {
			machine.TransitionTo(runState)
			Expect(machine.CurrentState()).To(Equal(runState))
			expectX(machine, 4.0)
		})

		It("jumps when there is no transition between the states", func() {
			machine.JumpTo(runState)
			machine.TransitionTo(idleState)
			Expect(machine.CurrentState()).To(Equal(idleState))
			expectX(machine, 1.0)
		})

		It("starts from the current state when a transition is in progress", func() {
			machine.TransitionTo(walkState)
			machine.Advance(0.5)
			expectX(machine, 1.5)
			machine.TransitionTo(runState)
			Expect(machine.CurrentState()).To(Equal(runState))
			Expect(machine.SourceState()).To(Equal(walkState))
			Expect(machine.TransitionProgress()).To(Equal(0.0))
			expectX(machine, 2.0)
			machine.Advance(0.5)
			expectX(machine, 3.0)
			machine.Advance(0.5)
			expectX(machine, 4.0)
		})

		It("jumps when a transition is in progress and there is no transition from the current state", func() {
			machine.TransitionTo(walkState)
			machine.Advance(0.5)
			machine.TransitionTo(jumpState)
			Expect(machine.CurrentState()).To(Equal(jumpState))
			expectX(machine, 8.0)
			machine.Advance(0.5)
			expectX(machine, 8.0)
		})

		It("reverses a transition in progress when transitioning back", func() {
			machine.TransitionTo(walkState)
			machine.Advance(0.25)
			expectX(machine, 1.25)
			machine.TransitionTo(idleState)
			Expect(machine.CurrentState()).To(Equal(idleState))
			Expect(machine.SourceState()).To(Equal(walkState))
			Expect(machine.TransitionProgress()).To(BeNumerically("~", 0.75, 1e-9))
			expectX(machine, 1.25)
			machine.Advance(0.25) // half of the remaining 0.5 seconds of the reverse transition
			expectX(machine, 1.125)
			machine.Advance(0.25)
			expectX(machine, 1.0)
		})

		It("jumps back when there is no transition in the opposite direction", func() {
			machine.JumpTo(walkState)
			machine.TransitionTo(runState)
			machine.Advance(0.5)
			expectX(machine, 3.0)
			machine.TransitionTo(walkState)
			Expect(machine.CurrentState()).To(Equal(walkState))
			expectX(machine, 2.0)
		})

		It("is canceled by JumpTo", func() {
			machine.TransitionTo(walkState)
			machine.Advance(0.5)
			machine.JumpTo(jumpState)
			Expect(machine.IsTransitioning()).To(BeFalse())
			expectX(machine, 8.0)
			machine.Advance(0.5)
			expectX(machine, 8.0)
		})

		It("is completed by JumpTo to the current state", func() {
			machine.TransitionTo(walkState)
			machine.Advance(0.5)
			machine.JumpTo(walkState)
			Expect(machine.CurrentState()).To(Equal(walkState))
			Expect(machine.IsTransitioning()).To(BeFalse())
			expectX(machine, 2.0)
		})
	})

	Describe("default transition", func() {

		It("panics for a negative duration", func() {
			Expect(func() {
				animation3.NewStateMachineNode(idleState, map[state]animation3.Node{
					idleState: idle,
				}, animation3.StateMachineTransitionSettings{Duration: -1.0}, nil)
			}).To(Panic())
		})

		Context("with a duration", func() {

			var machine *animation3.StateMachineNode[state]

			BeforeEach(func() {
				machine = animation3.NewStateMachineNode(idleState, map[state]animation3.Node{
					idleState: positionNode(1.0),
					walkState: positionNode(2.0),
					runState:  positionNode(4.0),
				}, animation3.StateMachineTransitionSettings{Duration: 1.0}, []animation3.StateMachineTransition[state]{
					{From: idleState, To: runState, Duration: 0.5},
				})
			})

			It("cross-fades between states without a transition of their own", func() {
				machine.TransitionTo(walkState)
				Expect(machine.IsTransitioning()).To(BeTrue())
				machine.Advance(0.5)
				expectX(machine, 1.5)
				machine.Advance(0.5)
				expectX(machine, 2.0)
			})

			It("is not used between states with a transition of their own", func() {
				machine.TransitionTo(runState)
				machine.Advance(0.25)
				expectX(machine, 2.5)
			})

			It("reverses a transition in progress", func() {
				machine.TransitionTo(walkState)
				machine.Advance(0.25)
				machine.TransitionTo(idleState)
				Expect(machine.IsTransitioning()).To(BeTrue())
				expectX(machine, 1.25)
				machine.Advance(0.25)
				expectX(machine, 1.0)
			})
		})

		It("can rewind the states that are switched to", func() {
			machine := animation3.NewStateMachineNode(walkState, map[state]animation3.Node{
				idleState: idle,
				walkState: walk,
			}, animation3.StateMachineTransitionSettings{Rewind: true}, nil)
			machine.Advance(1.0) // walk at 0.5
			machine.TransitionTo(idleState)
			machine.Advance(1.0) // idle at 0.25
			machine.TransitionTo(walkState)
			Expect(machine.IsTransitioning()).To(BeFalse())
			Expect(walk.Fraction()).To(Equal(0.0))
			machine.TransitionTo(idleState)
			Expect(idle.Fraction()).To(Equal(0.0))
		})
	})

	Describe("transition playback", func() {

		var machine *animation3.StateMachineNode[state]

		BeforeEach(func() {
			machine = animation3.NewStateMachineNode(walkState, map[state]animation3.Node{
				idleState: idle,
				walkState: walk,
			}, animation3.StateMachineTransitionSettings{}, []animation3.StateMachineTransition[state]{
				{From: walkState, To: idleState, Duration: 1.0},
			})
		})

		It("advances both states during a transition and only the current one afterwards", func() {
			machine.TransitionTo(idleState)
			machine.Advance(0.5)
			Expect(walk.Fraction()).To(BeNumerically("~", 0.25, 1e-9))
			Expect(idle.Fraction()).To(BeNumerically("~", 0.125, 1e-9))
			machine.Advance(0.5)
			Expect(walk.Fraction()).To(BeNumerically("~", 0.5, 1e-9))
			Expect(idle.Fraction()).To(BeNumerically("~", 0.25, 1e-9))
			machine.Advance(1.0)
			Expect(walk.Fraction()).To(BeNumerically("~", 0.5, 1e-9))
			Expect(idle.Fraction()).To(BeNumerically("~", 0.5, 1e-9))
		})

		Context("with a rewinding transition", func() {

			BeforeEach(func() {
				machine = animation3.NewStateMachineNode(walkState, map[state]animation3.Node{
					idleState: idle,
					walkState: walk,
				}, animation3.StateMachineTransitionSettings{}, []animation3.StateMachineTransition[state]{
					{From: walkState, To: idleState, Duration: 1.0, Rewind: true},
					{From: idleState, To: walkState, Duration: 1.0, Rewind: true},
				})
				machine.JumpTo(idleState)
				machine.Advance(1.0) // idle at 0.25
				machine.JumpTo(walkState)
				machine.Advance(1.0) // walk at 0.5
			})

			It("starts the state from its beginning", func() {
				machine.TransitionTo(idleState)
				Expect(idle.Fraction()).To(Equal(0.0))
			})

			It("does not relocate the state that is being faded out", func() {
				machine.TransitionTo(idleState)
				Expect(walk.Fraction()).To(BeNumerically("~", 0.5, 1e-9))
				machine.Advance(0.5)
				Expect(walk.Fraction()).To(BeNumerically("~", 0.75, 1e-9))
				Expect(idle.Fraction()).To(BeNumerically("~", 0.125, 1e-9))
			})

			It("starts the state from its beginning when jumping with zero duration", func() {
				machine := animation3.NewStateMachineNode(walkState, map[state]animation3.Node{
					idleState: idle,
					walkState: walk,
				}, animation3.StateMachineTransitionSettings{}, []animation3.StateMachineTransition[state]{
					{From: walkState, To: idleState, Duration: 0.0, Rewind: true},
				})
				machine.TransitionTo(idleState)
				Expect(machine.IsTransitioning()).To(BeFalse())
				Expect(machine.CurrentState()).To(Equal(idleState))
				Expect(idle.Fraction()).To(Equal(0.0))
			})

			It("keeps a synchronized state in phase instead", func() {
				idle.SetSynchronized(true)
				machine.TransitionTo(idleState)
				Expect(idle.Fraction()).To(BeNumerically("~", 0.5, 1e-9))
			})

			It("does not rewind when reversing a transition in progress", func() {
				machine.TransitionTo(idleState)
				machine.Advance(0.5) // walk at 0.75, idle at 0.125
				machine.TransitionTo(walkState)
				Expect(machine.IsTransitioning()).To(BeTrue())
				Expect(walk.Fraction()).To(BeNumerically("~", 0.75, 1e-9))
			})
		})

		It("does not rewind without a rewinding transition", func() {
			machine.JumpTo(idleState)
			machine.Advance(1.0) // idle at 0.25
			machine.JumpTo(walkState)
			machine.TransitionTo(idleState)
			Expect(idle.Fraction()).To(BeNumerically("~", 0.25, 1e-9))
		})

		It("moves a synchronized state to the fraction of the node when the transition starts", func() {
			idle.SetSynchronized(true)
			machine.Advance(0.5) // walk at 0.25
			machine.TransitionTo(idleState)
			Expect(machine.Fraction()).To(BeNumerically("~", 0.25, 1e-9))
			Expect(idle.Fraction()).To(BeNumerically("~", 0.25, 1e-9))
		})

		It("keeps synchronized states in phase at a blended rate", func() {
			idle.SetSynchronized(true)
			walk.SetSynchronized(true)
			machine.TransitionTo(idleState)
			Expect(machine.Rate()).To(BeNumerically("~", 0.5, 1e-9))
			machine.Advance(0.5)
			Expect(machine.Fraction()).To(BeNumerically("~", 0.25, 1e-9))
			Expect(walk.Fraction()).To(BeNumerically("~", 0.25, 1e-9))
			Expect(idle.Fraction()).To(BeNumerically("~", 0.25, 1e-9))
			Expect(machine.Rate()).To(BeNumerically("~", 1.0/3.0, 1e-9)) // halfway between 2 and 4 seconds
		})

		It("does not progress its fraction when neither state is synchronized", func() {
			machine.Advance(0.5) // walk at 0.25
			machine.TransitionTo(idleState)
			Expect(machine.Rate()).To(Equal(0.0))
			machine.Advance(0.5)
			Expect(machine.Fraction()).To(BeNumerically("~", 0.25, 1e-9))
		})

		DescribeTable("relocates the current state and a synchronized faded out state on SetFraction",
			func(walkSynchronized bool, expectedWalk float64) {
				walk.SetSynchronized(walkSynchronized)
				machine.TransitionTo(idleState)
				machine.SetFraction(0.5)
				Expect(machine.Fraction()).To(BeNumerically("~", 0.5, 1e-9))
				Expect(idle.Fraction()).To(BeNumerically("~", 0.5, 1e-9))
				Expect(walk.Fraction()).To(BeNumerically("~", expectedWalk, 1e-9))
			},
			Entry("synchronized faded out state", true, 0.5),
			Entry("faded out state that is not synchronized", false, 0.0),
		)

		It("blends the output of the two states", func() {
			machine.Advance(1.0) // walk at x=1
			machine.TransitionTo(idleState)
			machine.Advance(0.5) // walk at x=1.5, idle at y=0.5
			expectVec3(machine.Transform(bone).Translation, dprec.NewVec3(0.75, 0.25, 0.0))
		})

		It("blends the motion of the two states", func() {
			machine.TransitionTo(idleState)
			machine.Advance(0.5)
			delta := machine.DeltaTransform(bone, 0.5)
			expectVec3(delta.Translation, dprec.NewVec3(0.25, 0.25, 0.0))
		})
	})

	Describe("Rate", func() {

		It("returns the rate of the current state", func() {
			Expect(node.Rate()).To(BeNumerically("~", 0.5, 1e-9))
		})
	})

	Describe("SetFraction", func() {

		It("relocates the current state only", func() {
			node.SetFraction(0.25)
			Expect(node.Fraction()).To(BeNumerically("~", 0.25, 1e-9))
			Expect(walk.Fraction()).To(BeNumerically("~", 0.25, 1e-9))
			Expect(idle.Fraction()).To(Equal(0.0))
		})
	})

	Describe("Advance", func() {

		It("advances the current state only", func() {
			node.Advance(0.5)
			Expect(node.Fraction()).To(BeNumerically("~", 0.25, 1e-9))
			Expect(walk.Fraction()).To(BeNumerically("~", 0.25, 1e-9))
			Expect(idle.Fraction()).To(Equal(0.0))
		})
	})

	Describe("Transform", func() {

		It("returns the transformation of the current state", func() {
			node.Advance(0.5)
			expectVec3(node.Transform(bone).Translation, dprec.NewVec3(0.5, 0.0, 0.0))
		})
	})

	Describe("DeltaTransform", func() {

		It("returns the motion of the current state", func() {
			delta := node.DeltaTransform(bone, 0.5)
			expectVec3(delta.Translation, dprec.NewVec3(0.5, 0.0, 0.0))
		})

		It("does not change the playback position", func() {
			node.DeltaTransform(bone, 0.5)
			Expect(node.Fraction()).To(Equal(0.0))
		})
	})

	Describe("SetSynchronized", func() {

		It("configures the node without affecting its states", func() {
			node.SetSynchronized(true)
			Expect(node.IsSynchronized()).To(BeTrue())
			Expect(walk.IsSynchronized()).To(BeFalse())
			Expect(idle.IsSynchronized()).To(BeFalse())
		})
	})
})
