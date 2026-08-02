package physics

import (
	"maps"
	"math"
	"time"

	"github.com/mokiat/gog/ds"
	"github.com/mokiat/gog/opt"
	"github.com/mokiat/gomath/dprec"
	"github.com/mokiat/lacking/core/spatial/placement3d"
	"github.com/mokiat/lacking/core/spatial/shape3d"
	"github.com/mokiat/lacking/debug/metric"
	"github.com/mokiat/lacking/game/physics/constraint"
	"github.com/mokiat/lacking/game/physics/solver"
	"github.com/mokiat/lacking/util/observer"
)

// Scene represents a physics scene that contains
// a number of bodies that are independent on any
// bodies managed by other scene objects.
type Scene struct {
	collisionScene *placement3d.Scene[bodyData, shapeData, propRef]

	sbCollisionSubscriptions *observer.SubscriptionSet[SoloBodyCollisionCallback]
	dbCollisionSubscriptions *observer.SubscriptionSet[PairBodyCollisionCallback]

	timeSpeed float64

	props []propState

	sbCollisionConstraints []SBConstraint
	sbCollisionSolvers     []constraint.Collision

	dbCollisionConstraints []DBConstraint
	dbCollisionSolvers     []constraint.PairCollision

	collisionSet placement3d.ContactList

	oldSBCollisions map[sbCollisionPair]struct{}
	newSBCollisions map[sbCollisionPair]struct{}

	oldDBCollisions map[dbCollisionPair]struct{}
	newDBCollisions map[dbCollisionPair]struct{}

	freeCollisionRejectGroup uint32

	// ---------- NEW BELOW---------- (TODO: REMOVE COMMENT)
	mediumSolver MediumSolver

	freeGlobalAcceleratorIndices *ds.Stack[int32]
	freeBodyAcceleratorIndices   *ds.Stack[int32]
	freeSoloConstraintIndices    *ds.Stack[int32]
	freePairConstraintIndices    *ds.Stack[int32]
	freeBodyIndices              *ds.Stack[int32]

	globalAccelerators []globalAcceleratorState
	bodyAccelerators   []bodyAcceleratorState
	soloConstraints    []soloConstraintState
	pairConstraints    []pairConstraintState
	bodies             []bodyState

	maxLinearAcceleration  float64
	maxAngularAcceleration float64
	maxLinearVelocity      float64
	maxAngularVelocity     float64
}

func NewScene() *Scene {
	return &Scene{
		collisionScene: placement3d.NewScene[bodyData, shapeData, propRef](placement3d.SceneSettings{
			Size:                opt.V(16384.0),
			MaxDepth:            opt.V[uint32](12),
			InitialNodeCapacity: opt.V[uint32](1024),
			InitialItemCapacity: opt.V[uint32](1024),
		}),

		sbCollisionSubscriptions: observer.NewSubscriptionSet[SoloBodyCollisionCallback](),
		dbCollisionSubscriptions: observer.NewSubscriptionSet[PairBodyCollisionCallback](),

		timeSpeed: 1.0,

		props: make([]propState, 0, 1024),

		collisionSet: make(placement3d.ContactList, 0, 128),

		oldSBCollisions: make(map[sbCollisionPair]struct{}, 32),
		newSBCollisions: make(map[sbCollisionPair]struct{}, 32),

		oldDBCollisions: make(map[dbCollisionPair]struct{}, 32),
		newDBCollisions: make(map[dbCollisionPair]struct{}, 32),

		// ---------- NEW BELOW---------- (TODO: REMOVE COMMENT)
		mediumSolver: NewStaticAirSolver(),

		freeGlobalAcceleratorIndices: ds.EmptyStack[int32](),
		freeBodyAcceleratorIndices:   ds.EmptyStack[int32](),
		freeSoloConstraintIndices:    ds.EmptyStack[int32](),
		freePairConstraintIndices:    ds.EmptyStack[int32](),
		freeBodyIndices:              ds.EmptyStack[int32](),

		globalAccelerators: make([]globalAcceleratorState, 0),
		bodyAccelerators:   make([]bodyAcceleratorState, 0),
		soloConstraints:    make([]soloConstraintState, 0),
		pairConstraints:    make([]pairConstraintState, 0),
		bodies:             make([]bodyState, 0),

		maxLinearAcceleration:  math.MaxFloat64,
		maxAngularAcceleration: math.MaxFloat64,
		maxLinearVelocity:      math.MaxFloat64,
		maxAngularVelocity:     math.MaxFloat64,
	}
}

// MediumSolver returns the solver that is used to calculate the medium
// properties of the scene.
//
// The returned solver is never nil. A scene starts off with a default
// [StaticAirSolver].
func (s *Scene) MediumSolver() MediumSolver {
	return s.mediumSolver
}

// SetMediumSolver changes the solver that is used to calculate the medium
// properties of the scene.
//
// Passing nil is not an error and resets the scene to a default
// [StaticAirSolver], since the scene always needs a medium to sample.
func (s *Scene) SetMediumSolver(solver MediumSolver) {
	if solver != nil {
		s.mediumSolver = solver
	} else {
		s.mediumSolver = NewStaticAirSolver()
	}
}

// GlobalAccelerators returns a [GlobalAcceleratorView] through which the
// global accelerators of this scene can be created and managed.
func (s *Scene) GlobalAccelerators() GlobalAcceleratorView {
	return GlobalAcceleratorView{
		scene: s,
	}
}

// BodyAccelerators returns a [BodyAcceleratorView] through which the body
// accelerators of this scene can be created and managed.
func (s *Scene) BodyAccelerators() BodyAcceleratorView {
	return BodyAcceleratorView{
		scene: s,
	}
}

// SoloConstraints returns a [SoloConstraintView] through which the solo
// constraints of this scene can be created and managed.
func (s *Scene) SoloConstraints() SoloConstraintView {
	return SoloConstraintView{
		scene: s,
	}
}

// PairConstraints returns a [PairConstraintView] through which the pair
// constraints of this scene can be created and managed.
func (s *Scene) PairConstraints() PairConstraintView {
	return PairConstraintView{
		scene: s,
	}
}

// Bodies returns a [BodyView] through which the bodies of this scene can be
// created and managed.
func (s *Scene) Bodies() BodyView {
	return BodyView{
		scene: s,
	}
}

// MaxLinearAcceleration returns the maximum magnitude that the linear
// acceleration of a body can reach. Accelerations that exceed it are
// clamped on every simulation step.
//
// Defaults to math.MaxFloat64, which is effectively unbounded.
func (s *Scene) MaxLinearAcceleration() float64 {
	return s.maxLinearAcceleration
}

// SetMaxLinearAcceleration changes the maximum magnitude that the linear
// acceleration of a body can reach.
func (s *Scene) SetMaxLinearAcceleration(acceleration float64) {
	s.maxLinearAcceleration = acceleration
}

// MaxAngularAcceleration returns the maximum magnitude that the angular
// acceleration of a body can reach. Accelerations that exceed it are
// clamped on every simulation step.
//
// Defaults to math.MaxFloat64, which is effectively unbounded.
func (s *Scene) MaxAngularAcceleration() float64 {
	return s.maxAngularAcceleration
}

// SetMaxAngularAcceleration changes the maximum magnitude that the angular
// acceleration of a body can reach.
func (s *Scene) SetMaxAngularAcceleration(acceleration float64) {
	s.maxAngularAcceleration = acceleration
}

// MaxLinearVelocity returns the maximum magnitude that the linear velocity
// of a body can reach. Velocities that exceed it are clamped on every
// simulation step.
//
// Defaults to math.MaxFloat64, which is effectively unbounded.
func (s *Scene) MaxLinearVelocity() float64 {
	return s.maxLinearVelocity
}

// SetMaxLinearVelocity changes the maximum magnitude that the linear
// velocity of a body can reach.
func (s *Scene) SetMaxLinearVelocity(velocity float64) {
	s.maxLinearVelocity = velocity
}

// MaxAngularVelocity returns the maximum magnitude that the angular
// velocity of a body can reach. Velocities that exceed it are clamped on
// every simulation step.
//
// Defaults to math.MaxFloat64, which is effectively unbounded.
func (s *Scene) MaxAngularVelocity() float64 {
	return s.maxAngularVelocity
}

// SetMaxAngularVelocity changes the maximum magnitude that the angular
// velocity of a body can reach.
func (s *Scene) SetMaxAngularVelocity(velocity float64) {
	s.maxAngularVelocity = velocity
}

/////// OLD BELOW ------------ (TODO: DELETE COMMENT)

// SubscribeSingleBodyCollision registers a callback that is invoked when a body
// collides with a static object.
func (s *Scene) SubscribeSingleBodyCollision(callback SoloBodyCollisionCallback) *SoloBodyCollisionSubscription {
	return s.sbCollisionSubscriptions.Subscribe(callback)
}

// SubscribeDoubleBodyCollision registers a callback that is invoked when two
// bodies collide.
func (s *Scene) SubscribeDoubleBodyCollision(callback PairBodyCollisionCallback) *PairBodyCollisionSubscription {
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

// NextCollisionRejectGroup returns a collision reject group that is unique
// within this Scene. Bodies that are assigned the same reject group do not
// collide with each other, which is useful for objects that are meant to
// overlap, such as the chassis and the wheels of a vehicle.
//
// The returned value is always larger than zero, since zero indicates that a
// body does not belong to any reject group and hence can collide with
// everything.
//
// Reject groups are never recycled. Each call returns a new value, even if all
// bodies that used a previously returned group have been deleted.
func (s *Scene) NextCollisionRejectGroup() uint32 {
	s.freeCollisionRejectGroup++
	return s.freeCollisionRejectGroup
}

// CreateProp creates a new static Prop. A prop is an object
// that is static and rarely removed.
func (s *Scene) CreateProp(info PropInfo) {
	// objectID := s.shapeScene.CreateObject(placement3d.ObjectInfo[internalRef]{
	// 	Position: info.Position,
	// 	Rotation: info.Rotation,
	// 	UserData: internalRef{
	// 		index:  propIndex,
	// 		isProp: true,
	// 	},
	// })
	for _, mesh := range info.CollisionMeshes {
		propIndex := uint32(len(s.props))

		meshID := s.collisionScene.CreateMesh(placement3d.MeshInfo[propRef]{
			Position: info.Position,
			Rotation: info.Rotation,
			Mesh:     mesh,
			UserData: propRef{
				index: propIndex,
			},
		})

		s.props = append(s.props, propState{
			reference: newIndexReference(propIndex, s.nextRevision()),
			meshID:    meshID,
			name:      info.Name,
		})
	}
}

// Update runs a single physics iteration. This method should be called with
// fixed elapsed times, otherwise the physics may break.
func (s *Scene) Update(elapsedTime time.Duration) {
	elapsedSeconds := elapsedTime.Seconds()
	s.runSimulation(elapsedSeconds * s.timeSpeed)
	s.notifySingleBodyCollisions()
	s.notifyDoubleBodyCollisions()
}

func (s *Scene) Each(cb func(b Body)) {
	s.eachBody(func(_ int, b *bodyState) {
		cb(Body{
			scene:     s,
			reference: b.reference,
		})
	})
}

func (s *Scene) CheckSegmentIntersection(segment shape3d.Segment, mask uint32) (BodyID, bool) {
	intersection, ok := s.collisionScene.CheckSegmentIntersection(segment, placement3d.Filter{
		Mask: opt.V(mask),
	})
	if !ok {
		return NilBodyID, false
	}
	if intersection.TargetShapeID == placement3d.InvalidShapeID {
		// A prop.
		return NilBodyID, false // FIXME: This should handle props as well.
	}
	objectID := s.collisionScene.GetShapeObject(intersection.TargetShapeID)
	bData := s.collisionScene.GetObjectUserData(objectID)
	body := &s.bodies[bData.index]
	return BodyID{
		index:    bData.index,
		revision: body.revision,
	}, true
}

// func (s *Scene) Nearby(body Body, distance float64, cb func(b Body)) {
// 	state := s.resolveBodyState(body.reference)
// 	if state == nil {
// 		return
// 	}
// 	region := spatial.CuboidRegion(
// 		state.position,
// 		dprec.NewVec3(distance, distance, distance),
// 	)
// 	s.bodyOctree.VisitHexahedronRegion(&region, spatial.VisitorFunc[uint32](func(candidate uint32) {
// 		candidateState := &s.bodies[candidate]
// 		if candidateState != state {
// 			cb(Body{
// 				scene:     s,
// 				reference: candidateState.reference,
// 			})
// 		}
// 	}))
// }

func (s *Scene) runSimulation(elapsedSeconds float64) {
	if elapsedSeconds > 0.0001 {
		s.applyAcceleration(elapsedSeconds)
		s.applyImpulses(elapsedSeconds)
		s.applyMotion(elapsedSeconds)
		s.applyNudges(elapsedSeconds)
		s.applyPlacement()
		s.detectCollisions()
	}
}

func (s *Scene) applyAcceleration(elapsedSeconds float64) {
	defer metric.BeginRegion("acceleration").End()

	s.eachBody(func(index int, body *bodyState) {
		// Create acceleration context.
		ctx := AccelerationContext{
			DeltaSeconds:   elapsedSeconds,
			MediumVelocity: s.mediumSolver.Velocity(body.position),
			MediumDensity:  s.mediumSolver.Density(body.position),
		}
		target := newAccelerationTarget(body)

		// Reset accumulated accelerations.
		body.linearAcceleration = dprec.ZeroVec3()
		body.angularAcceleration = dprec.ZeroVec3()

		// Apply global accelerators.
		s.eachEnabledGlobalAccelerator(func(_ int, accelerator *globalAcceleratorState) {
			accelerator.solver.ApplyAcceleration(ctx, target)
		})

		// Apply body accelerators.
		// TODO: Implement body accelerators.

		// Constrain the accumulated accelerations to the maximum allowed values.
		body.clampLinearAcceleration(s.maxLinearAcceleration)
		body.clampAngularAcceleration(s.maxAngularAcceleration)

		// Update the body's velocity based on the accumulated accelerations.
		body.addLinearVelocity(dprec.Vec3Prod(body.linearAcceleration, elapsedSeconds))
		body.addAngularVelocity(dprec.Vec3Prod(body.angularAcceleration, elapsedSeconds))
	})
}

// func (s *Scene) applyAerodynamicAccelerations() {
// 	s.eachBody(func(index int, body *bodyState) {
// 		if len(body.aerodynamicShapes) == 0 {
// 			return
// 		}
// 		target := &s.bodyAccelerationTargets[index]
// 		mediumDensity := s.mediumSolver.Density(body.position)
// 		mediumVelocity := s.mediumSolver.Velocity(body.position)

// 		deltaVelocity := dprec.Vec3Diff(mediumVelocity, body.velocity)
// 		dragForce := dprec.Vec3Prod(deltaVelocity, deltaVelocity.Length()*mediumDensity*body.dragFactor)
// 		target.ApplyForce(dragForce)

// 		angularDragForce := dprec.Vec3Prod(body.angularVelocity, -body.angularVelocity.Length()*mediumDensity*body.angularDragFactor)
// 		target.ApplyTorque(angularDragForce)

// 		bodyTransform := NewTransform(body.position, body.rotation)
// 		for _, aerodynamicShape := range body.aerodynamicShapes {
// 			// TODO: Take shape velocity into account. This also means that wings should be
// 			// split into two, to benefit from that.

// 			aerodynamicShape = aerodynamicShape.Transformed(bodyTransform)
// 			relativeSpeed := dprec.QuatVec3Rotation(dprec.InverseQuat(aerodynamicShape.Rotation()), deltaVelocity)

// 			force := aerodynamicShape.solver.Force(relativeSpeed, mediumDensity)
// 			absoluteForce := dprec.QuatVec3Rotation(aerodynamicShape.Rotation(), force)

// 			offset := dprec.Vec3Diff(aerodynamicShape.Position(), bodyTransform.Position())
// 			target.ApplyOffsetForce(offset, absoluteForce)
// 			// target.ApplyOffsetForce(absoluteForce, aerodynamicShape.Position())
// 		}
// 	})
// }

func (s *Scene) applyImpulses(elapsedSeconds float64) {
	defer metric.BeginRegion("impulses").End()

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

	for range ImpulseIterationCount {
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
}

func (s *Scene) applyMotion(elapsedSeconds float64) {
	defer metric.BeginRegion("motion").End()

	s.eachBody(func(_ int, body *bodyState) {
		// Clamp the velocity to the maximum allowed values.
		body.clampLinearVelocity(s.maxLinearVelocity)
		body.clampAngularVelocity(s.maxAngularVelocity)

		// Apply the velocity to the body's position and rotation.
		body.translate(dprec.Vec3Prod(body.linearVelocity, elapsedSeconds))
		body.rotate(QuatFromVector(dprec.Vec3Prod(body.angularVelocity, elapsedSeconds)))
	})
}

func (s *Scene) applyNudges(elapsedSeconds float64) {
	defer metric.BeginRegion("nudges").End()

	for range NudgeIterationCount {
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
}

func (s *Scene) applyPlacement() {
	defer metric.BeginRegion("placement").End()

	s.eachBody(func(_ int, body *bodyState) {
		// Update the collision scene with the new position and rotation of the body.
		s.collisionScene.SetObjectTransform(body.objectID, shape3d.Transform{
			Translation: body.position,
			Rotation:    shape3d.RotationFromQuat(body.rotation),
		})
	})
}

func (s *Scene) detectCollisions() {
	defer metric.BeginRegion("collision").End()

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

	s.collisionSet.Reset()
	s.collisionScene.CollectIntersections(s.collisionSet.AddContact)
	for _, intersection := range s.collisionSet.Contacts() {
		srcBodyObject := s.collisionScene.GetShapeObject(intersection.SourceShapeID)
		srcBodyRef := s.collisionScene.GetObjectUserData(srcBodyObject)

		if intersection.TargetMeshID == placement3d.InvalidMeshID {
			tgtBodyObject := s.collisionScene.GetShapeObject(intersection.TargetShapeID)
			tgtBodyRef := s.collisionScene.GetObjectUserData(tgtBodyObject)
			s.detectBodyBodyCollision(srcBodyRef.index, tgtBodyRef.index, intersection)
		} else {
			tgtPropMesh := s.collisionScene.GetMeshUserData(intersection.TargetMeshID)
			s.detectBodyPropCollision(srcBodyRef.index, tgtPropMesh.index, intersection)
		}
	}
}

func (s *Scene) detectBodyBodyCollision(primaryIndex, secondaryIndex int32, intersection placement3d.Contact) {
	primary := &s.bodies[primaryIndex]
	secondary := &s.bodies[secondaryIndex]

	solver := s.allocateDualCollisionSolver()
	solver.Init(constraint.PairCollisionState{
		PrimaryNormal:                 intersection.TargetNormal,
		PrimaryPoint:                  intersection.EvalSourcePoint(),
		PrimaryFrictionCoefficient:    primary.frictionCoefficient,
		PrimaryRestitutionCoefficient: primary.restitutionCoefficient,

		SecondaryNormal:                 intersection.EvalSourceNormal(),
		SecondaryPoint:                  intersection.TargetPoint,
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

func (s *Scene) detectBodyPropCollision(bodyIndex, propIndex int32, intersection placement3d.Contact) {
	primary := &s.bodies[bodyIndex]
	secondary := &s.props[propIndex]

	solver := s.allocateGroundCollisionSolver()
	solver.Init(constraint.CollisionState{
		BodyNormal:                 intersection.TargetNormal,
		BodyPoint:                  intersection.EvalSourcePoint(),
		BodyFrictionCoefficient:    primary.frictionCoefficient,
		BodyRestitutionCoefficient: primary.restitutionCoefficient,

		PropFrictionCoefficient:    1.0, // TODO: Take from prop or shape material
		PropRestitutionCoefficient: 0.5, // TODO: Take from prop or shape material

		Depth: intersection.Depth,
	})

	pair := sbCollisionPair{
		BodyRef: primary.reference,
		PropRef: secondary.reference,
	}
	s.newSBCollisions[pair] = struct{}{}

	primaryBody := Body{
		scene:     s,
		reference: primary.reference,
	}
	s.sbCollisionConstraints = append(s.sbCollisionConstraints, s.CreateSingleBodyConstraint(primaryBody, solver))
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

// func (s *Scene) checkCollisionBodyWithProp(primary *bodyState, prop *propState) {
// 	s.collisionSet.Reset()
// 	collision.CheckIntersectionSetWithSet(primary.collisionSet, prop.collisionSet, s.collisionSet)
// 	for _, intersection := range s.collisionSet.Intersections() {
// 		solver := s.allocateGroundCollisionSolver()
// 		solver.Init(constraint.CollisionState{
// 			BodyNormal:                 intersection.FirstDisplaceNormal,
// 			BodyPoint:                  intersection.FirstContact,
// 			BodyFrictionCoefficient:    primary.frictionCoefficient,
// 			BodyRestitutionCoefficient: primary.restitutionCoefficient,

// 			PropFrictionCoefficient:    1.0, // TODO: Take from prop or shape material
// 			PropRestitutionCoefficient: 0.5, // TODO: Take from prop or shape material

// 			Depth: intersection.Depth,
// 		})

// 		pair := sbCollisionPair{
// 			BodyRef: primary.reference,
// 			PropRef: prop.reference,
// 		}
// 		s.newSBCollisions[pair] = struct{}{}

// 		primaryBody := Body{
// 			scene:     s,
// 			reference: primary.reference,
// 		}
// 		s.sbCollisionConstraints = append(s.sbCollisionConstraints, s.CreateSingleBodyConstraint(primaryBody, solver))
// 	}
// }

// func (s *Scene) checkCollisionTwoBodies(primary, secondary *bodyState) {
// 	s.collisionSet.Reset()
// 	collision.CheckIntersectionSetWithSet(primary.collisionSet, secondary.collisionSet, s.collisionSet)
// 	for _, intersection := range s.collisionSet.Intersections() {
// 		solver := s.allocateDualCollisionSolver()
// 		solver.Init(constraint.PairCollisionState{
// 			PrimaryNormal:                 intersection.FirstDisplaceNormal,
// 			PrimaryPoint:                  intersection.FirstContact,
// 			PrimaryFrictionCoefficient:    primary.frictionCoefficient,
// 			PrimaryRestitutionCoefficient: primary.restitutionCoefficient,

// 			SecondaryNormal:                 intersection.SecondDisplaceNormal,
// 			SecondaryPoint:                  intersection.SecondContact,
// 			SecondaryFrictionCoefficient:    secondary.frictionCoefficient,
// 			SecondaryRestitutionCoefficient: secondary.restitutionCoefficient,

// 			Depth: intersection.Depth,
// 		})

// 		pair := dbCollisionPair{
// 			PrimaryRef:   primary.reference,
// 			SecondaryRef: secondary.reference,
// 		}
// 		s.newDBCollisions[pair] = struct{}{}

// 		primaryBody := Body{
// 			scene:     s,
// 			reference: primary.reference,
// 		}
// 		secondaryBody := Body{
// 			scene:     s,
// 			reference: secondary.reference,
// 		}
// 		s.dbCollisionConstraints = append(s.dbCollisionConstraints, s.CreateDoubleBodyConstraint(primaryBody, secondaryBody, solver))
// 	}
// }

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
			s.sbCollisionSubscriptions.Each(func(callback SoloBodyCollisionCallback) {
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
			s.sbCollisionSubscriptions.Each(func(callback SoloBodyCollisionCallback) {
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
			s.dbCollisionSubscriptions.Each(func(callback PairBodyCollisionCallback) {
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
			s.dbCollisionSubscriptions.Each(func(callback PairBodyCollisionCallback) {
				callback(primary, secondary, false)
			})
		}
	}
	clear(s.oldDBCollisions)
	maps.Copy(s.oldDBCollisions, s.newDBCollisions)
	clear(s.newDBCollisions)
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

func (s *Scene) allocateGlobalAccelerator() (int32, *globalAcceleratorState) {
	var index int32
	if s.freeGlobalAcceleratorIndices.IsEmpty() {
		index = int32(len(s.globalAccelerators))
		s.globalAccelerators = append(s.globalAccelerators, globalAcceleratorState{})
	} else {
		index = s.freeGlobalAcceleratorIndices.Pop()
	}
	return index, &s.globalAccelerators[index]
}

func (s *Scene) releaseGlobalAccelerator(index int32) {
	s.freeGlobalAcceleratorIndices.Push(index)
}

func (s *Scene) eachGlobalAccelerator(cb func(index int, accelerator *globalAcceleratorState)) {
	for i := range s.globalAccelerators {
		accelerator := &s.globalAccelerators[i]
		if accelerator.isValid() {
			cb(i, accelerator)
		}
	}
}

func (s *Scene) eachEnabledGlobalAccelerator(cb func(index int, accelerator *globalAcceleratorState)) {
	for i := range s.globalAccelerators {
		accelerator := &s.globalAccelerators[i]
		if accelerator.isValid() && accelerator.isEnabled {
			cb(i, accelerator)
		}
	}
}

func (s *Scene) allocateBodyAccelerator() (int32, *bodyAcceleratorState) {
	var index int32
	if s.freeBodyAcceleratorIndices.IsEmpty() {
		index = int32(len(s.bodyAccelerators))
		s.bodyAccelerators = append(s.bodyAccelerators, bodyAcceleratorState{})
	} else {
		index = s.freeBodyAcceleratorIndices.Pop()
	}
	return index, &s.bodyAccelerators[index]
}

func (s *Scene) releaseBodyAccelerator(index int32) {
	panic("TODO")
}

func (s *Scene) allocateSoloConstraint() (int32, *soloConstraintState) {
	var index int32
	if s.freeSoloConstraintIndices.IsEmpty() {
		index = int32(len(s.soloConstraints))
		s.soloConstraints = append(s.soloConstraints, soloConstraintState{})
	} else {
		index = s.freeSoloConstraintIndices.Pop()
	}
	return index, &s.soloConstraints[index]
}

func (s *Scene) releaseSoloConstraint(index int32) {
	s.freeSoloConstraintIndices.Push(index)
}

func (s *Scene) allocatePairConstraint() (int32, *pairConstraintState) {
	var index int32
	if s.freePairConstraintIndices.IsEmpty() {
		index = int32(len(s.pairConstraints))
		s.pairConstraints = append(s.pairConstraints, pairConstraintState{})
	} else {
		index = s.freePairConstraintIndices.Pop()
	}
	return index, &s.pairConstraints[index]
}

func (s *Scene) releasePairConstraint(index int32) {
	s.freePairConstraintIndices.Push(index)
}

func (s *Scene) allocateBody() (int32, *bodyState) {
	var index int32
	if s.freeBodyIndices.IsEmpty() {
		index = int32(len(s.bodies))
		s.bodies = append(s.bodies, bodyState{})
	} else {
		index = s.freeBodyIndices.Pop()
	}
	return index, &s.bodies[index]
}

func (s *Scene) releaseBody(index int32) {
	s.freeBodyIndices.Push(index)
}

func (s *Scene) eachBody(cb func(index int, b *bodyState)) {
	for i := range s.bodies {
		if body := &s.bodies[i]; body.isValid() {
			cb(i, body)
		}
	}
}

type propRef struct {
	index uint32
}

type sbCollisionPair struct {
	BodyRef indexReference
	PropRef indexReference
}

type dbCollisionPair struct {
	PrimaryRef   indexReference
	SecondaryRef indexReference
}

var nilIndex int32 = -1
