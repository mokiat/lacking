package shape2d_test

import (
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"github.com/mokiat/gomath/dprec"
	"github.com/mokiat/gomath/testing/dprectest"
	"github.com/mokiat/lacking/util/shape2d"
)

var _ = Describe("IsCircleCircleIntersection", func() {

	Specify("circles with a gap return false", func() {
		source := shape2d.NewCircle(dprec.NewVec2(0.0, 0.0), 2.0)
		target := shape2d.NewCircle(dprec.NewVec2(5.0, 0.0), 2.0)
		Expect(shape2d.IsCircleCircleIntersection(source, target)).To(BeFalse())
	})

	Specify("circles touching exactly return true", func() {
		source := shape2d.NewCircle(dprec.NewVec2(0.0, 0.0), 2.0)
		target := shape2d.NewCircle(dprec.NewVec2(4.0, 0.0), 2.0)
		Expect(shape2d.IsCircleCircleIntersection(source, target)).To(BeTrue())
	})

	Specify("overlapping circles return true", func() {
		source := shape2d.NewCircle(dprec.NewVec2(0.0, 0.0), 2.0)
		target := shape2d.NewCircle(dprec.NewVec2(3.0, 0.0), 2.0)
		Expect(shape2d.IsCircleCircleIntersection(source, target)).To(BeTrue())
	})

	Specify("target circle fully inside source returns true", func() {
		source := shape2d.NewCircle(dprec.NewVec2(0.0, 0.0), 5.0)
		target := shape2d.NewCircle(dprec.NewVec2(1.0, 0.0), 1.0)
		Expect(shape2d.IsCircleCircleIntersection(source, target)).To(BeTrue())
	})

	Specify("concentric circles return true", func() {
		source := shape2d.NewCircle(dprec.NewVec2(0.0, 0.0), 2.0)
		target := shape2d.NewCircle(dprec.NewVec2(0.0, 0.0), 3.0)
		Expect(shape2d.IsCircleCircleIntersection(source, target)).To(BeTrue())
	})
})

var _ = Describe("CheckCircleCircleIntersection", func() {
	var intersections []shape2d.Intersection

	BeforeEach(func() {
		intersections = nil
	})

	yield := func(i shape2d.Intersection) {
		intersections = append(intersections, i)
	}

	Specify("circles with a gap produce no intersection", func() {
		source := shape2d.NewCircle(dprec.NewVec2(0.0, 0.0), 2.0)
		target := shape2d.NewCircle(dprec.NewVec2(5.0, 0.0), 2.0)
		shape2d.CheckCircleCircleIntersection(source, target, yield)
		Expect(intersections).To(BeEmpty())
	})

	Specify("touching circles produce no intersection", func() {
		// overlap == 0 is treated as no intersection
		source := shape2d.NewCircle(dprec.NewVec2(0.0, 0.0), 2.0)
		target := shape2d.NewCircle(dprec.NewVec2(4.0, 0.0), 2.0)
		shape2d.CheckCircleCircleIntersection(source, target, yield)
		Expect(intersections).To(BeEmpty())
	})

	Specify("overlapping circles along +X yield correct depth, normal and contact", func() {
		// source at origin r=2, target at (3,0) r=2: overlap = 2+2-3 = 1
		source := shape2d.NewCircle(dprec.NewVec2(0.0, 0.0), 2.0)
		target := shape2d.NewCircle(dprec.NewVec2(3.0, 0.0), 2.0)
		shape2d.CheckCircleCircleIntersection(source, target, yield)
		Expect(intersections).To(HaveLen(1))
		i := intersections[0]
		Expect(i.Depth).To(BeNumerically("~", 1.0, 1e-6))
		// normal points from target toward source (direction to push source out of target)
		Expect(i.TargetNormal).To(dprectest.HaveVec2Coords(-1.0, 0.0))
		// contact is on the target's boundary facing the source
		Expect(i.TargetContact).To(dprectest.HaveVec2Coords(1.0, 0.0))
	})

	Specify("overlapping circles along +Y yield correct depth, normal and contact", func() {
		// source at origin r=2, target at (0,3) r=2: overlap = 1
		source := shape2d.NewCircle(dprec.NewVec2(0.0, 0.0), 2.0)
		target := shape2d.NewCircle(dprec.NewVec2(0.0, 3.0), 2.0)
		shape2d.CheckCircleCircleIntersection(source, target, yield)
		Expect(intersections).To(HaveLen(1))
		i := intersections[0]
		Expect(i.Depth).To(BeNumerically("~", 1.0, 1e-6))
		Expect(i.TargetNormal).To(dprectest.HaveVec2Coords(0.0, -1.0))
		Expect(i.TargetContact).To(dprectest.HaveVec2Coords(0.0, 1.0))
	})

	Specify("target circle fully inside source yields correct depth and normal", func() {
		// source at origin r=5, target at (2,0) r=1: overlap = 5+1-2 = 4
		source := shape2d.NewCircle(dprec.NewVec2(0.0, 0.0), 5.0)
		target := shape2d.NewCircle(dprec.NewVec2(2.0, 0.0), 1.0)
		shape2d.CheckCircleCircleIntersection(source, target, yield)
		Expect(intersections).To(HaveLen(1))
		i := intersections[0]
		Expect(i.Depth).To(BeNumerically("~", 4.0, 1e-6))
		Expect(i.TargetNormal).To(dprectest.HaveVec2Coords(-1.0, 0.0))
		Expect(i.TargetContact).To(dprectest.HaveVec2Coords(1.0, 0.0))
	})

	Specify("swapping source and target flips normal and contact", func() {
		source := shape2d.NewCircle(dprec.NewVec2(0.0, 0.0), 2.0)
		target := shape2d.NewCircle(dprec.NewVec2(3.0, 0.0), 2.0)

		var fwd, rev []shape2d.Intersection
		shape2d.CheckCircleCircleIntersection(source, target, func(i shape2d.Intersection) {
			fwd = append(fwd, i)
		})
		shape2d.CheckCircleCircleIntersection(target, source, func(i shape2d.Intersection) {
			rev = append(rev, i)
		})

		Expect(fwd).To(HaveLen(1))
		Expect(rev).To(HaveLen(1))
		Expect(fwd[0].Depth).To(BeNumerically("~", rev[0].Depth, 1e-6))
		// normals should be opposite
		Expect(fwd[0].TargetNormal).To(dprectest.HaveVec2Coords(
			-rev[0].TargetNormal.X,
			-rev[0].TargetNormal.Y,
		))
	})

	Specify("concentric circles yield an intersection with a valid unit normal", func() {
		source := shape2d.NewCircle(dprec.NewVec2(0.0, 0.0), 2.0)
		target := shape2d.NewCircle(dprec.NewVec2(0.0, 0.0), 3.0)
		shape2d.CheckCircleCircleIntersection(source, target, yield)
		Expect(intersections).To(HaveLen(1))
		i := intersections[0]
		Expect(i.Depth).To(BeNumerically("~", 5.0, 1e-6))
		Expect(i.TargetNormal).To(dprectest.HaveVec2Coords(-1.0, 0.0))
	})
})
