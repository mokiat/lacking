package animation

import (
	"iter"
	"maps"
	"math"

	"github.com/mokiat/gog/opt"
	"github.com/mokiat/gomath/dprec"
)

// NewRecording creates a new empty recording.
func NewRecording() *Recording {
	return &Recording{
		name:      "",
		startTime: 0.0,
		endTime:   1.0,
		bindings:  make(map[string]KeyframeSet),
	}
}

var _ Source = (*Recording)(nil)

// Recording represents a pre-stored set of keyframmes that can be played back.
type Recording struct {
	name      string
	startTime float64
	endTime   float64
	bindings  map[string]KeyframeSet
}

// Name returns the name of the recording.
func (r *Recording) Name() string {
	return r.name
}

// SetName sets the name of the recording.
func (r *Recording) SetName(name string) *Recording {
	r.name = name
	return r
}

// StartTime returns the time (in seconds) at which the recording starts.
func (r *Recording) StartTime() float64 {
	return r.startTime
}

// SetStartTime sets the time (in seconds) at which the recording starts.
func (r *Recording) SetStartTime(startTime float64) *Recording {
	r.startTime = startTime
	return r
}

// EndTime returns the time (in seconds) at which the recording ends.
func (r *Recording) EndTime() float64 {
	return r.endTime
}

// SetEndTime sets the time (in seconds) at which the recording ends.
func (r *Recording) SetEndTime(endTime float64) *Recording {
	r.endTime = endTime
	return r
}

// Binding returns the keyframes for the node with the specified name.
func (r *Recording) Binding(name string) (KeyframeSet, bool) {
	if binding, ok := r.bindings[name]; ok {
		return binding, true
	}
	return KeyframeSet{}, false
}

// SetBinding sets the keyframes for the node with the specified name.
func (r *Recording) SetBinding(name string, keyframes KeyframeSet) *Recording {
	r.bindings[name] = keyframes
	return r
}

// RemoveBinding removes the binding for the node with the specified name.
func (r *Recording) RemoveBinding(name string) *Recording {
	delete(r.bindings, name)
	return r
}

// BoundNodesIter returns an iterator over the names of all nodes that have
// keyframes in this recording.
func (r *Recording) BoundNodesIter() iter.Seq[string] {
	return maps.Keys(r.bindings)
}

// Playback creates a new animtion node that plays back the animation.
func (r *Recording) Playback(loop bool) *PlaybackNode {
	return NewPlaybackNode(r, loop)
}

// Pose returns a new animation node that represents a fixed pose from the
// recording.
func (r *Recording) Pose(timestamp, length float64) *PoseNode {
	return NewPoseNode(r, timestamp, length)
}

// Length returns the length of the recording in seconds.
func (r *Recording) Length() float64 {
	return max(0.0, r.EndTime()-r.StartTime())
}

// BoneTransformDelta returns the transformation that occurred for the
// specified bone between the from and to animation timestamps .
//
// This is mostly used for root motion.
func (r *Recording) BoneTransformDelta(bone string, fromTimestamp, toTimestamp float64) NodeTransform {
	if toTimestamp < fromTimestamp {
		return InverseNodeTransform(r.BoneTransformDelta(bone, toTimestamp, fromTimestamp))
	}

	resultMatrix := dprec.IdentityMat4()

	length := r.Length()
	modFrom := math.Mod(fromTimestamp, length)
	modTo := math.Mod(toTimestamp, length)

	fromMatrix := r.getMatrixAt(bone, modFrom)

	for (modTo < modFrom) && (toTimestamp > fromTimestamp) {
		toMatrix := r.getMatrixAt(bone, length-0.00001) // prevent mod down to zero
		deltaMatrix := dprec.Mat4Prod(
			toMatrix,
			dprec.InverseMat4(fromMatrix),
		)
		resultMatrix = dprec.Mat4Prod(
			deltaMatrix,
			resultMatrix,
		)

		fromMatrix = r.getMatrixAt(bone, 0.0)
		toTimestamp -= length
	}

	toMatrix := r.getMatrixAt(bone, modTo)
	deltaMatrix := dprec.Mat4Prod(
		toMatrix,
		dprec.InverseMat4(fromMatrix),
	)
	resultMatrix = dprec.Mat4Prod(
		deltaMatrix,
		resultMatrix,
	)

	translation, rotation, scale := resultMatrix.TRS()
	return NodeTransform{
		Translation: opt.V(translation),
		Rotation:    opt.V(rotation),
		Scale:       opt.V(scale),
	}
}

// BoneTransformAt returns the transformation for the specified bone
// at the specified animation timestamp.
func (r *Recording) BoneTransformAt(bone string, timestamp float64) NodeTransform {
	binding, ok := r.bindings[bone]
	if !ok {
		return NodeTransform{}
	}
	timestamp = r.adjustTimestamp(timestamp)
	var result NodeTransform
	if len(binding.TranslationKeyframes) > 0 {
		result.Translation = opt.V(binding.Translation(timestamp))
	}
	if len(binding.RotationKeyframes) > 0 {
		result.Rotation = opt.V(binding.Rotation(timestamp))
	}
	if len(binding.ScaleKeyframes) > 0 {
		result.Scale = opt.V(binding.Scale(timestamp))
	}
	return result
}

func (r *Recording) adjustTimestamp(timestamp float64) float64 {
	length := r.Length()
	timestamp = dprec.Mod(timestamp, length)
	if timestamp < 0.0 {
		timestamp += length
	}
	return r.StartTime() + timestamp
}

func (r *Recording) getMatrixAt(bone string, t float64) dprec.Mat4 {
	t = r.adjustTimestamp(t)
	transform := r.BoneTransformAt(bone, t)
	return dprec.TRSMat4(
		transform.Translation.ValueOrDefault(dprec.ZeroVec3()),
		transform.Rotation.ValueOrDefault(dprec.IdentityQuat()),
		transform.Scale.ValueOrDefault(dprec.NewVec3(1.0, 1.0, 1.0)),
	)
}
