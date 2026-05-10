// Package ecs provides an Entity-Component-System (ECS) framework for
// game development. It supports efficient storage and querying of
// entities and their associated components.
//
// # Core concepts
//
// A [Scope] holds the component type registry. Register Go structs as
// component types with [Type] and pass the resulting [ComponentType]
// descriptors to API functions. Scopes are shared across scenes that
// operate over the same component types.
//
// A [Scene] is the central container. It manages entity lifetime,
// stores components in archetype-grouped tables, and dispatches
// structural-change events to subscribers.
//
// Entities are referred to by [ID] values, which remain valid until the
// entity is deleted. Deletion increments an internal revision so that
// stale IDs are detected automatically.
//
// Component reads and writes are performed through [ReadOperation] and
// [EditOperation] values passed to callbacks provided to
// [Scene.ReadEntity], [Scene.EditEntity], [Scene.CreateEntity], and
// [Scene.QueryEntities].
//
// # Systems
//
// This package does not define a System interface. System ordering,
// scheduling, and lifecycle management are the responsibility of the
// application.
package ecs
