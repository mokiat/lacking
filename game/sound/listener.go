package sound

import "github.com/mokiat/gomath/dprec"

// ListenerID uniquely identifies a listener that was created through
// [ListenerView.Create].
//
// The zero value, also available as [NilListenerID], does not reference a
// valid listener.
type ListenerID struct {
	index    int32
	revision int32
}

// NilListenerID is a [ListenerID] that never references a valid listener.
var NilListenerID = ListenerID{}

// ListenerView provides access to the listeners that belong to a [Scene].
//
// A listener is a point in space, with an orientation and a velocity, from
// which the spatial sounds of the scene are heard. Only the active listener
// of the scene (see [Scene.SetActiveListener]) is used.
//
// A ListenerView is a lightweight accessor around a [Scene] and can be
// obtained through [Scene.Listeners].
type ListenerView struct {
	scene *Scene
}

// Create allocates a new listener at the specified world position and
// rotation, and returns its ID. The rotation must be a unit quaternion.
//
// The new listener has zero velocity. If the scene has no active listener,
// the new listener becomes the active one.
func (v ListenerView) Create(position dprec.Vec3, rotation dprec.Quat) ListenerID {
	index, listener := allocateSlot(&v.scene.listeners, v.scene.freeListenerIndices)

	*listener = listenerState{
		revision: listener.revision + 1, // progress revision to valid (odd) value
		position: position,
		rotation: rotation,
		velocity: dprec.ZeroVec3(),
	}

	id := ListenerID{
		index:    index,
		revision: listener.revision,
	}
	if v.scene.activeListenerID == NilListenerID {
		v.scene.activeListenerID = id
	}
	return id
}

// CreateHandle behaves like [ListenerView.Create] but wraps the resulting
// ID in a [ListenerHandle], as returned by [ListenerView.Handle], for
// callers that want to keep acting on the new listener without holding
// onto its ID separately.
func (v ListenerView) CreateHandle(position dprec.Vec3, rotation dprec.Quat) ListenerHandle {
	return v.Handle(v.Create(position, rotation))
}

// Delete removes the listener with the specified ID. If it is the active
// listener, the scene is left without an active listener.
//
// It panics if the ID does not reference a valid listener. Use
// [ListenerView.IsValid] first if the ID's validity is not otherwise
// guaranteed.
func (v ListenerView) Delete(id ListenerID) {
	listener := v.resolve(id, true)

	*listener = listenerState{
		revision: listener.revision + 1, // progress revision to invalid (even) value
	}
	if v.scene.activeListenerID == id {
		v.scene.activeListenerID = NilListenerID
	}

	v.scene.freeListenerIndices.Push(id.index)
}

// Each calls cb once for every listener that is currently alive within
// this Scene, in unspecified order.
func (v ListenerView) Each(cb func(id ListenerID)) {
	for i := range v.scene.listeners {
		if listener := &v.scene.listeners[i]; listener.isValid() {
			cb(ListenerID{
				index:    int32(i),
				revision: listener.revision,
			})
		}
	}
}

// Handle returns a [ListenerHandle] that wraps the specified ID, as a more
// convenient means of repeatedly accessing the same listener without
// having to pass its ID to this view on every call.
func (v ListenerView) Handle(id ListenerID) ListenerHandle {
	return ListenerHandle{
		view: v,
		id:   id,
	}
}

// IsValid returns whether the specified ID references a listener that has
// not been deleted.
func (v ListenerView) IsValid(id ListenerID) bool {
	listener := v.resolve(id, false)
	return listener != nil
}

// Position returns the world position of the specified listener.
//
// It panics if the ID does not reference a valid listener.
func (v ListenerView) Position(id ListenerID) dprec.Vec3 {
	listener := v.resolve(id, true)
	return listener.position
}

// SetPosition changes the world position of the specified listener.
//
// It panics if the ID does not reference a valid listener.
func (v ListenerView) SetPosition(id ListenerID, position dprec.Vec3) {
	listener := v.resolve(id, true)
	listener.position = position
}

// Rotation returns the world orientation of the specified listener.
//
// It panics if the ID does not reference a valid listener.
func (v ListenerView) Rotation(id ListenerID) dprec.Quat {
	listener := v.resolve(id, true)
	return listener.rotation
}

// SetRotation changes the world orientation of the specified listener. The
// rotation must be a unit quaternion.
//
// It panics if the ID does not reference a valid listener.
func (v ListenerView) SetRotation(id ListenerID, rotation dprec.Quat) {
	listener := v.resolve(id, true)
	listener.rotation = rotation
}

// Velocity returns the velocity of the specified listener, in units per
// second.
//
// It panics if the ID does not reference a valid listener.
func (v ListenerView) Velocity(id ListenerID) dprec.Vec3 {
	listener := v.resolve(id, true)
	return listener.velocity
}

// SetVelocity changes the velocity of the specified listener, in units per
// second. The velocity does not move the listener. It is only used to
// calculate the Doppler shift of sounds on emitters that have the effect
// enabled (see [EmitterView.SetDoppler]).
//
// It panics if the ID does not reference a valid listener.
func (v ListenerView) SetVelocity(id ListenerID, velocity dprec.Vec3) {
	listener := v.resolve(id, true)
	listener.velocity = velocity
}

// resolve looks up the listenerState referenced by id. If id is stale or
// otherwise invalid, resolve panics when required is true, or returns nil
// otherwise.
func (v ListenerView) resolve(id ListenerID, required bool) *listenerState {
	if id.revision == 0 {
		if required {
			panic("invalid listener ID")
		}
		return nil
	}
	listener := &v.scene.listeners[id.index]
	if listener.revision != id.revision {
		if required {
			panic("invalid listener ID")
		}
		return nil
	}
	return listener
}

// ListenerHandle is an object-oriented alternative to [ListenerView] that
// is bound to a specific [ListenerID].
//
// It is obtained through [ListenerView.Handle].
//
// Wrapping an ID this way needs no allocation of its own, but unlike a
// plain [ListenerID] (which holds no pointers), a Handle keeps an internal
// reference to the owning [Scene]. This keeps that Scene reachable for as
// long as the handle is retained, and adds a pointer that the garbage
// collector must trace wherever the handle is stored. Prefer storing IDs
// over handles in long-lived collections unless the convenience is worth
// that cost.
type ListenerHandle struct {
	view ListenerView
	id   ListenerID
}

// ID returns the [ListenerID] wrapped by this handle.
func (h ListenerHandle) ID() ListenerID {
	return h.id
}

// Delete removes the wrapped listener.
//
// See [ListenerView.Delete] for further details.
func (h ListenerHandle) Delete() {
	h.view.Delete(h.id)
}

// IsValid returns whether the wrapped listener has not been deleted.
func (h ListenerHandle) IsValid() bool {
	return h.view.IsValid(h.id)
}

// Position returns the world position of the wrapped listener.
//
// It panics if the handle does not reference a valid listener.
func (h ListenerHandle) Position() dprec.Vec3 {
	return h.view.Position(h.id)
}

// SetPosition changes the world position of the wrapped listener.
//
// It panics if the handle does not reference a valid listener.
func (h ListenerHandle) SetPosition(position dprec.Vec3) {
	h.view.SetPosition(h.id, position)
}

// Rotation returns the world orientation of the wrapped listener.
//
// It panics if the handle does not reference a valid listener.
func (h ListenerHandle) Rotation() dprec.Quat {
	return h.view.Rotation(h.id)
}

// SetRotation changes the world orientation of the wrapped listener.
//
// See [ListenerView.SetRotation] for further details.
func (h ListenerHandle) SetRotation(rotation dprec.Quat) {
	h.view.SetRotation(h.id, rotation)
}

// Velocity returns the velocity of the wrapped listener.
//
// It panics if the handle does not reference a valid listener.
func (h ListenerHandle) Velocity() dprec.Vec3 {
	return h.view.Velocity(h.id)
}

// SetVelocity changes the velocity of the wrapped listener.
//
// See [ListenerView.SetVelocity] for further details.
func (h ListenerHandle) SetVelocity(velocity dprec.Vec3) {
	h.view.SetVelocity(h.id, velocity)
}

// listenerState holds the internal state of a single listener, as tracked
// by a [Scene].
type listenerState struct {
	revision int32
	position dprec.Vec3
	rotation dprec.Quat
	velocity dprec.Vec3
}

// isValid returns whether this state is currently backing a live listener,
// as opposed to a freed slot awaiting reuse.
func (s *listenerState) isValid() bool {
	return s.revision%2 == 1 // only odd revisions are valid
}
