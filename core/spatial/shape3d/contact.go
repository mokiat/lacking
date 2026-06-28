package shape3d

import "github.com/mokiat/gomath/dprec"

// Contact describes the intersection of a source shape with a target shape.
//
// Its fields are expressed relative to the target shape. The equivalent values
// for the source shape can be derived via EvalSourcePoint and EvalSourceNormal.
type Contact struct {

	// TargetPoint is the contact point on the surface of the target shape.
	TargetPoint dprec.Vec3

	// TargetNormal is the outward-facing surface normal of the target shape at
	// TargetPoint. It points away from the target and toward the source, and is
	// the direction along which the source shape must be moved by Depth to
	// resolve the intersection.
	TargetNormal dprec.Vec3

	// Depth is the penetration distance between the two shapes measured along
	// TargetNormal. It is always non-negative.
	//
	// Some producers report Depth in a different unit when a distance is not the
	// most useful measure. The segment resolves in the isec3d package, for
	// example, report it as the fraction of the segment lying beyond the contact
	// point (a unitless value in the range [0, 1]) so that contacts against
	// different shapes can be compared. For such contacts the distance-based
	// helpers EvalSourcePoint and Flipped are not meaningful; see their docs.
	Depth float64
}

// EvalSourcePoint returns the contact point on the surface of the source shape.
//
// It lies a distance of Depth from TargetPoint along the inverse of
// TargetNormal. This is only meaningful when Depth is a distance along
// TargetNormal; it does not apply to contacts whose Depth carries a different
// unit, such as those produced by the segment resolves in the isec3d package.
func (c Contact) EvalSourcePoint() dprec.Vec3 {
	return dprec.Vec3Diff(c.TargetPoint, dprec.Vec3Prod(c.TargetNormal, c.Depth))
}

// EvalSourceNormal returns the outward-facing surface normal of the source
// shape at its contact point.
//
// It is the inverse of TargetNormal and points in the direction along which the
// target shape must be moved by Depth to resolve the intersection.
func (c Contact) EvalSourceNormal() dprec.Vec3 {
	return dprec.InverseVec3(c.TargetNormal)
}

// Flipped returns a Contact with the source and target shapes swapped.
//
// The resulting contact describes the same intersection from the perspective of
// the opposite shape. As it relies on EvalSourcePoint, it is only meaningful
// when Depth is a distance along TargetNormal, and not for contacts whose Depth
// carries a different unit, such as those produced by the segment resolves in
// the isec3d package.
func (c Contact) Flipped() Contact {
	return Contact{
		TargetPoint:  c.EvalSourcePoint(),
		TargetNormal: c.EvalSourceNormal(),
		Depth:        c.Depth,
	}
}

// ContactCallback is invoked for each Contact discovered while testing shapes
// for intersection.
type ContactCallback func(contact Contact)

// LastContact is a contact sink that retains the most recently added Contact.
//
// Its AddContact method satisfies ContactCallback and can be passed directly to
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

// DeepestContact is a contact sink that retains the added Contact with the
// greatest Depth.
//
// Its AddContact method satisfies ContactCallback and can be passed directly to
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

// ShallowestContact is a contact sink that retains the added Contact with the
// smallest Depth.
//
// Its AddContact method satisfies ContactCallback and can be passed directly to
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

// ContactList is a contact sink that retains every added Contact in the order
// it was added.
//
// Its AddContact method satisfies ContactCallback and can be passed directly to
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
