package physics

import "github.com/mokiat/lacking/core/spatial/shape3d"

// TerrainContact describes an intersection between a query shape, such as
// a segment, and a collision shape of a terrain.
type TerrainContact struct {

	// TargetTerrainID is the ID of the terrain whose collision shape was
	// intersected.
	TargetTerrainID TerrainID

	// Contact holds the contact geometry, expressed relative to the
	// terrain's collision shape. For segment queries, its Depth is the
	// fraction of the segment that lies beyond the contact point, as
	// described by [shape3d.Contact].
	shape3d.Contact
}

// TerrainContactCallback is a function that receives the [TerrainContact]
// values produced by a query, such as
// [Scene.CollectSegmentTerrainIntersections].
//
// The AddContact methods of [DeepestTerrainContact],
// [ShallowestTerrainContact], [LastTerrainContact] and
// [TerrainContactList] can be used as such a callback.
type TerrainContactCallback func(contact TerrainContact)

// DeepestTerrainContact is a collector of [TerrainContact] values that
// retains only the one with the largest Depth.
//
// For segment queries, where Depth is the fraction of the segment that lies
// beyond the contact point, this is the contact closest to the start of
// the segment.
//
// The zero value is ready to use and holds no contact.
type DeepestTerrainContact struct {
	contact    TerrainContact
	hasContact bool
}

// Reset clears any retained contact.
func (c *DeepestTerrainContact) Reset() {
	c.hasContact = false
}

// AddContact retains the given contact if it is deeper than the currently
// retained one, or if no contact has been retained since the last Reset.
func (c *DeepestTerrainContact) AddContact(contact TerrainContact) {
	if !c.hasContact || contact.Depth > c.contact.Depth {
		c.contact = contact
		c.hasContact = true
	}
}

// Contact returns the retained contact and whether one was added since the
// last Reset.
func (c *DeepestTerrainContact) Contact() (TerrainContact, bool) {
	return c.contact, c.hasContact
}

// ShallowestTerrainContact is a collector of [TerrainContact] values that
// retains only the one with the smallest Depth.
//
// For segment queries, where Depth is the fraction of the segment that lies
// beyond the contact point, this is the contact farthest from the start of
// the segment.
//
// The zero value is ready to use and holds no contact.
type ShallowestTerrainContact struct {
	contact    TerrainContact
	hasContact bool
}

// Reset clears any retained contact.
func (c *ShallowestTerrainContact) Reset() {
	c.hasContact = false
}

// AddContact retains the given contact if it is shallower than the
// currently retained one, or if no contact has been retained since the
// last Reset.
func (c *ShallowestTerrainContact) AddContact(contact TerrainContact) {
	if !c.hasContact || contact.Depth < c.contact.Depth {
		c.contact = contact
		c.hasContact = true
	}
}

// Contact returns the retained contact and whether one was added since the
// last Reset.
func (c *ShallowestTerrainContact) Contact() (TerrainContact, bool) {
	return c.contact, c.hasContact
}

// TerrainContactList is a collector of [TerrainContact] values that
// retains all of them, in the order in which they were added.
//
// The zero value is ready to use and holds no contacts.
type TerrainContactList []TerrainContact

// Reset clears the retained contacts while preserving the underlying capacity
// so it can be reused without reallocating.
func (l *TerrainContactList) Reset() {
	*l = (*l)[:0]
}

// AddContact appends the given contact to the list.
func (l *TerrainContactList) AddContact(contact TerrainContact) {
	*l = append(*l, contact)
}

// Contacts returns the retained contacts in the order they were added.
//
// The result aliases the internal storage and remains valid until the next
// AddContact or Reset call.
func (l TerrainContactList) Contacts() []TerrainContact {
	return l
}

// LastTerrainContact is a collector of [TerrainContact] values that retains
// only the one that was added most recently.
//
// The zero value is ready to use and holds no contact.
type LastTerrainContact struct {
	contact    TerrainContact
	hasContact bool
}

// Reset clears any retained contact.
func (c *LastTerrainContact) Reset() {
	c.hasContact = false
}

// AddContact retains the given contact, replacing any previously retained one.
func (c *LastTerrainContact) AddContact(contact TerrainContact) {
	c.contact = contact
	c.hasContact = true
}

// Contact returns the retained contact and whether one was added since the
// last Reset.
func (c *LastTerrainContact) Contact() (TerrainContact, bool) {
	return c.contact, c.hasContact
}
