package game

import (
	"math"

	"github.com/mokiat/gog/opt"
	"github.com/mokiat/gomath/dprec"
	"github.com/mokiat/lacking/game/hierarchy"
)

// NodeTransform represents the transformation of a node.
type NodeTransform struct {

	// Translation, if specified, indicates the translation of the node.
	Translation opt.T[dprec.Vec3]

	// Rotation, if specified, indicates the rotation of the node.
	Rotation opt.T[dprec.Quat]

	// Scale, if specified, indicates the scale of the node.
	Scale opt.T[dprec.Vec3]
}

// AnimationSource represents a source of animation data.
type AnimationSource interface {

	// NodeTransform returns the transformation of the node with the
	// specified name.
	NodeTransform(name string) NodeTransform
}

// AnimationBlending represents an animation source that blends two animation
// sources.
type AnimationBlending struct {
	first  AnimationSource
	second AnimationSource
	factor float64
}

// Factor returns the blending factor of the node. A value of 0.0 means
// that the first source is used, a value of 1.0 means that the second
// source is used.
func (n *AnimationBlending) Factor() float64 {
	return n.factor
}

// SetFactor sets the blending factor of the node.
func (n *AnimationBlending) SetFactor(factor float64) {
	n.factor = dprec.Clamp(factor, 0.0, 1.0)
}

// NodeTransform returns the transformation of the node with the specified
// name. The transformation is a blend of the transformations of the two
// sources of the node.
func (n *AnimationBlending) NodeTransform(name string) NodeTransform {
	firstTransform := n.first.NodeTransform(name)
	secondTransform := n.second.NodeTransform(name)

	switch {
	case n.factor < 0.000001: // optimization
		return firstTransform
	case n.factor > 0.999999: // optimization
		return secondTransform
	default:
		return NodeTransform{
			Translation: combineLinear(firstTransform.Translation, secondTransform.Translation, n.factor),
			Rotation:    combineSpherical(firstTransform.Rotation, secondTransform.Rotation, n.factor),
			Scale:       combineLinear(firstTransform.Scale, secondTransform.Scale, n.factor),
		}
	}
}

type AnimationDefinitionInfo struct {
	Name      string
	StartTime float64
	EndTime   float64
	Bindings  []AnimationBindingDefinitionInfo
}

type AnimationBindingDefinitionInfo struct {
	NodeName             string
	TranslationKeyframes KeyframeList[dprec.Vec3]
	RotationKeyframes    KeyframeList[dprec.Quat]
	ScaleKeyframes       KeyframeList[dprec.Vec3]
}

type AnimationDefinition struct {
	name      string
	startTime float64
	endTime   float64
	bindings  []AnimationBindingDefinitionInfo
}

func (d *AnimationDefinition) Name() string {
	return d.name
}

type AnimationInfo struct {
	Root       *hierarchy.Node
	Definition *AnimationDefinition
}

type Animation struct {
	name       string
	definition *AnimationDefinition
	bindings   []animationBinding
}

func (a *Animation) Name() string {
	return a.name
}

func (a *Animation) StartTime() float64 {
	return a.definition.startTime
}

func (a *Animation) EndTime() float64 {
	return a.definition.endTime
}

func (a *Animation) Length() float64 {
	return a.EndTime() - a.StartTime()
}

func (a *Animation) BindingTransform(index int, timestamp float64) NodeTransform {
	var result NodeTransform
	if index < 0 || index >= len(a.bindings) {
		return result
	}
	binding := a.bindings[index]
	if binding.node == nil {
		return result
	}
	if len(binding.translationKeyframes) > 0 {
		result.Translation = opt.V(binding.Translation(timestamp))
	}
	if len(binding.rotationKeyframes) > 0 {
		result.Rotation = opt.V(binding.Rotation(timestamp))
	}
	if len(binding.scaleKeyframes) > 0 {
		result.Scale = opt.V(binding.Scale(timestamp))
	}
	return result
}

func combineLinear(first, second opt.T[dprec.Vec3], amount float64) opt.T[dprec.Vec3] {
	switch {
	case first.Specified && second.Specified:
		return opt.V(dprec.Vec3Lerp(first.Value, second.Value, amount))
	case first.Specified:
		return first
	case second.Specified:
		return second
	default:
		return opt.Unspecified[dprec.Vec3]()
	}
}

func combineSpherical(first, second opt.T[dprec.Quat], amount float64) opt.T[dprec.Quat] {
	switch {
	case first.Specified && second.Specified:
		return opt.V(dprec.QuatSlerp(first.Value, second.Value, amount))
	case first.Specified:
		return first
	case second.Specified:
		return second
	default:
		return opt.Unspecified[dprec.Quat]()
	}
}

type animationBinding struct {
	node                 *hierarchy.Node
	translationKeyframes KeyframeList[dprec.Vec3]
	rotationKeyframes    KeyframeList[dprec.Quat]
	scaleKeyframes       KeyframeList[dprec.Vec3]
}

func (b animationBinding) Translation(timestamp float64) dprec.Vec3 {
	left, right, t := b.translationKeyframes.Keyframe(timestamp)
	return dprec.Vec3Lerp(left.Value, right.Value, t)
}

func (b animationBinding) Rotation(timestamp float64) dprec.Quat {
	left, right, t := b.rotationKeyframes.Keyframe(timestamp)
	return dprec.QuatSlerp(left.Value, right.Value, t)
}

func (b animationBinding) Scale(timestamp float64) dprec.Vec3 {
	left, right, t := b.scaleKeyframes.Keyframe(timestamp)
	return dprec.Vec3Lerp(left.Value, right.Value, t)
}

type KeyframeList[T any] []Keyframe[T]

func (l KeyframeList[T]) Keyframe(timestamp float64) (Keyframe[T], Keyframe[T], float64) {
	leftIndex := 0
	rightIndex := len(l) - 1
	for leftIndex < rightIndex-1 {
		middleIndex := (leftIndex + rightIndex) / 2
		middle := l[middleIndex]
		if middle.Timestamp <= timestamp {
			leftIndex = middleIndex
		}
		if middle.Timestamp >= timestamp {
			rightIndex = middleIndex
		}
	}
	left := l[leftIndex]
	right := l[rightIndex]
	if leftIndex == rightIndex {
		return left, right, 0
	}
	t := dprec.Clamp((timestamp-left.Timestamp)/(right.Timestamp-left.Timestamp), 0.0, 1.0)
	return left, right, t
}

type Keyframe[T any] struct {
	Timestamp float64
	Value     T
}

type Playback struct {
	scene     *Scene
	animation *Animation

	name      string
	head      float64
	startTime float64
	endTime   float64
	speed     float64
	playing   bool
	loop      bool
}

func (p *Playback) Name() string {
	return p.name
}

func (p *Playback) SetName(name string) {
	p.name = name
}

func (p *Playback) Play() {
	p.playing = true
}

func (p *Playback) Pause() {
	p.playing = false
}

func (p *Playback) Stop() {
	p.Pause()
	p.head = p.animation.StartTime()
}

func (p *Playback) Loop() bool {
	return p.loop
}

func (p *Playback) SetLoop(loop bool) {
	p.loop = loop
}

func (p *Playback) Speed() float64 {
	return p.speed
}

func (p *Playback) SetSpeed(speed float64) {
	p.speed = speed
}

func (p *Playback) Advance(amount float64) {
	p.head += amount * p.speed
	if p.head > p.endTime {
		if p.loop {
			p.head = p.startTime + math.Mod(p.head, p.Length())
		} else {
			p.head = p.endTime
			p.Pause()
		}
	}
}

func (p *Playback) NodeTransform(name string) NodeTransform {
	for i, binding := range p.animation.bindings {
		if binding.node.Name() == name {
			return p.animation.BindingTransform(i, p.head)
		}
	}
	return NodeTransform{}
}

func (p *Playback) Seek(head float64) {
	p.head = head
}

func (p *Playback) Head() float64 {
	return p.head
}

func (p *Playback) StartTime() float64 {
	return p.startTime
}

func (p *Playback) SetStartTime(startTime float64) {
	p.startTime = startTime
}

func (p *Playback) EndTime() float64 {
	return p.endTime
}

func (p *Playback) SetEndTime(endTime float64) {
	p.endTime = endTime
}

func (p *Playback) Length() float64 {
	return p.endTime - p.startTime
}

func (p *Playback) Delete() {
	p.scene.playbacks.Remove(p)
	p.scene.playbackPool.Restore(p)
}
