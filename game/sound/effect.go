package sound

import (
	"github.com/mokiat/gog/opt"
	"github.com/mokiat/lacking/core/audio"
)

// EffectID uniquely identifies a sound effect that was started through
// [EffectView.Play].
//
// The zero value, also available as [NilEffectID], does not reference a
// valid effect.
type EffectID struct {
	index    int32
	revision int32
}

// NilEffectID is an [EffectID] that never references a valid effect.
var NilEffectID = EffectID{}

// EffectInfo describes a sound effect to be played through
// [EffectView.Play].
type EffectInfo struct {

	// Media is the sound to be played. It must be specified.
	Media audio.Media

	// Emitter is the emitter on which the effect is played. If it is
	// [NilEmitterID], the effect is not spatial and is heard the same
	// regardless of where the listener is.
	Emitter EmitterID

	// Gain is the gain of the effect.
	//
	// Defaults to 1.
	Gain opt.T[float64]

	// Pitch is the rate at which the effect is played, where 2 is twice as
	// fast and an octave higher. The value must be positive.
	//
	// Defaults to 1.
	Pitch opt.T[float64]

	// Loop indicates whether the effect should repeat until it is stopped
	// through [EffectView.Stop].
	Loop bool
}

// EffectView provides access to the individual sound effects that are
// playing in a [Scene], and to the mixing controls that all effects share
// (see [EffectView.Bus]), including the ones played through an effect pool
// (see [EffectPoolView]).
//
// An EffectView is a lightweight accessor around a [Scene] and can be
// obtained through [Scene.Effects].
type EffectView struct {
	scene *Scene
}

// Bus returns a [BusView] with the mixing controls that are shared by all
// effects, including the ones played through an effect pool.
func (v EffectView) Bus() BusView {
	return BusView{
		bus: &v.scene.effectBus,
	}
}

// Play starts a new sound effect and returns its ID.
//
// The ID stays valid for as long as the effect is playing. Once an effect
// that is not looping reaches its end, it is removed during the next
// [Scene.Update] and its ID becomes invalid.
//
// Each call allocates audio resources. Prefer an effect pool (see
// [EffectPoolView]) for sounds that are played often.
//
// It panics if no media is specified, or if the emitter is neither
// [NilEmitterID] nor a valid emitter.
func (v EffectView) Play(info EffectInfo) EffectID {
	playback := v.scene.createPlayback(&v.scene.effectBus, info.Media, info.Emitter)
	playback.gain = max(0.0, info.Gain.ValueOrDefault(1.0))
	playback.pitch = info.Pitch.ValueOrDefault(1.0)

	index, effect := allocateSlot(&v.scene.effects, v.scene.freeEffectIndices)
	*effect = effectState{
		revision: effect.revision + 1, // progress revision to valid (odd) value
		playback: playback,
	}
	id := EffectID{
		index:    index,
		revision: effect.revision,
	}

	scene := v.scene
	playback.setLooping(info.Loop)
	playback.setOnFinished(func() {
		if effect := scene.Effects().resolve(id, false); effect != nil {
			effect.finished = true
		}
	})
	scene.syncPlayback(&effect.playback)
	playback.start(0.0)

	return id
}

// PlayHandle behaves like [EffectView.Play] but wraps the resulting ID in
// an [EffectHandle], as returned by [EffectView.Handle], for callers that
// want to keep acting on the new effect without holding onto its ID
// separately.
func (v EffectView) PlayHandle(info EffectInfo) EffectHandle {
	return v.Handle(v.Play(info))
}

// Stop stops the effect with the specified ID, which makes the ID
// invalid. Subscribers of [EffectView.SubscribeFinished] are not notified.
//
// It panics if the ID does not reference a valid effect. Use
// [EffectView.IsValid] first, since an effect that is not looping becomes
// invalid on its own once it reaches its end.
func (v EffectView) Stop(id EffectID) {
	effect := v.resolve(id, true)
	effect.playback.stop()
	v.scene.releaseEffect(id.index)
}

// Handle returns an [EffectHandle] that wraps the specified ID, as a more
// convenient means of repeatedly accessing the same effect without having
// to pass its ID to this view on every call.
func (v EffectView) Handle(id EffectID) EffectHandle {
	return EffectHandle{
		view: v,
		id:   id,
	}
}

// IsValid returns whether the specified ID references an effect that is
// still playing.
func (v EffectView) IsValid(id EffectID) bool {
	effect := v.resolve(id, false)
	return effect != nil
}

// Gain returns the gain of the specified effect.
//
// It panics if the ID does not reference a valid effect.
func (v EffectView) Gain(id EffectID) float64 {
	effect := v.resolve(id, true)
	return effect.playback.gain
}

// SetGain changes the gain of the specified effect. Negative values are
// clamped to 0.
//
// It panics if the ID does not reference a valid effect.
func (v EffectView) SetGain(id EffectID, gain float64) {
	effect := v.resolve(id, true)
	effect.playback.gain = max(0.0, gain)
	v.scene.syncPlayback(&effect.playback)
}

// Pitch returns the rate at which the specified effect is played, not
// taking the Doppler effect into account.
//
// It panics if the ID does not reference a valid effect.
func (v EffectView) Pitch(id EffectID) float64 {
	effect := v.resolve(id, true)
	return effect.playback.pitch
}

// SetPitch changes the rate at which the specified effect is played, which
// is useful for sounds that follow a changing quantity, such as an engine.
// The value must be positive.
//
// It panics if the ID does not reference a valid effect.
func (v EffectView) SetPitch(id EffectID, pitch float64) {
	effect := v.resolve(id, true)
	effect.playback.pitch = pitch
	v.scene.syncPlayback(&effect.playback)
}

// SubscribeFinished registers a callback that is invoked, during
// [Scene.Update], whenever an effect has played to its end. Effects that
// are stopped, be it explicitly or because their emitter is deleted, are
// not reported.
//
// Call [EffectFinishedSubscription.Delete] on the returned subscription to
// stop receiving notifications.
func (v EffectView) SubscribeFinished(callback EffectFinishedCallback) *EffectFinishedSubscription {
	return v.scene.effectFinishedSubscriptions.Subscribe(callback)
}

// resolve looks up the effectState referenced by id. If id is stale or
// otherwise invalid, resolve panics when required is true, or returns nil
// otherwise.
func (v EffectView) resolve(id EffectID, required bool) *effectState {
	if id.revision == 0 {
		if required {
			panic("invalid effect ID")
		}
		return nil
	}
	effect := &v.scene.effects[id.index]
	if effect.revision != id.revision {
		if required {
			panic("invalid effect ID")
		}
		return nil
	}
	return effect
}

// EffectHandle is an object-oriented alternative to [EffectView] that is
// bound to a specific [EffectID].
//
// It is obtained through [EffectView.Handle].
//
// Wrapping an ID this way needs no allocation of its own, but unlike a
// plain [EffectID] (which holds no pointers), a Handle keeps an internal
// reference to the owning [Scene]. This keeps that Scene reachable for as
// long as the handle is retained, and adds a pointer that the garbage
// collector must trace wherever the handle is stored. Prefer storing IDs
// over handles in long-lived collections unless the convenience is worth
// that cost.
type EffectHandle struct {
	view EffectView
	id   EffectID
}

// ID returns the [EffectID] wrapped by this handle.
func (h EffectHandle) ID() EffectID {
	return h.id
}

// Stop stops the wrapped effect.
//
// See [EffectView.Stop] for further details.
func (h EffectHandle) Stop() {
	h.view.Stop(h.id)
}

// IsValid returns whether the wrapped effect is still playing.
func (h EffectHandle) IsValid() bool {
	return h.view.IsValid(h.id)
}

// Gain returns the gain of the wrapped effect.
//
// It panics if the handle does not reference a valid effect.
func (h EffectHandle) Gain() float64 {
	return h.view.Gain(h.id)
}

// SetGain changes the gain of the wrapped effect.
//
// See [EffectView.SetGain] for further details.
func (h EffectHandle) SetGain(gain float64) {
	h.view.SetGain(h.id, gain)
}

// Pitch returns the rate at which the wrapped effect is played.
//
// See [EffectView.Pitch] for further details.
func (h EffectHandle) Pitch() float64 {
	return h.view.Pitch(h.id)
}

// SetPitch changes the rate at which the wrapped effect is played.
//
// See [EffectView.SetPitch] for further details.
func (h EffectHandle) SetPitch(pitch float64) {
	h.view.SetPitch(h.id, pitch)
}

// effectState holds the internal state of a single playing effect, as
// tracked by a [Scene].
type effectState struct {
	revision int32
	playback playbackState
	finished bool // set once the audio API reports that playback has ended
}

// isValid returns whether this state is currently backing a playing
// effect, as opposed to a freed slot awaiting reuse.
func (s *effectState) isValid() bool {
	return s.revision%2 == 1 // only odd revisions are valid
}
