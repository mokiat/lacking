package shape2d_test

import (
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"github.com/mokiat/gomath/dprec"
	"github.com/mokiat/gomath/testing/dprectest"
	"github.com/mokiat/lacking/util/shape2d"
)

func unitSquarePolygon() shape2d.Polygon {
	return shape2d.NewPolygon([]shape2d.Edge{
		shape2d.NewEdge(dprec.NewVec2(-0.5, -0.5), dprec.NewVec2(0.5, -0.5)), // bottom-left
		shape2d.NewEdge(dprec.NewVec2(0.5, -0.5), dprec.NewVec2(0.5, 0.5)),   // bottom-right
		shape2d.NewEdge(dprec.NewVec2(0.5, 0.5), dprec.NewVec2(-0.5, 0.5)),   // top-right
		shape2d.NewEdge(dprec.NewVec2(-0.5, 0.5), dprec.NewVec2(-0.5, -0.5)), // top-left
	})
}

var _ = Describe("IsCirclePolygonIntersection", func() {
	var polygon shape2d.Polygon

	BeforeEach(func() {
		polygon = unitSquarePolygon()
	})

	Specify("circle fully outside returns false", func() {
		circle := shape2d.NewCircle(dprec.NewVec2(0.0, -2.0), 0.5)
		Expect(shape2d.IsCirclePolygonIntersection(circle, polygon)).To(BeFalse())
	})

	Specify("circle fully inside returns false", func() {
		circle := shape2d.NewCircle(dprec.NewVec2(0.0, 0.0), 0.1)
		Expect(shape2d.IsCirclePolygonIntersection(circle, polygon)).To(BeFalse())
	})

	Specify("circle overlapping the bottom edge returns true", func() {
		circle := shape2d.NewCircle(dprec.NewVec2(0.0, -0.8), 0.5)
		Expect(shape2d.IsCirclePolygonIntersection(circle, polygon)).To(BeTrue())
	})

	Specify("circle overlapping the right edge returns true", func() {
		circle := shape2d.NewCircle(dprec.NewVec2(0.8, 0.0), 0.5)
		Expect(shape2d.IsCirclePolygonIntersection(circle, polygon)).To(BeTrue())
	})

	Specify("empty polygon returns false", func() {
		circle := shape2d.NewCircle(dprec.NewVec2(0.0, 0.0), 1.0)
		Expect(shape2d.IsCirclePolygonIntersection(circle, shape2d.NewPolygon(nil))).To(BeFalse())
	})
})

var _ = Describe("CheckCirclePolygonIntersection", func() {
	var (
		intersections []shape2d.Intersection
		polygon       shape2d.Polygon
	)

	yield := func(i shape2d.Intersection) {
		intersections = append(intersections, i)
	}

	BeforeEach(func() {
		intersections = nil
		polygon = unitSquarePolygon()
	})

	Specify("circle fully outside produces no intersection", func() {
		circle := shape2d.NewCircle(dprec.NewVec2(0.0, -2.0), 0.5)
		shape2d.CheckCirclePolygonIntersection(circle, polygon, yield)
		Expect(intersections).To(BeEmpty())
	})

	Specify("circle fully inside produces no intersection", func() {
		circle := shape2d.NewCircle(dprec.NewVec2(0.0, 0.0), 0.1)
		shape2d.CheckCirclePolygonIntersection(circle, polygon, yield)
		Expect(intersections).To(BeEmpty())
	})

	Specify("empty polygon produces no intersection", func() {
		circle := shape2d.NewCircle(dprec.NewVec2(0.0, 0.0), 1.0)
		shape2d.CheckCirclePolygonIntersection(circle, shape2d.NewPolygon(nil), yield)
		Expect(intersections).To(BeEmpty())
	})

	Specify("circle overlapping the bottom face yields correct depth, normal and contact", func() {
		circle := shape2d.NewCircle(dprec.NewVec2(0.0, -0.8), 0.5)
		shape2d.CheckCirclePolygonIntersection(circle, polygon, yield)
		Expect(intersections).To(HaveLen(1))
		i := intersections[0]
		Expect(i.Depth).To(BeNumerically("~", 0.2, 1e-6))
		Expect(i.TargetNormal).To(dprectest.HaveVec2Coords(0.0, -1.0))
		Expect(i.TargetContact).To(dprectest.HaveVec2Coords(0.0, -0.5))
	})

	Specify("circle overlapping the right face yields correct depth, normal and contact", func() {
		circle := shape2d.NewCircle(dprec.NewVec2(0.8, 0.0), 0.5)
		shape2d.CheckCirclePolygonIntersection(circle, polygon, yield)
		Expect(intersections).To(HaveLen(1))
		i := intersections[0]
		Expect(i.Depth).To(BeNumerically("~", 0.2, 1e-6))
		Expect(i.TargetNormal).To(dprectest.HaveVec2Coords(1.0, 0.0))
		Expect(i.TargetContact).To(dprectest.HaveVec2Coords(0.5, 0.0))
	})

	Specify("only the deepest intersection is returned when multiple edges are hit", func() {
		corridor := shape2d.NewPolygon([]shape2d.Edge{
			shape2d.NewEdge(dprec.NewVec2(0.0, 0.0), dprec.NewVec2(4.0, 0.0)),
			shape2d.NewEdge(dprec.NewVec2(4.0, -0.8), dprec.NewVec2(0.0, -0.8)),
		})
		circle := shape2d.NewCircle(dprec.NewVec2(2.0, -0.2), 1.0)
		shape2d.CheckCirclePolygonIntersection(circle, corridor, yield)
		Expect(intersections).To(HaveLen(1))
		i := intersections[0]
		Expect(i.Depth).To(BeNumerically("~", 0.8, 1e-6))
		Expect(i.TargetNormal).To(dprectest.HaveVec2Coords(0.0, -1.0))
		Expect(i.TargetContact).To(dprectest.HaveVec2Coords(2.0, 0.0))
	})
})
