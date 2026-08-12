package physics

import "github.com/mokiat/gomath/dprec"

// DragSolverConfig holds the parameters with which a [DragSolver] is
// configured, either through [NewDragSolver] or [DragSolver.Configure].
type DragSolverConfig struct {

	// RelativePosition is the body-local-space offset, relative to the
	// target's center of mass, at which the box is centered. The drag
	// force is applied at this point, so an offset box induces torque on
	// the target in addition to decelerating it.
	RelativePosition dprec.Vec3

	// RelativeRotation is the body-local-space orientation of the box,
	// relative to the target's rotation. See [DragSolver] for the axis
	// convention that this orientation establishes.
	//
	// It must be a unit quaternion, as must every rotation handed to the
	// engine.
	RelativeRotation dprec.Quat

	// Width is the extent of the box along its X axis, in meters.
	// Negative values are clamped to zero.
	Width float64

	// Height is the extent of the box along its Y axis, in meters.
	// Negative values are clamped to zero.
	Height float64

	// Length is the extent of the box along its Z axis, in meters.
	// Negative values are clamped to zero.
	Length float64

	// DragCoefficientX is the dimensionless drag coefficient for wind that
	// flows along the box's X axis, and hence acts on its Height by Length
	// cross-section. Negative values are clamped to zero.
	DragCoefficientX float64

	// DragCoefficientY is the dimensionless drag coefficient for wind that
	// flows along the box's Y axis, and hence acts on its Width by Length
	// cross-section. Negative values are clamped to zero.
	DragCoefficientY float64

	// DragCoefficientZ is the dimensionless drag coefficient for wind that
	// flows along the box's Z axis, and hence acts on its Width by Height
	// cross-section. Negative values are clamped to zero.
	DragCoefficientZ float64
}

// DragSolver is an [AccelerationSolver] that models the aerodynamic drag
// of a box-shaped volume mounted at a fixed position and orientation on
// its target body.
//
// The box has its own coordinate frame, established by
// [DragSolver.RelativeRotation] on top of the target's rotation, in which
// [DragSolver.Width] is the extent along X, [DragSolver.Height] the
// extent along Y, and [DragSolver.Length] the extent along Z. The box is
// centered on [DragSolver.RelativePosition].
//
// Each of the three axes resists the medium independently, in proportion
// to the square of the wind component along that axis, the
// cross-sectional area perpendicular to it, and the drag coefficient
// configured for it. A box can therefore be made more streamlined along
// one axis than another, and the resulting force is in general not
// antiparallel to the wind.
//
// Both translation and rotation are resisted. The wind is sampled at the
// box's mounted position and the force is applied there, so a box mounted
// off-center also induces torque; on top of that, a couple opposes the
// target's rotation about the box's own center. A box that is centered on
// the target's center of mass is subject to the couple alone.
//
// Only pressure drag is modeled, never skin friction, so a face that the
// medium merely slides along contributes nothing. A thin plate spun about
// its own normal is the degenerate case of this, and meets almost no
// resistance.
//
// Only drag is modeled. See [AirfoilSolver] for a surface that produces
// lift, which is meant to be combined with this solver rather than
// replaced by it.
//
// A DragSolver must be configured, either through [NewDragSolver] or
// [DragSolver.Configure], before being registered with a [Scene] through
// [BodyAcceleratorView.Create].
type DragSolver struct {
	relativePosition dprec.Vec3
	relativeRotation dprec.Quat
	width            float64
	height           float64
	length           float64
	dragCoefficientX float64
	dragCoefficientY float64
	dragCoefficientZ float64
}

var _ AccelerationSolver = (*DragSolver)(nil)

// NewDragSolver creates a new [DragSolver] configured according to config.
func NewDragSolver(config DragSolverConfig) *DragSolver {
	result := &DragSolver{}
	result.Configure(config)
	return result
}

// Configure configures this solver according to config. Negative Width,
// Height, Length and drag coefficient values are clamped to zero, as with
// the respective setters.
//
// Configure must be called before this solver is registered with a
// [Scene] through [BodyAcceleratorView.Create]. Unlike [NewDragSolver], it
// can be called on an already-allocated solver, which allows solvers to be
// cached (e.g. in a slice) and configured on demand.
func (s *DragSolver) Configure(config DragSolverConfig) {
	s.relativePosition = config.RelativePosition
	s.relativeRotation = config.RelativeRotation
	s.width = max(0.0, config.Width)
	s.height = max(0.0, config.Height)
	s.length = max(0.0, config.Length)
	s.dragCoefficientX = max(0.0, config.DragCoefficientX)
	s.dragCoefficientY = max(0.0, config.DragCoefficientY)
	s.dragCoefficientZ = max(0.0, config.DragCoefficientZ)
}

// RelativePosition returns the body-local-space offset, relative to the
// target's center of mass, at which the box is centered.
func (s *DragSolver) RelativePosition() dprec.Vec3 {
	return s.relativePosition
}

// SetRelativePosition changes the body-local-space offset, relative to the
// target's center of mass, at which the box is centered.
//
// It returns the solver itself, so that calls can be chained.
func (s *DragSolver) SetRelativePosition(position dprec.Vec3) *DragSolver {
	s.relativePosition = position
	return s
}

// RelativeRotation returns the body-local-space orientation of the box,
// relative to the target's rotation.
func (s *DragSolver) RelativeRotation() dprec.Quat {
	return s.relativeRotation
}

// SetRelativeRotation changes the body-local-space orientation of the box,
// relative to the target's rotation. The provided rotation must be a unit
// quaternion.
//
// It returns the solver itself, so that calls can be chained.
func (s *DragSolver) SetRelativeRotation(rotation dprec.Quat) *DragSolver {
	s.relativeRotation = rotation
	return s
}

// Width returns the extent of the box along its X axis, in meters.
func (s *DragSolver) Width() float64 {
	return s.width
}

// SetWidth changes the extent of the box along its X axis, in meters.
// Negative values are clamped to zero.
//
// It returns the solver itself, so that calls can be chained.
func (s *DragSolver) SetWidth(width float64) *DragSolver {
	s.width = max(0.0, width)
	return s
}

// Height returns the extent of the box along its Y axis, in meters.
func (s *DragSolver) Height() float64 {
	return s.height
}

// SetHeight changes the extent of the box along its Y axis, in meters.
// Negative values are clamped to zero.
//
// It returns the solver itself, so that calls can be chained.
func (s *DragSolver) SetHeight(height float64) *DragSolver {
	s.height = max(0.0, height)
	return s
}

// Length returns the extent of the box along its Z axis, in meters.
func (s *DragSolver) Length() float64 {
	return s.length
}

// SetLength changes the extent of the box along its Z axis, in meters.
// Negative values are clamped to zero.
//
// It returns the solver itself, so that calls can be chained.
func (s *DragSolver) SetLength(length float64) *DragSolver {
	s.length = max(0.0, length)
	return s
}

// DragCoefficientX returns the dimensionless drag coefficient for wind
// that flows along the box's X axis.
func (s *DragSolver) DragCoefficientX() float64 {
	return s.dragCoefficientX
}

// SetDragCoefficientX changes the dimensionless drag coefficient for wind
// that flows along the box's X axis. Negative values are clamped to zero.
//
// It returns the solver itself, so that calls can be chained.
func (s *DragSolver) SetDragCoefficientX(coefficient float64) *DragSolver {
	s.dragCoefficientX = max(0.0, coefficient)
	return s
}

// DragCoefficientY returns the dimensionless drag coefficient for wind
// that flows along the box's Y axis.
func (s *DragSolver) DragCoefficientY() float64 {
	return s.dragCoefficientY
}

// SetDragCoefficientY changes the dimensionless drag coefficient for wind
// that flows along the box's Y axis. Negative values are clamped to zero.
//
// It returns the solver itself, so that calls can be chained.
func (s *DragSolver) SetDragCoefficientY(coefficient float64) *DragSolver {
	s.dragCoefficientY = max(0.0, coefficient)
	return s
}

// DragCoefficientZ returns the dimensionless drag coefficient for wind
// that flows along the box's Z axis.
func (s *DragSolver) DragCoefficientZ() float64 {
	return s.dragCoefficientZ
}

// SetDragCoefficientZ changes the dimensionless drag coefficient for wind
// that flows along the box's Z axis. Negative values are clamped to zero.
//
// It returns the solver itself, so that calls can be chained.
func (s *DragSolver) SetDragCoefficientZ(coefficient float64) *DragSolver {
	s.dragCoefficientZ = max(0.0, coefficient)
	return s
}

// ApplyAcceleration implements [AccelerationSolver.ApplyAcceleration].
//
// It applies both the force that resists the target's motion through the
// medium and the couple that resists the target's rotation.
//
// The force is derived from the wind that the box experiences at its
// mounted position, which includes the velocity that the target's own
// rotation imparts there, and is applied at that same position. It is
// resolved onto the box's three axes, each of which contributes
// independently. If the box is at rest relative to the medium, no force is
// applied.
//
// The couple is derived from the target's angular velocity alone, since a
// medium that moves uniformly carries no rotation of its own to be
// measured against. It is applied whether or not the target is also
// translating.
func (s *DragSolver) ApplyAcceleration(ctx AccelerationContext) {
	bodyRotation := ctx.Target.Rotation()

	boxOffsetWS := dprec.QuatVec3Rotation(bodyRotation, s.relativePosition)
	boxRotationWS := dprec.QuatProd(bodyRotation, s.relativeRotation)

	basisX := boxRotationWS.OrientationX()
	basisY := boxRotationWS.OrientationY()
	basisZ := boxRotationWS.OrientationZ()

	{ // Linear drag
		boxVelocity := dprec.Vec3Sum(
			ctx.Target.LinearVelocity(),
			dprec.Vec3Cross(
				ctx.Target.AngularVelocity(),
				boxOffsetWS,
			),
		)

		relWindVelocity := dprec.Vec3Diff(ctx.MediumVelocity, boxVelocity)
		relWindVelocityLng := relWindVelocity.Length()
		if relWindVelocityLng > Epsilon {
			windX := dprec.Vec3Dot(relWindVelocity, basisX)
			windY := dprec.Vec3Dot(relWindVelocity, basisY)
			windZ := dprec.Vec3Dot(relWindVelocity, basisZ)

			magnitudeX := s.dragCoefficientX * 0.5 * ctx.MediumDensity * windX * dprec.Abs(windX) * (s.height * s.length)
			magnitudeY := s.dragCoefficientY * 0.5 * ctx.MediumDensity * windY * dprec.Abs(windY) * (s.width * s.length)
			magnitudeZ := s.dragCoefficientZ * 0.5 * ctx.MediumDensity * windZ * dprec.Abs(windZ) * (s.width * s.height)

			ctx.Target.ApplyOffsetForce(boxOffsetWS, dprec.Vec3Sum(
				dprec.Vec3Prod(basisX, magnitudeX),
				dprec.Vec3Sum(
					dprec.Vec3Prod(basisY, magnitudeY),
					dprec.Vec3Prod(basisZ, magnitudeZ),
				),
			))
		}
	}

	{ // Angular drag
		angularVelocity := ctx.Target.AngularVelocity()
		angularWindX := dprec.Vec3Dot(angularVelocity, basisX)
		angularWindY := dprec.Vec3Dot(angularVelocity, basisY)
		angularWindZ := dprec.Vec3Dot(angularVelocity, basisZ)

		halfWidth := s.width * 0.5
		halfHeight := s.height * 0.5
		halfLength := s.length * 0.5

		factorWidth := dprec.Sqr(dprec.Sqr(halfWidth))
		factorHeight := dprec.Sqr(dprec.Sqr(halfHeight))
		factorLength := dprec.Sqr(dprec.Sqr(halfLength))

		torque := dprec.Vec3{
			X: -angularWindX * dprec.Abs(angularWindX) * halfWidth * (s.dragCoefficientY*factorLength + s.dragCoefficientZ*factorHeight),
			Y: -angularWindY * dprec.Abs(angularWindY) * halfHeight * (s.dragCoefficientZ*factorWidth + s.dragCoefficientX*factorLength),
			Z: -angularWindZ * dprec.Abs(angularWindZ) * halfLength * (s.dragCoefficientX*factorHeight + s.dragCoefficientY*factorWidth),
		}
		torque = dprec.Vec3Prod(torque, ctx.MediumDensity)

		ctx.Target.ApplyTorque(dprec.QuatVec3Rotation(boxRotationWS, torque))
	}
}
