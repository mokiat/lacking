package placement3d

import "github.com/mokiat/lacking/core/spatial/shape3d"

// TerrainContact describes the intersection of a source shape with a terrain
// shape.
//
// Its fields are expressed relative to the target shape. The equivalent values
// for the source shape can be derived via [shape3d.Contact.EvalSourcePoint] and
// [shape3d.Contact.EvalSourceNormal].
type TerrainContact struct {

	// SourceObjectID contains the ID of the object that owns the source shape.
	//
	// This ID is equal to [NilObjectID] when the intersection was produced by a
	// query primitive rather than by a shape in the scene.
	SourceObjectID ObjectID

	// SourceShapeID contains the ID of the shape that acted as the source of
	// the intersection.
	//
	// This ID is equal to [NilObjectShapeID] when the intersection was produced
	// by a query primitive rather than by a shape in the scene.
	//
	// The source of a terrain contact is always an object shape, since terrain
	// shapes are never tested against one another.
	SourceShapeID ObjectShapeID

	// TargetTerrainID contains the ID of the terrain that owns the target
	// shape.
	TargetTerrainID TerrainID

	// TargetShapeID contains the ID of the shape that was intersected.
	TargetShapeID TerrainShapeID

	// Contact holds the underlying raw shape intersection.
	shape3d.Contact
}

// TerrainContactCallback is invoked for each [TerrainContact] discovered while
// testing shapes for intersection.
type TerrainContactCallback func(contact TerrainContact)

// DeepestTerrainContact is a contact sink that retains the added
// [TerrainContact] with the greatest Depth.
//
// Its AddContact method satisfies [TerrainContactCallback] and can be passed
// directly to intersection routines.
type DeepestTerrainContact struct {
	contact    TerrainContact
	hasContact bool
}

// Reset clears any retained contact.
func (c *DeepestTerrainContact) Reset() {
	c.hasContact = false
}

// AddContact retains the given contact if it is deeper than any previously
// retained one.
func (c *DeepestTerrainContact) AddContact(contact TerrainContact) {
	if !c.hasContact || contact.Depth > c.contact.Depth {
		c.contact = contact
		c.hasContact = true
	}
}

// Contact returns the deepest retained contact and whether one was added since
// the last Reset.
func (c *DeepestTerrainContact) Contact() (TerrainContact, bool) {
	return c.contact, c.hasContact
}

// ShallowestTerrainContact is a contact sink that retains the added
// [TerrainContact] with the smallest Depth.
//
// Its AddContact method satisfies [TerrainContactCallback] and can be passed
// directly to intersection routines.
type ShallowestTerrainContact struct {
	contact    TerrainContact
	hasContact bool
}

// Reset clears any retained contact.
func (c *ShallowestTerrainContact) Reset() {
	c.hasContact = false
}

// AddContact retains the given contact if it is shallower than any previously
// retained one.
func (c *ShallowestTerrainContact) AddContact(contact TerrainContact) {
	if !c.hasContact || contact.Depth < c.contact.Depth {
		c.contact = contact
		c.hasContact = true
	}
}

// Contact returns the shallowest retained contact and whether one was added
// since the last Reset.
func (c *ShallowestTerrainContact) Contact() (TerrainContact, bool) {
	return c.contact, c.hasContact
}

// TerrainContactList is a contact sink that retains every added
// [TerrainContact] in the order it was added.
//
// Its AddContact method satisfies [TerrainContactCallback] and can be passed
// directly to intersection routines. As it is itself a slice, the retained
// contacts can be ranged over directly.
//
// Use make(TerrainContactList, 0, n) to pre-size it and avoid reallocations as
// contacts are added. With a constant n that does not escape, the compiler can
// keep the backing array on the stack.
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

// LastTerrainContact is a contact sink that retains the most recently added
// [TerrainContact].
//
// Its AddContact method satisfies [TerrainContactCallback] and can be passed
// directly to intersection routines.
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
