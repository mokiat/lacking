package placement3d

import "github.com/mokiat/lacking/core/spatial/shape3d"

// ObjectContact describes the intersection of a source shape with an object
// shape.
//
// Its fields are expressed relative to the target shape. The equivalent values
// for the source shape can be derived via [shape3d.Contact.EvalSourcePoint] and
// [shape3d.Contact.EvalSourceNormal].
type ObjectContact struct {

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
	SourceShapeID ObjectShapeID

	// TargetObjectID contains the ID of the object that owns the target shape.
	TargetObjectID ObjectID

	// TargetShapeID contains the ID of the shape that was intersected.
	TargetShapeID ObjectShapeID

	// Contact holds the underlying raw shape intersection.
	shape3d.Contact
}

// ObjectContactCallback is invoked for each [ObjectContact] discovered while
// testing shapes for intersection.
type ObjectContactCallback func(contact ObjectContact)

// DeepestObjectContact is a contact sink that retains the added
// [ObjectContact] with the greatest Depth.
//
// Its AddContact method satisfies [ObjectContactCallback] and can be passed
// directly to intersection routines.
type DeepestObjectContact struct {
	contact    ObjectContact
	hasContact bool
}

// Reset clears any retained contact.
func (c *DeepestObjectContact) Reset() {
	c.hasContact = false
}

// AddContact retains the given contact if it is deeper than any previously
// retained one.
func (c *DeepestObjectContact) AddContact(contact ObjectContact) {
	if !c.hasContact || contact.Depth > c.contact.Depth {
		c.contact = contact
		c.hasContact = true
	}
}

// Contact returns the deepest retained contact and whether one was added since
// the last Reset.
func (c *DeepestObjectContact) Contact() (ObjectContact, bool) {
	return c.contact, c.hasContact
}

// ShallowestObjectContact is a contact sink that retains the added
// [ObjectContact] with the smallest Depth.
//
// Its AddContact method satisfies [ObjectContactCallback] and can be passed
// directly to intersection routines.
type ShallowestObjectContact struct {
	contact    ObjectContact
	hasContact bool
}

// Reset clears any retained contact.
func (c *ShallowestObjectContact) Reset() {
	c.hasContact = false
}

// AddContact retains the given contact if it is shallower than any previously
// retained one.
func (c *ShallowestObjectContact) AddContact(contact ObjectContact) {
	if !c.hasContact || contact.Depth < c.contact.Depth {
		c.contact = contact
		c.hasContact = true
	}
}

// Contact returns the shallowest retained contact and whether one was added
// since the last Reset.
func (c *ShallowestObjectContact) Contact() (ObjectContact, bool) {
	return c.contact, c.hasContact
}

// ObjectContactList is a contact sink that retains every added [ObjectContact]
// in the order it was added.
//
// Its AddContact method satisfies [ObjectContactCallback] and can be passed
// directly to intersection routines. As it is itself a slice, the retained
// contacts can be ranged over directly.
//
// Use make(ObjectContactList, 0, n) to pre-size it and avoid reallocations as
// contacts are added. With a constant n that does not escape, the compiler can
// keep the backing array on the stack.
type ObjectContactList []ObjectContact

// Reset clears the retained contacts while preserving the underlying capacity
// so it can be reused without reallocating.
func (l *ObjectContactList) Reset() {
	*l = (*l)[:0]
}

// AddContact appends the given contact to the list.
func (l *ObjectContactList) AddContact(contact ObjectContact) {
	*l = append(*l, contact)
}

// Contacts returns the retained contacts in the order they were added.
//
// The result aliases the internal storage and remains valid until the next
// AddContact or Reset call.
func (l ObjectContactList) Contacts() []ObjectContact {
	return l
}

// LastObjectContact is a contact sink that retains the most recently added
// [ObjectContact].
//
// Its AddContact method satisfies [ObjectContactCallback] and can be passed
// directly to intersection routines.
type LastObjectContact struct {
	contact    ObjectContact
	hasContact bool
}

// Reset clears any retained contact.
func (c *LastObjectContact) Reset() {
	c.hasContact = false
}

// AddContact retains the given contact, replacing any previously retained one.
func (c *LastObjectContact) AddContact(contact ObjectContact) {
	c.contact = contact
	c.hasContact = true
}

// Contact returns the retained contact and whether one was added since the
// last Reset.
func (c *LastObjectContact) Contact() (ObjectContact, bool) {
	return c.contact, c.hasContact
}
