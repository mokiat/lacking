package easing

import "github.com/mokiat/gomath/dprec"

// HermiteSplineCoefficients returns the p0, v0, p1, v1 coefficients
// for a Hermite spline given the specified time and duration.
func HermiteSplineCoefficients(t, duration float64) (float64, float64, float64, float64) {
	invDuration := 1.0 / duration
	s1 := t * invDuration
	s2 := s1 * s1
	s3 := s1 * s2

	coefP0 := (2.0*s3 - 3.0*s2 + 1.0)
	coefV0 := (s3 - 2.0*s2 + s1) * duration
	coefP1 := (-2.0*s3 + 3.0*s2)
	coefV1 := (s3 - s2) * duration

	return coefP0, coefV0, coefP1, coefV1
}

// HermiteSplineDeriv1Coefficients returns the p0, v0, p1, v1 coefficients
// for the first-order derivative of a Hermite spline given the specified
// time and duration.
func HermiteSplineDeriv1Coefficients(t, duration float64) (float64, float64, float64, float64) {
	invDuration := 1.0 / duration
	s1 := t * invDuration
	s2 := s1 * s1

	coefP0 := (6.0*s2 - 6.0*s1) * invDuration
	coefV0 := (3.0*s2 - 4.0*s1 + 1.0)
	coefP1 := (-6.0*s2 + 6.0*s1) * invDuration
	coefV1 := (3.0*s2 - 2.0*s1)

	return coefP0, coefV0, coefP1, coefV1
}

// HermiteSplineDeriv2Coefficients returns the p0, v0, p1, v1 coefficients
// for the second-order derivative of a Hermite spline given the specified
// time and duration.
func HermiteSplineDeriv2Coefficients(t, duration float64) (float64, float64, float64, float64) {
	invDuration := 1.0 / duration
	invDuration2 := invDuration * invDuration
	s1 := t * invDuration

	coefP0 := (12.0*s1 - 6.0) * invDuration2
	coefV0 := (6.0*s1 - 4.0) * invDuration
	coefP1 := (-12.0*s1 + 6.0) * invDuration2
	coefV1 := (6.0*s1 - 2.0) * invDuration

	return coefP0, coefV0, coefP1, coefV1
}

// HermiteSpline1D returns the 1D Hermite spline value at the given time value,
// considering the specified position, velocity and duration parameters.
func HermiteSpline1D(p0, v0, p1, v1, t, duration float64) float64 {
	coefP0, coefV0, coefP1, coefV1 := HermiteSplineCoefficients(t, duration)
	return (coefP0 * p0) + (coefV0 * v0) + (coefP1 * p1) + (coefV1 * v1)
}

// HermiteSpline2D returns the 2D Hermite spline value at the given time value,
// considering the specified position, velocity and duration parameters.
func HermiteSpline2D(p0, v0, p1, v1 dprec.Vec2, t, duration float64) dprec.Vec2 {
	coefP0, coefV0, coefP1, coefV1 := HermiteSplineCoefficients(t, duration)
	return dprec.Vec2MultiSum(
		dprec.Vec2Prod(p0, coefP0),
		dprec.Vec2Prod(v0, coefV0),
		dprec.Vec2Prod(p1, coefP1),
		dprec.Vec2Prod(v1, coefV1),
	)
}

// HermiteSpline3D returns the 3D Hermite spline value at the given time value,
// considering the specified position, velocity and duration parameters.
func HermiteSpline3D(p0, v0, p1, v1 dprec.Vec3, t, duration float64) dprec.Vec3 {
	coefP0, coefV0, coefP1, coefV1 := HermiteSplineCoefficients(t, duration)
	return dprec.Vec3MultiSum(
		dprec.Vec3Prod(p0, coefP0),
		dprec.Vec3Prod(v0, coefV0),
		dprec.Vec3Prod(p1, coefP1),
		dprec.Vec3Prod(v1, coefV1),
	)
}

// HermiteSpline1DDeriv1 returns the 1D Hermite spline first-order derivative
// value at the given time value, considering the specified position,
// velocity and duration parameters.
func HermiteSpline1DDeriv1(p0, v0, p1, v1, t, duration float64) float64 {
	coefP0, coefV0, coefP1, coefV1 := HermiteSplineDeriv1Coefficients(t, duration)
	return (coefP0 * p0) + (coefV0 * v0) + (coefP1 * p1) + (coefV1 * v1)
}

// HermiteSpline2DDeriv1 returns the 2D Hermite spline first-order derivative
// value at the given time value, considering the specified position,
// velocity and duration parameters.
func HermiteSpline2DDeriv1(p0, v0, p1, v1 dprec.Vec2, t, duration float64) dprec.Vec2 {
	coefP0, coefV0, coefP1, coefV1 := HermiteSplineDeriv1Coefficients(t, duration)
	return dprec.Vec2MultiSum(
		dprec.Vec2Prod(p0, coefP0),
		dprec.Vec2Prod(v0, coefV0),
		dprec.Vec2Prod(p1, coefP1),
		dprec.Vec2Prod(v1, coefV1),
	)
}

// HermiteSpline3DDeriv1 returns the 3D Hermite spline first-order derivative
// value at the given time value, considering the specified position,
// velocity and duration parameters.
func HermiteSpline3DDeriv1(p0, v0, p1, v1 dprec.Vec3, t, duration float64) dprec.Vec3 {
	coefP0, coefV0, coefP1, coefV1 := HermiteSplineDeriv1Coefficients(t, duration)
	return dprec.Vec3MultiSum(
		dprec.Vec3Prod(p0, coefP0),
		dprec.Vec3Prod(v0, coefV0),
		dprec.Vec3Prod(p1, coefP1),
		dprec.Vec3Prod(v1, coefV1),
	)
}

// HermiteSpline1DDeriv2 returns the 1D Hermite spline second-order derivative
// value at the given time value, considering the specified position,
// velocity and duration parameters.
func HermiteSpline1DDeriv2(p0, v0, p1, v1, t, duration float64) float64 {
	coefP0, coefV0, coefP1, coefV1 := HermiteSplineDeriv2Coefficients(t, duration)
	return (coefP0 * p0) + (coefV0 * v0) + (coefP1 * p1) + (coefV1 * v1)
}

// HermiteSpline2DDeriv2 returns the 2D Hermite spline second-order derivative
// value at the given time value, considering the specified position,
// velocity and duration parameters.
func HermiteSpline2DDeriv2(p0, v0, p1, v1 dprec.Vec2, t, duration float64) dprec.Vec2 {
	coefP0, coefV0, coefP1, coefV1 := HermiteSplineDeriv2Coefficients(t, duration)
	return dprec.Vec2MultiSum(
		dprec.Vec2Prod(p0, coefP0),
		dprec.Vec2Prod(v0, coefV0),
		dprec.Vec2Prod(p1, coefP1),
		dprec.Vec2Prod(v1, coefV1),
	)
}

// HermiteSpline3DDeriv2 returns the 3D Hermite spline second-order derivative
// value at the given time value, considering the specified position,
// velocity and duration parameters.
func HermiteSpline3DDeriv2(p0, v0, p1, v1 dprec.Vec3, t, duration float64) dprec.Vec3 {
	coefP0, coefV0, coefP1, coefV1 := HermiteSplineDeriv2Coefficients(t, duration)
	return dprec.Vec3MultiSum(
		dprec.Vec3Prod(p0, coefP0),
		dprec.Vec3Prod(v0, coefV0),
		dprec.Vec3Prod(p1, coefP1),
		dprec.Vec3Prod(v1, coefV1),
	)
}
