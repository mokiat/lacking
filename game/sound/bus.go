package sound

import (
	"github.com/mokiat/gomath/dprec"
	"github.com/mokiat/lacking/core/audio"
)

// BusView provides access to the mixing controls of a single sound
// category (music, speech, effects or ambience). The controls apply to
// everything that plays in that category, on top of the settings of the
// individual sounds.
//
// A BusView is a lightweight accessor around a [Scene] and can be obtained
// through [MusicView.Bus], [SpeechView.Bus], [EffectView.Bus] or
// [AmbienceView.Bus].
type BusView struct {
	bus *busState
}

// Gain returns the gain that is applied to the category. The default is 1.
func (v BusView) Gain() float64 {
	return v.bus.gain
}

// SetGain changes the gain that is applied to the category. Negative
// values are clamped to 0.
func (v BusView) SetGain(gain float64) {
	v.bus.setGain(gain)
}

// Volume returns the perceived loudness of the category, which is the
// square root of the gain. The default is 1.
func (v BusView) Volume() float64 {
	return v.bus.volume()
}

// SetVolume changes the perceived loudness of the category. It is a more
// natural fit for volume sliders than [BusView.SetGain], since the gain is
// set to the square of the volume. Negative values are clamped to 0.
func (v BusView) SetVolume(volume float64) {
	v.bus.setVolume(volume)
}

// Muted returns whether the category is silenced.
func (v BusView) Muted() bool {
	return v.bus.muted
}

// SetMuted changes whether the category is silenced. Muting does not
// change the gain, which takes effect again once unmuted, and does not
// stop the sounds of the category from playing.
func (v BusView) SetMuted(muted bool) {
	v.bus.setMuted(muted)
}

// Paused returns whether the category is paused.
func (v BusView) Paused() bool {
	return v.bus.paused
}

// Pause suspends everything that plays in the category, including the
// sounds that are started while paused, until [BusView.Resume] is called.
func (v BusView) Pause() {
	v.bus.pause()
}

// Resume continues the sounds that were suspended through
// [BusView.Pause].
func (v BusView) Resume() {
	v.bus.resume()
}

// busState holds the mixing state of a single sound category (music,
// speech, effects or ambience) and the audio bus that backs it.
type busState struct {
	bus    audio.Bus
	gain   float64
	muted  bool
	paused bool
}

func newBusState(api audio.API) busState {
	return busState{
		// The category has neither reverb nor compression of its own. The
		// compression of the master bus is what keeps the mix in range.
		bus: api.CreateBus(audio.BusSettings{
			UseReverb:      false,
			UseCompression: false,
		}),
		gain: 1.0,
	}
}

func (b *busState) setGain(gain float64) {
	b.gain = max(0.0, gain)
	b.apply()
}

// volume returns the perceptual counterpart of the gain, where the gain is
// the square of the volume.
func (b *busState) volume() float64 {
	return dprec.Sqrt(b.gain)
}

func (b *busState) setVolume(volume float64) {
	volume = max(0.0, volume)
	b.setGain(volume * volume)
}

func (b *busState) setMuted(muted bool) {
	b.muted = muted
	b.apply()
}

func (b *busState) pause() {
	if !b.paused {
		b.paused = true
		b.bus.Pause()
	}
}

func (b *busState) resume() {
	if b.paused {
		b.paused = false
		b.bus.Resume()
	}
}

// release releases the underlying audio bus, which stops everything that
// is attached to it. The bus must not be used afterwards.
func (b *busState) release() {
	b.bus.Release()
}

// apply pushes the effective gain to the underlying audio bus.
func (b *busState) apply() {
	if b.muted {
		b.bus.SetGain(0.0)
	} else {
		b.bus.SetGain(float32(b.gain))
	}
}
