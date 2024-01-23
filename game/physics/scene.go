package physics

import (
	"time"

	"github.com/mokiat/gog/ds"
	"github.com/mokiat/gog/opt"
	"github.com/mokiat/gomath/dprec"
	"github.com/mokiat/lacking/debug/metric"
	"github.com/mokiat/lacking/game/physics/collision"
	"github.com/mokiat/lacking/game/physics/constraint"
	"github.com/mokiat/lacking/game/physics/solver"
	"github.com/mokiat/lacking/game/timestep"
	"github.com/mokiat/lacking/util/spatial"
	"golang.org/x/exp/maps"
)

func newScene(engine *Engine, interval time.Duration) *Scene {
	return &Scene{
		engine: engine,

		preUpdateSubscriptions:   NewUpdateSubscriptionSet(),
		postUpdateSubscriptions:  NewUpdateSubscriptionSet(),
		sbCollisionSubscriptions: NewStaticCollisionSubscriptionSet(),
		dbCollisionSubscriptions: NewDynamicCollisionSubscriptionSet(),

		timeSegmenter: timestep.NewSegmenter(interval),
		interval:      interval,
		timeSpeed:     1.0,

		maxAcceleration:        200.0,  // TODO: Measure something reasonable
		maxAngularAcceleration: 200.0,  // TODO: Measure something reasonable
		maxVelocity:            2000.0, // TODO: Measure something reasonable
		maxAngularVelocity:     2000.0, // TODO: Measure something reasonable

		propOctree: spatial.NewStaticOctree[uint32](spatial.StaticOctreeSettings{
			Size:                opt.V(32000.0),
			MaxDepth:            opt.V(int32(15)),
			BiasRatio:           opt.V(2.0),
			InitialNodeCapacity: opt.V(int32(4 * 1024)),
			InitialItemCapacity: opt.V(int32(8 * 1024)),
		}),

		bodyPool: ds.NewPool[Body](),
		bodyOctree: spatial.NewDynamicOctree[*Body](spatial.DynamicOctreeSettings{
			Size:                opt.V(32000.0),
			MaxDepth:            opt.V(int32(15)),
			BiasRatio:           opt.V(2.0),
			InitialNodeCapacity: opt.V(int32(4 * 1024)),
			InitialItemCapacity: opt.V(int32(8 * 1024)),
		}),
		dynamicBodies: make(map[*Body]struct{}),

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

		windVelocity: dprec.NewVec3(0.0, 0.0, 0.0),
		windDensity:  1.2,

		oldDBCollisions: make(map[uint64]dbCollisionPair, 32),
		newDBCollisions: make(map[uint64]dbCollisionPair, 32),

		collisionRevision: 1,
	}
}

// Scene represents a physics scene that contains
// a number of bodies that are independent on any
// bodies managed by other scene objects.
type Scene struct {
	engine *Engine

	preUpdateSubscriptions   *UpdateSubscriptionSet
	postUpdateSubscriptions  *UpdateSubscriptionSet
	sbCollisionSubscriptions *StaticCollisionSubscriptionSet
	dbCollisionSubscriptions *DynamicCollisionSubscriptionSet

	timeSegmenter *timestep.Segmenter
	interval      time.Duration
	timeSpeed     float64

	maxAcceleration        float64
	maxAngularAcceleration float64
	maxVelocity            float64
	maxAngularVelocity     float64

	props      []Prop
	propOctree *spatial.StaticOctree[uint32]

	// bodies        []Body // TODO
	bodyPool      *ds.Pool[Body]
	bodyOctree    *spatial.DynamicOctree[*Body]
	dynamicBodies map[*Body]struct{}

	// bodyAccelerators   []any // TOOD
	// areaAccelerators   []any // TODO
	globalAccelerators []globalAcceleratorState

	freeBodyAcceleratorIndices   *ds.Stack[uint32]
	freeAreaAcceleratorIndices   *ds.Stack[uint32]
	freeGlobalAcceleratorIndices *ds.Stack[uint32]

	sbConstraints []sbConstraintState
	dbConstraints []dbConstraintState

	freeSBConstraintIndices *ds.Stack[uint32]
	freeDBConstraintIndices *ds.Stack[uint32]

	sbCollisionConstraints []SBConstraint
	sbCollisionSolvers     []constraint.Collision
	dbCollisionConstraints []DBConstraint
	dbCollisionSolvers     []constraint.PairCollision
	collisionSet           *collision.IntersectionBucket

	oldDBCollisions map[uint64]dbCollisionPair
	newDBCollisions map[uint64]dbCollisionPair

	windVelocity dprec.Vec3
	windDensity  float64

	freeRevision uint32

	collisionRevision int
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
func (s *Scene) SubscribeSingleBodyCollision(callback StaticCollisionCallback) *StaticCollisionSubscription {
	return s.sbCollisionSubscriptions.Subscribe(callback)
}

// SubscribeDoubleBodyCollision registers a callback that is invoked when two
// bodies collide.
func (s *Scene) SubscribeDoubleBodyCollision(callback DynamicCollisionCallback) *DynamicCollisionSubscription {
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

// CreateGlobalAccelerator creates a new accelerator that affects the whole
// scene.
func (s *Scene) CreateGlobalAccelerator(logic solver.Acceleration) GlobalAccelerator {
	return createGlobalAccelerator(s, logic)
}

// CreateProp creates a new static Prop. A prop is an object
// that is static and rarely removed.
func (s *Scene) CreateProp(info PropInfo) {
	bs := info.CollisionSet.BoundingSphere()
	position := bs.Position()
	radius := bs.Radius()

	propIndex := uint32(len(s.props))
	s.props = append(s.props, Prop{
		name:         info.Name,
		collisionSet: info.CollisionSet,
	})
	s.propOctree.Insert(position, radius, propIndex)
}

// CreateBody creates a new physics body and places
// it within this scene.
func (s *Scene) CreateBody(info BodyInfo) *Body {
	body := s.bodyPool.Fetch()
	body.scene = s
	body.itemID = s.bodyOctree.Insert(
		info.Position, 1.0, body,
	)
	body.id = freeBodyID
	freeBodyID++

	// TODO: Don't use Setters, since these are intended for external use
	// and not all should be allowed as well as some might have invalidation
	// logic that we don't need to trigger here.
	def := info.Definition
	body.definition = def
	body.SetName(info.Name)
	body.SetPosition(info.Position)
	body.SetOrientation(info.Rotation)

	body.SetMass(def.mass)
	body.SetMomentOfInertia(def.momentOfInertia)

	// TODO: Move these to Material and have Material be assigned
	// to the collision shapes instead of the body.
	body.frictionCoefficient = def.frictionCoefficient
	body.SetRestitutionCoefficient(def.restitutionCoefficient)

	body.SetDragFactor(def.dragFactor)
	body.SetAngularDragFactor(def.angularDragFactor)
	body.SetCollisionGroup(def.collisionGroup)
	body.SetAerodynamicShapes(def.aerodynamicShapes)

	body.invalidateCollisionShapes()

	s.dynamicBodies[body] = struct{}{}
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
func (s *Scene) CreateSingleBodyConstraint(body *Body, logic solver.Constraint) SBConstraint {
	return createSBConstraint(s, logic, body)
}

// CreateDoubleBodyConstraint creates a new physics constraint that acts on
// two bodies and enables it for this scene.
func (s *Scene) CreateDoubleBodyConstraint(primary, secondary *Body, logic solver.PairConstraint) DBConstraint {
	return createDBConstraint(s, logic, primary, secondary)
}

// Update runs a number of physics iterations until the specified duration has
// been reached or surpassed (yes physics can get ahead).
func (s *Scene) Update(elapsedTime time.Duration) {
	s.timeSegmenter.Update(elapsedTime, s.onTickInterval, s.onTickLerp)
}

func (s *Scene) onTickInterval(elapsedTime time.Duration) {
	elapsedSeconds := elapsedTime.Seconds()
	s.notifyPreUpdate()
	s.runSimulation(elapsedSeconds * s.timeSpeed)
	s.notifyDynamicCollisions()
	s.notifyPostUpdate()
}

func (s *Scene) onTickLerp(alpha float64) {
	for body := range s.dynamicBodies {
		body.lerpPosition = dprec.Vec3Lerp(body.oldPosition, body.position, alpha)
		body.slerpOrientation = dprec.QuatSlerp(body.oldOrientation, body.orientation, alpha)
	}
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

func (s *Scene) notifyDynamicCollisions() {
	for newHash, newPair := range s.newDBCollisions {
		if _, ok := s.oldDBCollisions[newHash]; !ok {
			s.dbCollisionSubscriptions.Each(func(callback DynamicCollisionCallback) {
				callback(newPair.First, newPair.Second, true)
			})
		}
	}
	for oldHash, oldPair := range s.oldDBCollisions {
		if _, ok := s.newDBCollisions[oldHash]; !ok {
			s.dbCollisionSubscriptions.Each(func(callback DynamicCollisionCallback) {
				callback(oldPair.First, oldPair.Second, false)
			})
		}
	}
	clear(s.oldDBCollisions)
	maps.Copy(s.oldDBCollisions, s.newDBCollisions)
	clear(s.newDBCollisions)
}

func (s *Scene) Each(cb func(b *Body)) {
	for body := range s.dynamicBodies {
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
	s.propOctree = nil
	s.props = nil

	s.bodyPool = nil
	s.dynamicBodies = nil

	s.globalAccelerators = nil
	s.freeBodyAcceleratorIndices = nil
	s.freeAreaAcceleratorIndices = nil
	s.freeGlobalAcceleratorIndices = nil

	s.sbConstraints = nil
	s.dbConstraints = nil
	s.freeSBConstraintIndices = nil
	s.freeDBConstraintIndices = nil
}

func (s *Scene) runSimulation(elapsedSeconds float64) {
	for body := range s.dynamicBodies {
		body.oldPosition = body.position
		body.oldOrientation = body.orientation
	}
	if elapsedSeconds > 0.0001 {
		s.applyForces()
		s.applyAcceleration(elapsedSeconds)
		s.applyImpulses(elapsedSeconds)
		s.applyMotion(elapsedSeconds)
		s.applyNudges(elapsedSeconds)
		s.detectCollisions()
	}
}

func (s *Scene) applyForces() {
	defer metric.BeginRegion("forces").End()

	for body := range s.dynamicBodies {
		body.resetAcceleration()
		body.resetAngularAcceleration()
	}

	// TODO: The target needs to be global across all accelerator types.
	// The current implementation below works only while there is a single
	// accelerator type.

	for body := range s.dynamicBodies {
		target := solver.NewAccelerationTarget(
			body.mass,
			body.momentOfInertia,
			body.position,
			body.orientation,
			body.velocity,
			body.angularVelocity,
		)

		for _, accelerator := range s.globalAccelerators {
			if !accelerator.reference.IsValid() || !accelerator.enabled {
				continue
			}
			accelerator.logic.ApplyAcceleration(solver.AccelerationContext{
				Target: &target,
			})
		}

		// TODO: These should be moved outside, after all accelerator types have
		// had their way with the targets. Furthermore, it should apply directly
		// on the velocity, skipping the applyAcceleration step and
		// getting rid of the acceleration fields on the body.

		body.addAcceleration(target.AccumulatedLinearAcceleration())
		body.addAngularAcceleration(target.AccumulatedAngularAcceleration())
	}

	for body := range s.dynamicBodies {
		deltaWindVelocity := dprec.Vec3Diff(s.windVelocity, body.velocity)
		dragForce := dprec.Vec3Prod(deltaWindVelocity, deltaWindVelocity.Length()*s.windDensity*body.dragFactor)
		body.applyForce(dragForce)

		angularDragForce := dprec.Vec3Prod(body.angularVelocity, -body.angularVelocity.Length()*s.windDensity*body.angularDragFactor)
		body.applyTorque(angularDragForce)

		if len(body.aerodynamicShapes) > 0 {
			bodyTransform := NewTransform(body.position, body.orientation)

			for _, aerodynamicShape := range body.aerodynamicShapes {
				aerodynamicShape = aerodynamicShape.Transformed(bodyTransform)
				relativeWindSpeed := dprec.QuatVec3Rotation(dprec.InverseQuat(aerodynamicShape.Rotation()), deltaWindVelocity)

				// TODO: Apply at offset
				force := aerodynamicShape.solver.Force(relativeWindSpeed)
				absoluteForce := dprec.QuatVec3Rotation(aerodynamicShape.Rotation(), force)
				body.applyForce(absoluteForce)
				// body.applyOffsetForce(absoluteForce, aerodynamicShape.Position())
			}
		}
	}

	// TODO: Apply custom force fields
}

func (s *Scene) applyAcceleration(elapsedSeconds float64) {
	defer metric.BeginRegion("acceleration").End()
	for body := range s.dynamicBodies {
		body.clampAcceleration(s.maxAcceleration)
		body.clampAngularAcceleration(s.maxAngularAcceleration)

		deltaVelocity := dprec.Vec3Prod(body.acceleration, elapsedSeconds)
		body.addVelocity(deltaVelocity)
		deltaAngularVelocity := dprec.Vec3Prod(body.angularAcceleration, elapsedSeconds)
		body.addAngularVelocity(deltaAngularVelocity)
	}
}

var (
	bodies = make([]*Body, 0, 1024)

	placeholders = make([]solver.Placeholder, 0, 1024)

	// TODO: Maybe assign and unassign to/from a Body
	bodyToPlaceholder = make(map[*Body]*solver.Placeholder)
)

func initPlaceholder(placeholder *solver.Placeholder, body *Body) {
	placeholder.Init(solver.PlaceholderState{
		Mass:            body.mass,
		MomentOfInertia: body.momentOfInertia,
		LinearVelocity:  body.velocity,
		AngularVelocity: body.angularVelocity,
		Position:        body.position,
		Rotation:        body.orientation,
	})
}

func deinitPlaceholder(placeholder *solver.Placeholder, body *Body) {
	body.SetVelocity(placeholder.LinearVelocity())
	body.SetAngularVelocity(placeholder.AngularVelocity())
	body.SetPosition(placeholder.Position())
	body.SetOrientation(placeholder.Rotation())
}

func (s *Scene) applyImpulses(elapsedSeconds float64) {
	defer metric.BeginRegion("impulses").End()

	bodies = bodies[:0]
	placeholders = placeholders[:0]
	maps.Clear(bodyToPlaceholder)

	for _, constraint := range s.dbConstraints {
		if !constraint.CanUse() {
			continue
		}
		target := constraint.primary
		if _, ok := bodyToPlaceholder[target]; !ok {
			placeholders = append(placeholders, solver.Placeholder{})
			placeholder := &placeholders[len(placeholders)-1]
			initPlaceholder(placeholder, target)
			bodyToPlaceholder[target] = placeholder
			bodies = append(bodies, target)
		}
		source := constraint.secondary
		if _, ok := bodyToPlaceholder[source]; !ok {
			placeholders = append(placeholders, solver.Placeholder{})
			placeholder := &placeholders[len(placeholders)-1]
			initPlaceholder(placeholder, source)
			bodyToPlaceholder[source] = placeholder
			bodies = append(bodies, source)
		}
		constraint.logic.Reset(solver.PairContext{
			Target:      bodyToPlaceholder[constraint.primary],
			Source:      bodyToPlaceholder[constraint.secondary],
			DeltaTime:   elapsedSeconds,
			ImpulseBeta: ImpulseDriftAdjustmentRatio,
			NudgeBeta:   NudgeDriftAdjustmentRatio,
		})
	}
	for _, constraint := range s.sbConstraints {
		if !constraint.CanUse() {
			continue
		}
		target := constraint.body
		if _, ok := bodyToPlaceholder[target]; !ok {
			placeholders = append(placeholders, solver.Placeholder{})
			placeholder := &placeholders[len(placeholders)-1]
			initPlaceholder(placeholder, target)
			bodyToPlaceholder[target] = placeholder
			bodies = append(bodies, target)
		}
		constraint.logic.Reset(solver.Context{
			Target:      bodyToPlaceholder[constraint.body],
			DeltaTime:   elapsedSeconds,
			ImpulseBeta: ImpulseDriftAdjustmentRatio,
			NudgeBeta:   NudgeDriftAdjustmentRatio,
		})
	}

	for i := 0; i < ImpulseIterationCount; i++ {
		for _, constraint := range s.dbConstraints {
			if !constraint.CanUse() {
				continue
			}
			constraint.logic.ApplyImpulses(solver.PairContext{
				Target:      bodyToPlaceholder[constraint.primary],
				Source:      bodyToPlaceholder[constraint.secondary],
				DeltaTime:   elapsedSeconds,
				ImpulseBeta: ImpulseDriftAdjustmentRatio,
				NudgeBeta:   NudgeDriftAdjustmentRatio,
			})
		}
		for _, constraint := range s.sbConstraints {
			if !constraint.CanUse() {
				continue
			}
			constraint.logic.ApplyImpulses(solver.Context{
				Target:      bodyToPlaceholder[constraint.body],
				DeltaTime:   elapsedSeconds,
				ImpulseBeta: ImpulseDriftAdjustmentRatio,
				NudgeBeta:   NudgeDriftAdjustmentRatio,
			})
		}
	}

	for _, body := range bodies {
		deinitPlaceholder(bodyToPlaceholder[body], body)
	}
}

func (s *Scene) applyMotion(elapsedSeconds float64) {
	defer metric.BeginRegion("motion").End()
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
	defer metric.BeginRegion("nudges").End()

	for _, body := range bodies {
		initPlaceholder(bodyToPlaceholder[body], body)
	}

	for i := 0; i < NudgeIterationCount; i++ {
		for _, constraint := range s.dbConstraints {
			if !constraint.CanUse() {
				continue
			}
			ctx := solver.PairContext{
				Target:      bodyToPlaceholder[constraint.primary],
				Source:      bodyToPlaceholder[constraint.secondary],
				DeltaTime:   elapsedSeconds,
				ImpulseBeta: ImpulseDriftAdjustmentRatio,
				NudgeBeta:   NudgeDriftAdjustmentRatio,
			}
			constraint.logic.Reset(ctx)
			constraint.logic.ApplyNudges(ctx)
		}
		for _, constraint := range s.sbConstraints {
			if !constraint.CanUse() {
				continue
			}
			ctx := solver.Context{
				Target:      bodyToPlaceholder[constraint.body],
				DeltaTime:   elapsedSeconds,
				ImpulseBeta: ImpulseDriftAdjustmentRatio,
				NudgeBeta:   NudgeDriftAdjustmentRatio,
			}
			constraint.logic.Reset(ctx)
			constraint.logic.ApplyNudges(ctx)
		}
	}

	for _, body := range bodies {
		deinitPlaceholder(bodyToPlaceholder[body], body)
	}
}

func (s *Scene) detectCollisions() {
	defer metric.BeginRegion("collision").End()
	s.collisionRevision++

	for body := range s.dynamicBodies {
		body.invalidateCollisionShapes()
	}

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

	for primary := range s.dynamicBodies {
		primary.revision = s.collisionRevision
		if primary.collisionSet.IsEmpty() {
			continue
		}
		region := spatial.CuboidRegion(
			primary.position,
			dprec.NewVec3(primary.bsRadius*2.0, primary.bsRadius*2.0, primary.bsRadius*2.0),
		)

		s.propOctree.VisitHexahedronRegion(&region, spatial.VisitorFunc[uint32](func(propIndex uint32) {
			s.checkCollisionBodyWithProp(primary, &s.props[propIndex])
		}))

		s.bodyOctree.VisitHexahedronRegion(&region, spatial.VisitorFunc[*Body](func(secondary *Body) {
			if secondary == primary {
				return
			}
			if secondary.revision == s.collisionRevision {
				return // secondary already processed
			}
			if secondary.collisionSet.IsEmpty() {
				return
			}
			if (primary.collisionGroup == secondary.collisionGroup) && (primary.collisionGroup != 0) {
				return
			}
			s.checkCollisionTwoBodies(primary, secondary)
		}))
	}
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

func (s *Scene) checkCollisionBodyWithProp(primary *Body, prop *Prop) {
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
		s.sbCollisionConstraints = append(s.sbCollisionConstraints, s.CreateSingleBodyConstraint(primary, solver))
	}
}

func (s *Scene) checkCollisionTwoBodies(primary, secondary *Body) {
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
			First:  primary,
			Second: secondary,
		}
		s.newDBCollisions[pair.Hash()] = pair

		s.dbCollisionConstraints = append(s.dbCollisionConstraints, s.CreateDoubleBodyConstraint(primary, secondary, solver))
	}
}

func (s *Scene) nextRevision() uint32 {
	s.freeRevision++
	return s.freeRevision
}

type dbCollisionPair struct {
	First  *Body
	Second *Body
}

func (p dbCollisionPair) Hash() uint64 {
	hash1 := uint64(p.First.id)
	hash2 := uint64(p.Second.id)
	if hash1 < hash2 {
		return hash1 + uint64(hash2)*0xFFFFFFFF
	} else {
		return hash2 + uint64(hash1)*0xFFFFFFFF
	}
}
