package sound

import (
	"math/rand/v2"

	"github.com/mokiat/gog/opt"
	"github.com/mokiat/lacking/core/audio"
)

// EffectPoolID uniquely identifies an effect pool that was created through
// [EffectPoolView.Create].
//
// The zero value, also available as [NilEffectPoolID], does not reference
// a valid effect pool.
type EffectPoolID struct {
	index    int32
	revision int32
}

// NilEffectPoolID is an [EffectPoolID] that never references a valid
// effect pool.
var NilEffectPoolID = EffectPoolID{}

// EffectPoolOverflow determines what an effect pool does when it is asked
// to play a sound while all of its playbacks are in use.
type EffectPoolOverflow uint8

const (
	// EffectPoolOverflowSteal restarts the playback that has been playing
	// the longest.
	EffectPoolOverflowSteal EffectPoolOverflow = iota

	// EffectPoolOverflowSkip leaves the playing playbacks untouched and
	// does not play the new sound.
	EffectPoolOverflowSkip
)

// EffectPoolInfo describes an effect pool to be created through
// [EffectPoolView.Create].
type EffectPoolInfo struct {

	// Media is the sound that the pool plays. It must be specified.
	Media audio.Media

	// Capacity is the number of playbacks in the pool, which is the number
	// of times the sound can be heard simultaneously. It must be positive.
	Capacity int

	// Spatial indicates whether the pool plays its sound on emitters. If
	// it is false, the sound is heard the same regardless of where the
	// listener is.
	//
	// Defaults to true.
	Spatial opt.T[bool]

	// Gain is the gain at which the sound is played.
	//
	// Defaults to 1.
	Gain opt.T[float64]

	// Pitch is the rate at which the sound is played, where 2 is twice as
	// fast and an octave higher. The value must be positive.
	//
	// Defaults to 1.
	Pitch opt.T[float64]

	// GainVariance is the fraction of the gain by which each individual
	// play can randomly deviate, so that a repeated sound is less
	// monotonous. For example, 0.1 produces gains within 10% of Gain. The
	// value must be in the range [0, 1].
	GainVariance float64

	// PitchVariance is the fraction of the pitch by which each individual
	// play can randomly deviate. The value must be in the range [0, 1).
	PitchVariance float64

	// Overflow determines what happens when the pool is asked to play its
	// sound while all of its playbacks are in use.
	//
	// Defaults to [EffectPoolOverflowSteal].
	Overflow EffectPoolOverflow
}

// EffectPoolView provides access to the effect pools that belong to a
// [Scene].
//
// An effect pool is intended for a sound that is played often, such as a
// footstep or a gunshot. It allocates a fixed number of playbacks when it
// is created and reuses them afterwards, so that playing the sound neither
// allocates memory nor creates audio resources.
//
// Pooled effects are mixed together with the individual effects and are
// subject to the controls of [EffectView].
//
// An EffectPoolView is a lightweight accessor around a [Scene] and can be
// obtained through [Scene.EffectPools].
type EffectPoolView struct {
	scene *Scene
}

// Create allocates a new effect pool and returns its ID.
//
// It panics if no media is specified or if the capacity is not positive.
func (v EffectPoolView) Create(info EffectPoolInfo) EffectPoolID {
	if info.Media == nil {
		panic("media must be specified")
	}
	if info.Capacity <= 0 {
		panic("effect pool capacity must be positive")
	}
	spatial := info.Spatial.ValueOrDefault(true)

	playbacks := make([]pooledPlayback, info.Capacity)
	for i := range playbacks {
		pooled := &playbacks[i] // stable, since the slice is never resized
		pooled.playback = playbackState{
			gain:  1.0,
			fade:  1.0,
			pitch: 1.0,
		}
		if spatial {
			playback := v.scene.api.CreateSpatialPlayback(v.scene.effectBus.bus, info.Media, audio.PlaybackSettings{})
			pooled.playback.playback = playback
			pooled.playback.spatial = playback
		} else {
			pooled.playback.playback = v.scene.api.CreatePlayback(v.scene.effectBus.bus, info.Media, audio.PlaybackSettings{})
		}
		pooled.playback.setOnFinished(func() {
			pooled.active = false
		})
	}

	index, pool := allocateSlot(&v.scene.effectPools, v.scene.freeEffectPoolIndices)
	*pool = effectPoolState{
		revision:      pool.revision + 1, // progress revision to valid (odd) value
		playbacks:     playbacks,
		spatial:       spatial,
		gain:          max(0.0, info.Gain.ValueOrDefault(1.0)),
		pitch:         info.Pitch.ValueOrDefault(1.0),
		gainVariance:  info.GainVariance,
		pitchVariance: info.PitchVariance,
		overflow:      info.Overflow,
	}

	return EffectPoolID{
		index:    index,
		revision: pool.revision,
	}
}

// CreateHandle behaves like [EffectPoolView.Create] but wraps the
// resulting ID in an [EffectPoolHandle], as returned by
// [EffectPoolView.Handle], for callers that want to keep acting on the new
// effect pool without holding onto its ID separately.
func (v EffectPoolView) CreateHandle(info EffectPoolInfo) EffectPoolHandle {
	return v.Handle(v.Create(info))
}

// Delete removes the effect pool with the specified ID, stopping all of
// its playbacks.
//
// It panics if the ID does not reference a valid effect pool. Use
// [EffectPoolView.IsValid] first if the ID's validity is not otherwise
// guaranteed.
func (v EffectPoolView) Delete(id EffectPoolID) {
	pool := v.resolve(id, true)

	for i := range pool.playbacks {
		playback := &pool.playbacks[i].playback
		playback.stop()
		playback.release()
	}
	*pool = effectPoolState{
		revision: pool.revision + 1, // progress revision to invalid (even) value
	}

	v.scene.freeEffectPoolIndices.Push(id.index)
}

// Each calls cb once for every effect pool that is currently alive within
// this Scene, in unspecified order.
func (v EffectPoolView) Each(cb func(id EffectPoolID)) {
	for i := range v.scene.effectPools {
		if pool := &v.scene.effectPools[i]; pool.isValid() {
			cb(EffectPoolID{
				index:    int32(i),
				revision: pool.revision,
			})
		}
	}
}

// Handle returns an [EffectPoolHandle] that wraps the specified ID, as a
// more convenient means of repeatedly accessing the same effect pool
// without having to pass its ID to this view on every call.
func (v EffectPoolView) Handle(id EffectPoolID) EffectPoolHandle {
	return EffectPoolHandle{
		view: v,
		id:   id,
	}
}

// IsValid returns whether the specified ID references an effect pool that
// has not been deleted.
func (v EffectPoolView) IsValid(id EffectPoolID) bool {
	pool := v.resolve(id, false)
	return pool != nil
}

// Play plays the sound of the specified effect pool on the specified
// emitter and returns whether the sound was started. The latter is only
// ever false for pools that use [EffectPoolOverflowSkip], when all
// playbacks are in use.
//
// A pool that is not spatial ignores the emitter, hence [NilEmitterID]
// should be passed.
//
// It panics if the ID does not reference a valid effect pool, or if the
// pool is spatial and emitterID does not reference a valid emitter.
func (v EffectPoolView) Play(id EffectPoolID, emitterID EmitterID) bool {
	pool := v.resolve(id, true)
	if pool.spatial {
		v.scene.Emitters().resolve(emitterID, true)
	} else {
		emitterID = NilEmitterID
	}

	var selected *pooledPlayback
	for i := range pool.playbacks {
		pooled := &pool.playbacks[i]
		if !pooled.active {
			selected = pooled
			break
		}
		if selected == nil || pooled.order < selected.order {
			selected = pooled // oldest so far
		}
	}
	if selected.active && pool.overflow == EffectPoolOverflowSkip {
		return false
	}

	pool.playCount++
	selected.active = true
	selected.order = pool.playCount
	selected.playback.emitter = emitterID
	selected.playback.gain = pool.gain * (1.0 + pool.gainVariance*(2.0*rand.Float64()-1.0))
	selected.playback.pitch = pool.pitch * (1.0 + pool.pitchVariance*(2.0*rand.Float64()-1.0))
	v.scene.syncPlayback(&selected.playback)
	selected.playback.start(0.0)
	return true
}

// StopAll stops all playbacks of the specified effect pool.
//
// It panics if the ID does not reference a valid effect pool.
func (v EffectPoolView) StopAll(id EffectPoolID) {
	pool := v.resolve(id, true)
	for i := range pool.playbacks {
		if pooled := &pool.playbacks[i]; pooled.active {
			pooled.playback.stop()
			pooled.active = false
		}
	}
}

// PlayingCount returns the number of playbacks of the specified effect
// pool that are in use.
//
// It panics if the ID does not reference a valid effect pool.
func (v EffectPoolView) PlayingCount(id EffectPoolID) int {
	pool := v.resolve(id, true)
	count := 0
	for i := range pool.playbacks {
		if pool.playbacks[i].active {
			count++
		}
	}
	return count
}

// resolve looks up the effectPoolState referenced by id. If id is stale or
// otherwise invalid, resolve panics when required is true, or returns nil
// otherwise.
func (v EffectPoolView) resolve(id EffectPoolID, required bool) *effectPoolState {
	if id.revision == 0 {
		if required {
			panic("invalid effect pool ID")
		}
		return nil
	}
	pool := &v.scene.effectPools[id.index]
	if pool.revision != id.revision {
		if required {
			panic("invalid effect pool ID")
		}
		return nil
	}
	return pool
}

// EffectPoolHandle is an object-oriented alternative to [EffectPoolView]
// that is bound to a specific [EffectPoolID].
//
// It is obtained through [EffectPoolView.Handle].
//
// Wrapping an ID this way needs no allocation of its own, but unlike a
// plain [EffectPoolID] (which holds no pointers), a Handle keeps an
// internal reference to the owning [Scene]. This keeps that Scene
// reachable for as long as the handle is retained, and adds a pointer
// that the garbage collector must trace wherever the handle is stored.
// Prefer storing IDs over handles in long-lived collections unless the
// convenience is worth that cost.
type EffectPoolHandle struct {
	view EffectPoolView
	id   EffectPoolID
}

// ID returns the [EffectPoolID] wrapped by this handle.
func (h EffectPoolHandle) ID() EffectPoolID {
	return h.id
}

// Delete removes the wrapped effect pool.
//
// See [EffectPoolView.Delete] for further details.
func (h EffectPoolHandle) Delete() {
	h.view.Delete(h.id)
}

// IsValid returns whether the wrapped effect pool has not been deleted.
func (h EffectPoolHandle) IsValid() bool {
	return h.view.IsValid(h.id)
}

// Play plays the sound of the wrapped effect pool on the specified
// emitter.
//
// See [EffectPoolView.Play] for further details.
func (h EffectPoolHandle) Play(emitterID EmitterID) bool {
	return h.view.Play(h.id, emitterID)
}

// StopAll stops all playbacks of the wrapped effect pool.
//
// It panics if the handle does not reference a valid effect pool.
func (h EffectPoolHandle) StopAll() {
	h.view.StopAll(h.id)
}

// PlayingCount returns the number of playbacks of the wrapped effect
// pool that are in use.
//
// It panics if the handle does not reference a valid effect pool.
func (h EffectPoolHandle) PlayingCount() int {
	return h.view.PlayingCount(h.id)
}

// effectPoolState holds the internal state of a single effect pool, as
// tracked by a [Scene].
type effectPoolState struct {
	revision      int32
	spatial       bool
	playbacks     []pooledPlayback
	gain          float64
	pitch         float64
	gainVariance  float64
	pitchVariance float64
	overflow      EffectPoolOverflow
	playCount     uint64
}

// isValid returns whether this state is currently backing a live effect
// pool, as opposed to a freed slot awaiting reuse.
func (s *effectPoolState) isValid() bool {
	return s.revision%2 == 1 // only odd revisions are valid
}

// pooledPlayback is a reusable playback of an effect pool.
type pooledPlayback struct {
	playback playbackState
	active   bool   // whether the playback is playing
	order    uint64 // increases with every play; lower means older
}
