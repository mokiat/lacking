package internal

import "github.com/mokiat/gomath/sprec"

const (
	initialShapePointCount    = 1024
	initialShapeSubShapeCount = 4
)

func newShape() *Shape {
	return &Shape{
		points:    make([]ShapePoint, 0, initialShapePointCount),
		subShapes: make([]SubShape, 0, initialShapeSubShapeCount),
	}
}

type Shape struct {
	fill      Fill
	points    []ShapePoint
	subShapes []SubShape
}

func (s *Shape) Init(fill Fill) {
	s.fill = fill
	s.points = s.points[:0]
	s.subShapes = s.subShapes[:0]
}

func (s *Shape) MoveTo(position sprec.Vec2) {
	s.startSubShape()
	s.addPoint(ShapePoint{
		coords: position,
	})
}

func (s *Shape) LineTo(position sprec.Vec2) {
	s.addPoint(ShapePoint{
		coords: position,
	})
}

func (s *Shape) QuadTo(control, position sprec.Vec2) {
	// TODO: Evaluate tessellation based on curvature and size
	const tessellation = 30

	lastPoint := s.lastPoint()
	vecCS := sprec.Vec2Diff(lastPoint.coords, control)
	vecCE := sprec.Vec2Diff(position, control)

	// start and end are excluded from this loop on purpose
	for i := 1; i < tessellation; i++ {
		t := float32(i) / float32(tessellation)
		alpha := (1 - t) * (1 - t)
		beta := t * t
		s.addPoint(ShapePoint{
			coords: sprec.Vec2Sum(
				control,
				sprec.Vec2Sum(
					sprec.Vec2Prod(vecCS, alpha),
					sprec.Vec2Prod(vecCE, beta),
				),
			),
		})
	}

	s.addPoint(ShapePoint{
		coords: position,
	})
}

func (s *Shape) CubeTo(control1, control2, position sprec.Vec2) {
	// TODO: Evaluate tessellation based on curvature and size
	const tessellation = 30

	lastPoint := s.lastPoint()

	// start and end are excluded from this loop on purpose
	for i := 1; i < tessellation; i++ {
		t := float32(i) / float32(tessellation)
		alpha := (1 - t) * (1 - t) * (1 - t)
		beta := 3 * (1 - t) * (1 - t) * t
		gamma := 3 * (1 - t) * t * t
		delta := t * t * t
		s.addPoint(ShapePoint{
			coords: sprec.Vec2Sum(
				sprec.Vec2Sum(
					sprec.Vec2Prod(lastPoint.coords, alpha),
					sprec.Vec2Prod(control1, beta),
				),
				sprec.Vec2Sum(
					sprec.Vec2Prod(control2, gamma),
					sprec.Vec2Prod(position, delta),
				),
			),
		})
	}

	s.addPoint(ShapePoint{
		coords: position,
	})
}

func (s *Shape) startSubShape() {
	s.subShapes = append(s.subShapes, SubShape{
		pointOffset: len(s.points),
		pointCount:  0,
	})
}

func (s *Shape) addPoint(point ShapePoint) {
	s.points = append(s.points, point)
	s.subShapes[len(s.subShapes)-1].pointCount++
}

func (s *Shape) lastPoint() ShapePoint {
	return s.points[len(s.points)-1]
}

type ShapePoint struct {
	coords sprec.Vec2
}

type SubShape struct {
	pointOffset int
	pointCount  int
}
