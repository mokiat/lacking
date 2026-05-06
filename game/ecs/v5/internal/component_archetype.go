package internal

// NewArchetype creates a new Archetype.
func NewArchetype() *Archetype {
	return &Archetype{}
}

// Archetype represents a unique combination of component types. It is used to
// group entities that have the same set of component types together for
// efficient storage and querying.
type Archetype struct {
	mask       TypeMask
	size       uint32
	lookup     TypeLookup
	components []BaseColumn
}

// TypeMask returns the type mask associated with the archetype.
func (a *Archetype) TypeMask() TypeMask {
	return a.mask
}

// PlacementMap returns a mapping from component type identifiers to storage
// positions for the archetype and row.
func (a *Archetype) PlacementMap(row ArchetypeRow) TypePlacementMap {
	offset := uint32(row) % chunkSize
	chunkIndex := uint32(row) / chunkSize

	var result TypePlacementMap
	a.mask.EachType(func(id TypeID) {
		column := a.components[a.lookup[id]]
		result[id].ChunkIndex = uint32(row) / chunkSize
	})
	for i := range result {
		result[i].Offset = offset
	}
	return result
}

// func (a *componentArchetype) reset() {
// 	a.mask = emptyComponentMask()
// 	a.size = 0

// 	// for i := range a.lookup {
// 	// 	a.lookup[i] = -1
// 	// }
// 	// // TODO: Pool component chains as well?
// 	// clear(a.components)
// 	// a.components = a.components[:0]
// }

// func (a *componentArchetype) allocateOffset() uint32 {
// 	// offset := a.size
// 	// a.size++
// 	// return offset
// 	return 0
// }

// func (a *componentArchetype) releaseOffset(offset uint32) {
// 	panic("not implemented")
// }

// ArchetypeRow represents a single row in an archetype, which corresponds to a
// single entity's worth of component data.
type ArchetypeRow uint32
