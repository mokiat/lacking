package internal

import (
	"github.com/go-gl/gl/v4.6-core/gl"

	"github.com/mokiat/gomath/sprec"
	"github.com/mokiat/lacking/ui"
)

const (
	initialPointCount    = 1024
	initialSubShapeCount = 16
)

type Fill struct {
	// Rule            FillRule
	// Winding         Winding
	Color sprec.Vec4
	Image *Image
}

type Stroke struct {
	Size  float32
	Color sprec.Vec4
}

func MixStrokes(a, b Stroke, alpha float32) Stroke {
	return Stroke{
		Size: (1-alpha)*a.Size + alpha*b.Size,
		Color: sprec.Vec4Sum(
			sprec.Vec4Prod(a.Color, (1-alpha)),
			sprec.Vec4Prod(b.Color, alpha),
		),
	}
}

type Point struct {
	coords    sprec.Vec2
	inStroke  Stroke
	outStroke Stroke
}

type SubShape struct {
}

func NewCurveRenderer() *CurveRenderer {
	return &CurveRenderer{}
}

type CurveRenderer struct {
}

// TODO: Maybe split fill and stroke into separate APIs.
// The fill renderer could have a single stroke or none at all.

func NewShapeRenderer(target Target) *ShapeRenderer {
	return &ShapeRenderer{
		target:    target,
		points:    make([]Point, 0, initialPointCount),
		subShapes: make([]SubShape, 0, initialSubShapeCount),
	}
}

type ShapeRenderer struct {
	target Target

	mesh *Mesh

	fill Fill

	// startOffset int
	// start       sprec.Vec2

	points    []Point
	subShapes []SubShape
}

func (r *ShapeRenderer) BeginShape(fill Fill) {
	r.fill = fill
	r.points = r.points[:0]
	r.subShapes = r.subShapes[:0]
}

func (r *ShapeRenderer) MoveTo(position sprec.Vec2) {
	// r.startOffset = r.mesh.Offset()
	// r.start = position

	// r.mesh.Append(Vertex{
	// 	position: position,
	// })

	r.points = append(r.points, Point{
		coords: position,
	})
}

func (r *ShapeRenderer) LineTo(position sprec.Vec2, startStroke, endStroke Stroke) {
	r.points[len(r.points)-1].outStroke = startStroke
	r.points = append(r.points, Point{
		coords:   position,
		inStroke: endStroke,
	})
}

func (r *ShapeRenderer) QuadTo(control, position sprec.Vec2, startStroke, endStroke Stroke) {
	startPoint := r.points[len(r.points)-1]
	startPoint.outStroke = startStroke

	vecCS := sprec.Vec2Diff(startPoint.coords, control)
	vecCE := sprec.Vec2Diff(position, control)

	const tessellation = 30 // TODO: Evaluate based on points

	// Note: Start and end are excluded from this loop
	for i := 1; i < tessellation; i++ {
		t := float32(i) / float32(tessellation)
		alpha := (1 - t) * (1 - t)
		beta := t * t
		stroke := MixStrokes(startStroke, endStroke, t)
		r.points = append(r.points, Point{
			coords: sprec.Vec2Sum(
				control,
				sprec.Vec2Sum(
					sprec.Vec2Prod(vecCS, alpha),
					sprec.Vec2Prod(vecCE, beta),
				),
			),
			inStroke:  stroke,
			outStroke: stroke,
		})
	}

	r.points = append(r.points, Point{
		coords:   position,
		inStroke: endStroke,
	})
}

func (r *ShapeRenderer) CubeTo(control1, control2, position sprec.Vec2, startStroke, endStroke Stroke) {
	startPoint := r.points[len(r.points)-1]
	startPoint.outStroke = startStroke

	const tessellation = 30 // TODO: Evaluate based on points

	// Note: Start and end are excluded from this loop
	for i := 1; i < tessellation; i++ {
		t := float32(i) / float32(tessellation)
		alpha := (1 - t) * (1 - t) * (1 - t)
		beta := 3 * (1 - t) * (1 - t) * t
		gamma := 3 * (1 - t) * t * t
		delta := t * t * t
		stroke := MixStrokes(startStroke, endStroke, t)
		r.points = append(r.points, Point{
			coords: sprec.Vec2Sum(
				sprec.Vec2Sum(
					sprec.Vec2Prod(startPoint.coords, alpha),
					sprec.Vec2Prod(control1, beta),
				),
				sprec.Vec2Sum(
					sprec.Vec2Prod(control2, gamma),
					sprec.Vec2Prod(position, delta),
				),
			),
			inStroke:  stroke,
			outStroke: stroke,
		})
	}

	r.points = append(r.points, Point{
		coords:   position,
		inStroke: endStroke,
	})
}

func (r *ShapeRenderer) CloseLoop(startStroke, endStroke Stroke) {
	r.points[len(r.points)-1].outStroke = startStroke
	r.points[0].inStroke = endStroke
}

func (r *ShapeRenderer) EndShape() {
	translation := sprec.NewVec2(
		float32(c.currentLayer.Translation.X),
		float32(c.currentLayer.Translation.Y),
	)

	offset := c.mesh.Offset()
	// TODO: Only if background color or texture is set
	for _, point := range c.activeShape.Points {
		c.mesh.Append(Vertex{
			position: sprec.Vec2Sum(point.Vec2, translation),
			texCoord: sprec.NewVec2(0.0, 0.0),
			color:    c.activeShape.BackgroundColor,
		})
	}
	count := c.mesh.Offset() - offset

	cullFace := gl.BACK
	if c.activeShape.Winding == ui.WindingCW {
		cullFace = gl.FRONT
	}

	if c.activeShape.Rule == ui.FillRuleSimple {
		c.subMeshes = append(c.subMeshes, SubMesh{
			clipBounds: sprec.NewVec4(
				float32(c.currentLayer.ClipBounds.X),
				float32(c.currentLayer.ClipBounds.X+c.currentLayer.ClipBounds.Width),
				float32(c.currentLayer.ClipBounds.Y),
				float32(c.currentLayer.ClipBounds.Y+c.currentLayer.ClipBounds.Height),
			),
			material:     c.opaqueMaterial,
			texture:      c.whiteMask,
			vertexOffset: offset,
			vertexCount:  count,
			cullFace:     uint32(cullFace),
			primitive:    gl.TRIANGLE_FAN,
		})
	}

	if c.activeShape.Rule != ui.FillRuleSimple {
		// clear stencil
		c.subMeshes = append(c.subMeshes, SubMesh{
			clipBounds: sprec.NewVec4(
				float32(c.currentLayer.ClipBounds.X),
				float32(c.currentLayer.ClipBounds.X+c.currentLayer.ClipBounds.Width),
				float32(c.currentLayer.ClipBounds.Y),
				float32(c.currentLayer.ClipBounds.Y+c.currentLayer.ClipBounds.Height),
			),
			material:     c.opaqueMaterial,
			texture:      c.whiteMask,
			vertexOffset: offset,
			vertexCount:  count,
			culling:      false,
			cullFace:     uint32(cullFace),
			primitive:    gl.TRIANGLE_FAN,
			skipColor:    true,
			stencil:      true,
			stencilCfg: stencilConfig{
				stencilFuncFront: stencilFunc{
					fn:   gl.ALWAYS,
					ref:  0,
					mask: 0xFF,
				},
				stencilFuncBack: stencilFunc{
					fn:   gl.ALWAYS,
					ref:  0,
					mask: 0xFF,
				},
				stencilOpFront: stencilOp{
					sfail:  gl.REPLACE,
					dpfail: gl.REPLACE,
					dppass: gl.REPLACE,
				},
				stencilOpBack: stencilOp{
					sfail:  gl.REPLACE,
					dpfail: gl.REPLACE,
					dppass: gl.REPLACE,
				},
			},
		})

		// render stencil mask
		c.subMeshes = append(c.subMeshes, SubMesh{
			clipBounds: sprec.NewVec4(
				float32(c.currentLayer.ClipBounds.X),
				float32(c.currentLayer.ClipBounds.X+c.currentLayer.ClipBounds.Width),
				float32(c.currentLayer.ClipBounds.Y),
				float32(c.currentLayer.ClipBounds.Y+c.currentLayer.ClipBounds.Height),
			),
			material:     c.opaqueMaterial,
			texture:      c.whiteMask,
			vertexOffset: offset,
			vertexCount:  count,
			cullFace:     uint32(cullFace),
			primitive:    gl.TRIANGLE_FAN,
			skipColor:    true, // we don't want to render anything
			stencil:      true,
			stencilCfg: stencilConfig{
				stencilFuncFront: stencilFunc{
					fn:   gl.ALWAYS,
					ref:  0,
					mask: 0xFF,
				},
				stencilFuncBack: stencilFunc{
					fn:   gl.ALWAYS,
					ref:  0,
					mask: 0xFF,
				},
				stencilOpFront: stencilOp{
					sfail:  gl.KEEP,
					dpfail: gl.KEEP,
					dppass: gl.INCR_WRAP, // increase correct winding
				},
				stencilOpBack: stencilOp{
					sfail:  gl.KEEP,
					dpfail: gl.KEEP,
					dppass: gl.DECR_WRAP, // decrease incorrect winding
				},
			},
		})

		// render final polygon
		c.subMeshes = append(c.subMeshes, SubMesh{
			clipBounds: sprec.NewVec4(
				float32(c.currentLayer.ClipBounds.X),
				float32(c.currentLayer.ClipBounds.X+c.currentLayer.ClipBounds.Width),
				float32(c.currentLayer.ClipBounds.Y),
				float32(c.currentLayer.ClipBounds.Y+c.currentLayer.ClipBounds.Height),
			),
			material:     c.opaqueMaterial,
			texture:      c.whiteMask,
			vertexOffset: offset,
			vertexCount:  count,
			cullFace:     uint32(cullFace),
			primitive:    gl.TRIANGLE_FAN,
			skipColor:    false, // we want to render now
			stencil:      true,
			stencilCfg: stencilConfig{
				stencilFuncFront: stencilFunc{
					fn:   gl.LESS,
					ref:  0,
					mask: 0xFF,
				},
				stencilFuncBack: stencilFunc{
					fn:   gl.LESS,
					ref:  0,
					mask: 0xFF,
				},
				stencilOpFront: stencilOp{
					sfail:  gl.KEEP,
					dpfail: gl.KEEP,
					dppass: gl.KEEP,
				},
				stencilOpBack: stencilOp{
					sfail:  gl.KEEP,
					dpfail: gl.KEEP,
					dppass: gl.KEEP,
				},
			},
		})
	}

	// TODO: Draw stroke
}
