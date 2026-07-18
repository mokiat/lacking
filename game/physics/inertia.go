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

// RotatedMomentOfInertia returns the specified moment of inertia tensor
// as observed from a frame in which the object is rotated by the
// specified rotation.
//
// This is needed when a part is mounted at an angle relative to the
// body that contains it, as is the case with a dihedral wing or a
// canted tail fin.
func RotatedMomentOfInertia(tensor dprec.Mat3, rotation dprec.Quat) dprec.Mat3 {
	basisX := rotation.OrientationX()
	basisY := rotation.OrientationY()
	basisZ := rotation.OrientationZ()

	matrix := dprec.NewMat3(
		basisX.X, basisY.X, basisZ.X,
		basisX.Y, basisY.Y, basisZ.Y,
		basisX.Z, basisY.Z, basisZ.Z,
	)

	return dprec.Mat3MultiProd(
		matrix,
		tensor,
		dprec.TransposedMat3(matrix),
	)
}

// OffsetMomentOfInertia returns the moment of inertia tensor of an
// object with the specified mass, as measured around a point that is
// displaced by the specified offset from the object's center of mass.
//
// The tensor passed in has to be measured around the object's own
// center of mass. This implements the parallel axis theorem and is
// needed when combining parts that are positioned away from the center
// of mass of the body that contains them, like the engines of an
// airplane.
//
// Note that the result is generally not diagonal, even if the input is.
func OffsetMomentOfInertia(tensor dprec.Mat3, mass float64, offset dprec.Vec3) dprec.Mat3 {
	offsetSqr := offset.SqrLength()
	offsetXX := offset.X * offset.X
	offsetXY := offset.X * offset.Y
	offsetXZ := offset.X * offset.Z
	offsetYY := offset.Y * offset.Y
	offsetYZ := offset.Y * offset.Z
	offsetZZ := offset.Z * offset.Z

	return dprec.NewMat3(
		tensor.M11+mass*(offsetSqr-offsetXX),
		tensor.M12-mass*offsetXY,
		tensor.M13-mass*offsetXZ,

		tensor.M21-mass*offsetXY,
		tensor.M22+mass*(offsetSqr-offsetYY),
		tensor.M23-mass*offsetYZ,

		tensor.M31-mass*offsetXZ,
		tensor.M32-mass*offsetYZ,
		tensor.M33+mass*(offsetSqr-offsetZZ),
	)
}

// MomentOfInertiaSum returns the sum of the two specified moment of
// inertia tensors.
//
// The tensors need to be expressed in the same coordinate frame and
// around the same reference point for the result to be meaningful. Use
// [RotatedMomentOfInertia] and [OffsetMomentOfInertia] to bring the
// tensor of each part into the frame of the containing body first.
func MomentOfInertiaSum(first, second dprec.Mat3) dprec.Mat3 {
	return dprec.NewMat3(
		first.M11+second.M11, first.M12+second.M12, first.M13+second.M13,
		first.M21+second.M21, first.M22+second.M22, first.M23+second.M23,
		first.M31+second.M31, first.M32+second.M32, first.M33+second.M33,
	)
}

// MomentOfInertiaMultiSum returns the sum of the specified moment of
// inertia tensors.
//
// The same coordinate frame and reference point requirements as with
// [MomentOfInertiaSum] apply.
func MomentOfInertiaMultiSum(first dprec.Mat3, others ...dprec.Mat3) dprec.Mat3 {
	result := first
	for _, other := range others {
		result = MomentOfInertiaSum(result, other)
	}
	return result
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
