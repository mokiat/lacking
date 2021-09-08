package internal

import "github.com/mokiat/gomath/sprec"

const (
	initialContourPointCount      = 1024
	initialContourSubContourCount = 4
)

func newContour() *Contour {
	return &Contour{
		points:      make([]ContourPoint, 0, initialContourPointCount),
		subContours: make([]SubContour, 0, initialContourSubContourCount),
	}
}

type Contour struct {
	points      []ContourPoint
	subContours []SubContour
}

func (c *Contour) Init() {
	c.points = c.points[:0]
	c.subContours = c.subContours[:0]
}

func (c *Contour) MoveTo(position sprec.Vec2, stroke Stroke) {
	c.startSubContour()
	c.addPoint(ContourPoint{
		coords: position,
		stroke: stroke,
	})
}

func (c *Contour) LineTo(position sprec.Vec2, stroke Stroke) {
	c.addPoint(ContourPoint{
		coords: position,
		stroke: stroke,
	})
}

func (c *Contour) QuadTo(control, position sprec.Vec2, stroke Stroke) {
	// TODO: Evaluate tessellation based on curvature and size
	const tessellation = 30

	lastPoint := c.lastPoint()
	vecCS := sprec.Vec2Diff(lastPoint.coords, control)
	vecCE := sprec.Vec2Diff(position, control)

	// start and end are excluded from this loop on purpose
	for i := 1; i < tessellation; i++ {
		t := float32(i) / float32(tessellation)
		alpha := (1 - t) * (1 - t)
		beta := t * t
		c.addPoint(ContourPoint{
			coords: sprec.Vec2Sum(
				control,
				sprec.Vec2Sum(
					sprec.Vec2Prod(vecCS, alpha),
					sprec.Vec2Prod(vecCE, beta),
				),
			),
			stroke: MixStrokes(lastPoint.stroke, stroke, t),
		})
	}

	c.addPoint(ContourPoint{
		coords: position,
		stroke: stroke,
	})
}

func (c *Contour) CubeTo(control1, control2, position sprec.Vec2, stroke Stroke) {
	// TODO: Evaluate tessellation based on curvature and size
	const tessellation = 30

	lastPoint := c.lastPoint()

	// start and end are excluded from this loop on purpose
	for i := 1; i < tessellation; i++ {
		t := float32(i) / float32(tessellation)
		alpha := (1 - t) * (1 - t) * (1 - t)
		beta := 3 * (1 - t) * (1 - t) * t
		gamma := 3 * (1 - t) * t * t
		delta := t * t * t
		c.addPoint(ContourPoint{
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
			stroke: MixStrokes(lastPoint.stroke, stroke, t),
		})
	}

	c.addPoint(ContourPoint{
		coords: position,
		stroke: stroke,
	})
}

func (c *Contour) CloseLoop() {
	lastSubContour := c.subContours[len(c.subContours)-1]
	c.addPoint(c.points[lastSubContour.pointOffset])
}

func (c *Contour) startSubContour() {
	c.subContours = append(c.subContours, SubContour{
		pointOffset: len(c.points),
		pointCount:  0,
	})
}

func (c *Contour) addPoint(point ContourPoint) {
	c.points = append(c.points, point)
	c.subContours[len(c.subContours)-1].pointCount++
}

func (c *Contour) lastPoint() ContourPoint {
	return c.points[len(c.points)-1]
}

type ContourPoint struct {
	coords sprec.Vec2
	stroke Stroke
}

type SubContour struct {
	pointOffset int
	pointCount  int
}
