package physics

import (
	"github.com/mokiat/gomath/dprec"
	"github.com/mokiat/lacking/util/metrics"
	"github.com/mokiat/lacking/util/shape"
	"github.com/mokiat/lacking/util/spatial"
)

const (
	defaultImpulseIterations = 100
	defaultNudgeIterations   = 100
)

func newScene(stepSeconds float64) *Scene {
	return &Scene{
		bodyOctree: spatial.NewOctree[*Body](32000.0, 9, 2_000_000),

		dynamicBodies: make(map[*Body]struct{}),

		stepSeconds:       stepSeconds,
		impulseIterations: defaultImpulseIterations,
		nudgeIterations:   defaultNudgeIterations,

		maxAcceleration:        200.0, // TODO: Measure something reasonable
		maxAngularAcceleration: 200.0, // TODO: Measure something reasonable
		maxVelocity:            2000.0,
		maxAngularVelocity:     2000.0, // TODO: Measure something reasonable

		intersectionSet: shape.NewIntersectionResultSet(128),

		gravity:      dprec.NewVec3(0.0, -9.8, 0.0),
		windVelocity: dprec.NewVec3(0.0, 0.0, 0.0),
		windDensity:  1.2,
	}
}

// Scene represents a physics scene that contains
// a number of bodies that are independent on any
// bodies managed by other scene objects.
type Scene struct {
	bodyOctree *spatial.Octree[*Body]

	stepSeconds            float64
	impulseIterations      int
	nudgeIterations        int
	maxAcceleration        float64
	maxAngularAcceleration float64
	maxVelocity            float64
	maxAngularVelocity     float64

	dynamicBodies map[*Body]struct{}
	firstBody     *Body
	lastBody      *Body
	cachedBody    *Body

	firstSBConstraint  *SBConstraint
	lastSBConstraint   *SBConstraint
	cachedSBConstraint *SBConstraint

	firstDBConstraint  *DBConstraint
	lastDBConstraint   *DBConstraint
	cachedDBConstraint *DBConstraint

	collisionConstraints []*SBConstraint
	collisionSolvers     []groundCollisionSolver
	intersectionSet      *shape.IntersectionResultSet

	gravity      dprec.Vec3
	windVelocity dprec.Vec3
	windDensity  float64
}

// Gravity returns the gravity acceleration.
func (s *Scene) Gravity() dprec.Vec3 {
	return s.gravity
}

// SetGravity changes the gravity acceleration.
func (s *Scene) SetGravity(gravity dprec.Vec3) {
	s.gravity = gravity
}

// WindVelocity returns the wind speed.
func (s *Scene) WindVelocity() dprec.Vec3 {
	return s.windVelocity
}

// SetWindVelocity sets the wind speed.
func (s *Scene) SetWindVelocity(velocity dprec.Vec3) {
	s.windVelocity = velocity
}

// WindDensity returns the wind density.
func (s *Scene) WindDensity() float64 {
	return s.windDensity
}

// SetWindDensity changes the wind density.
func (s *Scene) SetWindDensity(density float64) {
	s.windDensity = density
}

func (s *Scene) CreateBody2(info BodyInfo) *Body {
	result := s.CreateBody()

	result.SetName(info.Name)
	result.SetPosition(info.Position)
	result.SetOrientation(info.Rotation)
	result.SetStatic(!info.IsDynamic)

	def := info.Definition
	result.SetMass(def.mass)
	result.SetMomentOfInertia(def.momentOfInertia)
	result.SetRestitutionCoefficient(def.restitutionCoefficient)
	result.SetDragFactor(def.dragFactor)
	result.SetAngularDragFactor(def.angularDragFactor)
	result.SetCollisionShapes(def.collisionShapes)
	result.SetAerodynamicShapes(def.aerodynamicShapes)

	return result
}

// CreateBody creates a new physics body and places
// it within this scene.
func (s *Scene) CreateBody() *Body {
	var body *Body
	if s.cachedBody != nil {
		body = s.cachedBody
		s.cachedBody = s.cachedBody.next
	} else {
		body = &Body{}
	}
	// TODO: Init Body with sane defaults
	body.scene = s
	body.item = s.bodyOctree.CreateItem(body)
	body.prev = nil
	body.next = nil
	body.SetStatic(false)
	s.appendBody(body)
	return body
}

// CreateConstraintSet creates a new ConstraintSet.
func (s *Scene) CreateConstraintSet() *ConstraintSet {
	return &ConstraintSet{
		scene: s,
	}
}

// CreateSingleBodyConstraint creates a new physics constraint that acts on
// a single body and enables it for this scene.
func (s *Scene) CreateSingleBodyConstraint(body *Body, solver SBConstraintSolver) *SBConstraint {
	var constraint *SBConstraint
	if s.cachedSBConstraint != nil {
		constraint = s.cachedSBConstraint
		s.cachedSBConstraint = s.cachedSBConstraint.next
	} else {
		constraint = &SBConstraint{}
	}
	constraint.scene = s
	constraint.solver = solver
	constraint.prev = nil
	constraint.next = nil
	constraint.body = body
	s.appendSBConstraint(constraint)
	return constraint
}

// CreateDoubleBodyConstraint creates a new physics constraint that acts on
// two bodies and enables it for this scene.
func (s *Scene) CreateDoubleBodyConstraint(primary, secondary *Body, solver DBConstraintSolver) *DBConstraint {
	var constraint *DBConstraint
	if s.cachedDBConstraint != nil {
		constraint = s.cachedDBConstraint
		s.cachedDBConstraint = s.cachedDBConstraint.next
	} else {
		constraint = &DBConstraint{}
	}
	constraint.scene = s
	constraint.solver = solver
	constraint.prev = nil
	constraint.next = nil
	constraint.primary = primary
	constraint.secondary = secondary
	s.appendDBConstraint(constraint)
	return constraint
}

// Update runs a number of physics iterations
// until the specified number of seconds worth
// of simulation have passed.
func (s *Scene) Update(elapsedSeconds float64) {
	defer metrics.BeginSpan("physics").End()

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

	s.firstSBConstraint = nil
	s.lastSBConstraint = nil

	s.firstDBConstraint = nil
	s.lastDBConstraint = nil
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

func (s *Scene) cacheBody(body *Body) {
	body.next = s.cachedBody
	s.cachedBody = body
}

func (s *Scene) appendSBConstraint(constraint *SBConstraint) {
	if s.firstSBConstraint == nil {
		s.firstSBConstraint = constraint
	}
	if s.lastSBConstraint != nil {
		s.lastSBConstraint.next = constraint
		constraint.prev = s.lastSBConstraint
	}
	constraint.next = nil
	s.lastSBConstraint = constraint
}

func (s *Scene) removeSBConstraint(constraint *SBConstraint) {
	if s.firstSBConstraint == constraint {
		s.firstSBConstraint = constraint.next
	}
	if s.lastSBConstraint == constraint {
		s.lastSBConstraint = constraint.prev
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

func (s *Scene) cacheSBConstraint(constraint *SBConstraint) {
	constraint.next = s.cachedSBConstraint
	s.cachedSBConstraint = constraint
}

func (s *Scene) appendDBConstraint(constraint *DBConstraint) {
	if s.firstDBConstraint == nil {
		s.firstDBConstraint = constraint
	}
	if s.lastDBConstraint != nil {
		s.lastDBConstraint.next = constraint
		constraint.prev = s.lastDBConstraint
	}
	constraint.next = nil
	s.lastDBConstraint = constraint
}

func (s *Scene) removeDBConstraint(constraint *DBConstraint) {
	if s.firstDBConstraint == constraint {
		s.firstDBConstraint = constraint.next
	}
	if s.lastDBConstraint == constraint {
		s.lastDBConstraint = constraint.prev
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

func (s *Scene) cacheDBConstraint(constraint *DBConstraint) {
	constraint.next = s.cachedDBConstraint
	s.cachedDBConstraint = constraint
}

func (s *Scene) runSimulation(elapsedSeconds float64) {
	s.resetConstraints(elapsedSeconds)
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

func (s *Scene) resetConstraints(elapsedSeconds float64) {
	for constraint := s.firstSBConstraint; constraint != nil; constraint = constraint.next {
		constraint.solver.Reset(SBSolverContext{
			Body:           constraint.body,
			ElapsedSeconds: elapsedSeconds,
		})
	}
	for constraint := s.firstDBConstraint; constraint != nil; constraint = constraint.next {
		constraint.solver.Reset(DBSolverContext{
			Primary:        constraint.primary,
			Secondary:      constraint.secondary,
			ElapsedSeconds: elapsedSeconds,
		})
	}
}

func (s *Scene) applyForces() {
	for body := range s.dynamicBodies {
		body.resetAcceleration()
		body.resetAngularAcceleration()

		body.addAcceleration(s.gravity)

		deltaWindVelocity := dprec.Vec3Diff(s.windVelocity, body.velocity)
		dragForce := dprec.Vec3Prod(deltaWindVelocity, deltaWindVelocity.Length()*s.windDensity*body.dragFactor)
		body.applyForce(dragForce)

		angularDragForce := dprec.Vec3Prod(body.angularVelocity, -body.angularVelocity.Length()*s.windDensity*body.angularDragFactor)
		body.applyTorque(angularDragForce)
	}

	// TODO: Apply custom force fields
}

func (s *Scene) integrate(elapsedSeconds float64) {
	for body := range s.dynamicBodies {
		body.clampAcceleration(s.maxAcceleration)
		body.clampAngularAcceleration(s.maxAngularAcceleration)

		deltaVelocity := dprec.Vec3Prod(body.acceleration, elapsedSeconds)
		body.addVelocity(deltaVelocity)
		deltaAngularVelocity := dprec.Vec3Prod(body.angularAcceleration, elapsedSeconds)
		body.addAngularVelocity(deltaAngularVelocity)
	}
}

func (s *Scene) applyImpulses(elapsedSeconds float64) {
	for constraint := s.firstDBConstraint; constraint != nil; constraint = constraint.next {
		solution := constraint.solver.CalculateImpulses(DBSolverContext{
			Primary:        constraint.primary,
			Secondary:      constraint.secondary,
			ElapsedSeconds: elapsedSeconds,
		})
		constraint.primary.applyImpulse(solution.Primary.Impulse)
		constraint.primary.applyAngularImpulse(solution.Primary.AngularImpulse)
		constraint.secondary.applyImpulse(solution.Secondary.Impulse)
		constraint.secondary.applyAngularImpulse(solution.Secondary.AngularImpulse)
	}
	for constraint := s.firstSBConstraint; constraint != nil; constraint = constraint.next {
		solution := constraint.solver.CalculateImpulses(SBSolverContext{
			Body:           constraint.body,
			ElapsedSeconds: elapsedSeconds,
		})
		constraint.body.applyImpulse(solution.Impulse)
		constraint.body.applyAngularImpulse(solution.AngularImpulse)
	}
}

func (s *Scene) applyMotion(elapsedSeconds float64) {
	for body := range s.dynamicBodies {
		body.clampVelocity(s.maxVelocity)
		body.clampAngularVelocity(s.maxAngularVelocity)

		deltaPosition := dprec.Vec3Prod(body.velocity, elapsedSeconds)
		body.translate(deltaPosition)
		deltaRotation := dprec.Vec3Prod(body.angularVelocity, elapsedSeconds)
		body.vectorRotate(deltaRotation)
	}
}

func (s *Scene) applyNudges(elapsedSeconds float64) {
	for constraint := s.firstDBConstraint; constraint != nil; constraint = constraint.next {
		solution := constraint.solver.CalculateNudges(DBSolverContext{
			Primary:        constraint.primary,
			Secondary:      constraint.secondary,
			ElapsedSeconds: elapsedSeconds,
		})
		constraint.primary.applyNudge(solution.Primary.Nudge)
		constraint.primary.applyAngularNudge(solution.Primary.AngularNudge)
		constraint.secondary.applyNudge(solution.Secondary.Nudge)
		constraint.secondary.applyAngularNudge(solution.Secondary.AngularNudge)
	}
	for constraint := s.firstSBConstraint; constraint != nil; constraint = constraint.next {
		solution := constraint.solver.CalculateNudges(SBSolverContext{
			Body:           constraint.body,
			ElapsedSeconds: elapsedSeconds,
		})
		constraint.body.applyNudge(solution.Nudge)
		constraint.body.applyAngularNudge(solution.AngularNudge)
	}
}

func (s *Scene) detectCollisions() {
	for _, constraint := range s.collisionConstraints {
		constraint.Delete()
	}
	s.collisionConstraints = s.collisionConstraints[:0]
	s.collisionSolvers = s.collisionSolvers[:0]

	for primary := range s.dynamicBodies {
		// FIXME: 50 is hardcoded range. Also, consider using a SphericalRegion
		// instead, once BoundingSphereRadius is exposed from CollisionShape.
		region := spatial.CuboidRegion(
			primary.position,
			dprec.NewVec3(50, 50, 50),
		)

		s.bodyOctree.VisitHexahedronRegion(&region, spatial.VisitorFunc[*Body](func(secondary *Body) {
			if secondary == primary {
				return
			}
			s.checkCollisionTwoBodies(secondary, primary) // FIXME: Reverse order does not work!
		}))
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
					s.collisionConstraints = append(s.collisionConstraints, s.CreateSingleBodyConstraint(primary, solver))
				}

				if !secondary.static {
					solver := s.allocateGroundCollisionSolver()
					solver.Normal = intersection.SecondDisplaceNormal
					solver.ContactPoint = intersection.SecondContact
					solver.Depth = intersection.Depth
					s.collisionConstraints = append(s.collisionConstraints, s.CreateSingleBodyConstraint(secondary, solver))
				}
			}
		}
	}
}
