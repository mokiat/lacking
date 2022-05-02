package ui

import (
	"github.com/mokiat/gomath/sprec"
	"github.com/mokiat/lacking/render"
)

const (
	initialContourPointCount      = 1024
	initialContourSubContourCount = 4
)

func newContour(state *canvasState, shaders ShaderCollection) *Contour {
	return &Contour{
		state: state,

		contourMesh:     newContourMesh(maxVertexCount),
		contourMaterial: newMaterial(shaders.ContourMaterial),

		points:      make([]contourPoint, 0, initialContourPointCount),
		subContours: make([]subContour, 0, initialContourSubContourCount),
	}
}

// Contour represents a module for drawing curved lines.
type Contour struct {
	state        *canvasState
	commandQueue render.CommandQueue

	contourMesh     *ContourMesh
	contourMaterial *Material
	contourPipeline render.Pipeline

	engaged bool

	clipBounds      sprec.Vec4
	transformMatrix sprec.Mat4

	points      []contourPoint
	subContours []subContour
}

func (c *Contour) onCreate(api render.API, commandQueue render.CommandQueue) {
	c.commandQueue = commandQueue
	c.contourMesh.Allocate(api)
	c.contourMaterial.Allocate(api)
	c.contourPipeline = api.CreatePipeline(render.PipelineInfo{
		Program:                     c.contourMaterial.program,
		VertexArray:                 c.contourMesh.vertexArray,
		Topology:                    render.TopologyTriangles,
		Culling:                     render.CullModeNone,
		FrontFace:                   render.FaceOrientationCCW,
		DepthTest:                   false,
		DepthWrite:                  false,
		StencilTest:                 false,
		ColorWrite:                  [4]bool{true, true, true, true},
		BlendEnabled:                true,
		BlendSourceColorFactor:      render.BlendFactorSourceAlpha,
		BlendSourceAlphaFactor:      render.BlendFactorSourceAlpha,
		BlendDestinationColorFactor: render.BlendFactorOneMinusSourceAlpha,
		BlendDestinationAlphaFactor: render.BlendFactorOneMinusSourceAlpha,
		BlendOpColor:                render.BlendOperationAdd,
		BlendOpAlpha:                render.BlendOperationAdd,
	})
}

func (c *Contour) onDestroy() {
	defer c.contourMesh.Release()
	defer c.contourMaterial.Release()
	defer c.contourPipeline.Release()
}

func (c *Contour) onBegin() {
	c.contourMesh.Reset()
}

func (c *Contour) onEnd() {
	c.contourMesh.Update()
}

// Begin starts a new contour.
// Make sure to use End when finished with the contour.
func (c *Contour) begin() {
	if c.engaged {
		panic("contour already started")
	}
	c.engaged = true

	currentLayer := c.state.currentLayer
	c.clipBounds = sprec.NewVec4(
		float32(currentLayer.ClipBounds.X),
		float32(currentLayer.ClipBounds.X+currentLayer.ClipBounds.Width),
		float32(currentLayer.ClipBounds.Y),
		float32(currentLayer.ClipBounds.Y+currentLayer.ClipBounds.Height),
	)
	c.transformMatrix = currentLayer.Transform

	c.points = c.points[:0]
	c.subContours = c.subContours[:0]
}

// End marks the end of the contour and pushes all collected data for
// drawing.
func (c *Contour) end() {
	if !c.engaged {
		panic("contour already ended")
	}
	c.engaged = false

	c.commandQueue.BindPipeline(c.contourPipeline)
	c.commandQueue.Uniform4f(c.contourMaterial.clipDistancesLocation, c.clipBounds.Array())
	c.commandQueue.UniformMatrix4f(c.contourMaterial.projectionMatrixLocation, c.state.projectionMatrix.ColumnMajorArray())
	c.commandQueue.UniformMatrix4f(c.contourMaterial.transformMatrixLocation, c.transformMatrix.ColumnMajorArray())
	c.commandQueue.Uniform1i(c.contourMaterial.textureLocation, 0)

	for _, subContour := range c.subContours {
		pointIndex := subContour.pointOffset
		current := c.points[pointIndex]
		next := c.points[pointIndex+1]
		currentNormal := endPointNormal(
			current.coords,
			next.coords,
		)
		pointIndex++

		vertexOffset := c.contourMesh.Offset()
		for pointIndex < subContour.pointOffset+subContour.pointCount {
			prev := current
			prevNormal := currentNormal

			current = c.points[pointIndex]
			if pointIndex != subContour.pointOffset+subContour.pointCount-1 {
				next := c.points[pointIndex+1]
				currentNormal = midPointNormal(
					prev.coords,
					current.coords,
					next.coords,
				)
			} else {
				currentNormal = endPointNormal(
					prev.coords,
					current.coords,
				)
			}

			prevLeft := ContourVertex{
				position: sprec.Vec2Sum(prev.coords, sprec.Vec2Prod(prevNormal, prev.stroke.innerSize)),
				color:    prev.stroke.color,
			}
			prevRight := ContourVertex{
				position: sprec.Vec2Diff(prev.coords, sprec.Vec2Prod(prevNormal, prev.stroke.outerSize)),
				color:    prev.stroke.color,
			}
			currentLeft := ContourVertex{
				position: sprec.Vec2Sum(current.coords, sprec.Vec2Prod(currentNormal, current.stroke.innerSize)),
				color:    prev.stroke.color,
			}
			currentRight := ContourVertex{
				position: sprec.Vec2Diff(current.coords, sprec.Vec2Prod(currentNormal, current.stroke.outerSize)),
				color:    prev.stroke.color,
			}

			c.contourMesh.Append(prevLeft)
			c.contourMesh.Append(prevRight)
			c.contourMesh.Append(currentLeft)

			c.contourMesh.Append(prevRight)
			c.contourMesh.Append(currentRight)
			c.contourMesh.Append(currentLeft)

			pointIndex++
		}
		vertexCount := c.contourMesh.Offset() - vertexOffset

		c.commandQueue.Draw(vertexOffset, vertexCount, 1)
	}
}

// MoveTo positions the cursor to the specified position and
// marks the specified stroke setting for that point.
func (c *Contour) moveTo(position sprec.Vec2, stroke Stroke) {
	c.startSubContour()
	c.addPoint(contourPoint{
		coords: position,
		stroke: uiStrokeToStroke(stroke),
	})
}

// LineTo creates a direct line from the last cursor position
// to the newly specified position and sets the specified
// stroke for the new position.
func (c *Contour) lineTo(position sprec.Vec2, stroke Stroke) {
	c.addPoint(contourPoint{
		coords: position,
		stroke: uiStrokeToStroke(stroke),
	})
}

// QuadTo creates a quadratic Bezier curve from the last cursor
// position to the newly specified position by going past the
// specified control point. The target position is assigned
// the specified stroke setting.
func (c *Contour) quadTo(control, position sprec.Vec2, stroke Stroke) {
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
		c.addPoint(contourPoint{
			coords: sprec.Vec2Sum(
				control,
				sprec.Vec2Sum(
					sprec.Vec2Prod(vecCS, alpha),
					sprec.Vec2Prod(vecCE, beta),
				),
			),
			stroke: mixStrokes(lastPoint.stroke, uiStrokeToStroke(stroke), t),
		})
	}

	c.addPoint(contourPoint{
		coords: position,
		stroke: uiStrokeToStroke(stroke),
	})
}

// CubeTo creates a cubic Bezier curve from the last cursor position
// to the newly specified position by going past the two specified
// control points. The target position is assigned
// the specified stroke setting.
func (c *Contour) cubeTo(control1, control2, position sprec.Vec2, stroke Stroke) {
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
		c.addPoint(contourPoint{
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
			stroke: mixStrokes(lastPoint.stroke, uiStrokeToStroke(stroke), t),
		})
	}

	c.addPoint(contourPoint{
		coords: position,
		stroke: uiStrokeToStroke(stroke),
	})
}

// CloseLoop makes an automatic line connection back to the starting
// point, as specified via MoveTo.
func (c *Contour) closeLoop() {
	lastSubContour := c.subContours[len(c.subContours)-1]
	c.addPoint(c.points[lastSubContour.pointOffset])
}

func (c *Contour) startSubContour() {
	c.subContours = append(c.subContours, subContour{
		pointOffset: len(c.points),
		pointCount:  0,
	})
}

func (c *Contour) addPoint(point contourPoint) {
	c.points = append(c.points, point)
	c.subContours[len(c.subContours)-1].pointCount++
}

func (c *Contour) lastPoint() contourPoint {
	return c.points[len(c.points)-1]
}

// Stroke configures how a contour is to be drawn.
type Stroke struct {

	// Size determines the size of the contour.
	InnerSize float32

	OuterSize float32

	// Color specifies the color of the contour.
	Color Color
}

type contourPoint struct {
	coords sprec.Vec2
	stroke contourStroke
}

type subContour struct {
	pointOffset int
	pointCount  int
}

type contourStroke struct {
	innerSize float32
	outerSize float32
	color     sprec.Vec4
}

func uiStrokeToStroke(stroke Stroke) contourStroke {
	return contourStroke{
		innerSize: stroke.InnerSize,
		outerSize: stroke.OuterSize,
		color:     uiColorToVec(stroke.Color),
	}
}

func mixStrokes(a, b contourStroke, alpha float32) contourStroke {
	return contourStroke{
		innerSize: (1-alpha)*a.innerSize + alpha*b.innerSize,
		outerSize: (1-alpha)*a.outerSize + alpha*b.outerSize,
		color: sprec.Vec4Sum(
			sprec.Vec4Prod(a.color, (1-alpha)),
			sprec.Vec4Prod(b.color, alpha),
		),
	}
}

func midPointNormal(prev, middle, next sprec.Vec2) sprec.Vec2 {
	normal1 := endPointNormal(prev, middle)
	normal2 := endPointNormal(middle, next)
	normalSum := sprec.Vec2Sum(normal1, normal2)
	dot := sprec.Vec2Dot(normal1, normalSum)
	return sprec.Vec2Quot(normalSum, dot)
}

func endPointNormal(prev, next sprec.Vec2) sprec.Vec2 {
	tangent := sprec.UnitVec2(sprec.Vec2Diff(next, prev))
	return sprec.NewVec2(tangent.Y, -tangent.X)
}
