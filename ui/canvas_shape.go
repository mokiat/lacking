package ui

import (
	"github.com/mokiat/gomath/sprec"
	"github.com/mokiat/lacking/render"
)

const (
	initialShapePointCount    = 1024
	initialShapeSubShapeCount = 4
)

func newShape(state *canvasState, shaders ShaderCollection) *Shape {
	return &Shape{
		state: state,

		mesh:          newShapeMesh(maxVertexCount),
		shadeMaterial: newMaterial(shaders.ShapeMaterial),
		blankMaterial: newMaterial(shaders.ShapeBlankMaterial),

		points:    make([]shapePoint, 0, initialShapePointCount),
		subShapes: make([]subShape, 0, initialShapeSubShapeCount),
	}
}

// Shape represents a module for drawing solid shapes.
type Shape struct {
	state        *canvasState
	commandQueue render.CommandQueue

	mesh                 *ShapeMesh
	shadeMaterial        *Material
	blankMaterial        *Material
	maskPipeline         render.Pipeline
	shadeSimplePipeline  render.Pipeline
	shadeNonZeroPipeline render.Pipeline
	shadeOddPipeline     render.Pipeline

	engaged bool

	clipBounds             sprec.Vec4
	transformMatrix        sprec.Mat4
	textureTransformMatrix sprec.Mat4
	rule                   FillRule
	color                  sprec.Vec4
	image                  *Image

	points    []shapePoint
	subShapes []subShape
}

func (s *Shape) onCreate(api render.API, commandQueue render.CommandQueue) {
	s.commandQueue = commandQueue

	s.mesh.Allocate(api)
	s.shadeMaterial.Allocate(api)
	s.blankMaterial.Allocate(api)
	s.maskPipeline = api.CreatePipeline(render.PipelineInfo{
		Program:         s.blankMaterial.program,
		VertexArray:     s.mesh.vertexArray,
		Topology:        render.TopologyTriangleFan,
		Culling:         render.CullModeNone,
		FrontFace:       render.FaceOrientationCCW,
		LineWidth:       1.0,
		DepthTest:       false,
		DepthWrite:      false,
		DepthComparison: render.ComparisonAlways,
		StencilTest:     true,
		StencilFront: render.StencilOperationState{
			StencilFailOp:  render.StencilOperationKeep,
			DepthFailOp:    render.StencilOperationKeep,
			PassOp:         render.StencilOperationIncreaseWrap,
			Comparison:     render.ComparisonAlways,
			ComparisonMask: 0xFF,
			Reference:      0x00,
			WriteMask:      0xFF,
		},
		StencilBack: render.StencilOperationState{
			StencilFailOp:  render.StencilOperationKeep,
			DepthFailOp:    render.StencilOperationKeep,
			PassOp:         render.StencilOperationDecreaseWrap,
			Comparison:     render.ComparisonAlways,
			ComparisonMask: 0xFF,
			Reference:      0x00,
			WriteMask:      0xFF,
		},
		ColorWrite:                  [4]bool{false, false, false, false},
		BlendEnabled:                false,
		BlendSourceColorFactor:      render.BlendFactorSourceAlpha,
		BlendSourceAlphaFactor:      render.BlendFactorSourceAlpha,
		BlendDestinationColorFactor: render.BlendFactorOneMinusSourceAlpha,
		BlendDestinationAlphaFactor: render.BlendFactorOneMinusSourceAlpha,
		BlendOpColor:                render.BlendOperationAdd,
		BlendOpAlpha:                render.BlendOperationAdd,
	})
	s.shadeSimplePipeline = api.CreatePipeline(render.PipelineInfo{
		Program:                     s.shadeMaterial.program,
		VertexArray:                 s.mesh.vertexArray,
		Topology:                    render.TopologyTriangleFan,
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
	s.shadeNonZeroPipeline = api.CreatePipeline(render.PipelineInfo{
		Program:         s.shadeMaterial.program,
		VertexArray:     s.mesh.vertexArray,
		Topology:        render.TopologyTriangleFan,
		Culling:         render.CullModeNone,
		FrontFace:       render.FaceOrientationCCW,
		LineWidth:       1.0,
		DepthTest:       false,
		DepthWrite:      false,
		DepthComparison: render.ComparisonAlways,
		StencilTest:     true,
		StencilFront: render.StencilOperationState{
			StencilFailOp:  render.StencilOperationKeep,
			DepthFailOp:    render.StencilOperationKeep,
			PassOp:         render.StencilOperationReplace,
			Comparison:     render.ComparisonNotEqual,
			ComparisonMask: 0xFF,
			Reference:      0x00,
			WriteMask:      0xFF,
		},
		StencilBack: render.StencilOperationState{
			StencilFailOp:  render.StencilOperationKeep,
			DepthFailOp:    render.StencilOperationKeep,
			PassOp:         render.StencilOperationReplace,
			Comparison:     render.ComparisonNotEqual,
			ComparisonMask: 0xFF,
			Reference:      0x00,
			WriteMask:      0xFF,
		},
		ColorWrite:                  [4]bool{true, true, true, true},
		BlendEnabled:                true,
		BlendSourceColorFactor:      render.BlendFactorSourceAlpha,
		BlendSourceAlphaFactor:      render.BlendFactorSourceAlpha,
		BlendDestinationColorFactor: render.BlendFactorOneMinusSourceAlpha,
		BlendDestinationAlphaFactor: render.BlendFactorOneMinusSourceAlpha,
		BlendOpColor:                render.BlendOperationAdd,
		BlendOpAlpha:                render.BlendOperationAdd,
	})
	s.shadeOddPipeline = api.CreatePipeline(render.PipelineInfo{
		Program:     s.shadeMaterial.program,
		VertexArray: s.mesh.vertexArray,
		Topology:    render.TopologyTriangleFan,
		Culling:     render.CullModeNone,
		FrontFace:   render.FaceOrientationCCW,
		DepthTest:   false,
		DepthWrite:  false,
		StencilTest: true,
		StencilFront: render.StencilOperationState{
			StencilFailOp:  render.StencilOperationKeep,
			DepthFailOp:    render.StencilOperationKeep,
			PassOp:         render.StencilOperationReplace,
			Comparison:     render.ComparisonNotEqual,
			ComparisonMask: 0x01,
			Reference:      0x00,
			WriteMask:      0xFF,
		},
		StencilBack: render.StencilOperationState{
			StencilFailOp:  render.StencilOperationKeep,
			DepthFailOp:    render.StencilOperationKeep,
			PassOp:         render.StencilOperationReplace,
			Comparison:     render.ComparisonNotEqual,
			ComparisonMask: 0x01,
			Reference:      0x00,
			WriteMask:      0xFF,
		},
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

func (s *Shape) onDestroy() {
	defer s.mesh.Release()
	defer s.shadeMaterial.Release()
	defer s.blankMaterial.Release()
	defer s.maskPipeline.Release()
	defer s.shadeSimplePipeline.Release()
	defer s.shadeNonZeroPipeline.Release()
	defer s.shadeOddPipeline.Release()
}

func (s *Shape) onBegin() {
	s.mesh.Reset()
}

func (s *Shape) onEnd() {
	s.mesh.Update()
}

// Begin starts a new solid shape using the specified fill settings.
// Make sure to use End when finished with the shape.
func (s *Shape) begin(fill Fill) {
	if s.engaged {
		panic("shape already started")
	}
	s.engaged = true

	currentLayer := s.state.currentLayer
	s.clipBounds = sprec.NewVec4(
		float32(currentLayer.ClipBounds.X),
		float32(currentLayer.ClipBounds.X+currentLayer.ClipBounds.Width),
		float32(currentLayer.ClipBounds.Y),
		float32(currentLayer.ClipBounds.Y+currentLayer.ClipBounds.Height),
	)
	s.transformMatrix = currentLayer.Transform
	s.textureTransformMatrix = sprec.Mat4MultiProd(
		sprec.ScaleMat4(
			1.0/fill.ImageSize.X,
			1.0/fill.ImageSize.Y,
			1.0,
		),
		sprec.TranslationMat4(
			-fill.ImageOffset.X,
			-fill.ImageOffset.Y,
			0.0,
		),
	)

	s.rule = fill.Rule
	s.color = uiColorToVec(fill.Color)
	s.image = fill.Image

	s.points = s.points[:0]
	s.subShapes = s.subShapes[:0]
}

// End marks the end of the shape and pushes all collected data for
// drawing.
func (s *Shape) end() {
	if !s.engaged {
		panic("shape already ended")
	}
	s.engaged = false

	vertexOffset := s.mesh.Offset()
	for _, point := range s.points {
		s.mesh.Append(ShapeVertex{
			position: point.coords,
		})
	}

	// draw stencil mask for all sub-shapes
	if s.rule != FillRuleSimple {
		s.commandQueue.BindPipeline(s.maskPipeline)
		s.commandQueue.Uniform4f(s.blankMaterial.clipDistancesLocation, s.clipBounds.Array())
		s.commandQueue.UniformMatrix4f(s.blankMaterial.projectionMatrixLocation, s.state.projectionMatrix.ColumnMajorArray())
		s.commandQueue.UniformMatrix4f(s.blankMaterial.transformMatrixLocation, s.transformMatrix.ColumnMajorArray())

		for _, subShape := range s.subShapes {
			s.commandQueue.Draw(vertexOffset+subShape.pointOffset, subShape.pointCount, 1)
		}
	}

	// render color shader shape and clear stencil buffer
	switch s.rule {
	case FillRuleSimple:
		s.commandQueue.BindPipeline(s.shadeSimplePipeline)
	case FillRuleNonZero:
		s.commandQueue.BindPipeline(s.shadeNonZeroPipeline)
	case FillRuleEvenOdd:
		s.commandQueue.BindPipeline(s.shadeOddPipeline)
	default:
		s.commandQueue.BindPipeline(s.shadeSimplePipeline)
	}

	texture := s.state.whiteMask
	if s.image != nil {
		texture = s.image.texture
	}

	s.commandQueue.Uniform4f(s.shadeMaterial.clipDistancesLocation, s.clipBounds.Array())
	s.commandQueue.Uniform4f(s.shadeMaterial.colorLocation, s.color.Array())
	s.commandQueue.UniformMatrix4f(s.shadeMaterial.projectionMatrixLocation, s.state.projectionMatrix.ColumnMajorArray())
	s.commandQueue.UniformMatrix4f(s.shadeMaterial.transformMatrixLocation, s.transformMatrix.ColumnMajorArray())
	s.commandQueue.UniformMatrix4f(s.shadeMaterial.textureTransformMatrixLocation, s.textureTransformMatrix.ColumnMajorArray())
	s.commandQueue.TextureUnit(0, texture)
	s.commandQueue.Uniform1i(s.shadeMaterial.textureLocation, 0)

	for _, subShape := range s.subShapes {
		s.commandQueue.Draw(vertexOffset+subShape.pointOffset, subShape.pointCount, 1)
	}
}

// MoveTo positions the cursor to the specified position.
func (s *Shape) moveTo(position sprec.Vec2) {
	s.startSubShape()
	s.addPoint(shapePoint{
		coords: position,
	})
}

// LineTo creates a direct line from the last cursor position
// to the newly specified position.
func (s *Shape) lineTo(position sprec.Vec2) {
	s.addPoint(shapePoint{
		coords: position,
	})
}

// QuadTo creates a quadratic Bezier curve from the last cursor
// position to the newly specified position by going past the
// specified control point.
func (s *Shape) quadTo(control, position sprec.Vec2) {
	// TODO: Evaluate tessellation based on curvature and size
	const tessellation = 5

	lastPoint := s.lastPoint()
	vecCS := sprec.Vec2Diff(lastPoint.coords, control)
	vecCE := sprec.Vec2Diff(position, control)

	// start and end are excluded from this loop on purpose
	for i := 1; i < tessellation; i++ {
		t := float32(i) / float32(tessellation)
		alpha := (1 - t) * (1 - t)
		beta := t * t
		s.addPoint(shapePoint{
			coords: sprec.Vec2Sum(
				control,
				sprec.Vec2Sum(
					sprec.Vec2Prod(vecCS, alpha),
					sprec.Vec2Prod(vecCE, beta),
				),
			),
		})
	}

	s.addPoint(shapePoint{
		coords: position,
	})
}

// CubeTo creates a cubic Bezier curve from the last cursor position
// to the newly specified position by going past the two specified
// control points.
func (s *Shape) cubeTo(control1, control2, position sprec.Vec2) {
	// TODO: Evaluate tessellation based on curvature and size
	const tessellation = 5

	lastPoint := s.lastPoint()

	// start and end are excluded from this loop on purpose
	for i := 1; i < tessellation; i++ {
		t := float32(i) / float32(tessellation)
		alpha := (1 - t) * (1 - t) * (1 - t)
		beta := 3 * (1 - t) * (1 - t) * t
		gamma := 3 * (1 - t) * t * t
		delta := t * t * t
		s.addPoint(shapePoint{
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
	s.addPoint(shapePoint{
		coords: position,
	})
}

func (s *Shape) startSubShape() {
	s.subShapes = append(s.subShapes, subShape{
		pointOffset: len(s.points),
		pointCount:  0,
	})
}

func (s *Shape) addPoint(point shapePoint) {
	s.points = append(s.points, point)
	s.subShapes[len(s.subShapes)-1].pointCount++
}

func (s *Shape) lastPoint() shapePoint {
	return s.points[len(s.points)-1]
}

// Fill configures how a solid shape is to be drawn.
type Fill struct {

	// Rule specifies the mechanism through which it is determined
	// which point is part of the shape in an overlapping or concave
	// polygon.
	Rule FillRule

	// Color specifies the color to use to fill the shape.
	Color Color

	// Image specifies an optional image to be used for filling
	// the shape.
	Image *Image

	// ImageOffset determines the offset of the origin of the
	// image relative to the current translation context.
	ImageOffset sprec.Vec2

	// ImageSize determines the size of the drawn image. In
	// essence, this size performs scaling.
	ImageSize sprec.Vec2
}

// FillRule represents the mechanism through which it is determined
// which point is part of the shape in an overlapping or concave
// polygon.
type FillRule int

const (
	// FillRuleSimple is the fastest approach and should be used
	// with non-overlapping concave shapes.
	FillRuleSimple FillRule = iota

	// FillRuleNonZero will fill areas that are covered by the
	// shape, regardless if it overlaps.
	FillRuleNonZero

	// FillRuleEvenOdd will fill areas that are covered by the
	// shape and it does not overlap or overlaps an odd number
	// of times.
	FillRuleEvenOdd
)

type shapePoint struct {
	coords sprec.Vec2
}

type subShape struct {
	pointOffset int
	pointCount  int
}
