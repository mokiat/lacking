---
title: Overview
---

# ECS

The `game/ecs` package provides an Entity-Component System (ECS) framework for managing game-world objects and their data. It is archetype-based, meaning entities that share the same set of component types are stored together for efficient iteration.

## Core Concepts

| Concept | Description |
|---|---|
| **Scope** | Registry of component types. Shared across scenes that use the same component vocabulary. |
| **Scene** | Central container for entities and their components. |
| **ID** | Versioned handle to a live entity. Becomes stale automatically when the entity is deleted. |
| **Component** | Plain Go struct value attached to an entity. |
| **ComponentType** | Typed descriptor for a component, obtained at registration time. |
| **Condition** | Predicate over an entity's component set, used for queries and subscriptions. |

## Setup

### Registering Component Types

Component types are registered in a `Scope` before any `Scene` is created from it. Registration is typically done with package-level variables so that `ComponentType` descriptors are accessible throughout the codebase.

```go
var scope = ecs.NewScope()

var (
    PositionType = ecs.Type[Position](scope)
    VelocityType = ecs.Type[Velocity](scope)
    HealthType   = ecs.Type[Health](scope)
)

type Position struct {
    X, Y, Z float32
}

type Velocity struct {
    X, Y, Z float32
}

type Health struct {
    Current, Max int
}
```

A scope is locked once it is passed to `NewScene`. Attempting to register additional types after that point panics. A single scope can back multiple scenes, but each scene maintains its own independent entity table.

> **Limit:** A scope supports at most 256 component types.

### Creating a Scene

```go
scene := ecs.NewScene(scope)
defer scene.Delete()
```

Call `scene.Delete()` when the scene is no longer needed to release all resources.

## Entities

### Creating Entities

`CreateEntity` allocates a new entity and returns its `ID`. Pass a callback to add initial components atomically:

```go
id := scene.CreateEntity(func(op *ecs.EditOperation) {
    ecs.AddComponent(op, PositionType, Position{X: 1, Y: 0, Z: 0})
    ecs.AddComponent(op, VelocityType, Velocity{X: 0, Y: 0, Z: 5})
})
```

Pass `nil` to create an entity with no components:

```go
id := scene.CreateEntity(nil)
```

### Deleting Entities

```go
scene.DeleteEntity(id)
```

After deletion the `ID` becomes stale and should not be used. Few methods in the library (e.g. `HasEntity`) accept a stale ID and won't panic.

### Checking Existence and Component Membership

```go
alive := scene.HasEntity(id)

isPhysical := scene.CheckEntity(id, ecs.Conditions(
    ecs.HasComponent(PositionType),
    ecs.HasComponent(VelocityType),
))
```

## Reading Component Data

### Reading a Single Entity

`ReadEntity` calls the provided function with a `ReadOperation` scoped to that entity. The operation is valid only for the duration of the call.

```go
scene.ReadEntity(id, func(op *ecs.ReadOperation) {
    pos := ecs.GetComponent(op, PositionType) // returns *Position or nil
    if pos != nil {
        fmt.Println(pos.X, pos.Y, pos.Z)
    }
})
```

`GetComponent` returns a pointer to the component value, or `nil` if the entity does not have that component. `InjectComponent` is a convenience wrapper that writes the pointer into a variable:

```go
scene.ReadEntity(id, func(op *ecs.ReadOperation) {
    var pos *Position
    ecs.InjectComponent(op, PositionType, &pos)
    if pos != nil {
        // use pos
    }
})
```

### Querying Multiple Entities

`QueryEntities` iterates every entity that satisfies a condition. Return `false` from the callback to stop early.

```go
movingEntities := ecs.Conditions(
    ecs.HasComponent(PositionType),
    ecs.HasComponent(VelocityType),
)

scene.QueryEntities(movingEntities, func(id ecs.ID, op *ecs.ReadOperation) bool {
    pos := ecs.GetComponent(op, PositionType)
    vel := ecs.GetComponent(op, VelocityType)
    fmt.Printf("%v: pos=%v vel=%v\n", id, pos, vel)
    return true // continue
})
```

`QueryEntitiesIter` provides the same traversal as a Go range iterator:

```go
for id, op := range scene.QueryEntitiesIter(movingEntities) {
    pos := ecs.GetComponent(op, PositionType)
    _ = pos
}
```

## Editing Component Data

### Editing a Single Entity

`EditEntity` calls the provided function with an `EditOperation` for the entity. Three operations are available:

| Function | Effect |
|---|---|
| `AddComponent` | Adds a new component. Panics if the entity already has one of that type. |
| `RemoveComponent` | Removes an existing component. Panics if the entity does not have one. |
| `ReplaceComponent` | Updates the value of an existing component without changing the entity's component set. |

```go
scene.EditEntity(id, func(op *ecs.EditOperation) {
    ecs.AddComponent(op, HealthType, Health{Current: 100, Max: 100})
})

scene.EditEntity(id, func(op *ecs.EditOperation) {
    ecs.ReplaceComponent(op, VelocityType, Velocity{X: 0, Y: 10, Z: 0})
})

scene.EditEntity(id, func(op *ecs.EditOperation) {
    ecs.RemoveComponent(op, VelocityType)
})
```

Multiple operations can be staged in a single `EditEntity` call:

```go
scene.EditEntity(id, func(op *ecs.EditOperation) {
    ecs.RemoveComponent(op, VelocityType)
    ecs.AddComponent(op, HealthType, Health{Current: 50, Max: 100})
})
```

> `ReplaceComponent` is more efficient than a remove-then-add sequence when only the value needs to change, because it does not move the entity to a different archetype.

## Conditions

Conditions are predicates over an entity's component set. They are used for queries, subscriptions, and `CheckEntity`.

```go
// Entity must have Position.
ecs.HasComponent(PositionType)

// Entity must not have Velocity.
ecs.LacksComponent(VelocityType)

// Entity must have Position and Health, but not Velocity.
ecs.Conditions(
    ecs.HasComponent(PositionType),
    ecs.HasComponent(HealthType),
    ecs.LacksComponent(VelocityType),
)
```

`Conditions` panics if the combined condition is contradictory (e.g., `HasComponent` and `LacksComponent` for the same type).

### Exclusive Conditions

`Exclusive()` derives a condition that additionally requires the entity to have *no other components* beyond those already required. It is useful for targeting a very specific archetype:

```go
// Entity must have exactly Position and Velocity, and nothing else.
exact := ecs.Conditions(
    ecs.HasComponent(PositionType),
    ecs.HasComponent(VelocityType),
).Exclusive()
```

## Subscriptions

Subscriptions fire a callback whenever an entity transitions into or out of satisfying a condition. This is useful for initialising or tearing down subsystem resources (e.g., physics bodies, render objects) in response to component changes.

```go
// Called when an entity gains both Position and Velocity.
sub := scene.SubscribeEnter(
    ecs.Conditions(
        ecs.HasComponent(PositionType),
        ecs.HasComponent(VelocityType),
    ),
    func(id ecs.ID) {
        fmt.Println("entity became dynamic:", id)
    },
)

// Called when an entity no longer satisfies the condition.
scene.SubscribeExit(
    ecs.Conditions(
        ecs.HasComponent(PositionType),
        ecs.HasComponent(VelocityType),
    ),
    func(id ecs.ID) {
        fmt.Println("entity left dynamic group:", id)
    },
)

// Cancel a subscription when it is no longer needed.
sub.Delete()
```

Callbacks are dispatched after structural changes are committed, not inline during the mutation. They fire in the order the subscriptions were registered; there is no priority mechanism.

## Deferred Mutations During Queries

Structural changes — `CreateEntity`, `DeleteEntity`, and `EditEntity` calls that add or remove components — are safe to make during a query. They are buffered and applied once iteration completes.

```go
toDelete := make([]ecs.ID, 0)

scene.QueryEntities(ecs.HasComponent(HealthType), func(id ecs.ID, op *ecs.ReadOperation) bool {
    h := ecs.GetComponent(op, HealthType)
    if h.Current <= 0 {
        toDelete = append(toDelete, id)
    }
    return true
})

for _, id := range toDelete {
    scene.DeleteEntity(id)
}
```

Alternatively, `DeleteEntity` (and `EditEntity`) may be called directly inside the query callback — the deletion will be buffered and executed after the query finishes:

```go
scene.QueryEntities(ecs.HasComponent(HealthType), func(id ecs.ID, op *ecs.ReadOperation) bool {
    h := ecs.GetComponent(op, HealthType)
    if h.Current <= 0 {
        scene.DeleteEntity(id) // safe; deferred until query completes
    }
    return true
})
```

## Systems

This package does not define a system interface or a scheduler. System ordering, execution, and lifecycle management are the responsibility of the consuming application.

## Limitations

The following features are not currently provided by this ECS implementation:

- **No change detection.** There is no built-in mechanism to query only entities whose component data changed since the last frame. Systems must iterate all matching entities unconditionally.
- **No parallel queries.** The scene is not thread-safe. All operations on a scene must occur from a single goroutine.
- **No system scheduler.** Ordering and parallelism are entirely up to the application.
- **No entity relations.** Modelling parent–child or other entity-to-entity relationships requires external bookkeeping (the `game/hierarchy` package may be used for scene-node hierarchies).
- **No prefabs or entity templates.** There is no built-in way to stamp out entities from a template; construction helpers must be written by the application.
