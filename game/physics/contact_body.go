package physics

import "github.com/mokiat/lacking/core/spatial/shape3d"

// BodyContact describes an intersection between a query shape, such as a
// segment, and a collision shape of a body.
type BodyContact struct {

	// TargetBodyID is the ID of the body whose collision shape was
	// intersected.
	TargetBodyID BodyID

	// Contact holds the contact geometry, expressed relative to the body's
	// collision shape. For segment queries, its Depth is the fraction of
	// the segment that lies beyond the contact point, as described by
	// [shape3d.Contact].
	shape3d.Contact
}

// BodyContactCallback is a function that receives the [BodyContact] values
// produced by a query, such as [Scene.CollectSegmentBodyIntersections].
//
// The AddContact methods of [DeepestBodyContact], [ShallowestBodyContact],
// [LastBodyContact] and [BodyContactList] can be used as such a callback.
type BodyContactCallback func(contact BodyContact)

// DeepestBodyContact is a collector of [BodyContact] values that retains
// only the one with the largest Depth.
//
// For segment queries, where Depth is the fraction of the segment that lies
// beyond the contact point, this is the contact closest to the start of
// the segment.
//
// The zero value is ready to use and holds no contact.
type DeepestBodyContact struct {
	contact    BodyContact
	hasContact bool
}

// Reset clears any retained contact.
func (c *DeepestBodyContact) Reset() {
	c.hasContact = false
}

// AddContact retains the given contact if it is deeper than the currently
// retained one, or if no contact has been retained since the last Reset.
func (c *DeepestBodyContact) AddContact(contact BodyContact) {
	if !c.hasContact || contact.Depth > c.contact.Depth {
		c.contact = contact
		c.hasContact = true
	}
}

// Contact returns the retained contact and whether one was added since the
// last Reset.
func (c *DeepestBodyContact) Contact() (BodyContact, bool) {
	return c.contact, c.hasContact
}

// ShallowestBodyContact is a collector of [BodyContact] values that retains
// only the one with the smallest Depth.
//
// For segment queries, where Depth is the fraction of the segment that lies
// beyond the contact point, this is the contact farthest from the start of
// the segment.
//
// The zero value is ready to use and holds no contact.
type ShallowestBodyContact struct {
	contact    BodyContact
	hasContact bool
}

// Reset clears any retained contact.
func (c *ShallowestBodyContact) Reset() {
	c.hasContact = false
}

// AddContact retains the given contact if it is shallower than the
// currently retained one, or if no contact has been retained since the
// last Reset.
func (c *ShallowestBodyContact) AddContact(contact BodyContact) {
	if !c.hasContact || contact.Depth < c.contact.Depth {
		c.contact = contact
		c.hasContact = true
	}
}

// Contact returns the retained contact and whether one was added since the
// last Reset.
func (c *ShallowestBodyContact) Contact() (BodyContact, bool) {
	return c.contact, c.hasContact
}

// BodyContactList is a collector of [BodyContact] values that retains all
// of them, in the order in which they were added.
//
// The zero value is ready to use and holds no contacts.
type BodyContactList []BodyContact

// Reset clears the retained contacts while preserving the underlying capacity
// so it can be reused without reallocating.
func (l *BodyContactList) Reset() {
	*l = (*l)[:0]
}

// AddContact appends the given contact to the list.
func (l *BodyContactList) AddContact(contact BodyContact) {
	*l = append(*l, contact)
}

// Contacts returns the retained contacts in the order they were added.
//
// The result aliases the internal storage and remains valid until the next
// AddContact or Reset call.
func (l BodyContactList) Contacts() []BodyContact {
	return l
}

// LastBodyContact is a collector of [BodyContact] values that retains only
// the one that was added most recently.
//
// The zero value is ready to use and holds no contact.
type LastBodyContact struct {
	contact    BodyContact
	hasContact bool
}

// Reset clears any retained contact.
func (c *LastBodyContact) Reset() {
	c.hasContact = false
}

// AddContact retains the given contact, replacing any previously retained one.
func (c *LastBodyContact) AddContact(contact BodyContact) {
	c.contact = contact
	c.hasContact = true
}

// Contact returns the retained contact and whether one was added since the
// last Reset.
func (c *LastBodyContact) Contact() (BodyContact, bool) {
	return c.contact, c.hasContact
}
