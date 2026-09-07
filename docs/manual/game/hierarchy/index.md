---
title: Overview
---

# Hierarchy

The `game/hierarchy` package provides a scene graph of 3D nodes arranged in parent-child relationships. Each node carries a local transformation (translation, rotation, and scale) relative to its parent, from which its absolute (world) transformation is composed. The package is designed for smooth rendering between fixed-rate updates and for connecting nodes to external systems (graphics, physics, animation) through bindings.

## Core Concepts

| Concept | Description |
|---|---|
| **Scene** | Central container that owns all nodes and their storage, and the set of bindings registered against them. |
| **NodeView** | Lightweight, freely-copyable handle to a scene through which nodes are created, deleted, traversed, and transformed. |
| **NodeID** | Stable, versioned identifier for a single node. Becomes invalid once the node is deleted, even if its storage slot is reused. |
| **NodeHandle** | Convenience wrapper binding a `NodeID` to its `NodeView`, so a node can be operated on without passing the two separately. |
| **Binding** | Association of arbitrary values of some type `T` with nodes, driven by a solver that transfers data between the node and the value. |

## Setup

### Creating a Scene

A scene requires no configuration and is created empty:

```go
scene := hierarchy.NewScene()
```

Nodes are never manipulated directly. Instead, `scene.Nodes()` returns a `NodeView`, the entry point for all node operations:

```go
nodes := scene.Nodes()
```

A `NodeView` is a small value that can be freely copied and passed around; every view obtained from the same scene observes the same underlying state.

## Nodes

### Creating Nodes

`Create` allocates a new node and returns its `NodeID`. The node starts as a root (no parent), with no children, an empty name, and an identity transformation.

```go
id := nodes.Create()
```

`CreateHandle` is a convenience that returns a `NodeHandle` instead:

```go
handle := nodes.CreateHandle()
```

### Deleting Nodes

```go
nodes.Delete(id)
```

Deleting a node also deletes all of its descendants. Any `NodeID` referring to a deleted node becomes invalid. Deletion notifies every binding so that bound values can be released (see [Bindings](#bindings)).

### Validity

A `NodeID` is stable for the lifetime of its node but is invalidated on deletion, even if the storage slot is later reused by a different node. `NilNodeID` (the zero value) is never valid.

```go
if nodes.IsValid(id) {
    // safe to operate on id
}
```

Passing an invalid ID to most operations panics. `IsValid` is the safe way to test an ID.

## Structure and Traversal

### Parent-Child Relationships

New nodes are roots. Use `AttachChild` and `Detach` to build and rearrange the hierarchy:

```go
parent := nodes.Create()
child := nodes.Create()

// Make child a child of parent.
nodes.AttachChild(parent, child, false)

// Turn child back into a root.
nodes.Detach(child, false)
```

If the child already has a parent, `AttachChild` first detaches it. Attaching a child to its current parent is a no-op.

The `preserveWorldTransform` argument controls what happens to the node's world-space placement. When `true`, the node's local transformation is adjusted so its absolute (world) position, rotation, and scale are unchanged by the reparenting; when `false`, the local transformation is kept and the node moves along with its new parent.

### Querying Structure

```go
nodes.IsRoot(id)          // true if the node has no parent
nodes.Parent(id)          // parent ID, or NilNodeID for a root
nodes.FirstChild(id)      // first child ID, or NilNodeID
nodes.NextSibling(id)     // next sibling ID, or NilNodeID
```

The children of a node are visited by starting at `FirstChild` and following `NextSibling` until `NilNodeID`:

```go
for child := nodes.FirstChild(id); child != hierarchy.NilNodeID; child = nodes.NextSibling(child) {
    // visit child
}
```

`SubtreeContains` reports whether one node lies within the subtree rooted at another (a node contains itself):

```go
if nodes.SubtreeContains(root, candidate) {
    // candidate is root or a descendant of root
}
```

### Traversal

Several traversal helpers are provided. Each callback returns `true` to continue or `false` to stop early. Each method has an iterator counterpart (suffixed `Iter`) for use with range-over-func.

| Method | Visits |
|---|---|
| `Each` | Every valid node, in unspecified order. |
| `EachRoot` | Every valid root node, in unspecified order. |
| `Walk` | Every node, depth-first and pre-order (each node before its children). |
| `WalkSubtree` | The subtree rooted at a given node, depth-first and pre-order. |

```go
// Callback form.
nodes.Walk(func(id hierarchy.NodeID) bool {
    fmt.Println(nodes.Name(id))
    return true
})

// Iterator form.
for id := range nodes.WalkIter() {
    fmt.Println(nodes.Name(id))
}

// Restrict to one subtree.
nodes.WalkSubtree(root, func(id hierarchy.NodeID) bool {
    return true
})
```

Depth-first traversal visits each node before its children; the relative order of separate root subtrees is unspecified.

### Finding Nodes by Name

Nodes carry a name (not required to be unique) set via `SetName`. `FindNode` searches the whole scene; `FindNodeInSubtree` restricts the search to one subtree (passing `NilNodeID` as the root searches the whole scene). Names are matched exactly, and if several nodes share a name, which one is returned is unspecified.

```go
nodes.SetName(id, "player")
found := nodes.FindNode("player")

inSubtree := nodes.FindNodeInSubtree(root, "weapon")
```

## Transformations

Every node has a local translation, rotation, and scale, expressed relative to its parent. The individual components can be read and written separately or together:

```go
nodes.SetPosition(id, dprec.NewVec3(1.0, 0.0, 0.0))
nodes.SetRotation(id, dprec.RotationQuat(dprec.Degrees(90), dprec.BasisYVec3()))
nodes.SetScale(id, dprec.NewVec3(2.0, 2.0, 2.0))

pos, rot, scale := nodes.TRS(id)
nodes.SetTRS(id, pos, rot, scale)
```

The local transformation is also available as a matrix, which is composed from (and decomposed back into) the position, rotation, and scale:

```go
local := nodes.Matrix(id)
nodes.SetMatrix(id, local)
```

### Absolute (World) Transformation

A node's absolute transformation is its local transformation composed with those of all its ancestors. It is exposed as a matrix and computed lazily and cached, so reading it is cheap when nothing has changed.

```go
world := nodes.AbsoluteMatrix(id)
```

`SetAbsoluteMatrix` sets the local transformation so that the node's world matrix equals the given matrix; the local transformation is derived relative to the parent, so moving an ancestor afterwards still moves the node with it. Descendants are updated to account for the change.

```go
nodes.SetAbsoluteMatrix(id, worldMatrix)
```

`ReferenceMatrix` returns the frame the local transformation is relative to: the parent's absolute matrix for a parented node, or the identity matrix for a root or independent node. By definition, absolute matrix = reference matrix * local matrix.

### Independent Nodes

A node can be made independent via `SetIndependent`. An independent node keeps its place in the hierarchy - it still has a parent, participates in traversal, and inherits hidden state - but its local transformation is used directly as its absolute transformation, ignoring its ancestors, as if it were a root. A root node is always effectively independent.

```go
nodes.SetIndependent(id, true, false)
```

As with attachment, the final `preserveWorldTransform` argument, when `true`, adjusts the local transformation so the node's world transformation is unchanged by the switch.

## Visibility

Each node has its own hidden flag, and an effective ("absolute") visibility that also accounts for ancestors.

| Method | Meaning |
|---|---|
| `SetHidden` / `SetVisible` | Set the node's own hidden flag. |
| `IsHidden` / `IsVisible` | Read only the node's own hidden flag. |
| `IsAbsoluteHidden` / `IsAbsoluteVisible` | Effective visibility, considering ancestors. |

Hiding a node makes it and all of its descendants absolutely hidden. A descendant remains absolutely hidden until every ancestor that hides it, as well as the descendant itself, is no longer hidden.

```go
nodes.SetHidden(parent, true)
nodes.IsHidden(child)         // false - the child itself is not hidden
nodes.IsAbsoluteHidden(child) // true  - an ancestor is hidden
```

## Interpolation

To support smooth rendering between fixed-rate updates, each node also retains its previous absolute transformation.

`Scene.AdvanceStep` records the current absolute transformations of all nodes as their previous ones. Call it once per fixed-rate step, before applying that step's new transformations:

```go
scene.AdvanceStep()
// ... apply this step's movement to the nodes ...
```

During rendering, `InterpolatedAbsoluteMatrix` blends between the previous and current pose by a fraction in the range `[0, 1]`, where `0` yields the previous matrix and `1` the current one. Translation and scale are interpolated linearly and rotation spherically.

```go
world := nodes.InterpolatedAbsoluteMatrix(id, fraction)
```

### Teleporting with Snap

Moving a node by a discontinuous amount would otherwise make it appear to slide from its old position to the new one, because interpolation blends across the gap. `Snap` records the current pose of a node and all of its descendants as their previous pose, collapsing that interpolation so the node renders at its new position immediately. Snap after applying the new transformation.

```go
nodes.SetAbsoluteMatrix(id, teleportTarget)
nodes.Snap(id)
```

`Snap` acts on a single node's subtree; `Scene.AdvanceStep` performs the equivalent step for the whole scene.

## Node Handles

A `NodeHandle` bundles a `NodeID` with the `NodeView` it belongs to, so a node can be operated on without passing both around. Every `NodeView` operation has an equivalent method on `NodeHandle`.

```go
handle := nodes.CreateHandle()
handle.SetName("root")
handle.SetPosition(dprec.NewVec3(0.0, 1.0, 0.0))

child := nodes.CreateHandle()
handle.AttachChild(child, false)

for descendant := range handle.WalkIter() {
    fmt.Println(descendant.Name())
}
```

Navigation methods return handles too, so traversal chains naturally:

```go
firstGrandchild := handle.FirstChild().FirstChild()
if firstGrandchild.IsValid() {
    // ...
}
```

A handle to `NilNodeID` or to a deleted node is not valid; test it with `handle.IsValid()`.

## Bindings

Bindings connect nodes to external data and systems. A `Binding[T]` associates values of type `T` with nodes and transfers data between the node and its value according to a solver.

### Defining a Solver

A solver is any value that implements one or more of the following interfaces; the binding detects which it implements and drives only those. A solver implementing none of them yields an inert binding that stores values but transfers nothing.

| Interface | Method | Direction |
|---|---|---|
| `SourceBindingSolver[T]` | `OnSourceToNode` | Value -> node (value is the source). |
| `TargetBindingSolver[T]` | `OnTargetFromNode` | Node -> value (node is the source). |
| `InterpolationBindingSolver[T]` | `OnInterpolationFromNode` | Node's interpolated pose -> value. |
| `LifecycleBindingSolver[T]` | `OnDelete` | Invoked when a bound node is deleted, to release the value. |

For example, a solver that copies a node's interpolated world matrix into a render mesh:

```go
type meshSolver struct{}

func (meshSolver) OnInterpolationFromNode(scene *hierarchy.Scene, id hierarchy.NodeID, mesh *RenderMesh, fraction float64) {
    mesh.SetMatrix(scene.Nodes().InterpolatedAbsoluteMatrix(id, fraction))
}

func (meshSolver) OnDelete(scene *hierarchy.Scene, id hierarchy.NodeID, mesh *RenderMesh) {
    mesh.Release()
}
```

### Creating and Populating a Binding

Because the value type cannot be inferred from the solver, it must be given explicitly. A new binding registers itself with the scene and starts with a priority of `0`.

```go
binding := hierarchy.NewBinding[*RenderMesh](scene, meshSolver{})

binding.Bind(id, mesh)   // associate a value with a node
binding.Has(id)          // whether the node has a bound value
binding.Get(id)          // the bound value, or the zero value of T
binding.Unbind(id)       // remove the association silently
```

`Unbind` removes an association without notifying the solver; `OnDelete` is reserved for actual node deletion. `Delete` removes the whole binding from the scene:

```go
binding.Delete()
```

### Applying Transfers

The transfers are driven from the scene, which invokes the appropriate solver method for each bound node. Each has a whole-scene form and a single-node form:

| Scene method | Solver method driven |
|---|---|
| `ApplySourcesToNodes` | `OnSourceToNode` |
| `ApplyTargetsFromNodes` | `OnTargetFromNode` |
| `ApplyInterpolationsFromNodes(fraction)` | `OnInterpolationFromNode` |

```go
// Push external state onto nodes (e.g. physics -> hierarchy).
scene.ApplyTargetsFromNodes()

// During rendering, pull interpolated poses into render objects.
scene.ApplyInterpolationsFromNodes(fraction)
```

The single-node variants take a node ID and a `recursive` flag; when recursive, the transfer is applied to the node and all of its descendants:

```go
scene.ApplySourceToNode(id, true)
scene.ApplyTargetFromNode(id, false)
scene.ApplyInterpolationFromNode(id, fraction, true)
```

### Priority

A scene can host multiple bindings. They are always processed in ascending priority order, which lets dependent transfers run after the ones they rely on.

```go
binding.SetPriority(10)
```

## Limitations

- **Not thread-safe.** A scene and its nodes must be accessed from a single goroutine.
- **No structural change detection.** There is no built-in way to query only the nodes whose transformation changed since the last step; absolute matrices are cached and recomputed lazily, but traversal visits all matching nodes.
- **No built-in systems.** The package supplies the scene graph and the binding mechanism; scheduling of transfers, per-frame update order, and integration with concrete graphics/physics/animation systems are the responsibility of the consuming application.
</content>
</invoke>
