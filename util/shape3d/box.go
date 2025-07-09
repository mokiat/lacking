package shape3d

import "github.com/mokiat/gomath/dprec"

func NewBox(position dprec.Vec3, rotation dprec.Quat, size dprec.Vec3) Box {
	return Box{
		Position:   position,
		Rotation:   rotation,
		HalfWidth:  size.X / 2.0,
		HalfHeight: size.Y / 2.0,
		HalfLength: size.Z / 2.0,
	}
}

func TransformedBox(source Box, transform Transform) Box {
	boxTransform := ChainedTransform(transform, Transform{
		Translation: source.Position,
		Rotation:    source.Rotation,
	})
	return Box{
		Position:   boxTransform.Translation,
		Rotation:   boxTransform.Rotation,
		HalfWidth:  source.HalfWidth,
		HalfHeight: source.HalfHeight,
		HalfLength: source.HalfLength,
	}
}

type Box struct {
	Position   dprec.Vec3
	Rotation   dprec.Quat
	HalfWidth  float64
	HalfHeight float64
	HalfLength float64
}
