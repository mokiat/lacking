---
title: Overview
---

# Animation

The `game/animation` package provides skeletal animation playback through trees of composable animation nodes. Pre-recorded animation data (clips) is played back by leaf nodes, and inner nodes combine the output of their children by blending, layering, switching between, or retiming them. The root of the tree is advanced once per update and produces the local transformation of every bone of an animated model.

## Core Concepts

| Concept | Description |
|---|---|
| **Transform** | Local translation, rotation, and scale of a single bone, relative to its parent. |
| **Skeleton** | Assigns a stable index and a rest transformation to each bone name, so per-frame work uses indices instead of names. |
| **Clip** | Pre-recorded animation: one channel per animated bone, covering a time range. |
| **Channel** | Animation data of a single bone in a clip; produces the bone's `Transform` at any keyframe timestamp. |
| **Node** | Element of an animation tree. Keeps track of its playback position and produces the `Transform` of each bone at that position. |
| **RootMotion** | Extracts the motion of a single bone from a tree, so a character can be moved in the world in agreement with its animation. |

Skeletons, clips, and channels are immutable once created and can be shared freely. Nodes hold playback state and belong to a single tree.

## Transforms

A `Transform` holds a translation, a rotation (unit quaternion), and a per-axis scale:

```go
t := animation.Transform{
    Translation: dprec.NewVec3(0.0, 1.0, 0.0),
    Rotation:    dprec.RotationQuat(dprec.Degrees(45), dprec.BasisYVec3()),
    Scale:       dprec.NewVec3(1.0, 1.0, 1.0),
}

matrix := t.Matrix()                     // scale, then rotate, then translate
back := animation.Mat4Transform(matrix) // inverse of Matrix
```

The package provides three ways of combining transforms:

| Function | Effect |
|---|---|
| `BlendTransforms(a, b, factor)` | Interpolates from `a` (factor `0.0`) to `b` (factor `1.0`). Translation and scale are linear, rotation is a normalized linear interpolation along the shorter arc. The factor is not clamped. |
| `AddTransforms(base, delta)` | Applies `delta` as an additive layer on top of `base`. Translations are added, scales multiplied, and rotations composed. The order matters. |
| `DiffTransforms(a, b)` | The additive difference of `a` relative to `b`, so that `AddTransforms(b, DiffTransforms(a, b)) == a`. |

`IdentityTransform()` returns a zero translation, an identity rotation, and a unit scale. It is the neutral element of `AddTransforms`.

## Animation Data

### Skeletons

A `Skeleton` is created from a list of bones. Bones receive consecutive indices in the order of their first occurrence; later duplicates are ignored.

```go
skeleton := animation.NewSkeleton([]animation.SkeletonBone{
    {Name: "root", RestTransform: animation.IdentityTransform()},
    {Name: "spine", RestTransform: spineRest},
    {Name: "head", RestTransform: headRest},
})

skeleton.BoneCount()           // 3
bone, ok := skeleton.BoneIndex("spine") // 1, true
skeleton.BoneName(bone)        // "spine"
skeleton.BoneRestTransform(bone)
```

The rest transformation is what a bone shows when an absolute node does not animate it.

When working with the `game` package, a loaded `Model` already exposes a skeleton built from its node hierarchy through `model.Skeleton()`.

### Keyframes and Channels

A `KeyframeList[T]` is a list of `Keyframe[T]` values ordered by strictly ascending timestamp. `Sample` returns the two keyframes surrounding a timestamp together with the interpolation factor between them, clamping to the nearest keyframe outside the covered range.

A `Channel` produces the complete local transformation of its bone at a keyframe timestamp. A timestamp outside the range of the data holds the value at the nearest end. Three implementations are provided, differing in how they store keyframes:

| Channel | Storage |
|---|---|
| `TransformKeyframeChannel` | A single keyframe list of complete transforms; all components share timestamps. |
| `ComponentKeyframeChannel` | Separate keyframe lists for translation, rotation, and scale, each keyed at its own timestamps. |
| `FixedRateKeyframeChannel` | Complete transforms sampled at a fixed rate from a start timestamp; no per-frame timestamps and no search on lookup. |

```go
channel := animation.NewComponentKeyframeChannel(
    animation.KeyframeList[dprec.Vec3]{
        {Timestamp: 0.0, Value: dprec.NewVec3(0.0, 0.0, 0.0)},
        {Timestamp: 1.0, Value: dprec.NewVec3(0.0, 1.0, 0.0)},
    },
    animation.KeyframeList[dprec.Quat]{
        {Timestamp: 0.0, Value: dprec.IdentityQuat()},
    },
    animation.KeyframeList[dprec.Vec3]{
        {Timestamp: 0.0, Value: dprec.NewVec3(1.0, 1.0, 1.0)},
    },
)

fixed := animation.NewFixedRateKeyframeChannel(0.0, 30.0, frames) // 30 frames per second
```

Every list must contain at least one keyframe (the constructors panic otherwise). A component that is not animated needs a single keyframe holding its constant value, typically the rest value of the bone. Channels take ownership of the slices they are given.

### Clips

A `Clip` groups channels, keyed by bone name, with a name and a time range:

```go
clip := animation.NewClip("walk", 0.0, 1.2, map[string]animation.Channel{
    "root":  rootChannel,
    "spine": spineChannel,
})

clip.Length()            // 1.2
clip.HasChannel("head")  // false
for name := range clip.ChannelNamesIter() {
    fmt.Println(name)
}
```

`Crop` derives a new clip with a different time range over the same channels, which is useful when several animations are stored back-to-back in a single recording. The times passed to `Crop` are keyframe timestamps, not offsets relative to the original start.

```go
idle := recording.Crop("idle", 0.0, 2.0)
wave := recording.Crop("wave", 2.0, 3.5)
```

When working with the `game` package, the clips of a loaded model are available through `model.Recordings()` and `model.FindRecording(name)`.

## Node Trees

Playback is performed by a tree of `Node` values. Leaf nodes produce animation data, either by playing back a clip or by holding a static pose. Inner nodes derive their output from that of their children.

```go
walk := animation.NewClipNode(skeleton, model.FindRecording("walk"), true)
run := animation.NewClipNode(skeleton, model.FindRecording("run"), true)

locomotion := animation.Blend(
    animation.Synchronized(walk),
    animation.Synchronized(run),
)
locomotion.SetWeight(0.3)
```

Each frame, the root of the tree is advanced and the resulting pose is read bone by bone:

```go
locomotion.Advance(deltaSeconds)

for bone := range skeleton.BoneCount() {
    t := locomotion.Transform(bone)
    // apply t to the scene node of this bone
}
```

A parent node takes control of the playback of its children. Once a node is part of a tree, it should not be advanced or relocated by any other means.

### The Node Interface

| Method | Meaning |
|---|---|
| `Rate` | Fraction of the animation length that elapses per second. Equals `1.0 / length` at normal speed, or `0.0` for nodes with no meaningful length (such as static poses). |
| `Fraction` / `SetFraction` | Playback position as a fraction of the animation length, in the range `[0.0..1.0]`. |
| `Advance` | Moves playback by a number of seconds. A negative delta moves backwards. |
| `IsSynchronized` / `SetSynchronized` | Whether the node should be kept in phase with its parent (see [Synchronization](#synchronization)). |
| `Transform` | Transformation of a bone at the current position. |
| `DeltaTransform` | Motion a bone experiences over the next delta seconds, without changing the playback position (see [Root Motion](#root-motion)). |

Bones are passed as indices in the skeleton that the leaf nodes were created with. All nodes in a tree are expected to share the same skeleton.

### Absolute and Additive Nodes

Nodes come in two kinds:

- An **absolute** node produces the local transformation of each bone, relative to its parent. Bones it does not animate return their rest transformation.
- An **additive** node produces a difference that is meant to be applied on top of an absolute transformation through `AddTransforms`. Bones it does not animate return the `IdentityTransform`.

A `DiffNode` turns absolute animations into an additive one and an `AddNode` applies an additive animation on top of an absolute one. Blending nodes expect all of their children to be of the same kind, which is also the kind of their result.

### Construction Shorthands

Several nodes have a shorthand function that makes it convenient to build a tree in a single expression:

| Shorthand | Equivalent |
|---|---|
| `Blend(first, second)` | `NewBlendNode(first, second)` |
| `Add(base, additive)` | `NewAddNode(base, additive)` |
| `Diff(source, reference)` | `NewDiffNode(source, reference)` |
| `Reverse(node)` | `NewReverseNode(node)` |
| `ScaleTime(node, scale)` | `NewTimeScaleNode(node)` followed by `SetTimeScale(scale)` |
| `Synchronized(node)` | Calls `SetSynchronized(true)` on the node and returns it. |

## Leaf Nodes

### Clip Playback

`ClipNode` plays back a single clip. The clip's start time maps to fraction `0.0` and its end time to fraction `1.0`. One second of `Advance` corresponds to one unit of keyframe time.

```go
node := animation.NewClipNode(skeleton, clip, true) // looping
node.SetFraction(0.5)                                // jump to the middle
```

A looping node wraps around at either end, whereas a non-looping node stops at the end it reaches. Clip channels are matched to skeleton bones by name: channels for unknown bones are ignored, and bones without a channel hold their rest transformation.

### Static Poses

Static pose nodes have no playback of their own. Their rate and fraction are always `0.0`, and advancing or relocating them has no effect.

| Node | Kind | Pose |
|---|---|---|
| `ClipPoseNode` | Absolute | The pose a clip has at a timestamp (relative to the clip start). The timestamp can be changed with `SetTimestamp`. |
| `FixedPoseNode` | Absolute | Bone transformations given at construction, keyed by bone name. Unlisted bones hold their rest transformation. |
| `AdditivePoseNode` | Additive | Bone differences given at construction, keyed by bone name. Unlisted bones hold the `IdentityTransform`. |

```go
// The first frame of the aim animation, used as a reference pose.
aimRest := animation.NewClipPoseNode(skeleton, aimClip, 0.0)

// A fixed offset that bends the spine forward.
lean := animation.NewAdditivePoseNode(skeleton, map[string]animation.Transform{
    "spine": {
        Translation: dprec.ZeroVec3(),
        Rotation:    dprec.RotationQuat(dprec.Degrees(15), dprec.BasisXVec3()),
        Scale:       dprec.NewVec3(1.0, 1.0, 1.0),
    },
})
```

## Blending Nodes

### Two-Way Blend

`BlendNode` blends two animations by a weight, clamped to `[0.0..1.0]`. A weight of `0.0` shows the first animation and `1.0` the second.

```go
locomotion := animation.Blend(
    animation.Synchronized(walk),
    animation.Synchronized(run),
)
locomotion.SetWeight(speedFactor)
```

A `BlendNode` keeps a playback position of its own and advances at a rate blended from the rates of its synchronized children, so that, for example, the feet of a walk and a run cycle stay aligned while the length of the blended cycle changes smoothly with the weight. Its fraction wraps around, so it is meant for cyclic animations.

### One-Dimensional Blend Space

`Blend1DNode` generalizes `BlendNode` to any number of animations placed at coordinates along a line. Setting the blend coordinate to a gameplay value, such as the current speed, shows a blend of the two entries surrounding it. A coordinate beyond either end shows the closest entry alone.

```go
locomotion := animation.NewBlend1DNode(
    animation.Blend1DEntry{Coord: 0.0, Node: animation.Synchronized(idle)},
    animation.Blend1DEntry{Coord: 1.5, Node: animation.Synchronized(walk)},
    animation.Blend1DEntry{Coord: 4.0, Node: animation.Synchronized(jog)},
    animation.Blend1DEntry{Coord: 7.0, Node: animation.Synchronized(run)},
)
locomotion.SetBlendCoord(currentSpeed)
```

Entries need not be sorted. Entries that share a coordinate produce an abrupt switch at that coordinate. At least one entry is required.

### Masked Blend

`MaskedBlendNode` blends a layer animation over a base animation bone by bone, according to a mask of per-bone weights. This makes it possible, for example, to play a different animation on the upper body while the legs keep playing the base animation.

```go
body := animation.NewMaskedBlendNode(skeleton, locomotion, aim, map[string]float64{
    "spine":     0.5,
    "chest":     1.0,
    "upper_arm": 1.0,
    "forearm":   1.0,
})
body.SetWeight(1.0) // overall weight; scales all mask weights
```

Bones without an entry in the mask have a weight of `0.0` and show the base animation. The overall weight allows the layer to be faded in and out as a whole. The base node drives the playback of the `MaskedBlendNode`.

## Layering Nodes

### Additive Layers

`DiffNode` produces the difference of a source animation relative to a reference animation. Both children are absolute and the result is additive. The reference is commonly a static pose, such as the first frame of the source, so that the result holds only the motion of the source.

`AddNode` applies an additive animation on top of an absolute base. Its weight controls the strength of the layer: `1.0` applies it in full, `0.0` leaves the base unchanged. The weight is not clamped and values outside `[0.0..1.0]` extrapolate.

```go
breathingClip := model.FindRecording("breathing")
breathing := animation.Diff(
    animation.NewClipNode(skeleton, breathingClip, true),
    animation.NewClipPoseNode(skeleton, breathingClip, 0.0),
)

layered := animation.Add(locomotion, breathing)
layered.SetWeight(0.8)
```

In both nodes, the first child (the source or the base) drives the playback.

### One-Shot Overlays

`OneShotNode` plays a base animation and, each time it is triggered, plays an overlay animation over it exactly once, such as a recoil or a wave on top of locomotion. The overlay is blended over the base according to a per-bone mask, in the same way as by a `MaskedBlendNode`. Passing a `nil` mask applies the overlay to all bones.

```go
upperBody := map[string]float64{
    "spine": 0.5, "chest": 1.0, "upper_arm": 1.0, "forearm": 1.0,
}

action := animation.NewOneShotNode(skeleton, locomotion, waveNode, upperBody)
action.SetFadeInDuration(0.2)
action.SetFadeOutDuration(0.3)

// Later, in response to a gameplay event.
action.Trigger()

// Optionally cut it short, fading out from the current weight.
action.Stop()

action.IsPlaying() // true while the overlay plays, including its fades
```

The overlay is advanced only while it plays and starts from its beginning each time it is triggered. It plays until one cycle has elapsed (as determined by its rate) or until it is stopped. Triggering has no effect while the overlay is playing, unless it was stopped and is fading out, in which case it starts over and fades back in from its current weight. The overlay must have a rate greater than `0.0`, otherwise `Trigger` panics.

If the overlay is shorter than the fade-in and fade-out durations combined, the two fades meet and the overlay does not reach full weight.

## State Machines

`StateMachineNode` plays one of a set of states, each represented by a node of its own, and cross-fades between them. States are identified by keys of any comparable type.

```go
type State int

const (
    StateIdle State = iota
    StateWalk
    StateJump
)

machine := animation.NewStateMachineNode(
    StateIdle,
    map[State]animation.Node{
        StateIdle: idle,
        StateWalk: walk,
        StateJump: jump,
    },
    animation.StateMachineTransitionSettings{Duration: 0.25}, // default transition
    []animation.StateMachineTransition[State]{
        {
            From: StateWalk,
            To:   StateJump,
            StateMachineTransitionSettings: animation.StateMachineTransitionSettings{
                Duration: 0.1,
                Rewind:   true,
            },
        },
    },
)
```

The default transition is used between any two states that have no transition of their own. A transition specifies:

| Setting | Meaning |
|---|---|
| `Duration` | Time in seconds of the cross-fade. `0.0` switches immediately. |
| `Rewind` | Whether the state transitioned to starts from its beginning instead of resuming from its own playback position. Has no effect on synchronized states. |

### Changing State

```go
machine.TransitionTo(StateWalk) // cross-fade using the configured transition
machine.JumpTo(StateIdle)       // switch immediately, canceling any transition
```

| Method | Meaning |
|---|---|
| `CurrentState` | The state the node is in. During a transition, the state it leads to. |
| `SourceState` | The state being faded out. Equal to `CurrentState` when not transitioning. |
| `IsTransitioning` | Whether a transition is in progress. |
| `TransitionProgress` | Progress of the transition, from `0.0` (only the source) to `1.0` (only the current state). |

Transitioning to the current state has no effect. If a transition is already in progress, the state it leads to becomes the source of the new transition, without a fade. As an exception, transitioning back to the state that is being faded out reverses the transition in progress from where it is.

A synchronized state that is entered is moved to the fraction of the node, so that it continues in phase with the state it replaces. A non-synchronized state resumes from its own playback position (or from its beginning, if the transition rewinds). To start a state from a particular position after `JumpTo`, call `SetFraction` afterwards.

Only the current state and the state being faded out are advanced. The remaining states keep their playback position until they are entered again.

`NewStateMachineNode` panics if the initial state or any transition endpoint is not among the states, if a transition leads to the state it starts from, if two transitions share the same direction between the same states, or if any duration is negative. `JumpTo` and `TransitionTo` panic on an unknown state.

## Playback Modifiers

### Time Scale

`TimeScaleNode` plays back another node with its time multiplied by a scale. A scale of `2.0` plays twice as fast, `0.5` at half speed, `0.0` freezes playback, and a negative scale plays backwards.

```go
slowMotion := animation.ScaleTime(walk, 0.5)
slowMotion.SetTimeScale(0.25)
```

### Reverse

`ReverseNode` plays back another node in reverse. Fraction `0.0` of a `ReverseNode` shows the end of the delegate, and advancing it moves the delegate backwards in time.

```go
walkBackwards := animation.Reverse(walk)
```

Both modifiers are transparent wrappers: they hold no playback position of their own, their synchronization setting is that of the delegate, and whether playback wraps or stops at either end is decided by the delegate.

## Synchronization

Playback progress is expressed as a fraction of the animation length, which allows animations of different lengths to be kept in phase. A node marked as synchronized is kept at the fraction of its parent; a node that is not synchronized plays at its own pace. Nodes start out as not synchronized.

```go
locomotion := animation.Blend(
    animation.Synchronized(walk), // 1.2s cycle
    animation.Synchronized(run),  // 0.8s cycle
)
```

How a parent treats its children falls into two groups:

| Parent | Fraction | Synchronized children |
|---|---|---|
| `BlendNode`, `Blend1DNode`, `StateMachineNode` (during a transition) | Own playback position, advancing at a rate blended from the synchronized children. Wraps around. | Kept at the parent's fraction. |
| `AddNode`, `DiffNode`, `MaskedBlendNode`, `OneShotNode` | That of the first child (base or source). | The second child is kept at the first child's fraction if synchronized. |

For the second group, the synchronization setting of the first child has no effect, since that child defines the fraction of the parent. Whether it is kept in phase with the grandparent is decided by the synchronization setting of the parent itself. In a `OneShotNode`, the synchronization setting of the overlay has no effect either, since the overlay always starts from its beginning when triggered.

All children are advanced in real time, whether synchronized or not; synchronized children are then snapped to the appropriate fraction. A synchronized child that does not loop is moved back to its start when the fraction of a wrapping parent wraps around.

The blended rate is computed so that the length of the cycle, rather than the rate itself, changes linearly with the weight. Only synchronized children with a rate greater than `0.0` take part; if none do, the parent's rate is `0.0` and its fraction does not advance.

## Root Motion

Root motion extracts the motion of a single bone, typically the root of a character, from an animation tree, so that the character itself can be moved in the world (for example through a character controller) in agreement with its animation.

`RootMotion` resolves a bone name against a skeleton once and reports the motion of that bone through `Node.DeltaTransform`. Call it before advancing the tree by the same delta:

```go
rootMotion := animation.NewRootMotion(skeleton, tree, "root")

// Each update.
delta := rootMotion.DeltaTransform(deltaSeconds)
tree.Advance(deltaSeconds)

// Move the character by delta (expressed in the space of the bone's parent).
```

The returned transformation `D` satisfies `D * A = B`, where `A` and `B` are the matrices of the bone at the current position and `deltaSeconds` later. If the bone is not in the skeleton, the `IdentityTransform` is returned.

How nodes report motion:

- A looping `ClipNode` measures motion across any number of wrap-arounds as if each repetition continued from where the previous one ended. A non-looping one stops measuring at the end it reaches.
- Static poses, `DiffNode`, and the additive layer of an `AddNode` contribute no motion.
- Blending nodes blend the motion of their children. A bone with an effective mask weight of `0.0` moves as in the base, so a mask that excludes the root bone leaves root motion entirely to the base animation.
- Blended motion is an approximation of the motion of the blended transformation, especially when synchronized children play at a rate different from that of their parent.

## Integration with the Game Scene

The `game` package connects animation trees to model hierarchies:

```go
model := scene.InstantiateModel(modelInfo)
skeleton := model.Skeleton()

tree := animation.NewClipNode(skeleton, model.FindRecording("idle"), true)

model.BindAnimation(tree) // drive the model's animated nodes from the tree
scene.PlayAnimation(tree) // advance the tree on each scene update

// Later.
scene.StopAnimation(tree)
```

`BindAnimation` binds the model's animated hierarchy nodes to the bones of the tree with matching names, and `BindAnimationSubtree` does the same for every node in the subtree of a named node. The bound nodes receive the local transformation of their bone through a hierarchy binding (see [Hierarchy - Bindings](../hierarchy/index.md#bindings)). `PlayAnimation` registers the root of the tree with the scene, which advances it at the fixed update rate. Only the root should be registered, since parents advance their children.

## Limitations

- **Single skeleton per tree.** Bones are addressed by index, so all nodes in a tree must be built from the same skeleton. Retargeting between skeletons is not provided.
- **No inverse kinematics.** The package produces forward-kinematic poses only. Procedural adjustments can be approximated with `AdditivePoseNode` and `AddNode`.
- **No animation events.** There is no mechanism to fire callbacks at specific timestamps (such as footsteps); applications can poll `Fraction` instead.
