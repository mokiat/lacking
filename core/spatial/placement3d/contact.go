package placement3d

import "github.com/mokiat/lacking/core/spatial/shape3d"

// Contact describes the intersection of a source shape with a target shape.
//
// Its fields are expressed relative to the target shape. The equivalent values
// for the source shape can be derived via [shape3d.Contact.EvalSourcePoint] and
// [shape3d.Contact.EvalSourceNormal].
type Contact struct {

	// SourceShapeID contains the ID of the shape from the first involved object.
	//
	// This ID is equal to [InvalidShapeID] if the check was not performed with
	// a scene object.
	SourceShapeID ShapeID

	// TargetShapeID contains the ID of the shape from the second involved object.
	//
	// This ID is equal to [InvalidShapeID] when the target of the intersection
	// was a mesh, in which case [Contact.TargetMeshID] identifies it instead.
	TargetShapeID ShapeID

	// TargetMeshID contains the ID of the mesh that was intersected.
	//
	// This ID is equal to [InvalidMeshID] when the target of the intersection
	// was a shape rather than a mesh.
	TargetMeshID MeshID

	// Contact holds the underlying raw shape intersection.
	shape3d.Contact
}

// ContactCallback is invoked for each [Contact] discovered while testing shapes
// for intersection.
type ContactCallback func(contact Contact)

// LastContact is a contact sink that retains the most recently added [Contact].
//
// Its AddContact method satisfies [ContactCallback] and can be passed directly to
// intersection routines.
type LastContact struct {
	contact    Contact
	hasContact bool
}

// Reset clears any retained contact.
func (c *LastContact) Reset() {
	c.hasContact = false
}

// AddContact retains the given contact, replacing any previously retained one.
func (c *LastContact) AddContact(contact Contact) {
	c.contact = contact
	c.hasContact = true
}

// Contact returns the retained contact and whether one was added since the last
// Reset.
func (c *LastContact) Contact() (Contact, bool) {
	return c.contact, c.hasContact
}

// DeepestContact is a contact sink that retains the added [Contact] with the
// greatest Depth.
//
// Its AddContact method satisfies [ContactCallback] and can be passed directly to
// intersection routines.
type DeepestContact struct {
	contact    Contact
	hasContact bool
}

// Reset clears any retained contact.
func (c *DeepestContact) Reset() {
	c.hasContact = false
}

// AddContact retains the given contact if it is deeper than any previously
// retained one.
func (c *DeepestContact) AddContact(contact Contact) {
	if !c.hasContact || contact.Depth > c.contact.Depth {
		c.contact = contact
		c.hasContact = true
	}
}

// Contact returns the deepest retained contact and whether one was added since
// the last Reset.
func (c *DeepestContact) Contact() (Contact, bool) {
	return c.contact, c.hasContact
}

// ShallowestContact is a contact sink that retains the added [Contact] with the
// smallest Depth.
//
// Its AddContact method satisfies [ContactCallback] and can be passed directly to
// intersection routines.
type ShallowestContact struct {
	contact    Contact
	hasContact bool
}

// Reset clears any retained contact.
func (c *ShallowestContact) Reset() {
	c.hasContact = false
}

// AddContact retains the given contact if it is shallower than any previously
// retained one.
func (c *ShallowestContact) AddContact(contact Contact) {
	if !c.hasContact || contact.Depth < c.contact.Depth {
		c.contact = contact
		c.hasContact = true
	}
}

// Contact returns the shallowest retained contact and whether one was added
// since the last Reset.
func (c *ShallowestContact) Contact() (Contact, bool) {
	return c.contact, c.hasContact
}

// ContactList is a contact sink that retains every added [Contact] in the order
// it was added.
//
// Its AddContact method satisfies [ContactCallback] and can be passed directly to
// intersection routines. As it is itself a slice, the retained contacts can be
// ranged over directly.
//
// Use make(ContactList, 0, n) to pre-size it and avoid reallocations as
// contacts are added. With a constant n that does not escape, the compiler can
// keep the backing array on the stack.
type ContactList []Contact

// Reset clears the retained contacts while preserving the underlying capacity
// so it can be reused without reallocating.
func (l *ContactList) Reset() {
	*l = (*l)[:0]
}

// AddContact appends the given contact to the list.
func (l *ContactList) AddContact(contact Contact) {
	*l = append(*l, contact)
}

// Contacts returns the retained contacts in the order they were added.
//
// The result aliases the internal storage and remains valid until the next
// AddContact or Reset call.
func (l ContactList) Contacts() []Contact {
	return l
}
