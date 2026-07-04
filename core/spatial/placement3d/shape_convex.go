package placement3d

import (
	"github.com/mokiat/gomath/dprec"
	"github.com/mokiat/lacking/core/spatial/gjk3d"
	"github.com/mokiat/lacking/core/spatial/shape3d"
)

type convexShape[S any] struct {
	baseShape[S]
	convexShapeRepresentation
}

type convexShapeRepresentation struct {
	lsBSphere shape3d.Sphere
	wsBSphere shape3d.Sphere

	lsTransform shape3d.Transform
	wsTransform shape3d.Transform

	kind       convexKind
	points     []dprec.Vec3
	skinRadius float64
}

func (s *convexShapeRepresentation) update(parentTransform shape3d.Transform) {
	s.wsBSphere = shape3d.TransformedSphere(s.lsBSphere, parentTransform)

	s.wsTransform = shape3d.ChainedTransform(
		parentTransform,
		s.lsTransform,
	)
}

func (s *convexShapeRepresentation) boundingSphere() shape3d.Sphere {
	return s.wsBSphere
}

func (s *convexShapeRepresentation) gjkShape() gjk3d.Shape {
	return gjk3d.Shape{
		Position:   s.wsTransform.Translation,
		Rotation:   s.wsTransform.Rotation,
		Points:     s.points,
		SkinRadius: s.skinRadius,
	}
}

func (s *convexShapeRepresentation) toSphere() shape3d.Sphere {
	return shape3d.Sphere{
		Center: s.wsTransform.Translation,
		Radius: s.skinRadius,
	}
}

func (s *convexShapeRepresentation) toBox() shape3d.Box {
	var halfWidth, halfHeight, halfLength float64
	for _, point := range s.points {
		halfWidth = max(halfWidth, point.X)
		halfHeight = max(halfHeight, point.Y)
		halfLength = max(halfLength, point.Z)
	}
	return shape3d.Box{
		Center:     s.wsTransform.Translation,
		Rotation:   s.wsTransform.Rotation,
		HalfWidth:  halfWidth,
		HalfHeight: halfHeight,
		HalfLength: halfLength,
	}
}

type convexKind uint32

const (
	convexKindSphere convexKind = iota
	convexKindBox
	convexKindCapsule
	convexKindConvexHull
)
