package spatial

import "github.com/mokiat/gomath/dprec"

// HexahedronRegion represents a region in space that is defined by six
// clipping planes pointing inwards.
// It is possible to use the HexahedronRegion to define a region through
// fewer planes by repeating one of the plane definitions one or more times.
type HexahedronRegion [6]Plane

// ProjectionRegion creates a new HexahedronRegion using the specified
// projection matrix. The provided matrix could be a transformed projection
// matrix (e.g. projection * view).
func ProjectionRegion(matrix dprec.Mat4) HexahedronRegion {
	return HexahedronRegion{
		// clipping planes are ordered in a way that they may reject as fast as possible
		Plane(dprec.Vec4Sum(matrix.Row4(), matrix.Row1())).Normalized(),  // left
		Plane(dprec.Vec4Diff(matrix.Row4(), matrix.Row1())).Normalized(), // right
		Plane(dprec.Vec4Diff(matrix.Row4(), matrix.Row3())).Normalized(), // far
		Plane(dprec.Vec4Sum(matrix.Row4(), matrix.Row2())).Normalized(),  // bottom
		Plane(dprec.Vec4Diff(matrix.Row4(), matrix.Row2())).Normalized(), // top
		Plane(dprec.Vec4Sum(matrix.Row4(), matrix.Row3())).Normalized(),  // near
	}
}

// CuboidRegion creates a new HexahedronRegion that represents a Cuboid shape
// using the specified position and size.
func CuboidRegion(position, size dprec.Vec3) HexahedronRegion {
	halfSize := dprec.Vec3Prod(size, 0.5)
	return HexahedronRegion{
		Plane(dprec.NewVec4(1, 0, 0, -(position.X - halfSize.X))),
		Plane(dprec.NewVec4(-1, 0, 0, (position.X + halfSize.X))),
		Plane(dprec.NewVec4(0, 1, 0, -(position.Y - halfSize.Y))),
		Plane(dprec.NewVec4(0, -1, 0, (position.Y + halfSize.Y))),
		Plane(dprec.NewVec4(0, 0, 1, -(position.Z - halfSize.Z))),
		Plane(dprec.NewVec4(0, 0, -1, (position.Z + halfSize.Z))),
	}
}

// SphericalRegion represents a region in space that is defined by a position
// and a radius around that position.
type SphericalRegion struct {
	Position dprec.Vec3
	Radius   float64
}
