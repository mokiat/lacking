package physics

import (
	"maps"
	"time"

	"github.com/mokiat/gog/ds"
	"github.com/mokiat/gog/opt"
	"github.com/mokiat/gomath/dprec"
	"github.com/mokiat/lacking/debug/metric"
	"github.com/mokiat/lacking/game/physics/collision"
	"github.com/mokiat/lacking/game/physics/constraint"
	"github.com/mokiat/lacking/game/physics/medium"
	"github.com/mokiat/lacking/game/physics/solver"
	"github.com/mokiat/lacking/game/timestep"
	"github.com/mokiat/lacking/util/spatial"
)

func newScene(engine *Engine, interval time.Duration) *Scene {
	return &Scene{
		engine: engine,

		preUpdateSubscriptions:   NewUpdateSubscriptionSet(),
		postUpdateSubscriptions:  NewUpdateSubscriptionSet(),
		sbCollisionSubscriptions: NewSingleBodyCollisionSubscriptionSet(),
		dbCollisionSubscriptions: NewDoubleBodyCollisionSubscriptionSet(),

		timeSegmenter: timestep.NewSegmenter(interval),
		interval:      interval,
		timeSpeed:     1.0,

		maxLinearAcceleration:  200.0,
		maxAngularAcceleration: 200.0,
		maxLinearVelocity:      2000.0,
		maxAngularVelocity:     2000.0,

		mediumSolver: medium.NewStaticAirMedium(),

		props: make([]propState, 0, 1024),
		propOctree: spatial.NewStaticOctree[uint32](spatial.StaticOctreeSettings{
			Size:                opt.V(32000.0),
			MaxDepth:            opt.V(int32(15)),
			BiasRatio:           opt.V(2.0),
			InitialNodeCapacity: opt.V(int32(4 * 1024)),
			InitialItemCapacity: opt.V(int32(8 * 1024)),
		}),

		freeBodyIndices:            ds.NewStack[uint32](16),
		bodies:                     make([]bodyState, 0, 64),
		bodyAccelerationTargets:    make([]solver.AccelerationTarget, 0, 64),
		bodyConstraintPlaceholders: make([]solver.Placeholder, 0, 64),
		bodyOctree: spatial.NewDynamicOctree[uint32](spatial.DynamicOctreeSettings{
			Size:                opt.V(32000.0),
			MaxDepth:            opt.V(int32(15)),
			BiasRatio:           opt.V(2.0),
			InitialNodeCapacity: opt.V(int32(4 * 1024)),
			InitialItemCapacity: opt.V(int32(8 * 1024)),
		}),

		// bodyAccelerators   []any // TOOD
		// areaAccelerators   []any // TODO
		globalAccelerators: make([]globalAcceleratorState, 0, 64),

		freeBodyAcceleratorIndices:   ds.NewStack[uint32](16),
		freeAreaAcceleratorIndices:   ds.NewStack[uint32](16),
		freeGlobalAcceleratorIndices: ds.NewStack[uint32](16),

		sbConstraints: make([]sbConstraintState, 0, 64),
		dbConstraints: make([]dbConstraintState, 0, 64),

		freeSBConstraintIndices: ds.NewStack[uint32](16),
		freeDBConstraintIndices: ds.NewStack[uint32](16),

		collisionSet: collision.NewIntersectionBucket(128),

		oldSBCollisions: make(map[sbCollisionPair]struct{}, 32),
		newSBCollisions: make(map[sbCollisionPair]struct{}, 32),

		oldDBCollisions: make(map[dbCollisionPair]struct{}, 32),
		newDBCollisions: make(map[dbCollisionPair]struct{}, 32),
	}
}

// Scene represents a physics scene that contains
// a number of bodies that are independent on any
// bodies managed by other scene objects.
type Scene struct {
	engine *Engine

	preUpdateSubscriptions   *UpdateSubscriptionSet
	postUpdateSubscriptions  *UpdateSubscriptionSet
	sbCollisionSubscriptions *SingleBodyCollisionSubscriptionSet
	dbCollisionSubscriptions *DoubleBodyCollisionSubscriptionSet

	timeSegmenter *timestep.Segmenter
	interval      time.Duration
	timeSpeed     float64

	maxLinearAcceleration  float64
	maxAngularAcceleration float64
	maxLinearVelocity      float64
	maxAngularVelocity     float64

	mediumSolver solver.Medium

	props      []propState
	propOctree *spatial.StaticOctree[uint32]

	bodies                     []bodyState
	bodyAccelerationTargets    []solver.AccelerationTarget
	bodyConstraintPlaceholders []solver.Placeholder
	freeBodyIndices            *ds.Stack[uint32]
	bodyOctree                 *spatial.DynamicOctree[uint32]

	// bodyAccelerators   []any // TOOD
	freeBodyAcceleratorIndices *ds.Stack[uint32]

	// areaAccelerators   []any // TODO
	freeAreaAcceleratorIndices *ds.Stack[uint32]

	globalAccelerators           []globalAcceleratorState
	freeGlobalAcceleratorIndices *ds.Stack[uint32]

	sbConstraints           []sbConstraintState
	freeSBConstraintIndices *ds.Stack[uint32]

	dbConstraints           []dbConstraintState
	freeDBConstraintIndices *ds.Stack[uint32]

	sbCollisionConstraints []SBConstraint
	sbCollisionSolvers     []constraint.Collision

	dbCollisionConstraints []DBConstraint
	dbCollisionSolvers     []constraint.PairCollision

	collisionSet *collision.IntersectionBucket

	oldSBCollisions map[sbCollisionPair]struct{}
	newSBCollisions map[sbCollisionPair]struct{}

	oldDBCollisions map[dbCollisionPair]struct{}
	newDBCollisions map[dbCollisionPair]struct{}

	freeRevision uint32
}

// Delete releases resources allocated by this scene. Users should not call
// any further methods on this object.
func (s *Scene) Delete() {
	s.engine = nil

	s.props = nil
	s.propOctree = nil

	s.freeBodyIndices = nil
	s.bodies = nil
	s.bodyAccelerationTargets = nil
	s.bodyConstraintPlaceholders = nil
	s.bodyOctree = nil

	s.globalAccelerators = nil
	s.freeBodyAcceleratorIndices = nil
	s.freeAreaAcceleratorIndices = nil
	s.freeGlobalAcceleratorIndices = nil

	s.sbConstraints = nil
	s.freeSBConstraintIndices = nil

	s.dbConstraints = nil
	s.freeDBConstraintIndices = nil

	s.sbCollisionConstraints = nil
	s.sbCollisionSolvers = nil

	s.dbCollisionConstraints = nil
	s.dbCollisionSolvers = nil

	s.collisionSet = nil

	s.oldSBCollisions = nil
	s.newSBCollisions = nil

	s.oldDBCollisions = nil
	s.newDBCollisions = nil
}

// Engine returns the physics Engine that owns this Scene.
func (s *Scene) Engine() *Engine {
	return s.engine
}

// SubscribePreUpdate registers a callback that is invoked before each physics
// iteration.
func (s *Scene) SubscribePreUpdate(callback UpdateCallback) *UpdateSubscription {
	return s.preUpdateSubscriptions.Subscribe(callback)
}

// SubscribePostUpdate registers a callback that is invoked after each physics
// iteration.
func (s *Scene) SubscribePostUpdate(callback UpdateCallback) *UpdateSubscription {
	return s.postUpdateSubscriptions.Subscribe(callback)
}

// SubscribeSingleBodyCollision registers a callback that is invoked when a body
// collides with a static object.
func (s *Scene) SubscribeSingleBodyCollision(callback SingleBodyCollisionCallback) *SingleBodyCollisionSubscription {
	return s.sbCollisionSubscriptions.Subscribe(callback)
}

// SubscribeDoubleBodyCollision registers a callback that is invoked when two
// bodies collide.
func (s *Scene) SubscribeDoubleBodyCollision(callback DoubleBodyCollisionCallback) *DoubleBodyCollisionSubscription {
	return s.dbCollisionSubscriptions.Subscribe(callback)
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

// MaxLinearAcceleration returns the maximum linear acceleration that a body
// can have.
func (s *Scene) MaxLinearAcceleration() float64 {
	return s.maxLinearAcceleration
}

// SetMaxLinearAcceleration changes the maximum linear acceleration that a body
// can have.
func (s *Scene) SetMaxLinearAcceleration(acceleration float64) {
	s.maxLinearAcceleration = acceleration
}

// MaxAngularAcceleration returns the maximum angular acceleration that a body
// can have.
func (s *Scene) MaxAngularAcceleration() float64 {
	return s.maxAngularAcceleration
}

// SetMaxAngularAcceleration changes the maximum angular acceleration that a
// body can have.
func (s *Scene) SetMaxAngularAcceleration(acceleration float64) {
	s.maxAngularAcceleration = acceleration
}

// MediumSolver returns the solver that is used to calculate the medium
// properties of the scene.
func (s *Scene) MediumSolver() solver.Medium {
	return s.mediumSolver
}

// SetMediumSolver changes the solver that is used to calculate the medium
// properties of the scene.
func (s *Scene) SetMediumSolver(solver solver.Medium) {
	s.mediumSolver = solver
}

// CreateGlobalAccelerator creates a new accelerator that affects the whole
// scene.
func (s *Scene) CreateGlobalAccelerator(logic solver.Acceleration) GlobalAccelerator {
	return createGlobalAccelerator(s, logic)
}

// CreateProp creates a new static Prop. A prop is an object
// that is static and rarely removed.
func (s *Scene) CreateProp(info PropInfo) {
	// TODO: createProp(s, info)

	bs := info.CollisionSet.BoundingSphere()
	position := bs.Position()
	radius := bs.Radius()

	propIndex := uint32(len(s.props))
	s.props = append(s.props, propState{
		reference:    newIndexReference(propIndex, s.nextRevision()),
		name:         info.Name,
		collisionSet: info.CollisionSet,
	})
	s.propOctree.Insert(position, radius, propIndex)
}

// CreateBody creates a new physics body and places
// it within this scene.
func (s *Scene) CreateBody(info BodyInfo) Body {
	return createBody(s, info)
}

// CreateConstraintSet creates a new ConstraintSet.
func (s *Scene) CreateConstraintSet() *ConstraintSet {
	return &ConstraintSet{
		scene: s,
	}
}

// CreateSingleBodyConstraint creates a new physics constraint that acts on
// a single body and enables it for this scene.
func (s *Scene) CreateSingleBodyConstraint(body Body, logic solver.Constraint) SBConstraint {
	return createSBConstraint(s, logic, body)
}

// CreateDoubleBodyConstraint creates a new physics constraint that acts on
// two bodies and enables it for this scene.
func (s *Scene) CreateDoubleBodyConstraint(primary, secondary Body, logic solver.PairConstraint) DBConstraint {
	return createDBConstraint(s, logic, primary, secondary)
}

// Update runs a number of physics iterations until the specified duration has
// been reached or surpassed (yes physics can get ahead).
func (s *Scene) Update(elapsedTime time.Duration) {
	s.timeSegmenter.Update(elapsedTime, s.onTickInterval, s.onTickLerp)
}

func (s *Scene) Each(cb func(b Body)) {
	s.eachBodyState(func(_ int, b *bodyState) {
		cb(Body{
			scene:     s,
			reference: b.reference,
		})
	})
}

func (s *Scene) Nearby(body Body, distance float64, cb func(b Body)) {
	state := s.resolveBodyState(body.reference)
	if state == nil {
		return
	}
	region := spatial.CuboidRegion(
		state.position,
		dprec.NewVec3(distance, distance, distance),
	)
	s.bodyOctree.VisitHexahedronRegion(&region, spatial.VisitorFunc[uint32](func(candidate uint32) {
		candidateState := &s.bodies[candidate]
		if candidateState != state {
			cb(Body{
				scene:     s,
				reference: candidateState.reference,
			})
		}
	}))
}

func (s *Scene) runSimulation(elapsedSeconds float64) {
	s.resetOldPositions()
	if elapsedSeconds > 0.0001 {
		s.applyAcceleration(elapsedSeconds)
		s.applyImpulses(elapsedSeconds)
		s.applyMotion(elapsedSeconds)
		s.applyNudges(elapsedSeconds)
		s.detectCollisions()
	}
}

func (s *Scene) resetOldPositions() {
	s.eachBodyState(func(_ int, body *bodyState) {
		body.oldPosition = body.position
		body.oldRotation = body.rotation
	})
}

func (s *Scene) applyAcceleration(elapsedSeconds float64) {
	defer metric.BeginRegion("acceleration").End()
	s.prepareAccelerationTargets()
	s.applyBodyAccelerators()
	s.applyAreaAccelerators()
	s.applyGlobalAccelerators()
	s.applyAerodynamicAccelerations()
	s.applyAccelerationTargets(elapsedSeconds)
}

func (s *Scene) prepareAccelerationTargets() {
	s.eachBodyState(func(index int, body *bodyState) {
		s.bodyAccelerationTargets[index] = solver.NewAccelerationTarget(
			body.mass,
			body.momentOfInertia,
			body.position,
			body.rotation,
			body.velocity,
			body.angularVelocity,
		)
	})
}

func (s *Scene) applyBodyAccelerators() {
	// TODO
}

func (s *Scene) applyAreaAccelerators() {
	// TODO
}

func (s *Scene) applyGlobalAccelerators() {
	s.eachBodyState(func(index int, body *bodyState) {
		target := &s.bodyAccelerationTargets[index]
		for _, accelerator := range s.globalAccelerators {
			if !accelerator.reference.IsValid() || !accelerator.enabled {
				continue
			}
			accelerator.logic.ApplyAcceleration(solver.AccelerationContext{
				Target: target,
			})
		}
	})
}

func (s *Scene) applyAerodynamicAccelerations() {
	if s.mediumSolver == nil {
		return
	}
	s.eachBodyState(func(index int, body *bodyState) {
		if len(body.aerodynamicShapes) == 0 {
			return
		}
		target := &s.bodyAccelerationTargets[index]
		mediumDensity := s.mediumSolver.Density(body.position)
		mediumVelocity := s.mediumSolver.Velocity(body.position)

		deltaVelocity := dprec.Vec3Diff(mediumVelocity, body.velocity)
		dragForce := dprec.Vec3Prod(deltaVelocity, deltaVelocity.Length()*mediumDensity*body.dragFactor)
		target.ApplyForce(dragForce)

		angularDragForce := dprec.Vec3Prod(body.angularVelocity, -body.angularVelocity.Length()*mediumDensity*body.angularDragFactor)
		target.ApplyTorque(angularDragForce)

		bodyTransform := NewTransform(body.position, body.rotation)
		for _, aerodynamicShape := range body.aerodynamicShapes {
			aerodynamicShape = aerodynamicShape.Transformed(bodyTransform)
			relativeSpeed := dprec.QuatVec3Rotation(dprec.InverseQuat(aerodynamicShape.Rotation()), deltaVelocity)

			force := aerodynamicShape.solver.Force(relativeSpeed, mediumDensity)
			absoluteForce := dprec.QuatVec3Rotation(aerodynamicShape.Rotation(), force)
			target.ApplyForce(absoluteForce) // TODO: Apply at offset
			// target.ApplyOffsetForce(absoluteForce, aerodynamicShape.Position())
		}
	})
}

func (s *Scene) applyAccelerationTargets(elapsedSeconds float64) {
	s.eachBodyState(func(index int, body *bodyState) {
		target := s.bodyAccelerationTargets[index]

		linearAcceleration := target.AccumulatedLinearAcceleration()
		if linearAcceleration.Length() > s.maxLinearAcceleration {
			linearAcceleration = dprec.ResizedVec3(linearAcceleration, s.maxLinearAcceleration)
		}
		body.AddVelocity(dprec.Vec3Prod(linearAcceleration, elapsedSeconds))

		angularAcceleration := target.AccumulatedAngularAcceleration()
		if angularAcceleration.Length() > s.maxAngularAcceleration {
			angularAcceleration = dprec.ResizedVec3(angularAcceleration, s.maxAngularAcceleration)
		}
		body.AddAngularVelocity(dprec.Vec3Prod(angularAcceleration, elapsedSeconds))
	})
}

func (s *Scene) applyImpulses(elapsedSeconds float64) {
	defer metric.BeginRegion("impulses").End()

	s.eachBodyState(func(i int, body *bodyState) {
		placeholder := &s.bodyConstraintPlaceholders[i]
		s.initPlaceholder(placeholder, body)
	})

	s.eachSBConstraintState(func(_ int, constraint *sbConstraintState) {
		if s.resolveBodyState(constraint.body.reference) == nil {
			deleteSBConstraint(s, constraint.reference)
			return
		}
	})
	s.eachDBConstraintState(func(_ int, constraint *dbConstraintState) {
		if s.resolveBodyState(constraint.primary.reference) == nil {
			deleteDBConstraint(s, constraint.reference)
			return
		}
		if s.resolveBodyState(constraint.secondary.reference) == nil {
			deleteDBConstraint(s, constraint.reference)
			return
		}
	})

	s.eachDBConstraintState(func(_ int, constraint *dbConstraintState) {
		target := &s.bodyConstraintPlaceholders[constraint.primary.reference.Index]
		source := &s.bodyConstraintPlaceholders[constraint.secondary.reference.Index]
		constraint.logic.Reset(solver.PairContext{
			Target:      target,
			Source:      source,
			DeltaTime:   elapsedSeconds,
			ImpulseBeta: ImpulseDriftAdjustmentRatio,
			NudgeBeta:   NudgeDriftAdjustmentRatio,
		})
	})
	s.eachSBConstraintState(func(_ int, constraint *sbConstraintState) {
		target := &s.bodyConstraintPlaceholders[constraint.body.reference.Index]
		constraint.logic.Reset(solver.Context{
			Target:      target,
			DeltaTime:   elapsedSeconds,
			ImpulseBeta: ImpulseDriftAdjustmentRatio,
			NudgeBeta:   NudgeDriftAdjustmentRatio,
		})
	})

	for i := 0; i < ImpulseIterationCount; i++ {
		s.eachDBConstraintState(func(_ int, constraint *dbConstraintState) {
			target := &s.bodyConstraintPlaceholders[constraint.primary.reference.Index]
			source := &s.bodyConstraintPlaceholders[constraint.secondary.reference.Index]
			constraint.logic.ApplyImpulses(solver.PairContext{
				Target:      target,
				Source:      source,
				DeltaTime:   elapsedSeconds,
				ImpulseBeta: ImpulseDriftAdjustmentRatio,
				NudgeBeta:   NudgeDriftAdjustmentRatio,
			})
		})
		s.eachSBConstraintState(func(_ int, constraint *sbConstraintState) {
			target := &s.bodyConstraintPlaceholders[constraint.body.reference.Index]
			constraint.logic.ApplyImpulses(solver.Context{
				Target:      target,
				DeltaTime:   elapsedSeconds,
				ImpulseBeta: ImpulseDriftAdjustmentRatio,
				NudgeBeta:   NudgeDriftAdjustmentRatio,
			})
		})
	}

	s.eachBodyState(func(i int, body *bodyState) {
		placeholder := &s.bodyConstraintPlaceholders[i]
		s.deinitPlaceholder(placeholder, body)
	})
}

func (s *Scene) applyMotion(elapsedSeconds float64) {
	defer metric.BeginRegion("motion").End()
	s.eachBodyState(func(_ int, body *bodyState) {
		body.ClampVelocity(s.maxLinearVelocity)
		body.ClampAngularVelocity(s.maxAngularVelocity)

		deltaPosition := dprec.Vec3Prod(body.velocity, elapsedSeconds)
		body.Translate(deltaPosition)
		deltaRotation := dprec.Vec3Prod(body.angularVelocity, elapsedSeconds)
		body.VectorRotate(deltaRotation)

		// FIXME
		s.bodyOctree.Update(body.itemID, body.position, body.bsRadius)
		body.InvalidateCollisionShapes(s)
	})
}

func (s *Scene) applyNudges(elapsedSeconds float64) {
	defer metric.BeginRegion("nudges").End()

	// TODO: Use Grow instead
	s.bodyConstraintPlaceholders = s.bodyConstraintPlaceholders[:0]
	for i := range s.bodies {
		placeholder := solver.Placeholder{}
		if body := &s.bodies[i]; body.IsActive() {
			s.initPlaceholder(&placeholder, body)
		}
		s.bodyConstraintPlaceholders = append(s.bodyConstraintPlaceholders, placeholder)
	}

	for i := 0; i < NudgeIterationCount; i++ {
		for _, constraint := range s.dbConstraints {
			if !constraint.IsActive() {
				continue
			}
			target := &s.bodyConstraintPlaceholders[constraint.primary.reference.Index]
			source := &s.bodyConstraintPlaceholders[constraint.secondary.reference.Index]
			ctx := solver.PairContext{
				Target:      target,
				Source:      source,
				DeltaTime:   elapsedSeconds,
				ImpulseBeta: ImpulseDriftAdjustmentRatio,
				NudgeBeta:   NudgeDriftAdjustmentRatio,
			}
			constraint.logic.Reset(ctx)
			constraint.logic.ApplyNudges(ctx)
		}
		for _, constraint := range s.sbConstraints {
			if !constraint.IsActive() {
				continue
			}
			target := &s.bodyConstraintPlaceholders[constraint.body.reference.Index]
			ctx := solver.Context{
				Target:      target,
				DeltaTime:   elapsedSeconds,
				ImpulseBeta: ImpulseDriftAdjustmentRatio,
				NudgeBeta:   NudgeDriftAdjustmentRatio,
			}
			constraint.logic.Reset(ctx)
			constraint.logic.ApplyNudges(ctx)
		}
	}

	for i := range s.bodies {
		placeholder := s.bodyConstraintPlaceholders[i]
		if body := &s.bodies[i]; body.IsActive() {
			s.deinitPlaceholder(&placeholder, body)
		}
	}
}

func (s *Scene) detectCollisions() {
	defer metric.BeginRegion("collision").End()

	s.eachBodyState(func(_ int, body *bodyState) {
		body.InvalidateCollisionShapes(s)
	})

	for _, constraint := range s.sbCollisionConstraints {
		constraint.Delete()
	}
	s.sbCollisionConstraints = s.sbCollisionConstraints[:0]
	s.sbCollisionSolvers = s.sbCollisionSolvers[:0]

	for _, constraint := range s.dbCollisionConstraints {
		constraint.Delete()
	}
	s.dbCollisionConstraints = s.dbCollisionConstraints[:0]
	s.dbCollisionSolvers = s.dbCollisionSolvers[:0]

	s.eachBodyState(func(_ int, primary *bodyState) {
		if primary.collisionSet.IsEmpty() {
			return
		}
		region := spatial.CuboidRegion(
			primary.position,
			dprec.NewVec3(primary.bsRadius*2.0, primary.bsRadius*2.0, primary.bsRadius*2.0),
		)

		s.propOctree.VisitHexahedronRegion(&region, spatial.VisitorFunc[uint32](func(propIndex uint32) {
			s.checkCollisionBodyWithProp(primary, &s.props[propIndex])
		}))

		s.bodyOctree.VisitHexahedronRegion(&region, spatial.VisitorFunc[uint32](func(secondaryIndex uint32) {
			secondary := &s.bodies[secondaryIndex]
			if secondary.reference.Index <= primary.reference.Index {
				return // secondary is the same or was already a primary
			}
			if secondary.collisionSet.IsEmpty() {
				return
			}
			if (primary.collisionGroup == secondary.collisionGroup) && (primary.collisionGroup != 0) {
				return
			}
			s.checkCollisionTwoBodies(primary, secondary)
		}))
	})
}

func (s *Scene) allocateGroundCollisionSolver() *constraint.Collision {
	if len(s.sbCollisionSolvers) < cap(s.sbCollisionSolvers) {
		s.sbCollisionSolvers = s.sbCollisionSolvers[:len(s.sbCollisionSolvers)+1]
	} else {
		s.sbCollisionSolvers = append(s.sbCollisionSolvers, constraint.Collision{})
	}
	return &s.sbCollisionSolvers[len(s.sbCollisionSolvers)-1]
}

func (s *Scene) allocateDualCollisionSolver() *constraint.PairCollision {
	if len(s.dbCollisionSolvers) < cap(s.dbCollisionSolvers) {
		s.dbCollisionSolvers = s.dbCollisionSolvers[:len(s.dbCollisionSolvers)+1]
	} else {
		s.dbCollisionSolvers = append(s.dbCollisionSolvers, constraint.PairCollision{})
	}
	return &s.dbCollisionSolvers[len(s.dbCollisionSolvers)-1]
}

func (s *Scene) checkCollisionBodyWithProp(primary *bodyState, prop *propState) {
	s.collisionSet.Reset()
	collision.CheckIntersectionSetWithSet(primary.collisionSet, prop.collisionSet, s.collisionSet)
	for _, intersection := range s.collisionSet.Intersections() {
		solver := s.allocateGroundCollisionSolver()
		solver.Init(constraint.CollisionState{
			BodyNormal:                 intersection.FirstDisplaceNormal,
			BodyPoint:                  intersection.FirstContact,
			BodyFrictionCoefficient:    primary.frictionCoefficient,
			BodyRestitutionCoefficient: primary.restitutionCoefficient,

			PropFrictionCoefficient:    1.0, // TODO: Take from prop or shape material
			PropRestitutionCoefficient: 0.5, // TODO: Take from prop or shape material

			Depth: intersection.Depth,
		})

		pair := sbCollisionPair{
			BodyRef: primary.reference,
			PropRef: prop.reference,
		}
		s.newSBCollisions[pair] = struct{}{}

		primaryBody := Body{
			scene:     s,
			reference: primary.reference,
		}
		s.sbCollisionConstraints = append(s.sbCollisionConstraints, s.CreateSingleBodyConstraint(primaryBody, solver))
	}
}

func (s *Scene) checkCollisionTwoBodies(primary, secondary *bodyState) {
	s.collisionSet.Reset()
	collision.CheckIntersectionSetWithSet(primary.collisionSet, secondary.collisionSet, s.collisionSet)
	for _, intersection := range s.collisionSet.Intersections() {
		solver := s.allocateDualCollisionSolver()
		solver.Init(constraint.PairCollisionState{
			PrimaryNormal:                 intersection.FirstDisplaceNormal,
			PrimaryPoint:                  intersection.FirstContact,
			PrimaryFrictionCoefficient:    primary.frictionCoefficient,
			PrimaryRestitutionCoefficient: primary.restitutionCoefficient,

			SecondaryNormal:                 intersection.SecondDisplaceNormal,
			SecondaryPoint:                  intersection.SecondContact,
			SecondaryFrictionCoefficient:    secondary.frictionCoefficient,
			SecondaryRestitutionCoefficient: secondary.restitutionCoefficient,

			Depth: intersection.Depth,
		})

		pair := dbCollisionPair{
			PrimaryRef:   primary.reference,
			SecondaryRef: secondary.reference,
		}
		s.newDBCollisions[pair] = struct{}{}

		primaryBody := Body{
			scene:     s,
			reference: primary.reference,
		}
		secondaryBody := Body{
			scene:     s,
			reference: secondary.reference,
		}
		s.dbCollisionConstraints = append(s.dbCollisionConstraints, s.CreateDoubleBodyConstraint(primaryBody, secondaryBody, solver))
	}
}

func (s *Scene) nextRevision() uint32 {
	s.freeRevision++
	return s.freeRevision
}

func (s *Scene) onTickInterval(elapsedTime time.Duration) {
	elapsedSeconds := elapsedTime.Seconds()
	s.notifyPreUpdate()
	s.runSimulation(elapsedSeconds * s.timeSpeed)
	s.notifySingleBodyCollisions()
	s.notifyDoubleBodyCollisions()
	s.notifyPostUpdate()
}

func (s *Scene) onTickLerp(alpha float64) {
	s.eachBodyState(func(_ int, body *bodyState) {
		body.intermediatePosition = dprec.Vec3Lerp(body.oldPosition, body.position, alpha)
		body.intermediateRotation = dprec.QuatSlerp(body.oldRotation, body.rotation, alpha)
	})
}

func (s *Scene) notifyPreUpdate() {
	s.preUpdateSubscriptions.Each(func(callback UpdateCallback) {
		callback(s.interval)
	})
}

func (s *Scene) notifyPostUpdate() {
	s.postUpdateSubscriptions.Each(func(callback UpdateCallback) {
		callback(s.interval)
	})
}

func (s *Scene) notifySingleBodyCollisions() {
	for newCollision := range s.newSBCollisions {
		if _, ok := s.oldSBCollisions[newCollision]; !ok {
			primary := Body{
				scene:     s,
				reference: newCollision.BodyRef,
			}
			prop := Prop{
				name: s.props[newCollision.PropRef.Index].name,
			}
			s.sbCollisionSubscriptions.Each(func(callback SingleBodyCollisionCallback) {
				callback(primary, prop, true)
			})
		}
	}
	for oldCollision := range s.oldSBCollisions {
		if _, ok := s.newSBCollisions[oldCollision]; !ok {
			primary := Body{
				scene:     s,
				reference: oldCollision.BodyRef,
			}
			prop := Prop{
				name: s.props[oldCollision.PropRef.Index].name,
			}
			s.sbCollisionSubscriptions.Each(func(callback SingleBodyCollisionCallback) {
				callback(primary, prop, false)
			})
		}
	}
	clear(s.oldSBCollisions)
	maps.Copy(s.oldSBCollisions, s.newSBCollisions)
	clear(s.newSBCollisions)
}

func (s *Scene) notifyDoubleBodyCollisions() {
	for newCollision := range s.newDBCollisions {
		if _, ok := s.oldDBCollisions[newCollision]; !ok {
			primary := Body{
				scene:     s,
				reference: newCollision.PrimaryRef,
			}
			secondary := Body{
				scene:     s,
				reference: newCollision.SecondaryRef,
			}
			s.dbCollisionSubscriptions.Each(func(callback DoubleBodyCollisionCallback) {
				callback(primary, secondary, true)
			})
		}
	}
	for oldCollision := range s.oldDBCollisions {
		if _, ok := s.newDBCollisions[oldCollision]; !ok {
			primary := Body{
				scene:     s,
				reference: oldCollision.PrimaryRef,
			}
			secondary := Body{
				scene:     s,
				reference: oldCollision.SecondaryRef,
			}
			s.dbCollisionSubscriptions.Each(func(callback DoubleBodyCollisionCallback) {
				callback(primary, secondary, false)
			})
		}
	}
	clear(s.oldDBCollisions)
	maps.Copy(s.oldDBCollisions, s.newDBCollisions)
	clear(s.newDBCollisions)
}

func (s *Scene) eachBodyState(cb func(index int, b *bodyState)) {
	for i := range s.bodies {
		if body := &s.bodies[i]; body.IsActive() {
			cb(i, body)
		}
	}
}

func (s *Scene) eachSBConstraintState(cb func(index int, constraint *sbConstraintState)) {
	for i := range s.sbConstraints {
		if constraint := &s.sbConstraints[i]; constraint.IsActive() {
			cb(i, constraint)
		}
	}
}

func (s *Scene) eachDBConstraintState(cb func(index int, constraint *dbConstraintState)) {
	for i := range s.dbConstraints {
		if constraint := &s.dbConstraints[i]; constraint.IsActive() {
			cb(i, constraint)
		}
	}
}

func (s *Scene) resolveBodyState(reference indexReference) *bodyState {
	state := &s.bodies[reference.Index]
	if !state.IsActive() || state.reference.Revision != reference.Revision {
		return nil
	}
	return state
}

func (s *Scene) initPlaceholder(placeholder *solver.Placeholder, body *bodyState) {
	placeholder.Init(solver.PlaceholderState{
		Mass:            body.mass,
		MomentOfInertia: body.momentOfInertia,
		LinearVelocity:  body.velocity,
		AngularVelocity: body.angularVelocity,
		Position:        body.position,
		Rotation:        body.rotation,
	})
}

func (s *Scene) deinitPlaceholder(placeholder *solver.Placeholder, body *bodyState) {
	body.velocity = placeholder.LinearVelocity()
	body.angularVelocity = placeholder.AngularVelocity()
	body.position = placeholder.Position()
	body.rotation = placeholder.Rotation()

	// FIXME
	s.bodyOctree.Update(body.itemID, body.position, body.bsRadius)
	body.InvalidateCollisionShapes(s)
}
