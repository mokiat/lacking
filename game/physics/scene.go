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
	"github.com/mokiat/lacking/util/observer"
)

// Scene represents a physics scene that contains
// a number of bodies that are independent on any
// bodies managed by other scene objects.
type Scene struct {
	collisionScene *placement3d.Scene[bodyData, shapeData, terrainData]

	soloCollisionSubscriptions *observer.SubscriptionSet[SoloCollisionCallback]
	pairCollisionSubscriptions *observer.SubscriptionSet[PairCollisionCallback]

	soloCollisionConstraintIDs []SoloConstraintID
	pairCollisionConstraintIDs []PairConstraintID

	soloCollisionSolvers []SoloCollisionSolver
	pairCollisionSolvers []PairCollisionSolver

	oldSoloCollisionRefs map[soloCollisionRef]struct{}
	oldPairCollisionRefs map[pairCollisionRef]struct{}

	newSoloCollisionRefs map[soloCollisionRef]struct{}
	newPairCollisionRefs map[pairCollisionRef]struct{}

	collisionContacts placement3d.ContactList

	freeCollisionRejectGroup uint32

	mediumSolver MediumSolver

	freeGlobalAcceleratorIndices *ds.Stack[int32]
	freeBodyAcceleratorIndices   *ds.Stack[int32]
	freeSoloConstraintIndices    *ds.Stack[int32]
	freePairConstraintIndices    *ds.Stack[int32]
	freeBodyIndices              *ds.Stack[int32]
	freeTerrainIndices           *ds.Stack[int32]

	globalAccelerators []globalAcceleratorState
	bodyAccelerators   []bodyAcceleratorState
	soloConstraints    []soloConstraintState
	pairConstraints    []pairConstraintState
	bodies             []bodyState
	terrains           []terrainState

	maxLinearAcceleration  float64
	maxAngularAcceleration float64
	maxLinearVelocity      float64
	maxAngularVelocity     float64

	impulseIterationCount       int
	impulseDriftAdjustmentRatio float64
	nudgeIterationCount         int
	nudgeDriftAdjustmentRatio   float64

	timeScale float64
}

func NewScene() *Scene {
	return &Scene{
		collisionScene: placement3d.NewScene[bodyData, shapeData, terrainData](placement3d.SceneSettings{
			Size:                opt.V(16384.0),
			MaxDepth:            opt.V[uint32](12),
			InitialNodeCapacity: opt.V[uint32](1024),
			InitialItemCapacity: opt.V[uint32](1024),
		}),

		soloCollisionSubscriptions: observer.NewSubscriptionSet[SoloCollisionCallback](),
		pairCollisionSubscriptions: observer.NewSubscriptionSet[PairCollisionCallback](),

		soloCollisionConstraintIDs: make([]SoloConstraintID, 0),
		pairCollisionConstraintIDs: make([]PairConstraintID, 0),

		soloCollisionSolvers: make([]SoloCollisionSolver, 0),
		pairCollisionSolvers: make([]PairCollisionSolver, 0),

		oldSoloCollisionRefs: make(map[soloCollisionRef]struct{}),
		oldPairCollisionRefs: make(map[pairCollisionRef]struct{}),

		newSoloCollisionRefs: make(map[soloCollisionRef]struct{}),
		newPairCollisionRefs: make(map[pairCollisionRef]struct{}),

		collisionContacts: make(placement3d.ContactList, 0),

		freeCollisionRejectGroup: 0,

		mediumSolver: NewStaticAirSolver(),

		freeGlobalAcceleratorIndices: ds.EmptyStack[int32](),
		freeBodyAcceleratorIndices:   ds.EmptyStack[int32](),
		freeSoloConstraintIndices:    ds.EmptyStack[int32](),
		freePairConstraintIndices:    ds.EmptyStack[int32](),
		freeBodyIndices:              ds.EmptyStack[int32](),
		freeTerrainIndices:           ds.EmptyStack[int32](),

		globalAccelerators: make([]globalAcceleratorState, 0),
		bodyAccelerators:   make([]bodyAcceleratorState, 0),
		soloConstraints:    make([]soloConstraintState, 0),
		pairConstraints:    make([]pairConstraintState, 0),
		bodies:             make([]bodyState, 0),
		terrains:           make([]terrainState, 0),

		maxLinearAcceleration:  math.MaxFloat64,
		maxAngularAcceleration: math.MaxFloat64,
		maxLinearVelocity:      math.MaxFloat64,
		maxAngularVelocity:     math.MaxFloat64,

		impulseIterationCount:       8,
		impulseDriftAdjustmentRatio: 0.2,
		nudgeIterationCount:         4,
		nudgeDriftAdjustmentRatio:   0.2,

		timeScale: 1.0,
	}
}

// SubscribeSoloCollision registers a callback that is invoked whenever
// a body starts or stops colliding with a terrain in the scene.
//
// Call [SoloCollisionSubscription.Delete] on the returned subscription
// to stop receiving notifications.
func (s *Scene) SubscribeSoloCollision(callback SoloCollisionCallback) *SoloCollisionSubscription {
	return s.soloCollisionSubscriptions.Subscribe(callback)
}

// SubscribePairCollision registers a callback that is invoked whenever
// two bodies start or stop colliding with each other.
//
// Call [PairCollisionSubscription.Delete] on the returned subscription
// to stop receiving notifications.
func (s *Scene) SubscribePairCollision(callback PairCollisionCallback) *PairCollisionSubscription {
	return s.pairCollisionSubscriptions.Subscribe(callback)
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

// Terrains returns a [TerrainView] through which the terrains of this
// scene can be created and managed.
func (s *Scene) Terrains() TerrainView {
	return TerrainView{
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

// ImpulseIterationCount returns the number of impulse resolution
// iterations performed per physics simulation step. A higher iteration
// count improves the accuracy with which constraints are jointly
// satisfied, at the cost of extra computation.
//
// Defaults to 8.
func (s *Scene) ImpulseIterationCount() int {
	return s.impulseIterationCount
}

// SetImpulseIterationCount changes the number of impulse resolution
// iterations performed per physics simulation step.
func (s *Scene) SetImpulseIterationCount(count int) {
	s.impulseIterationCount = count
}

// ImpulseDriftAdjustmentRatio returns the Baumgarte stabilization factor
// used to correct positional drift through impulses, i.e. the value
// passed as [SoloConstraintContext.ImpulseBeta] and
// [PairConstraintContext.ImpulseBeta] to constraint solvers.
//
// Defaults to 0.2.
func (s *Scene) ImpulseDriftAdjustmentRatio() float64 {
	return s.impulseDriftAdjustmentRatio
}

// SetImpulseDriftAdjustmentRatio changes the Baumgarte stabilization
// factor used to correct positional drift through impulses.
func (s *Scene) SetImpulseDriftAdjustmentRatio(ratio float64) {
	s.impulseDriftAdjustmentRatio = ratio
}

// NudgeIterationCount returns the number of nudge resolution iterations
// performed per physics simulation step. A higher iteration count
// improves the accuracy with which constraints are jointly satisfied, at
// the cost of extra computation.
//
// Defaults to 8.
func (s *Scene) NudgeIterationCount() int {
	return s.nudgeIterationCount
}

// SetNudgeIterationCount changes the number of nudge resolution
// iterations performed per physics simulation step.
func (s *Scene) SetNudgeIterationCount(count int) {
	s.nudgeIterationCount = count
}

// NudgeDriftAdjustmentRatio returns the Baumgarte stabilization factor
// used to correct positional drift through nudges, i.e. the value passed
// as [SoloConstraintContext.NudgeBeta] and [PairConstraintContext.NudgeBeta]
// to constraint solvers.
//
// Defaults to 0.2.
func (s *Scene) NudgeDriftAdjustmentRatio() float64 {
	return s.nudgeDriftAdjustmentRatio
}

// SetNudgeDriftAdjustmentRatio changes the Baumgarte stabilization factor
// used to correct positional drift through nudges.
func (s *Scene) SetNudgeDriftAdjustmentRatio(ratio float64) {
	s.nudgeDriftAdjustmentRatio = ratio
}

// TimeScale returns the multiplier applied to elapsed real time before it
// is fed into the physics simulation through [Scene.Update], where 1.0 is
// the default (real-time) rate and 0.0 pauses the simulation.
//
// The returned value is never negative.
func (s *Scene) TimeScale() float64 {
	return s.timeScale
}

// SetTimeScale changes the multiplier applied to elapsed real time before
// it is fed into the physics simulation through [Scene.Update].
//
// Negative values are clamped to 0, since the simulation does not support
// running time backwards.
func (s *Scene) SetTimeScale(scale float64) {
	s.timeScale = max(0.0, scale)
}

// Update advances the physics simulation by elapsedTime, scaled by
// [Scene.TimeScale], and notifies any collision subscribers registered
// through [Scene.SubscribeSoloCollision] and [Scene.SubscribePairCollision]
// of collisions that started or stopped as a result.
//
// elapsedTime must be a fixed, consistent duration across calls (e.g. the
// ticks produced by a fixed-interval segmenter), since the impulse and
// nudge resolution assume a stable step size; a varying elapsedTime will
// make the simulation inaccurate or unstable.
//
// A [Scene.TimeScale] of 0 effectively pauses the simulation: Update can
// still be called on a fixed schedule, but no motion is integrated.
func (s *Scene) Update(elapsedTime time.Duration) {
	elapsedSeconds := elapsedTime.Seconds()
	s.runSimulation(elapsedSeconds * s.timeScale)
	s.notifySoloCollisions()
	s.notifyPairCollisions()
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
	s.freeBodyAcceleratorIndices.Push(index)
}

func (s *Scene) eachBodyAccelerator(cb func(index int, accelerator *bodyAcceleratorState)) {
	for i := range s.bodyAccelerators {
		accelerator := &s.bodyAccelerators[i]
		if accelerator.isValid() {
			cb(i, accelerator)
		}
	}
}

func (s *Scene) eachEnabledBodyAccelerator(body *bodyState, cb func(index int, accelerator *bodyAcceleratorState)) {
	index := body.firstBodyAcceleratorIndex
	for index != nilIndex {
		accelerator := &s.bodyAccelerators[index]
		if accelerator.isValid() && accelerator.isEnabled {
			cb(int(index), accelerator)
		}
		index = accelerator.nextIndex
	}
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

func (s *Scene) eachSoloConstraint(cb func(index int, constraint *soloConstraintState)) {
	for i := range s.soloConstraints {
		constraint := &s.soloConstraints[i]
		if constraint.isValid() {
			cb(i, constraint)
		}
	}
}

func (s *Scene) eachEnabledSoloConstraint(cb func(index int, constraint *soloConstraintState)) {
	for i := range s.soloConstraints {
		constraint := &s.soloConstraints[i]
		if constraint.isValid() && constraint.isEnabled {
			cb(i, constraint)
		}
	}
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

func (s *Scene) eachPairConstraint(cb func(index int, constraint *pairConstraintState)) {
	for i := range s.pairConstraints {
		constraint := &s.pairConstraints[i]
		if constraint.isValid() {
			cb(i, constraint)
		}
	}
}

func (s *Scene) eachEnabledPairConstraint(cb func(index int, constraint *pairConstraintState)) {
	for i := range s.pairConstraints {
		constraint := &s.pairConstraints[i]
		if constraint.isValid() && constraint.isEnabled {
			cb(i, constraint)
		}
	}
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
		body := &s.bodies[i]
		if body.isValid() {
			cb(i, body)
		}
	}
}

func (s *Scene) allocateTerrain() (int32, *terrainState) {
	var index int32
	if s.freeTerrainIndices.IsEmpty() {
		index = int32(len(s.terrains))
		s.terrains = append(s.terrains, terrainState{})
	} else {
		index = s.freeTerrainIndices.Pop()
	}
	return index, &s.terrains[index]
}

func (s *Scene) releaseTerrain(index int32) {
	s.freeTerrainIndices.Push(index)
}

func (s *Scene) eachTerrain(cb func(index int, t *terrainState)) {
	for i := range s.terrains {
		terrain := &s.terrains[i]
		if terrain.isValid() {
			cb(i, terrain)
		}
	}
}

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
		s.eachEnabledBodyAccelerator(body, func(_ int, accelerator *bodyAcceleratorState) {
			accelerator.solver.ApplyAcceleration(ctx, target)
		})

		// Constrain the accumulated accelerations to the maximum allowed values.
		body.clampLinearAcceleration(s.maxLinearAcceleration)
		body.clampAngularAcceleration(s.maxAngularAcceleration)

		// Update the body's velocity based on the accumulated accelerations.
		body.addLinearVelocity(dprec.Vec3Prod(body.linearAcceleration, elapsedSeconds))
		body.addAngularVelocity(dprec.Vec3Prod(body.angularAcceleration, elapsedSeconds))
	})
}

func (s *Scene) applyImpulses(elapsedSeconds float64) {
	defer metric.BeginRegion("impulses").End()

	// Reset constraint solvers.
	s.eachEnabledPairConstraint(func(_ int, constraint *pairConstraintState) {
		primaryBody := &s.bodies[constraint.primaryBodyIndex]
		secondaryBody := &s.bodies[constraint.secondaryBodyIndex]

		ctx := PairConstraintContext{
			DeltaSeconds:    elapsedSeconds,
			ImpulseBeta:     s.impulseDriftAdjustmentRatio,
			NudgeBeta:       s.nudgeDriftAdjustmentRatio,
			PrimaryTarget:   newConstraintTarget(primaryBody),
			SecondaryTarget: newConstraintTarget(secondaryBody),
		}

		constraint.solver.Reset(ctx)
	})

	s.eachEnabledSoloConstraint(func(index int, constraint *soloConstraintState) {
		body := &s.bodies[constraint.bodyIndex]

		ctx := SoloConstraintContext{
			DeltaSeconds: elapsedSeconds,
			ImpulseBeta:  s.impulseDriftAdjustmentRatio,
			NudgeBeta:    s.nudgeDriftAdjustmentRatio,
			Target:       newConstraintTarget(body),
		}

		constraint.solver.Reset(ctx)
	})

	// Apply impulses multiple times in a row.
	for range s.impulseIterationCount {
		s.eachEnabledPairConstraint(func(index int, constraint *pairConstraintState) {
			primaryBody := &s.bodies[constraint.primaryBodyIndex]
			secondaryBody := &s.bodies[constraint.secondaryBodyIndex]

			ctx := PairConstraintContext{
				DeltaSeconds:    elapsedSeconds,
				ImpulseBeta:     s.impulseDriftAdjustmentRatio,
				NudgeBeta:       s.nudgeDriftAdjustmentRatio,
				PrimaryTarget:   newConstraintTarget(primaryBody),
				SecondaryTarget: newConstraintTarget(secondaryBody),
			}

			constraint.solver.ApplyImpulses(ctx)
		})

		s.eachEnabledSoloConstraint(func(index int, constraint *soloConstraintState) {
			body := &s.bodies[constraint.bodyIndex]

			ctx := SoloConstraintContext{
				DeltaSeconds: elapsedSeconds,
				ImpulseBeta:  s.impulseDriftAdjustmentRatio,
				NudgeBeta:    s.nudgeDriftAdjustmentRatio,
				Target:       newConstraintTarget(body),
			}

			constraint.solver.ApplyImpulses(ctx)
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
		body.recalculateInertia()
	})
}

func (s *Scene) applyNudges(elapsedSeconds float64) {
	defer metric.BeginRegion("nudges").End()

	for range s.nudgeIterationCount {
		s.eachEnabledPairConstraint(func(index int, constraint *pairConstraintState) {
			primaryBody := &s.bodies[constraint.primaryBodyIndex]
			secondaryBody := &s.bodies[constraint.secondaryBodyIndex]

			ctx := PairConstraintContext{
				DeltaSeconds:    elapsedSeconds,
				ImpulseBeta:     s.impulseDriftAdjustmentRatio,
				NudgeBeta:       s.nudgeDriftAdjustmentRatio,
				PrimaryTarget:   newConstraintTarget(primaryBody),
				SecondaryTarget: newConstraintTarget(secondaryBody),
			}

			constraint.solver.Reset(ctx)
			constraint.solver.ApplyNudges(ctx)
		})

		s.eachEnabledSoloConstraint(func(index int, constraint *soloConstraintState) {
			body := &s.bodies[constraint.bodyIndex]

			ctx := SoloConstraintContext{
				DeltaSeconds: elapsedSeconds,
				ImpulseBeta:  s.impulseDriftAdjustmentRatio,
				NudgeBeta:    s.nudgeDriftAdjustmentRatio,
				Target:       newConstraintTarget(body),
			}

			constraint.solver.Reset(ctx)
			constraint.solver.ApplyNudges(ctx)
		})

		s.eachBody(func(_ int, body *bodyState) {
			body.recalculateInertia()
		})
	}
}

func (s *Scene) applyPlacement() {
	defer metric.BeginRegion("placement").End()

	bodies := s.Bodies()
	s.eachBody(func(_ int, body *bodyState) {
		bodies.refreshPlacement(body)
	})
}

func (s *Scene) detectCollisions() {
	defer metric.BeginRegion("collision").End()

	// Purge old collision constraints and solvers.
	soloConstraints := s.SoloConstraints()
	for _, constraintID := range s.soloCollisionConstraintIDs {
		soloConstraints.Delete(constraintID)
	}

	pairConstraints := s.PairConstraints()
	for _, constraintID := range s.pairCollisionConstraintIDs {
		pairConstraints.Delete(constraintID)
	}

	s.soloCollisionConstraintIDs = s.soloCollisionConstraintIDs[:0]
	s.pairCollisionConstraintIDs = s.pairCollisionConstraintIDs[:0]

	s.soloCollisionSolvers = s.soloCollisionSolvers[:0]
	s.pairCollisionSolvers = s.pairCollisionSolvers[:0]

	// Collect new contacts.
	s.collisionContacts.Reset()
	s.collisionScene.CollectIntersections(s.collisionContacts.AddContact)

	// Handle contacts.
	for _, contact := range s.collisionContacts.Contacts() {
		srcBodyObject := s.collisionScene.GetShapeObject(contact.SourceShapeID)
		srcShapeData := s.collisionScene.GetShapeUserData(contact.SourceShapeID)
		srcBodyData := s.collisionScene.GetObjectUserData(srcBodyObject)

		if contact.TargetMeshID == placement3d.InvalidMeshID {
			tgtBodyObject := s.collisionScene.GetShapeObject(contact.TargetShapeID)
			tgtShapeData := s.collisionScene.GetShapeUserData(contact.TargetShapeID)
			tgtBodyData := s.collisionScene.GetObjectUserData(tgtBodyObject)
			s.handlePairCollision(
				bodyCollisionData{
					srcBodyData.index,
					srcShapeData.frictionCoefficient,
					srcShapeData.restitutionCoefficient,
				},
				bodyCollisionData{
					tgtBodyData.index,
					tgtShapeData.frictionCoefficient,
					tgtShapeData.restitutionCoefficient,
				},
				contact,
			)
		} else {
			tgtTerrainData := s.collisionScene.GetMeshUserData(contact.TargetMeshID)
			s.handleSoloCollision(
				bodyCollisionData{
					srcBodyData.index,
					srcShapeData.frictionCoefficient,
					srcShapeData.restitutionCoefficient,
				},
				terrainCollisionData{
					tgtTerrainData.index,
					tgtTerrainData.frictionCoefficient,
					tgtTerrainData.restitutionCoefficient,
				},
				contact,
			)
		}
	}
}

func (s *Scene) handlePairCollision(primaryData, secondaryData bodyCollisionData, contact placement3d.Contact) {
	solver := s.allocatePairCollisionSolver()
	solver.Configure(PairCollisionSolverConfig{
		PrimaryFrictionCoefficient:    primaryData.frictionCoefficient,
		PrimaryRestitutionCoefficient: primaryData.restitutionCoefficient,
		PrimaryContactNormal:          contact.EvalSourceNormal(),
		PrimaryContactPoint:           contact.EvalSourcePoint(),

		SecondaryFrictionCoefficient:    secondaryData.frictionCoefficient,
		SecondaryRestitutionCoefficient: secondaryData.restitutionCoefficient,
		SecondaryContactNormal:          contact.TargetNormal,
		SecondaryContactPoint:           contact.TargetPoint,

		ContactDepth: contact.Depth,
	})

	primaryID := s.Bodies().idFromIndex(primaryData.index)
	secondaryID := s.Bodies().idFromIndex(secondaryData.index)

	constraintID := s.PairConstraints().Create(primaryID, secondaryID, solver)
	s.pairCollisionConstraintIDs = append(s.pairCollisionConstraintIDs, constraintID)

	ref := pairCollisionRef{
		primaryBodyID:   primaryID,
		secondaryBodyID: secondaryID,
	}
	s.newPairCollisionRefs[ref] = struct{}{}
}

func (s *Scene) handleSoloCollision(bodyData bodyCollisionData, terrainData terrainCollisionData, contact placement3d.Contact) {
	solver := s.allocateSoloCollisionSolver()
	solver.Configure(SoloCollisionSolverConfig{
		TerrainFrictionCoefficient:    terrainData.frictionCoefficient,
		TerrainRestitutionCoefficient: terrainData.restitutionCoefficient,
		TerrainContactNormal:          contact.TargetNormal,

		BodyFrictionCoefficient:    bodyData.frictionCoefficient,
		BodyRestitutionCoefficient: bodyData.restitutionCoefficient,
		BodyContactPoint:           contact.EvalSourcePoint(),

		ContactDepth: contact.Depth,
	})

	bodyID := s.Bodies().idFromIndex(bodyData.index)
	terrainID := s.Terrains().idFromIndex(terrainData.index)

	constraintID := s.SoloConstraints().Create(bodyID, solver)
	s.soloCollisionConstraintIDs = append(s.soloCollisionConstraintIDs, constraintID)

	ref := soloCollisionRef{
		bodyID:    bodyID,
		terrainID: terrainID,
	}
	s.newSoloCollisionRefs[ref] = struct{}{}
}

func (s *Scene) allocatePairCollisionSolver() *PairCollisionSolver {
	index := len(s.pairCollisionSolvers)
	s.pairCollisionSolvers = append(s.pairCollisionSolvers, PairCollisionSolver{})
	return &s.pairCollisionSolvers[index]
}

func (s *Scene) allocateSoloCollisionSolver() *SoloCollisionSolver {
	index := len(s.soloCollisionSolvers)
	s.soloCollisionSolvers = append(s.soloCollisionSolvers, SoloCollisionSolver{})
	return &s.soloCollisionSolvers[index]
}

func (s *Scene) notifySoloCollisions() {
	for newRef := range s.newSoloCollisionRefs {
		if _, ok := s.oldSoloCollisionRefs[newRef]; !ok {
			s.soloCollisionSubscriptions.Each(func(callback SoloCollisionCallback) {
				callback(newRef.bodyID, newRef.terrainID, true)
			})
		}
	}
	for oldRef := range s.oldSoloCollisionRefs {
		if _, ok := s.newSoloCollisionRefs[oldRef]; !ok {
			s.soloCollisionSubscriptions.Each(func(callback SoloCollisionCallback) {
				callback(oldRef.bodyID, oldRef.terrainID, false)
			})
		}
	}
	clear(s.oldSoloCollisionRefs)
	maps.Copy(s.oldSoloCollisionRefs, s.newSoloCollisionRefs)
	clear(s.newSoloCollisionRefs)
}

func (s *Scene) notifyPairCollisions() {
	for newRef := range s.newPairCollisionRefs {
		if _, ok := s.oldPairCollisionRefs[newRef]; !ok {
			s.pairCollisionSubscriptions.Each(func(callback PairCollisionCallback) {
				callback(newRef.primaryBodyID, newRef.secondaryBodyID, true)
			})
		}
	}
	for oldRef := range s.oldPairCollisionRefs {
		if _, ok := s.newPairCollisionRefs[oldRef]; !ok {
			s.pairCollisionSubscriptions.Each(func(callback PairCollisionCallback) {
				callback(oldRef.primaryBodyID, oldRef.secondaryBodyID, false)
			})
		}
	}
	clear(s.oldPairCollisionRefs)
	maps.Copy(s.oldPairCollisionRefs, s.newPairCollisionRefs)
	clear(s.newPairCollisionRefs)
}

var nilIndex int32 = -1

type bodyData struct {
	index int32
}

type shapeData struct {
	frictionCoefficient    float64
	restitutionCoefficient float64
}

type terrainData struct {
	index                  int32
	frictionCoefficient    float64
	restitutionCoefficient float64
}

type bodyCollisionData struct {
	index                  int32
	frictionCoefficient    float64
	restitutionCoefficient float64
}

type terrainCollisionData struct {
	index                  int32
	frictionCoefficient    float64
	restitutionCoefficient float64
}

type soloCollisionRef struct {
	bodyID    BodyID
	terrainID TerrainID
}

type pairCollisionRef struct {
	primaryBodyID   BodyID
	secondaryBodyID BodyID
}

/////// OLD BELOW ------------ (TODO: DELETE COMMENT)

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
