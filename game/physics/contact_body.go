package physics

import "github.com/mokiat/lacking/core/spatial/shape3d"

type BodyContact struct {
	TargetBodyID BodyID
	shape3d.Contact
}

type BodyContactCallback func(contact BodyContact)

type DeepestBodyContact struct {
	contact    BodyContact
	hasContact bool
}

func (c *DeepestBodyContact) Reset() {
	c.hasContact = false
}

func (c *DeepestBodyContact) AddContact(contact BodyContact) {
	if !c.hasContact || contact.Depth > c.contact.Depth {
		c.contact = contact
		c.hasContact = true
	}
}

func (c *DeepestBodyContact) Contact() (BodyContact, bool) {
	return c.contact, c.hasContact
}

type ShallowestBodyContact struct {
	contact    BodyContact
	hasContact bool
}

func (c *ShallowestBodyContact) Reset() {
	c.hasContact = false
}

func (c *ShallowestBodyContact) AddContact(contact BodyContact) {
	if !c.hasContact || contact.Depth < c.contact.Depth {
		c.contact = contact
		c.hasContact = true
	}
}

func (c *ShallowestBodyContact) Contact() (BodyContact, bool) {
	return c.contact, c.hasContact
}

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
