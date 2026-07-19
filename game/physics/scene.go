package physics

import (
	"maps"
	"time"

	"github.com/mokiat/gog/ds"
	"github.com/mokiat/gog/opt"
	"github.com/mokiat/gomath/dprec"
	"github.com/mokiat/lacking/core/spatial/placement3d"
	"github.com/mokiat/lacking/core/spatial/shape3d"
	"github.com/mokiat/lacking/debug/metric"
	"github.com/mokiat/lacking/game/physics/constraint"
	"github.com/mokiat/lacking/game/physics/solver"
)

func NewScene() *Scene {
	return &Scene{
		shapeScene: placement3d.NewScene[bodyRef, struct{}, propRef](placement3d.SceneSettings{
			Size:                opt.V(16384.0),
			MaxDepth:            opt.V[uint32](12),
			InitialNodeCapacity: opt.V[uint32](1024),
			InitialItemCapacity: opt.V[uint32](1024),
		}),

		sbCollisionSubscriptions: NewSingleBodyCollisionSubscriptionSet(),
		dbCollisionSubscriptions: NewDoubleBodyCollisionSubscriptionSet(),

		timeSpeed: 1.0,

		maxLinearAcceleration:  200.0,
		maxAngularAcceleration: 200.0,
		maxLinearVelocity:      2000.0,
		maxAngularVelocity:     2000.0,

		mediumSolver: NewStaticAirSolver(),

		props: make([]propState, 0, 1024),

		freeBodyIndices:            ds.PreallocatedStack[uint32](16),
		bodies:                     make([]bodyState, 0, 64),
		bodyAccelerationTargets:    make([]AccelerationTarget, 0, 64),
		bodyConstraintPlaceholders: make([]solver.Placeholder, 0, 64),

		// bodyAccelerators   []any // TOOD
		// areaAccelerators   []any // TODO
		globalAccelerators: make([]globalAcceleratorState, 0, 64),

		freeBodyAcceleratorIndices:   ds.PreallocatedStack[uint32](16),
		freeAreaAcceleratorIndices:   ds.PreallocatedStack[uint32](16),
		freeGlobalAcceleratorIndices: ds.PreallocatedStack[uint32](16),

		sbConstraints: make([]sbConstraintState, 0, 64),
		dbConstraints: make([]dbConstraintState, 0, 64),

		freeSBConstraintIndices: ds.PreallocatedStack[uint32](16),
		freeDBConstraintIndices: ds.PreallocatedStack[uint32](16),

		collisionSet: make(placement3d.ContactList, 0, 128),

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
	shapeScene *placement3d.Scene[bodyRef, struct{}, propRef]

	sbCollisionSubscriptions *SingleBodyCollisionSubscriptionSet
	dbCollisionSubscriptions *DoubleBodyCollisionSubscriptionSet

	timeSpeed float64

	maxLinearAcceleration  float64
	maxAngularAcceleration float64
	maxLinearVelocity      float64
	maxAngularVelocity     float64

	mediumSolver MediumSolver

	props []propState

	bodies                     []bodyState
	bodyAccelerationTargets    []AccelerationTarget
	bodyConstraintPlaceholders []solver.Placeholder
	freeBodyIndices            *ds.Stack[uint32]

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

	collisionSet placement3d.ContactList

	oldSBCollisions map[sbCollisionPair]struct{}
	newSBCollisions map[sbCollisionPair]struct{}

	oldDBCollisions map[dbCollisionPair]struct{}
	newDBCollisions map[dbCollisionPair]struct{}

	freeCollisionRejectGroup uint32
	freeRevision             uint32
}

// Delete releases resources allocated by this scene. Users should not call
// any further methods on this object.
func (s *Scene) Delete() {
	s.props = nil

	s.freeBodyIndices = nil
	s.bodies = nil
	s.bodyAccelerationTargets = nil
	s.bodyConstraintPlaceholders = nil

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

// CreateGlobalAccelerator creates a new accelerator that affects the whole
// scene.
func (s *Scene) CreateGlobalAccelerator(logic AccelerationSolver) GlobalAccelerator {
	return createGlobalAccelerator(s, logic)
}

// CreateProp creates a new static Prop. A prop is an object
// that is static and rarely removed.
func (s *Scene) CreateProp(info PropInfo) {
	// TODO: createProp(s, info)

	// objectID := s.shapeScene.CreateObject(placement3d.ObjectInfo[internalRef]{
	// 	Position: info.Position,
	// 	Rotation: info.Rotation,
	// 	UserData: internalRef{
	// 		index:  propIndex,
	// 		isProp: true,
	// 	},
	// })
	// for _, sphere := range info.CollisionSpheres {
	// 	s.shapeScene.AttachSphere(objectID, placement3d.SphereInfo[struct{}]{
	// 		ShapeInfo: placement3d.ShapeInfo[struct{}]{},
	// 		Sphere:    sphere,
	// 	})
	// }
	// for _, box := range info.CollisionBoxes {
	// 	s.shapeScene.AttachBox(objectID, placement3d.BoxInfo[struct{}]{
	// 		ShapeInfo: placement3d.ShapeInfo[struct{}]{},
	// 		Box:       box,
	// 	})
	// }
	for _, mesh := range info.CollisionMeshes {
		propIndex := uint32(len(s.props))

		meshID := s.shapeScene.CreateMesh(placement3d.MeshInfo[propRef]{
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

// Update runs a single physics iteration. This method should be called with
// fixed elapsed times, otherwise the physics may break.
func (s *Scene) Update(elapsedTime time.Duration) {
	elapsedSeconds := elapsedTime.Seconds()
	s.runSimulation(elapsedSeconds * s.timeSpeed)
	s.notifySingleBodyCollisions()
	s.notifyDoubleBodyCollisions()
}

func (s *Scene) Each(cb func(b Body)) {
	s.eachBodyState(func(_ int, b *bodyState) {
		cb(Body{
			scene:     s,
			reference: b.reference,
		})
	})
}

func (s *Scene) CheckSegmentIntersection(segment shape3d.Segment, mask uint32) (Body, bool) {
	intersection, ok := s.shapeScene.CheckSegmentIntersection(segment, placement3d.Filter{
		Mask: opt.V(mask),
	})
	if !ok {
		return Body{}, false
	}
	if intersection.TargetShapeID == placement3d.InvalidShapeID {
		// A prop.
		return Body{}, false // FIXME: This should handle props as well.
	}
	objectID := s.shapeScene.GetShapeObject(intersection.TargetShapeID)
	ref := s.shapeScene.GetObjectUserData(objectID)
	return Body{
		scene:     s,
		reference: s.bodies[ref.index].reference,
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
	// TODO: body -> acceleration targets -> impulse targets -> positioning targets -> body -> check for collisions (maybe reposition  to first)

	if elapsedSeconds > 0.0001 {
		s.applyAcceleration(elapsedSeconds)
		s.applyImpulses(elapsedSeconds)
		s.applyMotion(elapsedSeconds)
		s.applyNudges(elapsedSeconds)
		s.detectCollisions()
	}
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
		s.bodyAccelerationTargets[index] = newAccelerationTarget(
			1.0/body.mass,
			dprec.InverseMat3(
				RotatedMomentOfInertia(body.momentOfInertia, body.rotation),
			),
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
	s.eachBodyState(func(index int, _ *bodyState) {
		target := &s.bodyAccelerationTargets[index]
		position := target.Position()
		for _, accelerator := range s.globalAccelerators {
			if !accelerator.reference.IsValid() || !accelerator.enabled {
				continue
			}
			ctx := AccelerationContext{
				MediumVelocity: s.mediumSolver.Velocity(position),
				MediumDensity:  s.mediumSolver.Density(position),
			}
			accelerator.logic.ApplyAcceleration(ctx, target)
		}
	})
}

func (s *Scene) applyAerodynamicAccelerations() {
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
			// TODO: Take shape velocity into account. This also means that wings should be
			// split into two, to benefit from that.

			aerodynamicShape = aerodynamicShape.Transformed(bodyTransform)
			relativeSpeed := dprec.QuatVec3Rotation(dprec.InverseQuat(aerodynamicShape.Rotation()), deltaVelocity)

			force := aerodynamicShape.solver.Force(relativeSpeed, mediumDensity)
			absoluteForce := dprec.QuatVec3Rotation(aerodynamicShape.Rotation(), force)

			offset := dprec.Vec3Diff(aerodynamicShape.Position(), bodyTransform.Position())
			target.ApplyOffsetForce(offset, absoluteForce)
			// target.ApplyOffsetForce(absoluteForce, aerodynamicShape.Position())
		}
	})
}

func (s *Scene) applyAccelerationTargets(elapsedSeconds float64) {
	s.eachBodyState(func(index int, body *bodyState) {
		target := s.bodyAccelerationTargets[index]

		linearAcceleration := target.LinearAcceleration()
		if linearAcceleration.Length() > s.maxLinearAcceleration {
			linearAcceleration = dprec.ResizedVec3(linearAcceleration, s.maxLinearAcceleration)
		}
		body.AddVelocity(dprec.Vec3Prod(linearAcceleration, elapsedSeconds))

		angularAcceleration := target.AngularAcceleration()
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

		s.shapeScene.SetObjectTransform(body.objectID, shape3d.Transform{
			Translation: body.position,
			Rotation:    shape3d.RotationFromQuat(body.rotation),
		})
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
	s.shapeScene.CollectIntersections(s.collisionSet.AddContact)
	for _, intersection := range s.collisionSet.Contacts() {
		srcBodyObject := s.shapeScene.GetShapeObject(intersection.SourceShapeID)
		srcBodyRef := s.shapeScene.GetObjectUserData(srcBodyObject)

		if intersection.TargetMeshID == placement3d.InvalidMeshID {
			tgtBodyObject := s.shapeScene.GetShapeObject(intersection.TargetShapeID)
			tgtBodyRef := s.shapeScene.GetObjectUserData(tgtBodyObject)
			s.detectBodyBodyCollision(srcBodyRef.index, tgtBodyRef.index, intersection)
		} else {
			tgtPropMesh := s.shapeScene.GetMeshUserData(intersection.TargetMeshID)
			s.detectBodyPropCollision(srcBodyRef.index, tgtPropMesh.index, intersection)
		}
	}
}

func (s *Scene) detectBodyBodyCollision(primaryIndex, secondaryIndex uint32, intersection placement3d.Contact) {
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

func (s *Scene) detectBodyPropCollision(bodyIndex, propIndex uint32, intersection placement3d.Contact) {
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

func (s *Scene) nextRevision() uint32 {
	s.freeRevision++
	return s.freeRevision
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

	s.shapeScene.SetObjectTransform(body.objectID, shape3d.Transform{
		Translation: body.position,
		Rotation:    shape3d.RotationFromQuat(body.rotation),
	})
}

type bodyRef struct {
	index uint32
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
