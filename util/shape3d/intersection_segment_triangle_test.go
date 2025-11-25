package shape3d_test

import (
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"github.com/mokiat/gomath/dprec"
	"github.com/mokiat/gomath/testing/dprectest"
	"github.com/mokiat/lacking/util/shape3d"
)

var _ = Describe("IntersectionSegmentTriangle", func() {

	Specify("handles intersection correctly", func() {
		segment := shape3d.NewSegment(
			dprec.NewVec3(-0.1, 10.0, -0.1),
			dprec.NewVec3(0.1, -10.0, 0.1),
		)
		triangle := shape3d.NewTriangle(
			dprec.NewVec3(0.0, 0.0, -1.0),
			dprec.NewVec3(-1.5, 0.0, 1.0),
			dprec.NewVec3(1.5, 0.0, 1.0),
		)
		var collection shape3d.SmallestIntersection
		shape3d.CheckSegmentTriangleIntersection(segment, triangle, collection.AddIntersection)
		intersection, ok := collection.Intersection()
		Expect(ok).To(BeTrue())
		Expect(intersection.TargetContact).To(dprectest.HaveVec3Coords(0.0, 0.0, 0.0))
		Expect(intersection.TargetNormal).To(dprectest.HaveVec3Coords(0.0, 1.0, 0.0))
		Expect(intersection.Depth).To(dprectest.EqualFloat64(10.0))
	})

	Specify("handles rotated intersection correctly", func() {
		segment := shape3d.NewSegment(
			dprec.NewVec3(-5.0, 0.6, -0.1),
			dprec.NewVec3(5.0, 0.4, 0.1),
		)
		triangle := shape3d.NewTriangle(
			dprec.NewVec3(0.0, 1.0, 0.0),
			dprec.NewVec3(0.0, -1.0, -1.0),
			dprec.NewVec3(0.0, -1.0, 1.0),
		)
		var collection shape3d.SmallestIntersection
		shape3d.CheckSegmentTriangleIntersection(segment, triangle, collection.AddIntersection)
		intersection, ok := collection.Intersection()
		Expect(ok).To(BeTrue())
		Expect(intersection.TargetContact).To(dprectest.HaveVec3Coords(0.0, 0.5, 0.0))
		Expect(intersection.TargetNormal).To(dprectest.HaveVec3Coords(-1.0, 0.0, 0.0))
		Expect(intersection.Depth).To(dprectest.EqualFloat64(5.0))
	})

	Specify("handles non-intersections correctly", func() {
		segment := shape3d.NewSegment(
			dprec.NewVec3(-0.1, 10.0, 1.0),
			dprec.NewVec3(0.1, -10.0, 1.2),
		)
		triangle := shape3d.NewTriangle(
			dprec.NewVec3(0.0, 0.0, -1.0),
			dprec.NewVec3(-1.5, 0.0, 1.0),
			dprec.NewVec3(1.5, 0.0, 1.0),
		)
		var collection shape3d.SmallestIntersection
		shape3d.CheckSegmentTriangleIntersection(segment, triangle, collection.AddIntersection)
		_, ok := collection.Intersection()
		Expect(ok).To(BeFalse())
	})

})
