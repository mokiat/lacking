package query3d

import "github.com/mokiat/lacking/core/spatial/shape3d"

// Area represents the spatial area of an object in the 3D space.
type Area struct {
	x float64
	y float64
	z float64
	r float64
}

// AreaFromSphere creates an [Area] from the given sphere's center and radius.
func AreaFromSphere(sphere shape3d.Sphere) Area {
	return Area{
		x: sphere.Center.X,
		y: sphere.Center.Y,
		z: sphere.Center.Z,
		r: sphere.Radius,
	}
}

// AreaFromBox creates an [Area] from the given box, using its largest
// half-extent as the area radius.
func AreaFromBox(box shape3d.Box) Area {
	return Area{
		x: box.Center.X,
		y: box.Center.Y,
		z: box.Center.Z,
		r: max(box.HalfWidth, box.HalfHeight, box.HalfLength),
	}
}
