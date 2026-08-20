package physics

import "github.com/mokiat/lacking/core/spatial/shape3d"

type TerrainContact struct {
	TargetTerrainID TerrainID
	shape3d.Contact
}

type TerrainContactCallback func(contact TerrainContact)

type DeepestTerrainContact struct {
	contact    TerrainContact
	hasContact bool
}

func (c *DeepestTerrainContact) Reset() {
	c.hasContact = false
}

func (c *DeepestTerrainContact) AddContact(contact TerrainContact) {
	if !c.hasContact || contact.Depth > c.contact.Depth {
		c.contact = contact
		c.hasContact = true
	}
}

func (c *DeepestTerrainContact) Contact() (TerrainContact, bool) {
	return c.contact, c.hasContact
}

type ShallowestTerrainContact struct {
	contact    TerrainContact
	hasContact bool
}

func (c *ShallowestTerrainContact) Reset() {
	c.hasContact = false
}

func (c *ShallowestTerrainContact) AddContact(contact TerrainContact) {
	if !c.hasContact || contact.Depth < c.contact.Depth {
		c.contact = contact
		c.hasContact = true
	}
}

func (c *ShallowestTerrainContact) Contact() (TerrainContact, bool) {
	return c.contact, c.hasContact
}

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
