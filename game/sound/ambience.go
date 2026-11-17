package sound

import (
	"time"

	"github.com/mokiat/gog/opt"
	"github.com/mokiat/lacking/core/audio"
)

// AmbienceID uniquely identifies an ambience that was started through
// [AmbienceView.Play].
//
// The zero value, also available as [NilAmbienceID], does not reference a
// valid ambience.
type AmbienceID struct {
	index    int32
	revision int32
}

// NilAmbienceID is an [AmbienceID] that never references a valid ambience.
var NilAmbienceID = AmbienceID{}

// AmbienceInfo describes an ambience to be played through
// [AmbienceView.Play].
type AmbienceInfo struct {

	// Media is the sound to be looped. It must be specified.
	Media audio.Media

	// Emitter is the emitter on which the ambience is played, for sounds
	// that have a location, such as a waterfall. If it is [NilEmitterID],
	// the ambience is not spatial and is heard the same regardless of
	// where the listener is, which suits sounds such as wind or rain.
	Emitter EmitterID

	// Gain is the gain of the ambience.
	//
	// Defaults to 1.
	Gain opt.T[float64]

	// FadeIn is the time over which the ambience rises from silence to
	// its gain. If it is not positive, the ambience starts at its gain.
	FadeIn time.Duration
}

// AmbienceView provides access to the ambiences that are playing in a
// [Scene], and to the mixing controls that they share (see
// [AmbienceView.Bus]).
//
// An ambience is a background sound that loops until it is stopped, and
// that can be faded in and out.
//
// An AmbienceView is a lightweight accessor around a [Scene] and can be
// obtained through [Scene.Ambience].
type AmbienceView struct {
	scene *Scene
}

// Bus returns a [BusView] with the mixing controls that are shared by all
// ambiences.
//
// Pausing the bus also suspends the fades of the ambiences.
func (v AmbienceView) Bus() BusView {
	return BusView{
		bus: &v.scene.ambienceBus,
	}
}

// Play starts a new ambience and returns its ID. The ambience loops until
// it is stopped through [AmbienceView.Stop].
//
// It panics if no media is specified, or if the emitter is neither
// [NilEmitterID] nor a valid emitter.
func (v AmbienceView) Play(info AmbienceInfo) AmbienceID {
	playback := v.scene.createPlayback(&v.scene.ambienceBus, info.Media, info.Emitter)
	playback.gain = max(0.0, info.Gain.ValueOrDefault(1.0))

	fade := newFader(0.0)
	fade.fadeTo(1.0, info.FadeIn)
	playback.fade = fade.value

	index, ambience := allocateSlot(&v.scene.ambiences, v.scene.freeAmbienceIndices)
	*ambience = ambienceState{
		revision: ambience.revision + 1, // progress revision to valid (odd) value
		playback: playback,
		fade:     fade,
	}

	playback.setLooping(true)
	v.scene.syncPlayback(&ambience.playback)
	playback.start(0.0)

	return AmbienceID{
		index:    index,
		revision: ambience.revision,
	}
}

// PlayHandle behaves like [AmbienceView.Play] but wraps the resulting ID
// in an [AmbienceHandle], as returned by [AmbienceView.Handle], for
// callers that want to keep acting on the new ambience without holding
// onto its ID separately.
func (v AmbienceView) PlayHandle(info AmbienceInfo) AmbienceHandle {
	return v.Handle(v.Play(info))
}

// Stop stops the ambience with the specified ID, fading it out over the
// specified time.
//
// If fadeOut is not positive, the ambience is stopped right away and its
// ID becomes invalid. Otherwise the ID stays valid until the fade has
// completed during a subsequent [Scene.Update].
//
// It panics if the ID does not reference a valid ambience.
func (v AmbienceView) Stop(id AmbienceID, fadeOut time.Duration) {
	ambience := v.resolve(id, true)
	if fadeOut <= 0 {
		ambience.playback.stop()
		v.scene.releaseAmbience(id.index)
		return
	}
	ambience.stopping = true
	ambience.fade.fadeTo(0.0, fadeOut)
}

// StopAll stops all ambiences, fading them out over the specified time.
//
// See [AmbienceView.Stop] for further details.
func (v AmbienceView) StopAll(fadeOut time.Duration) {
	for i := range v.scene.ambiences {
		if ambience := &v.scene.ambiences[i]; ambience.isValid() {
			v.Stop(AmbienceID{
				index:    int32(i),
				revision: ambience.revision,
			}, fadeOut)
		}
	}
}

// Handle returns an [AmbienceHandle] that wraps the specified ID, as a
// more convenient means of repeatedly accessing the same ambience without
// having to pass its ID to this view on every call.
func (v AmbienceView) Handle(id AmbienceID) AmbienceHandle {
	return AmbienceHandle{
		view: v,
		id:   id,
	}
}

// IsValid returns whether the specified ID references an ambience that
// is still playing, which includes one that is fading out.
func (v AmbienceView) IsValid(id AmbienceID) bool {
	ambience := v.resolve(id, false)
	return ambience != nil
}

// Gain returns the gain of the specified ambience, not taking fades into
// account.
//
// It panics if the ID does not reference a valid ambience.
func (v AmbienceView) Gain(id AmbienceID) float64 {
	ambience := v.resolve(id, true)
	return ambience.playback.gain
}

// SetGain changes the gain of the specified ambience. Negative values are
// clamped to 0.
//
// It panics if the ID does not reference a valid ambience.
func (v AmbienceView) SetGain(id AmbienceID, gain float64) {
	ambience := v.resolve(id, true)
	ambience.playback.gain = max(0.0, gain)
	v.scene.syncPlayback(&ambience.playback)
}

// resolve looks up the ambienceState referenced by id. If id is stale or
// otherwise invalid, resolve panics when required is true, or returns nil
// otherwise.
func (v AmbienceView) resolve(id AmbienceID, required bool) *ambienceState {
	if id.revision == 0 {
		if required {
			panic("invalid ambience ID")
		}
		return nil
	}
	ambience := &v.scene.ambiences[id.index]
	if ambience.revision != id.revision {
		if required {
			panic("invalid ambience ID")
		}
		return nil
	}
	return ambience
}

// AmbienceHandle is an object-oriented alternative to [AmbienceView] that
// is bound to a specific [AmbienceID].
//
// It is obtained through [AmbienceView.Handle].
//
// Wrapping an ID this way needs no allocation of its own, but unlike a
// plain [AmbienceID] (which holds no pointers), a Handle keeps an internal
// reference to the owning [Scene]. This keeps that Scene reachable for as
// long as the handle is retained, and adds a pointer that the garbage
// collector must trace wherever the handle is stored. Prefer storing IDs
// over handles in long-lived collections unless the convenience is worth
// that cost.
type AmbienceHandle struct {
	view AmbienceView
	id   AmbienceID
}

// ID returns the [AmbienceID] wrapped by this handle.
func (h AmbienceHandle) ID() AmbienceID {
	return h.id
}

// Stop stops the wrapped ambience, fading it out over the specified time.
//
// See [AmbienceView.Stop] for further details.
func (h AmbienceHandle) Stop(fadeOut time.Duration) {
	h.view.Stop(h.id, fadeOut)
}

// IsValid returns whether the wrapped ambience is still playing.
func (h AmbienceHandle) IsValid() bool {
	return h.view.IsValid(h.id)
}

// Gain returns the gain of the wrapped ambience.
//
// See [AmbienceView.Gain] for further details.
func (h AmbienceHandle) Gain() float64 {
	return h.view.Gain(h.id)
}

// SetGain changes the gain of the wrapped ambience.
//
// See [AmbienceView.SetGain] for further details.
func (h AmbienceHandle) SetGain(gain float64) {
	h.view.SetGain(h.id, gain)
}

// ambienceState holds the internal state of a single playing ambience, as
// tracked by a [Scene].
type ambienceState struct {
	revision int32
	playback playbackState
	fade     fader
	stopping bool // whether the ambience is removed once it has faded out
}

// isValid returns whether this state is currently backing a playing
// ambience, as opposed to a freed slot awaiting reuse.
func (s *ambienceState) isValid() bool {
	return s.revision%2 == 1 // only odd revisions are valid
}
