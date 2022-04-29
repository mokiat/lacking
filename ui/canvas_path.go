package ui

import "github.com/mokiat/gomath/sprec"

type canvasPoint struct {
	coords    sprec.Vec2
	innerSize float32
	outerSize float32
	color     Color
}

func newCanvasPath() *canvasPath {
	return &canvasPath{
		pointTemplate: canvasPoint{
			innerSize: 0.5,
			outerSize: 0.5,
			color:     White(),
		},
	}
}

type canvasPath struct {
	points         []canvasPoint
	subPathOffsets []int
	pointTemplate  canvasPoint
}

// Reset erases all path points and starts from the beginning.
func (p *canvasPath) Reset() {
	p.points = p.points[:0]
	p.subPathOffsets = p.subPathOffsets[:0]
}

// MoveTo starts a new sub-path from the specified position.
func (p *canvasPath) MoveTo(position sprec.Vec2) {
	p.startSubPath()
	p.pointTemplate.coords = position
	p.addPoint()
}

func (p *canvasPath) LineTo(position sprec.Vec2) {
	p.pointTemplate.coords = position
	p.addPoint()
}

// QuadTo creates a quadratic Bezier curve from the last path
// position to the newly specified position by going past the
// specified control point.
func (p *canvasPath) QuadTo(control, position sprec.Vec2) {
	// TODO: Evaluate tessellation based on curvature and size
	const tessellation = 5

	lastPoint := p.pointTemplate
	vecCS := sprec.Vec2Diff(lastPoint.coords, control)
	vecCE := sprec.Vec2Diff(position, control)

	// start and end are excluded from this loop on purpose
	for i := 1; i < tessellation; i++ {
		t := float32(i) / float32(tessellation)
		alpha := (1 - t) * (1 - t)
		beta := t * t

		p.pointTemplate.coords = sprec.Vec2Sum(
			control,
			sprec.Vec2Sum(
				sprec.Vec2Prod(vecCS, alpha),
				sprec.Vec2Prod(vecCE, beta),
			),
		)
		p.addPoint()
	}

	p.pointTemplate.coords = position
	p.addPoint()
}

// CubeTo creates a cubic Bezier curve from the last path position
// to the newly specified position by going past the two specified
// control points.
func (p *canvasPath) CubeTo(control1, control2, position sprec.Vec2) {
	// TODO: Evaluate tessellation based on curvature and size
	const tessellation = 5

	lastPoint := p.pointTemplate

	// start and end are excluded from this loop on purpose
	for i := 1; i < tessellation; i++ {
		t := float32(i) / float32(tessellation)
		alpha := (1 - t) * (1 - t) * (1 - t)
		beta := 3 * (1 - t) * (1 - t) * t
		gamma := 3 * (1 - t) * t * t
		delta := t * t * t

		p.pointTemplate.coords = sprec.Vec2Sum(
			sprec.Vec2Sum(
				sprec.Vec2Prod(lastPoint.coords, alpha),
				sprec.Vec2Prod(control1, beta),
			),
			sprec.Vec2Sum(
				sprec.Vec2Prod(control2, gamma),
				sprec.Vec2Prod(position, delta),
			),
		)
		p.addPoint()
	}

	p.pointTemplate.coords = position
	p.addPoint()
}

// CloseLoop makes an automatic line connection back to the starting
// sub-path point.
func (p *canvasPath) CloseLoop() {
	if subPathCount := len(p.subPathOffsets); subPathCount > 0 {
		lastSubPathOffset := p.subPathOffsets[subPathCount-1]
		p.pointTemplate = p.points[lastSubPathOffset]
		p.addPoint()
	}
}

// Rectangle is a helper function that draws a rectangle at the
// specified position and size using a sequence of MoveTo and LineTo
// instructions.
func (p *canvasPath) Rectangle(position, size sprec.Vec2) {
	p.MoveTo(
		position,
	)
	p.LineTo(
		sprec.NewVec2(position.X, position.Y+size.Y),
	)
	p.LineTo(
		sprec.NewVec2(position.X+size.X, position.Y+size.Y),
	)
	p.LineTo(
		sprec.NewVec2(position.X+size.X, position.Y),
	)
	p.CloseLoop()
}

// Triangle is a helper function that draws a triangle with the
// specified corners, using a sequence of MoveTo and LineTo
// instructions.
func (p *canvasPath) Triangle(a, b, c sprec.Vec2) {
	p.MoveTo(a)
	p.LineTo(b)
	p.LineTo(c)
	p.CloseLoop()
}

// Circle is a helper function that draws a circle at the
// specified position and with the specified radius using a
// sequence of Shape instructions.
func (p *canvasPath) Circle(position sprec.Vec2, radius float32) {
	// TODO
}

// RoundRectangle is a helper function that draws a rounded
// rectangle at the specified position and with the specified size
// and corner radiuses (top-left, top-right, bottom-left, bottom-right).
func (p *canvasPath) RoundRectangle(position, size sprec.Vec2, roundness sprec.Vec4) {
	topLeft := roundness.X
	topRight := roundness.Y
	bottomLeft := roundness.Z
	bottomRight := roundness.W

	p.MoveTo(
		sprec.NewVec2(0.0, size.Y-bottomLeft),
	)
	p.QuadTo(
		sprec.NewVec2(0.0, size.Y),
		sprec.NewVec2(bottomLeft, size.Y),
	)
	p.LineTo(
		sprec.NewVec2(size.X-bottomRight, size.Y),
	)
	p.QuadTo(
		sprec.NewVec2(size.X, size.Y),
		sprec.NewVec2(size.X, size.Y-bottomRight),
	)
	p.LineTo(
		sprec.NewVec2(size.X, topRight),
	)
	p.QuadTo(
		sprec.NewVec2(size.X, 0),
		sprec.NewVec2(size.X-topRight, 0),
	)
	p.LineTo(
		sprec.NewVec2(topLeft, 0),
	)
	p.QuadTo(
		sprec.NewVec2(0, 0),
		sprec.NewVec2(0, topLeft),
	)
	p.CloseLoop()
}

func (p *canvasPath) startSubPath() {
	p.subPathOffsets = append(p.subPathOffsets, len(p.points))
}

func (p *canvasPath) addPoint() {
	p.points = append(p.points, p.pointTemplate)
}
