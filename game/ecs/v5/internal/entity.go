package internal

type Entity struct {
	index        uint32
	revision     uint32
	archetype    *Archetype
	archetypeRow Row
}

// Revive revives the entity with the specified archetype and archetype row.
// This method is used to reuse entity slots in the scene when entities are
// deleted and recreated.
func (e *Entity) Revive(index uint32) {
	e.index = index
	e.revision++
}

// Destroy destroys the entity and returns the archetype and archetype row that
// were associated with the entity before it was destroyed.
func (e *Entity) Destroy() (*Archetype, Row) {
	archetype := e.archetype
	archetypeRow := e.archetypeRow

	e.archetype = nil
	e.archetypeRow = 0
	e.revision++

	return archetype, archetypeRow
}

// Revision returns the current revision of the entity, which is incremented
// each time the entity is deleted and recreated.
func (e *Entity) Revision() uint32 {
	return e.revision
}

// HasRevision returns whether the entity's current revision matches the
// specified revision.
func (e *Entity) HasRevision(revision uint32) bool {
	return e.revision == revision
}

// Archetype returns the archetype associated with the entity.
func (e *Entity) Archetype() *Archetype {
	return e.archetype
}

// ArchetypeRow returns the archetype row associated with the entity.
func (e *Entity) ArchetypeRow() Row {
	return e.archetypeRow
}

// Assign changes the archetype and archetype row associated with the entity.
func (e *Entity) Assign(archetype *Archetype, row Row) {
	e.archetype = archetype
	e.archetypeRow = row
	archetype.IDColumn().SetValue(row, NewID(e.index, e.revision))
}
