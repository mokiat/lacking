// Package physics provides a rigid-body physics simulation.
//
// # Scene
//
// A [Scene] is a self-contained simulation. It holds dynamic bodies, static
// terrains and everything that acts on them, and it is advanced through
// [Scene.Update], which should be called with a fixed time step. Bodies in
// one Scene never interact with bodies in another.
//
// All quantities use SI units: meters, kilograms, seconds and radians.
//
// # IDs, views and handles
//
// Every entity in a Scene is referred to by a small, comparable ID value,
// such as [BodyID] or [TerrainID]. An ID holds no pointers, so it is cheap
// to store, and it becomes invalid once its entity is deleted, even if the
// underlying storage is later reused for a new entity. The zero value of
// every ID type (e.g. [NilBodyID]) never refers to a valid entity.
//
// An ID is resolved through a view of the owning Scene, such as
// [Scene.Bodies] or [Scene.Terrains], which exposes all operations for that
// kind of entity. Most view methods panic when given an invalid ID; use the
// view's IsValid method if an ID might be stale.
//
// A handle, such as [BodyHandle], binds an ID together with its view, for
// callers that prefer an object-oriented API. Handles are convenient but
// keep the Scene reachable, so prefer storing IDs in long-lived data.
//
// # Bodies and terrains
//
// A body (see [BodyView]) is a rigid object with mass, moment of inertia,
// position, rotation and velocity, which moves under the influence of
// accelerators and constraints. Collision shapes, such as [CollisionSphere]
// and [CollisionBox], can be attached to it.
//
// A terrain (see [TerrainView]) is static, immovable geometry, made of
// [CollisionMesh] shapes, that bodies collide with.
//
// # Accelerators
//
// Accelerators apply continuous effects, such as gravity, thrust, drag or
// lift, by accumulating acceleration on a body. Each is backed by an
// [AccelerationSolver]. A global accelerator (see [GlobalAcceleratorView])
// acts on every body in the Scene, while a body accelerator (see
// [BodyAcceleratorView]) acts on a single body. The package provides
// solvers such as [GravitySolver], [AxisForceSolver], [DragSolver] and
// [AirfoilSolver].
//
// Aerodynamic solvers take the surrounding medium into account, which the
// Scene samples through its [MediumSolver] (see [Scene.SetMediumSolver]).
//
// # Constraints
//
// Constraints restrict the motion of bodies, by correcting their velocities
// through impulses and their positions through nudges. A solo constraint
// (see [SoloConstraintView]) acts on a single body and is backed by a
// [SoloConstraintSolver]. A pair constraint (see [PairConstraintView]) acts
// on two bodies and is backed by a [PairConstraintSolver]. The package
// provides solvers for common joints and limits, such as [BallJointSolver],
// [DistanceSolver], [AxisRangeSolver] and [CoiloverSolver], and custom ones
// can be built on top of [Jacobian], [SoloConstraintContext] and
// [PairConstraintContext].
//
// # Collisions
//
// Collisions between bodies, and between bodies and terrains, are detected
// by the Scene itself and resolved through constraints that it manages
// internally. Which shapes may collide is controlled through the filtering
// settings of each collision shape (see [CollisionShape.Filtering] and
// [Scene.NextCollisionRejectGroup]). Callers can be notified when
// collisions start and stop through [Scene.SubscribeSoloCollision] and
// [Scene.SubscribePairCollision].
//
// The Scene can also be queried directly for segment intersections, for
// example through [Scene.CheckSegmentBodyIntersection].
package physics
