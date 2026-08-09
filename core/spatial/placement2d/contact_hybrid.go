package placement2d

// DeepestContact is a contact sink that retains the added [ObjectContact] and
// [TerrainContact] with the greatest Depth across both kinds.
//
// Its AddObjectContact method satisfies [ObjectContactCallback] and its
// AddTerrainContact method satisfies [TerrainContactCallback], so both can be
// passed directly to intersection routines. This makes it possible to pick the
// single deepest intersection without having to know in advance whether it will
// be with an object shape or with a terrain shape.
//
// At most one of [DeepestContact.ObjectContact] and
// [DeepestContact.TerrainContact] returns a contact. Ties in Depth are resolved
// in favor of the object contact.
type DeepestContact struct {
	objectSink  DeepestObjectContact
	terrainSink DeepestTerrainContact
}

// Reset clears any retained contacts.
func (c *DeepestContact) Reset() {
	c.objectSink.Reset()
	c.terrainSink.Reset()
}

// AddObjectContact retains the given contact if it is deeper than any
// previously retained object contact.
func (c *DeepestContact) AddObjectContact(contact ObjectContact) {
	c.objectSink.AddContact(contact)
}

// AddTerrainContact retains the given contact if it is deeper than any
// previously retained terrain contact.
func (c *DeepestContact) AddTerrainContact(contact TerrainContact) {
	c.terrainSink.AddContact(contact)
}

// ObjectContact returns the deepest retained object contact and whether it is
// the deepest contact overall.
//
// The result is false when no object contact was added since the last Reset, as
// well as when a strictly deeper terrain contact was added, in which case
// [DeepestContact.TerrainContact] holds the deepest contact instead.
func (c *DeepestContact) ObjectContact() (ObjectContact, bool) {
	objectContact, ok := c.objectSink.Contact()
	if !ok {
		return ObjectContact{}, false
	}
	terrainContact, ok := c.terrainSink.Contact()
	if !ok {
		return objectContact, true
	}
	if objectContact.Depth < terrainContact.Depth {
		return ObjectContact{}, false
	}
	return objectContact, true
}

// TerrainContact returns the deepest retained terrain contact and whether it is
// the deepest contact overall.
//
// The result is false when no terrain contact was added since the last Reset,
// as well as when an equally deep or deeper object contact was added, in which
// case [DeepestContact.ObjectContact] holds the deepest contact instead.
func (c *DeepestContact) TerrainContact() (TerrainContact, bool) {
	terrainContact, ok := c.terrainSink.Contact()
	if !ok {
		return TerrainContact{}, false
	}
	objectContact, ok := c.objectSink.Contact()
	if !ok {
		return terrainContact, true
	}
	if terrainContact.Depth <= objectContact.Depth {
		return TerrainContact{}, false
	}
	return terrainContact, true
}

// ShallowestContact is a contact sink that retains the added [ObjectContact]
// and [TerrainContact] with the smallest Depth across both kinds.
//
// Its AddObjectContact method satisfies [ObjectContactCallback] and its
// AddTerrainContact method satisfies [TerrainContactCallback], so both can be
// passed directly to intersection routines. This makes it possible to pick the
// single shallowest intersection without having to know in advance whether it
// will be with an object shape or with a terrain shape.
//
// At most one of [ShallowestContact.ObjectContact] and
// [ShallowestContact.TerrainContact] returns a contact. Ties in Depth are
// resolved in favor of the object contact.
type ShallowestContact struct {
	objectSink  ShallowestObjectContact
	terrainSink ShallowestTerrainContact
}

// Reset clears any retained contacts.
func (c *ShallowestContact) Reset() {
	c.objectSink.Reset()
	c.terrainSink.Reset()
}

// AddObjectContact retains the given contact if it is shallower than any
// previously retained object contact.
func (c *ShallowestContact) AddObjectContact(contact ObjectContact) {
	c.objectSink.AddContact(contact)
}

// AddTerrainContact retains the given contact if it is shallower than any
// previously retained terrain contact.
func (c *ShallowestContact) AddTerrainContact(contact TerrainContact) {
	c.terrainSink.AddContact(contact)
}

// ObjectContact returns the shallowest retained object contact and whether it
// is the shallowest contact overall.
//
// The result is false when no object contact was added since the last Reset, as
// well as when a strictly shallower terrain contact was added, in which case
// [ShallowestContact.TerrainContact] holds the shallowest contact instead.
func (c *ShallowestContact) ObjectContact() (ObjectContact, bool) {
	objectContact, ok := c.objectSink.Contact()
	if !ok {
		return ObjectContact{}, false
	}
	terrainContact, ok := c.terrainSink.Contact()
	if !ok {
		return objectContact, true
	}
	if objectContact.Depth > terrainContact.Depth {
		return ObjectContact{}, false
	}
	return objectContact, true
}

// TerrainContact returns the shallowest retained terrain contact and whether it
// is the shallowest contact overall.
//
// The result is false when no terrain contact was added since the last Reset,
// as well as when an equally shallow or shallower object contact was added, in
// which case [ShallowestContact.ObjectContact] holds the shallowest contact
// instead.
func (c *ShallowestContact) TerrainContact() (TerrainContact, bool) {
	terrainContact, ok := c.terrainSink.Contact()
	if !ok {
		return TerrainContact{}, false
	}
	objectContact, ok := c.objectSink.Contact()
	if !ok {
		return terrainContact, true
	}
	if terrainContact.Depth >= objectContact.Depth {
		return TerrainContact{}, false
	}
	return terrainContact, true
}
