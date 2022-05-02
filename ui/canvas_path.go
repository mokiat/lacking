package ui

import "github.com/mokiat/gomath/sprec"

const (
	initialPointCapacity   = 1024
	initialSubPathCapacity = 16
)

type canvasPoint struct {
	coords    sprec.Vec2
	innerSize float32
	outerSize float32
	color     Color
}

func newCanvasPath() *canvasPath {
	return &canvasPath{
		points:         make([]canvasPoint, 0, initialPointCapacity),
		subPathOffsets: make([]int, 0, initialSubPathCapacity),

		pointTemplate: canvasPoint{
			innerSize: 1.0,
			outerSize: 0.0,
			color:     Black(),
		},
	}
}

type canvasPath struct {
	points         []canvasPoint
	subPathOffsets []int

	pointTemplate canvasPoint
}

// StrokeColor returns the configured stroke color.
func (p *canvasPath) StrokeColor() Color {
	return p.pointTemplate.color
}

// SetStrokeColor changes the stroke color of future-placed points.
func (p *canvasPath) SetStrokeColor(color Color) {
	p.pointTemplate.color = color
}

// StrokeSize returns the sum of the currently configured inner and outer
// stroke sizes.
func (p *canvasPath) StrokeSize() float32 {
	return p.pointTemplate.innerSize + p.pointTemplate.outerSize
}

// SetStrokeSize configures the inner stoke size to the specified size
// value and sets the outer stroke size to zero, which will be used in
// subsequent Path points.
//
// For more control consider using SetStrokeSizeSeparate instead.
func (p *canvasPath) SetStrokeSize(size float32) {
	p.pointTemplate.innerSize = size
	p.pointTemplate.outerSize = 0.0
}

// StrokeSizeSeparate returns the inner and outer stroke sizes.
func (p *canvasPath) StrokeSizeSeparate() (float32, float32) {
	return p.pointTemplate.innerSize, p.pointTemplate.outerSize
}

// SetStrokeSizeSeparate configures the inner and outer stroke sizes
// to be used for subsequent Path points.
func (p *canvasPath) SetStrokeSizeSeparate(inner, outer float32) {
	p.pointTemplate.innerSize = inner
	p.pointTemplate.outerSize = outer
}

// Reset erases all path points and starts from the beginning.
func (p *canvasPath) Reset() {
	p.points = p.points[:0]
	p.subPathOffsets = p.subPathOffsets[:0]
	p.pointTemplate.innerSize = 1.0
	p.pointTemplate.outerSize = 0.0
	p.pointTemplate.color = Black()
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

	firstPoint := p.lastPoint()
	lastPoint := p.pointTemplate
	lastPoint.coords = position

	vecCS := sprec.Vec2Diff(firstPoint.coords, control)
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
		p.pointTemplate.innerSize = sprec.Mix(firstPoint.innerSize, lastPoint.innerSize, t)
		p.pointTemplate.outerSize = sprec.Mix(firstPoint.outerSize, lastPoint.outerSize, t)
		p.pointTemplate.color = MixColors(firstPoint.color, lastPoint.color, t)
		p.addPoint()
	}

	p.pointTemplate = lastPoint
	p.addPoint()
}

// CubeTo creates a cubic Bezier curve from the last path position
// to the newly specified position by going past the two specified
// control points.
func (p *canvasPath) CubeTo(control1, control2, position sprec.Vec2) {
	// TODO: Evaluate tessellation based on curvature and size
	const tessellation = 5

	firstPoint := p.lastPoint()
	lastPoint := p.pointTemplate
	lastPoint.coords = position

	// start and end are excluded from this loop on purpose
	for i := 1; i < tessellation; i++ {
		t := float32(i) / float32(tessellation)
		alpha := (1 - t) * (1 - t) * (1 - t)
		beta := 3 * (1 - t) * (1 - t) * t
		gamma := 3 * (1 - t) * t * t
		delta := t * t * t

		p.pointTemplate.coords = sprec.Vec2Sum(
			sprec.Vec2Sum(
				sprec.Vec2Prod(firstPoint.coords, alpha),
				sprec.Vec2Prod(control1, beta),
			),
			sprec.Vec2Sum(
				sprec.Vec2Prod(control2, gamma),
				sprec.Vec2Prod(position, delta),
			),
		)
		p.pointTemplate.innerSize = sprec.Mix(firstPoint.innerSize, lastPoint.innerSize, t)
		p.pointTemplate.outerSize = sprec.Mix(firstPoint.outerSize, lastPoint.outerSize, t)
		p.pointTemplate.color = MixColors(firstPoint.color, lastPoint.color, t)
		p.addPoint()
	}

	p.pointTemplate = lastPoint
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
	p.MoveTo(sprec.NewVec2(
		position.X+radius,
		position.Y,
	))
	const step = float32(10) // gives good results
	for degrees := step; degrees < 360.0; degrees += step {
		angle := sprec.Degrees(degrees)
		p.LineTo(sprec.NewVec2(
			position.X+sprec.Cos(angle)*radius,
			position.Y+sprec.Sin(angle)*radius,
		))
	}
	p.CloseLoop()
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
		sprec.NewVec2(
			position.X,
			position.Y+size.Y-bottomLeft,
		),
	)
	p.QuadTo(
		sprec.NewVec2(
			position.X,
			position.Y+size.Y,
		),
		sprec.NewVec2(
			position.X+bottomLeft,
			position.Y+size.Y,
		),
	)
	p.LineTo(
		sprec.NewVec2(
			position.X+size.X-bottomRight,
			position.Y+size.Y,
		),
	)
	p.QuadTo(
		sprec.NewVec2(
			position.X+size.X,
			position.Y+size.Y,
		),
		sprec.NewVec2(
			position.X+size.X,
			position.Y+size.Y-bottomRight,
		),
	)
	p.LineTo(
		sprec.NewVec2(
			position.X+size.X,
			position.Y+topRight,
		),
	)
	p.QuadTo(
		sprec.NewVec2(
			position.X+size.X,
			position.Y,
		),
		sprec.NewVec2(
			position.X+size.X-topRight,
			position.Y,
		),
	)
	p.LineTo(
		sprec.NewVec2(
			position.X+topLeft,
			position.Y,
		),
	)
	p.QuadTo(
		sprec.NewVec2(
			position.X,
			position.Y,
		),
		sprec.NewVec2(
			position.X,
			position.Y+topLeft,
		),
	)
	p.CloseLoop()
}

func (p *canvasPath) startSubPath() {
	p.subPathOffsets = append(p.subPathOffsets, len(p.points))
}

func (p *canvasPath) addPoint() {
	p.points = append(p.points, p.pointTemplate)
}

func (p *canvasPath) lastPoint() canvasPoint {
	return p.points[len(p.points)-1]
}
