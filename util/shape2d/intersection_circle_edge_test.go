package shape2d_test

import (
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"github.com/mokiat/gomath/dprec"
	"github.com/mokiat/gomath/testing/dprectest"
	"github.com/mokiat/lacking/util/shape2d"
)

var _ = Describe("IsCircleEdgeIntersection", func() {
	var edge shape2d.Edge

	BeforeEach(func() {
		edge = shape2d.NewEdge(dprec.NewVec2(0.0, 0.0), dprec.NewVec2(4.0, 0.0))
	})

	Specify("circle overlapping the face returns true", func() {
		circle := shape2d.NewCircle(dprec.NewVec2(2.0, -0.5), 1.0)
		Expect(shape2d.IsCircleEdgeIntersection(circle, edge)).To(BeTrue())
	})

	Specify("circle too far from the face returns false", func() {
		circle := shape2d.NewCircle(dprec.NewVec2(2.0, -2.0), 1.0)
		Expect(shape2d.IsCircleEdgeIntersection(circle, edge)).To(BeFalse())
	})

	Specify("circle on the inactive side of the edge returns false", func() {
		circle := shape2d.NewCircle(dprec.NewVec2(2.0, 0.5), 1.0)
		Expect(shape2d.IsCircleEdgeIntersection(circle, edge)).To(BeFalse())
	})

	Specify("circle overlapping corner A returns true", func() {
		circle := shape2d.NewCircle(dprec.NewVec2(-0.5, -0.5), 1.0)
		Expect(shape2d.IsCircleEdgeIntersection(circle, edge)).To(BeTrue())
	})

	Specify("circle outside corner A returns false", func() {
		circle := shape2d.NewCircle(dprec.NewVec2(-2.0, -0.5), 1.0)
		Expect(shape2d.IsCircleEdgeIntersection(circle, edge)).To(BeFalse())
	})

	Specify("circle overlapping corner B returns true", func() {
		circle := shape2d.NewCircle(dprec.NewVec2(4.5, -0.5), 1.0)
		Expect(shape2d.IsCircleEdgeIntersection(circle, edge)).To(BeTrue())
	})

	Specify("circle outside corner B returns false", func() {
		circle := shape2d.NewCircle(dprec.NewVec2(6.0, -0.5), 1.0)
		Expect(shape2d.IsCircleEdgeIntersection(circle, edge)).To(BeFalse())
	})

	Specify("circle exactly on corner A (below guard threshold) returns false", func() {
		circle := shape2d.NewCircle(dprec.NewVec2(-0.000001, -0.000001), 1.0)
		Expect(shape2d.IsCircleEdgeIntersection(circle, edge)).To(BeFalse())
	})
})

var _ = Describe("CheckCircleEdgeIntersection", func() {
	var (
		intersections []shape2d.Intersection
		edge          shape2d.Edge
	)

	yield := func(i shape2d.Intersection) {
		intersections = append(intersections, i)
	}

	BeforeEach(func() {
		intersections = nil
		edge = shape2d.NewEdge(dprec.NewVec2(0.0, 0.0), dprec.NewVec2(4.0, 0.0))
	})

	Specify("circle overlapping face yields correct depth, normal and contact", func() {
		circle := shape2d.NewCircle(dprec.NewVec2(2.0, -0.5), 1.0)
		shape2d.CheckCircleEdgeIntersection(circle, edge, yield)
		Expect(intersections).To(HaveLen(1))
		i := intersections[0]
		Expect(i.Depth).To(BeNumerically("~", 0.5, 1e-6))
		Expect(i.TargetNormal).To(dprectest.HaveVec2Coords(0.0, -1.0))
		Expect(i.TargetContact).To(dprectest.HaveVec2Coords(2.0, 0.0))
	})

	Specify("circle too far from face produces no intersection", func() {
		circle := shape2d.NewCircle(dprec.NewVec2(2.0, -2.0), 1.0)
		shape2d.CheckCircleEdgeIntersection(circle, edge, yield)
		Expect(intersections).To(BeEmpty())
	})

	Specify("circle on inactive side of edge produces no intersection", func() {
		circle := shape2d.NewCircle(dprec.NewVec2(2.0, 0.5), 1.0)
		shape2d.CheckCircleEdgeIntersection(circle, edge, yield)
		Expect(intersections).To(BeEmpty())
	})

	Specify("circle overlapping corner A yields correct depth, normal and contact", func() {
		circle := shape2d.NewCircle(dprec.NewVec2(-0.5, -0.5), 1.0)
		shape2d.CheckCircleEdgeIntersection(circle, edge, yield)
		Expect(intersections).To(HaveLen(1))
		i := intersections[0]
		distance := dprec.NewVec2(-0.5, -0.5).Length()
		Expect(i.Depth).To(BeNumerically("~", 1.0-distance, 1e-6))
		Expect(i.TargetContact).To(dprectest.HaveVec2Coords(0.0, 0.0))
		Expect(i.TargetNormal.X).To(BeNumerically("<", 0.0))
		Expect(i.TargetNormal.Y).To(BeNumerically("<", 0.0))
	})

	Specify("circle outside corner A produces no intersection", func() {
		circle := shape2d.NewCircle(dprec.NewVec2(-2.0, -0.5), 1.0)
		shape2d.CheckCircleEdgeIntersection(circle, edge, yield)
		Expect(intersections).To(BeEmpty())
	})

	Specify("circle overlapping corner B yields correct depth, normal and contact", func() {
		circle := shape2d.NewCircle(dprec.NewVec2(4.5, -0.5), 1.0)
		shape2d.CheckCircleEdgeIntersection(circle, edge, yield)
		Expect(intersections).To(HaveLen(1))
		i := intersections[0]
		distance := dprec.NewVec2(0.5, -0.5).Length()
		Expect(i.Depth).To(BeNumerically("~", 1.0-distance, 1e-6))
		Expect(i.TargetContact).To(dprectest.HaveVec2Coords(4.0, 0.0))
		Expect(i.TargetNormal.X).To(BeNumerically(">", 0.0))
		Expect(i.TargetNormal.Y).To(BeNumerically("<", 0.0))
	})

	Specify("circle outside corner B produces no intersection", func() {
		circle := shape2d.NewCircle(dprec.NewVec2(6.0, -0.5), 1.0)
		shape2d.CheckCircleEdgeIntersection(circle, edge, yield)
		Expect(intersections).To(BeEmpty())
	})

	Specify("circle exactly on corner A triggers division-by-zero guard and produces no intersection", func() {
		circle := shape2d.NewCircle(dprec.NewVec2(-0.000001, -0.000001), 1.0)
		shape2d.CheckCircleEdgeIntersection(circle, edge, yield)
		Expect(intersections).To(BeEmpty())
	})
})
