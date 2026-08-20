package physics

import (
	"github.com/mokiat/gomath/dprec"
)

// AirfoilSolverConfig holds the parameters with which an [AirfoilSolver]
// is configured, either through [NewAirfoilSolver] or
// [AirfoilSolver.Configure].
type AirfoilSolverConfig struct {

	// RelativePosition is the body-local-space offset, relative to the
	// target's center of mass, at which the airfoil is mounted. The lift
	// force is applied at this point, so an offset airfoil induces torque
	// on the target in addition to accelerating it.
	RelativePosition dprec.Vec3

	// RelativeRotation is the body-local-space orientation of the airfoil,
	// relative to the target's rotation. See [AirfoilSolver] for the axis
	// convention that this orientation establishes.
	//
	// It must be a unit quaternion. Note that the zero value of this field
	// is the zero quaternion, which is not a valid rotation; use
	// [dprec.IdentityQuat] for an airfoil that is aligned with the target
	// itself.
	RelativeRotation dprec.Quat

	// SurfaceArea is the area of the airfoil, in square meters, that the
	// lift force is scaled by. Negative values are clamped to zero.
	SurfaceArea float64

	// StallAngle is the angle of attack at which the airfoil produces its
	// peak lift and past which it starts to stall. Negative values are
	// clamped to zero, which disables the pre-stall lift bump entirely and
	// leaves only the flat-plate behavior described on [AirfoilSolver].
	StallAngle dprec.Angle

	// LiftCoefficient is the dimensionless coefficient that scales the
	// lift produced by the airfoil. Negative values are clamped to zero.
	LiftCoefficient float64
}

// AirfoilSolver is an [AccelerationSolver] that models a lift-producing
// surface - a wing, a tail plane, a rudder - mounted at a fixed position
// and orientation on its target body.
//
// The airfoil has its own coordinate frame, established by
// [AirfoilSolver.RelativeRotation] on top of the target's rotation, in
// which the chord line runs along Z, the span runs along X, and lift acts
// along Y. Forward flight is along positive Z, which places the oncoming
// wind along negative Z and puts the airfoil at a zero angle of attack.
// The profile is symmetric, so a zero angle of attack produces no lift,
// and inverting the airfoil inverts the lift.
//
// The lift force is applied at [AirfoilSolver.RelativePosition] rather
// than at the center of mass, so an airfoil mounted off-center also
// induces torque. The oncoming wind is sampled at that same mounted
// position, and therefore includes the velocity the target's own rotation
// imparts there. That is what makes a pair of offset airfoils damp the
// target's rotation rather than merely turn it.
//
// The force always acts along the airfoil's Y axis, which is fixed to the
// airfoil rather than to the airflow. Only at a zero angle of attack is
// that axis perpendicular to the oncoming wind, making the force pure
// lift; as the angle of attack grows, the axis tilts back relative to the
// wind and an increasing share of the force opposes the direction of
// travel. The airfoil therefore produces drag that grows with the angle
// of attack, and keeps producing it once stalled, without drag having to
// be modeled as a separate term.
//
// What is not covered is the parasitic drag that a real airfoil produces
// even at a zero angle of attack, where the force is entirely
// perpendicular to the wind. That contribution is small enough to be
// folded into a drag contributor on the body itself, rather than being
// accounted for per airfoil.
//
// An AirfoilSolver must be configured, either through [NewAirfoilSolver]
// or [AirfoilSolver.Configure], before being registered with a [Scene]
// through [BodyAcceleratorView.Create].
type AirfoilSolver struct {
	relativePosition dprec.Vec3
	relativeRotation dprec.Quat
	surfaceArea      float64
	stallAngle       dprec.Angle
	liftCoefficient  float64
}

var _ AccelerationSolver = (*AirfoilSolver)(nil)

// NewAirfoilSolver creates a new [AirfoilSolver] configured according to
// config.
func NewAirfoilSolver(config AirfoilSolverConfig) *AirfoilSolver {
	result := &AirfoilSolver{}
	result.Configure(config)
	return result
}

// Configure configures this solver according to config. Negative
// SurfaceArea, StallAngle and LiftCoefficient values are clamped to zero,
// as with the respective setters.
//
// Configure must be called before this solver is registered with a
// [Scene] through [BodyAcceleratorView.Create]. Unlike [NewAirfoilSolver],
// it can be called on an already-allocated solver, which allows solvers to
// be cached (e.g. in a slice) and configured on demand.
func (s *AirfoilSolver) Configure(config AirfoilSolverConfig) {
	s.relativePosition = config.RelativePosition
	s.relativeRotation = config.RelativeRotation
	s.surfaceArea = max(0.0, config.SurfaceArea)
	s.stallAngle = max(0.0, config.StallAngle)
	s.liftCoefficient = max(0.0, config.LiftCoefficient)
}

// RelativePosition returns the body-local-space offset, relative to the
// target's center of mass, at which the airfoil is mounted.
func (s *AirfoilSolver) RelativePosition() dprec.Vec3 {
	return s.relativePosition
}

// SetRelativePosition changes the body-local-space offset, relative to the
// target's center of mass, at which the airfoil is mounted.
//
// It returns the solver itself, so that calls can be chained.
func (s *AirfoilSolver) SetRelativePosition(position dprec.Vec3) *AirfoilSolver {
	s.relativePosition = position
	return s
}

// RelativeRotation returns the body-local-space orientation of the
// airfoil, relative to the target's rotation.
func (s *AirfoilSolver) RelativeRotation() dprec.Quat {
	return s.relativeRotation
}

// SetRelativeRotation changes the body-local-space orientation of the
// airfoil, relative to the target's rotation. The provided rotation must
// be a unit quaternion.
//
// It returns the solver itself, so that calls can be chained.
func (s *AirfoilSolver) SetRelativeRotation(rotation dprec.Quat) *AirfoilSolver {
	s.relativeRotation = rotation
	return s
}

// SurfaceArea returns the area of the airfoil, in square meters, that the
// lift force is scaled by.
func (s *AirfoilSolver) SurfaceArea() float64 {
	return s.surfaceArea
}

// SetSurfaceArea changes the area of the airfoil, in square meters, that
// the lift force is scaled by. Negative values are clamped to zero.
//
// It returns the solver itself, so that calls can be chained.
func (s *AirfoilSolver) SetSurfaceArea(area float64) *AirfoilSolver {
	s.surfaceArea = max(0.0, area)
	return s
}

// StallAngle returns the angle of attack at which the airfoil produces
// its peak lift and past which it starts to stall.
func (s *AirfoilSolver) StallAngle() dprec.Angle {
	return s.stallAngle
}

// SetStallAngle changes the angle of attack at which the airfoil produces
// its peak lift and past which it starts to stall. Negative values are
// clamped to zero, which disables the pre-stall lift bump entirely.
//
// It returns the solver itself, so that calls can be chained.
func (s *AirfoilSolver) SetStallAngle(angle dprec.Angle) *AirfoilSolver {
	s.stallAngle = max(0.0, angle)
	return s
}

// LiftCoefficient returns the dimensionless coefficient that scales the
// lift produced by the airfoil.
func (s *AirfoilSolver) LiftCoefficient() float64 {
	return s.liftCoefficient
}

// SetLiftCoefficient changes the dimensionless coefficient that scales the
// lift produced by the airfoil. Negative values are clamped to zero.
//
// It returns the solver itself, so that calls can be chained.
func (s *AirfoilSolver) SetLiftCoefficient(coefficient float64) *AirfoilSolver {
	s.liftCoefficient = max(0.0, coefficient)
	return s
}

// ApplyAcceleration implements [AccelerationSolver.ApplyAcceleration].
//
// It determines the wind that the airfoil experiences at its mounted
// position, relative to the medium, and converts it into a force along
// the airfoil's Y axis, applied at that same mounted position.
//
// The wind is split into the component that flows along the chord line
// and the component that flows along the span, each of which contributes
// its own force. The chordwise component is the one that behaves like a
// wing, in that it stalls; the spanwise component only ever produces the
// flat-plate force described on [AirfoilSolver.localLiftCoefficient],
// since an airfoil is not shaped to exploit air flowing along its span.
//
// If the airfoil is at rest relative to the medium, it does nothing. Wind
// that flows purely along the Y axis, having no chordwise or spanwise
// component to split, is treated as flowing along the chord line.
func (s *AirfoilSolver) ApplyAcceleration(ctx AccelerationContext) {
	bodyRotation := ctx.Target.Rotation()

	airfoilOffsetWS := dprec.QuatVec3Rotation(bodyRotation, s.relativePosition)
	airfoilRotationWS := dprec.QuatProd(bodyRotation, s.relativeRotation)

	airfoilVelocity := dprec.Vec3Sum(
		ctx.Target.LinearVelocity(),
		dprec.Vec3Cross(
			ctx.Target.AngularVelocity(),
			airfoilOffsetWS,
		),
	)

	relWindVelocity := dprec.Vec3Diff(ctx.MediumVelocity, airfoilVelocity)
	relWindVelocityLng := relWindVelocity.Length()
	if relWindVelocityLng < Epsilon {
		return // no significant wind
	}

	basisX := airfoilRotationWS.OrientationX()
	basisY := airfoilRotationWS.OrientationY()
	basisZ := airfoilRotationWS.OrientationZ()

	windX := dprec.Vec3Dot(relWindVelocity, basisX)
	windY := dprec.Vec3Dot(relWindVelocity, basisY)
	windZ := dprec.Vec3Dot(relWindVelocity, basisZ)

	var (
		directAmount  float64
		lateralAmount float64
	)
	planarWindVelocity := dprec.NewVec3(windX, 0.0, windZ)
	if planarLng := planarWindVelocity.Length(); planarLng > Epsilon {
		directAmount = dprec.Abs(planarWindVelocity.Z / planarLng)
		lateralAmount = dprec.Abs(planarWindVelocity.X / planarLng)
	} else {
		directAmount = 1.0
		lateralAmount = 0.0
	}

	// Wind along the chord line (direct).
	if directAmount > Epsilon {
		effWindVelocity := relWindVelocityLng * directAmount
		angleOfAttack := dprec.Atan2(windY, -windZ)

		coef := s.localLiftCoefficient(angleOfAttack, true)
		magnitude := coef * 0.5 * ctx.MediumDensity * dprec.Sqr(effWindVelocity) * s.surfaceArea

		ctx.Target.ApplyOffsetForce(airfoilOffsetWS, dprec.Vec3Prod(basisY, magnitude))
	}

	// Wind along the span (lateral).
	if lateralAmount > Epsilon {
		effWindVelocity := relWindVelocityLng * lateralAmount
		angleOfAttack := dprec.Atan2(windY, dprec.Abs(windX)) // keep symmetric

		coef := s.localLiftCoefficient(angleOfAttack, false)
		magnitude := coef * 0.5 * ctx.MediumDensity * dprec.Sqr(effWindVelocity) * s.surfaceArea

		ctx.Target.ApplyOffsetForce(airfoilOffsetWS, dprec.Vec3Prod(basisY, magnitude))
	}
}

// localLiftCoefficient returns the coefficient that scales the force the
// airfoil produces along its Y axis at the specified angle of attack.
//
// The coefficient is odd-symmetric about a zero angle of attack, so an
// airfoil that is turned upside down produces the same force in the
// opposite direction.
//
// Every airfoil produces the flat-plate force that any inclined surface
// does, which grows with the sine of the angle of attack and is largest
// when the surface meets the wind broadside. When isDirect is set, and
// the airfoil has been given a stall angle, the profile additionally
// produces the extra force that its shape is meant to generate: a bump
// that grows from zero at a zero angle of attack, peaks at the stall
// angle, and decays back to zero at twice the stall angle, past which the
// airfoil is fully stalled and only the flat-plate force remains.
func (s *AirfoilSolver) localLiftCoefficient(angle dprec.Angle, isDirect bool) float64 {
	if angle < 0.0 {
		return -s.localLiftCoefficient(-angle, isDirect) // flipped symmetric
	}
	degrees := angle.Degrees()
	result := (s.liftCoefficient / 2.0) * dprec.Sin(angle)
	if isDirect && (s.stallAngle > 0.0) { // add the lift coefficient bump prior to stall
		stallDegrees := s.stallAngle.Degrees()
		result += (s.liftCoefficient / 2.0) * max(0.0, degrees*(2.0*stallDegrees-degrees)) / dprec.Sqr(stallDegrees)
	}
	return result
}
