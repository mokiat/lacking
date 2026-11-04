// Package animation provides skeletal animation playback through trees of
// composable animation nodes.
//
// # Animation data
//
// A [Skeleton] assigns a stable index to each bone of an animated model,
// along with its rest transformation, so that per-frame work can refer to
// bones by index instead of by name. A [Clip] is a pre-recorded animation
// made of one [Channel] per animated bone. A Channel produces the local
// [Transform] of its bone at any keyframe timestamp; [TransformKeyframeChannel],
// [ComponentKeyframeChannel] and [FixedRateKeyframeChannel] offer different
// trade-offs in how the keyframes are stored. Skeletons, clips and channels
// are immutable once created and can be shared freely.
//
// # Node trees
//
// Playback is performed by a tree of [Node] values. Leaf nodes produce
// animation data, either by playing back a clip ([ClipNode]) or by holding a
// static pose ([ClipPoseNode], [FixedPoseNode], [AdditivePoseNode]). Inner
// nodes derive their output from that of their children, for example by
// blending them ([BlendNode], [Blend1DNode], [MaskedBlendNode]), by layering
// them ([AddNode], [DiffNode], [OneShotNode]), by switching between them
// ([StateMachineNode]) or by altering their playback ([TimeScaleNode],
// [ReverseNode]). Shorthand functions such as [Blend], [Add] and
// [Synchronized] make it convenient to construct a tree in a single
// expression.
//
// Each frame, the root of the tree is advanced through [Node.Advance] and the
// resulting pose is read bone by bone through [Node.Transform]. A node takes
// control of the playback of its children, which should not be advanced or
// relocated through other means.
//
// # Absolute and additive nodes
//
// Nodes come in two kinds. An absolute node produces the local
// transformation of each bone, while an additive node produces a difference
// that is meant to be applied on top of an absolute transformation through
// [AddTransforms]. A [DiffNode] turns absolute animations into an additive
// one and an [AddNode] applies an additive animation on top of an absolute
// one. See [Node.Transform] for details.
//
// # Synchronization
//
// Playback progress is expressed as a fraction of the animation length,
// which allows animations of different lengths to be kept in phase. A node
// that is marked as synchronized (see [Node.SetSynchronized]) is kept at the
// fraction of its parent, so that, for example, the feet of a walk and a run
// cycle stay aligned while the two are blended. Nodes that are not
// synchronized play at their own pace.
//
// # Root motion
//
// A [RootMotion] extracts the motion of a single bone, typically the root of
// a character, from a node tree through [Node.DeltaTransform], so that the
// character can be moved in the world in agreement with its animation.
package animation
