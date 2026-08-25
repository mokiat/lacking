// Package hierarchy provides a scene graph of 3D nodes arranged in
// parent-child relationships.
//
// A [Scene] owns all nodes and their storage. Nodes are not manipulated
// directly; instead, [Scene.Nodes] returns a [NodeView], a lightweight,
// freely-copyable handle to the scene through which nodes are created,
// deleted, traversed, and transformed. Each node is referenced by a [NodeID],
// a stable identifier that is invalidated when the node is deleted, even if its
// storage slot is later reused. A [NodeHandle] binds a [NodeID] to its
// [NodeView] so that a single node can be operated on without passing the two
// around separately.
//
// # Transformations
//
// Every node has a local translation, rotation, and scale, expressed relative
// to its parent. Its absolute (world) transformation is the composition of its
// local transformation with those of all its ancestors, exposed as a matrix by
// [NodeView.AbsoluteMatrix]. Absolute matrices are computed lazily and cached,
// so reading them is cheap when nothing has changed. A node can be made
// independent (see [NodeView.SetIndependent]), in which case its local
// transformation is used directly as its absolute transformation, ignoring its
// ancestors while it otherwise remains part of the hierarchy.
//
// # Interpolation
//
// To support smooth rendering between fixed-rate updates, each node also
// retains its previous absolute transformation. [Scene.AdvanceStep] records the
// current transformations as the previous ones and should be called once per
// fixed-rate step; [NodeView.InterpolatedAbsoluteMatrix] then blends between the
// previous and current pose by a fraction. [NodeView.Snap] collapses that
// interpolation for a single node, which is used to teleport it without
// rendering sliding across the jump.
//
// # Traversal
//
// Nodes can be visited in several ways: [NodeView.Each] over every node,
// [NodeView.EachRoot] over root nodes, and [NodeView.Walk] and
// [NodeView.WalkSubtree] for depth-first, pre-order traversal of the whole
// scene or a single subtree. Each has an iterator counterpart for use with
// range-over-func. Nodes may also be located by name with [NodeView.FindNode]
// and [NodeView.FindNodeInSubtree].
package hierarchy
