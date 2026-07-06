---
title: Overview
---

# Spatial

The `core/spatial` package is the engine's geometry and collision stack. It
provides the primitive shapes, the algorithms that test them for overlap, the
broad-phase indexes that make those tests scale, and the scene abstraction that
ties everything together. The physics engine is built directly on top of it,
but the package is self-contained and can be used on its own for picking,
line-of-sight tests, trigger volumes, and similar spatial problems.

Everything comes in a 2D and a 3D flavour, split into parallel sub-packages that
mirror one another almost exactly. Pick the dimensionality you need and the same
concepts apply:

| Layer | 2D | 3D | Purpose |
|---|---|---|---|
| **Shapes** | `shape2d` | `shape3d` | Primitive geometric shapes and contact types. |
| **Convex solver** | `gjk2d` | `gjk3d` | GJK/EPA overlap and contact resolution for convex shapes. |
| **Intersections** | `isec2d` | `isec3d` | Direct intersection tests and contacts between specific primitives. |
| **Broad phase** | `query2d` | `query3d` | Loose quad/octree spatial index for fast candidate lookup. |
| **Scenes** | `placement2d` | `placement3d` | Full scenes of placed objects and meshes, combining all of the above. |

The layers stack: shapes are the vocabulary; `gjk*` and `isec*` are two
different narrow-phase strategies for testing pairs of shapes; `query*` is the
broad phase that avoids testing every pair; and `placement*` is the high-level
scene that wires the broad phase and narrow phase together so you rarely have to
touch the lower layers directly.

**Note:**
Both the 2D and 3D packages work in double precision, using the `dprec`
vectors, rotations and transforms from the `gomath` math library.

## Shapes

`shape2d` and `shape3d` define the primitive shapes and the shared `Contact`
type used throughout the stack.

- **2D**: `Circle`, `Rectangle`, `Capsule`, `Triangle`, `Edge`, `Segment`, and
  `Mesh` (a collection of edges).
- **3D**: `Sphere`, `Box`, `Triangle`, `Surface`, `Segment`, and `Mesh` (a
  collection of triangles).

Shapes are plain value types constructed with `New...` helpers and carry no
behaviour beyond simple queries such as `ContainsPoint`:

```go
sphere := shape3d.NewSphere(dprec.NewVec3(0, 1, 0), 0.5)
box := shape3d.NewBox(
    dprec.ZeroVec3(),            // center
    shape3d.IdentityRotation(),  // orientation
    dprec.NewVec3(1, 1, 1),      // half-extents (half-width, -height, -length)
)

if sphere.ContainsPoint(dprec.NewVec3(0, 1.2, 0)) {
    // ...
}
```

Each shape has a `Transformed...` helper that applies a rigid-body `Transform`
(translation plus rotation), which is how shapes are moved into world space.

### Contacts

A `Contact` describes an intersection of a *source* shape with a *target*
shape. Its fields are expressed relative to the target:

| Field | Meaning |
|---|---|
| `TargetPoint` | The contact point on the target's surface. |
| `TargetNormal` | The outward surface normal at that point, pointing toward the source. It is the direction the source must move to separate. |
| `Depth` | How far the shapes overlap along the normal (always non-negative). |

Intersection routines do not return contacts directly; they push them to a
`ContactCallback`. The packages ship ready-made sinks that satisfy this callback
and accumulate contacts in common ways:

| Sink | Behaviour |
|---|---|
| `LastContact` | Keeps the most recent contact. |
| `DeepestContact` | Keeps the contact with the greatest depth. |
| `ShallowestContact` | Keeps the contact with the least depth. |
| `ContactList` | Collects every contact into a slice. |

```go
var deepest shape3d.DeepestContact
isec3d.ResolveSphereSphere(first, second, deepest.AddContact)
if contact, ok := deepest.Contact(); ok {
    // move `first` out of `second` along contact.TargetNormal by contact.Depth
}
```

## Convex Solver (GJK)

`gjk2d` and `gjk3d` implement the Gilbert-Johnson-Keerthi algorithm (with EPA
for penetration depth) for arbitrary *convex* shapes. This is the general-purpose
narrow phase: it works on any convex shape without a hand-written test for that
specific pair.

A `gjk.Shape` is a convex point cloud (a polygon in 2D, a polyhedron in 3D) with
an optional **skin radius** that inflates it outward. The skin radius lets a
single representation capture rounded shapes exactly: a sphere is one point with
a radius, a capsule is a segment with a radius, and a rounded box is a box whose
corners are inflated. Convenience constructors build these:

```go
shapeA := gjk3d.ShapeFromSphere(sphere)
shapeB := gjk3d.ShapeFromCapsule(segment, 0.25)

solver := gjk3d.NewSolver()

if solver.Intersect(shapeA, shapeB) {
    // boolean overlap test only
}

if contact, ok := solver.Resolve(shapeA, shapeB); ok {
    // full contact with normal and penetration depth
}
```

Reuse a `Solver` across calls; it holds scratch buffers to avoid per-call
allocations.

## Intersections

`isec2d` and `isec3d` provide direct, specialised intersection tests between
specific pairs of primitives. Where GJK is general, these are hand-written for
each shape combination and are typically faster for the cases
they cover (for example sphere-versus-triangle or segment-versus-box).

Functions come in two forms:

- `Check...` returns a boolean overlap answer.
- `Resolve...` reports full `Contact` values through a `ContactCallback`.

```go
if isec3d.CheckSphereBox(sphere, box) {
    // fast boolean test
}

var contacts shape3d.ContactList
isec3d.ResolveSphereMesh(sphere, mesh, contacts.AddContact)
for _, c := range contacts.Contacts() {
    // ...
}
```

### Contact conventions

Two contact conventions are used depending on the shapes involved:

- **Volume-versus-volume** resolves (such as `ResolveSphereSphere`) report a
  mutual-penetration contact: the normal is the axis of least penetration and
  `Depth` is the overlap distance along it.

- **Segment** resolves (such as `ResolveSegmentBox`) treat the segment as a
  directed probe from A to B. They report where the segment first enters the
  shape, the outward surface normal there, and a `Depth` equal to the *fraction*
  of the segment lying beyond the entry point (1 at A, 0 at B). Expressing depth
  as a fraction keeps it comparable across shapes, so `DeepestContact` selects
  the earliest entry along the ray. Segment tests are face-culled and directional:
  a segment that starts inside the shape is not treated as an intersection.

## Broad Phase (Spatial Queries)

`query2d` and `query3d` provide the broad phase: a loose **quadtree** /
**octree** that indexes items by a spatial `Area` and answers region queries
quickly.

These queries are deliberately *conservative*. Every item is reduced to a center
and a half-extent (its bounding square/cube), so a query may return false
positives, but it will **never omit** an item that truly matches. Callers are
expected to run a precise narrow-phase test (via `isec*` or `gjk*`) on the
returned candidates.

```go
tree := query3d.NewOctree[Entity](query3d.OctreeSettings{})

// Insert returns an ID used for later Update / Remove.
id := tree.Insert(query3d.AreaFromSphere(bounds), entity)

// Query a region; the visitor returns false to stop early.
tree.QueryAABB(query3d.NewAABB(-10, -10, -10, 10, 10, 10), func(e Entity) bool {
    // narrow-phase test against e here
    return true
})

// Query along a ray.
tree.QuerySegment(query3d.NewSegment(from, to), func(e Entity) bool {
    return true
})
```

A `VisitorBucket` is a convenient visitor that collects matches into a reusable
slice when you would rather iterate results after the query returns.

## Scenes

`placement2d` and `placement3d` are the top of the stack. A `Scene` holds a
collection of placed **objects** and **meshes** and handles both the broad phase
and the narrow phase for you, so most applications interact only with this layer.

- **Objects** are dynamic entities. Each owns one or more convex shapes
  (circles/rectangles in 2D, spheres/boxes in 3D) and can be moved by updating
  its transform.
- **Meshes** are static entities made of edges (2D) or triangles (3D), suited to
  level geometry.

Both are indexed in their own broad-phase tree, and narrow-phase overlap is
resolved with GJK/EPA internally. Intersections are reported as `Contact` values
through the same `ContactCallback` sinks as the lower layers.

The `Scene` is generic over three user-data types, one for each kind of entity
(`O` for objects, `S` for shapes, `M` for meshes), letting you attach your own
identifiers to whatever a query returns:

```go
scene := placement3d.NewScene[*Body, *Collider, *Level](placement3d.SceneSettings{})

// A dynamic object with a single sphere collider.
objID := scene.CreateObject(placement3d.ObjectInfo[*Body]{
    Position: opt.V(dprec.NewVec3(0, 5, 0)),
    UserData: body,
})
scene.AttachSphere(objID, placement3d.SphereInfo[*Collider]{
    Sphere:   shape3d.NewSphere(dprec.ZeroVec3(), 0.5),
    UserData: collider,
})

// Move it around over time.
scene.SetObjectTransform(objID, newTransform)

// Cast a ray through the scene.
segment := shape3d.NewSegment(eye, target)
if contact, ok := scene.CheckSegmentIntersection(segment, placement3d.Filter{}); ok {
    // ...
}
```

### Filtering

Every shape and mesh carries `FilterInfo` describing which layers it lives in
(`SourceMask` / `TargetMask`) and an optional `RejectGroup` that prevents
entities in the same group from colliding (useful, for example, so the parts of
one articulated body ignore one another). Queries then take a `Filter` that can
restrict the search by layer mask and skip dynamic or static entities:

```go
scene.CollectSphereIntersections(probe, placement3d.Filter{
    Mask:        opt.V(uint32(LayerEnemies)),
    SkipStatic:  true,
}, func(contact placement3d.Contact) {
    // handle each overlap with a dynamic enemy
})
```
