package spatial

import "github.com/mokiat/gomath/sprec"

// HexahedronRegion represents a region in space that is defined by six
// clipping planes pointing inwards.
// It is possible to use the HexahedronRegion to define a region through
// fewer planes by repeating one of the plane definitions one or more times.
type HexahedronRegion [6]Plane

// ProjectionRegion creates a new HexahedronRegion using the specified
// projection matrix. The provided matrix could be a transformed projection
// matrix (e.g. projection * view).
func ProjectionRegion(matrix sprec.Mat4) HexahedronRegion {
	return HexahedronRegion{
		// clipping planes are ordered in a way that they may reject as fast as possible
		Plane(sprec.Vec4Sum(matrix.Row4(), matrix.Row1())).Normalized(),  // left
		Plane(sprec.Vec4Diff(matrix.Row4(), matrix.Row1())).Normalized(), // right
		Plane(sprec.Vec4Diff(matrix.Row4(), matrix.Row3())).Normalized(), // far
		Plane(sprec.Vec4Sum(matrix.Row4(), matrix.Row2())).Normalized(),  // bottom
		Plane(sprec.Vec4Diff(matrix.Row4(), matrix.Row2())).Normalized(), // top
		Plane(sprec.Vec4Sum(matrix.Row4(), matrix.Row3())).Normalized(),  // near
	}
}

// CuboidRegion creates a new HexahedronRegion that represents a Cuboid shape
// using the specified position and size.
func CuboidRegion(position, size sprec.Vec3) HexahedronRegion {
	halfSize := sprec.Vec3Prod(size, 0.5)
	return HexahedronRegion{
		Plane(sprec.NewVec4(1, 0, 0, -(position.X - halfSize.X))),
		Plane(sprec.NewVec4(-1, 0, 0, (position.X + halfSize.X))),
		Plane(sprec.NewVec4(0, 1, 0, -(position.Y - halfSize.Y))),
		Plane(sprec.NewVec4(0, -1, 0, (position.Y + halfSize.Y))),
		Plane(sprec.NewVec4(0, 0, 1, -(position.Z - halfSize.Z))),
		Plane(sprec.NewVec4(0, 0, -1, (position.Z + halfSize.Z))),
	}
}

// SphericalRegion represents a region in space that is defined by a position
// and a radius around that position.
type SphericalRegion struct {
	Position sprec.Vec3
	Radius   float32
}
