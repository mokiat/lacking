package sound

import "github.com/mokiat/gomath/dprec"

// EmitterID uniquely identifies an emitter that was created through
// [EmitterView.Create].
//
// The zero value, also available as [NilEmitterID], does not reference a
// valid emitter. Where a sound can be played on an emitter, NilEmitterID
// stands for no emitter at all, which makes the sound non-spatial.
type EmitterID struct {
	index    int32
	revision int32
}

// NilEmitterID is an [EmitterID] that never references a valid emitter.
var NilEmitterID = EmitterID{}

// EmitterView provides access to the emitters that belong to a [Scene].
//
// An emitter is a point in space, with an orientation and a velocity, on
// which effects, speech lines and ambiences are played. Everything that is
// playing on an emitter follows it as it moves, and is heard relative to
// the active listener of the scene.
//
// Changes to an emitter are applied to the sounds that are playing on it
// during the next [Scene.Update].
//
// An EmitterView is a lightweight accessor around a [Scene] and can be
// obtained through [Scene.Emitters].
type EmitterView struct {
	scene *Scene
}

// Create allocates a new emitter at the specified world position and
// returns its ID.
//
// The new emitter has an identity rotation, zero velocity, a gain of 1,
// the Doppler effect disabled and emits evenly in all directions.
func (v EmitterView) Create(position dprec.Vec3) EmitterID {
	index, emitter := allocateSlot(&v.scene.emitters, v.scene.freeEmitterIndices)

	*emitter = emitterState{
		revision:       emitter.revision + 1, // progress revision to valid (odd) value
		position:       position,
		rotation:       dprec.IdentityQuat(),
		velocity:       dprec.ZeroVec3(),
		gain:           1.0,
		doppler:        false,
		innerConeAngle: dprec.Degrees(360.0),
		outerConeAngle: dprec.Degrees(360.0),
		outerConeGain:  0.0,
	}

	return EmitterID{
		index:    index,
		revision: emitter.revision,
	}
}

// CreateHandle behaves like [EmitterView.Create] but wraps the resulting
// ID in an [EmitterHandle], as returned by [EmitterView.Handle], for
// callers that want to keep acting on the new emitter without holding
// onto its ID separately.
func (v EmitterView) CreateHandle(position dprec.Vec3) EmitterHandle {
	return v.Handle(v.Create(position))
}

// Delete removes the emitter with the specified ID and stops everything
// that is playing on it. Speech lines that get cut short this way are
// reported to the subscribers of [SpeechView.SubscribeFinished].
//
// It panics if the ID does not reference a valid emitter. Use
// [EmitterView.IsValid] first if the ID's validity is not otherwise
// guaranteed.
func (v EmitterView) Delete(id EmitterID) {
	emitter := v.resolve(id, true)

	*emitter = emitterState{
		revision: emitter.revision + 1, // progress revision to invalid (even) value
	}
	v.scene.freeEmitterIndices.Push(id.index)

	v.scene.stopEffectsOnEmitter(id)
	v.scene.stopPooledEffectsOnEmitter(id)
	v.scene.stopAmbiencesOnEmitter(id)
	v.scene.stopSpeechOnEmitter(id)
}

// Each calls cb once for every emitter that is currently alive within
// this Scene, in unspecified order.
func (v EmitterView) Each(cb func(id EmitterID)) {
	for i := range v.scene.emitters {
		if emitter := &v.scene.emitters[i]; emitter.isValid() {
			cb(EmitterID{
				index:    int32(i),
				revision: emitter.revision,
			})
		}
	}
}

// Handle returns an [EmitterHandle] that wraps the specified ID, as a more
// convenient means of repeatedly accessing the same emitter without
// having to pass its ID to this view on every call.
func (v EmitterView) Handle(id EmitterID) EmitterHandle {
	return EmitterHandle{
		view: v,
		id:   id,
	}
}

// IsValid returns whether the specified ID references an emitter that has
// not been deleted.
func (v EmitterView) IsValid(id EmitterID) bool {
	emitter := v.resolve(id, false)
	return emitter != nil
}

// Position returns the world position of the specified emitter.
//
// It panics if the ID does not reference a valid emitter.
func (v EmitterView) Position(id EmitterID) dprec.Vec3 {
	emitter := v.resolve(id, true)
	return emitter.position
}

// SetPosition changes the world position of the specified emitter.
//
// It panics if the ID does not reference a valid emitter.
func (v EmitterView) SetPosition(id EmitterID, position dprec.Vec3) {
	emitter := v.resolve(id, true)
	emitter.position = position
}

// Rotation returns the world orientation of the specified emitter.
//
// It panics if the ID does not reference a valid emitter.
func (v EmitterView) Rotation(id EmitterID) dprec.Quat {
	emitter := v.resolve(id, true)
	return emitter.rotation
}

// SetRotation changes the world orientation of the specified emitter. The
// rotation must be a unit quaternion. The orientation only matters for
// emitters that have a sound cone (see [EmitterView.SetOuterConeAngle]).
//
// It panics if the ID does not reference a valid emitter.
func (v EmitterView) SetRotation(id EmitterID, rotation dprec.Quat) {
	emitter := v.resolve(id, true)
	emitter.rotation = rotation
}

// Velocity returns the velocity of the specified emitter, in units per
// second.
//
// It panics if the ID does not reference a valid emitter.
func (v EmitterView) Velocity(id EmitterID) dprec.Vec3 {
	emitter := v.resolve(id, true)
	return emitter.velocity
}

// SetVelocity changes the velocity of the specified emitter, in units per
// second. The velocity does not move the emitter. It is only used to
// calculate the Doppler shift of the sounds playing on it.
//
// It panics if the ID does not reference a valid emitter.
func (v EmitterView) SetVelocity(id EmitterID, velocity dprec.Vec3) {
	emitter := v.resolve(id, true)
	emitter.velocity = velocity
}

// Doppler returns whether the Doppler effect is applied to the sounds
// playing on the specified emitter. It is disabled by default.
//
// It panics if the ID does not reference a valid emitter.
func (v EmitterView) Doppler(id EmitterID) bool {
	emitter := v.resolve(id, true)
	return emitter.doppler
}

// SetDoppler changes whether the Doppler effect is applied to the sounds
// playing on the specified emitter. When enabled, the pitch of those
// sounds shifts according to the velocities of the emitter and of the
// active listener along the line that connects them. Either velocity is
// enough for a shift, so the sounds of an emitter that does not move are
// shifted as well when the active listener does.
//
// It panics if the ID does not reference a valid emitter.
func (v EmitterView) SetDoppler(id EmitterID, doppler bool) {
	emitter := v.resolve(id, true)
	emitter.doppler = doppler
}

// Gain returns the gain of the specified emitter. The default is 1.
//
// It panics if the ID does not reference a valid emitter.
func (v EmitterView) Gain(id EmitterID) float64 {
	emitter := v.resolve(id, true)
	return emitter.gain
}

// SetGain changes the gain of the specified emitter, which scales the
// loudness of everything that is playing on it. Negative values are
// clamped to 0.
//
// It panics if the ID does not reference a valid emitter.
func (v EmitterView) SetGain(id EmitterID, gain float64) {
	emitter := v.resolve(id, true)
	emitter.gain = max(0.0, gain)
}

// InnerConeAngle returns the angle of the cone, centered on the
// orientation of the specified emitter, within which sounds are heard at
// full gain. The default is 360 degrees.
//
// It panics if the ID does not reference a valid emitter.
func (v EmitterView) InnerConeAngle(id EmitterID) dprec.Angle {
	emitter := v.resolve(id, true)
	return emitter.innerConeAngle
}

// SetInnerConeAngle changes the angle of the cone within which the sounds
// of the specified emitter are heard at full gain.
//
// It panics if the ID does not reference a valid emitter.
func (v EmitterView) SetInnerConeAngle(id EmitterID, angle dprec.Angle) {
	emitter := v.resolve(id, true)
	emitter.innerConeAngle = angle
}

// OuterConeAngle returns the angle of the cone, centered on the
// orientation of the specified emitter, outside of which sounds are heard
// at the outer cone gain. Between the inner and the outer cone the gain
// transitions between the two. The default is 360 degrees, which makes
// the emitter omnidirectional.
//
// It panics if the ID does not reference a valid emitter.
func (v EmitterView) OuterConeAngle(id EmitterID) dprec.Angle {
	emitter := v.resolve(id, true)
	return emitter.outerConeAngle
}

// SetOuterConeAngle changes the angle of the cone outside of which the
// sounds of the specified emitter are heard at the outer cone gain.
//
// It panics if the ID does not reference a valid emitter.
func (v EmitterView) SetOuterConeAngle(id EmitterID, angle dprec.Angle) {
	emitter := v.resolve(id, true)
	emitter.outerConeAngle = angle
}

// OuterConeGain returns the gain that is applied to the sounds of the
// specified emitter when the listener is outside of its outer cone. The
// default is 0.
//
// It panics if the ID does not reference a valid emitter.
func (v EmitterView) OuterConeGain(id EmitterID) float64 {
	emitter := v.resolve(id, true)
	return emitter.outerConeGain
}

// SetOuterConeGain changes the gain that is applied to the sounds of the
// specified emitter when the listener is outside of its outer cone.
// Negative values are clamped to 0.
//
// It panics if the ID does not reference a valid emitter.
func (v EmitterView) SetOuterConeGain(id EmitterID, gain float64) {
	emitter := v.resolve(id, true)
	emitter.outerConeGain = max(0.0, gain)
}

// resolve looks up the emitterState referenced by id. If id is stale or
// otherwise invalid, resolve panics when required is true, or returns nil
// otherwise.
func (v EmitterView) resolve(id EmitterID, required bool) *emitterState {
	if id.revision == 0 {
		if required {
			panic("invalid emitter ID")
		}
		return nil
	}
	emitter := &v.scene.emitters[id.index]
	if emitter.revision != id.revision {
		if required {
			panic("invalid emitter ID")
		}
		return nil
	}
	return emitter
}

// EmitterHandle is an object-oriented alternative to [EmitterView] that is
// bound to a specific [EmitterID].
//
// It is obtained through [EmitterView.Handle].
//
// Wrapping an ID this way needs no allocation of its own, but unlike a
// plain [EmitterID] (which holds no pointers), a Handle keeps an internal
// reference to the owning [Scene]. This keeps that Scene reachable for as
// long as the handle is retained, and adds a pointer that the garbage
// collector must trace wherever the handle is stored. Prefer storing IDs
// over handles in long-lived collections unless the convenience is worth
// that cost.
type EmitterHandle struct {
	view EmitterView
	id   EmitterID
}

// ID returns the [EmitterID] wrapped by this handle.
func (h EmitterHandle) ID() EmitterID {
	return h.id
}

// Delete removes the wrapped emitter.
//
// See [EmitterView.Delete] for further details.
func (h EmitterHandle) Delete() {
	h.view.Delete(h.id)
}

// IsValid returns whether the wrapped emitter has not been deleted.
func (h EmitterHandle) IsValid() bool {
	return h.view.IsValid(h.id)
}

// Position returns the world position of the wrapped emitter.
//
// It panics if the handle does not reference a valid emitter.
func (h EmitterHandle) Position() dprec.Vec3 {
	return h.view.Position(h.id)
}

// SetPosition changes the world position of the wrapped emitter.
//
// It panics if the handle does not reference a valid emitter.
func (h EmitterHandle) SetPosition(position dprec.Vec3) {
	h.view.SetPosition(h.id, position)
}

// Rotation returns the world orientation of the wrapped emitter.
//
// It panics if the handle does not reference a valid emitter.
func (h EmitterHandle) Rotation() dprec.Quat {
	return h.view.Rotation(h.id)
}

// SetRotation changes the world orientation of the wrapped emitter.
//
// See [EmitterView.SetRotation] for further details.
func (h EmitterHandle) SetRotation(rotation dprec.Quat) {
	h.view.SetRotation(h.id, rotation)
}

// Velocity returns the velocity of the wrapped emitter.
//
// It panics if the handle does not reference a valid emitter.
func (h EmitterHandle) Velocity() dprec.Vec3 {
	return h.view.Velocity(h.id)
}

// SetVelocity changes the velocity of the wrapped emitter.
//
// See [EmitterView.SetVelocity] for further details.
func (h EmitterHandle) SetVelocity(velocity dprec.Vec3) {
	h.view.SetVelocity(h.id, velocity)
}

// Doppler returns whether the Doppler effect is applied to the sounds
// playing on the wrapped emitter.
//
// It panics if the handle does not reference a valid emitter.
func (h EmitterHandle) Doppler() bool {
	return h.view.Doppler(h.id)
}

// SetDoppler changes whether the Doppler effect is applied to the sounds
// playing on the wrapped emitter.
//
// See [EmitterView.SetDoppler] for further details.
func (h EmitterHandle) SetDoppler(doppler bool) {
	h.view.SetDoppler(h.id, doppler)
}

// Gain returns the gain of the wrapped emitter.
//
// It panics if the handle does not reference a valid emitter.
func (h EmitterHandle) Gain() float64 {
	return h.view.Gain(h.id)
}

// SetGain changes the gain of the wrapped emitter.
//
// See [EmitterView.SetGain] for further details.
func (h EmitterHandle) SetGain(gain float64) {
	h.view.SetGain(h.id, gain)
}

// InnerConeAngle returns the inner cone angle of the wrapped emitter.
//
// See [EmitterView.InnerConeAngle] for further details.
func (h EmitterHandle) InnerConeAngle() dprec.Angle {
	return h.view.InnerConeAngle(h.id)
}

// SetInnerConeAngle changes the inner cone angle of the wrapped emitter.
//
// See [EmitterView.SetInnerConeAngle] for further details.
func (h EmitterHandle) SetInnerConeAngle(angle dprec.Angle) {
	h.view.SetInnerConeAngle(h.id, angle)
}

// OuterConeAngle returns the outer cone angle of the wrapped emitter.
//
// See [EmitterView.OuterConeAngle] for further details.
func (h EmitterHandle) OuterConeAngle() dprec.Angle {
	return h.view.OuterConeAngle(h.id)
}

// SetOuterConeAngle changes the outer cone angle of the wrapped emitter.
//
// See [EmitterView.SetOuterConeAngle] for further details.
func (h EmitterHandle) SetOuterConeAngle(angle dprec.Angle) {
	h.view.SetOuterConeAngle(h.id, angle)
}

// OuterConeGain returns the outer cone gain of the wrapped emitter.
//
// See [EmitterView.OuterConeGain] for further details.
func (h EmitterHandle) OuterConeGain() float64 {
	return h.view.OuterConeGain(h.id)
}

// SetOuterConeGain changes the outer cone gain of the wrapped emitter.
//
// See [EmitterView.SetOuterConeGain] for further details.
func (h EmitterHandle) SetOuterConeGain(gain float64) {
	h.view.SetOuterConeGain(h.id, gain)
}

// emitterState holds the internal state of a single emitter, as tracked by
// a [Scene].
type emitterState struct {
	revision       int32
	position       dprec.Vec3
	rotation       dprec.Quat
	velocity       dprec.Vec3
	gain           float64
	doppler        bool
	innerConeAngle dprec.Angle
	outerConeAngle dprec.Angle
	outerConeGain  float64
}

// isValid returns whether this state is currently backing a live emitter,
// as opposed to a freed slot awaiting reuse.
func (s *emitterState) isValid() bool {
	return s.revision%2 == 1 // only odd revisions are valid
}
