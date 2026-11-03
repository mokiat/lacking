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

	BeforeEach(func() {
		skeleton = animation3.NewSkeleton([]animation3.SkeletonBone{
			{Name: "bone", RestTransform: animation3.IdentityTransform()},
		})
		idle = cycleNode(4.0, dprec.BasisYVec3())
		walk = cycleNode(2.0, dprec.BasisXVec3())
		node = animation3.NewStateMachineNode(walkState, map[state]animation3.Node{
			idleState: idle,
			walkState: walk,
		})
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
				})
			}).To(Panic())
		})

		It("panics without states", func() {
			Expect(func() {
				animation3.NewStateMachineNode(idleState, nil)
			}).To(Panic())
		})

		It("does not retain the specified map", func() {
			states := map[state]animation3.Node{
				idleState: idle,
			}
			node := animation3.NewStateMachineNode(idleState, states)
			states[idleState] = walk
			node.Advance(1.0)
			Expect(idle.Fraction()).To(BeNumerically("~", 0.25, 1e-9))
			Expect(walk.Fraction()).To(Equal(0.0))
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
