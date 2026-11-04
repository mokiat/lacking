---
title: Overview
---

# Physics

The `game/physics` package provides an impulse-based rigid-body simulation. Dynamic bodies move under the influence of accelerators (gravity, thrust, drag, lift), are restricted by constraints (joints, limits, springs), and collide with one another and with static terrain. The simulation is advanced in fixed time steps, and each step corrects velocities through impulses and positions through nudges.

This page covers how to use the package. For the theory behind the solver, see the [Physics internals](../../../development/physics/index.md).

## Core Concepts

| Concept | Description |
|---|---|
| **Scene** | A self-contained simulation. Holds bodies, terrains, accelerators, and constraints, none of which interact with those of another scene. |
| **Body** | A dynamic rigid object with mass, moment of inertia, position, rotation, and velocity. Collision shapes can be attached to it. |
| **Terrain** | Static, immovable geometry made of triangle meshes, which bodies collide with. |
| **Accelerator** | A continuous effect, such as gravity or thrust, that accumulates acceleration on bodies. Either global (all bodies) or per-body. |
| **Constraint** | A restriction on the motion of one body (solo) or two bodies (pair), enforced through impulses and nudges. |
| **Solver** | The object that implements an accelerator or a constraint. The package ships many; custom ones implement a small interface. |
| **Medium** | The fluid (air, by default) surrounding the bodies. Aerodynamic solvers sample its velocity and density. |

All quantities use SI units: meters, kilograms, seconds, and radians.

## Setup

### Creating a Scene

A scene is created empty, with default settings:

```go
scene := physics.NewScene()
```

When working with the `game` package, every `game.Scene` includes a physics scene by default, available through `Physics()`. It is advanced automatically on each fixed update (see [Integration with the Game Scene](#integration-with-the-game-scene)).

```go
physicsScene := gameScene.Physics() // nil if created with IncludePhysics: opt.V(false)
```

A new scene has **no gravity**. Gravity is an accelerator like any other and has to be added explicitly (see [Accelerators](#accelerators)).

### Advancing the Simulation

`Update` advances the simulation by the given duration:

```go
scene.Update(16 * time.Millisecond)
```

The impulse and nudge resolution assume a stable step size, so `Update` should always be called with the same, fixed duration. The `game/timestep` package provides a `Segmenter` that turns variable frame times into fixed ticks; `game.Scene` already uses one.

Each step performs the following, in order:

1. Evaluates all enabled accelerators and turns the accumulated acceleration into velocity.
1. Resets all enabled constraints and applies their impulses, over several iterations.
1. Moves all bodies according to their velocities.
1. Applies the nudges of all enabled constraints, over several iterations.
1. Detects collisions, to be resolved during the next step.
1. Notifies collision subscribers of collisions that started or stopped.

### Scene Settings

| Setting | Default | Meaning |
|---|---|---|
| `SetTimeScale` | `1.0` | Multiplier applied to the duration passed to `Update`. `0.0` pauses the simulation; negative values are clamped to `0.0`. |
| `SetImpulseIterationCount` | `8` | Number of impulse iterations per step. More iterations satisfy multiple interacting constraints more accurately. |
| `SetNudgeIterationCount` | `4` | Number of nudge iterations per step. |
| `SetImpulseDriftAdjustmentRatio` | `0.2` | Baumgarte factor that controls how aggressively impulses correct positional drift. |
| `SetNudgeDriftAdjustmentRatio` | `0.2` | Baumgarte factor that controls how aggressively nudges correct positional drift. |
| `SetMaxLinearVelocity` | unbounded | Velocities above this magnitude are clamped every step. Also available for angular velocity and for linear and angular acceleration. |

While paused, `Update` can still be called on its regular schedule. No motion is integrated and no collision notifications are sent.

## IDs, Views, and Handles

Every entity in a scene is referred to by a small, comparable ID, such as `BodyID` or `TerrainID`. IDs hold no pointers, so they are cheap to store, for example in ECS components.

IDs are resolved through a **view** of the scene, which exposes all operations for one kind of entity:

```go
bodies := scene.Bodies()

id := bodies.Create(dprec.NewVec3(0.0, 5.0, 0.0), dprec.IdentityQuat())
bodies.SetMass(id, 10.0)
position := bodies.Position(id)
```

| View | Obtained through |
|---|---|
| `BodyView` | `scene.Bodies()` |
| `TerrainView` | `scene.Terrains()` |
| `GlobalAcceleratorView` | `scene.GlobalAccelerators()` |
| `BodyAcceleratorView` | `scene.BodyAccelerators()` |
| `SoloConstraintView` | `scene.SoloConstraints()` |
| `PairConstraintView` | `scene.PairConstraints()` |

An ID becomes invalid once its entity is deleted, even if the underlying storage is later reused for a new entity. Most view methods panic when given an invalid ID; use `IsValid` when an ID might be stale. The zero value of every ID type, also available as a `Nil...` variable (such as `physics.NilBodyID`), is never valid.

A **handle** binds an ID together with its view, for an object-oriented style:

```go
body := scene.Bodies().CreateHandle(dprec.NewVec3(0.0, 5.0, 0.0), dprec.IdentityQuat())
body.SetMass(10.0)
body.SetVelocity(dprec.NewVec3(1.0, 0.0, 0.0))
```

Handles need no allocation, but they hold a reference to the scene, which keeps it alive. Prefer storing IDs in long-lived data and creating handles on demand with `view.Handle(id)`.

Every view also provides `Each`, which visits all live entities of its kind in unspecified order.

## Bodies

### Creating Bodies

A body is created at a world position and rotation. It starts with a mass of `1.0`, an identity moment of inertia, zero velocity, and no collision shapes:

```go
body := scene.Bodies().CreateHandle(
    dprec.NewVec3(0.0, 2.0, 0.0),
    dprec.IdentityQuat(),
)
```

### Mass and Moment of Inertia

```go
const mass = 20.0

body.SetMass(mass)
body.SetMomentOfInertia(physics.SolidBoxMomentOfInertia(mass, 1.0, 0.5, 2.0))
```

The mass must be positive. A mass of `math.Inf(1)` makes a body immovable by forces and impulses, though it still moves at whatever velocity it is given. The moment of inertia is expressed in the body's local space and must be invertible.

The package provides helpers for common shapes and for combining parts:

| Function | Result |
|---|---|
| `SolidSphereMomentOfInertia(mass, radius)` | Solid sphere. |
| `HollowSphereMomentOfInertia(mass, radius)` | Thin spherical shell. |
| `SolidBoxMomentOfInertia(mass, width, height, length)` | Solid box; sizes are full extents along X, Y, and Z. |
| `HollowBoxMomentOfInertia(mass, width, height, length)` | Thin box shell. |
| `DiagonalMomentOfInertia(xx, yy, zz)` | Given principal moments along the local axes. |
| `SymmetricMomentOfInertia(value)` | The same principal moment along every axis. |
| `RotatedMomentOfInertia(tensor, rotation)` | A tensor as observed from a rotated frame, for parts mounted at an angle. |
| `OffsetMomentOfInertia(tensor, mass, offset)` | A tensor around a point away from the center of mass (parallel axis theorem). |
| `MomentOfInertiaSum(a, b)` / `MomentOfInertiaMultiSum(...)` | The sum of tensors expressed in the same frame. |

For example, the inertia of an airplane can be built from a fuselage and two offset engines:

```go
inertia := physics.MomentOfInertiaMultiSum(
    physics.SolidBoxMomentOfInertia(fuselageMass, 1.5, 1.5, 10.0),
    physics.OffsetMomentOfInertia(physics.SolidSphereMomentOfInertia(engineMass, 0.5), engineMass, dprec.NewVec3(-3.0, 0.0, 0.0)),
    physics.OffsetMomentOfInertia(physics.SolidSphereMomentOfInertia(engineMass, 0.5), engineMass, dprec.NewVec3(3.0, 0.0, 0.0)),
)
```

### Motion State

| Method | Meaning |
|---|---|
| `Position` / `SetPosition` | World position of the center of mass. |
| `Rotation` / `SetRotation` | World orientation, as a unit quaternion. |
| `Velocity` / `SetVelocity` | World linear velocity of the center of mass, in m/s. |
| `AngularVelocity` / `SetAngularVelocity` | World angular velocity, in radians per second around each axis. |

Setting the position or rotation teleports the body. Its collision shapes move along immediately, so segment queries observe the change right away.

### Collision Shapes

Spheres and boxes can be attached to a body. They are specified in the body's local space and move along with it:

```go
body.AttachCollisionSphere(physics.CollisionSphere{
    Shape:                  shape3d.NewSphere(dprec.NewVec3(0.0, 0.5, 0.0), 0.5),
    FrictionCoefficient:    0.8,
    RestitutionCoefficient: 0.2,
})

shapeID := body.AttachCollisionBox(physics.CollisionBox{
    Shape:               shape3d.NewBox(dprec.ZeroVec3(), shape3d.IdentityRotation(), dprec.NewVec3(1.0, 0.25, 2.0)),
    FrictionCoefficient: 0.5,
})

body.DetachCollisionShape(shapeID)
```

The box is given by its half-extents. A body may have any number of shapes; a body without shapes is still simulated but never collides. See [Spatial](../../core/spatial/index.md) for the shapes themselves.

### Deleting Bodies

```go
body.Delete()
```

Deleting a body also deletes its collision shapes, the body accelerators that act on it, and every solo and pair constraint that references it.

## Terrains

A terrain is static geometry made of triangle meshes. Unlike body shapes, its meshes are specified in world space, since a terrain has no transform of its own:

```go
terrain := scene.Terrains().CreateHandle()
terrain.AttachCollisionMesh(physics.CollisionMesh{
    Shape: shape3d.NewMesh([]shape3d.Triangle{
        shape3d.NewTriangle(
            dprec.NewVec3(-50.0, 0.0, 50.0),
            dprec.NewVec3(50.0, 0.0, 50.0),
            dprec.NewVec3(0.0, 0.0, -50.0),
        ),
    }),
    FrictionCoefficient: 0.9,
})
```

`AttachCollisionMesh` panics if the mesh is empty. When working with the `game` package, the terrains defined in a model asset are created automatically when the model is instantiated.

## Accelerators

An accelerator evaluates an `AccelerationSolver` on every step. A **global** accelerator acts on every body in the scene, while a **body** accelerator acts on a single body:

```go
// Gravity for every body.
scene.GlobalAccelerators().Create(physics.NewGravitySolver())

// A thruster mounted at the back of one body.
thruster := physics.NewAxisForceSolver(physics.AxisForceSolverConfig{
    BodyAnchorOffset: dprec.NewVec3(0.0, 0.0, -2.0),
    Axis:             dprec.BasisZVec3(),
    Magnitude:        500.0,
})
accelerator := scene.BodyAccelerators().CreateHandle(body.ID(), thruster)
```

Accelerators are enabled when created and can be toggled with `SetEnabled`. Solvers are ordinary objects, so their settings can be changed at any time:

```go
thruster.SetMagnitude(throttle * 500.0)
accelerator.SetEnabled(throttle > 0.0)
```

### Built-in Acceleration Solvers

| Solver | Effect |
|---|---|
| `GravitySolver` | Constant acceleration along a world direction, independent of mass. Defaults to 9.8 m/s^2 along negative Y. |
| `AxisForceSolver` | Constant force along a body-local axis, applied at a body-local offset. Rotates with the body, like a thruster. |
| `DirectionForceSolver` | Constant force along a world direction, applied at a body-local offset. Does not rotate with the body, like an external jet. |
| `DragSolver` | Aerodynamic pressure drag of a box-shaped volume, resisting both translation and rotation, with a separate drag coefficient per axis. |
| `AirfoilSolver` | Lift (and the drag induced by it) of a wing-like surface, including stall past a configurable angle of attack. |

Solvers that take a `Config` struct can be created with their `New...` constructor or allocated once and reconfigured through `Configure`, which allows them to be cached and reused.

### The Medium

`DragSolver` and `AirfoilSolver` act relative to the medium that surrounds the bodies. The scene samples the medium once per body per step, through a `MediumSolver`. The default is a `StaticAirSolver`, with sea-level air density (1.2 kg/m^3) and no wind:

```go
air := physics.NewStaticAirSolver().
    SetAirDensity(1.0).
    SetAirVelocity(dprec.NewVec3(5.0, 0.0, 0.0)) // constant wind

scene.SetMediumSolver(air)
```

Altitude-dependent air, water, or localized wind can be modeled with a custom `MediumSolver`:

```go
type AtmosphereSolver struct{}

func (AtmosphereSolver) Density(position dprec.Vec3) float64 {
    return 1.2 * math.Exp(-position.Y/8500.0)
}

func (AtmosphereSolver) Velocity(position dprec.Vec3) dprec.Vec3 {
    return dprec.ZeroVec3()
}
```

### Custom Acceleration Solvers

An `AccelerationSolver` receives an `AccelerationContext` with the time step, the medium at the body's location, and an `AccelerationTarget`. The target exposes the body's mass and motion state, and accumulates the acceleration the solver contributes:

```go
// SpringSolver pulls a body toward a fixed anchor.
type SpringSolver struct {
    Anchor    dprec.Vec3
    Stiffness float64
}

func (s *SpringSolver) ApplyAcceleration(ctx physics.AccelerationContext) {
    delta := dprec.Vec3Diff(s.Anchor, ctx.Target.Position())
    ctx.Target.ApplyForce(dprec.Vec3Prod(delta, s.Stiffness))
}
```

| Target method | Effect |
|---|---|
| `AddLinearAcceleration` / `AddAngularAcceleration` | Adds acceleration directly, independent of mass. |
| `ApplyForce` | Applies a force at the center of mass. |
| `ApplyTorque` | Applies a torque. |
| `ApplyOffsetForce` | Applies a force at a world-space offset from the center of mass, inducing torque. |

All vectors are in world space. The accumulated acceleration does not change the state observed through the target, so the order in which solvers are evaluated does not matter. A solver must not retain the context or the target beyond the call.

## Constraints

A constraint restricts the motion of bodies. A **solo** constraint acts on a single body and is backed by a `SoloConstraintSolver`; a **pair** constraint acts on two bodies (a primary and a secondary) and is backed by a `PairConstraintSolver`:

```go
// Pin a body to a fixed point in the world.
scene.SoloConstraints().Create(lamp.ID(), physics.NewFixedPositionSolver(physics.FixedPositionSolverConfig{
    Position: dprec.NewVec3(0.0, 10.0, 0.0),
}))

// Connect two bodies with a ball-and-socket joint.
joint := scene.PairConstraints().CreateHandle(upperArm.ID(), lowerArm.ID(), physics.NewBallJointSolver(physics.BallJointSolverConfig{
    PrimaryBodyAnchorOffset:   dprec.NewVec3(0.0, -0.5, 0.0),
    SecondaryBodyAnchorOffset: dprec.NewVec3(0.0, 0.5, 0.0),
}))
```

Constraints are enabled when created and can be toggled with `SetEnabled`. A pair constraint cannot connect a body to itself; `Create` panics in that case.

Offsets ("anchors") and axes in solver configurations are expressed in the local space of the body they belong to, relative to its center of mass.

### Built-in Constraint Solvers

Solo constraints:

| Solver | Effect |
|---|---|
| `FixedPositionSolver` | Pins the body to a world position. Kinematic: overrides the body's linear motion completely. |
| `FixedRotationSolver` | Pins the body to a world rotation. Kinematic: overrides the body's angular motion completely. |
| `FixedDistanceSolver` | Keeps an anchor point at a fixed distance from a world point, like a rigid rod to a fixed pivot. |

Pair constraints:

| Solver | Effect |
|---|---|
| `BallJointSolver` | Keeps an anchor on each body coincident, leaving rotation free. |
| `DistanceSolver` | Keeps an anchor on each body at a fixed distance apart, like a rigid rod. |
| `AxisDisplacementSolver` | Keeps the anchors at a fixed signed distance along an axis of the primary body. |
| `AxisRangeSolver` | Keeps that signed distance within a range, acting only at the limits. |
| `AxisAngleRangeSolver` | Keeps the angle between two axes, measured about a rotation axis, within a range. Adds end stops to a hinge. |
| `MatchAxesSolver` | Makes an axis of each body parallel. Combined with a `BallJointSolver`, it forms a hinge. |
| `MatchRotationsSolver` | Drives both bodies toward the same orientation, in proportion to their inertia. |
| `CoiloverSolver` | A soft spring-damper between two anchors, tuned by frequency and damping ratio. Suited to vehicle suspension. |
| `CopyPositionSolver` | Makes the primary body follow the secondary body's position. Kinematic. |
| `CopyRotationSolver` | Makes the primary body follow the secondary body's rotation. Kinematic. |
| `CopyAxisSolver` | Makes an axis of the primary body point in the same direction as an axis of the secondary body. Kinematic. |

Kinematic solvers overwrite the motion of the affected body outright and ignore its mass, so forces and other constraints have no effect on the motion they control. The remaining solvers apply corrections in proportion to the mass and inertia of each body, so both bodies react to one another.

### Combining Constraints

`CompositeSoloConstraintSolver` and `CompositePairConstraintSolver` group several solvers into a single constraint, so they share one ID and one enable toggle. For example, a hinge with end stops:

```go
hinge := physics.NewCompositePairConstraintSolver(
    physics.NewBallJointSolver(physics.BallJointSolverConfig{
        PrimaryBodyAnchorOffset:   dprec.NewVec3(0.5, 0.0, 0.0),
        SecondaryBodyAnchorOffset: dprec.NewVec3(-0.5, 0.0, 0.0),
    }),
    physics.NewMatchAxesSolver(physics.MatchAxesSolverConfig{
        PrimaryBodyAxis:   dprec.BasisYVec3(),
        SecondaryBodyAxis: dprec.BasisYVec3(),
    }),
    physics.NewAxisAngleRangeSolver(physics.AxisAngleRangeSolverConfig{
        PrimaryBodyAxis:   dprec.BasisXVec3(),
        SecondaryBodyAxis: dprec.BasisXVec3(),
        RotationAxis:      dprec.BasisYVec3(),
        MinAngle:          dprec.Degrees(-90.0),
        MaxAngle:          dprec.Degrees(0.0),
    }),
)
door := scene.PairConstraints().CreateHandle(frame.ID(), panel.ID(), hinge)
```

The combined solvers are driven in the order they are given.

### Custom Constraint Solvers

A constraint solver has three methods, which the scene calls on every step:

| Method | When | Purpose |
|---|---|---|
| `Reset` | Once, before the impulse iterations. | Compute anything that depends on the bodies' positions and rotations, such as Jacobians. |
| `ApplyImpulses` | Once per impulse iteration. | Correct velocities. Positions do not change between iterations, so the state computed in `Reset` stays valid. |
| `ApplyNudges` | Once per nudge iteration. | Correct positions. Positions may have changed since the previous call, so position-dependent state must be recomputed at the start of each call. |

Most constraints are expressed through a `Jacobian`, which describes how a body's linear and angular velocity affect a scalar constraint. The `SoloConstraintContext` and `PairConstraintContext` passed to the solver compute the corrective impulses and nudges for a Jacobian and a positional error (drift). The following solver keeps a body above a horizontal floor:

```go
type FloorSolver struct {
    Height float64

    jacobian physics.Jacobian
    drift    float64
}

func (s *FloorSolver) Reset(ctx physics.SoloConstraintContext) {
    s.recompute(ctx)
}

func (s *FloorSolver) ApplyImpulses(ctx physics.SoloConstraintContext) {
    if s.drift <= 0.0 {
        return // above the floor
    }
    bounce, baumgarte := ctx.ImpulseLambdaComponents(s.jacobian, s.drift, 0.0)
    if bounce < 0.0 {
        return // already moving away from the floor
    }
    ctx.Target.ApplyImpulse(s.jacobian.Impulse(bounce + baumgarte))
}

func (s *FloorSolver) ApplyNudges(ctx physics.SoloConstraintContext) {
    s.recompute(ctx)
    if s.drift > 0.0 {
        ctx.Target.ApplyNudge(ctx.NudgeSolution(s.jacobian, s.drift))
    }
}

func (s *FloorSolver) recompute(ctx physics.SoloConstraintContext) {
    s.jacobian = physics.Jacobian{LinearSlope: dprec.BasisYVec3()}
    s.drift = s.Height - ctx.Target.Position().Y
}
```

For a pair solver, the two Jacobians must describe the same scalar constraint from each body's point of view (typically with negated slopes), since the same correction magnitude is applied to both bodies. The [effective mass](../../../development/physics/theory/effective-mass.md) and [impulse derivation](../../../development/physics/derivations/impulse-derivation.md) pages describe the underlying math.

## Collisions

Collisions between bodies, and between bodies and terrains, are detected by the scene and resolved through constraints that it manages internally. Contacts detected at the end of one step are resolved during the next one.

### Friction and Restitution

When two shapes touch, their friction coefficients are combined through their geometric mean (`CombinedFrictionCoefficient`) and their restitution coefficients by taking the larger of the two (`CombinedRestitutionCoefficient`). Restitution is suppressed at low impact speeds, so resting objects do not jitter.

### Filtering

Which shapes may collide is controlled through the `Filtering` field of each collision shape:

| Field | Meaning |
|---|---|
| `SourceMask` | The layers the shape occupies. Defaults to the first layer only. |
| `TargetMask` | The layers the shape can collide with. Defaults to the first layer only. |
| `RejectGroup` | Shapes that share a non-zero reject group never collide with each other. |

Two shapes are tested against each other if either one's target mask overlaps the other's source mask, and they are not in the same reject group.

Reject groups are useful for objects whose parts are meant to overlap, such as the chassis and wheels of a vehicle. `NextCollisionRejectGroup` returns a group that is unique within the scene:

```go
group := scene.NextCollisionRejectGroup()

filtering := placement3d.FilterInfo{RejectGroup: group}
chassis.AttachCollisionBox(physics.CollisionBox{Shape: chassisBox, Filtering: filtering})
wheel.AttachCollisionSphere(physics.CollisionSphere{Shape: wheelSphere, Filtering: filtering})
```

### Collision Notifications

Subscribers are notified when a collision starts and when it stops:

```go
pairSubscription := scene.SubscribePairCollision(func(first, second physics.BodyID, active bool) {
    if active {
        playImpactSound(first, second)
    }
})

soloSubscription := scene.SubscribeSoloCollision(func(body physics.BodyID, terrain physics.TerrainID, active bool) {
    // A body started (active) or stopped touching a terrain.
})

// Later.
pairSubscription.Delete()
soloSubscription.Delete()
```

Notifications are sent from within `Update`, after collision detection. A collision is tracked per pair of bodies (or per body and terrain), not per shape, so a body touching another one with several shapes produces a single notification.

### Segment Queries

The scene can be queried for intersections with a line segment, for example for line-of-sight checks, picking, or wheel raycasts:

```go
segment := shape3d.NewSegment(
    dprec.NewVec3(0.0, 10.0, 0.0),
    dprec.NewVec3(0.0, -10.0, 0.0),
)

if contact, ok := scene.CheckSegmentTerrainIntersection(segment, physics.Filter{}); ok {
    groundPoint := contact.TargetPoint
    groundNormal := contact.TargetNormal
}
```

`CheckSegmentBodyIntersection` and `CheckSegmentTerrainIntersection` return the hit closest to the start of the segment. Their `Collect...` counterparts report every hit to a callback, which can be one of the provided collectors:

| Collector | Retains |
|---|---|
| `DeepestBodyContact` / `DeepestTerrainContact` | The hit closest to the start of the segment. |
| `ShallowestBodyContact` / `ShallowestTerrainContact` | The hit farthest from the start of the segment. |
| `LastBodyContact` / `LastTerrainContact` | The hit reported last. |
| `BodyContactList` / `TerrainContactList` | All hits, in the order reported. |

```go
var hits physics.BodyContactList
scene.CollectSegmentBodyIntersections(segment, physics.Filter{}, hits.AddContact)
for _, hit := range hits.Contacts() {
    fmt.Println(hit.TargetBodyID, hit.TargetPoint)
}
```

For segment queries, the `Depth` of a contact is the fraction of the segment that lies beyond the hit, rather than a distance. The `Filter` restricts the query to shapes in particular layers (`Mask`) or excludes a reject group (`RejectGroup`), so that, for example, a vehicle's own shapes can be ignored. The zero `Filter` considers every shape.

## Integration with the Game Scene

A `game.Scene` advances its physics scene once per fixed update, before the fixed-update callbacks run. To make a hierarchy node follow a body, bind the body to the node through the scene's body binding set:

```go
body := gameScene.Physics().Bodies().CreateHandle(position, rotation)
body.AttachCollisionSphere(physics.CollisionSphere{
    Shape: shape3d.NewSphere(dprec.ZeroVec3(), 0.5),
})

node := gameScene.Hierarchy().Nodes().CreateHandle()
gameScene.BodyBindingSet().Bind(node.ID(), body.ID())
```

On each fixed update, after the physics step and the fixed-update callbacks, the bound node's absolute transformation is set from the body's position and rotation, and the node's interpolation takes care of smooth rendering between fixed steps (see [Hierarchy - Bindings](../hierarchy/index.md#bindings)). Deleting the node also deletes the bound body.

Model assets can define terrains, which are created in the physics scene when the model is instantiated.

## Limitations

- **No continuous collision detection.** Collisions are detected at discrete steps, so small, fast-moving bodies can pass through thin geometry.
- **Limited body shapes.** Bodies support spheres and boxes; meshes are available only for terrains.
- **Static terrains.** Terrain meshes are fixed in world space and cannot be moved after they are attached.
- **No sleeping.** Every body is simulated on every step, even when at rest.
