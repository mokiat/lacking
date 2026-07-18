package physics

import "github.com/mokiat/gomath/dprec"

// DiagonalMomentOfInertia returns a moment of inertia tensor with the
// specified principal moments along the local X, Y, and Z axes. This
// assumes the axes are aligned with the object's principal axes, so the
// off-diagonal products of inertia are zero.
func DiagonalMomentOfInertia(xx, yy, zz float64) dprec.Mat3 {
	return dprec.NewMat3(
		xx, 0.0, 0.0,
		0.0, yy, 0.0,
		0.0, 0.0, zz,
	)
}

// SymmetricMomentOfInertia returns a moment of inertia
// tensor that represents a symmetric object across all
// axis.
func SymmetricMomentOfInertia(value float64) dprec.Mat3 {
	return DiagonalMomentOfInertia(value, value, value)
}

// SolidSphereMomentOfInertia returns the moment of inertia of a solid
// sphere with the specified mass and radius.
func SolidSphereMomentOfInertia(mass, radius float64) dprec.Mat3 {
	return SymmetricMomentOfInertia(mass * radius * radius * (2.0 / 5.0))
}

// HollowSphereMomentOfInertia returns the moment of inertia of a hollow
// sphere with the specified mass and radius.
func HollowSphereMomentOfInertia(mass, radius float64) dprec.Mat3 {
	return SymmetricMomentOfInertia(mass * radius * radius * (2.0 / 3.0))
}

// SolidBoxMomentOfInertia returns the moment of inertia of a solid box
// (rectangular cuboid) with the specified mass and dimensions. The width,
// height, and length are the full sizes of the box along its local X, Y,
// and Z axes respectively.
func SolidBoxMomentOfInertia(mass, width, height, length float64) dprec.Mat3 {
	factor := mass / 12.0
	return DiagonalMomentOfInertia(
		factor*(height*height+length*length),
		factor*(width*width+length*length),
		factor*(width*width+height*height),
	)
}

// HollowBoxMomentOfInertia returns the moment of inertia of a hollow box
// (rectangular cuboid shell) with the specified mass and dimensions. The
// width, height, and length are the full sizes of the box along its local
// X, Y, and Z axes respectively.
func HollowBoxMomentOfInertia(mass, width, height, length float64) dprec.Mat3 {
	area := 2.0 * (width*height + height*length + length*width)
	fraction := mass / area

	return DiagonalMomentOfInertia(
		fraction*hollowBoxAxisMoment(width, height, length),
		fraction*hollowBoxAxisMoment(height, length, width),
		fraction*hollowBoxAxisMoment(length, width, height),
	)
}

// hollowBoxAxisMoment returns the moment of inertia of a unit-density
// hollow box around the axis with size a, where n1 and n2 are the sizes
// along the two remaining axes.
func hollowBoxAxisMoment(a, n1, n2 float64) float64 {
	n1Sqr := n1 * n1
	n2Sqr := n2 * n2
	oneSixth := 1.0 / 6.0
	oneHalf := 1.0 / 2.0

	return n1*n2*(n1Sqr+n2Sqr)*oneSixth + // orhtogonal faces (n1 x n2 plane)
		a*n1*(n1Sqr*oneSixth+n2Sqr*oneHalf) + // parallel faces (a x n1 plane)
		a*n2*(n1Sqr*oneHalf+n2Sqr*oneSixth) // parallel faces (a x n2 plane)
}
