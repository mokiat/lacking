package sound

import (
	"time"

	"github.com/mokiat/gomath/sprec"
	"github.com/mokiat/lacking/core/audio"
)

// playbackState holds the internal state of a single playback that is
// driven by a [Scene]. It is the building block of effects, pooled effects,
// speech lines and ambiences.
//
// A playback is spatial when it was created on an emitter, in which case
// its placement and Doppler shift follow that emitter.
type playbackState struct {
	playback audio.Playback
	spatial  audio.SpatialEmitter // nil when the playback is not spatial
	emitter  EmitterID

	gain  float64
	fade  float32
	pitch float64

	// The following fields track what was last sent to the audio API, so
	// that unchanged values are not sent again.
	synced                bool
	appliedGain           float32
	appliedRate           float32
	appliedPosition       sprec.Vec3
	appliedRotation       sprec.Quat
	appliedInnerConeAngle sprec.Angle
	appliedOuterConeAngle sprec.Angle
	appliedOuterConeGain  float32
}

// start begins playing from the specified offset in seconds.
func (s *playbackState) start(at float64) {
	s.playback.Start(at)
}

// stop halts the playback.
func (s *playbackState) stop() {
	s.playback.Stop()
}

// release releases the audio resources of the playback, which must not be
// used afterwards.
func (s *playbackState) release() {
	s.playback.Release()
}

// setLooping changes whether the playback repeats once it reaches its end.
func (s *playbackState) setLooping(looping bool) {
	s.playback.SetLooping(looping)
}

// setOnFinished changes the callback that is invoked once the playback
// reaches its end on its own.
func (s *playbackState) setOnFinished(callback func()) {
	s.playback.SetOnFinished(callback)
}

// fader moves a value linearly towards a target over a period of time. It
// is used for fades and crossfades.
type fader struct {
	value    float32
	start    float32
	target   float32
	duration float64 // in seconds
	elapsed  float64 // in seconds
}

func newFader(value float32) fader {
	return fader{
		value:  value,
		start:  value,
		target: value,
	}
}

// fadeTo makes the fader reach target after the specified duration,
// starting from its current value. A non-positive duration makes the
// change immediate.
func (f *fader) fadeTo(target float32, duration time.Duration) {
	f.start = f.value
	f.target = target
	f.duration = duration.Seconds()
	f.elapsed = 0.0
	if duration <= 0 {
		f.value = target
	}
}

func (f *fader) advance(seconds float64) {
	if f.isDone() {
		return
	}
	f.elapsed += seconds
	if f.elapsed >= f.duration {
		f.value = f.target
	} else {
		f.value = sprec.Mix(f.start, f.target, float32(f.elapsed/f.duration))
	}
}

func (f *fader) isDone() bool {
	return f.value == f.target
}
