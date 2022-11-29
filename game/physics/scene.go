package physics

import (
	"time"

	"github.com/mokiat/gomath/dprec"
	"github.com/mokiat/lacking/util/metrics"
	"github.com/mokiat/lacking/util/shape"
	"github.com/mokiat/lacking/util/spatial"
)

// TODO: Nudges are slower because each nudge iteration moves the body
// a bit, causing it to be relocated in the octree. Either optimize
// octree relocations or better yet accumulate positional changes
// and apply them only at the end.
// This would mean that there needs to be a temporary replacement structure
// for a body that is used only during constraint evaluation.
// This might not be such a bad idea, since it could temporarily contain the
// inverse mass and moment of intertia, avoiding inverse matrix calculations.

func newScene(engine *Engine, timestep time.Duration) *Scene {
	return &Scene{
		engine:     engine,
		bodyOctree: spatial.NewOctree[*Body](32000.0, 9, 2_000_000),

		dynamicBodies: make(map[*Body]struct{}),

		maxAcceleration:        200.0, // TODO: Measure something reasonable
		maxAngularAcceleration: 200.0, // TODO: Measure something reasonable
		maxVelocity:            2000.0,
		maxAngularVelocity:     2000.0, // TODO: Measure something reasonable

		intersectionSet: shape.NewIntersectionResultSet(128),

		timestep:     timestep,
		timeSpeed:    1.0,
		gravity:      dprec.NewVec3(0.0, -9.8, 0.0),
		windVelocity: dprec.NewVec3(0.0, 0.0, 0.0),
		windDensity:  1.2,

		revision: 1,
	}
}

// Scene represents a physics scene that contains
// a number of bodies that are independent on any
// bodies managed by other scene objects.
type Scene struct {
	engine     *Engine
	bodyOctree *spatial.Octree[*Body]

	timestep               time.Duration
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

	collisionConstraints     []*SBConstraint
	collisionSolvers         []soloCollisionSolver
	dualCollisionConstraints []*DBConstraint
	dualCollisionSolvers     []dualCollisionSolver
	intersectionSet          *shape.IntersectionResultSet

	timeSpeed    float64
	gravity      dprec.Vec3
	windVelocity dprec.Vec3
	windDensity  float64

	revision int
}

// Engine returns the physics Engine that owns this Scene.
func (s *Scene) Engine() *Engine {
	return s.engine
}

// TimeSpeed returns the speed at which time runs, where 1.0 is the default
// and 0.0 is stopped.
func (s *Scene) TimeSpeed() float64 {
	return s.timeSpeed
}

// SetTimeSpeed changes the rate at which time runs.
func (s *Scene) SetTimeSpeed(timeSpeed float64) {
	s.timeSpeed = timeSpeed
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

// CreateBody creates a new physics body and places
// it within this scene.
func (s *Scene) CreateBody(info BodyInfo) *Body {
	var body *Body
	if s.cachedBody != nil {
		body = s.cachedBody
		s.cachedBody = s.cachedBody.next
	} else {
		body = &Body{}
	}
	body.scene = s
	body.item = s.bodyOctree.CreateItem(body)
	body.prev = nil
	body.next = nil

	body.SetName(info.Name)
	body.SetPosition(info.Position)
	body.SetOrientation(info.Rotation)
	body.SetStatic(!info.IsDynamic)

	def := info.Definition
	body.SetMass(def.mass)
	body.SetMomentOfInertia(def.momentOfInertia)
	body.SetRestitutionCoefficient(def.restitutionCoefficient)
	body.SetDragFactor(def.dragFactor)
	body.SetAngularDragFactor(def.angularDragFactor)
	body.SetCollisionGroup(def.collisionGroup)
	body.SetCollisionShapes(def.collisionShapes)
	body.SetAerodynamicShapes(def.aerodynamicShapes)

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

// Update runs a number of physics iterations until the specified number of
// seconds worth of simulation have passed.
func (s *Scene) Update(elapsedSeconds float64) {
	stepSeconds := s.timestep.Seconds()

	elapsedSeconds = elapsedSeconds * s.timeSpeed
	for elapsedSeconds > stepSeconds {
		s.runSimulation(stepSeconds)
		elapsedSeconds -= stepSeconds
	}
	s.runSimulation(elapsedSeconds)
}

func (s *Scene) Each(cb func(b *Body)) {
	for body := s.firstBody; body != nil; body = body.next {
		cb(body)
	}
}

func (s *Scene) Nearby(body *Body, distance float64, cb func(b *Body)) {
	region := spatial.CuboidRegion(
		body.position,
		dprec.NewVec3(distance, distance, distance),
	)
	s.bodyOctree.VisitHexahedronRegion(&region, spatial.VisitorFunc[*Body](func(candidate *Body) {
		if candidate != body {
			cb(candidate)
		}
	}))
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
	constraint.prev = s.lastSBConstraint
	constraint.next = nil
	if s.firstSBConstraint == nil {
		s.firstSBConstraint = constraint
	}
	if s.lastSBConstraint != nil {
		s.lastSBConstraint.next = constraint
	}
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
	if elapsedSeconds < epsilon {
		return
	}
	s.resetConstraints(elapsedSeconds)
	s.applyForces()
	s.integrate(elapsedSeconds)
	s.applyImpulses(elapsedSeconds)
	s.applyMotion(elapsedSeconds)
	s.applyNudges(elapsedSeconds)
	s.detectCollisions()
}

func (s *Scene) resetConstraints(elapsedSeconds float64) {
	defer metrics.BeginSpan("reset").End()

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
	defer metrics.BeginSpan("forces").End()
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
	defer metrics.BeginSpan("integrate").End()
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
	defer metrics.BeginSpan("impulses").End()

	for i := 0; i < ImpulseIterationCount; i++ {
		for constraint := s.firstDBConstraint; constraint != nil; constraint = constraint.next {
			constraint.solver.ApplyImpulses(DBSolverContext{
				Primary:        constraint.primary,
				Secondary:      constraint.secondary,
				ElapsedSeconds: elapsedSeconds,
			})
		}
		for constraint := s.firstSBConstraint; constraint != nil; constraint = constraint.next {
			constraint.solver.ApplyImpulses(SBSolverContext{
				Body:           constraint.body,
				ElapsedSeconds: elapsedSeconds,
			})
		}
	}
}

func (s *Scene) applyMotion(elapsedSeconds float64) {
	defer metrics.BeginSpan("motion").End()
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
	defer metrics.BeginSpan("nudges").End()

	for i := 0; i < NudgeIterationCount; i++ {
		for constraint := s.firstDBConstraint; constraint != nil; constraint = constraint.next {
			ctx := DBSolverContext{
				Primary:        constraint.primary,
				Secondary:      constraint.secondary,
				ElapsedSeconds: elapsedSeconds,
			}
			constraint.solver.Reset(ctx)
			constraint.solver.ApplyNudges(ctx)
		}
		for constraint := s.firstSBConstraint; constraint != nil; constraint = constraint.next {
			ctx := SBSolverContext{
				Body:           constraint.body,
				ElapsedSeconds: elapsedSeconds,
			}
			constraint.solver.Reset(ctx)
			constraint.solver.ApplyNudges(ctx)
		}
	}
}

func (s *Scene) detectCollisions() {
	defer metrics.BeginSpan("collision").End()
	s.revision++

	for _, constraint := range s.collisionConstraints {
		constraint.Delete()
	}
	s.collisionConstraints = s.collisionConstraints[:0]
	s.collisionSolvers = s.collisionSolvers[:0]

	for _, constraint := range s.dualCollisionConstraints {
		constraint.Delete()
	}
	s.dualCollisionConstraints = s.dualCollisionConstraints[:0]
	s.dualCollisionSolvers = s.dualCollisionSolvers[:0]

	for primary := range s.dynamicBodies {
		primary.revision = s.revision

		if len(primary.collisionShapes) == 0 {
			continue
		}

		region := spatial.CuboidRegion(
			primary.position,
			dprec.NewVec3(primary.bsRadius*2.0, primary.bsRadius*2.0, primary.bsRadius*2.0),
		)
		s.bodyOctree.VisitHexahedronRegion(&region, spatial.VisitorFunc[*Body](func(secondary *Body) {
			if secondary == primary {
				return
			}
			if secondary.revision == s.revision {
				return // secondary already processed
			}
			if len(secondary.collisionShapes) == 0 {
				return
			}
			s.checkCollisionTwoBodies(secondary, primary) // FIXME: Reverse order does not work!
		}))
	}
}

func (s *Scene) allocateGroundCollisionSolver() *soloCollisionSolver {
	if len(s.collisionSolvers) < cap(s.collisionSolvers) {
		s.collisionSolvers = s.collisionSolvers[:len(s.collisionSolvers)+1]
	} else {
		s.collisionSolvers = append(s.collisionSolvers, soloCollisionSolver{})
	}
	return &s.collisionSolvers[len(s.collisionSolvers)-1]
}

func (s *Scene) allocateDualCollisionSolver() *dualCollisionSolver {
	if len(s.dualCollisionSolvers) < cap(s.dualCollisionSolvers) {
		s.dualCollisionSolvers = s.dualCollisionSolvers[:len(s.dualCollisionSolvers)+1]
	} else {
		s.dualCollisionSolvers = append(s.dualCollisionSolvers, dualCollisionSolver{})
	}
	return &s.dualCollisionSolvers[len(s.dualCollisionSolvers)-1]
}

func (s *Scene) checkCollisionTwoBodies(primary, secondary *Body) {
	if primary.static && secondary.static {
		return
	}
	if (primary.collisionGroup == secondary.collisionGroup) && (primary.collisionGroup != 0) {
		return
	}

	primaryTransform := shape.NewTransform(primary.position, primary.orientation)
	secondaryTransform := shape.NewTransform(secondary.position, secondary.orientation)

	for _, primaryShape := range primary.collisionShapes {
		primaryShape = primaryShape.Transformed(primaryTransform)

		for _, secondaryShape := range secondary.collisionShapes {
			secondaryShape = secondaryShape.Transformed(secondaryTransform)

			s.intersectionSet.Reset()
			shape.CheckIntersection(primaryShape, secondaryShape, s.intersectionSet)

			for _, intersection := range s.intersectionSet.Intersections() {
				if !primary.static && !secondary.static {
					solver := s.allocateDualCollisionSolver()
					solver.primaryCollisionNormal = intersection.FirstDisplaceNormal
					solver.primaryCollisionPoint = intersection.FirstContact
					solver.secondaryCollisionNormal = intersection.SecondDisplaceNormal
					solver.secondaryCollisionPoint = intersection.SecondContact
					solver.collisionDepth = intersection.Depth
					s.dualCollisionConstraints = append(s.dualCollisionConstraints, s.CreateDoubleBodyConstraint(primary, secondary, solver))
					continue
				}

				if !primary.static {
					solver := s.allocateGroundCollisionSolver()
					solver.collisionNormal = intersection.FirstDisplaceNormal
					solver.collisionPoint = intersection.FirstContact
					solver.collisionDepth = intersection.Depth
					s.collisionConstraints = append(s.collisionConstraints, s.CreateSingleBodyConstraint(primary, solver))
				}

				if !secondary.static {
					solver := s.allocateGroundCollisionSolver()
					solver.collisionNormal = intersection.SecondDisplaceNormal
					solver.collisionPoint = intersection.SecondContact
					solver.collisionDepth = intersection.Depth
					s.collisionConstraints = append(s.collisionConstraints, s.CreateSingleBodyConstraint(secondary, solver))
				}
			}
		}
	}
}
