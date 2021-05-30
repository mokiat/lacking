package physics

import (
	"github.com/mokiat/gomath/sprec"
	"github.com/mokiat/lacking/shape"
)

const (
	defaultImpulseIterations = 100
	defaultNudgeIterations   = 100
)

func newScene(stepSeconds float32) *Scene {
	return &Scene{
		stepSeconds:       stepSeconds,
		impulseIterations: defaultImpulseIterations,
		nudgeIterations:   defaultNudgeIterations,

		maxAcceleration:        100.0, // TODO: Measure something reasonable
		maxAngularAcceleration: 100.0, // TODO: Measure something reasonable
		maxVelocity:            300.0,
		maxAngularVelocity:     300.0, // TODO: Measure something reasonable

		intersectionSet: shape.NewIntersectionResultSet(128),

		gravity:      sprec.NewVec3(0.0, -9.8, 0.0),
		windVelocity: sprec.NewVec3(0.0, 0.0, 0.0),
		windDensity:  1.2,
	}
}

// Scene represents a physics scene that contains
// a number of bodies that are independent on any
// bodies managed by other scene objects.
type Scene struct {
	stepSeconds            float32
	impulseIterations      int
	nudgeIterations        int
	maxAcceleration        float32
	maxAngularAcceleration float32
	maxVelocity            float32
	maxAngularVelocity     float32

	firstBody *Body
	lastBody  *Body
	cacheBody *Body

	firstConstraint *Constraint
	lastConstraint  *Constraint
	cacheConstraint *Constraint

	collisionConstraints []*Constraint
	collisionSolvers     []groundCollisionSolver
	intersectionSet      *shape.IntersectionResultSet

	gravity      sprec.Vec3
	windVelocity sprec.Vec3
	windDensity  float32
}

// Gravity returns the gravity acceleration.
func (s *Scene) Gravity() sprec.Vec3 {
	return s.gravity
}

// SetGravity changes the gravity acceleration.
func (s *Scene) SetGravity(gravity sprec.Vec3) {
	s.gravity = gravity
}

// WindVelocity returns the wind speed.
func (s *Scene) WindVelocity() sprec.Vec3 {
	return s.windVelocity
}

// SetWindVelocity sets the wind speed.
func (s *Scene) SetWindVelocity(velocity sprec.Vec3) {
	s.windVelocity = velocity
}

// WindDensity returns the wind density.
func (s *Scene) WindDensity() float32 {
	return s.windDensity
}

// SetWindDensity changes the wind density.
func (s *Scene) SetWindDensity(density float32) {
	s.windDensity = density
}

// CreateBody creates a new physics body and places
// it within this scene.
func (s *Scene) CreateBody() *Body {
	var body *Body
	if s.cacheBody != nil {
		body = s.cacheBody
		s.cacheBody = s.cacheBody.next
	} else {
		body = &Body{}
	}
	body.scene = s
	body.prev = nil
	body.next = nil
	s.appendBody(body)
	return body
}

// CreateSingleBodyConstraint creates a new physics constraint that acts on
// a single body and enables it for this scene.
func (s *Scene) CreateSingleBodyConstraint(solver ConstraintSolver, body *Body) *Constraint {
	var constraint *Constraint
	if s.cacheConstraint != nil {
		constraint = s.cacheConstraint
		s.cacheConstraint = s.cacheConstraint.next
	} else {
		constraint = &Constraint{}
	}
	constraint.scene = s
	constraint.solver = solver
	constraint.prev = nil
	constraint.next = nil
	constraint.enabled = true
	constraint.primary = body
	constraint.secondary = nil
	s.appendConstraint(constraint)
	return constraint
}

// CreateDoubleBodyConstraint creates a new physics constraint that acts on
// two bodies and enables it for this scene.
func (s *Scene) CreateDoubleBodyConstraint(solver ConstraintSolver, primary, secondary *Body) *Constraint {
	var constraint *Constraint
	if s.cacheConstraint != nil {
		constraint = s.cacheConstraint
		s.cacheConstraint = s.cacheConstraint.next
	} else {
		constraint = &Constraint{}
	}
	constraint.scene = s
	constraint.solver = solver
	constraint.prev = nil
	constraint.next = nil
	constraint.enabled = true
	constraint.primary = primary
	constraint.secondary = secondary
	s.appendConstraint(constraint)
	return constraint
}

// Update runs a number of physics iterations
// until the specified number of seconds worth
// of simulation have passed.
func (s *Scene) Update(elapsedSeconds float32) {
	for elapsedSeconds > s.stepSeconds {
		s.runSimulation(s.stepSeconds)
		elapsedSeconds -= s.stepSeconds
	}
	s.runSimulation(elapsedSeconds)
}

// Delete releases resources allocated by this
// scene. Users should not call any further methods
// on this object.
func (s *Scene) Delete() {
	s.firstBody = nil
	s.lastBody = nil

	s.firstConstraint = nil
	s.lastConstraint = nil
}

func (s *Scene) appendBody(body *Body) {
	if s.firstBody == nil {
		s.firstBody = body
	}
	if s.lastBody != nil {
		s.lastBody.next = body
		body.prev = s.lastBody
	}
	body.next = nil
	s.lastBody = body
}

func (s *Scene) removeBody(body *Body) {
	if s.firstBody == body {
		s.firstBody = body.next
	}
	if s.lastBody == body {
		s.lastBody = body.prev
	}
	if body.next != nil {
		body.next.prev = body.prev
	}
	if body.prev != nil {
		body.prev.next = body.next
	}
	body.prev = nil
	body.next = nil
}

func (s *Scene) appendConstraint(constraint *Constraint) {
	if s.firstConstraint == nil {
		s.firstConstraint = constraint
	}
	if s.lastConstraint != nil {
		s.lastConstraint.next = constraint
		constraint.prev = s.lastConstraint
	}
	constraint.next = nil
	s.lastConstraint = constraint
}

func (s *Scene) removeConstraint(constraint *Constraint) {
	if s.firstConstraint == constraint {
		s.firstConstraint = constraint.next
	}
	if s.lastConstraint == constraint {
		s.lastConstraint = constraint.prev
	}
	if constraint.next != nil {
		constraint.next.prev = constraint.prev
	}
	if constraint.prev != nil {
		constraint.prev.next = constraint.next
	}
	constraint.prev = nil
	constraint.next = nil
}

func (s *Scene) runSimulation(elapsedSeconds float32) {
	s.resetConstraints()
	s.applyForces()
	s.integrate(elapsedSeconds)
	for i := 0; i < s.impulseIterations; i++ {
		s.applyImpulses(elapsedSeconds)
	}
	s.applyMotion(elapsedSeconds)
	for i := 0; i < s.nudgeIterations; i++ {
		s.applyNudges(elapsedSeconds)
	}
	s.detectCollisions()
}

func (s *Scene) resetConstraints() {
	for constraint := s.firstConstraint; constraint != nil; constraint = constraint.next {
		constraint.solver.Reset()
	}
}

func (s *Scene) applyForces() {
	for body := s.firstBody; body != nil; body = body.next {
		if body.static {
			continue
		}

		body.resetAcceleration()
		body.resetAngularAcceleration()

		body.addAcceleration(s.gravity)

		deltaWindVelocity := sprec.Vec3Diff(s.windVelocity, body.velocity)
		dragForce := sprec.Vec3Prod(deltaWindVelocity, deltaWindVelocity.Length()*s.windDensity*body.dragFactor)
		body.applyForce(dragForce)

		angularDragForce := sprec.Vec3Prod(body.angularVelocity, -body.angularVelocity.Length()*s.windDensity*body.angularDragFactor)
		body.applyTorque(angularDragForce)
	}

	// TODO: Apply custom force fields
}

func (s *Scene) integrate(elapsedSeconds float32) {
	for body := s.firstBody; body != nil; body = body.next {
		if body.static {
			continue
		}

		body.clampAcceleration(s.maxAcceleration)
		body.clampAngularAcceleration(s.maxAngularAcceleration)

		deltaVelocity := sprec.Vec3Prod(body.acceleration, elapsedSeconds)
		body.addVelocity(deltaVelocity)
		deltaAngularVelocity := sprec.Vec3Prod(body.angularAcceleration, elapsedSeconds)
		body.addAngularVelocity(deltaAngularVelocity)
	}
}

func (s *Scene) applyImpulses(elapsedSeconds float32) {
	for constraint := s.firstConstraint; constraint != nil; constraint = constraint.next {
		solution := constraint.solver.CalculateImpulses(constraint.primary, constraint.secondary, elapsedSeconds)
		if body := constraint.primary; body != nil {
			body.applyImpulse(solution.PrimaryImpulse)
			body.applyAngularImpulse(solution.PrimaryAngularImpulse)
		}
		if body := constraint.secondary; body != nil {
			body.applyImpulse(solution.SecondaryImpulse)
			body.applyAngularImpulse(solution.SecondaryAngularImpulse)
		}
	}
}

func (s *Scene) applyMotion(elapsedSeconds float32) {
	for body := s.firstBody; body != nil; body = body.next {
		if body.static {
			continue
		}

		body.clampVelocity(s.maxVelocity)
		body.clampAngularVelocity(s.maxAngularVelocity)

		deltaPosition := sprec.Vec3Prod(body.velocity, elapsedSeconds)
		body.translate(deltaPosition)
		deltaRotation := sprec.Vec3Prod(body.angularVelocity, elapsedSeconds)
		body.vectorRotate(deltaRotation)
	}
}

func (s *Scene) applyNudges(elapsedSeconds float32) {
	for constraint := s.firstConstraint; constraint != nil; constraint = constraint.next {
		solution := constraint.solver.CalculateNudges(constraint.primary, constraint.secondary, elapsedSeconds)
		if body := constraint.primary; body != nil {
			body.applyNudge(solution.PrimaryNudge)
			body.applyAngularNudge(solution.PrimaryAngularNudge)
		}
		if body := constraint.secondary; body != nil {
			body.applyNudge(solution.SecondaryNudge)
			body.applyAngularNudge(solution.SecondaryAngularNudge)
		}
	}
}

func (s *Scene) detectCollisions() {
	for _, constraint := range s.collisionConstraints {
		constraint.Delete()
	}
	s.collisionConstraints = s.collisionConstraints[:0]
	s.collisionSolvers = s.collisionSolvers[:0]

	for primary := s.firstBody; primary != nil; primary = primary.next {
		for secondary := primary.next; secondary != nil; secondary = secondary.next {
			s.checkCollisionTwoBodies(primary, secondary)
		}
	}
}

func (s *Scene) allocateGroundCollisionSolver() *groundCollisionSolver {
	if len(s.collisionSolvers) < cap(s.collisionSolvers) {
		s.collisionSolvers = s.collisionSolvers[:len(s.collisionSolvers)+1]
	} else {
		s.collisionSolvers = append(s.collisionSolvers, groundCollisionSolver{})
	}
	return &s.collisionSolvers[len(s.collisionSolvers)-1]
}

func (s *Scene) checkCollisionTwoBodies(primary, secondary *Body) {
	if primary.static && secondary.static {
		return
	}

	// FIXME: Temporary, to prevent non-static entities from colliding for now
	// Currently, only static to non-static is supported
	if !primary.static && !secondary.static {
		return
	}

	for _, primaryPlacement := range primary.collisionShapes {
		primaryPlacementWS := (primaryPlacement.(shape.Placement)).Transformed(primary.position, primary.orientation)

		for _, secondaryPlacement := range secondary.collisionShapes {
			secondaryPlacementWS := (secondaryPlacement.(shape.Placement)).Transformed(secondary.position, secondary.orientation)

			s.intersectionSet.Reset()
			shape.CheckIntersection(primaryPlacementWS, secondaryPlacementWS, s.intersectionSet)

			for _, intersection := range s.intersectionSet.Intersections() {
				// TODO: Once both non-static are supported, a dual-body collision constraint
				// should be used instead of individual uni-body constraints

				if !primary.static {
					solver := s.allocateGroundCollisionSolver()
					solver.Normal = intersection.FirstDisplaceNormal
					solver.ContactPoint = intersection.FirstContact
					solver.Depth = intersection.Depth
					s.collisionConstraints = append(s.collisionConstraints, s.CreateSingleBodyConstraint(solver, primary))
				}

				if !secondary.static {
					solver := s.allocateGroundCollisionSolver()
					solver.Normal = intersection.SecondDisplaceNormal
					solver.ContactPoint = intersection.SecondContact
					solver.Depth = intersection.Depth
					s.collisionConstraints = append(s.collisionConstraints, s.CreateSingleBodyConstraint(solver, secondary))
				}
			}
		}
	}
}
