package animation_test

import (
	"github.com/mokiat/gomath/dprec"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"github.com/mokiat/lacking/game/animation"
)

var _ = Describe("Blend1DNode", func() {

	const bone = 0

	var skeleton *animation.Skeleton

	// positionNode creates a static node that places the bone at the
	// specified X coordinate.
	positionNode := func(x float64) *animation.FixedPoseNode {
		return animation.NewFixedPoseNode(skeleton, map[string]animation.Transform{
			"bone": {
				Translation: dprec.NewVec3(x, 0.0, 0.0),
				Rotation:    dprec.IdentityQuat(),
				Scale:       dprec.NewVec3(1.0, 1.0, 1.0),
			},
		})
	}

	// cycleNode creates a looping clip node of the specified length that
	// moves the bone along the specified axis by one unit per second.
	cycleNode := func(length float64, axis dprec.Vec3) *animation.ClipNode {
		clip := animation.NewClip("cycle", 0.0, length, map[string]animation.Channel{
			"bone": animation.NewComponentKeyframeChannel(
				animation.KeyframeList[dprec.Vec3]{
					{Timestamp: 0.0, Value: dprec.ZeroVec3()},
					{Timestamp: length, Value: dprec.Vec3Prod(axis, length)},
				},
				animation.KeyframeList[dprec.Quat]{
					{Timestamp: 0.0, Value: dprec.IdentityQuat()},
				},
				animation.KeyframeList[dprec.Vec3]{
					{Timestamp: 0.0, Value: dprec.NewVec3(1.0, 1.0, 1.0)},
				},
			),
		})
		return animation.NewClipNode(skeleton, clip, true)
	}

	expectX := func(node animation.Node, expected float64) {
		GinkgoHelper()
		expectVec3(node.Transform(bone).Translation, dprec.NewVec3(expected, 0.0, 0.0))
	}

	BeforeEach(func() {
		skeleton = animation.NewSkeleton([]animation.SkeletonBone{
			{Name: "bone", RestTransform: animation.IdentityTransform()},
		})
	})

	Describe("NewBlend1DNode", func() {

		It("panics without entries", func() {
			Expect(func() {
				animation.NewBlend1DNode()
			}).To(Panic())
		})

		It("starts at the coordinate of the first entry along the line", func() {
			node := animation.NewBlend1DNode(
				animation.Blend1DEntry{Coord: 2.0, Node: positionNode(2.0)},
				animation.Blend1DEntry{Coord: -1.0, Node: positionNode(-1.0)},
			)
			Expect(node.BlendCoord()).To(Equal(-1.0))
			expectX(node, -1.0)
		})

		It("starts at a fraction of 0.0 and is not synchronized", func() {
			node := animation.NewBlend1DNode(
				animation.Blend1DEntry{Coord: 0.0, Node: positionNode(0.0)},
			)
			Expect(node.Fraction()).To(Equal(0.0))
			Expect(node.IsSynchronized()).To(BeFalse())
		})

		It("does not retain the specified slice", func() {
			entries := []animation.Blend1DEntry{
				{Coord: 0.0, Node: positionNode(0.0)},
				{Coord: 1.0, Node: positionNode(1.0)},
			}
			node := animation.NewBlend1DNode(entries...)
			entries[1] = animation.Blend1DEntry{Coord: 1.0, Node: positionNode(5.0)}
			node.SetBlendCoord(1.0)
			expectX(node, 1.0)
		})
	})

	Describe("SetBlendCoord", func() {

		var node *animation.Blend1DNode

		BeforeEach(func() {
			// Each entry places the bone at its own coordinate, so the X
			// translation equals the coordinate within the range of entries.
			// The entries are listed out of order on purpose.
			node = animation.NewBlend1DNode(
				animation.Blend1DEntry{Coord: 3.0, Node: positionNode(3.0)},
				animation.Blend1DEntry{Coord: 0.0, Node: positionNode(0.0)},
				animation.Blend1DEntry{Coord: 1.0, Node: positionNode(1.0)},
			)
		})

		It("changes the blend coordinate", func() {
			node.SetBlendCoord(2.5)
			Expect(node.BlendCoord()).To(Equal(2.5))
		})

		DescribeTable("blends the entries that surround the coordinate",
			func(coord, expectedX float64) {
				node.SetBlendCoord(coord)
				expectX(node, expectedX)
			},
			Entry("at the first entry", 0.0, 0.0),
			Entry("between the first two entries", 0.25, 0.25),
			Entry("at a middle entry", 1.0, 1.0),
			Entry("between the last two entries", 2.5, 2.5),
			Entry("at the last entry", 3.0, 3.0),
			Entry("before the first entry", -5.0, 0.0),
			Entry("after the last entry", 7.0, 3.0),
		)

		It("blends rotation and scale as well", func() {
			node := animation.NewBlend1DNode(
				animation.Blend1DEntry{Coord: 0.0, Node: positionNode(0.0)},
				animation.Blend1DEntry{Coord: 2.0, Node: animation.NewFixedPoseNode(skeleton, map[string]animation.Transform{
					"bone": {
						Translation: dprec.ZeroVec3(),
						Rotation:    yRotation(90.0),
						Scale:       dprec.NewVec3(3.0, 3.0, 3.0),
					},
				})},
			)
			node.SetBlendCoord(1.0)
			transform := node.Transform(bone)
			expectRotation(transform.Rotation, yRotation(45.0))
			expectVec3(transform.Scale, dprec.NewVec3(2.0, 2.0, 2.0))
		})

		It("switches between entries that share a coordinate", func() {
			node := animation.NewBlend1DNode(
				animation.Blend1DEntry{Coord: 0.0, Node: positionNode(0.0)},
				animation.Blend1DEntry{Coord: 1.0, Node: positionNode(1.0)},
				animation.Blend1DEntry{Coord: 1.0, Node: positionNode(5.0)},
				animation.Blend1DEntry{Coord: 2.0, Node: positionNode(6.0)},
			)
			node.SetBlendCoord(0.5)
			expectX(node, 0.5)
			node.SetBlendCoord(1.0)
			expectX(node, 5.0)
			node.SetBlendCoord(1.5)
			expectX(node, 5.5)
		})

		It("shows a synchronized entry that becomes active in phase with the node", func() {
			first := animation.Synchronized(cycleNode(1.0, dprec.BasisXVec3()))
			second := animation.Synchronized(cycleNode(2.0, dprec.BasisXVec3()))
			third := animation.Synchronized(cycleNode(4.0, dprec.BasisXVec3()))
			node := animation.NewBlend1DNode(
				animation.Blend1DEntry{Coord: 0.0, Node: first},
				animation.Blend1DEntry{Coord: 1.0, Node: second},
				animation.Blend1DEntry{Coord: 2.0, Node: third},
			)
			node.Advance(0.25)
			Expect(node.Fraction()).To(BeNumerically("~", 0.25, 1e-9))

			node.SetBlendCoord(2.0)
			Expect(third.Fraction()).To(BeNumerically("~", 0.25, 1e-9))
			expectX(node, 1.0) // a quarter of the 4.0 second cycle
		})
	})

	Describe("Rate", func() {

		var node *animation.Blend1DNode

		BeforeEach(func() {
			node = animation.NewBlend1DNode(
				animation.Blend1DEntry{Coord: 0.0, Node: animation.Synchronized(cycleNode(1.0, dprec.BasisXVec3()))},
				animation.Blend1DEntry{Coord: 1.0, Node: animation.Synchronized(cycleNode(2.0, dprec.BasisXVec3()))},
				animation.Blend1DEntry{Coord: 2.0, Node: animation.Synchronized(cycleNode(4.0, dprec.BasisXVec3()))},
			)
		})

		DescribeTable("blends the cycle lengths of the active entries",
			func(coord, expectedLength float64) {
				node.SetBlendCoord(coord)
				Expect(node.Rate()).To(BeNumerically("~", 1.0/expectedLength, 1e-9))
			},
			Entry("at the first entry", 0.0, 1.0),
			Entry("between the first two entries", 0.5, 1.5),
			Entry("between the last two entries", 1.5, 3.0),
			Entry("after the last entry", 5.0, 4.0),
		)

		It("is 0.0 when no active entry is synchronized", func() {
			node := animation.NewBlend1DNode(
				animation.Blend1DEntry{Coord: 0.0, Node: cycleNode(1.0, dprec.BasisXVec3())},
				animation.Blend1DEntry{Coord: 1.0, Node: cycleNode(2.0, dprec.BasisXVec3())},
			)
			node.SetBlendCoord(0.5)
			Expect(node.Rate()).To(Equal(0.0))
		})
	})

	Describe("SetFraction", func() {

		It("wraps the fraction", func() {
			node := animation.NewBlend1DNode(
				animation.Blend1DEntry{Coord: 0.0, Node: cycleNode(1.0, dprec.BasisXVec3())},
			)
			node.SetFraction(1.25)
			Expect(node.Fraction()).To(BeNumerically("~", 0.25, 1e-9))
			node.SetFraction(-0.25)
			Expect(node.Fraction()).To(BeNumerically("~", 0.75, 1e-9))
		})

		It("relocates all synchronized entries and leaves the others in place", func() {
			active := animation.Synchronized(cycleNode(1.0, dprec.BasisXVec3()))
			inactive := animation.Synchronized(cycleNode(2.0, dprec.BasisXVec3()))
			free := cycleNode(4.0, dprec.BasisXVec3())
			node := animation.NewBlend1DNode(
				animation.Blend1DEntry{Coord: 0.0, Node: active},
				animation.Blend1DEntry{Coord: 1.0, Node: inactive},
				animation.Blend1DEntry{Coord: 2.0, Node: free},
			)
			node.SetFraction(0.5)
			Expect(active.Fraction()).To(BeNumerically("~", 0.5, 1e-9))
			Expect(inactive.Fraction()).To(BeNumerically("~", 0.5, 1e-9))
			Expect(free.Fraction()).To(Equal(0.0))
		})
	})

	Describe("Advance", func() {

		It("keeps the synchronized active entries in phase", func() {
			first := animation.Synchronized(cycleNode(1.0, dprec.BasisXVec3()))
			second := animation.Synchronized(cycleNode(2.0, dprec.BasisXVec3()))
			node := animation.NewBlend1DNode(
				animation.Blend1DEntry{Coord: 0.0, Node: first},
				animation.Blend1DEntry{Coord: 1.0, Node: second},
			)
			node.SetBlendCoord(0.5)
			node.Advance(0.75) // half of the blended cycle of 1.5 seconds
			Expect(node.Fraction()).To(BeNumerically("~", 0.5, 1e-9))
			Expect(first.Fraction()).To(BeNumerically("~", 0.5, 1e-9))
			Expect(second.Fraction()).To(BeNumerically("~", 0.5, 1e-9))
		})

		It("lets an active entry that is not synchronized play at its own pace", func() {
			first := animation.Synchronized(cycleNode(1.0, dprec.BasisXVec3()))
			second := cycleNode(4.0, dprec.BasisXVec3())
			node := animation.NewBlend1DNode(
				animation.Blend1DEntry{Coord: 0.0, Node: first},
				animation.Blend1DEntry{Coord: 1.0, Node: second},
			)
			node.SetBlendCoord(0.5)
			node.Advance(0.5)
			Expect(node.Fraction()).To(BeNumerically("~", 0.5, 1e-9))
			Expect(second.Fraction()).To(BeNumerically("~", 0.125, 1e-9))
		})

		It("advances inactive entries as well", func() {
			active := animation.Synchronized(cycleNode(1.0, dprec.BasisXVec3()))
			inactive := animation.Synchronized(cycleNode(2.0, dprec.BasisXVec3()))
			free := cycleNode(4.0, dprec.BasisXVec3())
			node := animation.NewBlend1DNode(
				animation.Blend1DEntry{Coord: 0.0, Node: active},
				animation.Blend1DEntry{Coord: 1.0, Node: inactive},
				animation.Blend1DEntry{Coord: 2.0, Node: free},
			)
			node.Advance(0.5)
			Expect(node.Fraction()).To(BeNumerically("~", 0.5, 1e-9))
			Expect(active.Fraction()).To(BeNumerically("~", 0.5, 1e-9))
			Expect(inactive.Fraction()).To(BeNumerically("~", 0.5, 1e-9))
			Expect(free.Fraction()).To(BeNumerically("~", 0.125, 1e-9))
		})

		It("advances each entry only once", func() {
			first := cycleNode(4.0, dprec.BasisXVec3())
			second := cycleNode(8.0, dprec.BasisXVec3())
			node := animation.NewBlend1DNode(
				animation.Blend1DEntry{Coord: 0.0, Node: first},
				animation.Blend1DEntry{Coord: 1.0, Node: second},
			)
			node.SetBlendCoord(0.5)
			node.Advance(1.0)
			Expect(first.Fraction()).To(BeNumerically("~", 0.25, 1e-9))
			Expect(second.Fraction()).To(BeNumerically("~", 0.125, 1e-9))
		})

		It("wraps around at the end of the cycle", func() {
			node := animation.NewBlend1DNode(
				animation.Blend1DEntry{Coord: 0.0, Node: animation.Synchronized(cycleNode(2.0, dprec.BasisXVec3()))},
			)
			node.Advance(2.5)
			Expect(node.Fraction()).To(BeNumerically("~", 0.25, 1e-9))
		})

		It("keeps its fraction when no active entry is synchronized", func() {
			node := animation.NewBlend1DNode(
				animation.Blend1DEntry{Coord: 0.0, Node: cycleNode(2.0, dprec.BasisXVec3())},
			)
			node.SetFraction(0.25)
			node.Advance(0.5)
			Expect(node.Fraction()).To(BeNumerically("~", 0.25, 1e-9))
		})
	})

	Describe("DeltaTransform", func() {

		DescribeTable("blends the motion of the active entries",
			func(coord float64, expected dprec.Vec3) {
				node := animation.NewBlend1DNode(
					animation.Blend1DEntry{Coord: 0.0, Node: cycleNode(2.0, dprec.BasisXVec3())},
					animation.Blend1DEntry{Coord: 1.0, Node: cycleNode(2.0, dprec.BasisYVec3())},
				)
				node.SetBlendCoord(coord)
				delta := node.DeltaTransform(bone, 0.5)
				expectVec3(delta.Translation, expected)
				expectRotation(delta.Rotation, dprec.IdentityQuat())
				expectVec3(delta.Scale, dprec.NewVec3(1.0, 1.0, 1.0))
			},
			Entry("first entry", 0.0, dprec.NewVec3(0.5, 0.0, 0.0)),
			Entry("second entry", 1.0, dprec.NewVec3(0.0, 0.5, 0.0)),
			Entry("halfway", 0.5, dprec.NewVec3(0.25, 0.25, 0.0)),
		)

		It("does not change the playback position", func() {
			entry := animation.Synchronized(cycleNode(2.0, dprec.BasisXVec3()))
			node := animation.NewBlend1DNode(
				animation.Blend1DEntry{Coord: 0.0, Node: entry},
			)
			node.SetFraction(0.25)
			node.DeltaTransform(bone, 1.0)
			Expect(node.Fraction()).To(BeNumerically("~", 0.25, 1e-9))
			Expect(entry.Fraction()).To(BeNumerically("~", 0.25, 1e-9))
		})
	})

	Describe("SetSynchronized", func() {

		It("configures the node without affecting its entries", func() {
			entry := cycleNode(1.0, dprec.BasisXVec3())
			node := animation.NewBlend1DNode(
				animation.Blend1DEntry{Coord: 0.0, Node: entry},
			)
			node.SetSynchronized(true)
			Expect(node.IsSynchronized()).To(BeTrue())
			Expect(entry.IsSynchronized()).To(BeFalse())
			node.SetSynchronized(false)
			Expect(node.IsSynchronized()).To(BeFalse())
		})
	})
})
