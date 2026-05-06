package internal

type Entity struct {
	revision uint32

	archetype    *Archetype
	archetypeRow ArchetypeRow
}

// Revive revives the entity with the specified archetype and archetype row.
// This method is used to reuse entity slots in the scene when entities are
// deleted and recreated.
func (e *Entity) Revive(archetype *Archetype, archetypeRow ArchetypeRow) {
	e.revision++
	e.archetype = archetype
	e.archetypeRow = archetypeRow
}

// Destroy destroys the entity and returns the archetype and archetype row that
// were associated with the entity before it was destroyed.
func (e *Entity) Destroy() (*Archetype, ArchetypeRow) {
	archetype := e.archetype
	archetypeRow := e.archetypeRow

	e.revision++
	e.archetype = nil
	e.archetypeRow = 0

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
func (e *Entity) ArchetypeRow() ArchetypeRow {
	return e.archetypeRow
}

// Relocate changes the archetype and archetype row associated with the entity.
func (e *Entity) Relocate(archetype *Archetype, archetypeRow ArchetypeRow) {
	e.archetype = archetype
	e.archetypeRow = archetypeRow
}

// PlacementMap returns a mapping from component type identifiers to storage
// positions for the entity's archetype and archetype row.
func (e *Entity) PlacementMap() TypePlacementMap {
	return e.archetype.PlacementMap(e.archetypeRow)
}
